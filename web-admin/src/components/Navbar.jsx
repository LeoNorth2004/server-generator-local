/**
 * 顶部导航栏
 * 集中提供：品牌标识、语言切换、主题切换、用户信息与登出
 */
import { useTheme } from '../contexts/ThemeContext';
import { useAuth } from '../contexts/AuthContext';
import { useI18n } from '../contexts/I18nContext';

const BRAND_NAME = 'Generator Platform';
const LANG_BUTTON_BG = 'rgba(147, 51, 234, 0.1)';
const LANG_BUTTON_COLOR = 'rgb(147 51 234)';
const FALLBACK_USERNAME = 'User';
const FALLBACK_ROLE = 'user';

const Icon = ({ d, className = '', style }) => (
  <svg
    className={className}
    style={style}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={d} />
  </svg>
);

const PATHS = {
  bolt: 'M13 10V3L4 14h7v7l9-11h-7z',
  sun: 'M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z',
  moon: 'M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z',
  logout: 'M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1',
};

export default function Navbar() {
  const { isDark, toggleTheme } = useTheme();
  const { user, logout } = useAuth();
  const { language, toggleLanguage } = useI18n();

  const username = user?.username || FALLBACK_USERNAME;
  const role = user?.role || FALLBACK_ROLE;
  const initial = username.charAt(0).toUpperCase();

  return (
    <nav
      className="fixed top-0 left-0 right-0 h-16 z-50 flex items-center px-6"
      style={{
        backgroundColor: 'var(--sidebar-bg)',
        borderBottom: '1px solid var(--border-color)',
      }}
    >
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center">
          <Icon d={PATHS.bolt} className="w-5 h-5 text-white" />
        </div>
        <span
          className="text-lg font-semibold"
          style={{ color: 'var(--text-primary)' }}
        >
          {BRAND_NAME}
        </span>
      </div>

      <div className="flex items-center gap-4 ml-auto">
        <button
          onClick={toggleLanguage}
          className="px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
          style={{
            backgroundColor: LANG_BUTTON_BG,
            color: LANG_BUTTON_COLOR,
          }}
        >
          {language === 'zh' ? 'EN' : '中'}
        </button>

        <button
          onClick={toggleTheme}
          className="p-2 rounded-xl transition-colors"
          style={{ backgroundColor: 'var(--hover-bg)' }}
        >
          {isDark ? (
            <Icon d={PATHS.sun} className="w-5 h-5 text-yellow-500" />
          ) : (
            <Icon
              d={PATHS.moon}
              className="w-5 h-5"
              style={{ color: 'var(--text-secondary)' }}
            />
          )}
        </button>

        <div
          className="flex items-center gap-3 pl-4"
          style={{ borderLeft: '1px solid var(--border-color)' }}
        >
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center">
            <span className="text-sm font-medium text-white">{initial}</span>
          </div>
          <div className="hidden md:block">
            <p
              className="text-sm font-medium"
              style={{ color: 'var(--text-primary)' }}
            >
              {username}
            </p>
            <p className="text-xs" style={{ color: 'var(--text-muted)' }}>
              {role}
            </p>
          </div>
          <button
            onClick={logout}
            className="ml-2 p-2 rounded-xl transition-colors"
            style={{ backgroundColor: 'var(--hover-bg)' }}
          >
            <Icon
              d={PATHS.logout}
              className="w-5 h-5"
              style={{ color: 'var(--text-muted)' }}
            />
          </button>
        </div>
      </div>
    </nav>
  );
}
