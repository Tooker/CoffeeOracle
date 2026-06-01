import { apiBaseUrl } from "@/lib/env";

export type OracleUploadPayload = {
  name: string;
  creativity: number;
  questionMode?: boolean;
  question?: string;
  file: File;
};

export type OracleStreamEvent =
  | { type: "chunk"; value: string }
  | { type: "error"; message: string }
  | { type: "share"; url: string }
  | { type: "complete" };

type StreamOptions = {
  formData: FormData;
  signal?: AbortSignal;
  onEvent: (event: OracleStreamEvent) => void;
};

type ParsedEvent = {
  event: string;
  data: string;
};

type CurrentParsedEvent = {
  event: string;
  dataLines: string[];
};

const SCRIPT_REGEX = /<script.*?>[\s\S]*?<\/script>/gi;
const URL_REGEX = /https?:\/\/\S+/gi;
const DISALLOWED_REGEX = /[^a-zA-Z0-9\s\-_'.,äöüÄÖÜß]/g;
const DISALLOWED_QUESTION_REGEX = /[^a-zA-Z0-9\s\-_'.,?!:;()äöüÄÖÜß]/g;
const MAX_NAME_LENGTH = 64;
const MAX_QUESTION_LENGTH = 280;

// sanitizeOracleName strips unsafe or unsupported characters before sending user name to backend.
export function sanitizeOracleName(input: string): string {
  const cleaned = input
    .replace(SCRIPT_REGEX, "")
    .replace(URL_REGEX, "")
    .replace(DISALLOWED_REGEX, "")
    .trim();
  return cleaned.slice(0, MAX_NAME_LENGTH);
}

// sanitizeOracleQuestion keeps normal questions readable while removing noisy prompt-injection content.
export function sanitizeOracleQuestion(input: string): string {
  const cleaned = input
    .replace(SCRIPT_REGEX, "")
    .replace(URL_REGEX, "")
    .replace(DISALLOWED_QUESTION_REGEX, "")
    .replace(/\s+/g, " ")
    .trim();
  return cleaned.slice(0, MAX_QUESTION_LENGTH);
}

// buildOracleFormData prepares multipart form payload in the format expected by /api/oracle.
export function buildOracleFormData(payload: OracleUploadPayload): FormData {
  const data = new FormData();
  data.append("name", sanitizeOracleName(payload.name));
  const boundedCreativity = Math.min(10, Math.max(0, payload.creativity));
  data.append("creativity", String(boundedCreativity));
  const questionMode = Boolean(payload.questionMode);
  data.append("questionMode", String(questionMode));
  if (questionMode) {
    data.append("question", sanitizeOracleQuestion(payload.question ?? ""));
  }
  data.append("file", payload.file);
  return data;
}

// streamOracleResponse sends the request and consumes the SSE stream until completion.
export async function streamOracleResponse({ formData, signal, onEvent }: StreamOptions) {
  const response = await fetch(`${apiBaseUrl}/api/oracle`, {
    method: "POST",
    body: formData,
    signal,
    headers: {
      Accept: "text/event-stream",
    },
  });

  if (!response.ok) {
    let message = `Oracle request failed (${response.status})`;
    try {
      const payload = await response.json();
      if (payload?.error) {
        message = payload.error;
      }
    } catch {
      // ignore JSON parsing issues; fallback to status text
    }
    throw new Error(message);
  }

  if (!response.body) {
    throw new Error("Streaming wird im aktuellen Browser nicht unterstützt.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { done, value } = await reader.read();
    if (value) {
      buffer += decoder.decode(value, { stream: true });
      buffer = drainSSEBuffer(buffer, onEvent);
    }
    if (done) {
      break;
    }
  }

  if (buffer.trim().length > 0) {
    drainSSEBuffer(buffer + "\n\n", onEvent);
  }
}

// drainSSEBuffer parses as many complete SSE blocks as possible and returns the unconsumed remainder.
export function drainSSEBuffer(buffer: string, onEvent: (event: OracleStreamEvent) => void): string {
  let remaining = buffer;
  while (true) {
    const { block, rest } = readNextBlock(remaining);
    if (!block) {
      break;
    }
    remaining = rest;
    const parsedEvents = parseSSEBlock(block);
    for (const event of parsedEvents) {
      dispatchEvent(event, onEvent);
    }
  }
  return remaining;
}

// readNextBlock finds one SSE message block separated by blank lines.
function readNextBlock(input: string): { block: string | null; rest: string } {
  const idxLF = input.indexOf("\n\n");
  const idxCRLF = input.indexOf("\r\n\r\n");

  let idx = -1;
  let delimiterLength = 2;

  if (idxLF >= 0 && (idxCRLF === -1 || idxLF < idxCRLF)) {
    idx = idxLF;
    delimiterLength = 2;
  } else if (idxCRLF >= 0) {
    idx = idxCRLF;
    delimiterLength = 4;
  }

  if (idx === -1) {
    return { block: null, rest: input };
  }

  const block = input.slice(0, idx);
  const rest = input.slice(idx + delimiterLength);
  return { block, rest };
}

// parseSSEBlock converts raw "event:"/"data:" lines into structured events.
function parseSSEBlock(block: string): ParsedEvent[] {
  if (!block.trim()) {
    return [];
  }

  const messages: ParsedEvent[] = [];
  const lines = block.split(/\r?\n/);
  let currentEvent: CurrentParsedEvent = { event: "message", dataLines: [] };

  const pushCurrentEvent = () => {
    if (currentEvent.dataLines.length > 0 || currentEvent.event !== "message") {
      messages.push({
        event: currentEvent.event,
        data: currentEvent.dataLines.join("\n"),
      });
    }
  };

  for (const line of lines) {
    if (line.startsWith("event:")) {
      pushCurrentEvent();
      currentEvent = {
        event: line.slice(6).trim() || "message",
        dataLines: [],
      };
    } else if (line.startsWith("data:")) {
      const rawChunk = line.slice(5);
      const chunk = rawChunk.startsWith(" ") ? rawChunk.slice(1) : rawChunk;
      currentEvent.dataLines.push(chunk);
    }
  }

  pushCurrentEvent();

  return messages;
}

// dispatchEvent maps low-level SSE events to app-level union events consumed by UI state.
function dispatchEvent(event: ParsedEvent, onEvent: (event: OracleStreamEvent) => void) {
  if (event.event === "error" || event.event === "response.error") {
    onEvent({ type: "error", message: event.data || "Unbekannter Fehler" });
    return;
  }

  if (event.event === "complete" || event.event === "response.completed") {
    onEvent({ type: "complete" });
    return;
  }

  if (event.event === "share") {
    try {
      const payload = JSON.parse(event.data) as { url?: string };
      if (payload.url) {
        onEvent({ type: "share", url: payload.url });
      }
    } catch {
      // ignore malformed share metadata; text streaming should still succeed
    }
    return;
  }

  if (event.data) {
    onEvent({ type: "chunk", value: normalizeChunk(event.data) });
  }
}

// mockSSEChunk builds deterministic SSE text for unit tests.
export function mockSSEChunk(event: string, data: string): string {
  return `event: ${event}\ndata: ${data}\n\n`;
}

// normalizeChunk preserves meaningful markdown whitespace while removing protocol-specific noise.
function normalizeChunk(input: string): string {
  const match = input.match(/^\s+/);
  if (!match) {
    return input;
  }

  const prefix = match[0];
  const trimmed = input.slice(prefix.length);
  const newlineCount = (prefix.match(/\n/g) ?? []).length;
  const spaceCount = prefix.replace(/\n/g, "").length;

  let normalizedPrefix = newlineCount > 0 ? "\n".repeat(newlineCount) : "";
  if (spaceCount > 0 && trimmed.length > 0) {
    normalizedPrefix += " ";
  }

  if (trimmed.length === 0) {
    return normalizedPrefix || "";
  }

  return normalizedPrefix + trimmed;
}
