/**
 * 侧边导航栏
 * 1. 根据角色过滤菜单（普通用户不可见 users / operations）
 * 2. 高亮当前路径
 */
import { useLocation, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../contexts/AuthContext';

const ADMIN_RESTRICTED_PATHS = new Set(['/users', '/operations']);
const SIDEBAR_WIDTH_CLASS = 'w-64';

const ICON_PROPS = {
  className: 'w-5 h-5',
  fill: 'none',
  viewBox: '0 0 24 24',
  stroke: 'currentColor',
};

const Icon = ({ d }) => (
  <svg {...ICON_PROPS}>
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={d} />
  </svg>
);

const MENU_ITEMS = [
  {
    titleKey: 'nav.dashboard',
    path: '/',
    iconPath:
      'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6',
  },
  {
    titleKey: 'nav.projects',
    path: '/projects',
    iconPath:
      'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z',
  },
  {
    titleKey: 'nav.generator',
    path: '/generator',
    iconPath:
      'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4',
  },
  {
    titleKey: 'nav.docs',
    path: '/docs',
    iconPath:
      'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
  },
  {
    titleKey: 'nav.users',
    path: '/users',
    iconPath:
      'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
  },
  {
    titleKey: 'nav.operations',
    path: '/operations',
    iconPath:
      'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z',
  },
];

const isAdmin = (user) => user?.role === 'admin';

export default function Sidebar() {
  const location = useLocation();
  const { t } = useTranslation();
  const { user } = useAuth();

  const visibleItems = MENU_ITEMS.filter((item) => {
    if (!isAdmin(user) && ADMIN_RESTRICTED_PATHS.has(item.path)) {
      return false;
    }
    return true;
  });

  return (
    <aside
      className={`fixed left-0 top-16 bottom-0 ${SIDEBAR_WIDTH_CLASS} z-40 overflow-y-auto`}
      style={{
        backgroundColor: 'var(--sidebar-bg)',
        borderRight: '1px solid var(--border-color)',
      }}
    >
      <nav className="p-4 space-y-1">
        {visibleItems.map((item) => {
          const isActive = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={isActive ? 'sidebar-link-active' : 'sidebar-link'}
            >
              <Icon d={item.iconPath} />
              <span>{t(item.titleKey)}</span>
            </Link>
          );
        })}
      </nav>

      <div
        className="absolute bottom-0 left-0 right-0 p-4"
        style={{ borderTop: '1px solid var(--border-color)' }}
      >
        <div className="glass-card p-4 rounded-xl">
          <p className="text-xs mb-2" style={{ color: 'var(--text-muted)' }}>
            {t('nav.allSystemsOperational')}
          </p>
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
            <span
              className="text-sm font-medium"
              style={{ color: 'var(--text-secondary)' }}
            >
              {t('common.operational')}
            </span>
          </div>
        </div>
      </div>
    </aside>
  );
}
