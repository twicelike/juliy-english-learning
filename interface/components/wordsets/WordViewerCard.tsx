import { WordDTO } from "@/lib/dictionary-api";
import { useEffect, useState } from "react";
import AudioButton from "@/components/dictionary/AudioButton";

interface WordViewerItemDTO extends WordDTO {
  progress_state?: string;
  sense_index?: number;
  translation_candidates?: string[];
  annotation?: {
    text: string;
    source: string;
    language_code: string;
  };
}

interface Props {
  word: WordViewerItemDTO;
  onNext: () => void;
  onPrev: () => void;
  hasNext: boolean;
  hasPrev: boolean;
  isActive: boolean;
}

export function WordViewerCard({ word, onNext, onPrev, hasNext, hasPrev, isActive }: Props) {
  const [isRevealed, setIsRevealed] = useState(false);

  // Reset reveal state when word changes
  useEffect(() => {
    setIsRevealed(false);
  }, [word.id]);

  useEffect(() => {
    if (!isActive) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't trigger if user is typing in an input/textarea
      if (
        document.activeElement?.tagName === "INPUT" ||
        document.activeElement?.tagName === "TEXTAREA" ||
        document.activeElement?.tagName === "BUTTON"
      ) {
        return;
      }

      if (e.code === "Space") {
        e.preventDefault();
        setIsRevealed((r) => !r);
      } else if (e.key === "ArrowLeft") {
        if (hasPrev) onPrev();
      } else if (e.key === "ArrowRight") {
        if (hasNext) onNext();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isActive, hasNext, hasPrev, onNext, onPrev]);

  // Determine translation to display on front
  let activeTranslation = "";
  if (word.annotation) {
    activeTranslation = word.annotation.text;
  } else if (word.translation_candidates && word.translation_candidates.length > 0) {
    activeTranslation = word.translation_candidates[0];
  }

  // Filter definitions for specific sense if applicable
  const targetSenseIndex = (word as any).sense_index;
  const filteredDefs = word.definitions
    ? word.definitions.filter((def: any) => targetSenseIndex === undefined || def.sense_index === targetSenseIndex)
    : [];

  return (
    <div className="flex flex-col h-full bg-transparent rounded-3xl overflow-hidden relative" style={{ perspective: "1200px" }}>
      
      {/* ─── Flashcard Area (3D Flip) ───────────────────────────────────── */}
      <div 
        className="flex-1 w-full h-full cursor-pointer relative"
        onClick={() => setIsRevealed(prev => !prev)}
      >
        <div
          className="relative w-full h-full transition-transform duration-600 shadow-2xl rounded-3xl"
          style={{
            transformStyle: "preserve-3d",
            transform: isRevealed ? "rotateY(180deg)" : "rotateY(0deg)",
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
              {word.pos && (
                <span className="mb-4 px-3 py-1 rounded-full bg-white/5 text-white/50 text-xs font-medium tracking-widest uppercase border border-white/10">
                  {word.pos.split(", ")[0]}
                </span>
              )}

              <h2 className="text-5xl sm:text-7xl font-bold text-white mb-4 tracking-tight">
                {word.word}
              </h2>

              <div className="flex items-center gap-4">
                {word.ipa && <span className="text-white/40 font-mono text-lg">{word.ipa}</span>}
                {word.audio && <AudioButton audioUrl={word.audio} />}
              </div>

            </div>

            <div className="absolute bottom-16 sm:bottom-6 text-white/30 text-xs uppercase tracking-widest flex items-center gap-2">
              <span>Press Space to reveal</span>
              <kbd className="px-2 py-1 rounded bg-white/10 text-white/50 font-mono text-[10px] hidden sm:inline-block">␣</kbd>
            </div>
          </div>

          {/* BACK */}
          <div
            className="absolute inset-0 rounded-3xl overflow-hidden flex flex-col border border-indigo-500/30 shadow-2xl bg-black/60"
            style={{
              backfaceVisibility: "hidden",
              WebkitBackfaceVisibility: "hidden",
              transform: "rotateY(180deg)",
              background: "linear-gradient(135deg, rgba(79, 70, 229, 0.15) 0%, rgba(17, 24, 39, 0.9) 100%)",
            }}
          >
            <div className="flex-1 overflow-y-auto custom-scrollbar p-6 pb-20 sm:p-10 space-y-6">
              
              {/* Header (Word + POS) */}
              <div className="text-center mb-8">
                 <h3 className="text-3xl font-bold text-white/90 mb-2">{word.word}</h3>
                 <div className="flex justify-center gap-2">
                    {word.pos && word.pos.split(", ").map((p, i) => (
                      <span key={i} className="px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 text-[10px] uppercase tracking-wider">
                        {p}
                      </span>
                    ))}
                 </div>
              </div>

              {/* Definitions */}
              <div className="space-y-4">
                <h4 className="text-xs uppercase tracking-widest text-white/40 font-semibold mb-3">Definitions</h4>
                <div className="space-y-4">
                  {filteredDefs.slice(0, 3).map((def: any, i: number) => (
                    <div key={i} className="space-y-2">
                      <p className="text-base sm:text-lg text-white/90 leading-relaxed">
                        <span className="text-indigo-400 font-bold mr-2">{i + 1}.</span>
                        {def.definition}
                      </p>

                    </div>
                  ))}
                </div>
              </div>

              {/* Translation Area */}
              <div className="pt-6 border-t border-white/10 mt-6">
                <h4 className="text-xs uppercase tracking-widest text-white/40 font-semibold mb-3">Translation</h4>
                
                {word.annotation ? (
                  <div className="bg-indigo-500/10 border border-indigo-500/30 rounded-xl p-4 flex items-start gap-3">
                     <div className="text-indigo-400 mt-0.5">⭐</div>
                     <div>
                       <div className="text-white font-medium text-lg">{word.annotation.text}</div>
                       <div className="text-indigo-400/60 text-[10px] uppercase tracking-wider mt-2">Your Annotation</div>
                     </div>
                  </div>
                ) : word.translation_candidates && word.translation_candidates.length > 0 ? (
                  <div className="flex flex-wrap gap-2">
                    <div className="w-full flex items-center gap-2 mb-1 text-white/50 text-xs">
                      <span>AI Candidates</span>
                    </div>
                    {word.translation_candidates.map((t, idx) => (
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

      {/* ─── Navigation Bar ───────────────────────────────────── */}
      <div className="absolute bottom-0 left-0 right-0 h-16 border-t border-white/10 bg-black/40 backdrop-blur-md flex items-center justify-between px-6 z-20 rounded-b-3xl">
        <button
          onClick={(e) => { e.stopPropagation(); onPrev(); }}
          disabled={!hasPrev}
          className="flex items-center gap-2 text-white/50 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-colors font-medium px-4 py-2 -ml-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" /></svg>
          Previous
        </button>



        <button
          onClick={(e) => { e.stopPropagation(); onNext(); }}
          disabled={!hasNext}
          className="flex items-center gap-2 text-white/50 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-colors font-medium px-4 py-2 -mr-4"
        >
          Next
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
        </button>
      </div>
    </div>
  );
}
