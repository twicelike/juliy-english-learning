"use client";

import { useState } from "react";
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

type Result =
  | { feature: "nuance"; data: NuanceResponse }
  | { feature: "examples"; data: ExamplesResponse };

const BUTTONS: { id: Feature; label: string; emoji: string }[] = [
  { id: "nuance", label: "Nuance", emoji: "🎯" },
  { id: "examples", label: "More Examples", emoji: "📝" },
];

export default function AITutor({ word, pos }: AITutorProps) {
  const [loading, setLoading] = useState<Feature | null>(null);
  const [result, setResult] = useState<Result | null>(null);
  const [hasFetched, setHasFetched] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleClick = async (feature: Feature) => {
    setLoading(feature);
    setError(null);
    setResult(null);
    try {
      const req = { word, pos: pos ?? "" };
      if (feature === "nuance") {
        const data = await AIAPI.explainNuance(req);
        setResult({ feature, data });
      } else if (feature === "examples") {
        const data = await AIAPI.generateExamples(req);
        setResult({ feature, data });
      }
      setHasFetched(true);
    } catch (e: any) {
      setError(e?.message ?? "AI is unavailable right now.");
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="mt-10 border-t border-white/10 pt-8">
      <div className="flex items-center gap-2 mb-5">
        <span className="text-lg">🤖</span>
        <h2 className="text-white font-semibold text-lg">AI Tutor</h2>
        <span className="ml-1 px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-xs">Powered by Gemini</span>
      </div>

      {/* Action buttons */}
      <div className="flex flex-wrap gap-2 mb-6">
        {BUTTONS.map((btn) => (
          <button
            key={btn.id}
            onClick={() => handleClick(btn.id)}
            disabled={loading !== null || hasFetched}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-medium transition-all border ${
              loading === btn.id
                ? "bg-indigo-500/30 border-indigo-500/50 text-indigo-300 cursor-wait"
                : hasFetched 
                  ? "bg-indigo-500/20 border-indigo-500/40 text-indigo-200 cursor-default"
                  : "bg-white/5 border-white/10 text-white/70 hover:bg-white/10 hover:text-white hover:border-indigo-500/40"
            }`}
          >
            {loading === btn.id ? (
              <span className="w-3 h-3 border border-indigo-400 border-t-transparent rounded-full animate-spin" />
            ) : hasFetched ? (
              <span>✅</span>
            ) : (
              <span>{btn.emoji}</span>
            )}
            {hasFetched ? "Explanation Loaded" : btn.label}
          </button>
        ))}
      </div>

      {error && (
        <div className="p-4 rounded-2xl bg-red-500/10 border border-red-500/20 text-red-300 text-sm">
          {error}
        </div>
      )}

      {/* Results */}
      {result && (
        <div className="rounded-3xl bg-white/5 border border-white/10 p-6 space-y-4 animate-in fade-in duration-300">
          {result.feature === "nuance" && (
            <Section title="Nuance" icon="🎯">
              <div className="space-y-4">
                <div>
                  <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Core Meaning</p>
                  <p>{result.data.coreMeaning}</p>
                </div>
                <div>
                  <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Subtle Meaning</p>
                  <p>{result.data.subtleMeaning}</p>
                </div>
                {result.data.usageNotes?.length > 0 && (
                  <div>
                    <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Usage Notes</p>
                    <ul className="list-disc pl-4 space-y-1">
                      {result.data.usageNotes.map((note, i) => <li key={i}>{note}</li>)}
                    </ul>
                  </div>
                )}
                {result.data.confusedWith?.length > 0 && (
                  <div>
                    <p className="text-white/50 text-xs font-semibold uppercase tracking-wider mb-1">Confused With</p>
                    <div className="flex gap-2">
                      {result.data.confusedWith.map((word, i) => (
                        <span key={i} className="px-2 py-1 bg-white/10 rounded text-xs">{word}</span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </Section>
          )}

          {result.feature === "examples" && (
            <Section title="More Examples" icon="📝">
              <ul className="space-y-3">
                {result.data.examples.map((ex, i) => (
                  <li key={i} className="text-white/80 italic border-l-2 border-indigo-500/30 pl-3">
                    "{ex}"
                  </li>
                ))}
              </ul>
            </Section>
          )}


        </div>
      )}
    </div>
  );
}

// ── Sub-components ────────────────────────────────────────────────────────────

function SectionHeader({ icon, title }: { icon: string; title: string }) {
  return (
    <p className="text-white/50 text-xs font-semibold uppercase tracking-wider flex gap-1.5 items-center">
      <span>{icon}</span>{title}
    </p>
  );
}

function Section({ title, icon, children }: { title: string; icon: string; children: React.ReactNode }) {
  return (
    <div>
      <SectionHeader icon={icon} title={title} />
      <div className="text-white/80 text-sm leading-relaxed mt-1">{children}</div>
    </div>
  );
}
