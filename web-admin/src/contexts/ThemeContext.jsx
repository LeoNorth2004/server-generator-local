/**
 * 主题上下文 - 暗色/亮色模式管理
 * 使用类单例模式 + 自定义 hook 封装
 */
import { createContext, useContext, useEffect, useState, useCallback, useMemo } from 'react';

const ThemeContext = createContext(null);

const THEME_STORAGE_KEY = 'theme';
const DARK_CLASS = 'dark';

const getInitialTheme = () => {
  const saved = localStorage.getItem(THEME_STORAGE_KEY);
  if (saved) return saved === 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
};

const applyThemeToDOM = (isDark) => {
  const root = document.documentElement;
  if (isDark) {
    root.classList.add(DARK_CLASS);
    localStorage.setItem(THEME_STORAGE_KEY, 'dark');
  } else {
    root.classList.remove(DARK_CLASS);
    localStorage.setItem(THEME_STORAGE_KEY, 'light');
  }
};

export function ThemeProvider({ children }) {
  const [isDark, setIsDark] = useState(getInitialTheme);

  useEffect(() => {
    applyThemeToDOM(isDark);
  }, [isDark]);

  const toggleTheme = useCallback(() => {
    setIsDark((prev) => !prev);
  }, []);

  const setDarkMode = useCallback((value) => {
    setIsDark(value);
  }, []);

  const value = useMemo(
    () => ({ isDark, toggleTheme, setDarkMode }),
    [isDark, toggleTheme, setDarkMode]
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
