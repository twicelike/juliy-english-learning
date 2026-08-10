"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";
import { WritingAPI, DashboardStatsDTO } from "@/lib/writing-api";
import Link from "next/link";
import { BookOpen, Target, BrainCircuit, PenTool, ArrowRight } from "lucide-react";

export default function WritingDashboardPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [stats, setStats] = useState<DashboardStatsDTO | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }
    loadStats();
  }, [user, router]);

  const loadStats = async () => {
    try {
      const data = await WritingAPI.getDashboardStats();
      setStats(data);
    } catch (error) {
      console.error("Failed to load dashboard stats", error);
    } finally {
      setLoading(false);
    }
  };

  if (!user || loading) {
    return (
      <div className="flex justify-center items-center h-[50vh]">
        <div className="w-8 h-8 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const averageScore = stats?.average_score ? Math.round(stats.average_score) : 0;
  const totalCompleted = stats?.total_completed || 0;
  
  // Mastery distribution math for circular chart (simplified CSS conical gradient)
  const totalM = (stats?.mastery_mastered || 0) + (stats?.mastery_familiar || 0) + (stats?.mastery_beginner || 0);
  const masteredPct = totalM ? ((stats?.mastery_mastered || 0) / totalM) * 100 : 0;
  const familiarPct = totalM ? ((stats?.mastery_familiar || 0) / totalM) * 100 : 0;

  return (
    <div className="min-h-[calc(100vh-4rem)] p-6 max-w-5xl mx-auto">
      <div className="mb-10 text-center space-y-4">
        <h1 className="text-4xl md:text-5xl font-black bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 via-purple-400 to-emerald-400">
          AI Writing Missions
        </h1>
        <p className="text-white/60 text-lg max-w-2xl mx-auto">
          Translate real-world scenarios from Vietnamese to English. Get instant feedback on grammar, vocabulary, and naturalness.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
        {/* Average Score Card */}
        <div className="bg-white/5 border border-white/10 rounded-3xl p-6 relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
          <div className="flex items-center gap-4 mb-4">
            <div className="p-3 bg-indigo-500/20 rounded-2xl">
              <Target className="w-6 h-6 text-indigo-400" />
            </div>
            <h2 className="text-white/80 font-medium">Average Score</h2>
          </div>
          <div className="flex items-end gap-2">
            <span className="text-5xl font-black text-white">{averageScore}</span>
            <span className="text-white/40 font-medium mb-1">/100</span>
          </div>
        </div>

        {/* Total Missions */}
        <div className="bg-white/5 border border-white/10 rounded-3xl p-6 relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
          <div className="flex items-center gap-4 mb-4">
            <div className="p-3 bg-emerald-500/20 rounded-2xl">
              <PenTool className="w-6 h-6 text-emerald-400" />
            </div>
            <h2 className="text-white/80 font-medium">Total Completed</h2>
          </div>
          <div className="flex items-end gap-2">
            <span className="text-5xl font-black text-white">{totalCompleted}</span>
            <span className="text-white/40 font-medium mb-1">missions</span>
          </div>
        </div>

        {/* Mastery Levels */}
        <div className="bg-white/5 border border-white/10 rounded-3xl p-6 relative overflow-hidden group flex items-center gap-6">
           <div className="flex-1 space-y-3">
             <h2 className="text-white/80 font-medium mb-1">Mastery Levels</h2>
             <div className="flex items-center justify-between text-sm">
               <div className="flex items-center gap-2">
                 <div className="w-3 h-3 rounded-full bg-emerald-400"></div>
                 <span className="text-white/70">Mastered</span>
               </div>
               <span className="text-white font-bold">{stats?.mastery_mastered || 0}</span>
             </div>
             <div className="flex items-center justify-between text-sm">
               <div className="flex items-center gap-2">
                 <div className="w-3 h-3 rounded-full bg-amber-400"></div>
                 <span className="text-white/70">Familiar</span>
               </div>
               <span className="text-white font-bold">{stats?.mastery_familiar || 0}</span>
             </div>
             <div className="flex items-center justify-between text-sm">
               <div className="flex items-center gap-2">
                 <div className="w-3 h-3 rounded-full bg-red-400"></div>
                 <span className="text-white/70">Beginner</span>
               </div>
               <span className="text-white font-bold">{stats?.mastery_beginner || 0}</span>
             </div>
           </div>
           
           <div className="w-24 h-24 rounded-full flex-shrink-0 relative flex items-center justify-center bg-white/5">
              {totalM > 0 ? (
                <div 
                  className="absolute inset-0 rounded-full" 
                  style={{
                    background: `conic-gradient(#34d399 0% ${masteredPct}%, #fbbf24 ${masteredPct}% ${masteredPct + familiarPct}%, #f87171 ${masteredPct + familiarPct}% 100%)`
                  }}
                >
                  <div className="absolute inset-2 bg-[#0B0C10] rounded-full"></div>
                </div>
              ) : (
                <div className="absolute inset-0 rounded-full border-4 border-white/5"></div>
              )}
              <BrainCircuit className="w-8 h-8 text-white/30 relative z-10" />
           </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Link 
          href="/learn/writing/play?mode=new"
          className="group relative flex flex-col items-center justify-center p-12 bg-indigo-500/10 border border-indigo-500/20 hover:border-indigo-500/50 hover:bg-indigo-500/20 rounded-3xl transition-all overflow-hidden"
        >
          <div className="absolute inset-0 bg-gradient-to-b from-indigo-500/5 to-transparent"></div>
          <div className="w-16 h-16 bg-indigo-500 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-indigo-500/30 group-hover:scale-110 transition-transform">
            <BookOpen className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-bold text-white mb-2">Start New Mission</h2>
          <p className="text-indigo-200/60 text-center max-w-sm">Translate a completely new scenario and learn new vocabulary.</p>
        </Link>

        <Link 
          href="/learn/writing/play?mode=review"
          className="group relative flex flex-col items-center justify-center p-12 bg-white/5 border border-white/10 hover:border-white/30 hover:bg-white/10 rounded-3xl transition-all overflow-hidden"
        >
          <div className="absolute inset-0 bg-gradient-to-b from-white/5 to-transparent"></div>
          <div className="w-16 h-16 bg-white/10 border border-white/20 rounded-2xl flex items-center justify-center mb-6 group-hover:scale-110 transition-transform">
            <Target className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-bold text-white mb-2">Review Past Missions</h2>
          <p className="text-white/50 text-center max-w-sm">Re-translate scenarios you struggled with using Spaced Repetition.</p>
        </Link>
      </div>
    </div>
  );
}
