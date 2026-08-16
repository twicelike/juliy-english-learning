"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, Mic, AlertCircle, Loader2 } from "lucide-react";
import { SpeakingAPI, SpeakingTask, SpeakingTaskType } from "@/lib/speaking-api";

function SpeakingPreviewContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const topicParam = searchParams.get("topic") || "random";
  const typeParam = (searchParams.get("type") as SpeakingTaskType) || "quick_practice";

  const [task, setTask] = useState<SpeakingTask | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isStarting, setIsStarting] = useState(false);

  const loadTask = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const fetchedTask = await SpeakingAPI.getRandomTask({
        topic: topicParam,
        type: typeParam,
      });
      setTask(fetchedTask);
    } catch (err: any) {
      console.error(err);
      setError("Unable to load this speaking task. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadTask();
  }, [topicParam, typeParam]);

  const handleStartSpeaking = async () => {
    if (!task || isStarting) return;
    setIsStarting(true);
    try {
      const session = await SpeakingAPI.startSession(task.id);
      router.push(`/speaking/session/${session.id}`);
    } catch (err) {
      console.error(err);
      alert("Failed to start session. Please try again.");
      setIsStarting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto px-6 pt-12 pb-24 flex flex-col items-center justify-center min-h-[50vh]">
        <Loader2 size={48} className="animate-spin text-indigo-500 mb-4" />
        <p className="text-white/60">Preparing your speaking task...</p>
      </div>
    );
  }

  if (error || !task) {
    return (
      <div className="max-w-2xl mx-auto px-6 pt-12 pb-24 flex flex-col items-center justify-center min-h-[50vh]">
        <AlertCircle size={48} className="text-red-500 mb-4" />
        <p className="text-white/80 text-lg mb-8">{error}</p>
        <div className="flex gap-4">
          <button
            onClick={loadTask}
            className="px-6 py-2.5 bg-indigo-500 hover:bg-indigo-600 text-white rounded-full font-medium transition-colors"
          >
            Try Again
          </button>
          <button
            onClick={() => router.push("/speaking")}
            className="px-6 py-2.5 bg-white/10 hover:bg-white/20 text-white rounded-full font-medium transition-colors"
          >
            Back to Speaking
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto px-6 pt-12 pb-24">
      {/* Header */}
      <div className="flex items-center justify-between mb-12">
        <button
          onClick={() => router.push("/speaking")}
          className="flex items-center gap-2 text-white/60 hover:text-white transition-colors"
        >
          <ArrowLeft size={20} />
          <span>Speaking</span>
        </button>
        <div className="px-4 py-1.5 bg-white/10 rounded-full text-sm font-medium text-white/80 uppercase tracking-wider">
          {task.topic}
        </div>
      </div>

      {/* Task Content */}
      <div className="bg-white/5 border border-white/10 rounded-3xl p-8 mb-8">
        <h2 className="text-xs font-bold text-indigo-400 uppercase tracking-widest mb-4 flex items-center gap-2">
          🎒 Your Task
        </h2>
        
        {task.scenario && (
          <div className="mb-6 p-4 bg-indigo-500/10 border border-indigo-500/20 rounded-xl">
            <p className="text-indigo-200 text-sm">{task.scenario}</p>
          </div>
        )}

        <p className="text-2xl font-bold text-white mb-6 leading-relaxed">
          {task.prompt}
        </p>

        {task.instructions.length > 0 && (
          <div className="mb-8">
            <p className="text-white/60 text-sm mb-3">Try to talk about:</p>
            <ul className="space-y-2">
              {task.instructions.map((inst, i) => (
                <li key={i} className="flex items-start gap-3 text-white/90">
                  <span className="text-indigo-400 mt-0.5">•</span>
                  <span>{inst}</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        {((task.hints && task.hints.length > 0) || (task.suggested_vocabulary && task.suggested_vocabulary.length > 0)) && (
          <div className="pt-6 border-t border-white/10">
            <h3 className="text-sm font-bold text-yellow-400 mb-4 flex items-center gap-2">
              💡 Helpful hints & words
            </h3>
            
            {task.hints && task.hints.length > 0 && (
              <ul className="space-y-1 mb-4">
                {task.hints.map((hint, i) => (
                  <li key={i} className="text-sm text-white/70 italic">
                    - {hint}
                  </li>
                ))}
              </ul>
            )}

            {task.suggested_vocabulary && task.suggested_vocabulary.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {task.suggested_vocabulary.map((word, i) => (
                  <span key={i} className="px-3 py-1 bg-white/10 rounded-md text-sm text-white/80">
                    {word}
                  </span>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Action Footer */}
      <div className="flex flex-col items-center gap-4">
        <div className="text-white/40 text-sm flex items-center gap-2">
          <span>⏱</span>
          <span>{task.duration_seconds} seconds</span>
        </div>
        <button
          onClick={handleStartSpeaking}
          disabled={isStarting}
          className="w-full sm:w-auto flex items-center justify-center gap-3 px-8 py-4 bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-400 hover:to-purple-400 text-white rounded-full font-bold text-lg shadow-lg shadow-indigo-500/25 transition-all disabled:opacity-50"
        >
          {isStarting ? (
            <Loader2 size={24} className="animate-spin" />
          ) : (
            <Mic size={24} />
          )}
          <span>Start Speaking</span>
        </button>
      </div>
    </div>
  );
}

export default function SpeakingPreviewPage() {
  return (
    <Suspense fallback={
      <div className="max-w-2xl mx-auto px-6 pt-12 pb-24 flex flex-col items-center justify-center min-h-[50vh]">
        <Loader2 size={48} className="animate-spin text-indigo-500 mb-4" />
        <p className="text-white/60">Preparing your speaking task...</p>
      </div>
    }>
      <SpeakingPreviewContent />
    </Suspense>
  );
}
