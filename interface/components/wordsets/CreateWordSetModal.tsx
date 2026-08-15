import { useEffect, useState } from "react";
import { CreateWordSetRequest, WordSetVisibility, WordSetDifficulty, WordSetAPI, Language } from "@/lib/wordset-api";

interface Props {
  onClose: () => void;
  onSubmit: (data: CreateWordSetRequest) => void;
  isSubmitting: boolean;
}

export function CreateWordSetModal({ onClose, onSubmit, isSubmitting }: Props) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [visibility, setVisibility] = useState<WordSetVisibility>("private");
  const [difficulty, setDifficulty] = useState<WordSetDifficulty>("Beginner");
  const [selectedTopics, setSelectedTopics] = useState<string[]>([]);
  const [availableTopics, setAvailableTopics] = useState<string[]>([]);
  const [annotationLanguage, setAnnotationLanguage] = useState<string>("vi");
  const [availableLanguages, setAvailableLanguages] = useState<Language[]>([]);

  useEffect(() => {
    WordSetAPI.getTopics().then(setAvailableTopics).catch(console.error);
    WordSetAPI.getSupportedLanguages().then(setAvailableLanguages).catch(console.error);
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    onSubmit({ name, description, visibility, difficulty, topics: selectedTopics, annotation_language: annotationLanguage, source_language: "en" });
  };

  const toggleTopic = (topic: string) => {
    setSelectedTopics(prev => 
      prev.includes(topic) ? prev.filter(t => t !== topic) : [...prev, topic]
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div className="bg-[#1e1e2d] border border-white/10 rounded-2xl p-6 max-w-md w-full shadow-2xl">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-xl font-bold text-white">Create Word Set</h3>
          <button 
            onClick={onClose}
            className="text-white/50 hover:text-white transition-colors"
          >✕</button>
        </div>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-white/70 mb-1">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:border-indigo-500 transition-colors"
              placeholder="e.g. IELTS Environment"
              required
              maxLength={100}
            />
          </div>
          <div>
            <label className="block text-sm text-white/70 mb-1">Description <span className="text-white/30">(optional)</span></label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:border-indigo-500 transition-colors min-h-[100px]"
              placeholder="What is this set about?"
              maxLength={500}
            />
          </div>
                    <div>
            <label className="block text-sm text-white/70 mb-1">Translation Language</label>
            <select
              value={annotationLanguage}
              onChange={(e) => setAnnotationLanguage(e.target.value)}
              className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:border-indigo-500 transition-colors appearance-none"
            >
              {availableLanguages.map(l => (
                <option key={l.code} value={l.code}>{l.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-sm text-white/70 mb-1">Visibility</label>
            <select
              value={visibility}
              onChange={(e) => setVisibility(e.target.value as WordSetVisibility)}
              className="w-full bg-black/20 border border-white/10 rounded-xl px-4 py-3 text-white focus:outline-none focus:border-indigo-500 transition-colors appearance-none"
            >
              <option value="private">Private (Only you)</option>
              <option value="public">Public (Everyone)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-white/70 mb-2">Difficulty</label>
            <div className="flex gap-4">
              {["Beginner", "Intermediate", "Advanced"].map((level) => (
                <label key={level} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="difficulty"
                    value={level}
                    checked={difficulty === level}
                    onChange={(e) => setDifficulty(e.target.value as WordSetDifficulty)}
                    className="w-4 h-4 text-indigo-500 bg-black/20 border-white/10 focus:ring-indigo-500"
                  />
                  <span className="text-white text-sm">{level}</span>
                </label>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-sm text-white/70 mb-2">Topics</label>
            <div className="flex flex-wrap gap-2">
              {availableTopics.map(topic => (
                <button
                  key={topic}
                  type="button"
                  onClick={() => toggleTopic(topic)}
                  className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                    selectedTopics.includes(topic) 
                      ? "bg-indigo-500/20 border-indigo-500 text-indigo-300" 
                      : "bg-black/20 border-white/10 text-white/50 hover:text-white/80"
                  }`}
                >
                  {topic}
                </button>
              ))}
            </div>
          </div>
          <div className="pt-4 flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-3 rounded-xl bg-white/5 hover:bg-white/10 text-white font-bold transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || !name.trim()}
              className="flex-1 py-3 rounded-xl bg-indigo-500 hover:bg-indigo-600 disabled:opacity-50 text-white font-bold transition-colors"
            >
              {isSubmitting ? "Creating..." : "Create"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
