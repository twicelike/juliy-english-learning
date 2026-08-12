import apiClient from "./api-client";
import { WordDTO } from "./dictionary-api";

export type WordSetVisibility = "private" | "public";

export interface WordSet {
  id: string;
  name: string;
  description: string;
  visibility: WordSetVisibility;
  word_count: number;
  created_at: string;
  updated_at: string;
}

export interface WordSetListItem extends WordSet {
  contains_word?: boolean;
}

export interface WordSetListResponse {
  items: WordSetListItem[];
  total: number;
}

export interface WordSetDetailResponse extends WordSet {
  item_ids: string[];
}

export interface CreateWordSetRequest {
  name: string;
  description: string;
  visibility: WordSetVisibility;
}

export interface UpdateWordSetRequest {
  name?: string;
  description?: string;
  visibility?: WordSetVisibility;
}

export type ProgressState = "new" | "learning" | "due" | "mastered";

export interface WordViewerItemDTO extends WordDTO {
  progress_state: ProgressState;
}

export const WordSetAPI = {
  getWordSets: async (containsWordId?: string): Promise<WordSetListResponse> => {
    const url = containsWordId ? `/word-sets?contains_word=${containsWordId}` : "/word-sets";
    const res = await apiClient.get<WordSetListResponse>(url);
    return res.data;
  },

  getWordSet: async (id: string): Promise<WordSetDetailResponse> => {
    const res = await apiClient.get<WordSetDetailResponse>(`/word-sets/${id}`);
    return res.data;
  },

  createWordSet: async (req: CreateWordSetRequest): Promise<WordSet> => {
    const res = await apiClient.post<WordSet>("/word-sets", req);
    return res.data;
  },

  updateWordSet: async (id: string, req: UpdateWordSetRequest): Promise<WordSet> => {
    const res = await apiClient.patch<WordSet>(`/word-sets/${id}`, req);
    return res.data;
  },

  deleteWordSet: async (id: string): Promise<void> => {
    await apiClient.delete(`/word-sets/${id}`);
  },

  addWordToSet: async (setId: string, wordId: string): Promise<void> => {
    await apiClient.post(`/word-sets/${setId}/words`, { word_id: wordId });
  },

  removeWordFromSet: async (setId: string, wordId: string): Promise<void> => {
    await apiClient.delete(`/word-sets/${setId}/words/${wordId}`);
  },

  reorderWords: async (setId: string, wordIds: string[]) => {
    const res = await apiClient.patch(`/word-sets/${setId}/words/reorder`, { word_ids: wordIds });
    return res.data;
  },

  getWordSetWords: async (setId: string): Promise<WordViewerItemDTO[]> => {
    const res = await apiClient.get<WordViewerItemDTO[]>(`/word-sets/${setId}/words`);
    return res.data;
  },

  bulkAddWords: async (setId: string, wordIds: string[]): Promise<{ added: number; word_count: number; limit: number }> => {
    const res = await apiClient.post(`/word-sets/${setId}/words/bulk`, { word_ids: wordIds });
    return res.data;
  },
};
