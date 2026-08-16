"use client";

import { useState } from "react";
import WordCard from "./WordCard";
import { WordDTO, DictionaryAPI } from "@/lib/dictionary-api";

interface RandomWordCardProps {
  initialWord: WordDTO;
}

export default function RandomWordCard({ initialWord }: RandomWordCardProps) {
  const [word, setWord] = useState<WordDTO>(initialWord);
  const [loading, setLoading] = useState(false);
  const [cooldown, setCooldown] = useState(false);

  const handleReload = async () => {
    if (cooldown) return;
    
    setLoading(true);
    setCooldown(true);
    try {
      console.log("Fetching new random word...");
      const newWord = await DictionaryAPI.getRandom();
      console.log("New word received:", newWord?.word);
      if (newWord) {
        setWord(newWord);
      }
    } catch (error) {
      console.error("Failed to fetch random word", error);
    } finally {
      setLoading(false);
      // Wait 500ms before allowing another click
      setTimeout(() => setCooldown(false), 500);
    }
  };

  return (
    <WordCard 
      word={word} 
      title={"Random Word"} 
      onReload={handleReload} 
      isLoading={loading} 
    />
  );
}
