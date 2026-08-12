"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { X, Play, RefreshCw, Layers, BrainCircuit } from "lucide-react";
import { CreateSessionParams } from "@/lib/learning-api";

interface ReviewSetModalProps {
  setId: string;
  isOpen: boolean;
  onClose: () => void;
}

export function ReviewSetModal({ setId, isOpen, onClose }: ReviewSetModalProps) {
  const router = useRouter();
  const [strategy, setStrategy] = useState<CreateSessionParams["strategy"]>("continue_learning");
  const [wordCount, setWordCount] = useState<number>(20);
  const [isLoading, setIsLoading] = useState(false);

  if (!isOpen) return null;

  const handleStartReview = () => {
    setIsLoading(true);
    // Navigate to session runner with proper query params
    // Wait, the specification states: "thay vì nhét strategy/word_count vào URL, mình khuyên truyền qua modal state → API trực tiếp. Flow: Word Set Viewer -> Review Modal -> POST createSession -> router.push(/learn/session?session_id=xxx)"
    // Let's implement that.
    import("@/lib/learning-api").then(({ LearningAPI }) => {
      LearningAPI.createSession({
        mode: "review",
        source: "word_set",
        word_set_id: setId,
        strategy,
        word_count: wordCount,
      })
        .then((s) => {
          router.push(`/learn/session?session_id=${s.id}`);
        })
        .catch((err: any) => {
          alert(err?.response?.data?.message || err.message || "Failed to create review session");
        })
        .finally(() => {
          setIsLoading(false);
        });
    });
  };

  const strategies = [
    {
      id: "continue_learning",
      icon: <Play className="w-5 h-5" />,
      title: "Continue Learning",
      desc: "Prioritize due words, then words you are learning, then new words.",
    },
    {
      id: "fresh_words",
      icon: <Layers className="w-5 h-5" />,
      title: "Fresh Words",
      desc: "Skip words you are currently learning. Prioritize due words, then new words.",
    },
    {
      id: "smart_mix",
      icon: <BrainCircuit className="w-5 h-5" />,
      title: "Smart Mix",
      desc: "System automatically balances due, learning, and new words for optimal retention.",
    },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />

      {/* Modal */}
      <div className="relative w-full max-w-md bg-neutral-900 border border-white/10 rounded-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div className="p-6 border-b border-white/5 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-white">Review Options</h2>
          <button
            onClick={onClose}
            className="p-2 text-white/40 hover:text-white hover:bg-white/5 rounded-full transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6 space-y-8">
          {/* Strategy Selection */}
          <div className="space-y-4">
            <h3 className="text-sm font-medium text-white/60 uppercase tracking-wider">Strategy</h3>
            <div className="space-y-3">
              {strategies.map((s) => (
                <button
                  key={s.id}
                  onClick={() => setStrategy(s.id as any)}
                  className={`w-full flex items-start p-4 rounded-xl border text-left transition-all ${
                    strategy === s.id
                      ? "bg-blue-500/10 border-blue-500/50 text-blue-400"
                      : "bg-white/5 border-transparent text-white hover:bg-white/10"
                  }`}
                >
                  <div className="shrink-0 mt-0.5">{s.icon}</div>
                  <div className="ml-3">
                    <div className="font-medium text-sm mb-1">{s.title}</div>
                    <div className="text-xs opacity-70 leading-relaxed">{s.desc}</div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Word Count */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-medium text-white/60 uppercase tracking-wider">Words to Review</h3>
              <span className="text-xl font-bold text-white">{wordCount}</span>
            </div>
            <input
              type="range"
              min="5"
              max="25"
              step="5"
              value={wordCount}
              onChange={(e) => setWordCount(parseInt(e.target.value))}
              className="w-full h-2 bg-white/10 rounded-lg appearance-none cursor-pointer accent-blue-500"
            />
            <div className="flex justify-between text-xs text-white/40 font-medium">
              <span>5</span>
              <span>10</span>
              <span>15</span>
              <span>20</span>
              <span>25</span>
            </div>
          </div>
        </div>

        <div className="p-6 border-t border-white/5 bg-black/20">
          <button
            onClick={handleStartReview}
            disabled={isLoading}
            className="w-full flex items-center justify-center py-3.5 px-4 bg-white text-black hover:bg-gray-200 rounded-xl font-semibold transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? (
              <RefreshCw className="w-5 h-5 animate-spin" />
            ) : (
              "Start Session"
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
