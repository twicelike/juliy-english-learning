import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { WordSetAPI, WordSetListItem } from "@/lib/wordset-api";
import { CreateWordSetModal } from "./CreateWordSetModal";
import { X, Plus, Folder, Check } from "lucide-react";

interface Props {
  wordId: string;
  senseId: string;
  onClose: () => void;
}

export function MiniAddToSetModal({ wordId, senseId, onClose }: Props) {
  const [sets, setSets] = useState<WordSetListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [addingTo, setAddingTo] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    loadSets();
  }, [wordId, senseId]);

  const loadSets = async () => {
    try {
      setLoading(true);
      const res = await WordSetAPI.getWordSets([wordId]);
      setSets(res.items || []);
    } catch (err) {
      console.error("Failed to load word sets:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleAddToSet = async (set: WordSetListItem) => {
    try {
      setAddingTo(set.id);
      
      const isAlreadyAdded = set.saved_senses?.some(s => s.word_id === wordId && s.sense_id === senseId);
      
      if (isAlreadyAdded) {
        await loadSets();
        setAddingTo(null);
        return;
      }
      
      if (set.word_count >= 50) {
        alert("This word set is full (maximum 50 words).");
        setAddingTo(null);
        return;
      }

      await WordSetAPI.bulkAddWords(set.id, [{ word_id: wordId, sense_id: senseId }]);
      await loadSets();
      
      // Close modal after brief success indication
      setTimeout(() => {
        onClose();
      }, 500);
      
    } catch (err) {
      console.error("Failed to add to set:", err);
      alert("Failed to add word to set.");
    } finally {
      setAddingTo(null);
    }
  };

  const handleCreateSet = async (data: import("@/lib/wordset-api").CreateWordSetRequest) => {
    try {
      setCreating(true);
      const newSet = await WordSetAPI.createWordSet(data);
      await WordSetAPI.bulkAddWords(newSet.id, [{ word_id: wordId, sense_id: senseId }]);
      setShowCreate(false);
      onClose(); // Close everything upon success
    } catch (err) {
      console.error("Failed to create set:", err);
      alert("Failed to create set.");
    } finally {
      setCreating(false);
    }
  };

  if (!mounted) return null;

  const modalContent = (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4" onClick={(e) => { e.stopPropagation(); onClose(); }}>
      <div 
        className="bg-white rounded-3xl w-full max-w-sm overflow-hidden flex flex-col shadow-xl text-black border-2 border-gray-800"
        style={{ borderRadius: "20px 255px 20px 20px/255px 20px 255px 20px" }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-4 border-b border-black/10">
          <h2 className="font-bold text-lg">Save to WordSet</h2>
          <button onClick={(e) => { e.stopPropagation(); onClose(); }} className="p-1 hover:bg-black/5 rounded-full transition-colors">
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>
        
        <div className="p-2 max-h-[60vh] overflow-y-auto">
          {loading ? (
            <div className="py-8 text-center text-gray-500">Loading your sets...</div>
          ) : sets.length === 0 ? (
            <div className="py-8 text-center text-gray-500">
              <p>You don't have any word sets yet.</p>
            </div>
          ) : (
            <div className="space-y-1">
              {sets.map(set => {
                const isAdded = set.saved_senses?.some(s => s.word_id === wordId && s.sense_id === senseId);
                const isAdding = addingTo === set.id;
                
                return (
                  <button
                    key={set.id}
                    onClick={(e) => { e.stopPropagation(); handleAddToSet(set); }}
                    disabled={isAdding || (isAdded ? true : false)}
                    className="w-full flex items-center gap-3 p-3 text-left rounded-xl hover:bg-black/5 transition-colors disabled:opacity-50"
                  >
                    <div className="bg-blue-100 p-2 rounded-lg text-blue-600">
                      <Folder className="w-5 h-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-gray-900 truncate">{set.name}</div>
                      <div className="text-xs text-gray-500">{set.word_count}/50 words</div>
                    </div>
                    
                    <div>
                      {isAdding ? (
                        <div className="w-5 h-5 rounded-full border-2 border-gray-300 border-t-gray-600 animate-spin" />
                      ) : isAdded ? (
                        <div className="w-6 h-6 rounded-full bg-green-100 text-green-600 flex items-center justify-center">
                          <Check className="w-4 h-4" />
                        </div>
                      ) : null}
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>
        
        <div className="p-3 border-t border-black/10 bg-gray-50">
          <button 
            onClick={(e) => { e.stopPropagation(); setShowCreate(true); }}
            className="w-full py-3 flex items-center justify-center gap-2 border-2 border-gray-800 text-gray-800 font-medium rounded-xl hover:bg-gray-200 transition-colors bg-white"
            style={{ borderRadius: "15px 225px 15px 255px/255px 15px 225px 15px" }}
          >
            <Plus className="w-5 h-5" />
            Create New WordSet
          </button>
        </div>
      </div>
      
      {showCreate && (
        <div onClick={(e) => e.stopPropagation()}>
          <CreateWordSetModal 
            onClose={() => setShowCreate(false)}
            onSubmit={handleCreateSet}
            isSubmitting={creating}
          />
        </div>
      )}
    </div>
  );

  return createPortal(modalContent, document.body);
}
