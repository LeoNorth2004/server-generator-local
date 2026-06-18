/**
 * 应用入口
 * 1. 配置 React Router v7 future flags
 * 2. 组合全局 Provider
 * 3. 集中定义受保护路由
 */
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { I18nProvider } from './contexts/I18nContext';
import Layout from './components/Layout';
import Home from './pages/Home';
import Login from './pages/Login';
import Projects from './pages/Projects';
import Generator from './pages/Generator';
import Users from './pages/Users';
import Operations from './pages/Operations';
import Docs from './pages/Docs';

const ROUTE_FUTURE_FLAGS = {
  v7_startTransition: true,
  v7_relativeSplatPath: true,
};

const PROTECTED_ROUTES = [
  { path: '/', element: <Home /> },
  { path: '/projects', element: <Projects /> },
  { path: '/generator', element: <Generator /> },
  { path: '/users', element: <Users /> },
  { path: '/operations', element: <Operations /> },
  { path: '/docs', element: <Docs /> },
];

const LOGIN_PATH = '/login';
const ROOT_PATH = '/';

function ProtectedRoute({ children }) {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? children : <Navigate to={LOGIN_PATH} />;
}

function LoginRoute() {
  const { isAuthenticated } = useAuth();
  return isAuthenticated ? <Navigate to={ROOT_PATH} /> : <Login />;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path={LOGIN_PATH} element={<LoginRoute />} />
      {PROTECTED_ROUTES.map(({ path, element }) => (
        <Route
          key={path}
          path={path}
          element={
            <ProtectedRoute>
              <Layout>{element}</Layout>
            </ProtectedRoute>
          }
        />
      ))}
    </Routes>
  );
}

function App() {
  return (
    <I18nProvider>
      <ThemeProvider>
        <AuthProvider>
          <BrowserRouter future={ROUTE_FUTURE_FLAGS}>
            <AppRoutes />
          </BrowserRouter>
        </AuthProvider>
      </ThemeProvider>
    </I18nProvider>
  );
}

export default App;
