"use client";

import Link from "next/link";
import { WordDTO } from "@/lib/dictionary-api";

interface WordCardProps {
  word: WordDTO;
  title?: string;
  onReload?: () => void;
  isLoading?: boolean;
}

export default function WordCard({ word, title = "Word of the Day", onReload, isLoading }: WordCardProps) {
  const firstDef = word.definitions[0]?.definition || "No definition available.";

  return (
    <div className="group relative p-6 bg-white/5 border border-white/10 rounded-3xl backdrop-blur-md overflow-hidden transition-all hover:bg-white/10 hover:border-indigo-500/50">
      <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/10 rounded-full blur-3xl group-hover:bg-indigo-500/20 transition-all"></div>
      
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-xs text-indigo-400 uppercase tracking-widest font-semibold">
          {title}
        </h3>
        {onReload && (
          <button
            onClick={(e) => {
              e.preventDefault();
              onReload();
            }}
            disabled={isLoading}
            className={`p-1.5 rounded-full bg-white/5 hover:bg-white/10 text-white/50 hover:text-white transition-all ${isLoading ? 'animate-spin opacity-50' : ''}`}
            title="Reload random word"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        )}
      </div>
      
      <div className={`mb-4 transition-opacity ${isLoading ? 'opacity-30' : 'opacity-100'}`}>
        <h4 className="text-3xl font-bold text-white mb-2">{word.word}</h4>
        <div className="flex gap-2 text-sm">
          {word.pos && <span className="text-indigo-300">{word.pos}</span>}
          {word.ipa && <span className="text-white/40">{word.ipa}</span>}
        </div>
      </div>
      
      <p className={`text-white/70 line-clamp-2 mb-6 min-h-[3rem] transition-opacity ${isLoading ? 'opacity-30' : 'opacity-100'}`}>
        {firstDef}
      </p>
      
      <Link 
        href={`/dictionary/${encodeURIComponent(word.word)}`}
        className="inline-flex items-center gap-2 text-sm font-medium text-white group-hover:text-indigo-300 transition-colors"
      >
        Xem chi tiết
        <svg className="w-4 h-4 transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
        </svg>
      </Link>
    </div>
  );
}
