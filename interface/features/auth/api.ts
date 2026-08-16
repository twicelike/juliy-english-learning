import apiClient from "@/lib/api-client";
import { User } from "./types";

export interface SessionResponse {
  id: string;
  user_agent: string;
  ip_address: string;
  created_at: string;
  expires_at: string;
  is_current: boolean;
}

export const authAPI = {
  getMe: async (): Promise<User> => {
    const res = await apiClient.get("/auth/me");
    return res.data.data;
  },

  login: async (credentials: Record<string, string>): Promise<User> => {
    const res = await apiClient.post("/auth/login", credentials);
    return res.data.data;
  },

  register: async (data: Record<string, string>): Promise<User> => {
    const res = await apiClient.post("/auth/register", data);
    return res.data.data;
  },

  logout: async (): Promise<void> => {
    await apiClient.post("/auth/logout");
  },
  
  logoutAll: async (): Promise<void> => {
    await apiClient.post("/auth/logout-all");
  },

  verifyEmail: async (token: string): Promise<void> => {
    await apiClient.post("/auth/verify-email", { token });
  },

  resendVerification: async (email: string): Promise<void> => {
    await apiClient.post("/auth/resend-verification", { email });
  },

  forgotPassword: async (email: string): Promise<void> => {
    await apiClient.post("/auth/forgot-password", { email });
  },

  resetPassword: async (token: string, new_password: string): Promise<void> => {
    await apiClient.post("/auth/reset-password", { token, new_password });
  },

  getSessions: async (): Promise<SessionResponse[]> => {
    const res = await apiClient.get("/auth/sessions");
    return res.data.data;
  },

  revokeSession: async (sessionId: string): Promise<void> => {
    await apiClient.delete(`/auth/sessions/${sessionId}`);
  },
};
