import en from './en.json';
import fr from './fr.json';
import ar from './ar.json';
import ru from './ru.json';
import zh from './zh.json';

export type Lang = 'en' | 'fr' | 'ar' | 'ru' | 'zh';

const messages: Record<Lang, Record<string, string>> = { en, fr, ar, ru, zh };

const STORAGE_KEY = 'voyara_lang';

export function getLang(): Lang {
  const stored = localStorage.getItem(STORAGE_KEY) as Lang | null;
  if (stored && messages[stored]) return stored;
  const browser = navigator.language.split('-')[0] as Lang;
  if (browser && messages[browser]) return browser;
  return 'en';
}

export function setLang(lang: Lang) {
  localStorage.setItem(STORAGE_KEY, lang);
}

export function t(key: string, fallback?: string): string {
  const lang = getLang();
  return messages[lang]?.[key] ?? fallback ?? key;
}

export const supportedLangs: { code: Lang; label: string }[] = [
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'Français' },
  { code: 'ar', label: 'العربية' },
  { code: 'ru', label: 'Русский' },
  { code: 'zh', label: '中文' },
];
