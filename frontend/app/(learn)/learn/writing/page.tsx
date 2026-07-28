"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";
import { AIAPI, GrammarResponse, VocabSuggestionResponse } from "@/lib/ai-api";

type Mode = "grammar" | "vocab";

export default function WritingPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [text, setText] = useState("");
  const [mode, setMode] = useState<Mode>("grammar");
  const [loading, setLoading] = useState(false);
  const [grammarResult, setGrammarResult] = useState<GrammarResponse | null>(null);
  const [vocabResult, setVocabResult] = useState<VocabSuggestionResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!user) router.push("/login");
  }, [user, router]);

  if (!user) return null;

  const handleSubmit = async () => {
    if (text.trim().length < 3) return;
    setLoading(true);
    setError(null);
    setGrammarResult(null);
    setVocabResult(null);
    try {
      if (mode === "grammar") {
        const r = await AIAPI.checkGrammar({ text });
        setGrammarResult(r);
      } else {
        const r = await AIAPI.suggestVocab({ text });
        setVocabResult(r);
      }
    } catch (e: any) {
      setError(e?.message ?? "Something went wrong.");
    } finally {
      setLoading(false);
    }
  };

  const scoreColor = (s: number) =>
    s >= 80 ? "text-emerald-400" : s >= 60 ? "text-amber-400" : "text-red-400";

  return (
    <div className="min-h-[calc(100vh-4rem)] px-6 py-12 max-w-2xl mx-auto">
      <h1 className="text-3xl font-bold text-white mb-2">Writing Lab</h1>
      <p className="text-white/50 mb-8 text-sm">Check your grammar or improve your vocabulary with AI.</p>

      {/* Mode Toggle */}
      <div className="flex gap-2 mb-6">
        {(["grammar", "vocab"] as Mode[]).map((m) => (
          <button
            key={m}
            onClick={() => setMode(m)}
            className={`px-5 py-2 rounded-xl text-sm font-medium transition-all ${
              mode === m
                ? "bg-indigo-500 text-white shadow-lg shadow-indigo-500/30"
                : "bg-white/5 text-white/50 hover:bg-white/10 border border-white/10"
            }`}
          >
            {m === "grammar" ? "✏️ Check Grammar" : "💡 Suggest Better Vocab"}
          </button>
        ))}
      </div>

      {/* Text input */}
      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="Type or paste your English text here…"
        rows={6}
        className="w-full rounded-2xl bg-white/5 border border-white/10 text-white/90 placeholder-white/20 p-4 text-sm resize-none focus:outline-none focus:border-indigo-500 transition-colors mb-4"
      />

      <button
        onClick={handleSubmit}
        disabled={loading || text.trim().length < 3}
        className="w-full py-3 rounded-2xl bg-indigo-500 text-white font-semibold hover:bg-indigo-600 disabled:opacity-40 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2"
      >
        {loading ? (
          <>
            <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            AI is analysing…
          </>
        ) : mode === "grammar" ? "Check Grammar" : "Suggest Better Vocab"}
      </button>

      {error && (
        <div className="mt-6 p-4 rounded-2xl bg-red-500/10 border border-red-500/20 text-red-300 text-sm">
          {error}
        </div>
      )}

      {/* ── Grammar Result ── */}
      {grammarResult && (
        <div className="mt-8 space-y-6">
          {/* Score */}
          <div className="flex items-center gap-4 p-6 rounded-3xl bg-white/5 border border-white/10">
            <div className={`text-6xl font-black ${scoreColor(grammarResult.score)}`}>
              {grammarResult.score}
            </div>
            <div>
              <p className="text-white font-semibold">Grammar Score</p>
              <p className="text-white/40 text-sm">out of 100</p>
            </div>
          </div>

          {/* Corrected text */}
          {grammarResult.correctedText !== text && (
            <div className="p-5 rounded-2xl bg-emerald-500/10 border border-emerald-500/20">
              <p className="text-emerald-400 text-xs font-semibold uppercase tracking-wider mb-2">Corrected version</p>
              <p className="text-white/80 text-sm leading-relaxed">{grammarResult.correctedText}</p>
            </div>
          )}

          {/* Mistakes */}
          {grammarResult.mistakes.length > 0 && (
            <div className="space-y-3">
              <p className="text-white/60 text-xs font-semibold uppercase tracking-wider">
                Mistakes ({grammarResult.mistakes.length})
              </p>
              {grammarResult.mistakes.map((m, i) => (
                <div key={i} className="p-4 rounded-2xl bg-white/5 border border-white/10">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-red-400 line-through text-sm">{m.wrong}</span>
                    <span className="text-white/40">→</span>
                    <span className="text-emerald-400 text-sm font-medium">{m.correct}</span>
                    <span className="ml-auto text-xs px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300">{m.reason}</span>
                  </div>
                  <p className="text-white/50 text-xs">{m.explanation}</p>
                </div>
              ))}
            </div>
          )}

          {/* Tips */}
          {grammarResult.tips.length > 0 && (
            <div className="p-5 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 space-y-2">
              <p className="text-indigo-300 text-xs font-semibold uppercase tracking-wider mb-3">Tips</p>
              {grammarResult.tips.map((t, i) => (
                <p key={i} className="text-white/60 text-sm flex gap-2">
                  <span className="text-indigo-400">💡</span>{t}
                </p>
              ))}
            </div>
          )}
        </div>
      )}

      {/* ── Vocab Result ── */}
      {vocabResult && (
        <div className="mt-8 space-y-6">
          <div className="p-5 rounded-2xl bg-indigo-500/10 border border-indigo-500/20">
            <p className="text-indigo-300 text-xs font-semibold uppercase tracking-wider mb-2">Rewritten version</p>
            <p className="text-white/80 text-sm leading-relaxed">{vocabResult.rewrittenText}</p>
          </div>

          {vocabResult.suggestions.length > 0 && (
            <div className="space-y-3">
              <p className="text-white/60 text-xs font-semibold uppercase tracking-wider">
                Suggestions ({vocabResult.suggestions.length})
              </p>
              {vocabResult.suggestions.map((s, i) => (
                <div key={i} className="p-4 rounded-2xl bg-white/5 border border-white/10">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-white/50 text-sm">{s.original}</span>
                    <span className="text-white/30">→</span>
                    <span className="text-amber-300 text-sm font-semibold">{s.suggested}</span>
                  </div>
                  <p className="text-white/50 text-xs mb-1">{s.reason}</p>
                  <p className="text-white/30 text-xs italic">"{s.example}"</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
