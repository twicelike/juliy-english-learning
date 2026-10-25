"use client";

import { useEffect } from "react";
import { AlertTriangle } from "lucide-react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error to an error reporting service
    console.error("Caught in error.tsx:", error);
  }, [error]);

  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center rounded-3xl bg-white/5 border border-white/10 p-8 text-center max-w-lg mx-auto mt-20">
      <div className="mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-red-500/20 text-red-500">
        <AlertTriangle className="h-10 w-10" />
      </div>
      <h2 className="mb-2 text-2xl font-bold">Oops, something went wrong!</h2>
      <p className="mb-8 text-sm opacity-60">
        We ran into an unexpected error. Please try again or refresh the page.
      </p>
      <button
        onClick={() => reset()}
        className="rounded-full bg-white px-8 py-3 text-sm font-semibold text-black transition-colors hover:bg-white/90"
      >
        Try again
      </button>
    </div>
  );
}
