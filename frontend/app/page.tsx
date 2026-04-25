import { OracleExperience } from "@/components/OracleExperience";

export default function Home() {
  return (
    <section className="space-y-8 rounded-3xl border border-white/10 bg-coffee-night/80 p-8 shadow-oracle">
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
