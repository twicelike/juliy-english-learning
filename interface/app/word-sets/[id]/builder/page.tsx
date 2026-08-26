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

  const [selectedItems, setSelectedItems] = useState<Map<string, {word_id: string, sense_id: string}>>(new Map());
  const [existingItems, setExistingItems] = useState<Set<string>>(new Set());
  
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadSet();
  }, [id]);

  const loadSet = async () => {
    try {
      setLoading(true);
      const res = await WordSetAPI.getWordSet(id);
      setSet(res);
      const existing = new Set<string>();
      if (res.saved_senses) {
        res.saved_senses.forEach((s) => {
          existing.add(`${s.word_id}:${s.sense_id}`);
        });
      }
      setExistingItems(existing);
    } catch (err) {
      console.error(err);
      router.push("/word-sets");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const q = query.trim();
    if (q.length === 0) {
      setSearchResults([]);
      setIsSearching(false);
      return;
    }
    if (q.length < 2) {
      return;
    }

    const delayDebounceFn = setTimeout(async () => {
      try {
        setIsSearching(true);
        const results = await DictionaryAPI.search(q);
        setSearchResults(results);
      } catch (err) {
        console.error(err);
      } finally {
        setIsSearching(false);
      }
    }, 400);

    return () => clearTimeout(delayDebounceFn);
  }, [query]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
  };

  const toggleItem = (wordId: string, senseIndex: string) => {
    setSelectedItems(prev => {
      const next = new Map(prev);
      const key = `${wordId}:${senseIndex}`;
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.set(key, { word_id: wordId, sense_id: senseIndex });
      }
      return next;
    });
  };

  const handleSave = async () => {
    if (selectedItems.size === 0) return;
    if (!set) return;

    setIsSaving(true);
    try {
      const items = Array.from(selectedItems.values());
      const res = await WordSetAPI.bulkAddWords(set.id, items);
      alert(`Added ${res.added} words! Word set now has ${res.word_count}/${res.limit} words.`);
      setSelectedItems(new Map()); // clear selection
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
              return (
                <div key={word.id} className="bg-white/5 border border-white/5 rounded-xl overflow-hidden">
                  <div className="p-4 bg-black/20 border-b border-white/5 flex items-center justify-between">
                    <div>
                      <div className="text-white font-medium text-lg">{word.word}</div>
                      <div className="text-white/40 text-sm mt-1">{word.pos} • {word.ipa}</div>
                    </div>
                  </div>
                  <div className="p-2 space-y-1">
                    {word.definitions
                      ?.filter((def, idx, arr) => arr.findIndex(d => d.sense_id === def.sense_id && (d.word_id || word.id) === (def.word_id || word.id)) === idx)
                      .map((def) => {
                      const key = `${def.word_id || word.id}:${def.sense_id}`;
                      const isSelected = selectedItems.has(key);
                      const isExisting = existingItems.has(key);
                      return (
                        <div 
                          key={key}
                          onClick={() => {
                            if (!isExisting) toggleItem(def.word_id || word.id, def.sense_id);
                          }}
                          className={`flex items-start gap-3 p-3 rounded-lg transition-colors ${
                            isExisting ? "opacity-50 cursor-not-allowed" :
                            isSelected ? "bg-indigo-500/20 cursor-pointer" : "hover:bg-white/5 cursor-pointer"
                          }`}
                        >
                          <div className={`mt-0.5 w-5 h-5 rounded flex-shrink-0 border flex items-center justify-center transition-colors ${
                            isExisting ? "border-white/20 bg-white/10" :
                            isSelected ? "bg-indigo-500 border-indigo-500" : "border-white/30"
                          }`}>
                            {(isSelected || isExisting) && <svg className="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              {(def.partOfSpeech || word.pos) && <span className="text-[10px] uppercase font-bold text-indigo-300 bg-indigo-500/20 px-1.5 py-0.5 rounded">{def.partOfSpeech || word.pos}</span>}
                              {isExisting && <span className="text-[10px] uppercase font-bold text-green-400 bg-green-500/20 px-1.5 py-0.5 rounded">Already in set</span>}
                            </div>
                            <div className={`text-sm ${isSelected ? "text-indigo-100" : "text-white/70"}`}>
                              {def.definition}
                            </div>
                            {def.examples?.[0] && (
                              <div className="text-xs text-white/30 mt-1 italic">&quot;{def.examples[0]}&quot;</div>
                            )}
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Floating Action Bar */}
        <div className="mt-8 flex items-center justify-between bg-black/40 p-4 rounded-2xl border border-white/10">
          <div className="text-white/70">
            Selected: <span className="text-white font-bold">{selectedItems.size}</span> words
          </div>
          <button
            onClick={handleSave}
            disabled={selectedItems.size === 0 || isSaving}
            className={`px-8 py-3 rounded-xl font-medium transition-all ${
              selectedItems.size > 0 && !isSaving
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
