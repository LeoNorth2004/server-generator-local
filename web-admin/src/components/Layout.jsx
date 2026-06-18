/**
 * 顶层布局
 * 组合 Navbar、Sidebar、Main 三个区域
 */
import Navbar from './Navbar';
import Sidebar from './Sidebar';

const NAV_HEIGHT_OFFSET = 'pt-16';
const SIDEBAR_WIDTH = 'ml-64';
const CONTAINER_MAX_WIDTH = 'max-w-[1600px]';

export default function Layout({ children }) {
  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--main-bg)' }}>
      <Navbar />
      <Sidebar />
      <main
        className={`${SIDEBAR_WIDTH} ${NAV_HEIGHT_OFFSET} p-6 min-h-screen`}
      >
        <div className={`${CONTAINER_MAX_WIDTH} mx-auto`}>{children}</div>
      </main>
    </div>
  );
}
