"use client";

import React, { useState } from "react";
import MatchDeck from "@/components/match/MatchDeck";
import MatchWordSetsSidebar from "@/components/match/MatchWordSetsSidebar";

export default function MatchPage() {
  const [selectedSetIds, setSelectedSetIds] = useState<string[]>([]);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  const handleToggleSet = (id: string) => {
    setSelectedSetIds(prev => 
      prev.includes(id) ? prev.filter(setId => setId !== id) : [...prev, id]
    );
  };

  const handleWordSetsUpdated = () => {
    setRefreshTrigger(prev => prev + 1);
  };

  return (
    <div className="min-h-screen bg-[#FDFBF7] flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-5xl h-[100dvh] md:h-[80vh] flex justify-center gap-8 relative">
        <div className="flex-1 max-w-[460px] relative flex flex-col items-center justify-center h-full">
          <MatchDeck 
            selectedSetIds={selectedSetIds} 
            onWordSetsUpdated={handleWordSetsUpdated} 
          />
        </div>
        
        {/* Sidebar Component handles its own desktop/mobile responsive rendering */}
        <MatchWordSetsSidebar 
          selectedSetIds={selectedSetIds} 
          onToggleSet={handleToggleSet}
          refreshTrigger={refreshTrigger}
        />
      </div>
    </div>
  );
}
