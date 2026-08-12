"use client";

import { WordSet } from "@/lib/wordset-api";
import Link from "next/link";
import { useState } from "react";

interface Props {
  set: WordSet;
  onEdit: (set: WordSet) => void;
  onDelete: (set: WordSet) => void;
  formatRelativeTime: (date: string) => string;
}

export function WordSetCard({ set, onEdit, onDelete, formatRelativeTime }: Props) {
  const [showMenu, setShowMenu] = useState(false);

  // Visual fill bar (word_count as relative fill, 50 is a visual reference only)
  const barFill = Math.min(Math.round((set.word_count / 50) * 100), 100);

  return (
    <div className="group relative flex flex-col rounded-2xl bg-white/[0.04] border border-white/10 hover:border-white/20 hover:bg-white/[0.07] transition-all duration-200 shadow-sm overflow-hidden">
      {/* Top accent bar by visibility */}
      <div
        className={`h-0.5 w-full ${set.visibility === "public" ? "bg-emerald-500/60" : "bg-indigo-500/40"}`}
      />

      <div className="p-5 flex flex-col flex-1 gap-4">
        {/* Header row */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <h3 className="text-white font-bold text-lg leading-snug truncate group-hover:text-indigo-300 transition-colors">
              {set.name}
            </h3>
            <div className="flex items-center gap-1.5 mt-1">
              {set.visibility === "private" ? (
                <>
                  <svg className="w-3 h-3 text-white/30" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-xs text-white/30 uppercase tracking-wider font-medium">Private</span>
                </>
              ) : (
                <>
                  <svg className="w-3 h-3 text-emerald-400/60" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM4.332 8.027a6.012 6.012 0 011.912-2.706C6.512 5.73 6.5 5.969 6.5 6c0 1.105.896 2 2 2a1 1 0 010 2A2 2 0 016.5 8c0-.168.012-.33.034-.49z" clipRule="evenodd" />
                  </svg>
                  <span className="text-xs text-emerald-400/60 uppercase tracking-wider font-medium">Public</span>
                </>
              )}
            </div>
          </div>

          {/* 3-dot menu */}
          <div className="relative z-10 shrink-0">
            <button
              onClick={() => setShowMenu(!showMenu)}
              className="w-8 h-8 rounded-full flex items-center justify-center text-white/30 hover:text-white hover:bg-white/10 transition-colors"
            >
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <circle cx="12" cy="5" r="1.5" /><circle cx="12" cy="12" r="1.5" /><circle cx="12" cy="19" r="1.5" />
              </svg>
            </button>
            {showMenu && (
              <>
                <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)} />
                <div className="absolute right-0 mt-2 w-36 bg-[#1a1a2e] border border-white/10 rounded-xl shadow-2xl z-20 py-1.5 overflow-hidden">
                  <button
                    onClick={() => { setShowMenu(false); onEdit(set); }}
                    className="w-full text-left px-4 py-2 text-sm text-white/60 hover:text-white hover:bg-white/5 transition-colors"
                  >
                    Rename
                  </button>
                  <div className="mx-3 my-1 h-px bg-white/10" />
                  <button
                    onClick={() => { setShowMenu(false); onDelete(set); }}
                    className="w-full text-left px-4 py-2 text-sm text-red-400 hover:text-red-300 hover:bg-red-400/5 transition-colors"
                  >
                    Delete
                  </button>
                </div>
              </>
            )}
          </div>
        </div>

        {/* Description */}
        {set.description && (
          <p className="text-white/40 text-sm leading-relaxed line-clamp-2">{set.description}</p>
        )}

        {/* Word count + bar */}
        <div className="space-y-2">
          <div className="flex justify-between items-baseline">
            <span className="text-white font-semibold text-sm">{set.word_count} words</span>
          </div>
          <div className="h-1.5 w-full bg-white/5 rounded-full overflow-hidden">
            <div
              className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500"
              style={{ width: `${barFill}%` }}
            />
          </div>
        </div>

        {/* Updated timestamp */}
        <p className="text-white/25 text-xs">Updated {formatRelativeTime(set.updated_at)}</p>

        {/* Actions */}
        <div className="flex gap-2 mt-auto pt-1">
          <Link
            href={`/word-sets/${set.id}`}
            className="flex-1 py-2.5 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-semibold text-center transition-all shadow-sm shadow-indigo-500/20"
          >
            Open Set
          </Link>
          <Link
            href={`/word-sets/${set.id}/builder`}
            className="px-4 py-2.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 text-white/60 hover:text-white text-sm font-medium text-center transition-all"
            title="Open Builder"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </Link>
        </div>
      </div>
    </div>
  );
}
