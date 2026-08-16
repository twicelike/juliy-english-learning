"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Sparkles, ArrowRight, Loader2 } from "lucide-react";
import { SpeakingAPI, RecommendedTaskResponse } from "@/lib/speaking-api";

export function RecommendedTask() {
  const router = useRouter();
  const [data, setData] = useState<RecommendedTaskResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isStarting, setIsStarting] = useState(false);

  useEffect(() => {
    SpeakingAPI.getRecommendedTask()
      .then(setData)
      .catch((err) => {
        console.error("Failed to fetch recommended task:", err);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  const handleStart = async () => {
    if (!data) return;
    setIsStarting(true);
    try {
      const session = await SpeakingAPI.startSession(data.task.id);
      router.push(`/speaking/session/${session.id}`);
    } catch (err) {
      console.error(err);
      setIsStarting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="w-full bg-indigo-500/10 border border-indigo-500/20 rounded-3xl p-6 mb-12 flex flex-col justify-center animate-pulse h-[160px]">
        <div className="h-6 w-48 bg-indigo-500/20 rounded-md mb-4"></div>
        <div className="h-4 w-full max-w-md bg-white/5 rounded-md mb-2"></div>
        <div className="h-4 w-64 bg-white/5 rounded-md"></div>
      </div>
    );
  }

  if (!data) {
    return null; // Don't show if failed or no recommendation available
  }

  return (
    <div className="w-full bg-gradient-to-br from-indigo-500/10 via-purple-500/5 to-transparent border border-indigo-500/20 rounded-3xl p-6 md:p-8 mb-12 relative overflow-hidden group">
      {/* Decorative gradient orb */}
      <div className="absolute top-0 right-0 w-64 h-64 bg-indigo-500/20 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 opacity-50 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none"></div>

      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative z-10">
        <div className="flex-1">
          <div className="flex items-center gap-2 text-indigo-400 font-bold text-sm uppercase tracking-wider mb-3">
            <Sparkles size={16} className="animate-pulse" />
            Recommended for you
          </div>
          <h3 className="text-xl md:text-2xl font-bold text-white mb-2 leading-tight">
            {data.task.title || data.task.topic || "Practice Task"}
          </h3>
          <p className="text-white/60 text-sm md:text-base leading-relaxed max-w-2xl mb-4">
            {data.task.prompt}
          </p>
          
          {data.reason && (
            <div className="flex items-center gap-3 mt-4">
              <span className="px-3 py-1 bg-white/5 border border-white/10 rounded-full text-xs text-white/60 flex items-center gap-1.5">
                Targeting: <span className="text-indigo-300 font-medium">{data.reason.skill}</span>
              </span>
              <span className="px-3 py-1 bg-white/5 border border-white/10 rounded-full text-xs text-white/60 flex items-center gap-1.5">
                Focus: <span className="text-purple-300 font-medium capitalize">{data.reason.focus.replace(/_/g, ' ')}</span>
              </span>
            </div>
          )}
        </div>

        <button
          onClick={handleStart}
          disabled={isStarting}
          className="shrink-0 flex items-center justify-center gap-2 px-6 py-3 bg-indigo-500 hover:bg-indigo-600 text-white font-bold rounded-xl transition-all shadow-lg shadow-indigo-500/25 w-full md:w-auto min-w-[140px] disabled:opacity-70 disabled:cursor-not-allowed"
        >
          {isStarting ? (
            <Loader2 size={18} className="animate-spin" />
          ) : (
            <>
              Start Task <ArrowRight size={18} />
            </>
          )}
        </button>
      </div>
    </div>
  );
}
