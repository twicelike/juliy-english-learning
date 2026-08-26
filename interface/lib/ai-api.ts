import apiClient from "@/lib/api-client";

// ─── Request types ────────────────────────────────────────────────────────────

export interface WordAIRequest {
  word: string;
  pos?: string;
  context?: string;
  userLevel?: string;      // A1 | A2 | B1 | B2 | C1 | C2
  nativeLanguage?: string; // e.g. "Vietnamese"
}

export interface CompareAIRequest {
  word1: string;
  word2: string;
  pos?: string;
  userLevel?: string;
  nativeLanguage?: string;
}

export interface TextAIRequest {
  text: string;
  userLevel?: string;
  nativeLanguage?: string;
}

export interface FillBlankAIRequest {
  words: Array<{ word: string; pos?: string }>;
  userLevel?: string;
  nativeLanguage?: string;
}

// ─── Response types ───────────────────────────────────────────────────────────

export interface ExplainSenseRequest {
  wordId: string;
  senseId: string;
  userLevel?: string;
  nativeLanguage?: string;
}

export interface BatchSenseExplanation {
  senseId: string;
  explanation: string;
  examples: string[];
  commonMistakes: string[];
  usageNote: string;
}

export interface NuanceResponse {
  coreMeaning: string;
  subtleMeaning: string;
  usageNotes: string[];
  confusedWith: string[];
}

export interface CompareResponse {
  word1: string;
  word2: string;
  similarity: string;
  differences: Array<{ aspect: string; word1: string; word2: string }>;
  recommendedUsage: string;
}

export interface ExamplesResponse {
  word: string;
  examples: string[];
}



export interface GrammarMistake {
  wrong: string;
  correct: string;
  reason: string;
  explanation: string;
}

export interface GrammarResponse {
  score: number;
  correctedText: string;
  mistakes: GrammarMistake[];
  tips: string[];
}

export interface VocabSuggestion {
  original: string;
  suggested: string;
  reason: string;
  example: string;
}

export interface VocabSuggestionResponse {
  originalText: string;
  rewrittenText: string;
  suggestions: VocabSuggestion[];
}

export interface FillBlankResponse {
  chosenWord: string;
  pos: string;
  sentence: string;
  answer: string;
  hint: string;
}

// ─── Default context ──────────────────────────────────────────────────────────
const DEFAULT_CTX = { userLevel: "B1", nativeLanguage: "Vietnamese" };

// ─── API client ───────────────────────────────────────────────────────────────
export const AIAPI = {
  explainSense: (req: ExplainSenseRequest) =>
    apiClient
      .post<{ data: BatchSenseExplanation }>("/ai/words/explain-sense", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),

  explainNuance: (req: WordAIRequest) =>
    apiClient
      .post<{ data: NuanceResponse }>("/ai/words/nuance", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),

  compareWords: (req: CompareAIRequest) =>
    apiClient
      .post<{ data: CompareResponse }>("/ai/words/compare", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),

  generateExamples: (req: WordAIRequest) =>
    apiClient
      .post<{ data: ExamplesResponse }>("/ai/words/examples", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),



  checkGrammar: (req: TextAIRequest) =>
    apiClient
      .post<{ data: GrammarResponse }>("/ai/writing/check", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),

  suggestVocab: (req: TextAIRequest) =>
    apiClient
      .post<{ data: VocabSuggestionResponse }>("/ai/writing/vocab", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),

  generateFillBlank: (req: FillBlankAIRequest) =>
    apiClient
      .post<{ data: FillBlankResponse }>("/ai/quiz/fill-blank", { ...DEFAULT_CTX, ...req })
      .then((r) => r.data.data),
};
