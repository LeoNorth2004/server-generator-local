/**
 * 用户管理页
 * 1. 用户列表 + 添加/编辑/删除
 * 2. 权限矩阵展示
 * 3. 通过独立的 UserFormModal 复用添加/编辑表单
 */
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { userAPI } from '../api';
import { GlassCard, StatCard } from '../components/Cards';

const MODAL_TYPES = { NONE: null, ADD: 'add', EDIT: 'edit' };
const ROLES = { ADMIN: 'admin', USER: 'user' };
const DEFAULT_NEW_USER = { username: '', email: '', password: '', role: ROLES.USER };
const TREND_VALUE = 5;

const PERMISSION_MATRIX = [
  { permKey: 'users.permViewDashboard', admin: true, user: true },
  { permKey: 'users.permManageProjects', admin: true, user: true },
  { permKey: 'users.permGenerateCode', admin: true, user: true },
  { permKey: 'users.permDownloadCode', admin: true, user: true },
  { permKey: 'users.permViewDocs', admin: true, user: true },
  { permKey: 'users.permManageUsers', admin: true, user: false },
  { permKey: 'users.permViewOperations', admin: true, user: false },
  { permKey: 'users.permRegenerateProject', admin: true, user: true },
];

const ICON_PROPS = {
  fill: 'none',
  viewBox: '0 0 24 24',
  stroke: 'currentColor',
  strokeLinecap: 'round',
  strokeLinejoin: 'round',
  strokeWidth: 2,
};

const ActionIcon = ({ d, className = 'w-5 h-5' }) => (
  <svg {...ICON_PROPS} className={className}>
    <path d={d} />
  </svg>
);

const ICON_PATHS = {
  add: 'M12 4v16m8-8H4',
  edit: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z',
  close: 'M6 18L18 6M6 6l12 12',
  delete: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16',
  chevron: 'M19 9l-7 7-7-7',
};

export default function Users() {
  const { t } = useTranslation();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [modalType, setModalType] = useState(MODAL_TYPES.NONE);
  const [newUser, setNewUser] = useState(DEFAULT_NEW_USER);
  const [editingUser, setEditingUser] = useState(null);
  const [showPermissionMatrix, setShowPermissionMatrix] = useState(false);

  const fetchUsers = useCallback(async () => {
    try {
      const response = await userAPI.list();
      const data = response.data?.data;
      if (Array.isArray(data)) setUsers(data);
      else if (Array.isArray(response.data)) setUsers(response.data);
      else setUsers([]);
    } catch (error) {
      console.error('Failed to fetch users:', error);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const openAddModal = useCallback(() => {
    setNewUser(DEFAULT_NEW_USER);
    setModalType(MODAL_TYPES.ADD);
  }, []);

  const openEditModal = useCallback((user) => {
    setEditingUser({ ...user });
    setModalType(MODAL_TYPES.EDIT);
  }, []);

  const closeModal = useCallback(() => {
    setModalType(MODAL_TYPES.NONE);
    setEditingUser(null);
  }, []);

  const handleAddUser = useCallback(async () => {
    if (!newUser.username || !newUser.password) return;
    setSaving(true);
    try {
      const response = await userAPI.create(newUser);
      const created = response.data?.data ?? response.data;
      if (created) setUsers((prev) => [...prev, created]);
      closeModal();
    } catch (error) {
      console.error('Failed to add user:', error);
      alert('添加用户失败');
    } finally {
      setSaving(false);
    }
  }, [newUser, closeModal]);

  const handleEditUser = useCallback(async () => {
    if (!editingUser) return;
    setSaving(true);
    try {
      const payload = { email: editingUser.email, role: editingUser.role };
      if (editingUser.password) payload.password = editingUser.password;
      const response = await userAPI.update(editingUser.id, payload);
      const updated = response.data?.data ?? response.data;
      if (updated) setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
      closeModal();
    } catch (error) {
      console.error('Failed to update user:', error);
      alert('更新用户失败');
    } finally {
      setSaving(false);
    }
  }, [editingUser, closeModal]);

  const handleDeleteUser = useCallback(async (userId) => {
    if (!window.confirm('确定要删除这个用户吗？')) return;
    try {
      await userAPI.delete(userId);
      setUsers((prev) => prev.filter((u) => u.id !== userId));
    } catch (error) {
      console.error('Failed to delete user:', error);
      alert('删除用户失败');
    }
  }, []);

  const stats = useMemo(
    () => ({
      total: users.length,
      admin: users.filter((u) => u.role === ROLES.ADMIN).length,
      active: users.length,
    }),
    [users]
  );

  return (
    <div className="space-y-6">
      <PageHeader onAdd={openAddModal} t={t} />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatCard title={t('users.totalUsers')} value={stats.total} trend={TREND_VALUE} />
        <StatCard title={t('users.administrators')} value={stats.admin} />
        <StatCard title={t('users.activeUsers')} value={stats.active} />
      </div>

      <PermissionMatrixCard
        expanded={showPermissionMatrix}
        onToggle={() => setShowPermissionMatrix((v) => !v)}
        t={t}
      />

      <UserTable users={users} loading={loading} onEdit={openEditModal} onDelete={handleDeleteUser} t={t} />

      {modalType === MODAL_TYPES.ADD && (
        <UserFormModal
          title={t('users.addUser')}
          onCancel={closeModal}
          onSubmit={handleAddUser}
          saving={saving}
          username={newUser.username}
          onUsernameChange={(value) => setNewUser((p) => ({ ...p, username: value }))}
          password={newUser.password}
          onPasswordChange={(value) => setNewUser((p) => ({ ...p, password: value }))}
          email={newUser.email}
          onEmailChange={(value) => setNewUser((p) => ({ ...p, email: value }))}
          role={newUser.role}
          onRoleChange={(value) => setNewUser((p) => ({ ...p, role: value }))}
          t={t}
        />
      )}

      {modalType === MODAL_TYPES.EDIT && editingUser && (
        <UserFormModal
          title={t('users.editUser')}
          onCancel={closeModal}
          onSubmit={handleEditUser}
          saving={saving}
          isEdit
          username={editingUser.username}
          password={editingUser.password || ''}
          onPasswordChange={(value) => setEditingUser((p) => ({ ...p, password: value }))}
          email={editingUser.email || ''}
          onEmailChange={(value) => setEditingUser((p) => ({ ...p, email: value }))}
          role={editingUser.role}
          onRoleChange={(value) => setEditingUser((p) => ({ ...p, role: value }))}
          t={t}
        />
      )}
    </div>
  );
}

function PageHeader({ onAdd, t }) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('users.title')}</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">{t('users.subtitle')}</p>
      </div>
      <button onClick={onAdd} className="btn-primary flex items-center gap-2">
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={ICON_PATHS.add} />
        </svg>
        {t('users.addUser')}
      </button>
    </div>
  );
}

function PermissionMatrixCard({ expanded, onToggle, t }) {
  return (
    <GlassCard>
      <div className="flex items-center justify-between mb-4 cursor-pointer" onClick={onToggle}>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">📋 {t('users.permissionMatrix')}</h3>
        <svg
          className={`w-5 h-5 text-gray-500 transition-transform ${expanded ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={ICON_PATHS.chevron} />
        </svg>
      </div>

      {expanded && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200 dark:border-gray-700">
                <th className="text-left py-2 px-3 font-medium text-gray-600 dark:text-gray-400">{t('users.permission')}</th>
                <th className="text-center py-2 px-3 font-medium text-purple-600 dark:text-purple-400">{t('common.admin')}</th>
                <th className="text-center py-2 px-3 font-medium text-blue-600 dark:text-blue-400">{t('common.user')}</th>
              </tr>
            </thead>
            <tbody>
              {PERMISSION_MATRIX.map((row, index) => (
                <tr key={index} className="border-b border-gray-100 dark:border-gray-800">
                  <td className="py-2 px-3 text-gray-700 dark:text-gray-300">{t(row.permKey)}</td>
                  <td className="py-2 px-3 text-center">{row.admin ? '✅' : '❌'}</td>
                  <td className="py-2 px-3 text-center">{row.user ? '✅' : '❌'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </GlassCard>
  );
}

function UserTable({ users, loading, onEdit, onDelete, t }) {
  return (
    <GlassCard>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700">
              <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">{t('users.user')}</th>
              <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">{t('common.email')}</th>
              <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">{t('common.role')}</th>
              <th className="text-left py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">{t('users.created')}</th>
              <th className="text-right py-3 px-4 text-sm font-medium text-gray-500 dark:text-gray-400">{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan="5" className="py-8 text-center text-gray-500">{t('common.loading')}</td>
              </tr>
            ) : users.length === 0 ? (
              <tr>
                <td colSpan="5" className="py-8 text-center text-gray-500">{t('users.noUsers')}</td>
              </tr>
            ) : (
              users.map((user) => (
                <UserRow key={user.id} user={user} onEdit={onEdit} onDelete={onDelete} t={t} />
              ))
            )}
          </tbody>
        </table>
      </div>
    </GlassCard>
  );
}

function UserRow({ user, onEdit, onDelete, t }) {
  return (
    <tr className="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-[#111]">
      <td className="py-4 px-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center">
            <span className="text-sm font-medium text-white">{user.username?.charAt(0).toUpperCase()}</span>
          </div>
          <span className="font-medium text-gray-900 dark:text-white">{user.username}</span>
        </div>
      </td>
      <td className="py-4 px-4 text-gray-600 dark:text-gray-400">{user.email}</td>
      <td className="py-4 px-4">
        <span
          className={`px-3 py-1 text-xs font-medium rounded-full ${
            user.role === ROLES.ADMIN
              ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
              : 'bg-gray-100 text-gray-700 dark:bg-[#111] dark:text-gray-300'
          }`}
        >
          {t(`common.${user.role}`)}
        </span>
      </td>
      <td className="py-4 px-4 text-gray-600 dark:text-gray-400">
        {new Date(user.created_at).toLocaleDateString()}
      </td>
      <td className="py-4 px-4 text-right">
        <div className="flex items-center justify-end gap-2">
          <button onClick={() => onEdit(user)} className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
            <ActionIcon d={ICON_PATHS.edit} className="w-5 h-5 text-gray-500" />
          </button>
          <button onClick={() => onDelete(user.id)} className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
            <ActionIcon d={ICON_PATHS.delete} className="w-5 h-5 text-red-500" />
          </button>
        </div>
      </td>
    </tr>
  );
}

function UserFormModal({
  title,
  onCancel,
  onSubmit,
  saving,
  isEdit = false,
  username,
  onUsernameChange,
  password,
  onPasswordChange,
  email,
  onEmailChange,
  role,
  onRoleChange,
  t,
}) {
  return (
    <ModalWrapper onClose={onCancel}>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>
        <button onClick={onCancel} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
          <ActionIcon d={ICON_PATHS.close} className="w-6 h-6" />
        </button>
      </div>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('auth.username')}</label>
          <input
            type="text"
            value={username || ''}
            onChange={(e) => onUsernameChange?.(e.target.value)}
            disabled={isEdit}
            className={`input-field ${isEdit ? 'bg-gray-100 dark:bg-gray-700 cursor-not-allowed' : ''}`}
            placeholder={t('auth.username')}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {isEdit ? t('auth.newPassword') : t('auth.password')}
          </label>
          <input
            type="password"
            value={password}
            onChange={(e) => onPasswordChange(e.target.value)}
            className="input-field"
            placeholder={isEdit ? t('auth.newPasswordPlaceholder') : t('auth.password')}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('common.email')}</label>
          <input
            type="email"
            value={email}
            onChange={(e) => onEmailChange(e.target.value)}
            className="input-field"
            placeholder={t('common.email')}
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('common.role')}</label>
          <select value={role} onChange={(e) => onRoleChange(e.target.value)} className="input-field">
            <option value={ROLES.USER}>{t('common.user')}</option>
            <option value={ROLES.ADMIN}>{t('common.admin')}</option>
          </select>
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={onCancel} className="flex-1 btn-secondary">{t('common.cancel')}</button>
          <button
            onClick={onSubmit}
            disabled={saving || (!isEdit && (!username || !password))}
            className="flex-1 btn-primary disabled:opacity-50"
          >
            {saving ? <SaveSpinner label={t('common.saving')} /> : t('common.save')}
          </button>
        </div>
      </div>
    </ModalWrapper>
  );
}

function SaveSpinner({ label }) {
  return (
    <span className="flex items-center justify-center gap-2">
      <svg className="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
      {label}
    </span>
  );
}

function ModalWrapper({ children, onClose }) {
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="bg-white dark:bg-[#111] rounded-xl p-6 w-full max-w-md">{children}</div>
    </div>
  );
}
