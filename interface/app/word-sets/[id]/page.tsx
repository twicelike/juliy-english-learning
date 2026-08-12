"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { WordSetAPI, WordSet, WordViewerItemDTO } from "@/lib/wordset-api";
import { WordViewerCard } from "@/components/wordsets/WordViewerCard";

const progressLabels: Record<string, string> = {
  new: "New",
  learning: "Learning",
  due: "Due",
  mastered: "Mastered",
};

const progressColors: Record<string, string> = {
  new: "text-white/40",
  learning: "text-blue-400",
  due: "text-amber-400",
  mastered: "text-emerald-400",
};

export default function WordSetViewerPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();

  const [set, setSet] = useState<WordSet | null>(null);
  const [words, setWords] = useState<WordViewerItemDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [searchQuery, setSearchQuery] = useState("");
  const [activeWordIndex, setActiveWordIndex] = useState(0);

  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [setRes, wordsRes] = await Promise.all([
        WordSetAPI.getWordSet(id),
        WordSetAPI.getWordSetWords(id),
      ]);
      setSet(setRes);
      setWords(wordsRes);
    } catch (err: any) {
      console.error(err);
      setError("Failed to load Word Set.");
    } finally {
      setLoading(false);
    }
  };

  // Memoize search results but keep their original index for correct navigation
  const filteredWords = useMemo(() => {
    if (!searchQuery.trim()) {
      return words.map((word, index) => ({ word, index }));
    }
    const q = searchQuery.toLowerCase();
    return words
      .map((word, index) => ({ word, index }))
      .filter((item) => item.word.word.toLowerCase().includes(q));
  }, [words, searchQuery]);

  // Calculate Progress Stats
  const stats = useMemo(() => {
    let newCount = 0, learningCount = 0, dueCount = 0, masteredCount = 0;
    words.forEach(w => {
      if (w.progress_state === "new") newCount++;
      else if (w.progress_state === "learning") learningCount++;
      else if (w.progress_state === "due") dueCount++;
      else if (w.progress_state === "mastered") masteredCount++;
    });
    const total = words.length;
    const masteryPercent = total === 0 ? 0 : Math.round((masteredCount / total) * 100);
    return { newCount, learningCount, dueCount, masteredCount, total, masteryPercent };
  }, [words]);

  if (loading) {
    return (
      <div className="h-screen w-full flex items-center justify-center bg-[#0f0f17]">
        <div className="w-10 h-10 border-4 border-indigo-500/20 border-t-indigo-500 rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error || !set) {
    return (
      <div className="h-screen w-full flex flex-col items-center justify-center bg-[#0f0f17]">
        <div className="text-white/50 mb-4">{error || "Word Set not found"}</div>
        <Link href="/word-sets" className="px-6 py-2 bg-white/10 rounded-xl text-white hover:bg-white/20 transition-colors">
          Back to Library
        </Link>
      </div>
    );
  }

  const activeWord = words[activeWordIndex];

  return (
    <div className="flex flex-col h-screen bg-[#0f0f17] overflow-hidden">
      {/* ─── Header ────────────────────────────────────────────── */}
      <header className="h-16 shrink-0 border-b border-white/10 flex items-center justify-between px-6 bg-black/20">
        <div className="flex items-center gap-6">
          <Link
            href="/word-sets"
            className="flex items-center gap-2 text-white/50 hover:text-white transition-colors text-sm font-medium"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back to Library
          </Link>
          <div className="h-6 w-px bg-white/10" />
          <h1 className="text-white font-bold text-lg">{set.name}</h1>
          <div className="flex items-center gap-2">
            <span className="text-white/30 text-sm">{set.word_count} words</span>
            <span className="w-1 h-1 rounded-full bg-white/20"></span>
            <span className={`text-xs font-medium uppercase tracking-wider ${set.visibility === 'public' ? 'text-emerald-500/70' : 'text-indigo-500/70'}`}>
              {set.visibility}
            </span>
          </div>
        </div>
        <div>
          <Link
            href={`/word-sets/${set.id}/builder`}
            className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 text-white/70 hover:text-white transition-colors text-sm font-medium flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
            Edit Set
          </Link>
        </div>
      </header>

      {/* ─── Main Content ──────────────────────────────────────── */}
      <div className="flex-1 flex overflow-hidden">
        
        {/* Sidebar */}
        <div className="w-80 shrink-0 border-r border-white/10 bg-black/10 flex flex-col hidden md:flex">
          
          {/* Progress Summary Header */}
          <div className="p-4 border-b border-white/10 space-y-4">
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div className="flex justify-between p-2 rounded-lg bg-white/5">
                <span className="text-white/50">Mastered</span>
                <span className="text-emerald-400 font-medium">{stats.masteredCount}</span>
              </div>
              <div className="flex justify-between p-2 rounded-lg bg-white/5">
                <span className="text-white/50">Learning</span>
                <span className="text-blue-400 font-medium">{stats.learningCount}</span>
              </div>
              <div className="flex justify-between p-2 rounded-lg bg-white/5">
                <span className="text-white/50">Due</span>
                <span className="text-amber-400 font-medium">{stats.dueCount}</span>
              </div>
              <div className="flex justify-between p-2 rounded-lg bg-white/5">
                <span className="text-white/50">New</span>
                <span className="text-white/40 font-medium">{stats.newCount}</span>
              </div>
            </div>

            <div className="space-y-1.5">
              <div className="flex justify-between text-xs">
                <span className="text-white/50 font-medium tracking-wider uppercase">Mastery</span>
                <span className="text-white font-mono">{stats.masteryPercent}%</span>
              </div>
              <div className="h-1.5 w-full bg-white/5 rounded-full overflow-hidden flex">
                <div 
                  className="h-full bg-emerald-500 transition-all duration-1000"
                  style={{ width: `${stats.masteryPercent}%` }}
                />
              </div>
            </div>
          </div>

          <div className="p-4 border-b border-white/10">
            <div className="relative">
              <svg className="w-4 h-4 text-white/30 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
              <input
                type="text"
                placeholder="Search words..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-lg py-2 pl-9 pr-3 text-sm text-white placeholder-white/30 focus:outline-none focus:border-indigo-500 transition-colors"
              />
            </div>
          </div>
          
          <div className="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-1">
            {filteredWords.length === 0 ? (
              <div className="p-4 text-center text-white/30 text-sm">No words found.</div>
            ) : (
              filteredWords.map(({ word, index }) => {
                const isActive = index === activeWordIndex;
                return (
                  <button
                    key={word.id}
                    onClick={() => setActiveWordIndex(index)}
                    className={`w-full text-left px-3 py-2.5 rounded-lg flex items-center justify-between transition-colors ${
                      isActive 
                        ? "bg-indigo-500/20 text-indigo-100 border border-indigo-500/30" 
                        : "text-white/60 hover:bg-white/5 hover:text-white border border-transparent"
                    }`}
                  >
                    <div className="flex items-center gap-3 truncate">
                      <span className={`text-xs font-mono w-5 opacity-50 ${isActive ? "text-indigo-300" : ""}`}>
                        {(index + 1).toString().padStart(2, '0')}
                      </span>
                      <span className="font-medium truncate">{word.word}</span>
                    </div>
                    <span className={`text-[10px] font-bold uppercase tracking-wider ${progressColors[word.progress_state] || "text-white/40"}`}>
                      {progressLabels[word.progress_state] || "New"}
                    </span>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {/* Card Area */}
        <div className="flex-1 p-6 md:p-12 overflow-hidden bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-[#1e1e2d] to-[#0f0f17]">
          {words.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center">
              <div className="text-6xl mb-6">📭</div>
              <h3 className="text-2xl font-bold text-white mb-2">This Set is Empty</h3>
              <p className="text-white/50 mb-8 max-w-md">Add some words to this set using the dictionary or the builder tool.</p>
              <Link
                href={`/word-sets/${set.id}/builder`}
                className="px-6 py-3 bg-indigo-500 hover:bg-indigo-600 text-white rounded-xl font-bold transition-all shadow-lg shadow-indigo-500/20"
              >
                Open Builder
              </Link>
            </div>
          ) : activeWord ? (
            <div className="h-full max-w-4xl mx-auto flex flex-col">
              <div className="flex-1 min-h-0">
                <WordViewerCard
                  word={activeWord}
                  isActive={true}
                  hasPrev={activeWordIndex > 0}
                  hasNext={activeWordIndex < words.length - 1}
                  onPrev={() => setActiveWordIndex(i => Math.max(0, i - 1))}
                  onNext={() => setActiveWordIndex(i => Math.min(words.length - 1, i + 1))}
                />
              </div>
              <div className="shrink-0 mt-6 text-center text-white/30 text-sm font-mono tracking-widest">
                {activeWordIndex + 1} / {words.length}
              </div>
            </div>
          ) : (
            <div className="h-full flex items-center justify-center text-white/30">
              Word not found
            </div>
          )}
        </div>

      </div>
    </div>
  );
}
