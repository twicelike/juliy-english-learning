"use client";

import { useState } from "react";
import {
  AIAPI,
  ExplainResponse,
  NuanceResponse,
  ExamplesResponse,
  MnemonicResponse,
} from "@/lib/ai-api";

interface AITutorProps {
  word: string;
  pos?: string;
}

type Feature = "explain" | "nuance" | "examples" | "mnemonic";

type Result =
  | { feature: "explain"; data: ExplainResponse }
  | { feature: "nuance"; data: NuanceResponse }
  | { feature: "examples"; data: ExamplesResponse }
  | { feature: "mnemonic"; data: MnemonicResponse };

const BUTTONS: { id: Feature; label: string; emoji: string }[] = [
  { id: "explain", label: "Explain Word", emoji: "🧠" },
  { id: "nuance", label: "Explain Nuances", emoji: "🎯" },
  { id: "examples", label: "More Examples", emoji: "📝" },
  { id: "mnemonic", label: "Memory Trick", emoji: "💡" },
];

export default function AITutor({ word, pos }: AITutorProps) {
  const [loading, setLoading] = useState<Feature | null>(null);
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleClick = async (feature: Feature) => {
    setLoading(feature);
    setError(null);
    setResult(null);
    try {
      const req = { word, pos: pos ?? "" };
      if (feature === "explain") {
        const data = await AIAPI.explainWord(req);
        setResult({ feature, data });
      } else if (feature === "nuance") {
        const data = await AIAPI.explainNuance(req);
        setResult({ feature, data });
      } else if (feature === "examples") {
        const data = await AIAPI.generateExamples(req);
        setResult({ feature, data });
      } else {
        const data = await AIAPI.createMnemonic(req);
        setResult({ feature, data });
      }
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
            disabled={loading !== null}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-medium transition-all border ${
              loading === btn.id
                ? "bg-indigo-500/30 border-indigo-500/50 text-indigo-300 cursor-wait"
                : "bg-white/5 border-white/10 text-white/70 hover:bg-white/10 hover:text-white hover:border-indigo-500/40 disabled:opacity-40 disabled:cursor-not-allowed"
            }`}
          >
            {loading === btn.id ? (
              <span className="w-3 h-3 border border-indigo-400 border-t-transparent rounded-full animate-spin" />
            ) : (
              <span>{btn.emoji}</span>
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
      {result && (
        <div className="rounded-3xl bg-white/5 border border-white/10 p-6 space-y-4 animate-in fade-in duration-300">
          {result.feature === "explain" && (
            <>
              <Section title="Summary" icon="📖">{result.data.summary}</Section>
              <Section title="Nuance" icon="🎯">{result.data.nuance}</Section>
              {result.data.commonMistakes.length > 0 && (
                <div>
                  <SectionHeader icon="⚠️" title="Common Mistakes" />
                  <ul className="space-y-1 mt-2">
                    {result.data.commonMistakes.map((m, i) => (
                      <li key={i} className="text-white/60 text-sm flex gap-2">
                        <span className="text-red-400 shrink-0">•</span>{m}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              {result.data.examples.length > 0 && (
                <div>
                  <SectionHeader icon="💬" title="Examples" />
                  <ul className="space-y-1 mt-2">
                    {result.data.examples.map((ex, i) => (
                      <li key={i} className="text-white/60 text-sm italic">"{ex}"</li>
                    ))}
                  </ul>
                </div>
              )}
              {result.data.register && (
                <p className="text-white/40 text-xs">
                  Register: <span className="text-indigo-300 capitalize">{result.data.register}</span>
                </p>
              )}
            </>
          )}

          {result.feature === "nuance" && (
            <>
              <Section title="Core Meaning" icon="🔑">{result.data.coreMeaning}</Section>
              <Section title="Subtle Meaning" icon="✨">{result.data.subtleMeaning}</Section>
              {result.data.usageNotes.length > 0 && (
                <div>
                  <SectionHeader icon="📌" title="Usage Notes" />
                  <ul className="space-y-1 mt-2">
                    {result.data.usageNotes.map((n, i) => (
                      <li key={i} className="text-white/60 text-sm flex gap-2"><span className="text-indigo-400">•</span>{n}</li>
                    ))}
                  </ul>
                </div>
              )}
              {result.data.confusedWith.length > 0 && (
                <div className="flex items-center gap-2 flex-wrap">
                  <p className="text-white/40 text-xs">Often confused with:</p>
                  {result.data.confusedWith.map((w, i) => (
                    <span key={i} className="px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 text-xs">{w}</span>
                  ))}
                </div>
              )}
            </>
          )}

          {result.feature === "examples" && (
            <div className="space-y-2">
              {result.data.examples.map((ex, i) => (
                <p key={i} className="text-white/70 text-sm italic border-l-2 border-indigo-500/40 pl-3">"{ex}"</p>
              ))}
            </div>
          )}

          {result.feature === "mnemonic" && (
            <>
              <Section title="Memory Trick" icon="💡">{result.data.mnemonic}</Section>
              <Section title="Story" icon="📖">{result.data.story}</Section>
              <Section title="Visual Image" icon="🖼️">{result.data.visual}</Section>
            </>
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
      <p className="text-white/80 text-sm leading-relaxed mt-1">{children}</p>
    </div>
  );
}
