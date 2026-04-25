import {
  sanitizeOracleName,
  drainSSEBuffer,
  mockSSEChunk,
} from "@/lib/api/oracleClient";

describe("sanitizeOracleName", () => {
  it("strips scripts, urls, and unsupported chars", () => {
    const dirty = "<script>alert(1)</script>http://evil.com Alex✨";
    expect(sanitizeOracleName(dirty)).toBe("Alex");
  });

  it("truncates to 64 characters", () => {
    const long = "a".repeat(80);
    expect(sanitizeOracleName(long)).toHaveLength(64);
  });
});

describe("drainSSEBuffer", () => {
  it("emits chunk and completion events", () => {
    const events: string[] = [];
    const buffer = `${mockSSEChunk("message", "Hallo")}${mockSSEChunk("complete", "done")}`;

    drainSSEBuffer(buffer, (event) => {
      if (event.type === "chunk") {
        events.push(event.value.trimStart());
      }
      if (event.type === "complete") {
        events.push("complete");
      }
    });

    expect(events).toEqual(["Hallo", "complete"]);
  });

  it("strips the optional sse space after data colon", () => {
    const events: string[] = [];

    drainSSEBuffer("event: message\ndata: H\n\nevent: message\ndata: a\n\n", (event) => {
      if (event.type === "chunk") {
        events.push(event.value);
      }
    });

    expect(events.join("")).toBe("Ha");
  });

  it("preserves leading newlines in chunk data", () => {
    const events: string[] = [];

    drainSSEBuffer("event: message\ndata: \ndata: ## Deutung\n\n", (event) => {
      if (event.type === "chunk") {
        events.push(event.value);
      }
    });

    expect(events.join("")).toBe("\n## Deutung");
  });

  it("returns remainder when block incomplete", () => {
    const remainder = drainSSEBuffer("data: hi", () => undefined);
    expect(remainder).toBe("data: hi");
  });
});
