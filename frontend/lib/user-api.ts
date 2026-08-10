import apiClient from "./api-client";

export interface BookmarkRequest {
  word: string;
  type: "word" | "expression";
}

export const UserAPI = {
  addBookmark: async (req: BookmarkRequest): Promise<void> => {
    await apiClient.post("/users/bookmarks", req);
  },
};
