/**
 * 国际化上下文 - 中英文切换
 * 封装 i18next 的 changeLanguage 操作，并持久化到 localStorage
 */
import { createContext, useContext, useState, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';

const I18nContext = createContext(null);

const LANG_STORAGE_KEY = 'language';
const SUPPORTED_LANGS = ['zh', 'en'];

const getInitialLanguage = () => {
  const stored = localStorage.getItem(LANG_STORAGE_KEY);
  return SUPPORTED_LANGS.includes(stored) ? stored : 'zh';
};

const persistLanguage = (lang) => {
  localStorage.setItem(LANG_STORAGE_KEY, lang);
};

export function I18nProvider({ children }) {
  const { i18n } = useTranslation();
  const [language, setLanguage] = useState(getInitialLanguage);

  const changeLanguage = useCallback(
    (lang) => {
      if (!SUPPORTED_LANGS.includes(lang)) return;
      setLanguage(lang);
      i18n.changeLanguage(lang);
      persistLanguage(lang);
    },
    [i18n]
  );

  const toggleLanguage = useCallback(() => {
    const next = language === 'zh' ? 'en' : 'zh';
    changeLanguage(next);
  }, [language, changeLanguage]);

  const value = useMemo(
    () => ({ language, changeLanguage, toggleLanguage }),
    [language, changeLanguage, toggleLanguage]
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
}
