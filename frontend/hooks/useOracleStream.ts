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
  error: string | null;
};

const initialState: StreamState = {
  status: "idle",
  text: "",
  error: null,
};

export function useOracleStream() {
  const [state, setState] = useState<StreamState>(initialState);
  const controllerRef = useRef<AbortController | null>(null);

  const submit = useCallback(async (payload: OracleUploadPayload) => {
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;

    setState({ status: "uploading", text: "", error: null });

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
              error: null,
            }));
          } else if (event.type === "error") {
            setState({ status: "error", text: "", error: event.message });
            controller.abort();
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
      setState({ status: "error", text: "", error: message });
    }
  }, []);

  const cancel = useCallback(() => {
    controllerRef.current?.abort();
    controllerRef.current = null;
    setState((prev) => ({ ...prev, status: "idle" }));
  }, []);

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
    status: state.status,
    error: state.error,
  };
}
