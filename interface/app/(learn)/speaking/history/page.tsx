"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { SpeakingAPI, SpeakingHistoryItem } from "@/lib/speaking-api";
import Link from "next/link";
import { ArrowLeft, Clock, Calendar } from "lucide-react";

export default function SpeakingHistoryPage() {
  const router = useRouter();
  const [items, setItems] = useState<SpeakingHistoryItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    SpeakingAPI.getHistory()
      .then((res) => {
        setItems(res.items || []);
      })
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  }, []);

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "scenario": return "🎭";
      case "free_talk": return "💬";
      default: return "🎯";
    }
  };

  const getTypeName = (type: string) => {
    switch (type) {
      case "scenario": return "Scenario";
      case "free_talk": return "Free Talk";
      case "quick_practice": return "Quick Practice";
      default: return "Practice";
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-6 pt-12 pb-24">
      <div className="flex items-center gap-4 mb-8">
        <Link 
          href="/speaking"
          className="p-2 hover:bg-white/10 rounded-full transition-colors text-white/70 hover:text-white"
        >
          <ArrowLeft className="w-6 h-6" />
        </Link>
        <h1 className="text-3xl font-bold text-white">Speaking History</h1>
      </div>

      {loading ? (
        <div className="text-white/50 text-center py-12">Loading history...</div>
      ) : items.length === 0 ? (
        <div className="text-white/50 text-center py-12 bg-white/5 rounded-2xl border border-white/10">
          <p className="mb-4">You haven't completed any speaking practices yet.</p>
          <Link href="/speaking" className="text-blue-400 hover:text-blue-300 font-medium">
            Start a practice session
          </Link>
        </div>
      ) : (
        <div className="grid gap-4">
          {items.map((item) => {
            const date = new Date(item.session.started_at).toLocaleDateString('en-US', {
              month: 'short', day: 'numeric', year: 'numeric'
            });
            const typeName = getTypeName(item.session.task_snapshot.task_type);
            const icon = getTypeIcon(item.session.task_snapshot.task_type);

            return (
              <div 
                key={item.session.id}
                onClick={() => router.push(`/speaking/session/${item.session.id}`)}
                className="bg-white/5 border border-white/10 hover:border-white/30 rounded-2xl p-5 cursor-pointer transition-colors flex items-center justify-between group"
              >
                <div className="flex items-start gap-4">
                  <div className="text-3xl w-12 h-12 flex items-center justify-center bg-white/5 rounded-xl border border-white/5">
                    {icon}
                  </div>
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <span className="text-xs font-semibold text-white/50 uppercase tracking-wider bg-white/5 px-2 py-0.5 rounded">
                        {typeName}
                      </span>
                    </div>
                    <h3 className="text-lg font-medium text-white mb-2 line-clamp-1">
                      {item.session.task_snapshot.topic}
                    </h3>
                    <div className="flex items-center gap-4 text-sm text-white/40">
                      <div className="flex items-center gap-1.5">
                        <Calendar className="w-3.5 h-3.5" />
                        {date}
                      </div>
                      <div className="flex items-center gap-1.5">
                        <Clock className="w-3.5 h-3.5" />
                        {item.attempts_count} attempt{item.attempts_count !== 1 ? 's' : ''}
                      </div>
                    </div>
                  </div>
                </div>

                <div className="text-right pl-4 border-l border-white/10 flex flex-col justify-center min-w-[100px]">
                  <div className="text-xs text-white/40 mb-1 uppercase tracking-wider font-medium">Best Score</div>
                  {item.best_score ? (
                    <div className="text-3xl font-bold text-white tabular-nums">
                      {item.best_score}<span className="text-base text-white/30 font-medium">/100</span>
                    </div>
                  ) : (
                    <div className="text-lg text-white/30 font-medium italic mt-1">Pending</div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
