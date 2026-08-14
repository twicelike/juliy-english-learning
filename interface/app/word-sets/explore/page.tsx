"use client";

import { useEffect, useState } from "react";
import { WordSetAPI, WordSetListItem, ExploreWordSetsRequest } from "@/lib/wordset-api";
import { WordSetCard } from "@/components/wordsets/WordSetCard";
import Link from "next/link";
import { useAuthStore } from "@/store/auth-store";
import { useRouter } from "next/navigation";

function formatRelativeTime(dateString: string) {
  const date = new Date(dateString);
  const now = new Date();
  const diff = Math.floor((now.getTime() - date.getTime()) / 1000);
  
  if (diff < 60) return "Just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  if (diff < 86400 * 30) return `${Math.floor(diff / 86400)}d ago`;
  return date.toLocaleDateString();
}

export default function ExploreWordSetsPage() {
  const router = useRouter();
  const { user } = useAuthStore();

  const [sets, setSets] = useState<WordSetListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [query, setQuery] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [topic, setTopic] = useState("");
  const [difficulty, setDifficulty] = useState("");
  const [sort, setSort] = useState("popular");
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  
  const [availableTopics, setAvailableTopics] = useState<string[]>([]);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }
    WordSetAPI.getTopics().then(setAvailableTopics).catch(console.error);
    loadSets(true);
  }, [user, router, searchQuery, topic, difficulty, sort]);

  const loadSets = async (reset = false) => {
    try {
      setLoading(true);
      setError(null);
      
      const currentPage = reset ? 1 : page + 1;
      const params: ExploreWordSetsRequest = {
        q: searchQuery,
        topic,
        difficulty,
        sort,
        page: currentPage,
        limit: 20
      };

      const res = await WordSetAPI.exploreWordSets(params);
      
      if (reset) {
        setSets(res.items);
      } else {
        setSets(prev => [...prev, ...res.items]);
      }
      
      setPage(res.page);
      setHasMore(res.has_more);
    } catch (err) {
      console.error(err);
      setError("Failed to load word sets");
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearchQuery(query);
  };

  if (!user) return null;

  return (
    <div className="flex flex-col h-screen bg-[#0f0f17] overflow-hidden">
      <header className="shrink-0 px-6 py-4 flex items-center justify-between border-b border-white/5 bg-[#12121c] z-10">
        <div className="flex items-center gap-4">
          <Link
            href="/word-sets"
            className="text-white/40 hover:text-white transition-colors text-sm font-medium flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            My Library
          </Link>
          <div className="h-6 w-px bg-white/10" />
          <div>
            <h1 className="text-white font-bold text-lg leading-tight">Explore</h1>
            <span className="text-xs text-white/40">Discover vocabulary from the community</span>
          </div>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-6">
        <div className="max-w-6xl mx-auto space-y-6">
          
          <form onSubmit={handleSearch} className="flex gap-2">
            <div className="relative flex-1">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg className="w-5 h-5 text-white/30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </div>
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search word sets..."
                className="w-full bg-[#1e1e2d] border border-white/10 rounded-xl pl-10 pr-4 py-3 text-white focus:outline-none focus:border-indigo-500 transition-colors"
              />
            </div>
            <button
              type="submit"
              className="px-6 py-3 bg-indigo-500 hover:bg-indigo-600 text-white font-semibold rounded-xl transition-colors"
            >
              Search
            </button>
          </form>

          <div className="flex gap-4">
            <select
              value={topic}
              onChange={e => setTopic(e.target.value)}
              className="bg-[#1e1e2d] border border-white/10 rounded-xl px-4 py-2.5 text-white text-sm focus:outline-none focus:border-indigo-500 transition-colors appearance-none pr-8 relative"
            >
              <option value="">All Topics</option>
              {availableTopics.map(t => <option key={t} value={t}>{t}</option>)}
            </select>

            <select
              value={difficulty}
              onChange={e => setDifficulty(e.target.value)}
              className="bg-[#1e1e2d] border border-white/10 rounded-xl px-4 py-2.5 text-white text-sm focus:outline-none focus:border-indigo-500 transition-colors appearance-none pr-8 relative"
            >
              <option value="">All Difficulties</option>
              <option value="Beginner">Beginner</option>
              <option value="Intermediate">Intermediate</option>
              <option value="Advanced">Advanced</option>
            </select>

            <select
              value={sort}
              onChange={e => setSort(e.target.value)}
              className="bg-[#1e1e2d] border border-white/10 rounded-xl px-4 py-2.5 text-white text-sm focus:outline-none focus:border-indigo-500 transition-colors appearance-none pr-8 relative ml-auto"
            >
              <option value="popular">Popular</option>
              <option value="newest">Newest</option>
            </select>
          </div>

          {loading && sets.length === 0 ? (
            <div className="py-20 flex justify-center">
              <div className="w-8 h-8 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin" />
            </div>
          ) : error ? (
            <div className="py-20 text-center text-red-400">
              {error}
            </div>
          ) : sets.length === 0 ? (
            <div className="py-20 text-center flex flex-col items-center">
              <svg className="w-16 h-16 text-white/10 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
              <p className="text-white/40">No word sets found.</p>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {sets.map(set => (
                  <WordSetCard
                    key={set.id}
                    set={set}
                    formatRelativeTime={formatRelativeTime}
                    mode="explore"
                  />
                ))}
              </div>
              
              {hasMore && (
                <div className="pt-8 pb-4 flex justify-center">
                  <button
                    onClick={() => loadSets(false)}
                    disabled={loading}
                    className="px-6 py-2.5 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-xl transition-colors text-sm font-semibold disabled:opacity-50"
                  >
                    {loading ? "Loading..." : "Load More"}
                  </button>
                </div>
              )}
            </>
          )}

        </div>
      </div>
    </div>
  );
}
