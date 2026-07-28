"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ProfileAPI, PlacementQuestionDTO, SubmitTestResponse } from "@/lib/profile-api";

export default function PlacementTestPage() {
  const router = useRouter();
  const [questions, setQuestions] = useState<PlacementQuestionDTO[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<SubmitTestResponse | null>(null);

  useEffect(() => {
    ProfileAPI.getPlacementTest()
      .then((data) => {
        setQuestions(data.questions);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Failed to load placement test", err);
        setLoading(false);
      });
  }, []);

  const handleSelectOption = (questionId: string, option: string) => {
    setAnswers((prev) => ({ ...prev, [questionId]: option }));
  };

  const handleNext = () => {
    if (currentIndex < questions.length - 1) {
      setCurrentIndex((prev) => prev + 1);
    } else {
      handleSubmit();
    }
  };

  const handleSubmit = async () => {
    setSubmitting(true);
    try {
      const resp = await ProfileAPI.submitPlacementTest({ answers });
      setResult(resp);
    } catch (err) {
      console.error("Failed to submit", err);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0a0a0f] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (result) {
    return (
      <div className="min-h-screen bg-[#0a0a0f] text-white flex flex-col items-center justify-center p-6">
        <div className="max-w-md text-center space-y-6">
          <div className="text-6xl mb-4">🎉</div>
          <h1 className="text-3xl font-bold font-display">Assessment Complete!</h1>
          <p className="text-white/70">
            You scored {result.score} / 15. Based on this, your estimated level is:
          </p>
          <div className="text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-cyan-400">
            {result.cefr}
          </div>
          <button
            onClick={() => router.push("/learn")}
            className="w-full mt-8 bg-indigo-500 hover:bg-indigo-600 text-white font-medium py-4 rounded-2xl transition-all"
          >
            Start Your Roadmap
          </button>
        </div>
      </div>
    );
  }

  const currentQ = questions[currentIndex];
  if (!currentQ) return null;

  const isLast = currentIndex === questions.length - 1;
  const currentAnswer = answers[currentQ.id];

  return (
    <div className="min-h-screen bg-[#0a0a0f] text-white flex flex-col p-6 max-w-2xl mx-auto">
      {/* Progress Bar */}
      <div className="w-full bg-white/5 rounded-full h-2 mb-12 mt-8">
        <div
          className="bg-indigo-500 h-2 rounded-full transition-all duration-300"
          style={{ width: `${((currentIndex + 1) / questions.length) * 100}%` }}
        />
      </div>

      <div className="text-sm font-medium text-indigo-400 mb-2 tracking-wider uppercase">
        Question {currentIndex + 1} of {questions.length}
      </div>
      <h2 className="text-2xl font-medium leading-relaxed mb-8">
        {currentQ.question.split("________").map((part, i, arr) => (
          <span key={i}>
            {part}
            {i !== arr.length - 1 && (
              <span className="inline-block w-16 border-b-2 border-white/40 mx-2" />
            )}
          </span>
        ))}
      </h2>

      <div className="space-y-3 flex-1">
        {currentQ.options.map((opt) => {
          const isSelected = currentAnswer === opt;
          return (
            <button
              key={opt}
              onClick={() => handleSelectOption(currentQ.id, opt)}
              className={`w-full text-left px-6 py-4 rounded-2xl border-2 transition-all ${
                isSelected
                  ? "border-indigo-500 bg-indigo-500/10 text-white"
                  : "border-white/10 hover:border-white/20 text-white/80"
              }`}
            >
              {opt}
            </button>
          );
        })}
      </div>

      <button
        onClick={handleNext}
        disabled={!currentAnswer || submitting}
        className="w-full mt-8 bg-white text-black hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed font-medium py-4 rounded-2xl transition-all"
      >
        {submitting ? "Evaluating..." : isLast ? "Finish Assessment" : "Next Question"}
      </button>
    </div>
  );
}
