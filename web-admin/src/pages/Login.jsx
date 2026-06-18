/**
 * 登录 / 注册页
 * 1. 统一表单状态（isLogin / isRegister）
 * 2. 抽离 handleLogin / handleRegister 业务逻辑
 * 3. 表单验证统一入口
 */
import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { authAPI, userAPI } from '../api';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';
import LanguageToggle from '../components/LanguageToggle';

const FORM_MODE = { LOGIN: 'login', REGISTER: 'register' };
const DEFAULT_EMAIL_DOMAIN = '@generator-platform.com';
const MIN_PASSWORD_LENGTH = 6;
const RESET_FORM = {
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
};

const INPUT_CLASS = 'input-field';
const BUTTON_BASE = 'w-full btn-primary py-3 disabled:opacity-50 disabled:cursor-not-allowed';
const ERROR_CLASS = 'p-3 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800';

const Spinner = () => (
  <span className="flex items-center justify-center gap-2">
    <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
    </svg>
  </span>
);

export default function Login() {
  const [mode, setMode] = useState(FORM_MODE.LOGIN);
  const [form, setForm] = useState(RESET_FORM);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const { login } = useAuth();
  const { toggleTheme } = useTheme();
  const { t } = useTranslation();
  const navigate = useNavigate();

  const isLogin = mode === FORM_MODE.LOGIN;

  const updateField = useCallback((key, value) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  }, []);

  const resetForm = useCallback(() => {
    setForm(RESET_FORM);
    setError('');
  }, []);

  const switchMode = useCallback(() => {
    setMode((prev) => (prev === FORM_MODE.LOGIN ? FORM_MODE.REGISTER : FORM_MODE.LOGIN));
    resetForm();
  }, [resetForm]);

  const extractAuthData = (response) => {
    const data = response?.data?.data ?? response?.data;
    if (!data) throw new Error('No data in response');
    if (!data.token) throw new Error('Missing token');
    if (!data.user) throw new Error('Missing user');
    return data;
  };

  const validatePassword = (pwd, confirmPwd) => {
    if (pwd !== confirmPwd) {
      setError(t('auth.passwordMismatch'));
      return false;
    }
    if (pwd.length < MIN_PASSWORD_LENGTH) {
      setError(t('auth.passwordTooShort'));
      return false;
    }
    return true;
  };

  const handleLogin = useCallback(
    async (e) => {
      e.preventDefault();
      setError('');
      setLoading(true);
      try {
        const response = await authAPI.login({ username: form.username, password: form.password });
        const { user: userData, token } = extractAuthData(response);
        login(userData, token);
        navigate('/');
      } catch {
        setError(t('auth.loginFailed'));
      } finally {
        setLoading(false);
      }
    },
    [form.username, form.password, login, navigate, t]
  );

  const handleRegister = useCallback(
    async (e) => {
      e.preventDefault();
      setError('');

      if (!validatePassword(form.password, form.confirmPassword)) return;

      setLoading(true);
      try {
        const email = form.email || `${form.username}${DEFAULT_EMAIL_DOMAIN}`;
        await userAPI.create({
          username: form.username,
          password: form.password,
          email,
          role: 'user',
        });
        const response = await authAPI.login({ username: form.username, password: form.password });
        const { user: userData, token } = extractAuthData(response);
        login(userData, token);
        navigate('/');
      } catch (err) {
        const msg = err.response?.data?.message || err.message;
        setError(`${t('auth.registerFailed')}: ${msg}`);
      } finally {
        setLoading(false);
      }
    },
    [form, login, navigate, t]
  );

  const handleSubmit = isLogin ? handleLogin : handleRegister;

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-primary-100 dark:from-gray-950 dark:via-gray-900 dark:to-gray-950 flex items-center justify-center p-4 relative">
      <TopBar toggleTheme={toggleTheme} />

      <div className="w-full max-w-md">
        <BrandSection isLogin={isLogin} />
        <FormCard>
          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className={ERROR_CLASS}>
                <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
              </div>
            )}

            <Field
              label={t('auth.username')}
              type="text"
              value={form.username}
              onChange={(e) => updateField('username', e.target.value)}
              placeholder={t('auth.username')}
              required
            />

            {!isLogin && (
              <Field
                label={t('auth.email')}
                type="email"
                value={form.email}
                onChange={(e) => updateField('email', e.target.value)}
                placeholder={t('auth.emailPlaceholder')}
              />
            )}

            <Field
              label={t('auth.password')}
              type="password"
              value={form.password}
              onChange={(e) => updateField('password', e.target.value)}
              placeholder={t('auth.password')}
              required
              minLength={MIN_PASSWORD_LENGTH}
            />

            {!isLogin && (
              <Field
                label={t('auth.confirmPassword')}
                type="password"
                value={form.confirmPassword}
                onChange={(e) => updateField('confirmPassword', e.target.value)}
                placeholder={t('auth.confirmPasswordPlaceholder')}
                required
                minLength={MIN_PASSWORD_LENGTH}
              />
            )}

            <button type="submit" disabled={loading} className={BUTTON_BASE}>
              {loading ? (
                <Spinner />
              ) : isLogin ? (
                t('auth.signIn')
              ) : (
                t('auth.signUp')
              )}
            </button>
          </form>

          <div className="mt-6 text-center space-y-3">
            <button onClick={switchMode} className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 font-medium transition-colors">
              {isLogin ? t('auth.noAccount') : t('auth.hasAccount')}
            </button>
            {isLogin && (
              <p className="text-xs text-gray-400 dark:text-gray-500">{t('auth.defaultCredentials')}</p>
            )}
          </div>
        </FormCard>

        <p className="text-center text-sm text-gray-500 dark:text-gray-400 mt-6">
          Generator Platform © 2024. All rights reserved.
        </p>
      </div>
    </div>
  );
}

function TopBar({ toggleTheme }) {
  return (
    <div className="absolute top-6 right-6 flex items-center gap-3">
      <LanguageToggle />
      <button onClick={toggleTheme} className="p-2 rounded-xl transition-colors" style={{ backgroundColor: 'var(--hover-bg)' }}>
        <svg className="w-5 h-5 text-yellow-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      </button>
    </div>
  );
}

function BrandSection({ isLogin }) {
  return (
    <div className="text-center mb-8">
      <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 shadow-lg shadow-primary-500/30 mb-4">
        <svg className="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      </div>
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Generator Platform</h1>
      <p className="text-gray-500 dark:text-gray-400 mt-2">{isLogin ? 'Sign in to your account' : 'Create a new account'}</p>
    </div>
  );
}

function FormCard({ children }) {
  return <div className="glass-card p-8">{children}</div>;
}

function Field({ label, type, value, onChange, placeholder, required, minLength }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{label}</label>
      <input
        type={type}
        value={value}
        onChange={onChange}
        className={INPUT_CLASS}
        placeholder={placeholder}
        required={required}
        minLength={minLength}
      />
    </div>
  );
}
