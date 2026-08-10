import apiClient from "./api-client";

export interface DashboardDTO {
  new: number;
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
  choices?: ChoiceDTO[];
}

export interface SessionDTO {
  mode: "flashcard" | "quiz";
  items: SessionItemDTO[];
}

export type ProgressQuality = 1 | 3 | 4 | 5;

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

  submitProgress: async (wordId: string, quality: ProgressQuality, responseTimeMs?: number): Promise<void> => {
    await apiClient.post("/learning/progress", { 
      word_id: wordId, 
      quality,
      response_time_ms: responseTimeMs
    });
  },
};
