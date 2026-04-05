import { useCallback, useEffect, useState } from "react";

type Theme = "light" | "dark" | "system";

export function useTheme() {
  const [theme, setTheme] = useState<Theme>((localStorage.getItem('theme') || 'system') as Theme);

  const resolveTheme = useCallback((theme: Theme) => {
    if (theme !== 'system') return theme;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }, []);

  const applyTheme = useCallback((theme: Theme) => {
    document.documentElement.classList.remove('light', 'dark');
    document.documentElement.classList.add(resolveTheme(theme));
  }, [resolveTheme]);

  const updateTheme = useCallback((theme: Theme) => {
    localStorage.setItem('theme', theme);
    applyTheme(theme);
    setTheme(theme);
  }, [applyTheme]);

  useEffect(() => {
    const defaultTheme = (localStorage.getItem('theme') || 'system') as Theme;
    applyTheme(defaultTheme);
  }, [applyTheme]);

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => { if (theme === 'system') applyTheme(theme); };
    mq.addEventListener('change', handler);

    return () => mq.removeEventListener('change', handler);
  }, [theme, applyTheme]);

  return { theme, updateTheme };
}