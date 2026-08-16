"use client";

import { useState } from "react";
import { AIAPI, BatchSenseExplanation } from "@/lib/ai-api";
import { useAuthStore } from "@/features/auth";

interface SenseExplanationProps {
  wordId: string;
  senseIndex: number;
}

export default function SenseExplanation({ wordId, senseIndex }: SenseExplanationProps) {
  const [loading, setLoading] = useState(false);
  const [explanation, setExplanation] = useState<BatchSenseExplanation | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { user } = useAuthStore();

  const handleExplain = async () => {
    if (!user) return; // Should be handled by parent, but just in case

    setLoading(true);
    setError(null);
    try {
      const data = await AIAPI.explainSense({ wordId, senseIndex });
      setExplanation(data);
    } catch (e: any) {
      setError(e?.message ?? "AI is unavailable right now.");
    } finally {
      setLoading(false);
    }
  };

  if (!user) return null;

  return (
    <div className="mt-3">
      {!explanation && !loading && !error && (
        <button
          onClick={handleExplain}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all border bg-indigo-500/10 border-indigo-500/30 text-indigo-300 hover:bg-indigo-500/20 hover:text-white"
        >
          Explain this meaning
        </button>
      )}

      {loading && (
        <div className="flex items-center gap-2 text-indigo-300 text-sm italic">
          <span className="w-3 h-3 border border-indigo-400 border-t-transparent rounded-full animate-spin" />
          Analyzing...
        </div>
      )}

      {error && (
        <div className="p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs">
          {error}
        </div>
      )}

      {explanation && (
        <div className="mt-2 p-4 rounded-2xl bg-indigo-500/5 border border-indigo-500/20 animate-in fade-in duration-300">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-indigo-300 text-xs font-medium uppercase tracking-wider">Explanation</span>
          </div>

          <p className="text-white/90 text-sm leading-relaxed mb-3">
            {explanation.explanation}
          </p>

          {explanation.usageNote && (
            <div className="mb-3">
              <p className="text-indigo-400/70 text-[10px] font-bold uppercase tracking-wider mb-1">Usage Note</p>
              <p className="text-white/70 text-sm italic border-l-2 border-indigo-500/30 pl-2">
                {explanation.usageNote}
              </p>
            </div>
          )}

          {explanation.examples && explanation.examples.length > 0 && (
            <div className="mb-3">
              <p className="text-indigo-400/70 text-[10px] font-bold uppercase tracking-wider mb-1">More Examples</p>
              <ul className="space-y-1">
                {explanation.examples.map((ex, i) => (
                  <li key={i} className="text-white/60 text-sm italic flex gap-2">
                    <span className="text-indigo-500/50 shrink-0">•</span> {ex}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {explanation.commonMistakes && explanation.commonMistakes.length > 0 && (
            <div>
              <p className="text-red-400/70 text-[10px] font-bold uppercase tracking-wider mb-1">⚠️ Common Mistakes</p>
              <ul className="space-y-1">
                {explanation.commonMistakes.map((m, i) => (
                  <li key={i} className="text-white/60 text-sm flex gap-2">
                    <span className="text-red-400/50 shrink-0">•</span> {m}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
