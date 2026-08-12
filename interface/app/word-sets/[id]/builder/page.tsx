"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { WordSetAPI, WordSet } from "@/lib/wordset-api";
import { DictionaryAPI, WordDTO } from "@/lib/dictionary-api";
import Link from "next/link";

export default function WordSetBuilderPage() {
  const { id } = useParams() as { id: string };
  const router = useRouter();

  const [set, setSet] = useState<WordSet | null>(null);
  const [loading, setLoading] = useState(true);

  const [query, setQuery] = useState("");
  const [searchResults, setSearchResults] = useState<WordDTO[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const [selectedWordIds, setSelectedWordIds] = useState<Set<string>>(new Set());
  
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadSet();
  }, [id]);

  const loadSet = async () => {
    try {
      setLoading(true);
      const res = await WordSetAPI.getWordSet(id);
      setSet(res);
    } catch (err) {
      console.error(err);
      router.push("/word-sets");
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    try {
      setIsSearching(true);
      const results = await DictionaryAPI.search(query);
      setSearchResults(results);
    } catch (err) {
      console.error(err);
    } finally {
      setIsSearching(false);
    }
  };

  const toggleWord = (wordId: string) => {
    setSelectedWordIds(prev => {
      const next = new Set(prev);
      if (next.has(wordId)) {
        next.delete(wordId);
      } else {
        next.add(wordId);
      }
      return next;
    });
  };

  const handleSave = async () => {
    if (selectedWordIds.size === 0) return;
    if (!set) return;

    setIsSaving(true);
    try {
      const res = await WordSetAPI.bulkAddWords(set.id, Array.from(selectedWordIds));
      alert(`Added ${res.added} words! Word set now has ${res.word_count}/${res.limit} words.`);
      setSelectedWordIds(new Set()); // clear selection
      loadSet(); // reload count
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to add words");
    } finally {
      setIsSaving(false);
    }
  };

  if (loading) {
    return <div className="p-10 text-white/50 text-center animate-pulse">Loading builder...</div>;
  }

  if (!set) return null;

  return (
    <div className="max-w-4xl mx-auto py-10 px-6">
      <Link href="/word-sets" className="text-indigo-400 hover:text-indigo-300 flex items-center gap-2 mb-8 w-max">
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" /></svg>
        Back to Word Sets
      </Link>

      <div className="bg-[#1e1e2d] border border-white/10 rounded-3xl p-8 shadow-2xl">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">Word Set Builder</h1>
            <p className="text-white/50">Add words to <span className="text-white font-medium">{set.name}</span></p>
          </div>
          <div className="text-right">
            <div className="text-white/40 text-sm uppercase tracking-wider mb-1">Current Size</div>
            <div className="text-2xl font-mono text-indigo-400">{set.word_count} / 50</div>
          </div>
        </div>

        <form onSubmit={handleSearch} className="relative mb-8">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search dictionary..."
            className="w-full bg-black/20 border border-white/10 rounded-2xl py-4 pl-14 pr-6 text-white placeholder-white/30 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all text-lg"
          />
          <svg className="w-6 h-6 text-white/30 absolute left-5 top-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
          <button type="submit" className="hidden">Search</button>
        </form>

        <div className="space-y-3 min-h-[300px]">
          {isSearching ? (
             <div className="flex justify-center items-center h-40">
               <div className="w-8 h-8 border-2 border-white/20 border-t-indigo-500 rounded-full animate-spin"></div>
             </div>
          ) : searchResults.length === 0 ? (
            <div className="text-center text-white/30 py-10">
              {query ? "No words found in dictionary." : "Search to find words to add."}
            </div>
          ) : (
            searchResults.map(word => {
              const isSelected = selectedWordIds.has(word.id);
              return (
                <div key={word.id} className={`flex items-center justify-between p-4 rounded-xl border transition-all ${isSelected ? "bg-indigo-500/10 border-indigo-500/50" : "bg-white/5 border-white/5 hover:bg-white/10"}`}>
                  <div>
                    <div className="text-white font-medium text-lg">{word.word}</div>
                    <div className="text-white/40 text-sm mt-1">{word.pos} • {word.definitions?.[0]?.definition}</div>
                  </div>
                  <button
                    onClick={() => toggleWord(word.id)}
                    className={`w-10 h-10 rounded-full flex items-center justify-center transition-all ${isSelected ? "bg-indigo-500 text-white" : "bg-black/40 text-white/50 hover:bg-white/10"}`}
                  >
                    {isSelected ? (
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" /></svg>
                    ) : (
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg>
                    )}
                  </button>
                </div>
              );
            })
          )}
        </div>

        {/* Floating Action Bar */}
        <div className="mt-8 flex items-center justify-between bg-black/40 p-4 rounded-2xl border border-white/10">
          <div className="text-white/70">
            Selected: <span className="text-white font-bold">{selectedWordIds.size}</span> words
          </div>
          <button
            onClick={handleSave}
            disabled={selectedWordIds.size === 0 || isSaving}
            className={`px-8 py-3 rounded-xl font-medium transition-all ${
              selectedWordIds.size > 0 && !isSaving
                ? "bg-indigo-500 text-white hover:bg-indigo-600 shadow-lg shadow-indigo-500/30" 
                : "bg-white/5 text-white/30 cursor-not-allowed"
            }`}
          >
            {isSaving ? "Saving..." : "Save Selection"}
          </button>
        </div>
      </div>
    </div>
  );
}
