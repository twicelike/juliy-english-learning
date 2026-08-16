import React from "react";
import { Sparkles } from "lucide-react";

export const SPEAKING_TOPICS = [
  "Travel",
  "Food",
  "Work",
  "Daily Life",
  "Study",
  "Technology",
  "Entertainment",
];

interface TopicSelectorProps {
  selectedTopic: string | null;
  onSelectTopic: (topic: string | null) => void;
}

export function TopicSelector({ selectedTopic, onSelectTopic }: TopicSelectorProps) {
  return (
    <div className="mb-10">
      <h2 className="text-xl font-bold text-white mb-4">What do you want to talk about?</h2>
      
      <div className="flex flex-wrap gap-3">
        {SPEAKING_TOPICS.map((topic) => {
          const isSelected = selectedTopic === topic;
          return (
            <button
              key={topic}
              onClick={() => onSelectTopic(topic)}
              className={`px-5 py-2.5 rounded-full text-sm font-medium transition-all ${
                isSelected
                  ? "bg-indigo-500 text-white shadow-lg shadow-indigo-500/30"
                  : "bg-white/5 text-white/70 hover:bg-white/10 hover:text-white border border-white/5"
              }`}
            >
              {topic}
            </button>
          );
        })}

        <button
          onClick={() => onSelectTopic(null)}
          className={`flex items-center gap-2 px-5 py-2.5 rounded-full text-sm font-medium transition-all ${
            selectedTopic === null
              ? "bg-gradient-to-r from-fuchsia-500 to-cyan-500 text-white shadow-lg shadow-cyan-500/20"
              : "bg-white/5 text-white/70 hover:bg-white/10 hover:text-white border border-white/5"
          }`}
        >
          <Sparkles size={16} className={selectedTopic === null ? "text-yellow-200" : "text-fuchsia-400"} />
          Surprise Me
        </button>
      </div>
    </div>
  );
}
