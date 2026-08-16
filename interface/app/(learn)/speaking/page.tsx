"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { SpeakingHeader } from "@/components/speaking/SpeakingHeader";
import { RecommendedTask } from "@/components/speaking/RecommendedTask";
import { TopicSelector } from "@/components/speaking/TopicSelector";
import { PracticeModes } from "@/components/speaking/PracticeModes";
import { RecentPractice, RecentPracticeItem } from "@/components/speaking/RecentPractice";
import { SpeakingAPI, SpeakingTaskType } from "@/lib/speaking-api";

export default function SpeakingHomePage() {
  const router = useRouter();
  const [selectedTopic, setSelectedTopic] = useState<string | null>(null);
  const [loadingMode, setLoadingMode] = useState<SpeakingTaskType | null>(null);
  const [recentPractices, setRecentPractices] = useState<RecentPracticeItem[]>([]);
  const [isLoadingHistory, setIsLoadingHistory] = useState(true);

  // Note: Streak should ideally come from global state/auth store. 
  // We're hardcoding for SP1 UI purpose.
  const streak = 7; 

  useEffect(() => {
    SpeakingAPI.getRecentPractices()
      .then((data) => {
        setRecentPractices(data);
      })
      .catch((err) => console.error(err))
      .finally(() => setIsLoadingHistory(false));
  }, []);

  const handleStartPractice = (type: SpeakingTaskType) => {
    if (loadingMode) return;
    setLoadingMode(type);

    try {
      const topicParam = selectedTopic ? encodeURIComponent(selectedTopic) : "random";
      router.push(`/speaking/preview?topic=${topicParam}&type=${type}`);
    } finally {
      // Just visually clear after a short delay since we are navigating
      setTimeout(() => setLoadingMode(null), 1000);
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-6 pt-12 pb-24">
      <SpeakingHeader streak={streak} />
      
      <RecommendedTask />
      
      <TopicSelector
        selectedTopic={selectedTopic}
        onSelectTopic={setSelectedTopic}
      />
      
      <PracticeModes
        onStart={handleStartPractice}
        loadingMode={loadingMode}
      />
      
      {!isLoadingHistory && (
        <RecentPractice items={recentPractices} />
      )}
    </div>
  );
}
