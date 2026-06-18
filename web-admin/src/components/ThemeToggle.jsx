/**
 * 主题切换开关（开关式 UI）
 */
import { useTheme } from '../contexts/ThemeContext';

const TRACK_BASE = 'relative w-14 h-7 rounded-full transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2';
const TRACK_DARK = 'dark:bg-gray-700';
const TRACK_LIGHT = 'bg-gray-200';

const KNOB_BASE = 'absolute top-1 left-1 w-5 h-5 rounded-full bg-white dark:bg-primary-600 shadow-md transform transition-transform duration-300 flex items-center justify-center';
const KNOB_ACTIVE = 'translate-x-7';
const KNOB_INACTIVE = 'translate-x-0';

const SUN_PATH =
  'M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z';

const MOON_PATH =
  'M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z';

const ICON_BASE = 'w-3 h-3';

const SunIcon = () => (
  <svg className={`${ICON_BASE} text-yellow-500`} fill="currentColor" viewBox="0 0 24 24">
    <path d={SUN_PATH} />
  </svg>
);

const MoonIcon = () => (
  <svg className={`${ICON_BASE} text-gray-600`} fill="currentColor" viewBox="0 0 24 24">
    <path d={MOON_PATH} />
  </svg>
);

export default function ThemeToggle() {
  const { isDark, toggleTheme } = useTheme();
  const trackClass = `${TRACK_BASE} ${isDark ? TRACK_DARK : TRACK_LIGHT}`;
  const knobClass = `${KNOB_BASE} ${isDark ? KNOB_ACTIVE : KNOB_INACTIVE}`;

  return (
    <button onClick={toggleTheme} className={trackClass}>
      <span className={knobClass}>{isDark ? <SunIcon /> : <MoonIcon />}</span>
    </button>
  );
}
