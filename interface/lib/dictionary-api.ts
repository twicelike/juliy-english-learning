import apiClient from "./api-client";

export interface DefinitionDTO {
  word_id?: string;
  sense_index: number;
  partOfSpeech?: string;
  definition: string;
  examples: string[];
}

export interface WordDTO {
  id: string;
  word: string;
  pos: string;
  ipa: string;
  audio?: string;
  forms: string[];
  definitions: DefinitionDTO[];
  antonyms: string[];
  synonyms: string[];
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
}

export const DictionaryAPI = {
  search: async (q: string): Promise<WordDTO[]> => {
    if (!q) return [];
    const res = await apiClient.get<ApiResponse<WordDTO[]>>(`/words/search?q=${encodeURIComponent(q)}`);
    return res.data.data || [];
  },
  
  getDetail: async (word: string): Promise<WordDTO | null> => {
    try {
      const res = await apiClient.get<ApiResponse<WordDTO>>(`/words/detail/${encodeURIComponent(word)}`);
      return res.data.data;
    } catch (error: unknown) {
      const err = error as any;
      if (err.message === "Word not found" || err.response?.status === 404) {
        return null; // Handle not found gracefully
      }
      throw err;
    }
  },

  getRandom: async (): Promise<WordDTO | null> => {
    const res = await apiClient.get<ApiResponse<WordDTO>>(`/words/random?t=${Date.now()}`);
    return res.data.data;
  },

  getWordOfTheDay: async (): Promise<WordDTO | null> => {
    const res = await apiClient.get<ApiResponse<WordDTO>>("/words/word-of-the-day");
    return res.data.data;
  },
};
