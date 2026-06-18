/**
 * 运维监控页
 * 1. 监控 / 日志两个 Tab
 * 2. 自动保活 + 联动开关
 * 3. 报告下载（生成 markdown 报告）
 */
import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { GlassCard, ChartCard } from '../components/Cards';
import { operationsAPI } from '../api';
import { useAuth } from '../contexts/AuthContext';
import OperationLogs from '../components/OperationLogs';

const TABS = { MONITOR: 'monitor', LOGS: 'logs' };
const AUTO_REFRESH_INTERVAL = 30000;
const HEALTHY_STATUSES = new Set([
  'Running',
  'Operational',
  'running',
  'healthy',
  'Healthy',
]);
const RESOURCE_THRESHOLDS = { WARNING: 60, CRITICAL: 80 };

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
  download: 'M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4',
  refresh: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15',
};

export default function Operations() {
  const { t } = useTranslation();
  const { user: currentUser } = useAuth();

  const [loading, setLoading] = useState(false);
  const [metrics, setMetrics] = useState(null);
  const [services, setServices] = useState([]);
  const [systemResources, setSystemResources] = useState(null);
  const [activeTab, setActiveTab] = useState(TABS.MONITOR);
  const [keepAliveEnabled, setKeepAliveEnabled] = useState(false);
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState(false);
  const intervalRef = useRef(null);

  const recordOperation = useCallback(async (action, resource, details) => {
    try {
      await operationsAPI.recordOperationLog({
        action,
        resource,
        details: JSON.stringify(details),
        status: 'success',
      });
    } catch (e) {
      console.log('Failed to record operation:', e);
    }
  }, []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [statsRes, servicesRes] = await Promise.all([
        operationsAPI.stats(),
        operationsAPI.getServices(),
      ]);

      const metricsData = statsRes.data?.data || statsRes.data || {};
      const servicesRaw = servicesRes.data?.data || servicesRes.data || [];
      const servicesList = Array.isArray(servicesRaw) ? servicesRaw : [];

      const resourcesData = {
        total_services: servicesList.length,
        running_services: servicesList.filter((s) => HEALTHY_STATUSES.has(s.status)).length,
        services: servicesList.map((s) => ({
          name: s.name || s.metadata?.name,
          status: s.status,
          healthy: HEALTHY_STATUSES.has(s.status),
          cpu_usage: s.cpu_usage || 0,
          memory_usage: s.memory_usage || 0,
        })),
      };

      setMetrics(metricsData);
      setSystemResources(resourcesData);
      setServices(resourcesData.services.length > 0 ? resourcesData.services : servicesList);
    } catch (error) {
      console.error('Failed to fetch operations data:', error);
      setMetrics(null);
      setServices([]);
      setSystemResources(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const startAutoRefresh = useCallback(() => {
    if (intervalRef.current) return;
    intervalRef.current = setInterval(() => fetchData(), AUTO_REFRESH_INTERVAL);
  }, [fetchData]);

  const stopAutoRefresh = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  const handleToggleKeepAlive = useCallback((enabled) => {
    setKeepAliveEnabled(enabled);
    setAutoRefreshEnabled(enabled);
    if (enabled) {
      startAutoRefresh();
      recordOperation('keep_alive', 'system', { enabled: true });
    } else {
      stopAutoRefresh();
      recordOperation('keep_alive', 'system', { enabled: false });
    }
  }, [recordOperation, startAutoRefresh, stopAutoRefresh]);

  const handleToggleAutoRefresh = useCallback((enabled) => {
    setAutoRefreshEnabled(enabled);
    setKeepAliveEnabled(enabled);
    if (enabled) {
      startAutoRefresh();
      recordOperation('refresh', 'system', { auto_refresh: true });
    } else {
      stopAutoRefresh();
      recordOperation('refresh', 'system', { auto_refresh: false });
    }
  }, [recordOperation, startAutoRefresh, stopAutoRefresh]);

  const handleRefresh = useCallback(async () => {
    const startTime = Date.now();
    await fetchData();
    try {
      await operationsAPI.recordOperationLog({
        action: 'refresh',
        resource: 'operations',
        details: JSON.stringify({ action: 'refresh_monitor', page: 'operations' }),
        status: 'success',
        duration: Date.now() - startTime,
      });
    } catch (e) {
      console.log('Failed to record refresh log:', e);
    }
  }, [fetchData]);

  const handleDownload = useCallback(async () => {
    setLoading(true);
    try {
      const report = generateReport({ metrics, services, resources: systemResources, t });
      const blob = new Blob([report], { type: 'text/markdown' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `operations-report-${new Date().toISOString().split('T')[0]}.md`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      alert('下载成功');
    } catch (error) {
      console.error('Failed to download report:', error);
      alert('下载失败');
    } finally {
      setLoading(false);
    }
  }, [metrics, services, systemResources, t]);

  useEffect(() => {
    fetchData();
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [fetchData]);

  const resourceMetrics = useMemo(() => buildResourceMetrics(systemResources, t), [systemResources, t]);

  return (
    <div className="space-y-6">
      <OperationsHeader
        activeTab={activeTab}
        loading={loading}
        onDownload={handleDownload}
        onRefresh={handleRefresh}
        keepAliveEnabled={keepAliveEnabled}
        autoRefreshEnabled={autoRefreshEnabled}
        onToggleKeepAlive={handleToggleKeepAlive}
        onToggleAutoRefresh={handleToggleAutoRefresh}
        t={t}
      />

      <TabSwitcher activeTab={activeTab} onChange={setActiveTab} t={t} />

      {activeTab === TABS.MONITOR ? (
        <MonitorTab
          metrics={metrics}
          services={services}
          resourceMetrics={resourceMetrics}
          currentUser={currentUser}
          t={t}
        />
      ) : (
        <OperationLogs />
      )}
    </div>
  );
}

function OperationsHeader({
  activeTab,
  loading,
  onDownload,
  onRefresh,
  keepAliveEnabled,
  autoRefreshEnabled,
  onToggleKeepAlive,
  onToggleAutoRefresh,
  t,
}) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">运维监控</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">系统状态、服务健康与操作日志</p>
      </div>
      <div className="flex gap-3">
        {activeTab === TABS.MONITOR && (
          <>
            <button onClick={onDownload} className="btn-primary flex items-center gap-2" disabled={loading}>
              <ActionIcon d={ICON_PATHS.download} />
              下载
            </button>
            <button onClick={onRefresh} className="btn-secondary flex items-center gap-2" disabled={loading}>
              <ActionIcon d={ICON_PATHS.refresh} className={loading ? 'animate-spin' : ''} />
              {loading ? '加载中' : '刷新'}
            </button>
          </>
        )}
      </div>
    </div>
  );
}

function TabSwitcher({ activeTab, onChange, t }) {
  const tabClass = (active) =>
    `px-4 py-2 text-sm font-medium rounded-md transition-colors ${
      active
        ? 'bg-white dark:bg-gray-700 text-primary-600 shadow-sm'
        : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
    }`;

  return (
    <div className="flex space-x-1 bg-gray-100 dark:bg-[#111] p-1 rounded-lg w-fit">
      <button onClick={() => onChange(TABS.MONITOR)} className={tabClass(activeTab === TABS.MONITOR)}>
        系统指标
      </button>
      <button onClick={() => onChange(TABS.LOGS)} className={tabClass(activeTab === TABS.LOGS)}>
        操作日志
      </button>
    </div>
  );
}

function MonitorTab({ metrics, services, resourceMetrics, t }) {
  return (
    <>
      <MetricsGrid metrics={metrics} t={t} />

      {metrics?.runtime && <RuntimeCard runtime={metrics.runtime} t={t} />}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {resourceMetrics && resourceMetrics.length > 0 && (
          <ChartCard title="系统资源" className="lg:col-span-2">
            <div className="grid grid-cols-2 gap-6">
              {resourceMetrics.map((metric, index) => (
                <ResourceBar key={index} metric={metric} />
              ))}
            </div>
          </ChartCard>
        )}
        <ServiceHealthCard services={services} t={t} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard title="服务详细状态">
          <ServiceStatusList services={services} t={t} />
        </ChartCard>
      </div>
    </>
  );
}

function MetricsGrid({ metrics, t }) {
  const cards = [
    { key: 'total_requests', label: '累计请求', color: 'blue' },
    { key: 'active_projects', label: '活跃项目', color: 'emerald' },
    { key: 'generated_codes', label: '代码生成次数', color: 'purple' },
    { key: 'total_users', label: '注册用户', color: 'orange' },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      {cards.map((card) => (
        <MetricCard key={card.key} card={card} value={metrics?.[card.key]} />
      ))}
    </div>
  );
}

function MetricCard({ card, value }) {
  const colorMap = {
    blue: { bg: 'from-blue-50 to-indigo-100', dark: 'dark:from-blue-900/30 dark:to-indigo-900/30', border: 'border-blue-200 dark:border-blue-800', text: 'text-blue-700 dark:text-blue-300', valueText: 'text-blue-900 dark:text-blue-100', subText: 'text-blue-600 dark:text-blue-400' },
    emerald: { bg: 'from-emerald-50 to-teal-100', dark: 'dark:from-emerald-900/30 dark:to-teal-900/30', border: 'border-emerald-200 dark:border-emerald-800', text: 'text-emerald-700 dark:text-emerald-300', valueText: 'text-emerald-900 dark:text-emerald-100', subText: 'text-emerald-600 dark:text-emerald-400' },
    purple: { bg: 'from-purple-50 to-pink-100', dark: 'dark:from-purple-900/30 dark:to-pink-900/30', border: 'border-purple-200 dark:border-purple-800', text: 'text-purple-700 dark:text-purple-300', valueText: 'text-purple-900 dark:text-purple-400', subText: 'text-purple-600 dark:text-purple-400' },
    orange: { bg: 'from-orange-50 to-red-100', dark: 'dark:from-orange-900/30 dark:to-red-900/30', border: 'border-orange-200 dark:border-orange-800', text: 'text-orange-700 dark:text-orange-300', valueText: 'text-orange-900 dark:text-orange-100', subText: 'text-orange-600 dark:text-orange-400' },
  };
  const palette = colorMap[card.color];
  const iconMap = { blue: '📊', emerald: '📁', purple: '⚡', orange: '👥' };
  const subLabelMap = { '累计请求': '累计 API 请求总数', '活跃项目': '已生成的项目数量', '代码生成次数': '成功生成代码的次数', '注册用户': '系统用户总数' };

  return (
    <div className={`bg-gradient-to-br ${palette.bg} ${palette.dark} rounded-2xl p-6 border ${palette.border}`}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
        <span className={`text-sm font-medium ${palette.text}`}>{card.label}</span>
        <span style={{ fontSize: '20px' }}>{iconMap[card.color]}</span>
      </div>
      <p className={`text-4xl font-bold ${palette.valueText}`}>{formatValue(value, '0', '')}</p>
      <p className={`text-xs ${palette.subText} mt-2`}>{subLabelMap[card.label] || ''}</p>
    </div>
  );
}

function RuntimeCard({ runtime, t }) {
  return (
    <GlassCard title="🖥️ 系统运行时信息" className="lg:col-span-2">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
        <RuntimeMetric label="Goroutines" value={formatValue(runtime.goroutines, '0', '')} color="blue" sub="并发协程数" />
        <RuntimeMetric label="内存分配" value={formatValue(runtime.memory_alloc_mb, '0', ' MB')} color="green" sub="当前使用" />
        <RuntimeMetric label="系统内存" value={formatValue(runtime.memory_sys_mb, '0', ' MB')} color="purple" sub="总分配" />
        <RuntimeMetric label="GC 次数" value={formatValue(runtime.gc_collections, '0', '')} color="orange" sub="垃圾回收" />
      </div>
      <div className="mt-4 p-4 bg-gray-50 dark:bg-[#111] rounded-xl">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-gray-700 dark:text-gray-300">服务运行时间</span>
          <span className="text-lg font-mono text-primary-600 dark:text-primary-400">
            {runtime.uptime_seconds ? formatUptime(runtime.uptime_seconds) : '-'}
          </span>
        </div>
      </div>
    </GlassCard>
  );
}

function RuntimeMetric({ label, value, color, sub }) {
  const colorMap = {
    blue: { bg: 'bg-blue-50 dark:bg-blue-900/20', text: 'text-blue-600 dark:text-blue-400' },
    green: { bg: 'bg-green-50 dark:bg-green-900/20', text: 'text-green-600 dark:text-green-400' },
    purple: { bg: 'bg-purple-50 dark:bg-purple-900/20', text: 'text-purple-600 dark:text-purple-400' },
    orange: { bg: 'bg-orange-50 dark:bg-orange-900/20', text: 'text-orange-600 dark:text-orange-400' },
  };
  return (
    <div className={`text-center p-4 ${colorMap[color].bg} rounded-xl`}>
      <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">{label}</p>
      <p className={`text-3xl font-bold ${colorMap[color].text}`}>{value}</p>
      <p className="text-xs text-gray-500 mt-1">{sub}</p>
    </div>
  );
}

function ResourceBar({ metric }) {
  return (
    <div className="space-y-2">
      <div className="flex justify-between">
        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{metric.label}</span>
        <span className="text-sm text-gray-500">{metric.value}%</span>
      </div>
      <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
        <div
          className={`h-full ${metric.color} rounded-full transition-all duration-500`}
          style={{ width: `${Math.min(metric.value, 100)}%` }}
        />
      </div>
    </div>
  );
}

function ServiceHealthCard({ services, t }) {
  return (
    <GlassCard>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">服务健康状态</h3>
      <div className="space-y-4">
        {Array.isArray(services) && services.length > 0 ? (
          services.map((service, index) => {
            const name = service.name || service.metadata?.name || `service-${index}`;
            const isOperational = HEALTHY_STATUSES.has(service.status);
            return (
              <div key={index} className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className={`w-2 h-2 rounded-full ${isOperational ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`} />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{name}</span>
                </div>
                <span className="text-sm text-gray-500">{service.latency || service.uptime || service.response_time || '-'}</span>
              </div>
            );
          })
        ) : (
          <p className="text-gray-500">暂无数据</p>
        )}
      </div>
    </GlassCard>
  );
}

function ServiceStatusList({ services, t }) {
  if (!Array.isArray(services) || services.length === 0) {
    return <p className="text-gray-500">暂无数据</p>;
  }
  return (
    <div className="space-y-4">
      {services.map((service, index) => {
        const name = service.name || service.metadata?.name || `service-${index}`;
        const status = service.status || 'Unknown';
        const uptime = service.uptime || service.ready || '-';
        const isOperational = HEALTHY_STATUSES.has(status);
        return (
          <div key={index} className="flex items-center justify-between p-3 rounded-xl bg-gray-50 dark:bg-[#111]">
            <div className="flex items-center gap-3">
              <span className={`w-3 h-3 rounded-full ${isOperational ? 'bg-green-500' : 'bg-yellow-500'}`} />
              <span className="font-medium text-gray-900 dark:text-white">{name}</span>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-500">{uptime}</span>
              <span
                className={`px-2 py-1 text-xs font-medium rounded-full ${
                  isOperational
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                    : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                }`}
              >
                {status}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function buildResourceMetrics(resources, t) {
  if (!resources) return null;
  if (Array.isArray(resources.services) && resources.services.length > 0) {
    const services = resources.services;
    const metrics = [];
    const totalCpu = services.reduce((sum, s) => sum + (s.cpu_usage || 0), 0);
    const totalMem = services.reduce((sum, s) => sum + (s.memory_usage || 0), 0);

    if (totalCpu > 0) {
      const avgCpu = Math.round(totalCpu / services.length);
      metrics.push({
        label: 'CPU',
        value: avgCpu,
        color: getResourceColor(avgCpu, 'primary'),
      });
    }
    if (totalMem > 0) {
      const avgMem = Math.round(totalMem / services.length);
      metrics.push({
        label: '内存',
        value: avgMem,
        color: getResourceColor(avgMem, 'purple'),
      });
    }
    if (resources.disk_usage && resources.disk_usage > 0) {
      metrics.push({
        label: '磁盘',
        value: resources.disk_usage,
        color: getResourceColor(resources.disk_usage, 'blue'),
      });
    }
    if (resources.network_usage && resources.network_usage > 0) {
      metrics.push({ label: '网络', value: resources.network_usage, color: 'bg-green-500' });
    }
    return metrics.length > 0 ? metrics : null;
  }
  return null;
}

function getResourceColor(value, base) {
  if (value > RESOURCE_THRESHOLDS.CRITICAL) return 'bg-red-500';
  if (value > RESOURCE_THRESHOLDS.WARNING) return 'bg-yellow-500';
  return `bg-${base}-500`;
}

function formatValue(value, fallback = '-', suffix = '') {
  if (value === null || value === undefined) return fallback;
  if (typeof value === 'number') return Number(value).toLocaleString() + suffix;
  return String(value) + suffix;
}

function formatUptime(seconds) {
  if (!seconds) return '-';
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = Math.floor(seconds % 60);
  if (days > 0) return `${days}天 ${hours}小时 ${minutes}分钟`;
  if (hours > 0) return `${hours}小时 ${minutes}分钟 ${secs}秒`;
  if (minutes > 0) return `${minutes}分钟 ${secs}秒`;
  return `${secs}秒`;
}

function generateReport({ metrics, services, resources, t }) {
  let report = '# 运维监控报告\n\n';
  report += `生成时间: ${new Date().toLocaleString()}\n\n`;

  report += `## 系统状态\n`;
  if (metrics && Object.keys(metrics).length > 0) {
    if (metrics.total_requests) report += `- 总请求数: ${Math.round(metrics.total_requests)}\n`;
    if (metrics.avg_response_time) report += `- 平均响应时间: ${Number(metrics.avg_response_time).toFixed(2)}ms\n`;
  } else {
    report += `- 暂无数据\n`;
  }

  report += `\n## 系统资源\n`;
  if (resources) {
    if (resources.cpu_usage !== undefined) report += `- CPU: ${resources.cpu_usage}%\n`;
    if (resources.memory_usage !== undefined) report += `- 内存: ${resources.memory_usage}%\n`;
    if (resources.disk_usage !== undefined) report += `- 磁盘: ${resources.disk_usage}%\n`;
    if (resources.network_usage !== undefined) report += `- 网络: ${resources.network_usage}%\n`;
    if (Object.keys(resources).length === 0) report += `- 暂无数据\n`;
  } else {
    report += `- 暂无数据\n`;
  }

  report += `\n## 服务状态\n`;
  if (services.length > 0) {
    services.forEach((service) => {
      const name = service.name || service.metadata?.name;
      const status = service.status || 'Unknown';
      report += `- ${name}: ${status}\n`;
    });
  } else {
    report += `- 暂无数据\n`;
  }

  return report;
}
