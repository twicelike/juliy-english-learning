"use client";

import Link from "next/link";
import { WordDTO } from "@/lib/dictionary-api";

interface WordCardProps {
  word: WordDTO;
  title?: string;
}

export default function WordCard({ word, title = "Word of the Day" }: WordCardProps) {
  const firstDef = word.definitions[0]?.definition || "No definition available.";

  return (
    <div className="group relative p-6 bg-white/5 border border-white/10 rounded-3xl backdrop-blur-md overflow-hidden transition-all hover:bg-white/10 hover:border-indigo-500/50">
      <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/10 rounded-full blur-3xl group-hover:bg-indigo-500/20 transition-all"></div>
      
      <h3 className="text-xs text-indigo-400 uppercase tracking-widest mb-4 font-semibold">
        {title}
      </h3>
      
      <div className="mb-4">
        <h4 className="text-3xl font-bold text-white mb-2">{word.word}</h4>
        <div className="flex gap-2 text-sm">
          {word.pos && <span className="text-indigo-300">{word.pos}</span>}
          {word.ipa && <span className="text-white/40">{word.ipa}</span>}
        </div>
      </div>
      
      <p className="text-white/70 line-clamp-2 mb-6 min-h-[3rem]">
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
