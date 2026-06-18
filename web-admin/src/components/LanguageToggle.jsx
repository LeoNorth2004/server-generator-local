/**
 * 语言切换按钮
 * 封装通用样式，只关注显示与交互
 */
import { useI18n } from '../contexts/I18nContext';

const BUTTON_BG = 'rgba(147, 51, 234, 0.1)';
const BUTTON_COLOR = 'rgb(147 51 234)';

export default function LanguageToggle() {
  const { language, toggleLanguage } = useI18n();
  const label = language === 'zh' ? '中' : 'En';

  return (
    <button
      onClick={toggleLanguage}
      className="px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
      style={{
        backgroundColor: BUTTON_BG,
        color: BUTTON_COLOR,
      }}
    >
      {label}
    </button>
  );
}
