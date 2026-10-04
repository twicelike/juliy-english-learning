export const ROLES = {
  SUPERIOR: "superior",
  ADMIN: "admin",
  USER: "user",
} as const;

export const hasAdminAccess = (role?: string) => 
  role === ROLES.ADMIN || role === ROLES.SUPERIOR;

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  email_verified: boolean;
  has_google_linked: boolean;
  role: string;
}

export type AuthStatus = "loading" | "authenticated" | "unauthenticated";

export interface AuthState {
  user: User | null;
  status: AuthStatus;
  setUser: (user: User | null) => void;
  setStatus: (status: AuthStatus) => void;
  clearAuth: () => void;
}
