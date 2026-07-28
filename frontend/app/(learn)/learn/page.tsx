"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { LearningAPI, DashboardDTO } from "@/lib/learning-api";
import { useAuthStore } from "@/store/auth-store";

export default function LearnDashboard() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [dashboard, setDashboard] = useState<DashboardDTO | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }
    LearningAPI.getDashboard()
      .then(setDashboard)
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
    { label: "New Words", value: dashboard?.new ?? 0, color: "from-sky-500 to-blue-600", icon: "✦", href: "/learn/flashcard?type=new" },
    { label: "Due Today", value: dashboard?.due_today ?? 0, color: "from-amber-400 to-orange-500", icon: "⟳", href: "/learn/flashcard?type=review" },
    { label: "Mature", value: dashboard?.mature_words ?? 0, color: "from-emerald-400 to-green-600", icon: "✓", href: null },
  ];

  const accuracyPct = dashboard?.accuracy ? Math.round(dashboard.accuracy * 100) : 0;

  return (
    <div className="max-w-4xl mx-auto px-6 pt-12 pb-16">
      <div className="mb-12 flex justify-between items-end">
        <div>
          <h1 className="text-4xl font-bold text-white mb-2">Study Dashboard</h1>
          <p className="text-white/50">Your personalised spaced repetition schedule</p>
        </div>
        <div className="text-right">
          <div className="text-sm text-white/50 mb-1">Review Accuracy</div>
          <div className="text-2xl font-bold text-emerald-400">{accuracyPct}%</div>
        </div>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mb-12">
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
        <Link
          href="/learn/flashcard"
          className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group"
        >
          <div className="text-3xl">🃏</div>
          <h2 className="text-lg font-bold text-white">Flashcards</h2>
          <p className="text-white/50 text-sm">Flip-card review with Pass / Fail scoring</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Start session →</span>
        </Link>
        <Link
          href="/learn/quiz"
          className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group"
        >
          <div className="text-3xl">🧠</div>
          <h2 className="text-lg font-bold text-white">Multiple Choice Quiz</h2>
          <p className="text-white/50 text-sm">Pick the correct definition from same-POS distractors</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Start quiz →</span>
        </Link>
        <Link
          href="/learn/writing"
          className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group"
        >
          <div className="text-3xl">✏️</div>
          <h2 className="text-lg font-bold text-white">Writing Lab</h2>
          <p className="text-white/50 text-sm">AI-powered grammar check and vocabulary improvement</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Open lab →</span>
        </Link>
        <Link
          href="/learn/fill-blank"
          className="flex flex-col gap-3 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors group"
        >
          <div className="text-3xl">🤖</div>
          <h2 className="text-lg font-bold text-white">AI Fill-in-the-Blank</h2>
          <p className="text-white/50 text-sm">AI generates sentences from your bookmarked words</p>
          <span className="text-indigo-400 text-sm group-hover:underline">Start quiz →</span>
        </Link>
      </div>
    </div>
  );
}
