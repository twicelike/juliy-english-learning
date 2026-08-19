"use client";

import { useEffect, useState, useCallback, Suspense, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { LearningAPI, SessionDTO, SessionItemDTO, WordResult, ProgressQuality } from "@/lib/learning-api";
import { useAuthStore } from "@/features/auth";
import { Flashcard } from "@/components/learn/Flashcard";
import { RatingBar } from "@/components/learn/RatingBar";
import { SessionProgress } from "@/components/learn/SessionProgress";

type Phase = "initializing" | "preparing" | "flashcard" | "mcq" | "fill_blank" | "submitting" | "done" | "error";

function SessionRunner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const mode = searchParams.get("mode") as "new_words" | "review" | null;
  const sessionId = searchParams.get("session_id");
  const { user } = useAuthStore();

  const [session, setSession] = useState<SessionDTO | null>(null);
  const [phase, setPhase] = useState<Phase>("initializing");
  const [errorMsg, setErrorMsg] = useState("");
  const [doneMsg, setDoneMsg] = useState({ title: "Session Complete!", desc: "Your progress has been saved and your learning algorithm is updated." });

  const [results, setResults] = useState<Record<string, WordResult>>({});
  const [currentIndex, setCurrentIndex] = useState(0);

  // Flashcard state
  const [flipped, setFlipped] = useState(false);

  // MCQ state
  const [selectedChoice, setSelectedChoice] = useState<number | null>(null);

  // Fill Blank state
  const [fbInput, setFbInput] = useState("");
  const [fbSubmitted, setFbSubmitted] = useState(false);

  const initRef = useRef(false);

  useEffect(() => {
    if (!user) {
      router.push("/login");
      return;
    }

    if (initRef.current) return;
    initRef.current = true;

    const loadSession = async () => {
      try {
        let s: SessionDTO;
        if (sessionId) {
          s = await LearningAPI.getSessionById(sessionId);
        } else {
          if (mode !== "new_words" && mode !== "review") {
            setErrorMsg("Invalid session mode");
            setPhase("error");
            return;
          }
          s = await LearningAPI.createSession(mode);
        }
        if (s.items.length === 0) {
          setPhase("done");
          return;
        }

        // Initial shuffle for flashcards
        s.items = [...s.items].sort(() => Math.random() - 0.5);

        setSession(s);

        const initRes: Record<string, WordResult> = {};
        s.items.forEach((item) => {
          const key = `${item.word_id}:${item.sense_index}`;
          initRes[key] = { word_id: item.word_id, sense_index: item.sense_index, flashcard_quality: 1 };
        });
        setResults(initRes);

        if (s.status === "preparing") {
          setPhase("preparing");
        } else {
          setPhase("flashcard");
        }
      } catch (e: any) {
        const errorCode = e.response?.data?.error?.code;
        if (errorCode === "NO_NEW_WORDS") {
          setDoneMsg({ title: "You're all caught up!", desc: "You have no new bookmarked words to learn right now." });
          setPhase("done");
        } else if (errorCode === "NO_DUE_WORDS") {
          setDoneMsg({ title: "You're all caught up!", desc: "There are no words due for review right now." });
          setPhase("done");
        } else if (errorCode === "NO_REVIEWABLE_WORDS") {
          setDoneMsg({ title: "Word Set Completed!", desc: "All words in this set are currently mastered or you have selected a strategy that has no applicable words." });
          setPhase("done");
        } else {
          console.error("Failed to load session:", e);
          setErrorMsg(e.message || "Failed to start session");
          setPhase("error");
        }
      }
    };

    loadSession();
  }, [user, mode, router]);

  // Polling if preparing
  useEffect(() => {
    if (phase !== "preparing" || !session) return;

    const interval = setInterval(async () => {
      try {
        const s = await LearningAPI.getSessionById(session.id);
        if (s.status === "active") {
          setSession(s);
          setPhase("flashcard");
          clearInterval(interval);
        }
      } catch (e) {
        console.error("Polling error", e);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [phase, session]);

  // Handle Submit Session
  useEffect(() => {
    if (phase === "submitting" && session) {
      const submit = async () => {
        try {
          const resultsArr = Object.values(results);
          await LearningAPI.submitSessionResults(session.id, resultsArr);
          setPhase("done");
        } catch (e: any) {
          setErrorMsg(e.message || "Failed to submit session");
          setPhase("error");
        }
      };
      submit();
    }
  }, [phase, session, results]);

  if (phase === "initializing") {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (phase === "preparing") {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-6 px-6 text-center">
        <div className="w-12 h-12 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        <h2 className="text-2xl font-bold text-white">Preparing your exercises...</h2>
        <p className="text-white/50 max-w-sm">
          Generating multiple choice and fill-in-the-blank questions. This will just take a moment.
        </p>
      </div>
    );
  }

  if (phase === "error") {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-4 text-center px-6">
        <h2 className="text-2xl font-bold text-white">Something went wrong</h2>
        <p className="text-white/50">{errorMsg}</p>
        <button onClick={() => router.push("/learn")} className="mt-4 px-6 py-3 rounded-xl bg-indigo-500 text-white hover:bg-indigo-600">
          Back to Dashboard
        </button>
      </div>
    );
  }

  if (phase === "done") {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-6 text-center px-6">
        <h1 className="text-3xl font-bold text-white">{doneMsg.title}</h1>
        <p className="text-white/50">{doneMsg.desc}</p>
        <button onClick={() => router.push("/learn")} className="px-6 py-3 rounded-xl bg-indigo-500 text-white font-medium hover:bg-indigo-600">
          Back to Dashboard
        </button>
      </div>
    );
  }

  if (phase === "submitting") {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-center gap-4">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        <p className="text-white/50">Saving your progress...</p>
      </div>
    );
  }

  const items = session?.items || [];
  const card = items[currentIndex];
  if (!card) return null;

  const handleFlashcardResult = (quality: ProgressQuality) => {
    const key = `${card.word_id}:${card.sense_index}`;
    setResults(prev => ({
      ...prev,
      [key]: { ...prev[key], flashcard_quality: quality }
    }));
    
    setFlipped(false);
    if (currentIndex + 1 >= items.length) {
      setCurrentIndex(0);
      setPhase("mcq");
      setSession(prev => prev ? { ...prev, items: [...prev.items].sort(() => Math.random() - 0.5) } : prev);
    } else {
      setCurrentIndex(i => i + 1);
    }
  };

  const handleMcqSelect = (choiceIndex: number) => {
    if (selectedChoice !== null) return;
    setSelectedChoice(choiceIndex);
    
    const choice = card.choices[choiceIndex];
    const isCorrect = choice.is_correct;
    
    const key = `${card.word_id}:${card.sense_index}`;
    setResults(prev => ({
      ...prev,
      [key]: { ...prev[key], multiple_choice_correct: isCorrect }
    }));
  };

  const handleMcqNext = () => {
    setSelectedChoice(null);
    if (currentIndex + 1 >= items.length) {
      setCurrentIndex(0);
      setPhase("fill_blank");
      setSession(prev => prev ? { ...prev, items: [...prev.items].sort(() => Math.random() - 0.5) } : prev);
    } else {
      setCurrentIndex(i => i + 1);
    }
  };

  const handleFbSubmit = () => {
    if (fbSubmitted) return;
    setFbSubmitted(true);
    
    const correctAnswer = card.fill_blank_answer || "";
    const isCorrect = fbInput.trim().toLowerCase() === correctAnswer.toLowerCase();
    
    const key = `${card.word_id}:${card.sense_index}`;
    setResults(prev => ({
      ...prev,
      [key]: { ...prev[key], fill_blank_correct: isCorrect }
    }));
  };

  const handleFbNext = () => {
    setFbSubmitted(false);
    setFbInput("");
    if (currentIndex + 1 >= items.length) {
      setPhase("submitting");
    } else {
      setCurrentIndex(i => i + 1);
    }
  };

  const handleSuspendWord = async () => {
    if (!session) return;
    try {
      // 1. Call API to suspend
      await LearningAPI.suspendWord(card.word_id);
      
      // 2. Remove the word from the session locally
      const newItems = session.items.filter(item => item.word_id !== card.word_id);
      if (newItems.length === 0) {
        setPhase("done"); // If no words left, we're done
        return;
      }
      
      // 3. Adjust current index & phase if needed
      let nextIndex = currentIndex;
      if (nextIndex >= newItems.length) {
        nextIndex = 0;
        if (phase === "flashcard") setPhase("mcq");
        else if (phase === "mcq") setPhase("fill_blank");
        else if (phase === "fill_blank") setPhase("submitting");
      }
      
      // 4. Update state without triggering full unmount
      setSession({ ...session, items: newItems });
      setCurrentIndex(nextIndex);
      
    } catch (error) {
      console.error(error);
      alert("Failed to suspend word.");
    }
  };

  const progressPct = Math.round((currentIndex / items.length) * 100);

  return (
    <div className="min-h-[calc(100vh-4rem)] flex flex-col items-center justify-start px-6 py-12">
      <SessionProgress 
        currentIndex={currentIndex} 
        totalItems={items.length} 
        onSuspend={() => {
          if (confirm("Are you sure you want to suspend this word? It will never appear in your learning sessions again.")) {
            handleSuspendWord();
          }
        }}
        phase={phase}
      />

      {/* PHASE 1: FLASHCARD */}
      {phase === "flashcard" && (
        <div className="w-full max-w-md flex flex-col items-center animate-in fade-in zoom-in-95 duration-500">
          <Flashcard 
            card={card} 
            flipped={flipped} 
            onFlip={() => setFlipped(prev => !prev)} 
          />
          {flipped && <RatingBar onRating={handleFlashcardResult} />}
        </div>
      )}

      {/* PHASE 2: MCQ */}
      {phase === "mcq" && (
        <div className="w-full max-w-lg flex flex-col items-center">
          <div className="w-full rounded-3xl bg-white/5 border border-white/10 p-8 mb-6 text-center">
            <p className="text-sm text-white/40 mb-3">What does this word mean?</p>
            <div className="flex items-center justify-center gap-3 flex-wrap">
              <span className="text-4xl font-bold text-white">{card.word}</span>
              {card.pos && (
                <span className="px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-xs">{card.pos}</span>
              )}
            </div>
          </div>

          <div className="w-full space-y-3">
            {(!card.choices || card.choices.length === 0) ? (
              <div className="text-center p-6 bg-white/5 border border-white/10 rounded-2xl">
                <p className="text-white/50 mb-4">No choices available for this word.</p>
                <button onClick={handleMcqNext} className="px-6 py-2 rounded-xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-colors">
                  Skip →
                </button>
              </div>
            ) : (
              card.choices.map((choice, i) => {
                let cls = "w-full p-4 rounded-2xl border text-left text-sm leading-relaxed transition-all ";
                if (selectedChoice === null) {
                  cls += "bg-white/5 border-white/10 text-white/80 hover:bg-white/10 cursor-pointer";
                } else if (choice.is_correct) {
                  cls += "bg-emerald-500/20 border-emerald-500/40 text-emerald-300";
                } else if (selectedChoice === i) {
                  cls += "bg-red-500/20 border-red-500/40 text-red-300";
                } else {
                  cls += "bg-white/5 border-white/10 text-white/30";
                }
                return (
                  <button key={i} className={cls} onClick={() => handleMcqSelect(i)}>
                    <span className="font-mono text-white/40 mr-3">{String.fromCharCode(65 + i)}.</span>
                    {choice.definition}
                  </button>
                );
              })
            )}
          </div>

          {selectedChoice !== null && (
            <button onClick={handleMcqNext} className="mt-8 px-8 py-3 rounded-2xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-colors">
              Next →
            </button>
          )}
        </div>
      )}

      {/* PHASE 3: FILL IN THE BLANK */}
      {phase === "fill_blank" && (
        <div className="w-full max-w-lg flex flex-col items-center">
          <div className="w-full rounded-3xl bg-white/5 border border-white/10 p-8 mb-6">
            <p className="text-sm text-white/40 mb-6 text-center">Fill in the blank with the correct word.</p>
            
            {!card.fill_blank_question ? (
              <div className="text-center p-6">
                <p className="text-white/50 mb-4">No sentence available for this word.</p>
                <button onClick={handleFbNext} className="px-6 py-2 rounded-xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-colors">
                  Skip →
                </button>
              </div>
            ) : (
              <div className="text-lg text-white/90 leading-relaxed text-center space-x-2">
                {card.fill_blank_question.split("___").map((part, i, arr) => (
                  <span key={i}>
                    {part}
                    {i < arr.length - 1 && (
                      <input
                        type="text"
                        value={fbInput}
                        onChange={(e) => setFbInput(e.target.value)}
                        disabled={fbSubmitted}
                        className={`inline-block w-32 px-3 py-1 mx-2 text-center bg-transparent border-b-2 outline-none transition-colors ${
                          !fbSubmitted ? "border-indigo-500 text-indigo-300" :
                          fbInput.trim().toLowerCase() === (card.fill_blank_answer || "").toLowerCase() ? "border-emerald-500 text-emerald-400" : "border-red-500 text-red-400"
                        }`}
                        placeholder="..."
                        autoFocus
                        onKeyDown={(e) => {
                          if (e.key === "Enter" && fbInput.trim()) {
                            fbSubmitted ? handleFbNext() : handleFbSubmit();
                          }
                        }}
                      />
                    )}
                  </span>
                ))}
              </div>
            )}

            {fbSubmitted && fbInput.trim().toLowerCase() !== card.fill_blank_answer.toLowerCase() && (
              <div className="mt-6 p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-center">
                <p className="text-sm text-red-400/80 mb-1">Incorrect. The correct answer was:</p>
                <p className="text-lg font-bold text-white">{card.fill_blank_answer}</p>
              </div>
            )}
            {fbSubmitted && fbInput.trim().toLowerCase() === card.fill_blank_answer.toLowerCase() && (
              <div className="mt-6 p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-center">
                <p className="text-emerald-400 font-bold">Excellent!</p>
              </div>
            )}
          </div>

          {card.fill_blank_question && (
            <button 
              onClick={fbSubmitted ? handleFbNext : handleFbSubmit} 
              disabled={!fbInput.trim()}
              className="mt-4 px-8 py-3 rounded-2xl bg-indigo-500 text-white font-medium hover:bg-indigo-600 transition-colors disabled:opacity-50"
            >
              {fbSubmitted ? (currentIndex + 1 >= items.length ? "Submit Session" : "Next →") : "Check Answer"}
            </button>
          )}
        </div>
      )}

    </div>
  );
}

export default function SessionPage() {
  return (
    <Suspense fallback={
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
      </div>
    }>
      <SessionRunner />
    </Suspense>
  );
}
