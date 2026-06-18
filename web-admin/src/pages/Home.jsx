/**
 * 首页 - 仪表盘
 * 1. 并发拉取统计、健康与服务数据
 * 2. 通过数据适配器统一处理不同响应格式
 * 3. 渲染统计卡片与最近项目
 */
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { StatCard, ChartCard, GlassCard } from '../components/Cards';
import { projectAPI, userAPI } from '../api';
import { apiClient } from '../api';

const RECENT_PROJECTS_LIMIT = 5;
const SERVICES_DISPLAY_LIMIT = 6;
const RECENT_LIST_MAX_HEIGHT = 'max-h-[320px]';
const STAT_GRID_COLS = 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6';
const QUICK_ACTIONS_COLS = 'grid-cols-1 lg:grid-cols-3 gap-6';

const ICON_PROPS = {
  className: 'w-6 h-6',
  fill: 'none',
  viewBox: '0 0 24 24',
  stroke: 'currentColor',
};

const StatIcon = ({ d }) => (
  <svg {...ICON_PROPS}>
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={d} />
  </svg>
);

/** 把 axios/fetch 响应统一规范化为数组或空数组 */
const normalizeList = (rawData) => {
  if (Array.isArray(rawData)) return rawData;
  if (rawData && Array.isArray(rawData.data)) return rawData.data;
  return [];
};

/** 从 axios 响应中提取 data 字段 */
const extractData = (response) => response?.data?.data ?? response?.data;

/** 从 fetch Response 中提取 data 字段 */
const extractFetchData = async (response) => {
  if (!response?.ok) return null;
  try {
    return await response.json();
  } catch {
    return null;
  }
};

export default function Home() {
  const [stats, setStats] = useState({ user_count: 0, project_count: 0, generation_count: 0 });
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(false);
  const [recentProjects, setRecentProjects] = useState([]);
  const [services, setServices] = useState([]);
  const { t } = useTranslation();
  const navigate = useNavigate();

  useEffect(() => {
    const loadDashboard = async () => {
      setLoading(true);
      try {
        const [usersRes, projectsRes, healthRes, servicesRes] = await Promise.allSettled([
          userAPI.list(),
          projectAPI.list(),
          fetch('/api/v1/operations/health'),
          fetch('/api/v1/operations/services').catch(() => null),
        ]);

        if (usersRes.status === 'fulfilled' && usersRes.value.status === 200) {
          const usersData = normalizeList(extractData(usersRes.value));
          setStats((prev) => ({ ...prev, user_count: usersData.length }));
        }

        if (projectsRes.status === 'fulfilled' && projectsRes.value.status === 200) {
          const projectsData = normalizeList(extractData(projectsRes.value));
          setStats((prev) => ({ ...prev, project_count: projectsData.length }));
          setRecentProjects(projectsData.slice(0, RECENT_PROJECTS_LIMIT));
        }

        if (healthRes.status === 'fulfilled' && healthRes.value.ok) {
          const healthData = await extractFetchData(healthRes.value);
          setHealth(healthData);
        }

        if (servicesRes.status === 'fulfilled' && servicesRes.value?.ok) {
          const servicesData = await extractFetchData(servicesRes.value);
          const servicesList = normalizeList(servicesData);
          setServices(servicesList.slice(0, SERVICES_DISPLAY_LIMIT));
        }
      } catch (error) {
        console.log('Some data not available:', error.message);
      } finally {
        setLoading(false);
      }
    };
    loadDashboard();
  }, []);

  const handleNewProject = useCallback(() => navigate('/generator'), [navigate]);
  const handleGenerateCode = useCallback(() => navigate('/generator'), [navigate]);

  const statCards = useMemo(
    () => [
      {
        title: t('dashboard.totalUsers'),
        value: stats.user_count,
        icon: <StatIcon d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />,
      },
      {
        title: t('dashboard.activeProjects'),
        value: stats.project_count,
        icon: <StatIcon d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />,
      },
      {
        title: t('dashboard.codeGenerations'),
        value: stats.generation_count,
        icon: <StatIcon d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />,
      },
    ],
    [t, stats]
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {t('dashboard.title')}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">
            {t('dashboard.welcome')} - Generator Platform
          </p>
        </div>
        <div className="flex gap-3">
          <button onClick={handleNewProject} className="btn-primary">
            {t('dashboard.newProject')}
          </button>
        </div>
      </div>

      <div className={`grid ${STAT_GRID_COLS}`}>
        {statCards.map((stat, index) => (
          <StatCard key={index} {...stat} />
        ))}
      </div>

      <div className={`grid ${QUICK_ACTIONS_COLS}`}>
        <GlassCard>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            {t('dashboard.quickActions')}
          </h3>
          <div className="space-y-3">
            <button
              onClick={handleNewProject}
              className="w-full btn-primary flex items-center justify-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              {t('dashboard.newProject')}
            </button>
            <button
              onClick={handleGenerateCode}
              className="w-full btn-secondary flex items-center justify-center gap-2"
            >
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
              </svg>
              {t('dashboard.generateCode')}
            </button>
          </div>
        </GlassCard>

        <ChartCard title={t('dashboard.recentProjects')} className="lg:col-span-2">
          <div className={`space-y-4 ${RECENT_LIST_MAX_HEIGHT} overflow-y-auto`}>
            {recentProjects.length > 0 ? (
              recentProjects.map((project, index) => (
                <ProjectListItem
                  key={index}
                  project={project}
                  generatedLabel={t('dashboard.generated')}
                  notGeneratedLabel={t('dashboard.notGenerated')}
                />
              ))
            ) : (
              <EmptyProjectState
                noDataText={t('common.noData')}
                hintText={t('dashboard.noProjectsYet')}
              />
            )}
          </div>
        </ChartCard>
      </div>
    </div>
  );
}

function formatProjectDate(project) {
  if (project.created_at) return new Date(project.created_at).toLocaleDateString();
  if (project.updated_at) return new Date(project.updated_at).toLocaleDateString();
  return '-';
}

function ProjectListItem({ project, generatedLabel, notGeneratedLabel }) {
  const isGenerated = Boolean(project.generated_code);
  return (
    <div className="flex items-center justify-between p-3 rounded-xl bg-gray-50 dark:bg-[#111] hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center">
          <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </div>
        <div>
          <p className="font-medium text-gray-900 dark:text-white">{project.name}</p>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {formatProjectDate(project)}
          </p>
        </div>
      </div>
      <span
        className={`px-3 py-1 text-xs font-medium rounded-full ${
          isGenerated
            ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
            : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
        }`}
      >
        {isGenerated ? generatedLabel : notGeneratedLabel}
      </span>
    </div>
  );
}

function EmptyProjectState({ noDataText, hintText }) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-gray-400">
      <svg className="w-12 h-12 mb-3 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
      </svg>
      <p className="font-medium">{noDataText}</p>
      <p className="text-sm mt-1">{hintText}</p>
    </div>
  );
}
