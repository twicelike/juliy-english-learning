"use client";

import React, { useEffect } from "react";
import { useAuthStore } from "./auth-store";
import { authAPI } from "./api";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { setUser } = useAuthStore();

  useEffect(() => {
    let mounted = true;

    const initAuth = async () => {
      try {
        const user = await authAPI.getMe();
        if (mounted) {
          setUser(user);
        }
      } catch {
        if (mounted) {
          setUser(null);
        }
      }
    };

    initAuth();

    return () => {
      mounted = false;
    };
  }, [setUser]);

  return <>{children}</>;
}
