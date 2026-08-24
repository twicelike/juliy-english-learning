"use client";

import { WordSetListItem } from "@/lib/wordset-api";
import Link from "next/link";
import { useState, useEffect } from "react";
import { createPortal } from "react-dom";

interface Props {
  set: WordSetListItem;
  onEdit?: (set: WordSetListItem) => void;
  onDelete?: (set: WordSetListItem) => void;
  formatRelativeTime: (date: string) => string;
  mode?: "library" | "explore";
}

export function WordSetCard({ set, onEdit, onDelete, formatRelativeTime, mode = "library" }: Props) {
  const [showMenu, setShowMenu] = useState(false);
  const [showDetails, setShowDetails] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Visual fill bar (word_count as relative fill, 50 is a visual reference only)
  const barFill = Math.min(Math.round((set.word_count / 50) * 100), 100);

  return (
    <div className="card group relative flex flex-col rounded-2xl bg-white/[0.04] border border-white/10 hover:border-white/20 hover:bg-white/[0.07] transition-all duration-200 shadow-sm overflow-hidden">
      {/* Top accent bar by visibility */}
      <div
        className={`h-0.5 w-full ${set.visibility === "public" ? "bg-emerald-500/60" : "bg-indigo-500/40"}`}
      />

      <div className="p-5 flex flex-col flex-1 gap-4">
        {/* Header row */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1 min-w-0">
            <h3 className="font-bold text-lg leading-snug truncate group-hover:text-indigo-400 transition-colors">
              <Link href={`/word-sets/${set.id}`} className="after:absolute after:inset-0 focus:outline-none">
                {set.name}
              </Link>
            </h3>
            <div className="flex items-center gap-1.5 mt-1">
              {set.visibility === "private" ? (
                <>
                  <svg className="w-3 h-3 opacity-30" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-xs opacity-30 uppercase tracking-wider font-medium">Private</span>
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

          {/* Eye Icon & 3-dot menu */}
          <div className="relative z-10 shrink-0 flex items-center gap-1">
            <button
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setShowDetails(true);
              }}
              className="w-8 h-8 rounded-full flex items-center justify-center opacity-60 hover:opacity-100 hover:bg-white/10 transition-colors"
              title="View full details"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            </button>
            {mode === "library" && (
              <div className="relative">
                <button
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setShowMenu(!showMenu);
                  }}
                  className="w-8 h-8 rounded-full flex items-center justify-center opacity-60 hover:opacity-100 hover:bg-white/10 transition-colors"
                >
                  <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                    <circle cx="12" cy="5" r="1.5" /><circle cx="12" cy="12" r="1.5" /><circle cx="12" cy="19" r="1.5" />
                  </svg>
                </button>
                {showMenu && (
                  <>
                    <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)} />
                    <div className="absolute right-0 mt-2 w-36 bg-white/5 border border-white/10 rounded-xl shadow-2xl z-20 py-1.5 overflow-hidden backdrop-blur-md">
                      <button
                        onClick={(e) => { e.preventDefault(); e.stopPropagation(); setShowMenu(false); if (onEdit) onEdit(set); }}
                        className="w-full text-left px-4 py-2 text-sm opacity-60 hover:opacity-100 hover:bg-white/5 transition-colors"
                      >
                        Edit Details
                      </button>
                      <div className="mx-3 my-1 h-px bg-white/10" />
                      <button
                        onClick={(e) => { e.preventDefault(); e.stopPropagation(); setShowMenu(false); if (onDelete) onDelete(set); }}
                        className="w-full text-left px-4 py-2 text-sm text-red-400 hover:text-red-300 hover:bg-red-400/5 transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Description */}
        {set.description && (
          <p className="opacity-40 text-sm leading-relaxed line-clamp-2 break-all">{set.description}</p>
        )}

        {/* Metadata for Explore Mode */}
        {mode === "explore" && (
          <div className="flex flex-col gap-2 mt-1">
            <div className="flex items-center gap-2">
              <span className={`text-xs font-semibold px-2 py-0.5 rounded-md ${
                set.difficulty === 'Beginner' ? 'bg-emerald-500/10 text-emerald-400' :
                set.difficulty === 'Intermediate' ? 'bg-blue-500/10 text-blue-400' :
                'bg-rose-500/10 text-rose-400'
              }`}>
                {set.difficulty || "Beginner"}
              </span>
              <span className="opacity-40 text-xs flex items-center gap-1">
                <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
                {set.creator_name || "Unknown"}
              </span>
            </div>
            {set.topics && set.topics.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {set.topics.map(t => (
                  <span key={t} className="text-[10px] uppercase tracking-wider opacity-30 border border-white/10 px-1.5 py-0.5 rounded">
                    {t}
                  </span>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Word count + bar */}
        <div className="space-y-2 mt-auto pt-2">
          <div className="flex justify-between items-baseline">
            <span className="font-semibold text-sm">{set.word_count} words</span>
            {mode === "explore" && set.clone_count > 0 && (
              <span className="text-indigo-300/80 text-xs font-medium">{set.clone_count} learners</span>
            )}
          </div>
          <div className="progress-track h-1.5 w-full bg-white/5 rounded-full overflow-hidden">
            <div
              className="progress-fill h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500"
              style={{ width: `${barFill}%` }}
            />
          </div>
        </div>

        {/* Updated timestamp */}
        <p className="opacity-30 text-xs">Updated {formatRelativeTime(set.updated_at)}</p>

        {/* Actions */}
        {mode === "library" && (
          <div className="flex justify-end pt-1">
            <Link
              href={`/word-sets/${set.id}/builder`}
              className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 opacity-80 hover:opacity-100 text-sm font-medium transition-all relative z-10 flex items-center gap-2"
              title="Open Builder"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              <span>Add Words</span>
            </Link>
          </div>
        )}
      </div>

      {/* Details Modal */}
      {mounted && showDetails && createPortal(
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" onClick={() => setShowDetails(false)}>
          <div 
            className="card no-hover bg-[#1a1a1a] border border-white/10 rounded-2xl p-6 max-w-lg w-full shadow-2xl relative animate-in fade-in zoom-in-95 duration-200"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              onClick={() => setShowDetails(false)}
              className="absolute top-5 right-5 opacity-40 hover:opacity-100 transition-opacity text-white"
            >
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
            <h2 className="text-2xl font-bold mb-4 pr-8 leading-tight break-all text-white">{set.name}</h2>
            {set.description ? (
              <div className="opacity-80 whitespace-pre-wrap leading-relaxed max-h-[60vh] overflow-y-auto custom-scrollbar pr-2 text-base break-all text-white">
                {set.description}
              </div>
            ) : (
              <p className="opacity-30 italic text-white">No description provided.</p>
            )}
          </div>
        </div>,
        document.body
      )}
    </div>
  );
}
