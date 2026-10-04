"use client";

import React, { useState, useEffect, useRef } from "react";
import MatchCard, { MatchCardData } from "./MatchCard";
import { motion, AnimatePresence } from "framer-motion";
import apiClient from "@/lib/api-client";
import { WordSetAPI } from "@/lib/wordset-api";

interface Props {
  selectedSetIds?: string[];
  onWordSetsUpdated?: () => void;
}

export default function MatchDeck({ selectedSetIds = [], onWordSetsUpdated }: Props) {
  const [cards, setCards] = useState<MatchCardData[]>([]);
  const [loading, setLoading] = useState(true);
  const [sessionKept, setSessionKept] = useState(0);
  const [sessionPassed, setSessionPassed] = useState(0);
  const [showFeedback, setShowFeedback] = useState(false);
  const isFetchingRef = useRef(false);

  const fetchCards = async (append = false) => {
    if (isFetchingRef.current) return;
    isFetchingRef.current = true;
    if (!append) setLoading(true);
    try {
      const res = await apiClient.get("/match/recommendations");
      if (res.data?.success) {
        const newCards = Array.isArray(res.data.data) ? res.data.data : [];
        if (append) {
          setCards(prev => {
            // Filter out cards that are already in the deck to prevent duplicates
            // (e.g., if API returns words we just swiped but haven't been saved in DB yet)
            const uniqueNewCards = newCards.filter(
              (newCard: MatchCardData) => !prev.some(p => p.word_id === newCard.word_id)
            );
            return [...prev, ...uniqueNewCards];
          });
        } else {
          setCards(newCards);
        }
      }
    } catch (err) {
      console.error("Failed to fetch match recommendations", err);
      if (!append) setCards([]);
    } finally {
      setLoading(false);
      isFetchingRef.current = false;
    }
  };

  useEffect(() => {
    // In dev mode or without a token this will fail, 
    // you might want to add mock data here if no token for testing
    fetchCards();
  }, []);

  const handleSwipe = async (action: "keep" | "pass", card: MatchCardData) => {
    // Remove top card immediately and prefetch if running low
    setCards((prev) => {
      const newCards = prev.slice(1);
      if (newCards.length <= 3 && !isFetchingRef.current) {
        fetchCards(true);
      }
      return newCards;
    });

    if (action === "keep") {
      setSessionKept((prev) => prev + 1);

      if (selectedSetIds.length === 0) {
        // Do not record the swipe to the API and do not show save feedback.
        // The word will reappear in future sessions.
      } else {
        setShowFeedback(true);
        setTimeout(() => setShowFeedback(false), 800);

        try {
          // 1. Add to all selected word sets
          for (const setId of selectedSetIds) {
            try {
              await WordSetAPI.bulkAddWords(setId, [{ word_id: card.word_id, sense_id: card.sense_id }]);
            } catch (setErr) {
              console.warn(`Failed to save to set ${setId}, it might be full:`, setErr);
            }
          }

          // Trigger sidebar refresh
          if (onWordSetsUpdated) onWordSetsUpdated();

          // 2. Record swipe so it doesn't show in match again
          await apiClient.post("/match/swipe", {
            word_id: card.word_id,
            action: action,
          });
        } catch (err) {
          console.error("Failed to save word or record swipe", err);
        }
      }
    } else {
      setSessionPassed((prev) => prev + 1);
      try {
        await apiClient.post("/match/swipe", {
          word_id: card.word_id,
          action: action,
        });
      } catch (err) {
        console.error("Failed to record swipe pass", err);
      }
    }

    // Note: Infinite scrolling, so we don't show a summary break anymore.
  };

  const totalDiscovered = sessionKept + sessionPassed;

  if (loading && cards.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center font-handwritten text-xl text-gray-500 animate-pulse">
        Shuffling the deck...
      </div>
    );
  }

  return (
    <div className="w-full h-full flex flex-col items-center">
      <div className="w-full px-8 mb-4">
        <div className="flex justify-between items-center font-handwritten text-gray-500 mb-2">
          <span>Discover words you'll want to know</span>
          {totalDiscovered > 0 && <span>Discovered: {totalDiscovered}</span>}
        </div>
      </div>

      <div className="relative w-full max-w-md flex-1 min-h-[500px] mt-8 flex justify-center perspective-[1000px]">
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
            <button onClick={() => fetchCards(false)} className="mt-8 font-handwritten border-b-2 border-gray-800">Refresh</button>
          </div>
        )}
      </div>
    </div>
  );
}
