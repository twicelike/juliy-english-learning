"use client";

import { useState } from "react";

interface AudioButtonProps {
  audioUrl: string;
}

export default function AudioButton({ audioUrl }: AudioButtonProps) {
  const [isPlaying, setIsPlaying] = useState(false);

  const playAudio = (e: React.MouseEvent) => {
    e.preventDefault();
    if (!audioUrl || isPlaying) return;
    
    setIsPlaying(true);
    
    // If the audioUrl is just a filename like "En-uk-free.ogg", prefix it with Wikimedia Commons path
    const finalUrl = audioUrl.startsWith("http") 
      ? audioUrl 
      : `https://commons.wikimedia.org/wiki/Special:FilePath/${encodeURIComponent(audioUrl)}`;
      
    const audio = new Audio(finalUrl);
    
    audio.onended = () => setIsPlaying(false);
    audio.onerror = () => setIsPlaying(false);
    
    audio.play().catch((err) => {
      console.error("Audio playback failed", err);
      setIsPlaying(false);
    });
  };

  return (
    <button
      onClick={playAudio}
      disabled={isPlaying}
      className={`p-1.5 rounded-full border transition-all ${
        isPlaying 
          ? "bg-indigo-500/20 border-indigo-500/50 text-indigo-400" 
          : "bg-white/5 border-white/10 text-white/60 hover:bg-white/10 hover:text-white"
      }`}
      title="Phát âm thanh"
    >
      <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
        {isPlaying ? (
          <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" />
        ) : (
          <path d="M8 5v14l11-7z" />
        )}
      </svg>
    </button>
  );
}
