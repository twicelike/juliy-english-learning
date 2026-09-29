"use client";

import React, { useState } from "react";
import { motion, useMotionValue, useTransform, useAnimation, PanInfo } from "framer-motion";
import { Pencil, Coffee, Plane, Heart, X, BookOpen, Star } from "lucide-react";

export interface MatchCardData {
  word_id: string;
  word: string;
  pos: string;
  sense_id: string;
  definition: string;
  translation: string;
  total_swipes?: number;
  match_rate?: number;
}

interface MatchCardProps {
  card: MatchCardData;
  onSwipe: (action: "keep" | "pass", card: MatchCardData) => void;
  isTop: boolean;
  index: number;
}

// Randomly select a doodle based on word length for consistent randomness
const getDoodle = (word: string) => {
  const icons = [Coffee, Plane, Pencil, BookOpen, Star];
  const Icon = icons[word.length % icons.length];
  return <Icon className="w-5 h-5 text-gray-500" />;
};

export default function MatchCard({ card, onSwipe, isTop, index }: MatchCardProps) {
  const x = useMotionValue(0);
  const controls = useAnimation();
  
  // Transform values for rotation and opacity based on drag distance
  const rotate = useTransform(x, [-200, 200], [-15, 15]);
  const opacity = useTransform(x, [-200, -150, 0, 150, 200], [0, 1, 1, 1, 0]);
  const passOpacity = useTransform(x, [-100, -50, 0], [1, 0.5, 0]);
  const keepOpacity = useTransform(x, [0, 50, 100], [0, 0.5, 1]);

  const [isSwiping, setIsSwiping] = useState(false);

  const handleDragEnd = async (e: any, info: PanInfo) => {
    if (isSwiping) return;
    const offset = info.offset.x;
    const velocity = info.velocity.x;

    if (offset > 100 || velocity > 500) {
      // Keep
      setIsSwiping(true);
      await controls.start({ x: 300, opacity: 0, transition: { duration: 0.3 } });
      onSwipe("keep", card);
    } else if (offset < -100 || velocity < -500) {
      // Pass
      setIsSwiping(true);
      await controls.start({ x: -300, opacity: 0, transition: { duration: 0.3 } });
      onSwipe("pass", card);
    } else {
      // Reset
      controls.start({ x: 0, transition: { type: "spring", stiffness: 300, damping: 20 } });
    }
  };

  const handleManualSwipe = async (action: "keep" | "pass") => {
    if (isSwiping) return;
    setIsSwiping(true);
    if (action === "keep") {
      await controls.start({ x: 300, opacity: 0, transition: { duration: 0.3 } });
      onSwipe("keep", card);
    } else {
      await controls.start({ x: -300, opacity: 0, transition: { duration: 0.3 } });
      onSwipe("pass", card);
    }
  };

  // Stacked effect calculations
  const yOffset = isTop ? 0 : index * 8;
  const scale = isTop ? 1 : 1 - index * 0.05;
  const zIndex = 100 - index;

  return (
    <motion.div
      className="absolute w-full h-full flex flex-col justify-center items-center transition-all duration-300"
      style={{
        zIndex,
        y: yOffset,
        scale,
        opacity: isTop ? 1 : Math.max(1 - index * 0.25, 0.3),
        filter: isTop ? "none" : `blur(${index * 2}px)`,
      }}
      animate={controls}
    >
      <motion.div
        drag={isTop ? "x" : false}
        dragConstraints={{ left: 0, right: 0, top: 0, bottom: 0 }}
        onDragEnd={handleDragEnd}
        className={`w-full max-w-md min-h-[300px] h-auto shrink-0 bg-white rounded-3xl p-6 flex flex-col relative shadow-sm border-2 border-gray-800 ${
          isTop ? "cursor-grab active:cursor-grabbing touch-none" : ""
        }`}
        // Wobbly border effect using border-radius
        style={{
          x,
          rotate,
          opacity,
          borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px",
        }}
      >
        {/* Pass Label */}
        <motion.div
          style={{ opacity: passOpacity }}
          className="absolute top-8 right-8 border-4 border-red-500 text-red-500 font-bold text-3xl px-4 py-1 rounded-xl rotate-12 z-10"
        >
          PASS
        </motion.div>

        {/* Keep Label */}
        <motion.div
          style={{ opacity: keepOpacity }}
          className="absolute top-8 left-8 border-4 border-green-500 text-green-500 font-bold text-3xl px-4 py-1 rounded-xl -rotate-12 z-10"
        >
          KEEP
        </motion.div>

        <div className="flex-1 flex flex-col items-center justify-center text-center w-full pt-4 pb-2">
          <h2 className="text-5xl font-bold text-black tracking-wide">
            {card.word}
          </h2>
          <span className="text-black italic text-lg mt-2">{card.pos}</span>

          <div className="mt-8 flex flex-col justify-center w-full space-y-4">
            <p className="text-xl text-black">
              {card.definition}
            </p>
            <p className="text-lg text-black border-t border-black/20 border-dashed pt-4">
              {card.translation}
            </p>
          </div>
        </div>

        <div className="flex justify-between items-end mt-4">
          <div className="text-black">{getDoodle(card.word)}</div>
          
          {card.match_rate !== undefined && card.match_rate > 0 && (
            <div className="flex items-center gap-1 text-black bg-black/5 px-3 py-1 rounded-full text-sm border border-black/20" style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}>
              <Pencil className="w-3 h-3" />
              <span>{card.match_rate}% matched</span>
            </div>
          )}
        </div>
      </motion.div>

      {/* Manual buttons for desktop/accessibility (only show for top card) */}
      {isTop && (
        <div className="absolute -bottom-24 w-full flex justify-center gap-8 z-50">
          <button
            onClick={() => handleManualSwipe("pass")}
            className="w-16 h-16 bg-white border-2 border-gray-800 rounded-full flex items-center justify-center shadow-sm hover:bg-red-50 text-red-500 transition-colors"
            style={{ borderRadius: "45% 55% 45% 55% / 55% 45% 55% 45%" }}
          >
            <X className="w-8 h-8" strokeWidth={3} />
          </button>
          
          <button
            onClick={() => handleManualSwipe("keep")}
            className="w-16 h-16 bg-white border-2 border-gray-800 rounded-full flex items-center justify-center shadow-sm hover:bg-green-50 text-green-500 transition-colors"
            style={{ borderRadius: "55% 45% 55% 45% / 45% 55% 45% 55%" }}
          >
            <Heart className="w-8 h-8" strokeWidth={3} />
          </button>
        </div>
      )}
    </motion.div>
  );
}
