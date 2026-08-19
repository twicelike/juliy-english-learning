'use client';

import { createContext, useContext, useEffect, useState } from 'react';
import type { ThemeName } from '@/lib/theme';

const STORAGE_KEY = 'app-theme';

interface ThemeContextValue {
  theme: ThemeName;
  setTheme: (theme: ThemeName) => void;
  mounted: boolean;
}

export const ThemeContext = createContext<ThemeContextValue>({
  theme: 'default',
  setTheme: () => {},
  mounted: false,
});

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<ThemeName>('default');
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const saved = localStorage.getItem(STORAGE_KEY) as ThemeName | null;
    if (saved) {
      applyTheme(saved);
    }
  }, []);

  const setTheme = (t: ThemeName) => {
    applyTheme(t);
    localStorage.setItem(STORAGE_KEY, t);
  };

  const applyTheme = (t: ThemeName) => {
    document.documentElement.setAttribute('data-theme', t);
    setThemeState(t);
  };

  // Render children immediately to avoid hydration mismatch, but some UI relying on theme might flash.
  // The 'mounted' flag allows consumers to avoid rendering theme-dependent UI until client-side hydration is complete.
  return (
    <ThemeContext.Provider value={{ theme, setTheme, mounted }}>
      {children}
    </ThemeContext.Provider>
  );
}

export const useTheme = () => useContext(ThemeContext);
