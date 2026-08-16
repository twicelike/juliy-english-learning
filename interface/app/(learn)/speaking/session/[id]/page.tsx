"use client";

import { useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Mic, Pause, Play, Check, AlertCircle, Loader2, HelpCircle, Trophy, RefreshCcw } from "lucide-react";
import { SpeakingAPI, SpeakingSessionDetails, SpeakingAttempt } from "@/lib/speaking-api";
import { UserAPI } from "@/lib/user-api";
import { useAudioRecorder } from "@/hooks/use-audio-recorder";
import { AttemptAudioPlayer } from "@/components/speaking/AttemptAudioPlayer";

// Formats duration in seconds to MM:SS
const formatDuration = (seconds: number) => {
  const m = Math.floor(seconds / 60).toString().padStart(2, "0");
  const s = (seconds % 60).toString().padStart(2, "0");
  return `${m}:${s}`;
};

export default function SpeakingSessionRoom() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.id as string;

  const [sessionDetails, setSessionDetails] = useState<SpeakingSessionDetails | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [showHelp, setShowHelp] = useState(false);
  
  const [attemptStatus, setAttemptStatus] = useState<"idle" | "uploading" | "processing" | "completed" | "failed">("idle");
  const [finalAttempt, setFinalAttempt] = useState<SpeakingAttempt | null>(null);
  const [pollError, setPollError] = useState<string | null>(null);
  
  const [savedWords, setSavedWords] = useState<Set<string>>(new Set());

  // For visual volume effect
  const [volume, setVolume] = useState(0);

  const {
    status,
    duration,
    audioBlob,
    error: recorderError,
    start,
    pause,
    resume,
    stop,
    reset,
  } = useAudioRecorder({
    onVolumeChange: (vol) => setVolume(vol),
  });

  const loadSession = async () => {
    setIsLoading(true);
    setFetchError(null);
    try {
      const details = await SpeakingAPI.getSessionDetails(sessionId);
      if (details.session.status === "completed") {
        setFetchError("This session has already been completed.");
      } else {
        setSessionDetails(details);
      }
    } catch (err: any) {
      console.error(err);
      setFetchError("Unable to load speaking session.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (sessionId) {
      loadSession();
    }
  }, [sessionId]);

  const handleFinish = () => {
    stop(); // Stop recording, wait for blob
    setAttemptStatus("uploading");
  };

  useEffect(() => {
    let active = true;

    const processAudio = async () => {
      if (!audioBlob || attemptStatus !== "uploading") return;
      
      try {
        // Submit the audio
        const submitRes = await SpeakingAPI.submitAttempt(sessionId, audioBlob, duration * 1000);
        const attemptId = submitRes.attempt_id;
        
        if (!active) return;
        setAttemptStatus("processing");

        // Poll for results
        const pollInterval = setInterval(async () => {
          try {
            const attempt = await SpeakingAPI.getAttempt(attemptId);
            if (!active) {
              clearInterval(pollInterval);
              return;
            }
            if (attempt.status === "completed") {
              clearInterval(pollInterval);
              setFinalAttempt(attempt);
              setAttemptStatus("completed");
              loadSession(); // reload session to get updated attempts array
            } else if (attempt.status === "failed") {
              clearInterval(pollInterval);
              setPollError(attempt.error_message || "Failed to process audio.");
              setAttemptStatus("failed");
            }
          } catch (err) {
            console.error("Polling error:", err);
            // Optionally we can retry a few times before failing
          }
        }, 2000);

        return () => clearInterval(pollInterval);

      } catch (err: any) {
        if (!active) return;
        setPollError(err.message || "Failed to upload recording.");
        setAttemptStatus("failed");
      }
    };

    const cleanup = processAudio();
    return () => {
      active = false;
      cleanup.then(c => c && c());
    };
  }, [audioBlob, attemptStatus, sessionId, duration]);

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto px-6 pt-12 pb-24 flex flex-col items-center justify-center min-h-[50vh]">
        <Loader2 size={48} className="animate-spin text-indigo-500 mb-4" />
        <p className="text-white/60">Loading session...</p>
      </div>
    );
  }

  if (fetchError || !sessionDetails) {
    return (
      <div className="max-w-2xl mx-auto px-6 pt-12 pb-24 flex flex-col items-center justify-center min-h-[50vh]">
        <AlertCircle size={48} className="text-red-500 mb-4" />
        <p className="text-white/80 text-lg mb-8">{fetchError}</p>
        <div className="flex gap-4">
          <button
            onClick={loadSession}
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

  const handleSaveVocabulary = async (word: string) => {
    try {
      await UserAPI.addBookmark({ word, type: "word" });
      setSavedWords((prev) => new Set(prev).add(word));
    } catch (err) {
      console.error("Failed to save vocabulary", err);
    }
  };

  const snapshot = sessionDetails.session.task_snapshot;
  const isBeforeRecording = status === "idle" || status === "requesting_permission";
  const isRecording = status === "recording";
  const isPaused = status === "paused";

  // When finished and showing success/result screen
  if (attemptStatus === "completed" && finalAttempt?.evaluation) {
    const ev = finalAttempt.evaluation;
    const completedAttempts = (sessionDetails.attempts || []).filter(a => a.status === "completed" && a.evaluation);
    
    return (
      <div className="max-w-2xl mx-auto px-6 pt-16 pb-24 flex flex-col items-center min-h-[85vh]">
        <div className="w-20 h-20 bg-green-500/20 rounded-full flex items-center justify-center mb-6">
          <Trophy size={40} className="text-green-400" />
        </div>
        <h2 className="text-3xl font-bold text-white mb-2">Evaluation Complete</h2>
        <p className="text-lg text-white/70 mb-10">You scored <span className="text-indigo-400 font-bold">{ev.overall}/100</span></p>

        {/* Progress Timeline */}
        {completedAttempts.length > 0 && (
          <div className="w-full bg-white/5 border border-white/10 rounded-2xl p-6 mb-8 text-left">
            <h3 className="text-white/60 font-bold mb-6 text-sm uppercase tracking-wider">Your Progress</h3>
            <div className="flex items-center justify-between">
              {completedAttempts.map((atmpt, i, arr) => (
                <div key={atmpt.id} className="flex flex-col items-center flex-1 relative group">
                  <div className={`w-12 h-12 rounded-full flex items-center justify-center font-bold text-base border-2 relative z-10 transition-colors ${atmpt.id === finalAttempt.id ? 'border-indigo-400 bg-indigo-500 text-white shadow-[0_0_15px_rgba(99,102,241,0.5)]' : 'border-white/20 bg-[#1a1b26] text-white/60 group-hover:border-white/40 group-hover:text-white/80'}`}>
                    {atmpt.evaluation?.overall}
                  </div>
                  <div className={`mt-3 text-xs font-medium ${atmpt.id === finalAttempt.id ? 'text-indigo-300' : 'text-white/40'}`}>
                    Attempt {i + 1}
                  </div>
                  {i < arr.length - 1 && (
                    <div className="absolute top-6 left-1/2 w-full h-[2px] bg-white/10 -z-0"></div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="w-full bg-white/5 border border-white/10 rounded-2xl p-6 mb-8">
          <div className="mb-8">
            <h3 className="text-white/60 font-bold mb-4 text-sm uppercase tracking-wider">Your Recording</h3>
            <AttemptAudioPlayer attemptId={finalAttempt.id} />
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <div className="text-center p-3 bg-white/5 rounded-xl border border-white/5">
              <div className="text-sm text-white/50 mb-1">Fluency</div>
              <div className="text-xl font-bold text-white">{ev.fluency}</div>
            </div>
            <div className="text-center p-3 bg-white/5 rounded-xl border border-white/5">
              <div className="text-sm text-white/50 mb-1">Grammar</div>
              <div className="text-xl font-bold text-white">{ev.grammar}</div>
            </div>
            <div className="text-center p-3 bg-white/5 rounded-xl border border-white/5">
              <div className="text-sm text-white/50 mb-1">Vocabulary</div>
              <div className="text-xl font-bold text-white">{ev.vocabulary}</div>
            </div>
            <div className="text-center p-3 bg-white/5 rounded-xl border border-white/5">
              <div className="text-sm text-white/50 mb-1">Pronunciation</div>
              <div className="text-xl font-bold text-white">{ev.pronunciation}</div>
            </div>
          </div>

          <div className="space-y-6 text-left">
            {ev.strengths && ev.strengths.length > 0 && (
              <div>
                <h3 className="text-green-400 font-bold mb-2 flex items-center gap-2">
                  <Check size={18} /> Strengths
                </h3>
                <ul className="space-y-1 text-white/80 text-sm list-disc list-inside">
                  {ev.strengths.map((s, i) => <li key={i}>{s}</li>)}
                </ul>
              </div>
            )}
            
            {ev.improvements && ev.improvements.length > 0 && (
              <div>
                <h3 className="text-yellow-400 font-bold mb-2 flex items-center gap-2">
                  <RefreshCcw size={18} /> Areas to Improve
                </h3>
                <ul className="space-y-1 text-white/80 text-sm list-disc list-inside">
                  {ev.improvements.map((s, i) => <li key={i}>{s}</li>)}
                </ul>
              </div>
            )}

            {ev.new_used_vocabulary && ev.new_used_vocabulary.length > 0 && (
              <div className="pt-4 border-t border-white/10">
                <h3 className="text-purple-400 font-bold mb-3 flex items-center gap-2">
                  ✨ Vocabulary you used!
                </h3>
                <div className="grid gap-3">
                  {ev.new_used_vocabulary.map((vocab, i) => {
                    const isSaved = savedWords.has(vocab.word);
                    return (
                      <div key={i} className="flex items-center justify-between bg-white/5 border border-white/10 p-3 rounded-xl">
                        <div>
                          <div className="font-bold text-white mb-0.5">{vocab.word}</div>
                          <div className="text-xs text-white/60">{vocab.reason}</div>
                        </div>
                        <button
                          onClick={() => !isSaved && handleSaveVocabulary(vocab.word)}
                          disabled={isSaved}
                          className={`ml-4 shrink-0 px-3 py-1.5 text-xs font-bold rounded-lg transition-colors ${
                            isSaved 
                              ? 'bg-green-500/20 text-green-400 border border-green-500/30' 
                              : 'bg-white/10 hover:bg-white/20 text-white'
                          }`}
                        >
                          {isSaved ? "Saved" : "Save"}
                        </button>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {finalAttempt.transcript && (
              <div className="pt-4 border-t border-white/10">
                <h3 className="text-white/60 font-bold mb-2 text-sm uppercase tracking-wider">Transcript</h3>
                <p className="text-white/80 text-sm leading-relaxed bg-black/20 p-4 rounded-xl italic">
                  "{finalAttempt.transcript}"
                </p>
              </div>
            )}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 w-full max-w-md">
          <button
            onClick={() => {
              setAttemptStatus("idle");
              setFinalAttempt(null);
              setPollError(null);
              reset(); // Reset audio hook
              loadSession();
            }}
            className="flex-1 px-8 py-3 bg-indigo-500 hover:bg-indigo-600 text-white rounded-full font-bold transition-all shadow-lg shadow-indigo-500/25 flex items-center justify-center gap-2"
          >
            <RefreshCcw size={18} /> Try Again
          </button>
          <button
            onClick={() => router.push("/speaking")}
            className="flex-1 px-8 py-3 bg-white/10 hover:bg-white/20 text-white rounded-full font-bold transition-all flex items-center justify-center gap-2"
          >
            <ArrowLeft size={18} /> Back
          </button>
        </div>
      </div>
    );
  }

  if (attemptStatus === "failed") {
    return (
      <div className="max-w-2xl mx-auto px-6 pt-24 pb-24 flex flex-col items-center justify-center text-center">
        <AlertCircle size={48} className="text-red-500 mb-4" />
        <h2 className="text-3xl font-bold text-white mb-4">Analysis Failed</h2>
        <p className="text-lg text-white/80 mb-8">{pollError}</p>
        <button
          onClick={() => router.push("/speaking")}
          className="px-8 py-3 bg-white/10 hover:bg-white/20 text-white rounded-full font-bold transition-all"
        >
          Back to Speaking
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto px-6 pt-8 pb-24 flex flex-col min-h-[85vh]">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <button
          onClick={() => router.push("/speaking")}
          className="flex items-center gap-2 text-white/60 hover:text-white transition-colors"
        >
          <ArrowLeft size={20} />
          <span className="capitalize">{snapshot.topic}</span>
        </button>
        <div className={`font-mono text-xl font-bold ${isRecording ? "text-red-400" : "text-white/80"}`}>
          {formatDuration(duration)}
        </div>
      </div>

      {recorderError && (
        <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl mb-6 text-sm flex items-start gap-3">
          <AlertCircle size={18} className="mt-0.5 shrink-0" />
          <p>{recorderError}</p>
        </div>
      )}

      {/* Task Content */}
      <div className="flex-1 flex flex-col items-center justify-center text-center mb-12">
        <h2 className="text-xs font-bold text-indigo-400 uppercase tracking-widest mb-6">
          {isPaused ? "PAUSED" : snapshot.task_type === "free_talk" ? "💬 FREE TALK" : "YOUR TASK"}
        </h2>
        
        <p className="text-2xl sm:text-3xl font-bold text-white leading-relaxed mb-8 max-w-lg mx-auto">
          {snapshot.prompt}
        </p>

        {snapshot.task_type !== "free_talk" && (
          <button
            onClick={() => setShowHelp(!showHelp)}
            className="flex items-center gap-2 px-5 py-2.5 bg-white/5 hover:bg-white/10 rounded-full text-white/70 text-sm font-medium transition-colors"
          >
            <HelpCircle size={16} />
            {showHelp ? "Hide Help" : "💡 Help"}
          </button>
        )}

        {showHelp && snapshot.task_type !== "free_talk" && (
          <div className="w-full max-w-lg mt-6 bg-white/5 border border-white/10 rounded-2xl p-6 text-left">
            {snapshot.instructions.length > 0 && (
              <div className="mb-6 last:mb-0">
                <h4 className={`text-sm font-bold uppercase tracking-wider mb-3 ${snapshot.task_type === 'scenario' ? 'text-indigo-400' : 'text-white/60'}`}>
                  {snapshot.task_type === "scenario" ? "🎯 YOUR GOAL" : "Instructions"}
                </h4>
                <ul className="space-y-2">
                  {snapshot.instructions.map((inst, i) => (
                    <li key={i} className="flex items-start gap-2 text-white/90 text-sm">
                      <span className="text-indigo-400 mt-0.5">•</span>
                      <span>{inst}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
            
            {((snapshot.hints && snapshot.hints.length > 0) || (snapshot.suggested_vocabulary && snapshot.suggested_vocabulary.length > 0)) && (
              <div>
                <h4 className="text-sm font-bold text-yellow-500/80 uppercase tracking-wider mb-3">Helpful Words</h4>
                <div className="flex flex-wrap gap-2">
                  {snapshot.hints?.map((hint, i) => (
                    <span key={`hint-${i}`} className="px-2.5 py-1 bg-yellow-500/10 border border-yellow-500/20 rounded-md text-xs text-yellow-200">
                      {hint}
                    </span>
                  ))}
                  {snapshot.suggested_vocabulary?.map((word, i) => (
                    <span key={`vocab-${i}`} className="px-2.5 py-1 bg-white/10 border border-white/10 rounded-md text-xs text-white/80">
                      {word}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Recording Controls */}
      <div className="flex flex-col items-center">
        {attemptStatus === "uploading" || attemptStatus === "processing" ? (
          <div className="flex flex-col items-center gap-4 text-indigo-400">
            <Loader2 size={32} className="animate-spin" />
            <p className="font-medium animate-pulse">
              {attemptStatus === "uploading" ? "Uploading recording..." : "AI is analyzing your speaking..."}
            </p>
          </div>
        ) : isBeforeRecording ? (
          <button
            onClick={start}
            disabled={status === "requesting_permission"}
            className="group flex items-center justify-center gap-3 px-10 py-5 bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-400 hover:to-purple-400 text-white rounded-full font-bold text-xl shadow-xl shadow-indigo-500/25 transition-all disabled:opacity-50"
          >
            {status === "requesting_permission" ? (
              <Loader2 size={28} className="animate-spin" />
            ) : (
              <Mic size={28} className="group-hover:scale-110 transition-transform" />
            )}
            <span>Start Speaking</span>
          </button>
        ) : (
          <div className="w-full flex flex-col items-center gap-8">
            {/* Waveform Visualization (CSS based) */}
            <div className="flex items-center gap-1.5 h-16">
              {[...Array(9)].map((_, i) => {
                // Calculate scale based on volume and index to make a wave shape
                const distance = Math.abs(4 - i);
                const maxScale = 1 - (distance * 0.15); // Center is highest
                
                // Base height + volume reactivity
                const height = isRecording ? Math.max(15, 100 * maxScale * (0.3 + volume * 1.5)) : 10;
                
                return (
                  <div
                    key={i}
                    className={`w-2 rounded-full transition-all duration-75 ${isRecording ? 'bg-indigo-400' : 'bg-white/20'}`}
                    style={{ height: `${height}%` }}
                  />
                );
              })}
            </div>

            {/* Actions */}
            <div className="flex items-center gap-6">
              {isRecording ? (
                <button
                  onClick={pause}
                  className="flex items-center gap-2 px-6 py-3 bg-white/10 hover:bg-white/20 text-white rounded-full font-medium transition-colors"
                >
                  <Pause size={20} />
                  <span>Pause</span>
                </button>
              ) : isPaused ? (
                <button
                  onClick={resume}
                  className="flex items-center gap-2 px-6 py-3 bg-indigo-500/20 hover:bg-indigo-500/30 text-indigo-300 rounded-full font-medium border border-indigo-500/30 transition-colors"
                >
                  <Play size={20} className="fill-current" />
                  <span>Resume</span>
                </button>
              ) : null}

              <button
                onClick={handleFinish}
                className="flex items-center gap-2 px-8 py-3 bg-white text-indigo-900 hover:bg-indigo-50 rounded-full font-bold transition-colors shadow-lg shadow-white/10"
              >
                <Check size={20} />
                <span>I'm Done</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
