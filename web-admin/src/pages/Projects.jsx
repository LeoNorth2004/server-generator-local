/**
 * 项目管理页
 * 1. 列表 + 创建 / 编辑 / 删除
 * 2. 下载 / 重新生成操作
 * 3. 通过 Modal 子组件复用弹窗
 */
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { projectAPI, operationsAPI, generatorAPI } from '../api';
import { GlassCard, StatCard } from '../components/Cards';

const MODAL_TYPES = {
  NONE: null,
  CREATE: 'create',
  EDIT: 'edit',
  DELETE: 'delete',
};

const EMPTY_PROJECT = { name: '', description: '' };
const ZIP_MIME = 'application/zip';
const DOWNLOAD_SUCCESS_STATUS = 'generated';
const ISO_DATE_PREFIX_LENGTH = 10;

const ACTION_LABELS = {
  download: 'download',
};

const ICON_PROPS = {
  fill: 'none',
  viewBox: '0 0 24 24',
  stroke: 'currentColor',
  strokeLinecap: 'round',
  strokeLinejoin: 'round',
  strokeWidth: 2,
};

const ActionIcon = ({ d, className }) => (
  <svg {...ICON_PROPS} className={className}>
    <path d={d} />
  </svg>
);

const ACTION_ICONS = {
  edit: { d: 'M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z', className: 'w-5 h-5 text-blue-500' },
  download: { d: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4', className: 'w-5 h-5' },
  delete: { d: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16', className: 'w-5 h-5 text-red-500' },
};

export default function Projects() {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalType, setModalType] = useState(MODAL_TYPES.NONE);
  const [draftProject, setDraftProject] = useState(EMPTY_PROJECT);
  const [editingProject, setEditingProject] = useState(null);
  const [pendingDeleteId, setPendingDeleteId] = useState(null);

  const fetchProjects = useCallback(async () => {
    setLoading(true);
    try {
      const response = await projectAPI.list();
      setProjects(response.data?.data ?? response.data ?? []);
    } catch (error) {
      console.error('Failed to fetch projects:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const openCreateModal = useCallback(() => {
    setDraftProject(EMPTY_PROJECT);
    setModalType(MODAL_TYPES.CREATE);
  }, []);

  const openEditModal = useCallback((project) => {
    setEditingProject(project);
    setModalType(MODAL_TYPES.EDIT);
  }, []);

  const openDeleteModal = useCallback((id) => {
    setPendingDeleteId(id);
    setModalType(MODAL_TYPES.DELETE);
  }, []);

  const closeModal = useCallback(() => {
    setModalType(MODAL_TYPES.NONE);
    setEditingProject(null);
    setPendingDeleteId(null);
  }, []);

  const handleCreate = useCallback(async () => {
    if (!draftProject.name.trim()) return;
    try {
      await projectAPI.create(draftProject);
      closeModal();
      fetchProjects();
    } catch (error) {
      console.error('Failed to create project:', error);
    }
  }, [draftProject, fetchProjects, closeModal]);

  const handleUpdate = useCallback(async () => {
    if (!editingProject?.name.trim()) return;
    try {
      await projectAPI.update(editingProject.id, editingProject);
      closeModal();
      fetchProjects();
    } catch (error) {
      console.error('Failed to update project:', error);
    }
  }, [editingProject, fetchProjects, closeModal]);

  const handleDelete = useCallback(async () => {
    if (!pendingDeleteId) return;
    try {
      await projectAPI.delete(pendingDeleteId);
      closeModal();
      fetchProjects();
    } catch (error) {
      console.error('Failed to delete project:', error);
    }
  }, [pendingDeleteId, fetchProjects, closeModal]);

  const triggerDownload = (url, filename) => {
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  };

  const recordDownloadLog = (project, status, details, errorMsg) => {
    return operationsAPI.recordOperationLog({
      action: ACTION_LABELS.download,
      resource: 'project',
      resource_id: project.id,
      details: JSON.stringify(details),
      status,
      duration: 0,
      ...(errorMsg ? { error: errorMsg } : {}),
    }).catch(() => {});
  };

  const handleDownload = useCallback(async (project) => {
    if (project.status !== DOWNLOAD_SUCCESS_STATUS) {
      alert(t('projects.downloadNotReady'));
      return;
    }
    const startTime = Date.now();
    try {
      const response = await generatorAPI.download(project.id);
      const blob = new Blob([response.data], { type: ZIP_MIME });
      const url = window.URL.createObjectURL(blob);
      const filename = `${project.name || 'project'}_${new Date().toISOString().slice(0, ISO_DATE_PREFIX_LENGTH)}.zip`;
      triggerDownload(url, filename);

      recordDownloadLog(project, 'success', { project_name: project.name });
    } catch (error) {
      const msg = error.response?.data?.message || error.message;
      recordDownloadLog(project, 'failed', { project_name: project.name, error: msg }, msg);
      alert(`${t('projects.downloadFailed')}: ${msg}`);
    }
  }, [t]);

  const stats = useMemo(() => {
    const total = projects.length;
    const generated = projects.filter((p) => p.status === 'generated').length;
    const pending = projects.filter((p) => p.status === 'pending').length;
    return { total, generated, pending };
  }, [projects]);

  return (
    <div className="space-y-6">
      <PageHeader onCreate={openCreateModal} navigate={navigate} t={t} />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
        <StatCard title={t('projects.totalProjects')} value={stats.total} />
        <StatCard title={t('projects.generated')} value={stats.generated} />
        <StatCard title={t('projects.pending')} value={stats.pending} />
      </div>

      <GlassCard>
        <ProjectTable
          projects={projects}
          loading={loading}
          onEdit={openEditModal}
          onDownload={handleDownload}
          onDelete={openDeleteModal}
          navigate={navigate}
          t={t}
        />
      </GlassCard>

      {modalType === MODAL_TYPES.CREATE && (
        <ProjectFormModal
          title={t('projects.createProject')}
          name={draftProject.name}
          description={draftProject.description}
          onChangeName={(value) => setDraftProject((p) => ({ ...p, name: value }))}
          onChangeDescription={(value) => setDraftProject((p) => ({ ...p, description: value }))}
          onSubmit={handleCreate}
          onCancel={closeModal}
          submitLabel={t('common.create')}
        />
      )}

      {modalType === MODAL_TYPES.EDIT && editingProject && (
        <ProjectFormModal
          title={t('projects.editProject')}
          name={editingProject.name}
          description={editingProject.description}
          onChangeName={(value) => setEditingProject((p) => ({ ...p, name: value }))}
          onChangeDescription={(value) => setEditingProject((p) => ({ ...p, description: value }))}
          onSubmit={handleUpdate}
          onCancel={closeModal}
          submitLabel={t('common.save')}
          withLabels
        />
      )}

      {modalType === MODAL_TYPES.DELETE && (
        <DeleteConfirmModal
          onConfirm={handleDelete}
          onCancel={closeModal}
          t={t}
        />
      )}
    </div>
  );
}

function PageHeader({ onCreate, navigate, t }) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('projects.title')}</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">{t('projects.subtitle')}</p>
      </div>
      <button
        onClick={() => navigate('/generator')}
        className="btn-primary flex items-center gap-2"
      >
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
        </svg>
        {t('projects.newProject')}
      </button>
    </div>
  );
}

function ProjectTable({ projects, loading, onEdit, onDownload, onDelete, navigate, t }) {
  if (loading) {
    return <TableSpinner />;
  }
  if (projects.length === 0) {
    return <EmptyState navigate={navigate} t={t} />;
  }
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead>
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('projects.projectName')}</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.creator')}</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.description')}</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.status')}</th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.createdAt')}</th>
            <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.actions')}</th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-[#111] divide-y divide-gray-200 dark:divide-gray-700">
          {projects.map((project) => (
            <ProjectRow
              key={project.id}
              project={project}
              onEdit={onEdit}
              onDownload={onDownload}
              onDelete={onDelete}
              navigate={navigate}
              t={t}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TableSpinner() {
  return (
    <div className="px-6 py-12 text-center text-gray-400 dark:text-gray-500">
      <svg className="w-8 h-8 animate-spin mx-auto text-primary-500" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
    </div>
  );
}

function EmptyState({ navigate, t }) {
  return (
    <div className="px-6 py-12 text-center text-gray-400 dark:text-gray-500">
      <p>{t('projects.noProjects')}</p>
      <button onClick={() => navigate('/generator')} className="mt-3 btn-secondary text-sm px-4 py-2">
        {t('projects.createFirstProject')}
      </button>
    </div>
  );
}

function ProjectRow({ project, onEdit, onDownload, onDelete, navigate, t }) {
  const isGenerated = project.status === 'generated';
  return (
    <tr className="hover:bg-gray-50 dark:hover:bg-[#111] transition-colors">
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center gap-3">
          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${isGenerated ? 'bg-green-100 dark:bg-green-900/30' : 'bg-yellow-100 dark:bg-yellow-900/30'}`}>
            <svg className={`w-5 h-5 ${isGenerated ? 'text-green-600' : 'text-yellow-600'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <span className="font-medium text-gray-900 dark:text-white">{project.name}</span>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
            <span className="text-sm font-medium text-blue-600 dark:text-blue-400">
              {(project.user?.username || '?')[0].toUpperCase()}
            </span>
          </div>
          <div>
            <span className="text-sm font-medium text-gray-900 dark:text-white block">
              {project.user?.username || '-'}
            </span>
            <span className="text-xs text-gray-500 dark:text-gray-400">
              {project.user?.role === 'admin' ? '管理员' : '用户'}
            </span>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 max-w-[200px]">
        <span className="text-sm text-gray-500 dark:text-gray-400 truncate block">
          {project.description || '-'}
        </span>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${isGenerated ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400' : 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'}`}>
          {isGenerated ? t('common.generated') : t('common.pending')}
        </span>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
        {new Date(project.created_at).toLocaleDateString()}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-right">
        <div className="flex items-center justify-end gap-2">
          <button onClick={() => navigate(`/generator?projectId=${project.id}`)} className="p-2 hover:bg-blue-100 dark:hover:bg-blue-900/30 rounded-lg transition-colors" title={t('common.edit')}>
            <ActionIcon {...ACTION_ICONS.edit} />
          </button>
          <button
            onClick={() => onDownload(project)}
            className={`p-2 rounded-lg transition-colors ${isGenerated ? 'hover:bg-green-100 dark:hover:bg-green-900/30' : 'opacity-40 cursor-not-allowed'}`}
            title={isGenerated ? t('projects.download') : t('projects.downloadNotReady')}
          >
            <ActionIcon d={ACTION_ICONS.download.d} className={`w-5 h-5 ${isGenerated ? 'text-green-500' : 'text-gray-400'}`} />
          </button>
          <button onClick={() => onDelete(project.id)} className="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg transition-colors" title={t('common.delete')}>
            <ActionIcon {...ACTION_ICONS.delete} />
          </button>
        </div>
      </td>
    </tr>
  );
}

function ProjectFormModal({ title, name, description, onChangeName, onChangeDescription, onSubmit, onCancel, submitLabel, withLabels = false }) {
  return (
    <ModalWrapper onClose={onCancel}>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>
      {withLabels ? (
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Project Name</label>
            <input type="text" value={name} onChange={(e) => onChangeName(e.target.value)} className="input-field" placeholder="Project Name" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Description</label>
            <textarea value={description} onChange={(e) => onChangeDescription(e.target.value)} className="input-field" placeholder="Description" rows={3} />
          </div>
        </div>
      ) : (
        <>
          <input type="text" value={name} onChange={(e) => onChangeName(e.target.value)} className="input-field" placeholder="Project Name" />
          <textarea value={description} onChange={(e) => onChangeDescription(e.target.value)} className="input-field" placeholder="Description" rows={3} />
        </>
      )}
      <div className="flex gap-3 pt-2">
        <button onClick={onCancel} className="flex-1 btn-secondary">Cancel</button>
        <button onClick={onSubmit} disabled={!name.trim()} className="flex-1 btn-primary disabled:opacity-50">{submitLabel}</button>
      </div>
    </ModalWrapper>
  );
}

function DeleteConfirmModal({ onConfirm, onCancel, t }) {
  return (
    <ModalWrapper onClose={onCancel} size="sm">
      <h3 className="text-lg font-semibold text-red-600 dark:text-red-400">{t('projects.deleteConfirmTitle')}</h3>
      <p className="text-gray-600 dark:text-gray-400 text-sm">{t('projects.deleteConfirmMessage')}</p>
      <div className="flex gap-3 pt-2">
        <button onClick={onCancel} className="flex-1 btn-secondary">{t('common.cancel')}</button>
        <button onClick={onConfirm} className="flex-1 bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors">{t('common.delete')}</button>
      </div>
    </ModalWrapper>
  );
}

function ModalWrapper({ children, onClose, size = 'md' }) {
  const maxWidth = size === 'sm' ? 'max-w-sm' : 'max-w-md';
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className={`bg-white dark:bg-[#111] rounded-xl shadow-xl w-full ${maxWidth} p-6 space-y-4`}>
        {children}
      </div>
    </div>
  );
}
