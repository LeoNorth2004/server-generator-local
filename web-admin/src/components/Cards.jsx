/**
 * 通用卡片组件
 * GlassCard：玻璃态容器
 * StatCard：统计数字卡片
 * ChartCard：图表容器（带标题与可选操作）
 */
import { useMemo } from 'react';

const BASE_GLASS_CLASS = 'glass-card p-6';
const HOVER_CLASS = 'glass-card-hover';
const DEFAULT_GLASS_CLASS = 'glass-card';

const STAT_LABEL_CLASS = 'stat-label';
const STAT_VALUE_CLASS = 'stat-value break-all overflow-wrap-anywhere';

const isNumeric = (value) => {
  if (typeof value === 'number') return true;
  if (typeof value === 'string' && value !== '' && !Number.isNaN(Number(value))) {
    return true;
  }
  return false;
};

export function GlassCard({ children, className = '', hover = false, ...rest }) {
  const composedClass = `${hover ? HOVER_CLASS : DEFAULT_GLASS_CLASS} ${className}`.trim();
  return (
    <div className={composedClass} {...rest}>
      {children}
    </div>
  );
}

export function StatCard({ title, value, icon, trend, subtitle }) {
  const showTrend = useMemo(() => isNumeric(trend), [trend]);
  const trendDirection = useMemo(() => (Number(trend) > 0 ? '+' : ''), [trend]);
  const trendColor = useMemo(
    () => (Number(trend) > 0 ? 'text-green-500' : 'text-red-500'),
    [trend]
  );

  return (
    <GlassCard hover>
      <div className="flex items-start justify-between">
        <div className="flex flex-col gap-1 min-w-0 flex-1">
          <span className={STAT_LABEL_CLASS}>{title}</span>
          <span className={STAT_VALUE_CLASS}>{value}</span>
          {showTrend && (
            <span className={`text-sm ${trendColor}`}>
              {trendDirection}
              {trend}%
            </span>
          )}
          {subtitle && !showTrend && (
            <span className="text-sm text-gray-500 dark:text-gray-400">{subtitle}</span>
          )}
        </div>
        {icon && (
          <div className="p-3 rounded-xl bg-primary-50 dark:bg-primary-950/50 text-primary-600 dark:text-primary-400 flex-shrink-0 ml-2">
            {icon}
          </div>
        )}
      </div>
    </GlassCard>
  );
}

export function ChartCard({ title, children, action }) {
  return (
    <GlassCard className="h-full">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>
        {action}
      </div>
      {children}
    </GlassCard>
  );
}

export default { GlassCard, StatCard, ChartCard };
