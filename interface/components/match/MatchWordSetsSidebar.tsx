"use client";

import React, { useEffect, useState } from "react";
import { WordSetAPI, WordSetListItem } from "@/lib/wordset-api";
import { Folder, Check, ChevronLeft, ChevronRight, X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

interface Props {
  selectedSetIds: string[];
  onToggleSet: (id: string) => void;
  refreshTrigger: number;
}

export default function MatchWordSetsSidebar({ selectedSetIds, onToggleSet, refreshTrigger }: Props) {
  const [sets, setSets] = useState<WordSetListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [isOpenMobile, setIsOpenMobile] = useState(false);

  useEffect(() => {
    const loadSets = async () => {
      try {
        setLoading(true);
        // Using an empty array fetches all user's sets
        const res = await WordSetAPI.getWordSets([]);
        setSets(res.items || []);
      } catch (err) {
        console.error("Failed to load word sets:", err);
      } finally {
        setLoading(false);
      }
    };
    loadSets();
  }, [refreshTrigger]);

  const SidebarInner = () => (
    <>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-handwritten text-black font-bold">Save to Word Sets</h2>
        {/* Mobile close button */}
        <button
          className="md:hidden p-1 text-gray-500 hover:text-black rounded-full hover:bg-black/5"
          onClick={() => setIsOpenMobile(false)}
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 pr-2 [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">
        {loading ? (
          <div className="text-gray-500 font-handwritten text-center mt-8">Loading sets...</div>
        ) : sets.length === 0 ? (
          <div className="text-gray-500 font-handwritten text-center mt-8">
            You don't have any Word Sets yet. Create one in your profile!
          </div>
        ) : (
          sets.map(set => {
            const isSelected = selectedSetIds.includes(set.id);
            const isFull = set.word_count >= 100;

            return (
              <button
                key={set.id}
                onClick={() => {
                  if (!isFull) {
                    onToggleSet(set.id);
                  }
                }}
                disabled={isFull}
                className={`w-full flex items-center gap-3 p-3 text-left rounded-xl transition-colors border-2 ${isSelected
                    ? "bg-green-50 border-green-500"
                    : "bg-white hover:bg-gray-50 border-transparent hover:border-gray-200"
                  } ${isFull ? "opacity-50 cursor-not-allowed" : ""}`}
                style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}
              >
                <div className={`flex items-center justify-center w-6 h-6 rounded-md border-2 shrink-0 transition-colors ${isSelected ? "bg-green-500 border-green-500 text-white" : "border-gray-300 bg-white"
                  }`}>
                  {isSelected && <Check className="w-4 h-4" strokeWidth={3} />}
                </div>

                <div className="flex-1 min-w-0">
                  <div className={`font-medium truncate font-handwritten text-lg ${isFull ? "text-red-500" : "text-gray-900"}`}>
                    {set.name}
                  </div>
                  <div className={`text-xs ${isFull ? "text-red-400 font-bold" : "text-gray-500"}`}>
                    {set.word_count}/100 words {isFull && "(Full)"}
                  </div>
                </div>
              </button>
            );
          })
        )}
      </div>

      <div className="mt-4 pt-4 border-t-2 border-gray-100 text-sm text-gray-500 font-handwritten text-center">
        Words will be saved to ticked sets when you swipe KEEP.
      </div>
    </>
  );

  return (
    <>
      {/* Desktop Sidebar */}
      <div className="hidden md:flex flex-col fixed right-4 top-1/2 -translate-y-1/2 h-[80vh] w-80 z-40">
        <div className="flex flex-col h-full bg-white rounded-3xl shadow-sm border-2 border-gray-800 p-4"
          style={{ borderRadius: "15px 225px 15px 255px/255px 15px 225px 15px" }}>
          <SidebarInner />
        </div>
      </div>

      {/* Mobile Toggle Button */}
      <button
        onClick={() => setIsOpenMobile(true)}
        className="md:hidden fixed bottom-6 right-6 z-40 bg-gray-900 text-white p-4 rounded-full shadow-lg border-2 border-black flex items-center justify-center"
        style={{ borderRadius: "255px 15px 225px 15px/15px 225px 15px 255px" }}
      >
        <Folder className="w-6 h-6" />
        {selectedSetIds.length > 0 && (
          <span className="absolute -top-2 -right-2 bg-green-500 text-white text-xs font-bold w-6 h-6 rounded-full flex items-center justify-center border-2 border-white">
            {selectedSetIds.length}
          </span>
        )}
      </button>

      {/* Mobile Drawer */}
      <AnimatePresence>
        {isOpenMobile && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setIsOpenMobile(false)}
              className="md:hidden fixed inset-0 bg-black/40 z-50 backdrop-blur-sm"
            />
            <motion.div
              initial={{ x: "100%" }}
              animate={{ x: 0 }}
              exit={{ x: "100%" }}
              transition={{ type: "spring", bounce: 0, duration: 0.4 }}
              className="md:hidden fixed top-0 right-0 bottom-0 w-80 max-w-[85vw] z-50 p-6 bg-white border-l-2 border-gray-800 shadow-xl flex flex-col"
            >
              <SidebarInner />
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}
