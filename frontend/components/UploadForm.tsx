"use client";

import { useEffect, useId, useRef, useState } from "react";

export type UploadPayload = {
  name: string;
  creativity: number;
  file: File;
};

type UploadFormProps = {
  onSubmit?: (payload: UploadPayload) => void;
  onReset?: () => void;
  isSubmitting?: boolean;
};

type MotionPermissionState = "idle" | "active" | "denied" | "unsupported";

type DeviceOrientationEventWithPermission = typeof DeviceOrientationEvent & {
  requestPermission?: () => Promise<PermissionState>;
};

// UploadForm collects user input (name, creativity, image) and validates minimal client-side requirements.
export function UploadForm({ onSubmit, onReset, isSubmitting = false }: UploadFormProps) {
  const formId = useId();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [name, setName] = useState("");
  const [creativity, setCreativity] = useState(5);
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [fileInputResetKey, setFileInputResetKey] = useState(0);
  const [motionStatus, setMotionStatus] = useState<MotionPermissionState>("idle");

  // handleSubmit blocks empty uploads and forwards normalized form values to the parent component.
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);

    if (!file) {
      setError("Bitte lade zuerst ein Kaffeeschaumfoto hoch.");
      return;
    }

    onSubmit?.({ name: name.trim(), creativity, file });
  };

  // enableMotionControl asks the browser for motion access, then lets device tilt drive the slider.
  const enableMotionControl = async () => {
    if (motionStatus === "active") {
      setMotionStatus("idle");
      return;
    }

    if (typeof window === "undefined" || !("DeviceOrientationEvent" in window)) {
      setMotionStatus("unsupported");
      return;
    }

    const orientationEvent = DeviceOrientationEvent as DeviceOrientationEventWithPermission;
    if (typeof orientationEvent.requestPermission === "function") {
      const permission = await orientationEvent.requestPermission();
      if (permission !== "granted") {
        setMotionStatus("denied");
        return;
      }
    }

    setMotionStatus("active");
  };

  useEffect(() => {
    if (motionStatus !== "active") {
      return undefined;
    }

    const handleOrientation = (event: DeviceOrientationEvent) => {
      const rawTilt = event.gamma ?? (event.beta === null ? null : event.beta / 2);
      if (rawTilt === null) {
        return;
      }

      const boundedTilt = Math.max(-45, Math.min(45, rawTilt));
      const nextCreativity = Math.round(((boundedTilt + 45) / 90) * 10);
      setCreativity(nextCreativity);
    };

    window.addEventListener("deviceorientation", handleOrientation);
    return () => window.removeEventListener("deviceorientation", handleOrientation);
  }, [motionStatus]);

  return (
    <form aria-label="Coffee Oracle Upload Form" className="space-y-7" onSubmit={handleSubmit}>
      <div className="grid gap-5 lg:grid-cols-2">
        <div className="space-y-3">
          <label
            htmlFor={`${formId}-name`}
            className="text-sm font-semibold uppercase tracking-[0.18em] text-coffee-crema"
          >
            Dein Name
          </label>
          <input
            id={`${formId}-name`}
            type="text"
            maxLength={64}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            placeholder="z. B. Alex"
            className="w-full rounded-xl border border-white/25 bg-black/20 px-4 py-3 text-lg text-coffee-foam placeholder:text-coffee-foam/40 outline-none transition focus:border-coffee-crema/80 focus:ring-2 focus:ring-coffee-crema/30"
          />
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <label
              htmlFor={`${formId}-creativity`}
              className="text-sm font-semibold uppercase tracking-[0.18em] text-coffee-crema"
            >
              Esoterik-Stufe
            </label>
            <span className="rounded-full border border-white/20 bg-white/5 px-3 py-1 text-xs font-semibold text-coffee-foam/80">
              {creativity} / 10
            </span>
          </div>
          <input
            className="oracle-range"
            id={`${formId}-creativity`}
            type="range"
            min={0}
            max={10}
            step={1}
            value={creativity}
            onChange={(e) => setCreativity(Number(e.target.value))}
          />
          <div className="flex flex-wrap items-center gap-3 text-xs text-coffee-foam/65 md:hidden">
            <button
              type="button"
              onClick={enableMotionControl}
              className="rounded-full border border-coffee-crema/35 px-3 py-1.5 font-semibold text-coffee-crema transition hover:border-coffee-crema hover:bg-coffee-crema/10"
            >
              {motionStatus === "active" ? "Neigung ausschalten" : "Neigung nutzen"}
            </button>
            <span>
              {motionStatus === "active"
                ? "Aktiv: Kippe dein Gerät nach links oder rechts."
                : motionStatus === "denied"
                  ? "Bewegungssensor wurde nicht freigegeben."
                  : motionStatus === "unsupported"
                    ? "Dieses Gerät oder dieser Browser unterstützt das nicht."
                    : "Optional: steuere die Stufe per Geräte-Neigung."}
            </span>
          </div>
        </div>
      </div>

      <div className="space-y-3">
        <label
          htmlFor={`${formId}-file`}
          className="text-sm font-semibold uppercase tracking-[0.18em] text-coffee-crema"
        >
          Kaffeeschaumfoto
        </label>
        <div className="rounded-2xl border border-coffee-crema/20 bg-gradient-to-r from-white/5 via-white/[0.03] to-white/5 p-4">
          <input
            key={fileInputResetKey}
            ref={fileInputRef}
            id={`${formId}-file`}
            type="file"
            accept="image/png,image/jpeg,image/webp"
            className="sr-only"
            onChange={(event) => {
              const nextFile = event.target.files?.[0];
              setFile(nextFile ?? null);
            }}
          />
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className="inline-flex items-center rounded-lg border border-coffee-crema/50 bg-coffee-bean/50 px-4 py-2 text-sm font-semibold text-coffee-foam transition hover:border-coffee-crema hover:bg-coffee-bean/70"
          >
            Datei auswahlen
          </button>
          <p className="mt-3 truncate text-sm text-coffee-foam/75">
            {file ? file.name : "Keine Datei ausgewahlt"}
          </p>
        </div>
      </div>

      {error ? (
        <div className="rounded-xl border border-red-400/40 bg-red-900/20 px-4 py-3 text-sm text-red-300">
          {error}
        </div>
      ) : null}

      <div className="flex flex-wrap gap-3 pt-1">
        <button
          type="submit"
          className="oracle-ritual-button min-w-[220px] rounded-xl border border-coffee-crema/50 bg-coffee-crema/15 px-6 py-3 text-base font-semibold text-coffee-foam transition hover:bg-coffee-crema/25 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSubmitting}
        >
          {isSubmitting ? "Orakel wird kontaktiert" : "Coffee Fortune anfragen"}
        </button>
        <button
          type="button"
          className="rounded-xl border border-white/35 bg-transparent px-6 py-3 text-base font-semibold text-coffee-foam/90 transition hover:border-white/60 hover:bg-white/5"
          onClick={() => {
            setName("");
            setCreativity(5);
            setFile(null);
            setError(null);
            setFileInputResetKey((prev) => prev + 1);
            onReset?.();
          }}
        >
          Formular zurücksetzen
        </button>
      </div>
    </form>
  );
}
