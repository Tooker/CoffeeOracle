import { OracleExperience } from "@/components/OracleExperience";

// Home renders the landing page content and mounts the interactive oracle workflow.
export default function Home() {
  return (
    <section className="space-y-6 rounded-3xl border-0 bg-coffee-night/80 p-2 shadow-oracle sm:space-y-8 sm:p-4 lg:border lg:border-white/10 lg:p-8">
      <div className="space-y-4">
        <p className="text-sm uppercase tracking-[0.4em] text-coffee-crema">Coffee Oracle</p>
        <h1 className="text-4xl font-semibold leading-tight text-coffee-foam sm:text-5xl">
          Lies im Kaffeeschaum,
          <br /> was die Zukunft bringt.
        </h1>
        <p className="max-w-3xl text-base text-coffee-foam/80">
          Lade ein Bild hoch, gib deinen Namen ein und starte die Orakel-Lesung.
        </p>
      </div>

      <OracleExperience />
    </section>
  );
}
