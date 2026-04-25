"use client";

import { Button } from "flowbite-react";
import Link from "next/link";

// HeroActions shows optional CTA buttons for navigation/marketing entry points.
export function HeroActions() {
  return (
    <div className="flex flex-wrap gap-3">
      <Button color="pink" className="min-w-[180px]">
        Launch upcoming UI
      </Button>
      <Link
        href="https://nextjs.org/docs"
        className="inline-flex items-center rounded-full border border-white/20 px-5 py-3 text-sm font-medium text-coffee-foam/80 transition hover:border-white/50"
      >
        View platform docs
      </Link>
    </div>
  );
}
