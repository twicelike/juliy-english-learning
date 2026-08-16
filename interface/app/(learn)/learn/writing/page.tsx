"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/features/auth";
import { WritingAPI, DashboardStatsDTO } from "@/lib/writing-api";
import Link from "next/link";
import { BookOpen, Target, BrainCircuit, PenTool, ArrowRight, X } from "lucide-react";

export default function WritingDashboardPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [stats, setStats] = useState<DashboardStatsDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);

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

  const handleStartNew = () => setShowModal(true);
  const handleDifficultySelect = (diff: string) => {
    router.push(`/learn/writing/play?mode=new&difficulty=${diff}`);
  };

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
          Writing Missions
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
        <div className="bg-white/5 border border-white/10 rounded-3xl p-6 relative overflow-hidden group flex flex-col justify-between">
          <div className="absolute inset-0 bg-gradient-to-br from-emerald-500/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
          <div>
            <div className="flex items-center gap-4 mb-4">
              <div className="p-3 bg-emerald-500/20 rounded-2xl relative z-10">
                <PenTool className="w-6 h-6 text-emerald-400" />
              </div>
              <h2 className="text-white/80 font-medium relative z-10">Total Completed</h2>
            </div>
            <div className="flex items-end gap-2 mb-6 relative z-10">
              <span className="text-5xl font-black text-white">{totalCompleted}</span>
              <span className="text-white/40 font-medium mb-1">missions</span>
            </div>
          </div>
          
          <div className="grid grid-cols-3 gap-2 mt-auto relative z-10">
             <div className="bg-white/5 rounded-xl p-2 flex flex-col items-center border border-white/5 hover:border-emerald-500/30 transition-colors">
               <span className="text-[10px] uppercase tracking-wider text-emerald-400/70 mb-1 font-semibold">Easy</span>
               <span className="font-bold text-white">{stats?.completed_easy || 0}</span>
             </div>
             <div className="bg-white/5 rounded-xl p-2 flex flex-col items-center border border-white/5 hover:border-indigo-500/30 transition-colors">
               <span className="text-[10px] uppercase tracking-wider text-indigo-400/70 mb-1 font-semibold">Medium</span>
               <span className="font-bold text-white">{stats?.completed_medium || 0}</span>
             </div>
             <div className="bg-white/5 rounded-xl p-2 flex flex-col items-center border border-white/5 hover:border-rose-500/30 transition-colors">
               <span className="text-[10px] uppercase tracking-wider text-rose-400/70 mb-1 font-semibold">Hard</span>
               <span className="font-bold text-white">{stats?.completed_hard || 0}</span>
             </div>
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
        <button 
          onClick={handleStartNew}
          className="group relative flex flex-col items-center justify-center p-12 bg-indigo-500/10 border border-indigo-500/20 hover:border-indigo-500/50 hover:bg-indigo-500/20 rounded-3xl transition-all overflow-hidden text-left w-full"
        >
          <div className="absolute inset-0 bg-gradient-to-b from-indigo-500/5 to-transparent"></div>
          <div className="w-16 h-16 bg-indigo-500 rounded-2xl flex items-center justify-center mb-6 shadow-lg shadow-indigo-500/30 group-hover:scale-110 transition-transform">
            <BookOpen className="w-8 h-8 text-white" />
          </div>
          <h2 className="text-2xl font-bold text-white mb-2">Start New Mission</h2>
          <p className="text-indigo-200/60 text-center max-w-sm">Translate a completely new scenario and learn new vocabulary.</p>
        </button>

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

      {stats?.recent_submissions && stats.recent_submissions.length > 0 && (
        <div className="mt-16">
          <h2 className="text-2xl font-bold text-white mb-6 flex items-center gap-3">
            <span className="w-8 h-1 bg-indigo-500 rounded-full"></span>
            Recent Submissions
          </h2>
          <div className="space-y-6">
            {stats.recent_submissions.map((sub, idx) => (
              <SubmissionItem key={idx} sub={sub} />
            ))}
          </div>
        </div>
      )}

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-[#1a1b26] border border-white/10 p-8 rounded-3xl max-w-md w-full relative animate-in zoom-in-95 duration-200">
            <button 
              onClick={() => setShowModal(false)}
              className="absolute top-4 right-4 text-white/50 hover:text-white transition-colors"
            >
              <X className="w-6 h-6" />
            </button>
            <h2 className="text-2xl font-bold text-white mb-2">Select Difficulty</h2>
            <p className="text-white/60 mb-8">Choose the challenge level for your next writing mission.</p>
            
            <div className="grid grid-cols-1 gap-4">
              {[
                { id: "Easy", label: "Easy", desc: "Shorter sentences, basic vocabulary", color: "from-emerald-500/20 to-emerald-500/5", border: "border-emerald-500/20 hover:border-emerald-500/50" },
                { id: "Medium", label: "Medium", desc: "Standard everyday scenarios", color: "from-indigo-500/20 to-indigo-500/5", border: "border-indigo-500/20 hover:border-indigo-500/50" },
                { id: "Hard", label: "Hard", desc: "Complex structures, advanced vocabulary", color: "from-rose-500/20 to-rose-500/5", border: "border-rose-500/20 hover:border-rose-500/50" },
                { id: "Random", label: "Random", desc: "Surprise me with any level", color: "from-purple-500/20 to-purple-500/5", border: "border-purple-500/20 hover:border-purple-500/50" },
              ].map((diff) => (
                <button
                  key={diff.id}
                  onClick={() => handleDifficultySelect(diff.id)}
                  className={`flex flex-col text-left p-4 rounded-2xl border bg-gradient-to-r transition-all group/btn ${diff.color} ${diff.border}`}
                >
                  <span className="text-lg font-bold text-white group-hover/btn:translate-x-1 transition-transform">{diff.label}</span>
                  <span className="text-white/50 text-sm mt-1">{diff.desc}</span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function SubmissionItem({ sub }: { sub: any }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div 
      className={`bg-white/5 border border-white/10 rounded-3xl p-6 md:p-8 transition-colors ${!expanded ? "cursor-pointer hover:bg-white/10 hover:border-indigo-500/30" : ""}`}
      onClick={() => {
        if (!expanded) setExpanded(true);
      }}
    >
      <div className="flex justify-between items-center group">
        <div>
          <div className="text-indigo-400 font-semibold mb-1">{sub.topic}</div>
          <div className="text-white/60 text-sm">{sub.scenario}</div>
        </div>
        <div className="flex items-center gap-4">
          <div className={`text-2xl font-black ${
            sub.score >= 90 ? "text-emerald-400" :
            sub.score >= 70 ? "text-amber-400" : "text-rose-400"
          }`}>
            {sub.score}<span className="text-white/30 text-base">/100</span>
          </div>
          <button 
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
            className="w-8 h-8 flex items-center justify-center rounded-full bg-white/5 hover:bg-white/10 text-white/50 transition-colors"
          >
            <svg 
              className={`w-5 h-5 transition-transform duration-300 ${expanded ? "rotate-180" : ""}`} 
              fill="none" viewBox="0 0 24 24" stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        </div>
      </div>

      {expanded && (
        <div className="mt-6 border-t border-white/10 pt-6 animate-in fade-in slide-in-from-top-4 duration-300">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-[#12131C] rounded-2xl p-5 border border-white/5">
              <div className="text-xs uppercase tracking-widest text-white/40 font-semibold mb-3">Your Answer</div>
              <p className="text-white/80 leading-relaxed italic">&quot;{sub.user_content}&quot;</p>
            </div>
            <div className="bg-emerald-500/10 rounded-2xl p-5 border border-emerald-500/20">
              <div className="text-xs uppercase tracking-widest text-emerald-400/70 font-semibold mb-3">Corrected Version</div>
              <p className="text-white leading-relaxed">&quot;{sub.corrected_version}&quot;</p>
            </div>
          </div>

          <div className="mt-6 bg-white/5 rounded-2xl p-5 text-sm text-white/70 leading-relaxed">
            <span className="text-indigo-400 font-semibold mr-2">Feedback:</span>
            {sub.detailed_feedback}
          </div>
        </div>
      )}
    </div>
  );
}
