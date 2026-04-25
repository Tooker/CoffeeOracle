import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

type ResultsPanelProps = {
  streamingText?: string;
  isLoading?: boolean;
  error?: string;
  previewImageUrl?: string;
  expanded?: boolean;
  onNewReading?: () => void;
};

// ResultsPanel renders stream progress, markdown output, and fallback/error states.
export function ResultsPanel({
  streamingText = "",
  isLoading = false,
  error,
  previewImageUrl,
  expanded = false,
  onNewReading,
}: ResultsPanelProps) {
  const hasResponse = Boolean(streamingText.trim()) && !error;

  return (
    <aside
      className={`space-y-5 rounded-3xl border-0 bg-coffee-bean/70 text-sm text-coffee-foam/80 lg:border lg:border-white/10 ${expanded ? "oracle-panel-enter p-2 sm:p-4 md:p-8" : "p-2 sm:p-4"} ${hasResponse ? "oracle-response-reveal" : ""}`}
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.35em] text-coffee-crema">Orakel-Ausgabe</p>
          <p className={`oracle-heading-sweep ${expanded ? "text-xl sm:text-2xl md:text-3xl" : "text-lg"} font-semibold text-white`}>
            Prophezeiung
          </p>
        </div>
        {onNewReading ? (
          <button
            type="button"
            onClick={onNewReading}
            className="rounded-xl border border-white/30 px-4 py-2 text-sm font-medium text-coffee-foam/90 transition hover:border-white/60 hover:bg-white/5"
          >
            Neue Lesung
          </button>
        ) : null}
      </div>
      <div
        className={`space-y-3 rounded-2xl border-0 bg-black/30 lg:border lg:border-white/10 ${expanded ? "min-h-[280px] p-2 sm:min-h-[320px] sm:p-3 md:p-6" : "min-h-[180px] p-2 sm:p-3"}`}
      >
        {previewImageUrl ? (
          <figure className="oracle-cup-drift overflow-hidden rounded-xl border-0 bg-black/30 lg:border lg:border-white/10">
            <img
              src={previewImageUrl}
              alt="Hochgeladenes Kaffeeschaumbild"
              className="h-44 w-full object-cover md:h-56"
            />
            <figcaption className="border-t border-white/10 px-3 py-2 text-xs uppercase tracking-[0.16em] text-coffee-foam/65">
              Deine Tasse
            </figcaption>
          </figure>
        ) : null}

        {error ? (
          <p className="text-red-200">{error}</p>
        ) : streamingText ? (
          <div className="space-y-3">
            <div className={`space-y-3 text-coffee-foam ${expanded ? "sm:text-lg" : ""}`}>
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  h1: ({ ...props }) => <h1 className="text-xl font-semibold text-white sm:text-2xl md:text-3xl" {...props} />,
                  h2: ({ ...props }) => <h2 className="text-lg font-semibold text-white sm:text-xl md:text-2xl" {...props} />,
                  h3: ({ ...props }) => <h3 className="text-base font-semibold text-white sm:text-lg md:text-xl" {...props} />,
                  p: ({ ...props }) => <p className="leading-relaxed" {...props} />,
                  ul: ({ ...props }) => <ul className="list-disc space-y-1 pl-6" {...props} />,
                  ol: ({ ...props }) => <ol className="list-decimal space-y-1 pl-6" {...props} />,
                  li: ({ ...props }) => <li className="leading-relaxed" {...props} />,
                  blockquote: ({ ...props }) => (
                    <blockquote
                      className="border-l-2 border-coffee-crema/60 pl-4 italic text-coffee-foam/90"
                      {...props}
                    />
                  ),
                  code: ({ className, children, ...props }) => {
                    const isBlock = Boolean(className);
                    if (isBlock) {
                      return (
                        <code
                          className="my-2 block overflow-x-auto rounded-lg bg-black/35 px-3 py-2 text-sm text-coffee-crema"
                          {...props}
                        >
                          {children}
                        </code>
                      );
                    }
                    return (
                      <code className="rounded bg-black/35 px-1.5 py-0.5 text-coffee-crema" {...props}>
                        {children}
                      </code>
                    );
                  },
                  a: ({ ...props }) => (
                    <a
                      className="text-coffee-crema underline decoration-coffee-crema/60 underline-offset-2"
                      target="_blank"
                      rel="noreferrer"
                      {...props}
                    />
                  ),
                  hr: ({ ...props }) => <hr className="border-white/15" {...props} />,
                }}
              >
                {streamingText}
              </ReactMarkdown>
            </div>
            {isLoading && <p className="text-xs text-coffee-crema/70">Das Orakel orakelt …</p>}
          </div>
        ) : isLoading ? (
          <div className="flex min-h-[220px] flex-col items-center justify-center gap-4 text-center">
            <div className="oracle-reading-aura" aria-hidden>
              <div className="oracle-orb" />
            </div>
            <p className="oracle-pulse-text text-base text-coffee-crema/90 md:text-lg">
              Das Orakel sammelt Zeichen aus dem Schaum ...
            </p>
          </div>
        ) : (
          <p className="text-coffee-foam/60">
            Lade ein Bild hoch, um hier die erste Vision zu empfangen. Der Stream aktualisiert
            sich live, sobald das Go-Backend antwortet.
          </p>
        )}
      </div>
    </aside>
  );
}
