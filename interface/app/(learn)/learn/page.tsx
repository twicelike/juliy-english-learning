"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { DashboardAPI, UnifiedDashboardResponse } from "@/lib/dashboard-api";
import { DashboardSkeleton } from "@/components/skeletons/DashboardSkeleton";
import { LazySection } from "@/components/LazySection";
import { RecommendationAction } from "@/lib/profile-api";
import { useAuthStore } from "@/store/auth-store";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

// Map action -> href for the Recommendation Card CTA
const ACTION_HREF: Record<RecommendationAction, string> = {
  review: "/learn/session?mode=review",
  learn_new: "/learn/session?mode=new_words",
  grammar: "/learn/writing",
  fill_blank: "/learn/session?mode=review", // Just fall back to review for now
  done: "/learn",
};

const ACTION_ICON: Record<RecommendationAction, string> = {
  review: "⟳",
  learn_new: "✦",
  grammar: "✏️",
  fill_blank: "🤖",
  done: "✅",
};

export default function LearnDashboard() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [data, setData] = useState<UnifiedDashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [showXpInfo, setShowXpInfo] = useState(false);
  const [activeFeatureModal, setActiveFeatureModal] = useState<string | null>(null);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }

    DashboardAPI.getUnifiedDashboard()
      .then((dash) => {
        setData(dash);

        // Redirect new users to placement test
        if (!dash.profile.completed_placement) {
          router.push("/placement");
        }
      })
      .catch(() => router.push("/login"))
      .finally(() => setLoading(false));
  }, [user, router]);

  if (loading) {
    return (
      <div className="p-4 md:p-8">
        <DashboardSkeleton />
      </div>
    );
  }

  if (!data || !data.gamification) return null;

  const {
    profile,
    recommendation,
    gamification,
    statistics,
    charts = [],
  } = data;

  const {
    level,
    current_xp,
    next_level_xp,
    streak: current_streak,
    achievements = [], // default to empty array
    today_progress,
    daily_goal,
  } = gamification;

  const dailyGoal = daily_goal || 20;
  const progressPct = Math.min(100, Math.round((today_progress / dailyGoal) * 100));
  const levelProgressPct = Math.min(100, Math.round((current_xp / next_level_xp) * 100));

  // Format dates for chart (e.g. "2026-08-05" -> "Aug 05")
  const formattedChartData = charts.map((d) => {
    const parts = d.date.split("-");
    const label = parts.length === 3 ? `${parts[1]}/${parts[2]}` : d.date;
    return { ...d, label };
  });

  return (
    <div className="max-w-4xl mx-auto px-6 pt-12 pb-16 space-y-8">
      {/* Header */}
      <div className="flex justify-between items-end">
        <div>
          <h1 className="text-4xl font-bold text-white mb-2">Study Dashboard</h1>
          <p className="text-white/50">Your personalized spaced repetition schedule</p>
        </div>
        <div className="flex items-center gap-4 text-right">
          {current_streak > 0 && (
            <div className="text-right">
              <div className="text-2xl font-bold text-orange-400">🔥 {current_streak}</div>
              <div className="text-xs text-white/40">day streak</div>
            </div>
          )}
        </div>
      </div>

      {/* Gamification Bar (Level & XP) */}
      <div className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10 flex items-center gap-6">
        <div className="w-16 h-16 rounded-full bg-gradient-to-br from-yellow-400 to-orange-500 flex items-center justify-center text-2xl font-bold text-black shadow-[0_0_20px_rgba(251,191,36,0.3)] shrink-0">
          L{level}
        </div>
        <div className="flex-1">
          <div className="flex justify-between text-sm mb-2">
            <span className="font-semibold text-white/80 flex items-center gap-2">
              Level {level}
              <button 
                onClick={() => setShowXpInfo(true)}
                className="w-5 h-5 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-xs text-white/60 hover:text-white transition-colors"
                title="How to gain XP"
              >
                i
              </button>
            </span>
            <span className="text-white/50">{current_xp} / {next_level_xp} XP</span>
          </div>
          <div className="w-full bg-white/10 rounded-full h-3 overflow-hidden">
            <div
              className="bg-gradient-to-r from-yellow-400 to-orange-500 h-3 rounded-full transition-all duration-1000"
              style={{ width: `${levelProgressPct}%` }}
            />
          </div>
        </div>
        {achievements && achievements.length > 0 && (
          <div className="flex gap-2 shrink-0 border-l border-white/10 pl-6">
            {achievements.slice(0, 3).map((a) => (
              <div key={a.id} className="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center text-xl tooltip-trigger" title={a.title}>
                {a.icon}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {/* Proficiency Profile Card */}
        <div className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10">
          <div className="absolute inset-0 bg-gradient-to-br from-violet-500 to-indigo-600 opacity-10" />
          <div className="relative flex flex-col justify-center h-full gap-4">
            {profile.weaknesses.length > 0 && (
              <div>
                <div className="text-xs uppercase tracking-widest text-red-400/70 font-semibold mb-1">Areas to Improve (Weak)</div>
                <div className="text-white/80 font-medium text-lg">
                  {profile.weaknesses.slice(0, 2).join(", ")}
                </div>
              </div>
            )}
            {profile.strengths.length > 0 && (
              <div>
                <div className="text-xs uppercase tracking-widest text-emerald-400/70 font-semibold mb-1">Strengths</div>
                <div className="text-white/80 font-medium text-lg">
                  {profile.strengths.slice(0, 2).join(", ")}
                </div>
              </div>
            )}
            {profile.weaknesses.length === 0 && profile.strengths.length === 0 && (
              <div className="text-white/50 text-sm italic text-center">
                Keep studying to discover your strengths and weaknesses.
              </div>
            )}
          </div>
        </div>

        {/* Word Bank Breakdown */}
        <div className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10">
          <div className="absolute inset-0 bg-gradient-to-br from-blue-500 to-cyan-600 opacity-10" />
          <div className="relative">
            <div className="text-white/50 text-sm mb-4">Word Bank</div>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <div className="text-sm text-white/80 flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-orange-400"></span> Review Due</div>
                <div className="font-bold text-white">{statistics.reviews_due}</div>
              </div>
              <div className="flex justify-between items-center">
                <div className="text-sm text-white/80 flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-yellow-400"></span> Pending</div>
                <div className="font-bold text-white">{statistics.pending_words}</div>
              </div>
              <div className="flex justify-between items-center">
                <div className="text-sm text-white/80 flex items-center gap-2"><span className="w-2 h-2 rounded-full bg-indigo-400"></span> New Available</div>
                <div className="font-bold text-white">{statistics.new_words}</div>
              </div>
            </div>
          </div>
        </div>

        {/* Today's Progress */}
        <div className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10">
          <div className="absolute inset-0 bg-gradient-to-br from-emerald-500 to-green-600 opacity-10" />
          <div className="relative">
            <div className="text-white/50 text-sm mb-1">Today&apos;s Goal Progress</div>
            <div className="text-2xl font-bold text-white mb-3">
              {today_progress} / {dailyGoal} <span className="text-white/40 text-base font-normal">items</span>
            </div>
            <div className="w-full bg-white/5 rounded-full h-2.5">
              <div
                className="bg-gradient-to-r from-emerald-500 to-green-400 h-2.5 rounded-full transition-all duration-500"
                style={{ width: `${progressPct}%` }}
              />
            </div>
            <div className="text-right text-xs text-white/40 mt-1">{progressPct}%</div>
          </div>
        </div>
      </div>

      {/* Recommendation Card */}
      {recommendation.action !== "done" && (
        <Link
          href={ACTION_HREF[recommendation.action]}
          className="flex items-center gap-5 p-6 rounded-2xl bg-indigo-500/10 border border-indigo-500/40 hover:bg-indigo-500/20 transition-all group"
        >
          <div className="text-4xl">{ACTION_ICON[recommendation.action]}</div>
          <div className="flex-1 min-w-0">
            <div className="text-xs uppercase tracking-widest text-indigo-400 mb-1">Recommended Next</div>
            <div className="text-lg font-bold text-white">{recommendation.title}</div>
            <div className="text-white/50 text-sm">{recommendation.description}</div>
            {recommendation.ai_message && (
              <div className="mt-2 text-xs text-indigo-300/70 italic">💬 &quot;{recommendation.ai_message}&quot;</div>
            )}
          </div>
          <div className="text-indigo-400 group-hover:translate-x-1 transition-transform text-xl">→</div>
        </Link>
      )}

      {/* Session launchers */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        <div className="relative flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <button 
            onClick={(e) => { e.preventDefault(); setActiveFeatureModal('flashcard'); }}
            className="absolute top-4 right-4 w-6 h-6 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-xs text-white/60 hover:text-white transition-colors z-10"
            title="How to use Flashcards"
          >i</button>
          <Link href="/learn/session?mode=new_words" className="absolute inset-0 z-0"></Link>
          <div className="text-3xl relative z-10 pointer-events-none">🃏</div>
          <h2 className="text-lg font-bold text-white relative z-10 pointer-events-none">Learn New Words</h2>
          <p className="text-white/50 text-sm relative z-10 pointer-events-none">Unified session: Flashcard → Quiz → Context</p>
          <span className="text-indigo-400 text-sm group-hover:underline relative z-10 pointer-events-none">Start session →</span>
        </div>
        <div className="relative flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <button 
            onClick={(e) => { e.preventDefault(); setActiveFeatureModal('quiz'); }}
            className="absolute top-4 right-4 w-6 h-6 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-xs text-white/60 hover:text-white transition-colors z-10"
            title="How to use Quiz"
          >i</button>
          <Link href="/learn/session?mode=review" className="absolute inset-0 z-0"></Link>
          <div className="text-3xl relative z-10 pointer-events-none">🧠</div>
          <h2 className="text-lg font-bold text-white relative z-10 pointer-events-none">Review Due Words</h2>
          <p className="text-white/50 text-sm relative z-10 pointer-events-none">Keep your memory fresh with a unified review session</p>
          <span className="text-indigo-400 text-sm group-hover:underline relative z-10 pointer-events-none">Start quiz →</span>
        </div>
        <div className="relative flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <button 
            onClick={(e) => { e.preventDefault(); setActiveFeatureModal('writing'); }}
            className="absolute top-4 right-4 w-6 h-6 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-xs text-white/60 hover:text-white transition-colors z-10"
            title="How to use Writing Lab"
          >i</button>
          <Link href="/learn/writing" className="absolute inset-0 z-0"></Link>
          <div className="text-3xl relative z-10 pointer-events-none">✏️</div>
          <h2 className="text-lg font-bold text-white relative z-10 pointer-events-none">Writing Lab</h2>
          <p className="text-white/50 text-sm relative z-10 pointer-events-none">AI-powered grammar check and vocabulary improvement</p>
          <span className="text-indigo-400 text-sm group-hover:underline relative z-10 pointer-events-none">Open lab →</span>
        </div>
      </div>

      {/* Activity Chart Section */}
      <div className="rounded-2xl p-6 bg-white/5 border border-white/10">
        <h3 className="text-lg font-bold text-white mb-6">Learning Activity (Last 7 Days)</h3>
        <LazySection fallback={<div className="h-[250px] bg-white/5 animate-pulse rounded-xl"></div>}>
          <div className="h-[250px] w-full mt-6">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={formattedChartData}>
                <defs>
                  <linearGradient id="colorXp" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#818cf8" stopOpacity={0.8} />
                    <stop offset="95%" stopColor="#818cf8" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
                <XAxis 
                  dataKey="label" 
                  stroke="rgba(255,255,255,0.3)" 
                  tick={{fill: 'rgba(255,255,255,0.5)', fontSize: 12}}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis 
                  stroke="rgba(255,255,255,0.3)" 
                  tick={{fill: 'rgba(255,255,255,0.5)', fontSize: 12}}
                  axisLine={false}
                  tickLine={false}
                />
                <Tooltip 
                  contentStyle={{ backgroundColor: '#1e1e2d', borderColor: 'rgba(255,255,255,0.1)', borderRadius: '12px', color: '#fff' }}
                  itemStyle={{ color: '#818cf8' }}
                  cursor={{ stroke: 'rgba(255,255,255,0.1)', strokeWidth: 1 }}
                />
                <Line 
                  type="monotone" 
                  dataKey="xp_gained" 
                  name="XP Earned" 
                  stroke="#818cf8" 
                  strokeWidth={3}
                  dot={{ fill: '#818cf8', r: 4, strokeWidth: 0 }}
                  activeDot={{ r: 6, fill: '#6366f1', stroke: '#fff', strokeWidth: 2 }}
                  fillOpacity={1} 
                  fill="url(#colorXp)" 
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </LazySection>
      </div>

      {/* XP Info Modal */}
      {showXpInfo && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" onClick={() => setShowXpInfo(false)}>
          <div className="bg-[#1e1e2d] border border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-center mb-6">
              <h3 className="text-xl font-bold text-white flex items-center gap-2">
                <span className="text-orange-400">🔥</span> How to gain XP & Level Up
              </h3>
              <button 
                onClick={() => setShowXpInfo(false)}
                className="text-white/50 hover:text-white transition-colors"
              >
                ✕
              </button>
            </div>
            
            <div className="space-y-4">
              <div className="flex justify-between items-center p-3 rounded-lg bg-white/5">
                <div className="text-white/80">Learn a New Word</div>
                <div className="text-emerald-400 font-bold">+10 XP</div>
              </div>
              <div className="flex justify-between items-center p-3 rounded-lg bg-white/5">
                <div className="text-white/80">Review a Word Correctly</div>
                <div className="text-emerald-400 font-bold">+5 XP</div>
              </div>
              <div className="flex justify-between items-center p-3 rounded-lg bg-white/5">
                <div className="text-white/80">Complete Daily Goal ({dailyGoal} items)</div>
                <div className="text-emerald-400 font-bold">+20 XP Bonus</div>
              </div>
            </div>
            
            <div className="mt-6 text-sm text-white/50 bg-indigo-500/10 p-4 rounded-xl border border-indigo-500/20">
              <p>Keep your daily streak alive and complete your daily goals to rank up your CEFR level faster! The more XP you earn, the higher your level becomes.</p>
            </div>
            
            <button 
              onClick={() => setShowXpInfo(false)}
              className="mt-6 w-full py-3 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white font-bold transition-colors"
            >
              Got it, let&apos;s study!
            </button>
          </div>
        </div>
      )}

      {/* Feature Info Modal */}
      {activeFeatureModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" onClick={() => setActiveFeatureModal(null)}>
          <div className="bg-[#1e1e2d] border border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl" onClick={e => e.stopPropagation()}>
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-bold text-white flex items-center gap-2">
                {activeFeatureModal === 'flashcard' && <><span className="text-xl">🃏</span> Flashcards Guide</>}
                {activeFeatureModal === 'quiz' && <><span className="text-xl">🧠</span> Quiz Guide</>}
                {activeFeatureModal === 'writing' && <><span className="text-xl">✏️</span> Writing Lab Guide</>}
                {activeFeatureModal === 'fill-blank' && <><span className="text-xl">🤖</span> AI Fill-in Guide</>}
              </h3>
              <button 
                onClick={() => setActiveFeatureModal(null)}
                className="text-white/50 hover:text-white transition-colors"
              >
                ✕
              </button>
            </div>
            
            <div className="text-white/70 space-y-4 leading-relaxed text-sm">
              {activeFeatureModal === 'flashcard' && (
                <>
                  <p>Flashcards use <strong>Spaced Repetition (SM-2)</strong> to help you memorize words efficiently.</p>
                  <ul className="list-disc pl-5 space-y-2">
                    <li>Look at the word and try to recall its meaning.</li>
                    <li>Flip the card to check if you were right.</li>
                    <li>Rate your recall: <strong>Hard</strong>, <strong>Good</strong>, or <strong>Easy</strong>.</li>
                    <li>Words you struggle with will appear more often, while easy words will be spaced out to save your time.</li>
                  </ul>
                </>
              )}
              {activeFeatureModal === 'quiz' && (
                <>
                  <p>Test your knowledge with multiple-choice questions!</p>
                  <ul className="list-disc pl-5 space-y-2">
                    <li>The system generates questions based on your saved words.</li>
                    <li>Distractors (wrong answers) are intelligently chosen to be the same part-of-speech to challenge your understanding.</li>
                    <li>Completing quizzes helps reinforce your memory and earns you XP.</li>
                  </ul>
                </>
              )}
              {activeFeatureModal === 'writing' && (
                <>
                  <p>The Writing Lab is your personal AI English tutor.</p>
                  <ul className="list-disc pl-5 space-y-2">
                    <li>Type paragraphs or sentences using your new vocabulary.</li>
                    <li>The AI will check your grammar, spelling, and sentence structure.</li>
                    <li>You will receive detailed feedback, corrections, and suggestions for more natural phrasing.</li>
                  </ul>
                </>
              )}
            </div>
            
            <button 
              onClick={() => setActiveFeatureModal(null)}
              className="mt-6 w-full py-3 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white font-bold transition-colors"
            >
              Close Guide
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
