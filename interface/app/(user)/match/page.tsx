"use client";

import React from "react";
import MatchDeck from "@/components/match/MatchDeck";

export default function MatchPage() {
  return (
    <div className="min-h-screen bg-[#FDFBF7] flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-[460px] h-[100dvh] md:h-[80vh] flex flex-col items-center relative">
        <div className="flex-1 w-full relative flex flex-col items-center justify-center h-full">
          <MatchDeck />
        </div>
      </div>
    </div>
  );
}
