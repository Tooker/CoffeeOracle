"use client";

import { useEffect, useRef, useState } from "react";

import { UploadForm, UploadPayload } from "@/components/UploadForm";
import { ResultsPanel } from "@/components/ResultsPanel";
import { useOracleStream } from "@/hooks/useOracleStream";

// OracleExperience orchestrates the UI state between upload form and streaming result panel.
export function OracleExperience() {
  const { submit, reset, status, streamText, error } = useOracleStream();
  const [formVersion, setFormVersion] = useState(0);
  const [previewImageUrl, setPreviewImageUrl] = useState<string | null>(null);
  const previewUrlRef = useRef<string | null>(null);

  // setPreviewFromFile manages browser object URLs so image previews work without memory leaks.
  const setPreviewFromFile = (file: File | null) => {
    if (previewUrlRef.current) {
      URL.revokeObjectURL(previewUrlRef.current);
      previewUrlRef.current = null;
    }

    if (!file) {
      setPreviewImageUrl(null);
      return;
    }

    const nextUrl = URL.createObjectURL(file);
    previewUrlRef.current = nextUrl;
    setPreviewImageUrl(nextUrl);
  };

  useEffect(() => {
    return () => {
      if (previewUrlRef.current) {
        URL.revokeObjectURL(previewUrlRef.current);
      }
    };
  }, []);

  // handleSubmit starts a new reading and stores a local preview of the uploaded image.
  const handleSubmit = (payload: UploadPayload) => {
    setPreviewFromFile(payload.file);
    submit(payload);
  };

  // handleNewReading resets everything so the user can start with a clean form.
  const handleNewReading = () => {
    reset();
    setPreviewFromFile(null);
    setFormVersion((prev) => prev + 1);
  };

  const isBusy = status === "uploading" || status === "streaming";
  const showOracle = isBusy || streamText.trim().length > 0 || Boolean(error);

  return (
    <div className="space-y-8">
      {!showOracle ? (
        <div className="rounded-3xl border border-white/10 bg-white/5 p-8 shadow-xl transition-all duration-500 md:p-10">
          <UploadForm
            key={`upload-form-${formVersion}`}
            onSubmit={handleSubmit}
            onReset={reset}
            isSubmitting={isBusy}
          />
        </div>
      ) : null}

      {showOracle ? (
        <ResultsPanel
          expanded
          streamingText={streamText}
          isLoading={isBusy}
          error={error ?? undefined}
          previewImageUrl={previewImageUrl ?? undefined}
          onNewReading={handleNewReading}
        />
      ) : null}
    </div>
  );
}
