"use client";

import React, { useState, useEffect } from "react";
import MatchCard, { MatchCardData } from "./MatchCard";
import { motion, AnimatePresence } from "framer-motion";
import apiClient from "@/lib/api-client";

export default function MatchDeck() {
  const [cards, setCards] = useState<MatchCardData[]>([]);
  const [loading, setLoading] = useState(true);
  const [sessionKept, setSessionKept] = useState(0);
  const [sessionPassed, setSessionPassed] = useState(0);
  const [showSummary, setShowSummary] = useState(false);
  const [showFeedback, setShowFeedback] = useState(false);

  const fetchCards = async () => {
    setLoading(true);
    setShowSummary(false);
    try {
      const res = await apiClient.get("/match/recommendations");
      if (res.data?.success) {
        setCards(Array.isArray(res.data.data) ? res.data.data : []);
      }
    } catch (err) {
      console.error("Failed to fetch match recommendations", err);
      setCards([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // In dev mode or without a token this will fail, 
    // you might want to add mock data here if no token for testing
    fetchCards();
  }, []);

  const handleSwipe = async (action: "keep" | "pass", card: MatchCardData) => {
    // Remove top card immediately
    setCards((prev) => prev.slice(1));

    if (action === "keep") {
      setSessionKept((prev) => prev + 1);
      // Show micro-animation
      setShowFeedback(true);
      setTimeout(() => setShowFeedback(false), 800);
    } else {
      setSessionPassed((prev) => prev + 1);
    }

    try {
      await apiClient.post("/match/swipe", {
        word_id: card.word_id,
        action: action,
      });
    } catch (err) {
      console.error("Failed to record swipe", err);
    }

    // Check if we need to show summary or fetch next batch
    if (cards.length <= 1) {
      setShowSummary(true);
    }
  };

  const totalCards = cards.length + sessionKept + sessionPassed;
  const currentCardIndex = sessionKept + sessionPassed + 1;

  if (loading && cards.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center font-handwritten text-xl text-gray-500 animate-pulse">
        Shuffling the deck...
      </div>
    );
  }

  if (showSummary) {
    return (
      <div className="w-full max-w-md bg-white rounded-3xl p-8 flex flex-col items-center justify-center text-center shadow-sm border-2 border-gray-800"
        style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}
      >
        <h2 className="text-3xl font-bold font-handwritten mb-6">You've discovered {sessionKept + sessionPassed} words!</h2>
        
        <div className="space-y-4 mb-8 w-full">
          <div className="flex justify-between items-center text-xl font-handwritten px-4 py-2 border-2 border-green-200 bg-green-50 rounded-xl"
            style={{ borderRadius: "15px 225px 15px 255px/255px 15px 225px 15px" }}>
            <span>♥ Kept</span>
            <span className="font-bold text-green-600">{sessionKept}</span>
          </div>
          <div className="flex justify-between items-center text-xl font-handwritten px-4 py-2 border-2 border-gray-200 bg-gray-50 rounded-xl"
             style={{ borderRadius: "225px 15px 255px 15px/15px 255px 15px 225px" }}>
            <span>× Passed</span>
            <span className="font-bold text-gray-600">{sessionPassed}</span>
          </div>
        </div>

        <div className="flex flex-col gap-4 w-full">
          <button 
            className="w-full py-4 bg-gray-900 text-white font-handwritten text-xl rounded-xl hover:bg-gray-800 transition-colors border-2 border-black"
            style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}
            onClick={() => window.location.href = '/learning/session'}
          >
            Learn kept words
          </button>
          
          <button 
            className="w-full py-4 bg-transparent text-gray-900 font-handwritten text-xl rounded-xl hover:bg-gray-100 transition-colors border-2 border-gray-800"
            style={{ borderRadius: "15px 225px 15px 255px/255px 15px 225px 15px" }}
            onClick={fetchCards}
          >
            Keep discovering
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full flex flex-col items-center">
      <div className="w-full px-8 mb-4">
        <div className="flex justify-between items-center font-handwritten text-gray-500 mb-2">
          <span>Discover words you'll want to know</span>
          {totalCards > 0 && <span>{Math.min(currentCardIndex, totalCards)} / {totalCards}</span>}
        </div>
        {totalCards > 0 && (
          <div className="w-full h-1 bg-gray-200 relative" style={{ borderRadius: "5px" }}>
            <div 
              className="absolute top-0 left-0 h-full bg-gray-800 transition-all duration-300"
              style={{ width: `${Math.min((currentCardIndex / totalCards) * 100, 100)}%`, borderRadius: "5px" }}
            />
          </div>
        )}
      </div>

      <div className="relative w-full max-w-md h-[500px] mt-8 flex justify-center perspective-[1000px]">
        <AnimatePresence>
          {showFeedback && (
            <motion.div
              initial={{ opacity: 0, y: -20, scale: 0.8 }}
              animate={{ opacity: 1, y: -40, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8 }}
              className="absolute top-0 z-50 text-green-500 font-handwritten font-bold text-xl flex items-center gap-2 bg-white px-4 py-2 border-2 border-green-500 shadow-md"
              style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}
            >
              <span className="text-2xl">✨</span> ✓ Saved to your words
            </motion.div>
          )}
        </AnimatePresence>

        {cards.length > 0 ? (
          cards.slice(0, 3).reverse().map((card, idx, arr) => (
            <MatchCard
              key={card.word_id}
              card={card}
              onSwipe={handleSwipe}
              isTop={idx === arr.length - 1}
              index={arr.length - 1 - idx}
            />
          ))
        ) : (
          <div className="absolute w-full h-full bg-white rounded-3xl p-6 flex flex-col justify-center items-center shadow-sm border-2 border-gray-800"
            style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}>
            <span className="text-2xl font-handwritten">No more words to discover right now.</span>
            <button onClick={fetchCards} className="mt-8 font-handwritten border-b-2 border-gray-800">Refresh</button>
          </div>
        )}
      </div>
    </div>
  );
}
