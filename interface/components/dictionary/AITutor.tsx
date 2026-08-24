"use client";

import { useState, useRef } from "react";
import {
  AIAPI,
  NuanceResponse,
  ExamplesResponse,
} from "@/lib/ai-api";

interface AITutorProps {
  word: string;
  pos?: string;
}

type Feature = "nuance" | "examples";

const BUTTONS: { id: Feature; label: string }[] = [
  { id: "nuance", label: "Nuance" },
  { id: "examples", label: "More Examples" },
];

export default function AITutor({ word, pos }: AITutorProps) {
  const [loading, setLoading] = useState<Feature | null>(null);
  const [nuanceData, setNuanceData] = useState<NuanceResponse | null>(null);
  const [examplesData, setExamplesData] = useState<ExamplesResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const nuanceRef = useRef<HTMLDivElement>(null);
  const examplesRef = useRef<HTMLDivElement>(null);

  const handleClick = async (feature: Feature) => {
    setLoading(feature);
    setError(null);
    try {
      const req = { word, pos: pos ?? "" };
      if (feature === "nuance") {
        const data = await AIAPI.explainNuance(req);
        setNuanceData(data);
        setTimeout(() => nuanceRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" }), 100);
      } else if (feature === "examples") {
        const data = await AIAPI.generateExamples(req);
        setExamplesData(data);
        setTimeout(() => examplesRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" }), 100);
      }
    } catch (e: any) {
      setError(e?.message ?? "Service is temporarily unavailable.");
    } finally {
      setLoading(null);
    }
  };

  const isFetched = (id: Feature) => id === "nuance" ? !!nuanceData : !!examplesData;

  return (
    <div className="mt-10 border-t border-white/10 pt-8">
      <div className="flex items-center gap-2 mb-5">
        <h2 className="text-white font-semibold text-lg">Word Analysis</h2>
      </div>

      {/* Action buttons */}
      <div className="flex flex-wrap gap-2 mb-6">
        {BUTTONS.map((btn) => (
          <button
            key={btn.id}
            onClick={() => handleClick(btn.id)}
            disabled={loading !== null || isFetched(btn.id)}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-medium transition-all border ${
              loading === btn.id || isFetched(btn.id)
                ? "bg-white/10 border-white/20 text-white cursor-default"
                : loading !== null
                  ? "bg-white/5 border-white/5 text-white/30 cursor-default"
                  : "bg-white/5 border-white/10 text-white/70 hover:bg-white/10 hover:text-white hover:border-white/20"
            }`}
          >
            {loading === btn.id && (
              <span className="w-3 h-3 border border-indigo-400 border-t-transparent rounded-full animate-spin" />
            )}
            {btn.label}
          </button>
        ))}
      </div>

      {error && (
        <div className="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 text-red-300 text-sm">
          {error}
        </div>
      )}

      {/* Results */}
      {(nuanceData || examplesData) && (
        <div className="rounded-3xl bg-white/5 border border-white/10 p-6 space-y-6 animate-in fade-in duration-300">
          {nuanceData && (
            <div ref={nuanceRef}>
              <Section title="Nuance">
                <div className="space-y-4">
                  <div>
                    <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Core Meaning</p>
                    <p>{nuanceData.coreMeaning}</p>
                  </div>
                  <div>
                    <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Subtle Meaning</p>
                    <p>{nuanceData.subtleMeaning}</p>
                  </div>
                  {nuanceData.usageNotes?.length > 0 && (
                    <div>
                      <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Usage Notes</p>
                      <ul className="list-disc pl-4 space-y-1">
                        {nuanceData.usageNotes.map((note, i) => <li key={i}>{note}</li>)}
                      </ul>
                    </div>
                  )}
                  {nuanceData.confusedWith?.length > 0 && (
                    <div>
                      <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Confused With</p>
                      <div className="flex gap-2">
                        {nuanceData.confusedWith.map((word, i) => (
                          <span key={i} className="px-2 py-1 bg-white/10 rounded text-xs">{word}</span>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </Section>
            </div>
          )}

          {nuanceData && examplesData && (
            <hr className="border-white/10" />
          )}

          {examplesData && (
            <div ref={examplesRef}>
              <Section title="More Examples">
                <ul className="space-y-3">
                  {examplesData.examples.map((ex, i) => (
                    <li key={i} className="text-white/80 italic border-l-2 border-indigo-500/30 pl-3">
                      "{ex}"
                    </li>
                  ))}
                </ul>
              </Section>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// ── Sub-components ────────────────────────────────────────────────────────────

function SectionHeader({ title }: { title: string }) {
  return (
    <p className="text-white/50 text-xs font-semibold uppercase tracking-wider">
      {title}
    </p>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <SectionHeader title={title} />
      <div className="text-white/80 text-sm leading-relaxed mt-1">{children}</div>
    </div>
  );
}
