import { useEffect, useState } from "react";
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

  useEffect(() => {
    loadSets();
  }, [word.id]);

  const loadSets = async () => {
    try {
      setLoading(true);
      const res = await WordSetAPI.getWordSets(word.id);
      setSets(res.items || []);
    } catch (error) {
      console.error("Failed to load sets:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (e: React.ChangeEvent<HTMLInputElement>, set: WordSetListItem) => {
    e.preventDefault();
    if (addingTo) return;
    
    try {
      setAddingTo(set.id);
      if (set.contains_word) {
        await WordSetAPI.removeWordFromSet(set.id, word.id);
        setSets(sets.map(s => s.id === set.id ? { ...s, contains_word: false, word_count: s.word_count - 1 } : s));
      } else {
        if (set.word_count >= 50) {
          alert("This word set can contain up to 50 words.");
          return;
        }
        await WordSetAPI.bulkAddWords(set.id, [word.id]);
        setSets(sets.map(s => s.id === set.id ? { ...s, contains_word: true, word_count: s.word_count + 1 } : s));
      }
    } catch (error: any) {
      alert(error.response?.data?.error || "Failed to update word set");
    } finally {
      setAddingTo(null);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div className="bg-[#1e1e2d] border border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl flex flex-col max-h-[80vh]">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-bold text-white">Add &quot;{word.word}&quot; to Set</h3>
          <button 
            onClick={onClose}
            className="text-white/50 hover:text-white transition-colors"
          >✕</button>
        </div>
        
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
                className={`flex items-center justify-between p-4 rounded-xl border transition-colors cursor-pointer ${
                  set.contains_word 
                    ? "bg-indigo-500/10 border-indigo-500/50" 
                    : "bg-black/20 border-white/5 hover:border-white/20"
                } ${addingTo === set.id ? "opacity-50 pointer-events-none" : ""}`}
              >
                <div className="flex items-center gap-4">
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
                    <div className="text-white/40 text-xs mt-0.5">{set.word_count} / 50 words</div>
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
    </div>
  );
}
