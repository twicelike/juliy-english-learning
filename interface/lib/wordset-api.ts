import apiClient from "./api-client";
import { WordDTO } from "./dictionary-api";

export type WordSetVisibility = "private" | "public";

export type WordSetDifficulty = "Beginner" | "Intermediate" | "Advanced";

export interface WordSet {
  id: string;
  name: string;
  description: string;
  visibility: WordSetVisibility;
  topics: string[];
  difficulty: WordSetDifficulty;
  source_language: string;
  annotation_language: string;
  word_count: number;
  clone_count: number;
  created_at: string;
  updated_at: string;
}

export interface Language {
  code: string;
  name: string;
}

export interface SavedSenseRef {
  word_id: string;
  sense_index: number;
}

export interface WordSetListItem extends WordSet {
  contains_word?: boolean;
  saved_senses?: SavedSenseRef[];
  creator_name?: string;
}

export interface WordSetListResponse {
  items: WordSetListItem[];
  total: number;
}

export interface WordSetDetailResponse extends WordSet {
  item_ids: string[];
  saved_senses: SavedSenseRef[];
  is_owner: boolean;
  creator_name?: string;
}

export interface CreateWordSetRequest {
  name: string;
  description: string;
  visibility: WordSetVisibility;
  topics?: string[];
  difficulty?: WordSetDifficulty;
  source_language?: string;
  annotation_language?: string;
}

export interface UpdateWordSetRequest {
  name?: string;
  description?: string;
  visibility?: WordSetVisibility;
  topics?: string[];
  difficulty?: WordSetDifficulty;
  annotation_language?: string;
}

export interface ExploreWordSetsRequest {
  q?: string;
  topic?: string;
  difficulty?: string;
  sort?: string;
  page?: number;
  limit?: number;
}

export interface ExploreWordSetsResponse {
  items: WordSetListItem[];
  page: number;
  limit: number;
  total: number;
  has_more: boolean;
}

export type ProgressState = "new" | "learning" | "due" | "mastered";

export interface WordViewerItemDTO extends WordDTO {
  progress_state: ProgressState;
  sense_index?: number;
  translation_candidates?: string[];
  annotation?: {
    text: string;
    source: string;
    language_code: string;
  };
}

export const WordSetAPI = {
  getWordSets: async (containsWordIds?: string[]): Promise<WordSetListResponse> => {
    let url = "/word-sets";
    if (containsWordIds && containsWordIds.length > 0) {
      const params = containsWordIds.map(id => `contains_word=${encodeURIComponent(id)}`).join("&");
      url = `/word-sets?${params}`;
    }
    const res = await apiClient.get<WordSetListResponse>(url);
    return res.data;
  },

  getTopics: async (): Promise<string[]> => {
    const res = await apiClient.get<{ topics: string[] }>("/word-sets/topics");
    return res.data.topics;
  },

  getSupportedLanguages: async (): Promise<Language[]> => {
    const res = await apiClient.get<{ languages: Language[] }>("/word-sets/languages");
    return res.data.languages;
  },

  exploreWordSets: async (params: ExploreWordSetsRequest): Promise<ExploreWordSetsResponse> => {
    const query = new URLSearchParams();
    if (params.q) query.append("q", params.q);
    if (params.topic) query.append("topic", params.topic);
    if (params.difficulty) query.append("difficulty", params.difficulty);
    if (params.sort) query.append("sort", params.sort);
    if (params.page) query.append("page", params.page.toString());
    if (params.limit) query.append("limit", params.limit.toString());
    
    const res = await apiClient.get<ExploreWordSetsResponse>(`/word-sets/explore?${query.toString()}`);
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

  cloneWordSet: async (id: string): Promise<WordSet> => {
    const res = await apiClient.post<{data: WordSet}>(`/word-sets/${id}/clone`);
    return res.data.data;
  },

  addWordToSet: async (setId: string, wordId: string, senseIndex: number): Promise<void> => {
    await apiClient.post(`/word-sets/${setId}/words`, { word_id: wordId, sense_index: senseIndex });
  },

  removeWordFromSet: async (setId: string, wordId: string): Promise<void> => {
    await apiClient.delete(`/word-sets/${setId}/words/${wordId}`);
  },

  removeWordsFromSet: async (
    setId: string,
    wordIds: string[]
  ): Promise<{ deleted_word_ids: string[]; word_count: number }> => {
    const res = await apiClient.delete<{ deleted_word_ids: string[]; word_count: number }>(
      `/word-sets/${setId}/words`,
      { data: { word_ids: wordIds } }
    );
    return res.data;
  },

  restoreWordsToSet: async (
    setId: string,
    wordIds: string[]
  ): Promise<{ restored_word_ids: string[]; word_count: number }> => {
    const res = await apiClient.post<{ restored_word_ids: string[]; word_count: number }>(
      `/word-sets/${setId}/words/restore`,
      { word_ids: wordIds }
    );
    return res.data;
  },

  reorderWords: async (setId: string, wordIds: string[]) => {
    const res = await apiClient.patch(`/word-sets/${setId}/words/reorder`, { word_ids: wordIds });
    return res.data;
  },

  getWordSetWords: async (setId: string): Promise<WordViewerItemDTO[]> => {
    const res = await apiClient.get<WordViewerItemDTO[]>(`/word-sets/${setId}/words`);
    return res.data;
  },

  bulkAddWords: async (setId: string, items: {word_id: string; sense_index: number}[]): Promise<{ added: number; word_count: number; limit: number }> => {
    const res = await apiClient.post(`/word-sets/${setId}/words/bulk`, { items });
    return res.data;
  },
};
