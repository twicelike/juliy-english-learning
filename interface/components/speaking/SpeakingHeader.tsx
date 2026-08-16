import React from "react";
import Link from "next/link";
import { ChevronRight } from "lucide-react";

interface SpeakingHeaderProps {
  streak: number;
}

export function SpeakingHeader({ streak }: SpeakingHeaderProps) {
  return (
    <div className="flex justify-between items-start mb-8">
      <div>
        <div className="flex items-center gap-4 mb-2">
          <h1 className="text-4xl font-bold text-white">Speaking</h1>
          <Link 
            href="/speaking/history" 
            className="group flex items-center gap-1 text-sm font-medium text-white/50 hover:text-white/90 transition-colors bg-white/5 hover:bg-white/10 px-3 py-1.5 rounded-full"
          >
            History <ChevronRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
          </Link>
        </div>
        <p className="text-white/60 text-lg">
          Practice speaking naturally. <br className="hidden md:block" />
          Build confidence, one conversation at a time.
        </p>
      </div>
      <div className="flex items-center gap-4 text-right">
        {streak > 0 && (
          <div className="text-right">
            <div className="text-2xl font-bold text-orange-400">🔥 {streak}</div>
            <div className="text-xs text-white/40 uppercase tracking-widest">day streak</div>
          </div>
        )}
      </div>
    </div>
  );
}
