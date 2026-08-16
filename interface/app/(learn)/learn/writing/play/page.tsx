"use client";

import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/features/auth";
import { WritingAPI, MissionDTO, EvaluationResponseDTO } from "@/lib/writing-api";
import { UserAPI } from "@/lib/user-api";
import Link from "next/link";
import { ArrowLeft, Send, Check, Bookmark, Target, Loader2, Sparkles, AlertCircle, Info } from "lucide-react";

function WritingPlayContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const mode = searchParams.get("mode") || "new"; // "new" or "review"
  const difficulty = searchParams.get("difficulty") || undefined;
  
  const { user } = useAuthStore();
  
  const [mission, setMission] = useState<MissionDTO | null>(null);
  const [loadingMission, setLoadingMission] = useState(true);
  
  const [content, setContent] = useState("");
  const [evaluating, setEvaluating] = useState(false);
  const [result, setResult] = useState<EvaluationResponseDTO | null>(null);
  
  const [savingExp, setSavingExp] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (!user) return;
    loadMission();
  }, [user, mode, difficulty]);

  const loadMission = async () => {
    setLoadingMission(true);
    try {
      const apiDifficulty = difficulty === 'Random' ? undefined : difficulty;
      if (mode === "review") {
        const data = await WritingAPI.getReviewMission(undefined, apiDifficulty);
        setMission(data);
      } else {
        const data = await WritingAPI.getNewMission(undefined, apiDifficulty);
        setMission(data);
      }
    } catch (err) {
      console.error(err);
      // fallback or handle error
    } finally {
      setLoadingMission(false);
    }
  };

  const handleSubmit = async () => {
    if (!mission || content.trim().length < 3) return;
    setEvaluating(true);
    try {
      const data = await WritingAPI.submitMission(mission.id, content);
      setResult(data);
    } catch (err) {
      console.error(err);
    } finally {
      setEvaluating(false);
    }
  };

  const handleSaveExpression = async (expression: string) => {
    setSavingExp(true);
    try {
      await UserAPI.addBookmark({ word: expression, type: "expression" });
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      console.error(err);
    } finally {
      setSavingExp(false);
    }
  };

  if (!user || loadingMission) {
    return (
      <div className="flex justify-center items-center h-[50vh]">
        <Loader2 className="w-10 h-10 text-indigo-500 animate-spin" />
      </div>
    );
  }

  if (!mission) {
    return (
      <div className="text-center p-12 text-white/50">
        No missions available right now. Please try again later.
      </div>
    );
  }

  // Determine score color
  const getScoreColor = (s: number) => {
    if (s >= 90) return "text-emerald-400 border-emerald-400";
    if (s >= 70) return "text-amber-400 border-amber-400";
    return "text-red-400 border-red-400";
  };
  
  const getScoreBg = (s: number) => {
    if (s >= 90) return "bg-emerald-400/10";
    if (s >= 70) return "bg-amber-400/10";
    return "bg-red-400/10";
  };

  return (
    <div className="max-w-4xl mx-auto p-6 min-h-[calc(100vh-4rem)]">
      <Link href="/learn/writing" className="inline-flex items-center gap-2 text-white/40 hover:text-white mb-8 transition-colors">
        <ArrowLeft className="w-4 h-4" /> Back to Dashboard
      </Link>

      {!result && !evaluating && (
        <div className="space-y-6">
          <div className="bg-white/5 border border-white/10 p-8 rounded-3xl">
            <div className="flex items-center gap-2 text-indigo-400 text-sm font-semibold uppercase tracking-wider mb-4">
              <Target className="w-4 h-4" /> {mission.topic} • {mission.scenario}
            </div>
            <h1 className="text-2xl md:text-3xl font-bold text-white mb-6 leading-relaxed">
              {mission.prompt_vi?.join(" ")}
            </h1>
            
            <div className="mb-2">
              <span className="text-white/40 text-sm font-medium">Target Vocabulary:</span>
            </div>
            <div className="flex flex-wrap gap-2 mb-8">
              {mission.target_vocabulary.map((tv, idx) => (
                <span key={idx} className="px-3 py-1 bg-white/10 text-white/80 rounded-full text-sm">
                  {tv}
                </span>
              ))}
            </div>

            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="Translate to English here..."
              rows={6}
              className="w-full bg-[#0B0C10] border border-white/10 rounded-2xl p-6 text-white text-lg resize-none focus:outline-none focus:border-indigo-500 transition-colors shadow-inner"
            ></textarea>
            
            <div className="flex justify-end mt-6">
              <button 
                onClick={handleSubmit}
                disabled={content.trim().length < 5}
                className="px-8 py-4 bg-indigo-500 hover:bg-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-2xl flex items-center gap-2 transition-all shadow-lg shadow-indigo-500/30"
              >
                Submit Translation <Send className="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {evaluating && (
        <div className="flex flex-col items-center justify-center py-20">
          <div className="relative">
             <div className="w-24 h-24 border-4 border-indigo-500/20 rounded-full"></div>
             <div className="absolute top-0 left-0 w-24 h-24 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
             <Sparkles className="w-8 h-8 text-indigo-400 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 animate-pulse" />
          </div>
          <h2 className="mt-8 text-xl font-medium text-white">Evaluating your translation...</h2>
          <p className="text-white/40 mt-2">Checking grammar, vocabulary, and naturalness</p>
        </div>
      )}

      {result && (
        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
          {/* Main Score Header */}
          <div className="flex flex-col md:flex-row items-center gap-8 bg-white/5 border border-white/10 p-8 rounded-3xl">
            <div className={`w-32 h-32 flex-shrink-0 rounded-full border-4 flex flex-col items-center justify-center shadow-lg ${getScoreColor(result.score)} ${getScoreBg(result.score)}`}>
               <span className="text-4xl font-black">{result.score}</span>
               <span className="text-xs font-bold uppercase opacity-80">Score</span>
            </div>
            <div className="flex-1 text-center md:text-left">
               <h2 className="text-2xl font-bold text-white mb-2">Evaluation Complete</h2>
               <p className="text-white/60 mb-4">{result.detailed_feedback}</p>
               <button 
                 onClick={() => {
                   setResult(null);
                   setContent("");
                   loadMission();
                 }}
                 className="px-6 py-2 bg-white/10 hover:bg-white/20 text-white rounded-xl transition-colors font-medium text-sm"
               >
                 Next Mission
               </button>
            </div>
          </div>

          {/* Metrics Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
             {[
               { label: "Grammar", desc: "Evaluates correct usage of the target grammar and checks for any structural, tense, or preposition errors.", data: result.breakdown.grammar },
               { label: "Vocabulary", desc: "Evaluates the precise and appropriate use of the required target vocabulary words in your sentence.", data: result.breakdown.vocabulary },
               { label: "Naturalness", desc: "Assesses how native-like and contextually appropriate your phrasing is, penalizing awkward word-for-word translations.", data: result.breakdown.naturalness },
               { label: "Completeness", desc: "Verifies that 100% of the original Vietnamese meaning is conveyed without missing any details or nuances.", data: result.breakdown.completeness }
             ].map((metric, idx) => (
               <div key={idx} className="bg-white/5 border border-white/10 p-5 rounded-2xl flex gap-4 relative">
                  <div className={`w-12 h-12 flex-shrink-0 rounded-xl flex items-center justify-center font-bold text-lg border ${getScoreColor(metric.data.score)} ${getScoreBg(metric.data.score)}`}>
                    {metric.data.score}
                  </div>
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="text-white/90 font-bold">{metric.label}</h3>
                      <div className="group/tooltip relative flex items-center">
                         <Info className="w-3.5 h-3.5 text-white/30 hover:text-white/70 cursor-help transition-colors" />
                         <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 w-64 p-3 bg-[#1a1b23] border border-white/10 rounded-xl text-xs text-white/80 opacity-0 group-hover/tooltip:opacity-100 transition-opacity pointer-events-none z-10 text-center shadow-2xl leading-relaxed">
                            {metric.desc}
                            {/* Little triangle for tooltip */}
                            <div className="absolute top-full left-1/2 -translate-x-1/2 -mt-[1px] border-4 border-transparent border-t-[#1a1b23]"></div>
                         </div>
                      </div>
                    </div>
                    <p className="text-white/50 text-sm leading-relaxed">{metric.data.feedback}</p>
                  </div>
               </div>
             ))}
          </div>

          {/* Corrected Text Box */}
          <div className="bg-gradient-to-br from-indigo-500/10 to-emerald-500/10 border border-indigo-500/20 p-8 rounded-3xl relative overflow-hidden">
             <div className="absolute top-0 left-0 w-1 h-full bg-indigo-500"></div>
             <h3 className="text-indigo-300 font-bold uppercase tracking-wider text-sm mb-4">Recommended Translation</h3>
             <p className="text-xl text-white leading-relaxed mb-6">
               {result.corrected_version}
             </p>
             <button
                onClick={() => handleSaveExpression(result.corrected_version)}
                disabled={savingExp || saved}
                className={`px-5 py-2.5 rounded-xl font-medium flex items-center gap-2 transition-all ${
                  saved 
                    ? "bg-emerald-500 text-white shadow-lg shadow-emerald-500/20" 
                    : "bg-white/10 hover:bg-white/20 text-white"
                }`}
             >
                {saved ? (
                  <><Check className="w-4 h-4" /> Saved to Bookmarks!</>
                ) : (
                  <><Bookmark className="w-4 h-4" /> Save Expression</>
                )}
             </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default function WritingPlayPage() {
  return (
    <Suspense fallback={<div className="flex justify-center items-center h-[50vh]"><Loader2 className="w-10 h-10 text-indigo-500 animate-spin" /></div>}>
      <WritingPlayContent />
    </Suspense>
  );
}
