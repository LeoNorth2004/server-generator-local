import { useI18n } from '../contexts/I18nContext';

export default function LanguageToggle() {
  const { language, toggleLanguage } = useI18n();

  return (
    <button
      onClick={toggleLanguage}
      className="px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
      style={{
        backgroundColor: 'rgba(147, 51, 234, 0.1)',
        color: 'rgb(147, 51, 234)'
      }}
    >
      {language === 'zh' ? '中' : 'En'}
    </button>
  );
}
