import apiClient from "./api-client";
import { User } from "@/store/auth-store";

export interface BookmarkDTO {
  word: string;
  added_at: string;
}

export const AuthAPI = {
  register: async (data: any) => {
    const res = await apiClient.post("/auth/register", data);
    return res.data.data;
  },
  
  login: async (data: any) => {
    const res = await apiClient.post("/auth/login", data);
    return res.data.data;
  },

  logout: async () => {
    await apiClient.post("/auth/logout");
  },

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
