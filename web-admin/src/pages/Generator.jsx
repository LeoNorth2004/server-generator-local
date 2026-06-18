/**
 * 代码生成器页面
 * 1. 表单驱动：tables / fields
 * 2. 拖拽排序字段
 * 3. 区分创建/编辑模式（?projectId）
 * 4. 模块拆分：SortableField / DbSchemaSection / GenerationResultSection
 */
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { generatorAPI, projectAPI, operationsAPI } from '../api';
import { GlassCard, StatCard } from '../components/Cards';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

const FIELD_TYPES = ['int', 'bigint', 'varchar', 'text', 'decimal', 'timestamp', 'boolean'];
const DEFAULT_FIELD_TYPE = 'varchar';
const SENSOR_ACTIVATION_DISTANCE = 5;
const DB_CONFIG_DEFAULTS = {
  host: 'localhost',
  port: 5432,
  user: 'postgres',
  password: '123456',
};
const ZIP_MIME = 'application/zip';
const GENERATION_DELAY = 800;

const buildEmptyConfig = () => ({
  project_name: '',
  tables: [{ name: '', fields: [{ name: '', type: DEFAULT_FIELD_TYPE, comment: '' }] }],
});

const buildEmptyField = () => ({ name: '', type: DEFAULT_FIELD_TYPE, comment: '' });

/* ========== 可拖拽字段 ========== */
function SortableField({ field, fieldIndex, tableIndex, config, setConfig, onRemove }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `${tableIndex}-${fieldIndex}`,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    zIndex: isDragging ? 100 : 'auto',
  };

  const isPrimaryKey = fieldIndex === 0;
  const tableFields = config.tables[tableIndex].fields;

  const updateField = useCallback(
    (key, value) => {
      setConfig((prev) => {
        const nextTables = [...prev.tables];
        nextTables[tableIndex] = {
          ...nextTables[tableIndex],
          fields: [...nextTables[tableIndex].fields],
        };
        nextTables[tableIndex].fields[fieldIndex] = {
          ...nextTables[tableIndex].fields[fieldIndex],
          [key]: value,
        };
        return { ...prev, tables: nextTables };
      });
    },
    [fieldIndex, tableIndex, setConfig]
  );

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`flex gap-3 items-center p-3 rounded-lg border hover:shadow-md transition-all ${
        isPrimaryKey
          ? 'bg-yellow-50 dark:bg-yellow-900/30 border-yellow-400 dark:border-yellow-500'
          : 'bg-white dark:bg-[#111] border-gray-200 dark:border-gray-700'
      }`}
    >
      <DragHandle {...attributes} {...listeners} />
      <FieldIndexBadge isPrimaryKey={isPrimaryKey} fieldIndex={fieldIndex} />
      <input
        type="text"
        value={field.name}
        onChange={(e) => updateField('name', e.target.value)}
        className="input-field flex-1 min-w-[150px] h-9 px-3 text-sm"
        placeholder="字段名称 (如: username)"
      />
      <select
        value={field.type}
        onChange={(e) => updateField('type', e.target.value)}
        className="input-field w-36 h-9 px-2 text-sm cursor-pointer"
      >
        {FIELD_TYPES.map((type) => (
          <option key={type} value={type}>{type}</option>
        ))}
      </select>
      <input
        type="text"
        value={field.comment}
        onChange={(e) => updateField('comment', e.target.value)}
        className="input-field flex-1 min-w-[150px] h-9 px-3 text-sm"
        placeholder="注释说明"
      />
      {tableFields.length > 1 && (
        <button
          onClick={() => onRemove(tableIndex, fieldIndex)}
          className="p-1.5 hover:bg-red-100 dark:hover:bg-red-900/30 rounded transition-colors"
          title="删除字段"
        >
          <svg className="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
}

function DragHandle(props) {
  return (
    <div
      {...props}
      className="cursor-grab active:cursor-grabbing p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors touch-none"
      title="拖拽调整顺序"
    >
      <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 8h16M4 16h16" />
      </svg>
    </div>
  );
}

function FieldIndexBadge({ isPrimaryKey, fieldIndex }) {
  return (
    <div
      className={`flex items-center justify-center min-w-[60px] px-2 py-1 rounded text-xs font-medium ${
        isPrimaryKey
          ? 'bg-yellow-200 text-yellow-800 dark:bg-yellow-700 dark:text-yellow-100'
          : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
      }`}
    >
      {isPrimaryKey ? '🔑 主键' : `#${fieldIndex + 1}`}
    </div>
  );
}

/* ========== 数据库表配置区 ========== */
function DbSchemaSection({ config, setConfig }) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: SENSOR_ACTIVATION_DISTANCE } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  );

  const addTable = useCallback(() => {
    setConfig((prev) => ({
      ...prev,
      tables: [...prev.tables, { name: '', fields: [buildEmptyField()] }],
    }));
  }, [setConfig]);

  const addField = useCallback(
    (tableIndex) => {
      setConfig((prev) => {
        const nextTables = prev.tables.map((table, idx) =>
          idx === tableIndex ? { ...table, fields: [...table.fields, buildEmptyField()] } : table
        );
        return { ...prev, tables: nextTables };
      });
    },
    [setConfig]
  );

  const removeTable = useCallback(
    (tableIndex) => {
      setConfig((prev) => {
        if (prev.tables.length <= 1) return prev;
        return { ...prev, tables: prev.tables.filter((_, idx) => idx !== tableIndex) };
      });
    },
    [setConfig]
  );

  const removeField = useCallback(
    (tableIndex, fieldIndex) => {
      setConfig((prev) => {
        const nextTables = prev.tables.map((table, idx) => {
          if (idx !== tableIndex) return table;
          if (table.fields.length <= 1) return table;
          return { ...table, fields: table.fields.filter((_, i) => i !== fieldIndex) };
        });
        return { ...prev, tables: nextTables };
      });
    },
    [setConfig]
  );

  const updateTableName = useCallback(
    (tableIndex, name) => {
      setConfig((prev) => {
        const nextTables = prev.tables.map((table, idx) =>
          idx === tableIndex ? { ...table, name } : table
        );
        return { ...prev, tables: nextTables };
      });
    },
    [setConfig]
  );

  const handleDragEnd = useCallback(
    (event, tableIndex) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;
      const oldIndex = parseInt(active.id.split('-')[1], 10);
      const newIndex = parseInt(over.id.split('-')[1], 10);
      if (oldIndex === newIndex) return;
      setConfig((prev) => {
        const nextTables = prev.tables.map((table, idx) => {
          if (idx !== tableIndex) return table;
          return { ...table, fields: arrayMove(table.fields, oldIndex, newIndex) };
        });
        return { ...prev, tables: nextTables };
      });
    },
    [setConfig]
  );

  return (
    <GlassCard>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">数据库表配置</h3>
        <button onClick={addTable} className="btn-secondary text-sm px-4 py-2">添加表</button>
      </div>
      <div className="space-y-6">
        {config.tables.map((table, tableIndex) => (
          <TableEditor
            key={tableIndex}
            table={table}
            tableIndex={tableIndex}
            config={config}
            setConfig={setConfig}
            onAddField={addField}
            onRemoveTable={removeTable}
            onRemoveField={removeField}
            onUpdateTableName={updateTableName}
            onDragEnd={handleDragEnd}
            sensors={sensors}
            canRemoveTable={config.tables.length > 1}
          />
        ))}
        {config.tables.length === 0 && (
          <div className="text-center py-12 text-gray-400">
            <div className="text-5xl mb-4">📊</div>
            <p>暂无数据表，请点击上方按钮添加</p>
          </div>
        )}
      </div>
    </GlassCard>
  );
}

function TableEditor({
  table,
  tableIndex,
  config,
  setConfig,
  onAddField,
  onRemoveTable,
  onRemoveField,
  onUpdateTableName,
  onDragEnd,
  sensors,
  canRemoveTable,
}) {
  const itemIds = useMemo(() => table.fields.map((_, idx) => `${tableIndex}-${idx}`), [table.fields, tableIndex]);

  return (
    <div className="p-6 bg-gray-50 dark:bg-[#111] rounded-xl border border-gray-200 dark:border-gray-700">
      <div className="flex gap-4 mb-4 items-center">
        <label className="text-sm font-medium text-gray-700 dark:text-gray-300 min-w-[80px]">表名:</label>
        <input
          type="text"
          value={table.name}
          onChange={(e) => onUpdateTableName(tableIndex, e.target.value)}
          className="input-field flex-1 min-w-[200px] h-10 px-3 text-base"
          placeholder="例如: users"
        />
        <button onClick={() => onAddField(tableIndex)} className="btn-secondary text-sm px-4 py-2 whitespace-nowrap">
          + 添加字段
        </button>
        {canRemoveTable && (
          <button onClick={() => onRemoveTable(tableIndex)} className="p-2 hover:bg-red-100 dark:hover:bg-red-900/30 rounded-lg transition-colors" title="删除表">
            <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        )}
      </div>
      <div className="space-y-3 ml-[96px]">
        <div className="flex items-center justify-between mb-2">
          <div className="text-sm font-medium text-gray-600 dark:text-gray-400">字段列表 (拖拽调整顺序):</div>
          <div className="text-xs text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/30 px-2 py-1 rounded">
            💡 第一个字段自动设为主键
          </div>
        </div>
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={(event) => onDragEnd(event, tableIndex)}>
          <SortableContext items={itemIds} strategy={verticalListSortingStrategy}>
            {table.fields.map((field, fieldIndex) => (
              <SortableField
                key={`${tableIndex}-${fieldIndex}`}
                field={field}
                fieldIndex={fieldIndex}
                tableIndex={tableIndex}
                config={config}
                setConfig={setConfig}
                onRemove={onRemoveField}
              />
            ))}
          </SortableContext>
        </DndContext>
        {table.fields.length === 0 && (
          <div className="text-gray-400 text-sm italic p-3 bg-gray-100 dark:bg-[#111] rounded-lg">
            暂无字段，请点击"添加字段"按钮
          </div>
        )}
      </div>
    </div>
  );
}

/* ========== 生成结果区 ========== */
function GenerationResultSection({ result, onDownload }) {
  if (!result) return null;
  const files = result?.code?.files ? Object.keys(result.code.files) : [];
  return (
    <GlassCard>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">生成结果</h3>
      <div className="space-y-4">
        <div className="flex justify-between items-center flex-wrap gap-3">
          <span className="text-sm text-gray-600 dark:text-gray-400">
            生成了 <strong>{files.length}</strong> 个文件
          </span>
          {result?.project_id && (
            <button className="btn-primary text-sm flex items-center gap-2" onClick={onDownload}>
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              下载
            </button>
          )}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-h-[300px] overflow-y-auto">
          {files.length > 0 ? (
            files.map((file) => <GeneratedFileRow key={file} name={file} />)
          ) : (
            <p className="text-gray-500">没有生成的文件</p>
          )}
        </div>
      </div>
    </GlassCard>
  );
}

function GeneratedFileRow({ name }) {
  return (
    <div className="p-3 bg-gray-50 dark:bg-[#111] rounded-lg flex items-center gap-2">
      <svg className="w-4 h-4 text-primary-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <span className="text-sm font-medium text-gray-900 dark:text-white truncate">{name}</span>
    </div>
  );
}

/* ========== 主页面 ========== */
export default function Generator() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const editProjectId = searchParams.get('projectId');
  const { t } = useTranslation();

  const [config, setConfig] = useState(buildEmptyConfig);
  const [generating, setGenerating] = useState(false);
  const [result, setResult] = useState(null);
  const [loadingProject, setLoadingProject] = useState(Boolean(editProjectId));

  useEffect(() => {
    if (!editProjectId) return;
    const load = async () => {
      setLoadingProject(true);
      try {
        const response = await projectAPI.get(editProjectId);
        const project = response.data?.data ?? response.data;
        if (!project?.table_config) return;
        try {
          const tables = JSON.parse(project.table_config);
          if (Array.isArray(tables) && tables.length > 0) {
            setConfig({ project_name: project.name || '', tables });
          }
        } catch (e) {
          console.warn('Failed to parse table_config:', e);
        }
      } catch (error) {
        console.error('Failed to load project:', error);
        alert('加载项目失败');
      } finally {
        setLoadingProject(false);
      }
    };
    load();
  }, [editProjectId]);

  const buildDbConfig = useCallback(() => ({
    ...DB_CONFIG_DEFAULTS,
    db_name: config.project_name.toLowerCase().replace(/\s+/g, '_') + '_db',
  }), [config.project_name]);

  const recordGenerateLog = useCallback(
    async (status, projectId, details, errorMsg) => {
      await operationsAPI
        .recordOperationLog({
          action: status === 'success' ? (editProjectId ? 'regenerate' : 'generate') : 'generate',
          resource: 'project',
          resource_id: projectId || 0,
          details: JSON.stringify(details),
          status,
          duration: details.duration || 0,
          ...(errorMsg ? { error: errorMsg } : {}),
        })
        .catch(() => {});
    },
    [editProjectId]
  );

  const handleGenerate = useCallback(async () => {
    if (!config.project_name) return;
    setGenerating(true);
    setResult(null);
    const startTime = Date.now();
    try {
      const requestData = { ...config, db_config: config.db_config || buildDbConfig() };
      const response = await generatorAPI.generate(requestData);
      const data = response.data?.data ?? response.data;
      setResult(data);
      const duration = Date.now() - startTime;
      if (data?.project_id) {
        await recordGenerateLog('success', data.project_id, {
          project_name: config.project_name,
          tables_count: config.tables.length,
          files_generated: data.code?.files ? Object.keys(data.code.files).length : 0,
          duration,
        });
      }
      alert('代码生成成功！');
    } catch (error) {
      const msg = error.response?.data?.message || error.message || 'Unknown error';
      await recordGenerateLog('failed', 0, { project_name: config.project_name, error: msg, duration: Date.now() - startTime }, msg);
      alert(`代码生成失败: ${msg}`);
    } finally {
      setGenerating(false);
    }
  }, [config, buildDbConfig, recordGenerateLog]);

  const handleDownload = useCallback(async () => {
    if (!result?.project_id) return;
    try {
      const response = await generatorAPI.download(result.project_id);
      if (response.data instanceof Blob && response.data.type === ZIP_MIME) {
        const url = URL.createObjectURL(response.data);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${config.project_name}-generated-code.zip`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        await operationsAPI.recordOperationLog({
          action: 'download',
          resource: 'project',
          resource_id: result.project_id,
          details: JSON.stringify({ project_name: config.project_name }),
          status: 'success',
          duration: 0,
        }).catch(() => {});
        alert('下载成功！');
      } else {
        const errorText = await response.data.text();
        alert(`下载失败: ${errorText}`);
      }
    } catch (error) {
      console.error('Download failed:', error);
      alert('下载失败');
    }
  }, [result, config.project_name]);

  const handleRegenerate = useCallback(async () => {
    if (!editProjectId) return;
    if (!window.confirm(`重新生成 "${config.project_name}"?`)) return;
    setGenerating(true);
    setResult(null);
    const startTime = Date.now();
    try {
      const response = await generatorAPI.generateFromProject(editProjectId);
      const data = response.data?.data ?? response.data;
      setResult(data);
      await operationsAPI.recordOperationLog({
        action: 'regenerate',
        resource: 'project',
        resource_id: parseInt(editProjectId, 10),
        details: JSON.stringify({ project_name: config.project_name }),
        status: 'success',
        duration: Date.now() - startTime,
      }).catch(() => {});
      alert('重新生成成功');
    } catch (error) {
      const msg = error.response?.data?.message || error.message;
      await operationsAPI.recordOperationLog({
        action: 'regenerate',
        resource: 'project',
        resource_id: parseInt(editProjectId, 10),
        details: JSON.stringify({ project_name: config.project_name, error: msg }),
        status: 'failed',
        duration: Date.now() - startTime,
        error: msg,
      }).catch(() => {});
      alert(`重新生成失败: ${msg}`);
    } finally {
      setGenerating(false);
    }
  }, [editProjectId, config.project_name]);

  if (loadingProject) {
    return <PageLoader t={t} />;
  }

  return (
    <div className="space-y-6">
      <PageHeader editProjectId={editProjectId} projectName={config.project_name} onBack={() => navigate('/projects')} t={t} />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <ProjectInfoCard config={config} setConfig={setConfig} t={t} />
          <DbSchemaSection config={config} setConfig={setConfig} />
          <GeneratorActions
            generating={generating}
            hasProjectName={Boolean(config.project_name)}
            onGenerate={handleGenerate}
            onRegenerate={handleRegenerate}
            isEditing={Boolean(editProjectId)}
            t={t}
          />
        </div>

        <div className="space-y-6">
          <StructurePreviewCard />
          <StatCard title="预估文件" value="25+" />
          <StatCard title="代码行数" value="~2,500" />
        </div>

        <GenerationResultSection result={result} onDownload={handleDownload} />
      </div>
    </div>
  );
}

function PageHeader({ editProjectId, projectName, onBack, t }) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">代码生成器</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          {editProjectId ? `编辑项目: ${projectName}` : '配置项目并生成代码'}
        </p>
      </div>
      {editProjectId && (
        <button onClick={onBack} className="btn-secondary flex items-center gap-2">
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          返回项目列表
        </button>
      )}
    </div>
  );
}

function ProjectInfoCard({ config, setConfig }) {
  return (
    <GlassCard>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">项目配置</h3>
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">项目名称</label>
          <input
            type="text"
            value={config.project_name}
            onChange={(e) => setConfig({ ...config, project_name: e.target.value })}
            className="input-field"
            placeholder="my-service"
          />
        </div>
      </div>
    </GlassCard>
  );
}

function StructurePreviewCard() {
  const items = [
    'go.mod',
    'main.go',
    'config/',
    'database/',
    'internal/models/',
    'internal/dao/',
    'internal/service/',
    'internal/controller/',
    'internal/router/',
    'migrations/',
  ];
  return (
    <GlassCard>
      <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">生成的项目结构</h3>
      <div className="space-y-3 text-sm">
        {items.map((item) => (
          <div key={item} className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
            <svg className="w-4 h-4 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            {item}
          </div>
        ))}
      </div>
    </GlassCard>
  );
}

function GeneratorActions({ generating, hasProjectName, onGenerate, onRegenerate, isEditing, t }) {
  return (
    <div className="flex gap-3">
      <button onClick={onGenerate} disabled={generating || !hasProjectName} className="flex-1 btn-primary py-3 disabled:opacity-50">
        {generating ? <Spinner label="生成中..." /> : '生成代码'}
      </button>
      {isEditing && (
        <button onClick={onRegenerate} disabled={generating} className="flex-1 btn-secondary py-3 disabled:opacity-50">
          {generating ? <Spinner label="重新生成中..." /> : '重新生成'}
        </button>
      )}
    </div>
  );
}

function Spinner({ label }) {
  return (
    <span className="flex items-center justify-center gap-2">
      <svg className="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
      {label}
    </span>
  );
}

function PageLoader() {
  return (
    <div className="flex items-center justify-center py-20">
      <svg className="w-8 h-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
      <span className="ml-3 text-gray-600 dark:text-gray-400">加载中...</span>
    </div>
  );
}
