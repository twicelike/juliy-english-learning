import { WordDTO } from "@/lib/dictionary-api";
import { useEffect, useState } from "react";
import AudioButton from "@/components/dictionary/AudioButton";

interface Props {
  word: WordDTO;
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

  return (
    <div className="flex flex-col h-full bg-[#1e1e2d] rounded-3xl border border-white/10 shadow-2xl overflow-hidden relative">
      {/* ─── Flashcard Area ───────────────────────────────────── */}
      <div 
        className={`flex-1 overflow-y-auto custom-scrollbar p-10 flex flex-col transition-all cursor-pointer ${
          !isRevealed ? "justify-center items-center hover:bg-white/[0.02]" : ""
        }`}
        onClick={() => !isRevealed && setIsRevealed(true)}
      >
        {!isRevealed ? (
          <div className="text-center">
            <h2 className="text-5xl md:text-7xl font-bold text-white mb-6 tracking-tight">
              {word.word}
            </h2>
            <div className="text-white/30 text-sm tracking-widest uppercase flex items-center justify-center gap-2">
              <span>Press Space to reveal</span>
              <kbd className="px-2 py-1 rounded bg-white/10 text-white/50 font-mono text-xs hidden sm:inline-block">␣</kbd>
            </div>
          </div>
        ) : (
          <div className="w-full max-w-2xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-300">
            <div>
              <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
                {word.word}
              </h2>
              <div className="flex items-center gap-4 flex-wrap">
                {word.pos && word.pos.split(", ").map((p, i) => (
                  <span key={i} className="px-3 py-1 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 text-sm font-medium uppercase tracking-wider">
                    {p}
                  </span>
                ))}
                {word.ipa && (
                  <span className="text-white/60 font-mono text-lg flex items-center gap-2">
                    {word.ipa}
                    {word.audio && <AudioButton audioUrl={word.audio} />}
                  </span>
                )}
                {!word.ipa && word.audio && <AudioButton audioUrl={word.audio} />}
              </div>
            </div>

            <div className="h-px bg-white/10 w-full" />

            <div className="space-y-6">
              {word.definitions.slice(0, 3).map((def, i) => (
                <div key={i} className="space-y-2">
                  <p className="text-lg text-white/90 leading-relaxed">
                    <span className="text-indigo-400 font-bold mr-2">{i + 1}.</span>
                    {def.definition}
                  </p>
                  {def.examples && def.examples.length > 0 && (
                    <div className="space-y-1 mt-1">
                      {def.examples.map((example, idx) => (
                        <p key={idx} className="text-white/50 text-base italic pl-6 border-l-2 border-white/10">
                          "{example}"
                        </p>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {word.synonyms && word.synonyms.length > 0 && (
              <div className="pt-4 space-y-2">
                <h4 className="text-sm font-bold text-white/40 uppercase tracking-widest">Synonyms</h4>
                <div className="flex flex-wrap gap-2">
                  {word.synonyms.map((syn, i) => (
                    <span key={i} className="px-3 py-1.5 rounded-lg bg-white/5 text-white/70 text-sm border border-white/5">
                      {syn}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* ─── Navigation Bar ───────────────────────────────────── */}
      <div className="h-16 border-t border-white/10 bg-black/20 flex items-center justify-between px-6 shrink-0">
        <button
          onClick={(e) => { e.stopPropagation(); onPrev(); }}
          disabled={!hasPrev}
          className="flex items-center gap-2 text-white/50 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-colors font-medium px-4 py-2 -ml-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" /></svg>
          Previous
        </button>

        <button
          onClick={(e) => { e.stopPropagation(); setIsRevealed(!isRevealed); }}
          className="px-6 py-2 rounded-full bg-white/5 hover:bg-white/10 text-white/70 hover:text-white transition-colors text-sm font-medium border border-white/10"
        >
          {isRevealed ? "Hide Details" : "Reveal Details"}
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
