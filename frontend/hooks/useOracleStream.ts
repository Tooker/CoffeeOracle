"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import {
  OracleUploadPayload,
  buildOracleFormData,
  streamOracleResponse,
  OracleStreamEvent,
} from "@/lib/api/oracleClient";

type OracleStatus = "idle" | "uploading" | "streaming" | "error";

type StreamState = {
  status: OracleStatus;
  text: string;
  shareUrl: string | null;
  error: string | null;
};

const initialState: StreamState = {
  status: "idle",
  text: "",
  shareUrl: null,
  error: null,
};

// useOracleStream is a stateful hook that manages upload lifecycle, SSE streaming, and cancellation.
export function useOracleStream() {
  const [state, setState] = useState<StreamState>(initialState);
  const controllerRef = useRef<AbortController | null>(null);

  // submit aborts any running request, sends a new one, and incrementally appends streamed chunks.
  const submit = useCallback(async (payload: OracleUploadPayload) => {
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;

    setState({ status: "uploading", text: "", shareUrl: null, error: null });

    const formData = buildOracleFormData(payload);

    try {
      await streamOracleResponse({
        formData,
        signal: controller.signal,
        onEvent: (event: OracleStreamEvent) => {
          if (event.type === "chunk") {
            setState((prev) => ({
              status: "streaming",
              text: prev.text + event.value,
              shareUrl: prev.shareUrl,
              error: null,
            }));
          } else if (event.type === "error") {
            setState({ status: "error", text: "", shareUrl: null, error: event.message });
            controller.abort();
          } else if (event.type === "share") {
            const shareUrl = new URL(event.url, window.location.origin).toString();
            setState((prev) => ({ ...prev, shareUrl }));
          } else if (event.type === "complete") {
            setState((prev) => ({ ...prev, status: "idle" }));
          }
        },
      });
    } catch (error) {
      if ((error as DOMException)?.name === "AbortError") {
        return;
      }
      const message = error instanceof Error ? error.message : "Unbekannter Fehler";
      setState({ status: "error", text: "", shareUrl: null, error: message });
    }
  }, []);

  // cancel stops the active request but keeps current text in case the user wants to keep reading it.
  const cancel = useCallback(() => {
    controllerRef.current?.abort();
    controllerRef.current = null;
    setState((prev) => ({ ...prev, status: "idle" }));
  }, []);

  // reset fully clears stream output and error state.
  const reset = useCallback(() => {
    controllerRef.current?.abort();
    controllerRef.current = null;
    setState(initialState);
  }, []);

  useEffect(() => {
    return () => {
      controllerRef.current?.abort();
    };
  }, []);

  return {
    submit,
    cancel,
    reset,
    streamText: state.text,
    shareUrl: state.shareUrl,
    status: state.status,
    error: state.error,
  };
}
