"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { DictionaryAPI, WordDTO } from "@/lib/dictionary-api";

export default function SearchBar() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<WordDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const router = useRouter();
  const wrapperRef = useRef<HTMLDivElement>(null);

  // Debounce search
  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setLoading(false);
      return;
    }

    setLoading(true);
    const timer = setTimeout(async () => {
      try {
        const data = await DictionaryAPI.search(query);
        setResults(data);
      } catch (err) {
        setResults([]); // Handle error gracefully
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [query]);

  // Click outside to close dropdown
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target as Node)) {
        setShowDropdown(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSelect = (word: string) => {
    setShowDropdown(false);
    setQuery("");
    router.push(`/dictionary/${encodeURIComponent(word)}`);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && query.trim()) {
      handleSelect(query.trim().toLowerCase());
    }
  };

  return (
    <div ref={wrapperRef} className="relative w-full max-w-2xl mx-auto z-50">
      <div className="relative flex items-center">
        <input
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setShowDropdown(true);
          }}
          onKeyDown={handleKeyDown}
          onFocus={() => setShowDropdown(true)}
          placeholder="Tra từ điển (VD: hello, run...)"
          className="w-full h-14 pl-12 pr-4 bg-white/10 border border-white/20 rounded-2xl text-white placeholder:text-white/50 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 backdrop-blur-md transition-all shadow-lg"
          spellCheck={false}
          autoComplete="off"
        />
        <svg
          className="absolute left-4 w-5 h-5 text-white/50"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        {loading && (
          <div className="absolute right-4">
            <div className="w-5 h-5 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
          </div>
        )}
      </div>

      {/* Dropdown */}
      {showDropdown && query.trim() && !loading && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-[#12121c] border border-white/10 rounded-xl shadow-2xl overflow-hidden max-h-[60vh] overflow-y-auto">
          {results.length > 0 ? (
            <ul className="py-2">
              {results.map((w, index) => (
                <li key={`${w.word}-${w.pos}-${index}`}>
                  <button
                    onClick={() => handleSelect(w.word)}
                    className="w-full text-left px-4 py-3 hover:bg-white/5 transition-colors flex items-center justify-between group"
                  >
                    <div>
                      <span className="text-white font-medium group-hover:text-indigo-300">{w.word}</span>
                      {w.pos && <span className="ml-2 text-xs text-white/40 italic">{w.pos}</span>}
                    </div>
                    {w.ipa && <span className="text-xs text-indigo-200/50">{w.ipa}</span>}
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <div className="p-6 text-center text-white/50 text-sm">
              Không tìm thấy từ "{query}"
            </div>
          )}
        </div>
      )}
    </div>
  );
}
