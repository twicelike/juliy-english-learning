import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { WordSetAPI, WordSetListItem } from "@/lib/wordset-api";
import { WordDTO } from "@/lib/dictionary-api";

interface Props {
  word: WordDTO;
  onClose: () => void;
}

export function AddToSetModal({ word, onClose }: Props) {
  const [sets, setSets] = useState<WordSetListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [addingTo, setAddingTo] = useState<string | null>(null);
  
  const uniqueDefs = word.definitions?.filter((def, i, arr) => arr.findIndex(d => d.sense_id === def.sense_id && (d.word_id || word.id) === (def.word_id || word.id)) === i) || [];
  const posList = Array.from(new Set(uniqueDefs.map(d => d.partOfSpeech || "Other")));
  
  const [activeTab, setActiveTab] = useState<string>(posList[0] || "Other");
  const [selectedDefs, setSelectedDefs] = useState<Set<string>>(new Set(
    uniqueDefs.length > 0 ? [`${uniqueDefs[0].word_id || word.id}:${uniqueDefs[0].sense_id}`] : []
  ));
  
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    loadSets();
  }, [word.id]);

  const loadSets = async () => {
    try {
      setLoading(true);
      const allWordIds = Array.from(new Set(uniqueDefs.map(d => d.word_id || word.id)));
      const res = await WordSetAPI.getWordSets(allWordIds);
      const items = res.items || [];
      setSets(items);

      // Pre-populate selectedDefs from already-saved senses
      const alreadySaved = new Set<string>();
      items.forEach(set => {
        (set.saved_senses || []).forEach(s => {
          alreadySaved.add(`${s.word_id}:${s.sense_id}`);
        });
      });
      // If we found any saved senses across the sets, pre-select them.
      if (alreadySaved.size > 0) {
        setSelectedDefs(alreadySaved);
      } else {
        setSelectedDefs(new Set());
      }
    } catch (error) {
      console.error("Failed to load sets:", error);
    } finally {
      setLoading(false);
    }
  };

  const toggleDef = (key: string) => {
    const next = new Set(selectedDefs);
    if (next.has(key)) {
      next.delete(key);
    } else {
      next.add(key);
    }
    setSelectedDefs(next);
  };

  const handleToggle = async (e: React.ChangeEvent<HTMLInputElement>, set: WordSetListItem) => {
    e.preventDefault();
    if (addingTo) return;
    
    try {
      setAddingTo(set.id);
      if (set.contains_word) {
        // Remove them all to start fresh
        const currentSavedSenses = set.saved_senses || [];
        const wordIdsToRemove = Array.from(new Set(currentSavedSenses.map(s => s.word_id)));
        await Promise.all(wordIdsToRemove.map(wId => WordSetAPI.removeWordFromSet(set.id, wId).catch(() => {})));
        
        if (selectedDefs.size === 0) {
          // They unchecked the whole set, just remove and we're done
          await loadSets();
          return;
        }
        
        // Otherwise, they just wanted to update the meanings. Add the selected ones.
        const items = Array.from(selectedDefs).map(key => {
          const [wId, sIdx] = key.split(":");
          return { word_id: wId, sense_id: sIdx };
        });
        
        await WordSetAPI.bulkAddWords(set.id, items);
        await loadSets();
      } else {
        if (selectedDefs.size === 0) {
          alert("Please select at least one meaning to save.");
          return;
        }
        if (set.word_count + selectedDefs.size > 100) {
          alert(`This word set can contain up to 100 words. You are trying to add ${selectedDefs.size} meaning(s), but only ${100 - set.word_count} slot(s) are available.`);
          return;
        }
        
        const items = Array.from(selectedDefs).map(key => {
          const [wId, sIdx] = key.split(":");
          return { word_id: wId, sense_id: sIdx };
        });
        
        await WordSetAPI.bulkAddWords(set.id, items);
        await loadSets();
      }
    } catch (error: any) {
      alert(error.response?.data?.error || "Failed to update word set");
    } finally {
      setAddingTo(null);
    }
  };

  if (!mounted) return null;

  return createPortal(
    <div className="fixed inset-0 z-[9999] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div className="bg-[#1e1e2d] border border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl flex flex-col max-h-[90vh]">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-bold text-white">Add &quot;{word.word}&quot; to Set</h3>
          <button 
            onClick={onClose}
            className="text-white/50 hover:text-white transition-colors"
          >✕</button>
        </div>

        {uniqueDefs.length > 0 && (
          <div className="mb-6 flex flex-col">
            <label className="block text-sm font-medium text-white/70 mb-3">Select Meanings to Save:</label>
            
            {/* POS Tabs */}
            {posList.length > 1 && (
              <div className="flex gap-2 mb-3 overflow-x-auto custom-scrollbar pb-1">
                {posList.map(pos => (
                  <button
                    key={pos}
                    onClick={() => setActiveTab(pos)}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                      activeTab === pos 
                        ? "bg-indigo-500 text-white" 
                        : "bg-white/5 text-white/50 hover:bg-white/10"
                    }`}
                  >
                    {pos}
                  </button>
                ))}
              </div>
            )}
            
            {/* Definitions List */}
            <div className="space-y-2 max-h-48 overflow-y-auto custom-scrollbar pr-2">
              {uniqueDefs
                .filter(def => (def.partOfSpeech || "Other") === activeTab)
                .map(def => {
                  const key = `${def.word_id || word.id}:${def.sense_id}`;
                  const isSelected = selectedDefs.has(key);
                  return (
                    <div 
                      key={key}
                      onClick={() => toggleDef(key)}
                      className={`flex items-start gap-3 p-3 rounded-xl cursor-pointer transition-colors border ${
                        isSelected 
                          ? "bg-indigo-500/10 border-indigo-500/50" 
                          : "bg-black/20 border-white/5 hover:border-white/20"
                      }`}
                    >
                      <div className={`mt-0.5 w-5 h-5 rounded border flex-shrink-0 flex items-center justify-center transition-colors ${
                        isSelected ? "bg-indigo-500 border-indigo-500" : "border-white/30"
                      }`}>
                        {isSelected && <svg className="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                      </div>
                      <div className="text-sm text-white/90 line-clamp-2">
                        {def.definition}
                      </div>
                    </div>
                  );
                })}
            </div>
            
            <div className="text-xs text-white/40 mt-3 text-right">
              {selectedDefs.size} meaning{selectedDefs.size !== 1 ? 's' : ''} selected
            </div>
          </div>
        )}
        
        <div className="h-px bg-white/10 w-full my-4"></div>
        
        <div className="flex-1 overflow-y-auto space-y-2 pr-2 custom-scrollbar">
          {loading ? (
            <div className="animate-pulse space-y-3">
              {[1, 2, 3].map(i => (
                <div key={i} className="h-14 bg-white/5 rounded-xl"></div>
              ))}
            </div>
          ) : sets.length === 0 ? (
            <div className="text-center py-8 text-white/50">
              No word sets found. Create one in My Word Sets.
            </div>
          ) : (
            sets.map(set => (
              <label 
                key={set.id} 
                className={`flex items-center justify-between p-4 rounded-xl border transition-colors ${
                  set.contains_word 
                    ? "bg-indigo-500/10 border-indigo-500/50" 
                    : "bg-black/20 border-white/5 hover:border-white/20"
                } ${addingTo === set.id ? "opacity-50 pointer-events-none" : ""}`}
              >
                <div className="flex items-center gap-4 flex-1 cursor-pointer" onClick={(e) => handleToggle(e as any, set)}>
                  <div className={`w-5 h-5 rounded border flex items-center justify-center transition-colors ${
                    set.contains_word ? "bg-indigo-500 border-indigo-500" : "border-white/30"
                  }`}>
                    {set.contains_word && <span className="text-white text-xs">✓</span>}
                  </div>
                  <input
                    type="checkbox"
                    className="hidden"
                    checked={!!set.contains_word}
                    onChange={(e) => handleToggle(e, set)}
                  />
                  <div>
                    <div className="text-white font-medium">{set.name}</div>
                    <div className="text-white/40 text-xs mt-0.5">{set.word_count} / 100 words</div>
                  </div>
                  <div onClick={(e) => e.stopPropagation()}>
                    <button
                      disabled={addingTo === set.id}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleToggle(e as any, set);
                      }}
                      className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ml-auto ${
                        set.contains_word 
                          ? "bg-white/10 hover:bg-white/20 text-white" 
                          : "bg-indigo-500 hover:bg-indigo-600 text-white"
                      }`}
                    >
                      {set.contains_word ? 'Update' : 'Save'}
                    </button>
                  </div>
                </div>
                {addingTo === set.id && (
                  <div className="w-4 h-4 border-2 border-white/20 border-t-indigo-500 rounded-full animate-spin"></div>
                )}
              </label>
            ))
          )}
        </div>
      </div>
    </div>,
    document.body
  );
}
