/**
 * 认证上下文 - 用户登录状态管理
 * 提供 login/logout/getMe 等能力，统一维护 token 与 user 信息
 */
import { createContext, useContext, useState, useCallback, useMemo } from 'react';

const AuthContext = createContext(null);

const STORAGE_KEYS = {
  token: 'token',
  user: 'user',
};

const safeParseUser = (raw) => {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    localStorage.removeItem(STORAGE_KEYS.user);
    return null;
  }
};

const getInitialState = () => ({
  user: safeParseUser(localStorage.getItem(STORAGE_KEYS.user)),
  token: localStorage.getItem(STORAGE_KEYS.token),
});

export function AuthProvider({ children }) {
  const [{ user, token }, setAuthState] = useState(getInitialState);

  const login = useCallback((userData, authToken) => {
    setAuthState({ user: userData, token: authToken });
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(userData));
    localStorage.setItem(STORAGE_KEYS.token, authToken);
  }, []);

  const logout = useCallback(() => {
    setAuthState({ user: null, token: null });
    localStorage.removeItem(STORAGE_KEYS.user);
    localStorage.removeItem(STORAGE_KEYS.token);
  }, []);

  const updateUser = useCallback((patch) => {
    setAuthState((prev) => {
      const nextUser = typeof patch === 'function' ? patch(prev.user) : { ...prev.user, ...patch };
      if (nextUser) {
        localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(nextUser));
      }
      return { ...prev, user: nextUser };
    });
  }, []);

  const isAuthenticated = Boolean(token);

  const value = useMemo(
    () => ({ user, token, login, logout, updateUser, isAuthenticated }),
    [user, token, login, logout, updateUser, isAuthenticated]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
