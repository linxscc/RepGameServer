import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { getLang, setLang as setStoredLang, t, type Lang } from '../i18n';

interface LanguageContextType {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: string, fallback?: string) => string;
}

const LanguageContext = createContext<LanguageContextType>({
  lang: 'en',
  setLang: () => {},
  t: (key: string, _fallback?: string) => key,
});

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(getLang());

  const setLang = useCallback((newLang: Lang) => {
    setStoredLang(newLang);
    setLangState(newLang);
  }, []);

  const translate = useCallback((key: string, fallback?: string) => {
    return t(key, fallback);
  }, [lang]);

  useEffect(() => {
    const handleStorage = (e: StorageEvent) => {
      if (e.key === 'voyara_lang') {
        setLangState(getLang());
      }
    };
    window.addEventListener('storage', handleStorage);
    return () => window.removeEventListener('storage', handleStorage);
  }, []);

  return (
    <LanguageContext.Provider value={{ lang, setLang, t: translate }}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLanguage() {
  return useContext(LanguageContext);
}
