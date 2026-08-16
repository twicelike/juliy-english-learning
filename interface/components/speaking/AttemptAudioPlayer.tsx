import { useAttemptAudio } from "@/hooks/use-attempt-audio";
import { Loader2, Play, Pause, AlertCircle } from "lucide-react";
import { useState, useRef, useEffect } from "react";

export function AttemptAudioPlayer({ attemptId }: { attemptId: string }) {
  const { audioUrl, isLoading, error } = useAttemptAudio(attemptId);
  const audioRef = useRef<HTMLAudioElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const updateProgress = () => {
      setProgress((audio.currentTime / (audio.duration || 1)) * 100);
    };

    const handleEnded = () => setIsPlaying(false);

    audio.addEventListener("timeupdate", updateProgress);
    audio.addEventListener("ended", handleEnded);

    return () => {
      audio.removeEventListener("timeupdate", updateProgress);
      audio.removeEventListener("ended", handleEnded);
    };
  }, [audioUrl]);

  const togglePlay = () => {
    if (!audioRef.current) return;
    if (isPlaying) {
      audioRef.current.pause();
    } else {
      audioRef.current.play();
    }
    setIsPlaying(!isPlaying);
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-3 p-3 bg-white/5 rounded-xl text-white/50 text-sm">
        <Loader2 size={16} className="animate-spin" />
        <span>Loading recording...</span>
      </div>
    );
  }

  if (error || !audioUrl) {
    return (
      <div className="flex items-center gap-3 p-3 bg-red-500/10 rounded-xl text-red-400 text-sm">
        <AlertCircle size={16} />
        <span>{error || "Audio not available"}</span>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-4 p-3 bg-white/5 border border-white/10 rounded-xl">
      <button
        onClick={togglePlay}
        className="w-10 h-10 flex shrink-0 items-center justify-center bg-indigo-500 hover:bg-indigo-600 rounded-full text-white transition-colors"
      >
        {isPlaying ? <Pause size={18} className="fill-current" /> : <Play size={18} className="fill-current ml-1" />}
      </button>
      
      <div className="flex-1 relative h-2 bg-white/10 rounded-full overflow-hidden cursor-pointer"
        onClick={(e) => {
          if (!audioRef.current) return;
          const rect = e.currentTarget.getBoundingClientRect();
          const x = e.clientX - rect.left;
          const percentage = x / rect.width;
          audioRef.current.currentTime = percentage * audioRef.current.duration;
        }}
      >
        <div 
          className="absolute top-0 left-0 h-full bg-indigo-500 transition-all duration-100 ease-linear"
          style={{ width: `${progress}%` }}
        />
      </div>

      <audio ref={audioRef} src={audioUrl} className="hidden" />
    </div>
  );
}
