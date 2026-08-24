import React from "react";

interface SessionProgressProps {
  currentIndex: number;
  totalItems: number;
  onSuspend: () => void;
  phase: string;
}

export function SessionProgress({ currentIndex, totalItems, onSuspend, phase }: SessionProgressProps) {
  const progressPct = Math.round((currentIndex / totalItems) * 100) || 0;

  return (
    <div className="w-full max-w-2xl mb-8 space-y-4">
      {/* Header Row */}
      <div className="flex items-center justify-between relative">
        {/* Phase Indicators */}
        <div className="flex gap-2">
          <span className={`px-3 py-1 rounded-full text-xs font-bold ${phase === "flashcard" ? "bg-stone-800 text-white" : "bg-white/10 text-white/40"}`}>
            1. Recall
          </span>
          <span className={`px-3 py-1 rounded-full text-xs font-bold ${phase === "mcq" ? "bg-stone-800 text-white" : "bg-white/10 text-white/40"}`}>
            2. Quiz
          </span>
          <span className={`px-3 py-1 rounded-full text-xs font-bold ${phase === "fill_blank" ? "bg-stone-800 text-white" : "bg-white/10 text-white/40"}`}>
            3. Context
          </span>
        </div>
        
        {/* Suspend Button */}
        <button
          onClick={onSuspend}
          className="px-3 py-1.5 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors border border-red-500/20 text-xs font-medium uppercase tracking-wider"
          title="Suspend this word permanently"
        >
          🚫 Suspend
        </button>
      </div>
      
      {/* Progress Bar */}
      <div className="w-full flex justify-between items-center text-white/50 text-[10px] font-mono uppercase tracking-widest px-1">
        <span>{progressPct}% Completed</span>
        <span>{currentIndex + 1} / {totalItems}</span>
      </div>
      <div className="h-1 w-full bg-white/10 rounded-full overflow-hidden">
        <div
          className="h-full bg-stone-800 rounded-full transition-all duration-500 ease-out"
          style={{ width: `${progressPct}%` }}
        />
      </div>
    </div>
  );
}
