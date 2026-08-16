import React from "react";
import { SessionItemDTO } from "@/lib/learning-api";
import AudioButton from "@/components/dictionary/AudioButton";

interface FlashcardProps {
  card: SessionItemDTO;
  flipped: boolean;
  onFlip: () => void;
}

export function Flashcard({ card, flipped, onFlip }: FlashcardProps) {
  // Determine translation to display on front
  let activeTranslation = "";
  if (card.annotation) {
    activeTranslation = card.annotation.translation;
  } else if (card.translation_candidates && card.translation_candidates.length > 0) {
    // Only show the first translation on front if no annotation
    activeTranslation = card.translation_candidates[0];
  }

  return (
    <div
      className="w-full h-80 sm:h-96 cursor-pointer"
      style={{ perspective: "1200px" }}
      onClick={() => {
        onFlip();
      }}
    >
      <div
        className="relative w-full h-full transition-transform duration-600"
        style={{
          transformStyle: "preserve-3d",
          transform: flipped ? "rotateY(180deg)" : "rotateY(0deg)",
        }}
      >
        {/* FRONT */}
        <div
          className="absolute inset-0 rounded-3xl overflow-hidden flex flex-col items-center justify-center p-8 gap-4 border border-white/10 shadow-2xl bg-black/40"
          style={{
            backfaceVisibility: "hidden",
            WebkitBackfaceVisibility: "hidden",
            background: "linear-gradient(135deg, rgba(255,255,255,0.08) 0%, rgba(255,255,255,0.02) 100%)",
          }}
        >
          <div className="z-10 flex flex-col items-center">
            {/* POS */}
            {card.pos && (
              <span className="mb-2 px-3 py-1 rounded-full bg-white/5 text-white/50 text-xs font-medium tracking-widest uppercase border border-white/10">
                {card.pos}
              </span>
            )}

            {/* Word */}
            <h2 className="text-5xl sm:text-6xl font-bold text-white mb-2 tracking-tight">
              {card.word}
            </h2>

            {/* IPA & Audio */}
            <div className="flex items-center gap-3">
              {card.ipa && <span className="text-white/40 font-mono text-lg">{card.ipa}</span>}
              <AudioButton audioUrl={`https://api.dictionaryapi.dev/media/pronunciations/en/${card.word}-us.mp3`} />
            </div>

          </div>

          <p className="absolute bottom-6 text-white/30 text-xs uppercase tracking-widest flex items-center gap-2">
            Tap to reveal
          </p>
        </div>

        {/* BACK */}
        <div
          className="absolute inset-0 rounded-3xl overflow-hidden flex flex-col p-6 sm:p-8 border border-indigo-500/30 shadow-2xl custom-scrollbar overflow-y-auto bg-black/60"
          style={{
            backfaceVisibility: "hidden",
            WebkitBackfaceVisibility: "hidden",
            transform: "rotateY(180deg)",
            background: "linear-gradient(135deg, rgba(79, 70, 229, 0.15) 0%, rgba(17, 24, 39, 0.9) 100%)",
          }}
        >
          {/* Word (Small) & POS */}
          <div className="text-center mb-2 flex flex-col items-center">
            <h3 className="text-3xl font-bold text-white tracking-tight">{card.word}</h3>
            {card.pos && (
              <span className="mt-2 px-3 py-1 rounded-full bg-indigo-500/20 text-indigo-300 text-[10px] font-semibold tracking-widest uppercase border border-indigo-500/30">
                {card.pos}
              </span>
            )}
          </div>

          <div className="flex-1 flex flex-col justify-center space-y-8">
            {/* Definitions */}
            <div className="space-y-3">
              <h4 className="text-xs uppercase tracking-widest text-white/40 font-semibold mb-2">Definitions</h4>
              <ul className="space-y-4 mt-2">
                {card.definitions.slice(0, 3).map((d, i) => (
                  <li key={i} className="flex gap-4 text-white/90 text-base sm:text-lg leading-relaxed bg-white/5 p-4 rounded-xl border border-white/5">
                    <span className="text-indigo-400 font-bold">{i + 1}.</span>
                    <span>{d}</span>
                  </li>
                ))}
              </ul>
            </div>


            {/* Translation & Annotations */}
            <div className="pt-4 border-t border-white/10">
              <h4 className="text-xs uppercase tracking-widest text-white/40 font-semibold mb-3">Translation</h4>

              {card.annotation ? (
                <div className="bg-indigo-500/10 border border-indigo-500/30 rounded-xl p-3 flex items-start gap-3">
                  <div className="text-indigo-400 mt-0.5">⭐</div>
                  <div>
                    <div className="text-white font-medium">{card.annotation.translation}</div>
                    {card.annotation.note && (
                      <div className="text-white/50 text-xs mt-1">{card.annotation.note}</div>
                    )}
                    <div className="text-indigo-400/60 text-[10px] uppercase tracking-wider mt-2">Your Annotation</div>
                  </div>
                </div>
              ) : card.translation_candidates && card.translation_candidates.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  <div className="w-full flex items-center gap-2 mb-1 text-white/50 text-xs">
                    <span>Translations</span>
                  </div>
                  {card.translation_candidates.map((t, idx) => (
                    <div key={idx} className="px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-white/80 text-sm">
                      {t}
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-white/30 text-sm italic">No translations available.</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
