"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";
import { AIAPI, FillBlankResponse } from "@/lib/ai-api";
import { LearningAPI } from "@/lib/learning-api";

export default function FillBlankPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [quiz, setQuiz] = useState<FillBlankResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [answer, setAnswer] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [isCorrect, setIsCorrect] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [score, setScore] = useState({ correct: 0, total: 0 });

  const loadQuiz = useCallback(async () => {
    setLoading(true);
    setError(null);
    setQuiz(null);
    setAnswer("");
    setSubmitted(false);

    try {
      // Get due/new words from the learning session
      const session = await LearningAPI.getSession("flashcard", 10);
      if (session.items.length === 0) {
        setError("No bookmarked words found. Bookmark some words first!");
        setLoading(false);
        return;
      }

      const words = session.items.map((item) => ({ word: item.word, pos: item.pos }));
      const result = await AIAPI.generateFillBlank({ words });
      setQuiz(result);
    } catch (e: any) {
      setError(e?.message ?? "Failed to generate quiz.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!user) { router.push("/login"); return; }
    loadQuiz();
  }, [user, router, loadQuiz]);

  const handleSubmit = () => {
    if (!quiz || !answer.trim()) return;
    const correct = answer.trim().toLowerCase() === quiz.answer.toLowerCase();
    setIsCorrect(correct);
    setSubmitted(true);
    setScore((s) => ({ correct: s.correct + (correct ? 1 : 0), total: s.total + 1 }));
  };

  const handleNext = () => loadQuiz();

  // ── Render ────────────────────────────────────────────────────────────────

  if (loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-4">
        <div className="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-white/40 text-sm">AI is generating your quiz…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-4 text-center px-6">
        <div className="text-5xl">🤖</div>
        <p className="text-white/60 text-sm">{error}</p>
        <button onClick={loadQuiz} className="px-6 py-2 rounded-xl bg-indigo-500 text-white text-sm hover:bg-indigo-600">
          Try Again
        </button>
        <button onClick={() => router.push("/learn")} className="text-white/30 text-sm hover:text-white/60 transition-colors">
          Back to Dashboard
        </button>
      </div>
    );
  }

  if (!quiz) return null;

  // Build display sentence — split on "___"
  const parts = quiz.sentence.split("___");

  return (
    <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-start px-6 py-12">
      <div className="w-full max-w-xl">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-white">Fill in the Blank</h1>
            <p className="text-white/40 text-sm">AI-generated from your bookmarks</p>
          </div>
          {score.total > 0 && (
            <div className="text-right">
              <p className="text-white font-bold">{score.correct}/{score.total}</p>
              <p className="text-white/30 text-xs">Score</p>
            </div>
          )}
        </div>

        {/* Quiz card */}
        <div className="rounded-3xl bg-white/5 border border-white/10 p-8 mb-6">
          {/* POS badge */}
          {quiz.pos && (
            <span className="inline-block mb-4 px-3 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-xs">
              {quiz.pos}
            </span>
          )}

          {/* Sentence with blank */}
          <p className="text-white text-xl leading-relaxed mb-6">
            {parts[0]}
            <span className={`inline-block border-b-2 mx-2 min-w-[80px] text-center align-bottom transition-colors ${
              submitted
                ? isCorrect ? "border-emerald-400 text-emerald-400" : "border-red-400 text-red-400"
                : "border-indigo-400 text-indigo-300"
            }`}>
              {submitted ? quiz.answer : answer || "\u00A0\u00A0\u00A0\u00A0\u00A0"}
            </span>
            {parts[1]}
          </p>

          {/* Hint */}
          {quiz.hint && !submitted && (
            <p className="text-white/30 text-xs">💡 Hint: {quiz.hint}</p>
          )}

          {/* Answer after submission */}
          {submitted && !isCorrect && (
            <p className="text-white/50 text-sm mt-2">
              Your answer: <span className="text-red-400">{answer}</span>
            </p>
          )}
        </div>

        {/* Input */}
        {!submitted ? (
          <div className="flex gap-3">
            <input
              autoFocus
              value={answer}
              onChange={(e) => setAnswer(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
              placeholder="Type your answer…"
              className="flex-1 px-4 py-3 rounded-2xl bg-white/5 border border-white/10 text-white placeholder-white/20 focus:outline-none focus:border-indigo-500 transition-colors text-sm"
            />
            <button
              onClick={handleSubmit}
              disabled={!answer.trim()}
              className="px-6 py-3 rounded-2xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 disabled:opacity-40 disabled:cursor-not-allowed transition-all"
            >
              Submit
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {/* Result banner */}
            <div className={`p-4 rounded-2xl text-center font-semibold ${
              isCorrect
                ? "bg-emerald-500/20 border border-emerald-500/30 text-emerald-400"
                : "bg-red-500/20 border border-red-500/30 text-red-400"
            }`}>
              {isCorrect ? "🎉 Correct!" : `❌ Incorrect — answer was "${quiz.answer}"`}
            </div>
            <button
              onClick={handleNext}
              className="w-full py-3 rounded-2xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-all"
            >
              Next Question →
            </button>
            <button
              onClick={() => router.push("/learn")}
              className="w-full py-2 rounded-2xl text-white/30 text-sm hover:text-white/60 transition-colors"
            >
              Back to Dashboard
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
