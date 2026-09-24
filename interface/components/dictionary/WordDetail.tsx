"use client";

import { useEffect, useState } from "react";
import { WordDTO } from "@/lib/dictionary-api";
import { AuthAPI } from "@/lib/auth-api";
import { useAuthStore } from "@/features/auth";
import { useRouter } from "next/navigation";
import AITutor from "@/components/dictionary/AITutor";
import SenseExplanation from "@/components/dictionary/SenseExplanation";
import AudioButton from "./AudioButton";
import { AddToSetModal } from "@/components/wordsets/AddToSetModal";

const getPosColor = (pos: string | undefined) => {
  if (!pos) return { bg: "bg-indigo-500/20", text: "text-indigo-300", border: "border-indigo-500/30", bar: "bg-indigo-500" };
  const p = pos.toLowerCase();
  if (p.includes("noun")) return { bg: "bg-blue-500/20", text: "text-blue-300", border: "border-blue-500/30", bar: "bg-blue-500" };
  if (p.includes("verb")) return { bg: "bg-rose-500/20", text: "text-rose-300", border: "border-rose-500/30", bar: "bg-rose-500" };
  if (p.includes("adj")) return { bg: "bg-emerald-500/20", text: "text-emerald-300", border: "border-emerald-500/30", bar: "bg-emerald-500" };
  if (p.includes("adv")) return { bg: "bg-amber-500/20", text: "text-amber-300", border: "border-amber-500/30", bar: "bg-amber-500" };
  if (p.includes("prep")) return { bg: "bg-purple-500/20", text: "text-purple-300", border: "border-purple-500/30", bar: "bg-purple-500" };
  if (p.includes("conj")) return { bg: "bg-pink-500/20", text: "text-pink-300", border: "border-pink-500/30", bar: "bg-pink-500" };
  return { bg: "bg-indigo-500/20", text: "text-indigo-300", border: "border-indigo-500/30", bar: "bg-indigo-500" };
};

interface WordDetailProps {
  word: WordDTO;
}

export default function WordDetail({ word }: WordDetailProps) {
  const [showAll, setShowAll] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [isBookmarked, setIsBookmarked] = useState(false);
  const [loadingBookmark, setLoadingBookmark] = useState(false);

  const { user } = useAuthStore();
  const router = useRouter();

  // Check if already bookmarked
  useEffect(() => {
    if (user) {
      AuthAPI.getBookmarks()
        .then((bookmarks) => {
          setIsBookmarked(bookmarks.some((b) => b.word === word.word));
        })
        .catch(() => { });
    }
  }, [user, word.word]);

  const toggleBookmark = async () => {
    if (!user) {
      router.push("/login");
      return;
    }

    setLoadingBookmark(true);
    try {
      if (isBookmarked) {
        await AuthAPI.removeBookmark(word.word);
        setIsBookmarked(false);
      } else {
        await AuthAPI.addBookmark(word.word);
        setIsBookmarked(true);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingBookmark(false);
    }
  };

  const initialDefCount = 3;
  const hasMore = word.definitions.length > initialDefCount;
  const displayedDefs = showAll ? word.definitions : word.definitions.slice(0, initialDefCount);

  // Group definitions by POS
  const groupedDefs = displayedDefs.reduce((acc, def) => {
    const pos = def.partOfSpeech || 'other';
    if (!acc[pos]) acc[pos] = [];
    acc[pos].push(def);
    return acc;
  }, {} as Record<string, typeof displayedDefs>);

  return (
    <div className="w-full max-w-3xl mx-auto p-6 md:p-10 bg-white/5 border border-white/10 rounded-3xl backdrop-blur-md shadow-2xl">
      {/* Header: Word, POS, IPA */}
      <div className="mb-10 text-center md:text-left">
        <h1 className="text-4xl md:text-6xl font-bold text-white tracking-tight mb-4">
          {word.word}
        </h1>
        <div className="flex flex-wrap items-center justify-center md:justify-start gap-3">
          {word.pos && word.pos.split(', ').map((p, idx) => {
            const colors = getPosColor(p);
            return (
              <span key={idx} className={`px-3 py-1 rounded-full ${colors.bg} ${colors.text} text-sm font-medium border ${colors.border} uppercase tracking-wider`}>
                {p}
              </span>
            );
          })}
          {word.ipa && (
            <span className="text-white/60 font-mono text-lg flex items-center gap-2">
              {word.ipa}
              {word.audio && <AudioButton audioUrl={word.audio} />}
            </span>
          )}
          {!word.ipa && word.audio && <AudioButton audioUrl={word.audio} />}

          <button
            onClick={toggleBookmark}
            disabled={loadingBookmark}
            className={`ml-4 p-2 rounded-full border transition-all ${isBookmarked
              ? "bg-indigo-500/20 border-indigo-500/50 text-indigo-400"
              : "bg-white/5 border-white/10 text-white/40 hover:bg-white/10 hover:text-white"
              }`}
            title={isBookmarked ? "Remove from saved words" : "Save word"}
          >
            <svg className="w-5 h-5" fill={isBookmarked ? "currentColor" : "none"} stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={isBookmarked ? 0 : 2} d={isBookmarked ? "M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" : "M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"} />
            </svg>
          </button>
          <button
            onClick={() => {
              if (!user) router.push("/login");
              else setShowAddModal(true);
            }}
            className="ml-2 px-4 py-2 rounded-full border border-white/10 bg-white/5 text-white/70 hover:bg-white/10 hover:text-white transition-all text-sm font-medium flex items-center gap-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add to Set
          </button>
        </div>
      </div>

      {showAddModal && <AddToSetModal word={word} onClose={() => setShowAddModal(false)} />}

      {/* Forms */}
      {word.forms.length > 0 && (
        <div className="mb-7 p-3 rounded-2xl bg-black/20 border border-white/5">
          <h3 className="text-white/40 uppercase tracking-widest mb-2 font-semibold text-xs">Other Forms</h3>
          <div className="flex flex-wrap gap-2">
            {word.forms.map((f, i) => (
              <span key={i} className="px-3 py-1.5 rounded-lg bg-white/5 text-white/80 text-sm">
                {f}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Meanings */}
      <div className="mb-10">

        <div className="space-y-10">
          {Object.entries(groupedDefs).map(([pos, defs]) => {
            const colors = getPosColor(pos);
            return (
              <div key={pos} className="space-y-6">
                <h3 className={`text-sm uppercase tracking-widest font-bold flex items-center gap-3 ${colors.text}`}>
                  <span className={`w-8 h-[2px] ${colors.bar}`}></span>
                  {pos}
                </h3>

                <div className="space-y-8 pl-4 md:pl-8">
                  {defs.map((def, idx) => (
                    <div key={idx} className="group">
                      <p className="text-white/90 text-lg leading-relaxed mb-3 flex items-start">
                        <span className={`font-bold mr-3 mt-0.5 ${colors.text}`}>{idx + 1}.</span>
                        <span>{def.definition}</span>
                      </p>

                      {def.examples.length > 0 && (
                        <div className={`ml-6 pl-4 border-l-2 ${colors.border} space-y-2 mt-3`}>
                          {def.examples.map((ex, i) => (
                            <p key={i} className="text-white/50 text-base italic">
                              "{ex}"
                            </p>
                          ))}
                        </div>
                      )}

                      <div className="ml-6 mt-2">
                        <SenseExplanation wordId={def.word_id || word.id} senseIndex={def.sense_index ?? 0} />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>

        {/* View more button */}
        {hasMore && (
          <button
            onClick={() => setShowAll(!showAll)}
            className="mt-8 px-6 py-2.5 rounded-full border border-indigo-500/30 text-indigo-300 hover:bg-indigo-500/10 transition-colors font-medium text-sm flex items-center gap-2 mx-auto md:mx-0"
          >
            {showAll ? "Thu gọn" : `Xem thêm ${word.definitions.length - initialDefCount} nghĩa khác`}
            <svg
              className={`w-4 h-4 transition-transform ${showAll ? "rotate-180" : ""}`}
              fill="none" stroke="currentColor" viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        )}
      </div>

      {/* Antonyms */}
      {word.antonyms.length > 0 && (
        <div>
          <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
            <span className="w-6 h-1 bg-violet-500 rounded-full"></span>
            Antonyms
          </h2>
          <div className="flex flex-wrap gap-2">
            {word.antonyms.map((ant, i) => (
              <span key={i} className="px-3 py-1.5 rounded-lg bg-violet-500/10 text-violet-300 text-sm border border-violet-500/20">
                {ant}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* AI Tutor Section — visible only when logged in */}
      {user && <AITutor word={word.word} pos={word.pos} />}
    </div>
  );
}
