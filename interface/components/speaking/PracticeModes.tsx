import React from "react";
import { Zap, Theater, MessageCircle, Loader2 } from "lucide-react";
import { SpeakingTaskType } from "@/lib/speaking-api";

interface PracticeModeCardProps {
  type: SpeakingTaskType;
  title: string;
  description: string;
  icon: React.ReactNode;
  gradient: string;
  onStart: (type: SpeakingTaskType) => void;
  isLoading: boolean;
}

function PracticeModeCard({ type, title, description, icon, gradient, onStart, isLoading }: PracticeModeCardProps) {
  return (
    <button
      onClick={() => onStart(type)}
      disabled={isLoading}
      className={`relative flex flex-col items-start gap-4 p-6 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-all text-left overflow-hidden group ${isLoading ? "opacity-50 cursor-not-allowed" : ""}`}
    >
      <div className={`absolute inset-0 bg-gradient-to-br ${gradient} opacity-0 group-hover:opacity-10 transition-opacity`} />
      
      <div className="flex items-center justify-center w-12 h-12 rounded-full bg-white/10 text-white group-hover:scale-110 transition-transform">
        {icon}
      </div>
      
      <div>
        <h3 className="text-xl font-bold text-white mb-1 flex items-center gap-2">
          {title}
          {isLoading && <Loader2 size={16} className="animate-spin text-white/50" />}
        </h3>
        <p className="text-white/50 text-sm">{description}</p>
      </div>
    </button>
  );
}

interface PracticeModesProps {
  onStart: (type: SpeakingTaskType) => void;
  loadingMode: SpeakingTaskType | null;
}

export function PracticeModes({ onStart, loadingMode }: PracticeModesProps) {
  return (
    <div className="mb-12">
      <h2 className="text-xl font-bold text-white mb-4">Choose your mode</h2>
      
      {/* Mobile: stack cards (grid-cols-1), Desktop: grid-cols-3 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        <PracticeModeCard
          type="quick_practice"
          title="Quick Practice"
          description="Answer a single question. Fast and easy."
          icon={<Zap size={24} className="text-yellow-400" />}
          gradient="from-yellow-400 to-orange-500"
          onStart={onStart}
          isLoading={loadingMode === "quick_practice"}
        />
        
        <PracticeModeCard
          type="scenario"
          title="Scenario"
          description="Roleplay a real-life situation."
          icon={<Theater size={24} className="text-indigo-400" />}
          gradient="from-indigo-400 to-purple-500"
          onStart={onStart}
          isLoading={loadingMode === "scenario"}
        />
        
        <PracticeModeCard
          type="free_talk"
          title="Free Talk"
          description="Practice conversation freely on any topic."
          icon={<MessageCircle size={24} className="text-emerald-400" />}
          gradient="from-emerald-400 to-teal-500"
          onStart={onStart}
          isLoading={loadingMode === "free_talk"}
        />
      </div>
    </div>
  );
}
