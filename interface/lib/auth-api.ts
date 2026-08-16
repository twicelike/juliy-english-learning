import apiClient from "./api-client";
import { User } from "@/features/auth";

export interface BookmarkDTO {
  word: string;
  added_at: string;
}

export const AuthAPI = {
  getProfile: async (): Promise<User> => {
    const res = await apiClient.get("/users/me");
    return res.data.data;
  },

  getBookmarks: async (): Promise<BookmarkDTO[]> => {
    const res = await apiClient.get("/users/bookmarks");
    return res.data.data;
  },

  addBookmark: async (word: string) => {
    await apiClient.post("/users/bookmarks", { word });
  },

  removeBookmark: async (word: string) => {
    await apiClient.delete(`/users/bookmarks/${encodeURIComponent(word)}`);
  },
};
