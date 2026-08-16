import { create } from "zustand";
import { AuthState } from "./types";

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  status: "loading",
  setUser: (user) => set({ user, status: user ? "authenticated" : "unauthenticated" }),
  setStatus: (status) => set({ status }),
  clearAuth: () => set({ user: null, status: "unauthenticated" }),
}));
