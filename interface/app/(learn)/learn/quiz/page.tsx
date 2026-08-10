"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { LearningAPI, SessionItemDTO, ChoiceDTO } from "@/lib/learning-api";
import { useAuthStore } from "@/store/auth-store";

export default function QuizPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [cards, setCards] = useState<SessionItemDTO[]>([]);
  const [index, setIndex] = useState(0);
  const [selected, setSelected] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [done, setDone] = useState(false);
  const [score, setScore] = useState(0);

  useEffect(() => {
    if (!user) { router.push("/login"); return; }
    LearningAPI.getSession("quiz", 10)
      .then((s) => {
        if (s.items.length === 0) setDone(true);
        else setCards(s.items);
      })
      .catch((e) => {
        console.error("Failed to load quiz session:", e);
      })
      .finally(() => setLoading(false));
  }, [user, router]);

  const handleSelect = useCallback(async (choiceIndex: number) => {
    if (selected !== null) return; // Already answered
    setSelected(choiceIndex);

    const card = cards[index];
    const choice = card.choices![choiceIndex];

    if (choice.is_correct) setScore((s) => s + 1);

    try {
      await LearningAPI.submitProgress(card.word_id, choice.is_correct ? 4 : 1);
    } catch (e) { console.error(e); }
  }, [selected, cards, index]);

  const handleNext = () => {
    setSelected(null);
    if (index + 1 >= cards.length) setDone(true);
    else setIndex((i) => i + 1);
  };

  if (loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (done) {
    const pct = cards.length > 0 ? Math.round((score / cards.length) * 100) : 0;
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-6 text-center px-6">
        <div className="text-6xl">{pct >= 80 ? "🏆" : pct >= 50 ? "👍" : "📚"}</div>
        <h1 className="text-3xl font-bold text-white">Quiz Complete!</h1>
        <p className="text-white/60">You scored <span className="text-white font-bold">{score} / {cards.length}</span> ({pct}%)</p>
        <button onClick={() => router.push("/learn")} className="px-6 py-3 rounded-xl bg-indigo-500 text-white font-medium hover:bg-indigo-600">
          Back to Dashboard
        </button>
      </div>
    );
  }

  const card = cards[index];
  if (!card) return null;
  const choices: ChoiceDTO[] = card.choices ?? [];

  return (
    <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-start px-6 py-12">
      {/* Progress */}
      <div className="w-full max-w-lg mb-8">
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

      {/* Question */}
      <div className="w-full max-w-lg rounded-3xl bg-white/5 border border-white/10 p-8 mb-6 text-center">
        <p className="text-sm text-white/40 mb-3">What does this word mean?</p>
        <div className="flex items-center justify-center gap-3 flex-wrap">
          <span className="text-4xl font-bold text-white">{card.word}</span>
          {card.pos && (
            <span className="px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-xs">{card.pos}</span>
          )}
        </div>
        {card.ipa && <p className="text-white/30 font-mono mt-2 text-sm">{card.ipa}</p>}
      </div>

      {/* Choices */}
      <div className="w-full max-w-lg space-y-3">
        {choices.map((choice, i) => {
          let cls = "w-full p-4 rounded-2xl border text-left text-sm leading-relaxed transition-all ";
          if (selected === null) {
            cls += "bg-white/5 border-white/10 text-white/80 hover:bg-white/10 cursor-pointer";
          } else if (choice.is_correct) {
            cls += "bg-emerald-500/20 border-emerald-500/40 text-emerald-300";
          } else if (selected === i) {
            cls += "bg-red-500/20 border-red-500/40 text-red-300";
          } else {
            cls += "bg-white/5 border-white/10 text-white/30";
          }
          return (
            <button key={i} className={cls} onClick={() => handleSelect(i)}>
              <span className="font-mono text-white/40 mr-3">{String.fromCharCode(65 + i)}.</span>
              {choice.definition}
            </button>
          );
        })}
      </div>

      {/* Next button */}
      {selected !== null && (
        <button
          onClick={handleNext}
          className="mt-8 px-8 py-3 rounded-2xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-colors"
        >
          {index + 1 >= cards.length ? "See Results" : "Next →"}
        </button>
      )}
    </div>
  );
}
