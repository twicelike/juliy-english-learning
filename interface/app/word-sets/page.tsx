"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/features/auth";
import {
  WordSetAPI,
  WordSet,
  WordSetListItem,
  CreateWordSetRequest,
  UpdateWordSetRequest,
} from "@/lib/wordset-api";
import { WordSetCard } from "@/components/wordsets/WordSetCard";
import { CreateWordSetModal } from "@/components/wordsets/CreateWordSetModal";
import { EditWordSetModal } from "@/components/wordsets/EditWordSetModal";

type SortMode = "updated" | "created" | "words";

const SORT_LABELS: Record<SortMode, string> = {
  updated: "Recently Updated",
  created: "Recently Created",
  words: "Most Words",
};

function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = Date.now();
  const diff = Math.floor((now - date.getTime()) / 1000);
  if (diff < 60) return "Just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 30) return `${Math.floor(diff / 86400)}d ago`;
  return date.toLocaleDateString();
}

export default function WordSetsPage() {
  const router = useRouter();
  const { user } = useAuthStore();

  const [wordSets, setWordSets] = useState<WordSetListItem[]>([]);
  const [loading, setLoading] = useState(true);

  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState<SortMode>("updated");
  const [showSortMenu, setShowSortMenu] = useState(false);

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingWordSet, setEditingWordSet] = useState<WordSet | null>(null);
  const [deletingWordSet, setDeletingWordSet] = useState<WordSet | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }
    loadWordSets();
  }, [user, router]);

  const loadWordSets = async () => {
    try {
      setLoading(true);
      const res = await WordSetAPI.getWordSets();
      setWordSets(res.items || []);
    } catch (error) {
      console.error("Failed to load word sets:", error);
    } finally {
      setLoading(false);
    }
  };

  const totalWords = useMemo(
    () => wordSets.reduce((acc, s) => acc + s.word_count, 0),
    [wordSets]
  );

  const filteredAndSorted = useMemo(() => {
    let result = [...wordSets];

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter((s) => s.name.toLowerCase().includes(q));
    }

    result.sort((a, b) => {
      if (sortBy === "words") return b.word_count - a.word_count;
      if (sortBy === "created")
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
      // default: updated
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
    });

    return result;
  }, [wordSets, searchQuery, sortBy]);

  const handleCreate = async (req: CreateWordSetRequest) => {
    try {
      setIsSubmitting(true);
      const newSet = await WordSetAPI.createWordSet(req);
      setWordSets([newSet, ...wordSets]);
      setIsCreateModalOpen(false);
    } catch {
      alert("Failed to create word set");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEdit = async (req: UpdateWordSetRequest) => {
    if (!editingWordSet) return;
    try {
      setIsSubmitting(true);
      const updatedSet = await WordSetAPI.updateWordSet(editingWordSet.id, req);
      setWordSets(wordSets.map((ws) => (ws.id === updatedSet.id ? { ...ws, ...updatedSet } : ws)));
      setEditingWordSet(null);
    } catch {
      alert("Failed to update word set");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingWordSet) return;
    try {
      setIsSubmitting(true);
      await WordSetAPI.deleteWordSet(deletingWordSet.id);
      setWordSets(wordSets.filter((ws) => ws.id !== deletingWordSet.id));
      setDeletingWordSet(null);
    } catch {
      alert("Failed to delete word set");
    } finally {
      setIsSubmitting(false);
    }
  };



  if (loading) {
    return (
      <div className="max-w-6xl mx-auto px-6 pt-12 pb-16 animate-pulse space-y-8">
        <div className="h-28 bg-white/5 rounded-3xl" />
        <div className="h-12 bg-white/5 rounded-2xl" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="h-52 bg-white/5 rounded-2xl" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 pt-10 pb-20 space-y-6">
      {/* ─── Hero Header ──────────────────────────────────────────── */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-indigo-600/30 via-purple-600/20 to-transparent border border-white/10 px-8 py-8">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-indigo-500/10 via-transparent to-transparent pointer-events-none" />
        <div className="relative flex flex-col sm:flex-row sm:items-center justify-between gap-6">
          <div>
            <h1 className="text-3xl sm:text-4xl font-bold text-white mb-1 tracking-tight">
              Vocabulary Library
            </h1>
            <p className="text-white/50 text-sm sm:text-base">
              Build, organize, and grow your vocabulary.
            </p>
            <div className="flex gap-6 mt-4">
              <div>
                <div className="text-2xl font-bold text-white">{wordSets.length}</div>
                <div className="text-white/40 text-xs uppercase tracking-wider">Sets</div>
              </div>
              <div className="w-px bg-white/10" />
              <div>
                <div className="text-2xl font-bold text-white">{totalWords}</div>
                <div className="text-white/40 text-xs uppercase tracking-wider">Words</div>
              </div>
            </div>
          </div>

          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="shrink-0 inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white font-bold transition-all shadow-lg shadow-indigo-500/30 text-sm"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            New Set
          </button>
        </div>
      </div>

      {/* ─── My Sets Content ────────────────────────────────── */}
      <>
        {/* Search + Sort Bar */}
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <svg
                className="w-4 h-4 text-white/30 absolute left-4 top-1/2 -translate-y-1/2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                />
              </svg>
              <input
                type="text"
                placeholder="Search your word sets..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-11 pr-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/30 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all text-sm"
              />
            </div>

            {/* Sort Dropdown */}
            <div className="relative shrink-0">
              <button
                onClick={() => setShowSortMenu(!showSortMenu)}
                className="flex items-center gap-2 px-4 py-3 bg-white/5 border border-white/10 rounded-xl text-white/70 hover:text-white text-sm transition-colors whitespace-nowrap"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12" />
                </svg>
                {SORT_LABELS[sortBy]}
                <svg className={`w-3 h-3 transition-transform ${showSortMenu ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>

              {showSortMenu && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setShowSortMenu(false)} />
                  <div className="absolute right-0 mt-2 w-52 bg-[#1a1a2e] border border-white/10 rounded-xl shadow-2xl z-20 py-2 overflow-hidden">
                    {(Object.entries(SORT_LABELS) as [SortMode, string][]).map(([key, label]) => (
                      <button
                        key={key}
                        onClick={() => { setSortBy(key); setShowSortMenu(false); }}
                        className={`w-full text-left px-4 py-2.5 text-sm transition-colors ${
                          sortBy === key
                            ? "text-indigo-400 bg-indigo-500/10"
                            : "text-white/60 hover:text-white hover:bg-white/5"
                        }`}
                      >
                        {label}
                        {sortBy === key && <span className="float-right text-indigo-400">✓</span>}
                      </button>
                    ))}
                  </div>
                </>
              )}
            </div>
          </div>

          {/* Cards Grid */}
          {filteredAndSorted.length === 0 ? (
            <div className="py-20 text-center rounded-2xl border border-white/5 bg-white/[0.02]">
              {searchQuery ? (
                <>
                  <div className="text-5xl mb-4">🔍</div>
                  <h2 className="text-lg font-bold text-white mb-2">No results</h2>
                  <p className="text-white/40 text-sm">No word sets match &quot;{searchQuery}&quot;</p>
                </>
              ) : (
                <>
                  <div className="text-5xl mb-4">📚</div>
                  <h2 className="text-xl font-bold text-white mb-2">No word sets yet</h2>
                  <p className="text-white/40 mb-6 text-sm">Create your first set to start organizing your vocabulary.</p>
                  <button
                    onClick={() => setIsCreateModalOpen(true)}
                    className="px-6 py-2.5 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white font-medium transition-colors text-sm"
                  >
                    Create Word Set
                  </button>
                </>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
              {filteredAndSorted.map((set) => (
                <WordSetCard
                  key={set.id}
                  set={set}
                  onEdit={setEditingWordSet}
                  onDelete={setDeletingWordSet}
                  formatRelativeTime={formatRelativeTime}
                />
              ))}
            </div>
          )}
      </>

      {/* ─── Modals ────────────────────────────────────────────── */}
      {isCreateModalOpen && (
        <CreateWordSetModal
          onClose={() => setIsCreateModalOpen(false)}
          onSubmit={handleCreate}
          isSubmitting={isSubmitting}
        />
      )}

      {editingWordSet && (
        <EditWordSetModal
          set={editingWordSet}
          onClose={() => setEditingWordSet(null)}
          onSubmit={handleEdit}
          isSubmitting={isSubmitting}
        />
      )}

      {deletingWordSet && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <div className="bg-[#1e1e2d] border border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl">
            <h3 className="text-xl font-bold text-white mb-2">
              Delete &quot;{deletingWordSet.name}&quot;?
            </h3>
            <p className="text-white/60 text-sm mb-6">
              This will delete the word set and its organization. The original dictionary words will{" "}
              <strong className="text-white">not</strong> be deleted.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setDeletingWordSet(null)}
                className="flex-1 py-3 rounded-xl bg-white/5 hover:bg-white/10 text-white font-bold transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={isSubmitting}
                className="flex-1 py-3 rounded-xl bg-red-500 hover:bg-red-600 disabled:opacity-50 text-white font-bold transition-colors"
              >
                {isSubmitting ? "Deleting..." : "Delete Set"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
