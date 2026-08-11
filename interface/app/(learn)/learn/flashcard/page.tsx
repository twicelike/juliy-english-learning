"use client";

import { useEffect, useState, useCallback, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { LearningAPI, SessionItemDTO } from "@/lib/learning-api";
import { useAuthStore } from "@/store/auth-store";

function FlashcardContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const type = searchParams.get("type");
  const sessionId = searchParams.get("sessionId");
  const { user } = useAuthStore();
  const [cards, setCards] = useState<SessionItemDTO[]>([]);
  const [index, setIndex] = useState(0);
  const [flipped, setFlipped] = useState(false);
  const [loading, setLoading] = useState(true);
  const [done, setDone] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!user) { router.push("/login"); return; }
    
    const loadSession = async () => {
      try {
        if (type === "new") {
          const s = await LearningAPI.createSession("new_words");
          router.replace(`/learn/flashcard?sessionId=${s.id}`);
          return;
        } else if (type === "review") {
          const s = await LearningAPI.createSession("review");
          router.replace(`/learn/flashcard?sessionId=${s.id}`);
          return;
        } else if (sessionId) {
          const s = await LearningAPI.getSessionById(sessionId);
          if (s.items.length === 0) setDone(true);
          else {
            setCards(s.items);
            setIndex(s.current_index || 0);
          }
        } else {
          const s = await LearningAPI.getSession("flashcard", 10);
          if (s.items.length === 0) setDone(true);
          else setCards(s.items);
        }
      } catch (e: any) {
        console.error("Failed to load flashcard session:", e);
        if (e.response?.data?.error?.code === "NO_NEW_WORDS" || e.response?.data?.error?.code === "NO_DUE_WORDS") {
          setDone(true);
        }
      } finally {
        setLoading(false);
      }
    };

    loadSession();
  }, [user, router, type, sessionId]);

  const handleResult = useCallback(async (quality: 1 | 3 | 4 | 5) => {
    if (submitting || !cards[index]) return;
    setSubmitting(true);
    try {
      await LearningAPI.submitProgress(cards[index].word_id, quality, undefined, sessionId || undefined);
    } catch (e) { console.error(e); }

    setSubmitting(false);
    setFlipped(false);
    if (index + 1 >= cards.length) setDone(true);
    else setIndex((i) => i + 1);
  }, [submitting, cards, index, sessionId]);

  const handlePrevious = useCallback(() => {
    if (index === 0) return;
    setFlipped(false);
    setIndex((i) => i - 1);
  }, [index]);

  if (loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (done) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-6 text-center px-6">
        <div className="text-6xl">🎉</div>
        <h1 className="text-3xl font-bold text-white">
          {type === "review" ? "You're all caught up for today!" : "Session Complete!"}
        </h1>
        <p className="text-white/50">
          {type === "review" 
            ? "There are no more words due for review right now." 
            : "You've reviewed all your cards for this session."}
        </p>
        <button onClick={() => router.push("/learn")} className="px-6 py-3 rounded-xl bg-indigo-500 text-white font-medium hover:bg-indigo-600">
          Back to Dashboard
        </button>
      </div>
    );
  }

  const card = cards[index];
  if (!card) return null;

  return (
    <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center px-6 py-12">
      {/* Progress */}
      <div className="w-full max-w-md mb-6">
        <div className="flex justify-between text-sm text-white/40 mb-2">
          <span>{index + 1} / {cards.length}</span>
          <span>{Math.round((index / cards.length) * 100)}% done</span>
        </div>
        <div className="h-1.5 w-full bg-white/10 rounded-full overflow-hidden">
          <div
            className="h-full bg-indigo-500 rounded-full transition-all duration-300"
            style={{ width: `${(index / cards.length) * 100}%` }}
          />
        </div>
      </div>

      {/* Flashcard with 3D flip */}
      <div
        className="w-full max-w-md h-72 cursor-pointer"
        style={{ perspective: "1200px" }}
        onClick={() => !flipped && setFlipped(true)}
      >
        <div
          className="relative w-full h-full transition-transform duration-500"
          style={{
            transformStyle: "preserve-3d",
            transform: flipped ? "rotateY(180deg)" : "rotateY(0deg)",
          }}
        >
          {/* Front */}
          <div
            className="absolute inset-0 rounded-3xl bg-white/5 border border-white/10 flex flex-col items-center justify-center p-8 gap-3"
            style={{ backfaceVisibility: "hidden" }}
          >
            <span className="px-3 py-1 rounded-full bg-indigo-500/20 text-indigo-300 text-xs font-medium">{card.pos}</span>
            <h2 className="text-5xl font-bold text-white">{card.word}</h2>
            {card.ipa && <p className="text-white/40 font-mono">{card.ipa}</p>}
            <p className="text-white/30 text-sm mt-4">Tap to reveal definition</p>
          </div>

          {/* Back */}
          <div
            className="absolute inset-0 rounded-3xl bg-indigo-900/40 border border-indigo-500/30 flex flex-col items-center justify-center p-8 gap-4"
            style={{ backfaceVisibility: "hidden", transform: "rotateY(180deg)" }}
          >
            <ul className="list-disc list-inside space-y-1.5 text-center">
              {card.definitions.slice(0, 3).map((d, i) => (
                <li key={i} className="text-white/80 text-sm">{d}</li>
              ))}
            </ul>
            {card.example && (
              <p className="text-white/40 text-xs italic text-center">"{card.example}"</p>
            )}
          </div>
        </div>
      </div>

      {/* Navigation & Action buttons */}
      <div className="mt-8 w-full max-w-md space-y-3">
        {/* Pass / Fail — only after flip */}
        {flipped && (
          <div className="flex gap-2">
            <button
              onClick={() => handleResult(1)}
              disabled={submitting}
              className="flex-1 py-3 rounded-xl bg-red-500/15 border border-red-500/30 text-red-400 font-semibold hover:bg-red-500/25 transition-colors disabled:opacity-50"
            >
              <div className="text-sm opacity-60 font-normal mb-1">Again</div>
              <div><span className="opacity-50 text-sm">≤10m</span></div>
            </button>
            <button
              onClick={() => handleResult(3)}
              disabled={submitting}
              className="flex-1 py-3 rounded-xl bg-orange-500/15 border border-orange-500/30 text-orange-400 font-semibold hover:bg-orange-500/25 transition-colors disabled:opacity-50"
            >
              <div className="text-sm opacity-60 font-normal mb-1">Hard</div>
              <div><span className="opacity-50 text-sm">1d</span></div>
            </button>
            <button
              onClick={() => handleResult(4)}
              disabled={submitting}
              className="flex-1 py-3 rounded-xl bg-blue-500/15 border border-blue-500/30 text-blue-400 font-semibold hover:bg-blue-500/25 transition-colors disabled:opacity-50"
            >
              <div className="text-sm opacity-60 font-normal mb-1">Good</div>
              <div><span className="opacity-50 text-sm">3d</span></div>
            </button>
            <button
              onClick={() => handleResult(5)}
              disabled={submitting}
              className="flex-1 py-3 rounded-xl bg-green-500/15 border border-green-500/30 text-green-400 font-semibold hover:bg-green-500/25 transition-colors disabled:opacity-50"
            >
              <div className="text-sm opacity-60 font-normal mb-1">Easy</div>
              <div><span className="opacity-50 text-sm">5d+</span></div>
            </button>
          </div>
        )}

        {/* Previous — always visible, disabled on first card */}
        <div className="flex justify-center">
          <button
            onClick={handlePrevious}
            disabled={index === 0}
            className="flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm text-white/40 hover:text-white/70 disabled:opacity-20 disabled:cursor-not-allowed transition-colors"
          >
            ← Previous card
          </button>
        </div>
      </div>
    </div>
  );
}

export default function FlashcardPage() {
  return (
    <Suspense fallback={
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
      </div>
    }>
      <FlashcardContent />
    </Suspense>
  );
}
