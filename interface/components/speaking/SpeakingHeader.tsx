import React from "react";
import Link from "next/link";
import { ChevronRight, Flame } from "lucide-react";

export function SpeakingHeader() {
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
      </div>
    </div>
  );
}
