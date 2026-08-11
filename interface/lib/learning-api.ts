import apiClient from "./api-client";

export interface DashboardDTO {
  new: number;
  pending: number;
  review: number;
  learned: number;
  due_today: number;
  mature_words: number;
  accuracy: number;
}

export interface ChoiceDTO {
  definition: string;
  is_correct: boolean;
}

export interface SessionItemDTO {
  word_id: string;
  word: string;
  pos: string;
  ipa: string;
  definitions: string[];
  example: string;
  choices: ChoiceDTO[]; // MCQ
  fill_blank_question: string; // FB
  fill_blank_answer: string; // FB
}

export type SessionStatus = "preparing" | "active" | "completed" | "expired";

export interface SessionDTO {
  id: string;
  mode: "flashcard" | "quiz" | "new_words" | "review";
  status: SessionStatus;
  current_index: number;
  items: SessionItemDTO[];
}

export type ProgressQuality = 1 | 3 | 4 | 5;

export interface WordResult {
  word_id: string;
  flashcard_quality: ProgressQuality;
  multiple_choice_correct?: boolean;
  fill_blank_correct?: boolean;
}

export const LearningAPI = {
  getDashboard: async (): Promise<DashboardDTO> => {
    const res = await apiClient.get("/learning/dashboard");
    return res.data.data;
  },

  getSession: async (mode: "flashcard" | "quiz", limit = 10): Promise<SessionDTO> => {
    const res = await apiClient.get("/learning/session", {
      params: { mode, limit },
    });
    return res.data.data;
  },

  createSession: async (mode: "new_words" | "review"): Promise<SessionDTO> => {
    const res = await apiClient.post("/learning/session", { mode });
    return res.data.data;
  },

  debugAdvanceTime: async (days: number = 1): Promise<void> => {
    await apiClient.post("/learning/debug/advance-time", { days });
  },

  getSessionById: async (id: string): Promise<SessionDTO> => {
    const res = await apiClient.get(`/learning/session/${id}`);
    return res.data.data;
  },

  submitSessionResults: async (sessionId: string, results: WordResult[]): Promise<void> => {
    await apiClient.post(`/learning/sessions/${sessionId}/submit`, { results });
  },

  suspendWord: async (wordId: string): Promise<void> => {
    await apiClient.put(`/learning/progress/suspend/${wordId}`);
  },
};
