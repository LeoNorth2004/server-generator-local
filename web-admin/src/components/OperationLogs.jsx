import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import api from '../api';

// 常量配置
const LOGS_PER_PAGE = 5;
const DEFAULT_FILTERS = { action: '', resource: '', status: '' };

// 格式化工具函数
const formatTime = (timeStr) => {
  if (!timeStr) return '-';
  return new Date(timeStr).toLocaleString();
};

const formatDuration = (ms) => {
  if (!ms && ms !== 0) return '<1ms';
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
};

const normalizeIPv6 = (ip) => {
  if (!ip) return '-';
  if (ip === '::1' || ip === '0000:0000:0000:0000:0000:0000:0000:0001') {
    return '127.0.0.1';
  }
  return ip;
};

const parseDetails = (details) => {
  if (!details) return null;
  try {
    return typeof details === 'string' ? JSON.parse(details) : details;
  } catch {
    return null;
  }
};

const sortLogsByTime = (logs) => [...logs].sort((a, b) => 
  new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
);

const calculateTotalPages = (total, pageSize) => Math.ceil(total / pageSize) || 1;

// 状态徽章配置
const STATUS_CONFIG = {
  success: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-800 dark:text-green-400' },
  failed: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-800 dark:text-red-400' },
  error: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-800 dark:text-red-400' }
};

// 资源类型映射
const RESOURCE_MAP = (t) => ({
  project: t('operations.project'),
  code: t('operations.code'),
  auth: t('operations.auth'),
  user: t('operations.user'),
  system: t('operations.system'),
});

// 操作类型映射
const ACTION_MAP = (t) => ({
  generate: t('operations.generateCode'),
  regenerate: t('operations.regenerateCode'),
  download: t('operations.downloadCode'),
  preview: t('operations.previewCode'),
  login: t('operations.login'),
  register: t('operations.register')
});

// 主组件
export default function OperationLogs() {
  const { t } = useTranslation();
  
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(LOGS_PER_PAGE);
  const [total, setTotal] = useState(0);
  const [filters, setFilters] = useState(DEFAULT_FILTERS);

  const totalPages = useMemo(() => calculateTotalPages(total, pageSize), [total, pageSize]);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      params.append('page', page.toString());
      params.append('page_size', pageSize.toString());
      
      if (filters.action) params.append('action', filters.action);
      if (filters.resource) params.append('resource', filters.resource);
      if (filters.status) params.append('status', filters.status);

      const response = await api.get(`/operations/operation-logs?${params.toString()}`);

      if (response.data.code === 0 || response.data.code === 200) {
        const data = response.data.data;
        const rawLogs = data.items || data.list || [];
        const sortedLogs = sortLogsByTime(rawLogs);
        
        setLogs(sortedLogs);
        setTotal(data.total || 0);
      } else {
        setError(t('operations.loadFailed'));
      }
    } catch (err) {
      console.error('Failed to fetch operation logs:', err);
      setError(t('operations.loadFailed'));
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, filters, t]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleFilterChange = useCallback((key, value) => {
    setFilters(prev => ({ ...prev, [key]: value }));
    setPage(1);
  }, []);

  return (
    <div className="space-y-6">
      <LogsHeader onRefresh={fetchLogs} loading={loading} t={t} />
      <FilterSection filters={filters} onFilterChange={handleFilterChange} t={t} />
      <LogsTable 
        logs={logs} 
        loading={loading} 
        error={error} 
        page={page}
        totalPages={totalPages}
        total={total}
        pageSize={pageSize}
        onPageChange={setPage}
        t={t}
      />
    </div>
  );
}

// 日志页头组件
function LogsHeader({ onRefresh, loading, t }) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          {t('operations.operationLogs')}
        </h2>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {t('operations.operationLogsSubtitle')}
        </p>
      </div>
      <button
        onClick={onRefresh}
        disabled={loading}
        className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors disabled:opacity-50"
      >
        {t('common.refresh')}
      </button>
    </div>
  );
}

// 筛选区域组件
function FilterSection({ filters, onFilterChange, t }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 p-4 bg-white dark:bg-[#111] rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
      <FilterSelect
        label={t('operations.filterByAction')}
        value={filters.action}
        onChange={(v) => onFilterChange('action', v)}
        options={[
          { value: '', label: t('operations.allActions') },
          { value: 'generate', label: t('operations.generateCode') },
          { value: 'regenerate', label: t('operations.regenerateCode') },
          { value: 'download', label: t('operations.downloadCode') },
          { value: 'preview', label: t('operations.previewCode') },
          { value: 'login', label: t('operations.login') },
        ]}
      />
      <FilterSelect
        label={t('operations.filterByResource')}
        value={filters.resource}
        onChange={(v) => onFilterChange('resource', v)}
        options={[
          { value: '', label: t('operations.allResources') },
          { value: 'project', label: t('operations.project') },
          { value: 'code', label: t('operations.code') },
        ]}
      />
      <FilterSelect
        label={t('operations.filterByStatus')}
        value={filters.status}
        onChange={(v) => onFilterChange('status', v)}
        options={[
          { value: '', label: t('operations.allStatuses') },
          { value: 'success', label: t('operations.success') },
          { value: 'failed', label: t('operations.failed') },
          { value: 'error', label: t('operations.error') },
        ]}
      />
    </div>
  );
}

// 筛选选择器组件
function FilterSelect({ label, value, onChange, options }) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
        {label}
      </label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 dark:bg-gray-700 dark:text-white"
      >
        {options.map(opt => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  );
}

// 日志表格组件
function LogsTable({ logs, loading, error, page, totalPages, total, pageSize, onPageChange, t }) {
  return (
    <div className="bg-white dark:bg-[#111] rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      {loading ? (
        <LoadingState t={t} />
      ) : error ? (
        <ErrorState error={error} />
      ) : logs.length === 0 ? (
        <EmptyState t={t} />
      ) : (
        <>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-[#111]">
                <tr>
                  <HeaderCell label="ID" />
                  <HeaderCell label={t('operations.user')} />
                  <HeaderCell label={t('operations.action')} />
                  <HeaderCell label={t('operations.resource')} />
                  <HeaderCell label={t('operations.details')} />
                  <HeaderCell label={t('operations.status')} />
                  <HeaderCell label={t('operations.duration')} />
                  <HeaderCell label={t('operations.ipAddress')} />
                  <HeaderCell label={t('operations.time')} />
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-[#111] divide-y divide-gray-200 dark:divide-gray-700">
                {logs.map((log) => (
                  <LogRow key={log.id} log={log} t={t} />
                ))}
              </tbody>
            </table>
          </div>
          <Pagination
            page={page}
            totalPages={totalPages}
            total={total}
            pageSize={pageSize}
            onPageChange={onPageChange}
            t={t}
          />
        </>
      )}
    </div>
  );
}

// 表头单元格
function HeaderCell({ label }) {
  return (
    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
      {label}
    </th>
  );
}

// 日志行组件
function LogRow({ log, t }) {
  return (
    <tr className="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        #{log.id}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        <div className="font-medium">{log.username}</div>
        <div className="text-xs text-gray-500 dark:text-gray-400">ID: {log.user_id}</div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        {getActionDisplay(log.action, t)}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        {getResourceDisplay(log.target_type, t)}
      </td>
      <td className="px-6 py-4 text-sm text-gray-900 dark:text-gray-100">
        <DetailsCell details={log.details} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <StatusBadge status={log.status} t={t} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
        {formatDuration(log.duration) || '<1ms'}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
        {normalizeIPv6(log.ip_address)}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
        {formatTime(log.created_at)}
      </td>
    </tr>
  );
}

// 操作类型显示
function getActionDisplay(action, t) {
  const actionLabels = ACTION_MAP(t);
  return actionLabels[action] || action;
}

// 资源类型显示
function getResourceDisplay(resource, t) {
  const resourceLabels = RESOURCE_MAP(t);
  return resourceLabels[resource] || resource || '-';
}

// 详情单元格
function DetailsCell({ details }) {
  const parsed = parseDetails(details);
  if (!parsed) return <span className="text-sm">{String(details)}</span>;
  
  return (
    <div className="text-xs text-gray-600 dark:text-gray-400 max-w-md">
      {Object.entries(parsed).map(([key, value]) => (
        <div key={key}>
          <span className="font-medium">{key}:</span> {String(value)}
        </div>
      ))}
    </div>
  );
}

// 状态徽章
function StatusBadge({ status, t }) {
  const config = STATUS_CONFIG[status] || STATUS_CONFIG.success;
  const label = status === 'success' ? t('operations.success') : 
                status === 'failed' ? t('operations.failed') : 
                status === 'error' ? t('operations.error') : status;
  
  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${config.bg} ${config.text}`}>
      {label}
    </span>
  );
}

// 加载状态
function LoadingState({ t }) {
  return (
    <div className="flex items-center justify-center h-64">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      <span className="ml-3 text-gray-500 dark:text-gray-400">{t('operations.loadingLogs')}</span>
    </div>
  );
}

// 错误状态
function ErrorState({ error }) {
  return (
    <div className="flex items-center justify-center h-64 text-red-500">
      {error}
    </div>
  );
}

// 空状态
function EmptyState({ t }) {
  return (
    <div className="flex items-center justify-center h-64 text-gray-400 dark:text-gray-500">
      {t('operations.noOperationLogs')}
    </div>
  );
}

// 分页组件
function Pagination({ page, totalPages, total, pageSize, onPageChange, t }) {
  const handlePrev = () => onPageChange(p => Math.max(1, p - 1));
  const handleNext = () => onPageChange(p => p + 1);
  
  return (
    <div className="px-6 py-3 flex items-center justify-between border-t border-gray-200 dark:border-gray-700">
      <div className="text-sm text-gray-700 dark:text-gray-300">
        {t('operations.pageInfo').replace('{current}', page).replace('{total}', total)}
      </div>
      <div className="flex items-center space-x-2">
        <span className="text-sm text-gray-500 dark:text-gray-400">
          {t('operations.pageSize').replace('{count}', pageSize)}
        </span>
        <PageButton
          onClick={handlePrev}
          disabled={page <= 1}
          label={t('operations.previousPage')}
        />
        <PageButton
          onClick={handleNext}
          disabled={page >= totalPages || total === 0}
          label={t('operations.nextPage')}
        />
      </div>
    </div>
  );
}

// 分页按钮
function PageButton({ onClick, disabled, label }) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded text-sm hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
    >
      {label}
    </button>
  );
}
