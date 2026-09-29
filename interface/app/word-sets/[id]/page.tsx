"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { WordSetAPI, WordViewerItemDTO, WordSetDetailResponse } from "@/lib/wordset-api";
import { WordViewerCard } from "@/components/wordsets/WordViewerCard";
import { ReviewSetModal } from "@/components/wordsets/ReviewSetModal";

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

type LastDeletedState = {
  words: WordViewerItemDTO[];
  wordIds: string[];
  previousActiveWordKey: string | null;
} | null;

export default function WordSetViewerPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();

  const [set, setSet] = useState<WordSetDetailResponse | null>(null);
  const [words, setWords] = useState<WordViewerItemDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Composite key-based active word navigation (id:senseIndex)
  const [activeWordKey, setActiveWordKey] = useState<string | null>(null);

  const [isCloning, setIsCloning] = useState(false);
  const [isUpdatingVisibility, setIsUpdatingVisibility] = useState(false);
  const [showTopicsDropdown, setShowTopicsDropdown] = useState(false);
  const [forceReveal, setForceReveal] = useState(false);
  const [isReviewModalOpen, setIsReviewModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  // Batch delete — use composite key "wordId:senseIndex" to avoid collision when same word has multiple senses
  const [selectedWordIds, setSelectedWordIds] = useState<Set<string>>(new Set());
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [lastDeleted, setLastDeleted] = useState<LastDeletedState>(null);
  const [undoVisible, setUndoVisible] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const undoTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [setRes, wordsRes] = await Promise.all([
        WordSetAPI.getWordSet(id) as Promise<WordSetDetailResponse>,
        WordSetAPI.getWordSetWords(id),
      ]);
      setSet(setRes);
      setWords(wordsRes);
      if (wordsRes.length > 0) setActiveWordKey(`${wordsRes[0].id}:${wordsRes[0].sense_id ?? ""}`);
    } catch (err: any) {
      console.error(err);
      setError("Failed to load Word Set.");
    } finally {
      setLoading(false);
    }
  };

  const activeWord = useMemo(() => words.find((w) => `${w.id}:${w.sense_id ?? ""}` === activeWordKey) ?? null, [words, activeWordKey]);
  const activeWordIndex = useMemo(() => words.findIndex((w) => `${w.id}:${w.sense_id ?? ""}` === activeWordKey), [words, activeWordKey]);

  // Memoize search results
  const filteredWords = useMemo(() => {
    if (!searchQuery.trim()) return words;
    const q = searchQuery.toLowerCase();
    return words.filter((item) => item.word.toLowerCase().includes(q));
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

  const handleClone = async () => {
    try {
      setIsCloning(true);
      const newSet = await WordSetAPI.cloneWordSet(id);
      router.push(`/word-sets/${newSet.id}`);
    } catch (err) {
      console.error(err);
      alert("Failed to clone Word Set");
    } finally {
      setIsCloning(false);
    }
  };

  const toggleVisibility = async () => {
    if (!set.is_owner || isUpdatingVisibility) return;
    const newVisibility = set.visibility === "public" ? "private" : "public";
    try {
      setIsUpdatingVisibility(true);
      await WordSetAPI.updateWordSet(id, { visibility: newVisibility });
      setSet({ ...set, visibility: newVisibility });
    } catch { alert("Failed to update visibility"); }
    finally { setIsUpdatingVisibility(false); }
  };

  const toggleWordSelection = (wordId: string, senseId: string) => {
    const key = `${wordId}:${senseId}`;
    setSelectedWordIds((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key); else next.add(key);
      return next;
    });
  };

  const clearSelection = () => { setSelectedWordIds(new Set()); setShowDeleteConfirm(false); };

  const clearUndo = () => {
    if (undoTimerRef.current) clearTimeout(undoTimerRef.current);
    setUndoVisible(false); setLastDeleted(null);
  };

  const resolveNextActiveWordKey = (remainingWords: WordViewerItemDTO[], deletedKeys: Set<string>, currentActiveKey: string | null): string | null => {
    if (remainingWords.length === 0) return null;
    if (currentActiveKey && !deletedKeys.has(currentActiveKey)) return currentActiveKey;
    const currentIdx = words.findIndex((w) => `${w.id}:${w.sense_id ?? ""}` === currentActiveKey);
    for (let i = currentIdx + 1; i < words.length; i++) { 
      const key = `${words[i].id}:${words[i].sense_id ?? ""}`;
      if (!deletedKeys.has(key)) return key; 
    }
    for (let i = currentIdx - 1; i >= 0; i--) { 
      const key = `${words[i].id}:${words[i].sense_id ?? ""}`;
      if (!deletedKeys.has(key)) return key; 
    }
    return `${remainingWords[0].id}:${remainingWords[0].sense_id ?? ""}`;
  };

  const handleRemoveWords = async () => {
    if (selectedWordIds.size === 0 || isDeleting) return;
    // Extract word IDs from composite keys "wordId:senseIndex"
    const wordIdsList = Array.from(selectedWordIds).map((key) => key.split(":")[0]);
    const deletedKeys = new Set(selectedWordIds);
    const deletedWords = words.filter((w) => deletedKeys.has(`${w.id}:${w.sense_id ?? ""}`));
    const snapshot: LastDeletedState = { words: deletedWords, wordIds: wordIdsList, previousActiveWordKey: activeWordKey };
    const prevWords = words;
    const prevActiveKey = activeWordKey;
    const newWords = words.filter((w) => !deletedKeys.has(`${w.id}:${w.sense_id ?? ""}`));
    const deletedIdSet = new Set(deletedWords.map((w) => `${w.id}:${w.sense_id ?? ""}`));
    const nextActiveKey = resolveNextActiveWordKey(newWords, deletedIdSet, activeWordKey);
    setWords(newWords);
    setActiveWordKey(nextActiveKey);
    setSet((prev) => prev ? { ...prev, word_count: Math.max(0, prev.word_count - deletedWords.length) } : prev);
    setSelectedWordIds(new Set());
    setShowDeleteConfirm(false);
    setIsDeleting(true);
    try {
      await WordSetAPI.removeWordsFromSet(id, wordIdsList);
      clearUndo();
      setLastDeleted(snapshot);
      setUndoVisible(true);
      undoTimerRef.current = setTimeout(() => { setUndoVisible(false); setLastDeleted(null); }, 8000);
    } catch (err) {
      console.error(err);
      setWords(prevWords);
      setActiveWordKey(prevActiveKey);
      setSet((prev) => prev ? { ...prev, word_count: prev.word_count + deletedWords.length } : prev);
      alert("Failed to remove words. Please try again.");
    } finally { setIsDeleting(false); }
  };

  const handleUndo = async () => {
    if (!lastDeleted || isRestoring) return;
    const snapshot = lastDeleted;
    const prevWords = words;
    const prevActiveKey = activeWordKey;
    const restoredWords = snapshot.words.filter((w) => !words.find((e) => e.id === w.id && e.sense_id === w.sense_id));
    const finalWords = [...words, ...restoredWords];
    clearUndo();
    setWords(finalWords);
    setActiveWordKey(snapshot.previousActiveWordKey ?? (finalWords.length > 0 ? `${finalWords[0].id}:${finalWords[0].sense_id ?? ""}` : null));
    setSet((prev) => prev ? { ...prev, word_count: prev.word_count + snapshot.wordIds.length } : prev);
    setIsRestoring(true);
    try {
      await WordSetAPI.restoreWordsToSet(id, snapshot.wordIds);
    } catch (err) {
      console.error(err);
      setWords(prevWords);
      setActiveWordKey(prevActiveKey);
      setSet((prev) => prev ? { ...prev, word_count: prev.word_count - snapshot.wordIds.length } : prev);
      setLastDeleted(snapshot);
      setUndoVisible(true);
      alert("Failed to undo deletion. Please try again.");
    } finally { setIsRestoring(false); }
  };

  return (
    <div className="flex flex-col h-screen overflow-hidden">
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
          <div className="flex flex-col min-w-0">
            <h1 
              className="text-white font-bold text-lg leading-tight truncate max-w-[200px] sm:max-w-[300px] lg:max-w-[500px]"
              title={set.name}
            >
              {set.name}
            </h1>
            {!set.is_owner && set.creator_name && (
              <span className="text-xs text-white/40 truncate">Created by {set.creator_name}</span>
            )}
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <span className="text-white/30 text-sm">{set.word_count} words</span>
            {set.clone_count > 0 && (
              <>
                <span className="w-1 h-1 rounded-full bg-white/20"></span>
                <span className="text-indigo-300/80 text-sm font-medium">{set.clone_count} learners</span>
              </>
            )}
            <span className="w-1 h-1 rounded-full bg-white/20"></span>
            <span className={`text-xs font-semibold px-2 py-0.5 rounded-md ${
              set.difficulty === 'Beginner' ? 'bg-emerald-500/10 text-emerald-400' :
              set.difficulty === 'Intermediate' ? 'bg-blue-500/10 text-blue-400' :
              'bg-rose-500/10 text-rose-400'
            }`}>
              {set.difficulty || "Beginner"}
            </span>
            <span className="w-1 h-1 rounded-full bg-white/20"></span>
            {set.is_owner ? (
              <button
                onClick={toggleVisibility}
                disabled={isUpdatingVisibility}
                className={`text-xs font-medium uppercase tracking-wider px-2 py-0.5 rounded transition-colors disabled:opacity-50 ${
                  set.visibility === 'public' 
                    ? 'text-emerald-500/70 bg-emerald-500/10 hover:bg-emerald-500/20' 
                    : 'text-indigo-500/70 bg-indigo-500/10 hover:bg-indigo-500/20'
                }`}
                title="Click to toggle visibility"
              >
                {set.visibility}
              </button>
            ) : (
              <span className={`text-xs font-medium uppercase tracking-wider ${set.visibility === 'public' ? 'text-emerald-500/70' : 'text-indigo-500/70'}`}>
                {set.visibility}
              </span>
            )}
          </div>
          {set.topics && set.topics.length > 0 && (
            <div className="relative ml-4">
              <button 
                onClick={() => setShowTopicsDropdown(!showTopicsDropdown)}
                className="text-xs font-semibold uppercase tracking-wider text-white/50 border border-white/10 px-2 py-0.5 rounded-md bg-white/5 hover:bg-white/10 transition-colors flex items-center gap-1"
              >
                +{set.topics.length} Tags
                <svg className="w-3 h-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {showTopicsDropdown && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setShowTopicsDropdown(false)} />
                  <div className="absolute top-full left-0 mt-2 min-w-[140px] bg-[#1a1a2e] border border-white/10 rounded-xl shadow-2xl z-20 py-2 flex flex-col">
                    {set.topics.map(t => (
                      <span key={t} className="px-4 py-1.5 text-xs text-white/70 hover:text-white hover:bg-white/5 transition-colors">
                        {t}
                      </span>
                    ))}
                  </div>
                </>
              )}
            </div>
          )}
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setForceReveal(!forceReveal)}
            className={`px-4 py-2 rounded-xl transition-colors text-sm font-medium flex items-center gap-2 ${
              forceReveal 
                ? 'bg-amber-500/20 text-amber-400 hover:bg-amber-500/30 hover:text-amber-300' 
                : 'bg-white/5 text-white/50 hover:bg-white/10 hover:text-white'
            }`}
            title={forceReveal ? "Deactivate Flip Mode" : "Activate Flip Mode"}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
            {forceReveal ? "Hide Details" : "Show Details"}
          </button>
          <button
            onClick={() => setIsReviewModalOpen(true)}
            className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/30 text-blue-400 hover:text-blue-300 transition-colors text-sm font-medium flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Review Set
          </button>
          
          {set.is_owner ? (
            <Link
              href={`/word-sets/${set.id}/builder`}
              className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 text-white/70 hover:text-white transition-colors text-sm font-medium flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
              Edit Set
            </Link>
          ) : (
            <button
              onClick={handleClone}
              disabled={isCloning}
              className="px-4 py-2 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white transition-colors text-sm font-semibold flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isCloning ? (
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
              )}
              Add to My Sets
            </button>
          )}
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
              filteredWords.map((word) => {
                const isActive = `${word.id}:${word.sense_id ?? ""}` === activeWordKey;
                const isSelected = selectedWordIds.has(`${word.id}:${word.sense_id ?? ""}`);
                const wordOriginalIndex = words.indexOf(word);
                return (
                  <div
                    key={`${word.id}:${word.sense_id ?? ""}:${wordOriginalIndex}`}
                    className={`group w-full flex items-center gap-2 px-2 py-2 rounded-lg transition-colors border ${
                      isSelected ? "bg-red-500/10 border-red-500/20" :
                      isActive ? "bg-indigo-500/20 border-indigo-500/30" :
                      "border-transparent hover:bg-white/5"
                    }`}
                  >
                    {set.is_owner && (
                      <button
                        onClick={() => toggleWordSelection(word.id, word.sense_id ?? "")}
                        className={`shrink-0 w-4 h-4 rounded border flex items-center justify-center transition-all ${
                          isSelected ? "bg-red-500 border-red-500 opacity-100" : "border-white/20 bg-transparent opacity-0 group-hover:opacity-100 hover:border-white/50"
                        }`}
                        title={isSelected ? "Deselect" : "Select for deletion"}
                      >
                        {isSelected && <svg className="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}><path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" /></svg>}
                      </button>
                    )}
                    <button onClick={() => setActiveWordKey(`${word.id}:${word.sense_id ?? ""}`)} className="flex-1 text-left flex items-center justify-between min-w-0">
                      <div className="flex flex-col truncate pr-2">
                        <div className="flex items-center gap-3 truncate">
                          <span className={`text-xs font-mono w-5 opacity-50 shrink-0 ${isActive ? "text-indigo-300" : "text-white"}`}>{(wordOriginalIndex + 1).toString().padStart(2, "0")}</span>
                          <span className={`font-medium truncate ${isActive ? "text-indigo-100" : isSelected ? "text-red-200" : "text-white/60"}`}>{word.word}</span>
                        </div>
                        <span className="text-white/30 text-[10px] pl-8 truncate">
                          {word.definitions?.find((d) => d.sense_id === word.sense_id)?.definition || word.definitions?.[0]?.definition}
                        </span>
                      </div>
                      <span className={`text-[10px] font-bold uppercase tracking-wider shrink-0 ${progressColors[word.progress_state] || "text-white/40"}`}>{progressLabels[word.progress_state] || "New"}</span>
                    </button>
                  </div>
                );
              })
            )}
          </div>

          {/* Selection Action Bar */}
          {set.is_owner && selectedWordIds.size > 0 && (
            <div className="border-t border-white/10 bg-black/30 p-3">
              {!showDeleteConfirm ? (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-white/60"><span className="font-semibold text-white">{selectedWordIds.size}</span> selected</span>
                  <div className="flex items-center gap-2">
                    <button onClick={clearSelection} className="text-xs text-white/40 hover:text-white transition-colors px-2 py-1">Cancel</button>
                    <button onClick={() => setShowDeleteConfirm(true)} disabled={isDeleting} className="text-xs font-semibold bg-red-500/20 hover:bg-red-500/30 text-red-400 hover:text-red-300 px-3 py-1.5 rounded-lg transition-colors disabled:opacity-50 flex items-center gap-1.5">
                      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                      Delete
                    </button>
                  </div>
                </div>
              ) : (
                <div className="space-y-2">
                  <p className="text-sm text-white/70">Remove <span className="font-semibold text-red-400">{selectedWordIds.size} {selectedWordIds.size === 1 ? "word" : "words"}</span> from this set?</p>
                  <div className="flex gap-2">
                    <button onClick={() => setShowDeleteConfirm(false)} className="flex-1 text-xs py-2 rounded-lg bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-colors">Cancel</button>
                    <button onClick={handleRemoveWords} disabled={isDeleting} className="flex-1 text-xs py-2 rounded-lg bg-red-500 hover:bg-red-600 text-white font-semibold transition-colors disabled:opacity-50 flex items-center justify-center gap-1.5">
                      {isDeleting ? <div className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin" /> : <><svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>Confirm Delete</>}
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Card Area */}
        <div className="flex-1 p-6 md:p-12 overflow-hidden bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-[#1e1e2d] to-[#0f0f17] relative">
          {words.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center">
              <div className="text-6xl mb-6">📭</div>
              <h3 className="text-2xl font-bold text-white mb-2">This Set is Empty</h3>
              <p className="text-white/50 mb-8 max-w-md">Add some words to this set using the dictionary or the builder tool.</p>
              <Link href={`/word-sets/${set.id}/builder`} className="px-6 py-3 bg-indigo-500 hover:bg-indigo-600 text-white rounded-xl font-bold transition-all shadow-lg shadow-indigo-500/20">Open Builder</Link>
            </div>
          ) : activeWord ? (
            <div className="h-full max-w-4xl mx-auto flex flex-col">
              <div className="flex-1 min-h-0">
                <WordViewerCard
                  word={activeWord}
                  isActive={true}
                  forceReveal={forceReveal}
                  hasPrev={activeWordIndex > 0}
                  hasNext={activeWordIndex < words.length - 1}
                  onPrev={() => { if (activeWordIndex > 0) setActiveWordKey(`${words[activeWordIndex - 1].id}:${words[activeWordIndex - 1].sense_id ?? ""}`); }}
                  onNext={() => { if (activeWordIndex < words.length - 1) setActiveWordKey(`${words[activeWordIndex + 1].id}:${words[activeWordIndex + 1].sense_id ?? ""}`); }}
                />
              </div>
              <div className="shrink-0 mt-6 text-center text-white/30 text-sm font-mono tracking-widest">{activeWordIndex + 1} / {words.length}</div>
            </div>
          ) : (
            <div className="h-full flex items-center justify-center text-white/30">Word not found</div>
          )}

          {/* Undo Snackbar */}
          {undoVisible && lastDeleted && (
            <div className="absolute bottom-8 left-1/2 -translate-x-1/2 flex items-center gap-4 bg-[#1e1e2d] border border-white/10 rounded-xl px-5 py-3 shadow-2xl shadow-black/50">
              <span className="text-sm text-white/70">
                <span className="font-semibold text-white">{lastDeleted.wordIds.length}</span>{" "}
                {lastDeleted.wordIds.length === 1 ? "word" : "words"} removed
              </span>
              <button onClick={handleUndo} disabled={isRestoring} className="text-sm font-bold text-indigo-400 hover:text-indigo-300 transition-colors disabled:opacity-50 flex items-center gap-1.5">
                {isRestoring && <div className="w-3.5 h-3.5 border-2 border-indigo-400/30 border-t-indigo-400 rounded-full animate-spin" />}
                Undo
              </button>
              <button onClick={clearUndo} className="text-white/20 hover:text-white/60 transition-colors" title="Dismiss">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
          )}
        </div>
      </div>

      <ReviewSetModal setId={set.id} isOpen={isReviewModalOpen} onClose={() => setIsReviewModalOpen(false)} />
    </div>
  );
}
