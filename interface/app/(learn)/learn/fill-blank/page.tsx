"use client";

import { useEffect, useState, useCallback, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { LearningAPI, SessionItemDTO } from "@/lib/learning-api";
import { useAuthStore } from "@/store/auth-store";

function FillBlankContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const type = searchParams.get("type");
  const sessionId = searchParams.get("sessionId");
  const { user } = useAuthStore();
  const [cards, setCards] = useState<SessionItemDTO[]>([]);
  const [index, setIndex] = useState(0);
  const [loading, setLoading] = useState(true);
  const [done, setDone] = useState(false);
  const [answer, setAnswer] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [isCorrect, setIsCorrect] = useState(false);
  const [score, setScore] = useState(0);

  useEffect(() => {
    if (!user) { router.push("/login"); return; }

    const loadSession = async () => {
      try {
        if (type === "new") {
          const s = await LearningAPI.createSession("new_words", "fill_blank");
          router.replace(`/learn/fill-blank?sessionId=${s.id}`);
          return;
        } else if (type === "review") {
          const s = await LearningAPI.createSession("review", "fill_blank");
          router.replace(`/learn/fill-blank?sessionId=${s.id}`);
          return;
        } else if (sessionId) {
          const s = await LearningAPI.getSessionById(sessionId, "fill_blank");
          if (s.items.length === 0) setDone(true);
          else {
            setCards(s.items);
            setIndex(s.current_index || 0);
          }
        } else {
          // fallback
          const s = await LearningAPI.getSession("quiz", 10);
          if (s.items.length === 0) setDone(true);
          else setCards(s.items);
        }
      } catch (e: any) {
        console.error("Failed to load fill-blank session:", e);
        if (e.response?.data?.error?.code === "NO_NEW_WORDS" || e.response?.data?.error?.code === "NO_DUE_WORDS") {
          setDone(true);
        }
      } finally {
        setLoading(false);
      }
    };

    loadSession();
  }, [user, router, type, sessionId]);

  const handleSubmit = useCallback(async () => {
    if (submitted) return;
    const card = cards[index];
    if (!card || !card.fill_blank_answer || !answer.trim()) return;

    const correct = answer.trim().toLowerCase() === card.fill_blank_answer.toLowerCase();
    setIsCorrect(correct);
    setSubmitted(true);
    
    if (correct) setScore((s) => s + 1);

    try {
      await LearningAPI.submitProgress(card.word_id, correct ? 4 : 1, undefined, sessionId || undefined);
    } catch (e) { console.error(e); }
  }, [answer, submitted, cards, index, sessionId]);

  const handleNext = () => {
    setSubmitted(false);
    setAnswer("");
    if (index + 1 >= cards.length) setDone(true);
    else setIndex((i) => i + 1);
  };

  if (loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-4">
        <div className="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-white/40 text-sm">Loading your exercises…</p>
      </div>
    );
  }

  if (done) {
    const pct = cards.length > 0 ? Math.round((score / cards.length) * 100) : 0;
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-6 text-center px-6">
        <div className="text-6xl">{pct >= 80 ? "🏆" : pct >= 50 ? "👍" : "📚"}</div>
        <h1 className="text-3xl font-bold text-white">Exercise Complete!</h1>
        <p className="text-white/60">You scored <span className="text-white font-bold">{score} / {cards.length}</span> ({pct}%)</p>
        <button onClick={() => router.push("/learn")} className="px-6 py-3 rounded-xl bg-indigo-500 text-white font-medium hover:bg-indigo-600">
          Back to Dashboard
        </button>
      </div>
    );
  }

  const card = cards[index];
  if (!card || !card.fill_blank_question) return null;

  const parts = card.fill_blank_question.split("___");
  if (parts.length < 2) return null; // fallback if invalid

  return (
    <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-start px-6 py-12">
      {/* Progress */}
      <div className="w-full max-w-xl mb-8">
        <div className="flex justify-between text-sm text-white/40 mb-2">
          <span>{index + 1} / {cards.length}</span>
          <span>Score: {score}</span>
        </div>
        <div className="h-1.5 w-full bg-white/10 rounded-full overflow-hidden">
          <div
            className="h-full bg-indigo-500 rounded-full transition-all duration-300"
            style={{ width: `${(index / cards.length) * 100}%` }}
          />
        </div>
      </div>

      <div className="w-full max-w-xl">
        {/* Quiz card */}
        <div className="rounded-3xl bg-white/5 border border-white/10 p-8 mb-6">
          {/* POS badge */}
          {card.pos && (
            <span className="inline-block mb-4 px-3 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-xs">
              {card.pos}
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
              {submitted ? card.fill_blank_answer : answer || "\u00A0\u00A0\u00A0\u00A0\u00A0"}
            </span>
            {parts.slice(1).join("___")}
          </p>

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
              {isCorrect ? "🎉 Correct!" : `❌ Incorrect — answer was "${card.fill_blank_answer}"`}
            </div>
            <button
              onClick={handleNext}
              className="w-full py-3 rounded-2xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-all"
            >
              {index + 1 >= cards.length ? "See Results" : "Next Question →"}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default function FillBlankPage() {
  return (
    <Suspense fallback={
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-4">
        <div className="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-white/40 text-sm">Loading your exercises…</p>
      </div>
    }>
      <FillBlankContent />
    </Suspense>
  );
}
