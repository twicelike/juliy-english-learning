"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { LearningAPI, DashboardDTO } from "@/lib/learning-api";
import { ProfileAPI, DashboardResponse, RecommendationAction } from "@/lib/profile-api";
import { useAuthStore } from "@/store/auth-store";

// Map action -> href for the Recommendation Card CTA
const ACTION_HREF: Record<RecommendationAction, string> = {
  review: "/learn/flashcard?type=review",
  learn_new: "/learn/flashcard?type=new",
  grammar: "/learn/writing",
  fill_blank: "/learn/fill-blank",
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
  const [dashboard, setDashboard] = useState<DashboardDTO | null>(null);
  const [profileDash, setProfileDash] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }

    Promise.all([
      LearningAPI.getDashboard(),
      ProfileAPI.getDashboard(),
    ])
      .then(([learn, profile]) => {
        setDashboard(learn);
        setProfileDash(profile);

        // Redirect new users to placement test
        if (!profile.profile.completed_placement) {
          router.push("/placement");
        }
      })
      .catch(() => router.push("/login"))
      .finally(() => setLoading(false));
  }, [user, router]);

  if (loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  const stats = [
    { label: "New Words", value: dashboard?.new ?? 0, color: "from-sky-500 to-blue-600", href: "/learn/flashcard?type=new" },
    { label: "Due Today", value: dashboard?.due_today ?? 0, color: "from-amber-400 to-orange-500", href: "/learn/flashcard?type=review" },
    { label: "Mature", value: dashboard?.mature_words ?? 0, color: "from-emerald-400 to-green-600", href: null },
  ];

  const accuracyPct = dashboard?.accuracy ? Math.round(dashboard.accuracy * 100) : 0;
  const profile = profileDash?.profile;
  const rec = profileDash?.recommendation;
  const todayProgress = profileDash?.today_progress ?? 0;
  const dailyGoal = profile?.daily_goal ?? 20;
  const progressPct = Math.min(100, Math.round((todayProgress / dailyGoal) * 100));
  const streak = profileDash?.streak ?? 0;

  return (
    <div className="max-w-4xl mx-auto px-6 pt-12 pb-16 space-y-8">
      {/* Header */}
      <div className="flex justify-between items-end">
        <div>
          <h1 className="text-4xl font-bold text-white mb-2">Study Dashboard</h1>
          <p className="text-white/50">Your personalised spaced repetition schedule</p>
        </div>
        <div className="flex items-center gap-4 text-right">
          {streak > 0 && (
            <div className="text-right">
              <div className="text-2xl font-bold text-orange-400">🔥 {streak}</div>
              <div className="text-xs text-white/40">day streak</div>
            </div>
          )}
          <div className="text-right">
            <div className="text-sm text-white/50 mb-1">Review Accuracy</div>
            <div className="text-2xl font-bold text-emerald-400">{accuracyPct}%</div>
          </div>
        </div>
      </div>

      {/* CEFR Level + Today's Progress */}
      {profile && profile.completed_placement && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {/* CEFR Card */}
          <div className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10">
            <div className="absolute inset-0 bg-gradient-to-br from-violet-500 to-indigo-600 opacity-10" />
            <div className="relative flex items-center justify-between">
              <div>
                <div className="text-white/50 text-sm mb-1">Current Level</div>
                <div className="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-cyan-400">
                  {profile.current_cefr}
                </div>
              </div>
              <div className="text-right text-sm text-white/40 max-w-[160px]">
                {profile.weaknesses.length > 0 && (
                  <div>
                    <span className="text-red-400 font-medium">Weak: </span>
                    {profile.weaknesses.slice(0, 2).join(", ")}
                  </div>
                )}
                {profile.strengths.length > 0 && (
                  <div className="mt-1">
                    <span className="text-emerald-400 font-medium">Strong: </span>
                    {profile.strengths.slice(0, 2).join(", ")}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Today's Progress */}
          <div className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10">
            <div className="absolute inset-0 bg-gradient-to-br from-emerald-500 to-green-600 opacity-10" />
            <div className="relative">
              <div className="text-white/50 text-sm mb-1">Today&apos;s Progress</div>
              <div className="text-2xl font-bold text-white mb-3">
                {todayProgress} / {dailyGoal} <span className="text-white/40 text-base font-normal">words</span>
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
      )}

      {/* Recommendation Card */}
      {rec && rec.action !== "done" && (
        <Link
          href={ACTION_HREF[rec.action]}
          className="flex items-center gap-5 p-6 rounded-2xl bg-indigo-500/10 border border-indigo-500/40 hover:bg-indigo-500/20 transition-all group"
        >
          <div className="text-4xl">{ACTION_ICON[rec.action]}</div>
          <div className="flex-1 min-w-0">
            <div className="text-xs uppercase tracking-widest text-indigo-400 mb-1">Recommended Next</div>
            <div className="text-lg font-bold text-white">{rec.title}</div>
            <div className="text-white/50 text-sm">{rec.description}</div>
            {rec.ai_message && (
              <div className="mt-2 text-xs text-indigo-300/70 italic">💬 &quot;{rec.ai_message}&quot;</div>
            )}
          </div>
          <div className="text-indigo-400 group-hover:translate-x-1 transition-transform text-xl">→</div>
        </Link>
      )}

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        {stats.map((s) => (
          <div key={s.label} className="relative overflow-hidden rounded-2xl p-6 bg-white/5 border border-white/10 group">
            <div className={`absolute inset-0 bg-gradient-to-br ${s.color} opacity-10 group-hover:opacity-20 transition-opacity`} />
            <div className="relative">
              <div className="text-4xl font-bold text-white mb-1">{s.value}</div>
              <div className="text-white/60 text-sm">{s.label}</div>
              {s.href && s.value > 0 && (
                <Link
                  href={s.href}
                  className={`mt-4 inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r ${s.color} text-white text-sm font-medium hover:opacity-90 transition-opacity`}
                >
                  Study now →
                </Link>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Session launchers */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        <Link href="/learn/flashcard" className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <div className="text-3xl">🃏</div>
          <h2 className="text-lg font-bold text-white">Flashcards</h2>
          <p className="text-white/50 text-sm">Flip-card review with Pass / Fail scoring</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Start session →</span>
        </Link>
        <Link href="/learn/quiz" className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <div className="text-3xl">🧠</div>
          <h2 className="text-lg font-bold text-white">Multiple Choice Quiz</h2>
          <p className="text-white/50 text-sm">Pick the correct definition from same-POS distractors</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Start quiz →</span>
        </Link>
        <Link href="/learn/writing" className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <div className="text-3xl">✏️</div>
          <h2 className="text-lg font-bold text-white">Writing Lab</h2>
          <p className="text-white/50 text-sm">AI-powered grammar check and vocabulary improvement</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Open lab →</span>
        </Link>
        <Link href="/learn/fill-blank" className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group">
          <div className="text-3xl">🤖</div>
          <h2 className="text-lg font-bold text-white">AI Fill-in-the-Blank</h2>
          <p className="text-white/50 text-sm">AI generates sentences from your bookmarked words</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Start quiz →</span>
        </Link>
      </div>
    </div>
  );
}
