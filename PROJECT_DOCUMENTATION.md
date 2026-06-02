# 🚀 Generator Platform - 低代码代码生成平台

## 📖 项目简介

**Generator Platform** 是一个基于 **Go + React** 的全栈低代码平台，旨在通过可视化配置和元数据驱动的方式，自动生成完整的后端应用程序代码。该平台采用前后端分离架构，支持多用户管理、权限控制、实时监控等功能。

### ✨ 核心特性

- 🔥 **一键生成代码**: 根据数据库表结构自动生成 Go/Gin/GORM 后端代码
- 👥 **多用户系统**: 支持管理员和普通用户角色，完善的权限控制
- 📊 **实时运维监控**: 动态统计请求数、用户数、生成次数等关键指标
- 🌐 **国际化支持**: 完整的中英文界面切换
- 🎨 **现代化UI**: 基于 React 19 + Tailwind CSS 4 的响应式设计
- 💾 **数据持久化**: PostgreSQL 数据库存储，Docker 容器化部署
- 📝 **文档自动生成**: 自动生成 API 文档、配置指南、开发手册

---

## 🛠️ 技术栈

### 后端技术

| 技术 | 版本 | 用途 |
|------|------|------|
| **Go (Golang)** | 1.21+ | 主要编程语言 |
| **Gin Web Framework** | v1.10.0 | HTTP 路由框架 |
| **GORM** | v1.25.x | ORM 数据库操作 |
| **PostgreSQL Driver** | v1.5.7 | 数据库连接 |
| **JWT (golang-jwt)** | v5.2.1 | 用户认证令牌 |
| **bcrypt** | - | 密码加密 |
| **Reverse Proxy** | 内置 | API 网关代理 |

### 前端技术

| 技术 | 版本 | 用途 |
|------|------|------|
| **React** | 19.2.4 | UI 框架 |
| **Vite** | 8.0.1 | 构建工具 |
| **Tailwind CSS** | 4.2.2 | CSS 框架 |
| **React Router** | 6.22.0 | 路由管理 |
| **Axios** | 1.6.7 | HTTP 客户端 |
| **i18next** | 26.0.3 | 国际化 |
| **Recharts** | 3.8.1 | 图表可视化 |
| **MUI (Material UI)** | 7.3.9 | 组件库 |

### 基础设施

| 组件 | 说明 |
|------|------|
| **PostgreSQL 15+** | 关系型数据库（Docker 部署） |
| **Docker** | 容器化运行环境 |
| **Nginx**（可选） | 反向代理/生产部署 |

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户浏览器                               │
│                   React SPA (:3000)                            │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTPS/HTTP
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Backend Service (:8080)                      │
│  ┌─────────────┬─────────────┬─────────────┬────────────────┐   │
│  │ Auth Module │ User Module │Project Module│ Operations Mod │   │
│  │  /api/v1/*  │  /api/v1/*  │  /api/v1/*  │   /api/v1/*    │   │
│  └─────────────┴─────────────┴─────────────┴────────────────┘   │
│                         │                                       │
│              ┌──────────▼──────────┐                             │
│         Reverse Proxy (API Gateway)                              │
│         /api/v1/generator/* → :8084                             │
│  └─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                 Generator Service (:8084)                       │
│  ┌──────────────┬──────────────┬──────────────┬───────────────┐  │
│  │ Code Engine  │ Metadata     │ Workflow     │ Form Builder  │  │
│  │ Generator    │ Engine       │ Engine       │ Engine        │  │
│  └──────────────┴──────────────┴──────────────┴───────────────┘  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              PostgreSQL Database (:5432)                        │
│  ┌──────────┬──────────┬──────────┬──────────────────────────┐  │
│  │ users    │ projects │operation_│ generated_code (JSONB)    │  │
│  │ table    │ table    │ logs     │                          │  │
│  └──────────┴──────────┴──────────┴──────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 架构设计原则

1. **微服务思想**: Backend 和 Generator 分离，独立部署
2. **API Gateway 模式**: Backend 作为统一入口，代理 Generator 服务
3. **关注点分离**: 认证、业务逻辑、代码生成各自独立
4. **水平扩展**: 各服务可独立扩展

---

## 📁 项目结构详解

```
generator/
│
├── 📂 backend-service/          # 后端主服务（API 网关）
│   ├── main.go                  # 入口文件 & 路由注册
│   ├── auth.go                  # 认证模块（登录/注册/JWT）
│   ├── user.go                  # 用户管理 CRUD
│   ├── project.go               # 项目管理 CRUD
│   ├── operations.go            # 运维监控 & 统计
│   ├── init_data.go             # 初始化默认数据
│   ├── .env                     # 环境变量配置
│   └── go.mod / go.sum          # Go 依赖管理
│
├── 📂 generator-service/        # 代码生成引擎服务
│   ├── main.go                  # 入口 & 服务初始化
│   ├── handlers.go              # HTTP 请求处理器
│   ├── config.go                # 配置加载
│   ├── types.go                 # 类型定义
│   ├── models.go                # 数据模型
│   ├── core.go                  # 核心代码生成逻辑
│   ├── engine_adapter.go        # 多数据库适配器
│   ├── engine_metadata.go       # 元数据引擎
│   ├── engine_workflow.go       # 工作流引擎
│   ├── engine_form.go           # 表单引擎
│   ├── docs.go                  # 文档生成器
│   ├── middleware.go            # 中间件（认证/CORS）
│   └── router.go                # 路由定义
│
├── 📂 libs/go-common/           # 公共共享库
│   ├── config/config.go         # 统一配置加载
│   ├── database/database.go     # 数据库连接池
│   ├── middleware/middleware.go  # JWT 认证中间件
│   ├── models/models.go         # 公共数据模型
│   └── response/response.go     # 统一响应格式
│
├── 📂 web-admin/                # 前端管理界面
│   ├── src/
│   │   ├── main.jsx             # 应用入口
│   │   ├── App.jsx              # 根组件 & 路由
│   │   ├── api.jsx              # API 封装层
│   │   ├── pages/               # 页面组件
│   │   │   ├── Home.jsx         # 仪表盘首页
│   │   │   ├── Login.jsx        # 登录页
│   │   │   ├── Projects.jsx     # 项目管理
│   │   │   ├── Generator.jsx    # 代码生成器
│   │   │   ├── Users.jsx        # 用户管理
│   │   │   ├── Operations.jsx   # 运维监控
│   │   │   └── Docs.jsx         # 文档中心
│   │   ├── components/          # 可复用组件
│   │   │   ├── Layout.jsx       # 页面布局
│   │   │   ├── Sidebar.jsx      # 侧边栏导航
│   │   │   ├── Navbar.jsx       # 顶部导航栏
│   │   │   └── Cards.jsx        # 卡片组件
│   │   ├── contexts/            # React Context
│   │   │   ├── AuthContext.jsx  # 认证状态
│   │   │   ├── ThemeContext.jsx # 主题状态
│   │   │   └── I18nContext.jsx  # 国际化状态
│   │   └── i18n/locales/        # 翻译文件
│   │       ├── zh.json         # 中文
│   │       └── en.json         # 英文
│   ├── package.json             # NPM 依赖
│   └── vite.config.js           # Vite 配置（含代理）
│
├── start-all.bat                # Windows 一键启动脚本
├── start.sh                     # Linux/Mac 启动脚本
├── stop.bat                     # 停止所有服务
└── README.md                    # 本文档
```

---

## 🗄️ 数据库设计

### ER 关系图

```
┌─────────────────┐       ┌──────────────────┐       ┌─────────────────────┐
│     users       │       │     projects      │       │   operation_logs    │
├─────────────────┤       ├──────────────────┤       ├─────────────────────┤
│ PK id (uint)    │──┐    │ PK id (uint)      │       │ PK id (bigint)      │
│ username (varchar)│ │    │ FK user_id (uint) │◄──┐   │ FK user_id (uint)   │
│ email (varchar)  │  │    │ name (varchar)    │   │   │ action (varchar)    │
│ password_hash    │  └───►│ description(text) │   │   │ resource (varchar)   │
│ role (enum)      │       │ db_config(jsonb) │   │   │ resource_id (uint)  │
│ status (enum)    │       │ table_config(jsonb)│   │   │ details (jsonb)     │
│ created_at       │       │ generated_code   │   │   │ status (varchar)     │
│ updated_at       │       │ status (varchar)  │   └──►│ error_msg (text)     │
└─────────────────┘       │ created_at       │       │ duration_ms (int)    │
                          │ updated_at       │       │ ip_address (varchar) │
                          └──────────────────┘       │ user_agent (text)    │
                                                       │ created_at          │
                                                       └─────────────────────┘
```

### 核心表说明

#### `users` 表 - 用户信息
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,  -- bcrypt 加密
    role VARCHAR(20) DEFAULT 'user',       -- 'admin' | 'user'
    status VARCHAR(20) DEFAULT 'active',   -- 'active' | 'disabled'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### `projects` 表 - 项目信息
```sql
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),  -- 创建者
    name VARCHAR(200) NOT NULL,
    description TEXT,
    db_config JSONB,                      -- 数据库配置
    table_config JSONB,                    -- 表结构定义
    generated_code JSONB,                  -- 生成的代码内容
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending' | 'generated'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### `operation_logs` 表 - 操作日志
```sql
CREATE TABLE operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50),                    -- 'login'|'generate'|'download'...
    resource VARCHAR(50),                  -- 'project'|'code'|'auth'...
    resource_id INTEGER,
    details JSONB,                         -- 操作详情
    status VARCHAR(20),                    -- 'success'|'failed'
    error_msg TEXT,
    duration_ms INTEGER,                   -- 耗时（毫秒）
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 🔌 API 接口文档

### 认证模块 (`/api/v1/auth`)

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| POST | `/auth/login` | 用户登录 | 公开 |
| POST | `/auth/register` | 用户注册 | 公开 |
| GET | `/auth/me` | 获取当前用户信息 | 需登录 |

**请求示例**:
```bash
# 登录
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}

# 响应
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": { "id": 1, "username": "admin", "role": "admin" }
  }
}
```

### 项目管理 (`/api/v1/projects`)

| 方法 | 路径 | 描述 | 权限 |
|------|------|------|------|
| GET | `/projects` | 获取项目列表 | 需登录 |
| GET | `/projects/:id` | 获取项目详情 | 需登录 |
| POST | `/projects` | 创建新项目 | 需登录 |
| PUT | `/projects/:id` | 更新项目 | 需登录 |
| DELETE | `/projects/:id` | 删除项目 | 仅管理员 |

**权限规则**:
- **普通用户**: 只能查看/操作自己创建的项目
- **管理员**: 可以查看所有用户的项目

### 代码生成 (`/api/v1/generator/*`)

> ⚠️ 这些请求会被 Backend Service 代理到 Generator Service (:8084)

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/generator/generate` | 生成代码 |
| POST | `/generator/generate/:project_id` | 重新生成 |
| GET | `/generator/download/:project_id` | 下载 ZIP 包 |
| GET | `/generator/preview/:project_id` | 预览代码文件 |
| POST | `/generator/docs/generate` | 生成文档 |

**代码生成请求示例**:
```bash
POST /api/v1/generator/generate
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_name": "MyECommerce",
  "db_config": {
    "host": "localhost",
    "port": 5432,
    "user": "postgres",
    "password": "123456",
    "db_name": "myapp"
  },
  "tables": [
    {
      "name": "users",
      "fields": [
        { "name": "id", "type": "int", "primary": true },
        { "name": "username", "type": "varchar(100)" },
        { "name": "email", "type": "varchar(200)" },
        { "name": "created_at", "type": "timestamp" }
      ]
    },
    {
      "name": "products",
      "fields": [
        { "name": "id", "type": "int", "primary": true },
        { "name": "name", "type": "varchar(200)" },
        { "name": "price", "type": "decimal(10,2)" }
      ]
    }
  ]
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "project_id": 15,
    "code": {
      "language": "go",
      "files": {
        "main.go": "package main\n\nimport ...",
        "config/config.go": "...",
        "internal/models/user.go": "..."
      }
    }
  }
}
```

### 运维监控 (`/api/v1/operations`)

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/operations/stats` | 获取统计数据 |
| GET | `/operations/metrics` | 获取系统指标 |
| GET | `/operations/health` | 健康检查 |
| GET | `/operation-logs` | 获取操作日志 |

---

## 🚀 快速开始

### 环境要求

- **Go**: 1.21 或更高版本
- **Node.js**: 18+ 或 20+
- **Docker Desktop**: 用于运行 PostgreSQL
- **Git**: 版本控制（可选）

### 一键启动（推荐）

#### Windows
```bash
# 双击运行或命令行执行：
start-all.bat
```

#### Linux/macOS
```bash
chmod +x start.sh
./start.sh
```

启动脚本会自动：
1. ✅ 检查并启动 PostgreSQL Docker 容器
2. ✅ 编译并启动 Backend Service (:8080)
3. ✅ 编译并启动 Generator Service (:8084)
4. ✅ 安装依赖并启动前端开发服务器 (:3000)
5. ✅ 打开浏览器访问 http://localhost:3000

### 手动启动步骤

#### 1️⃣ 启动数据库
```bash
# 启动 PostgreSQL 容器（如果未运行）
docker start local-postgres

# 或创建新容器
docker run -d \
  --name local-postgres \
  -e POSTGRES_PASSWORD=123456 \
  -e POSTGRES_DB=generator_platform \
  -p 5432:5432 \
  postgres:15
```

#### 2️⃣ 启动后端服务
```bash
cd backend-service

# 设置环境变量并编译运行
$env:DB_PASSWORD="123456"
go build -o backend-service.exe .
.\backend-service.exe
```

Backend Service 将在 **http://localhost:8080** 启动

#### 3️⃣ 启动代码生成服务
```bash
cd generator-service

$env:DB_PASSWORD="123456"
go build -o generator-service.exe .
.\generator-service.exe
```

Generator Service 将在 **http://localhost:8084** 启动

#### 4️⃣ 启动前端
```bash
cd web-admin
npm install
npm run dev
```

前端将在 **http://localhost:3000** 启动

### 默认账户

| 角色 | 用户名 | 密码 | 权限 |
|------|--------|------|------|
| 管理员 | `admin` | `admin123` | 全部功能 |
| 普通用户 | 自定义 | 自定义 | 只能查看自己的项目 |

---

## 🎯 核心功能使用指南

### 1️⃣ 代码生成流程

```
登录 → 进入"代码生成"页面 → 配置表结构 → 点击"生成"
→ 查看生成的代码 → 下载 ZIP 包 或 在项目管理中查看
```

**详细步骤**:

1. **登录系统**
   - 打开 http://localhost:3000
   - 使用 admin/admin123 登录

2. **进入代码生成器**
   - 点击侧边栏 **"代码生成"** 或顶部 **"+ 新建项目"**

3. **配置项目信息**
   ```
   - 项目名称：例如 "电商后台管理系统"
   
   - 数据库配置（可选）：
     * Host: localhost
     * Port: 5432
     * User: postgres
     * Password: 123456
     * DB Name: myapp
   ```

4. **添加数据表**
   - 点击 **"+ 添加表"**
   - 输入表名：如 `users`, `orders`, `products`
   - 为每个表添加字段：
     * 字段名：`id`, `username`, `email`
     * 字段类型：`int`, `varchar(100)`, `timestamp`
     * 勾选主键（通常第一个字段）
     * 可选：是否允许为空、注释说明

5. **生成代码**
   - 点击底部 **"🚀 生成代码"** 按钮
   - 等待几秒后显示生成结果
   - 查看：
     * 生成的文件列表
     * 预览单个文件内容
     * 代码行数统计

6. **下载或保存**
   - **立即下载**: 点击 **"📥 下载代码"** 获取 ZIP 包
   - **稍后访问**: 进入 **"项目管理"** 页面查看历史记录

### 2️⃣ 项目管理

在项目管理页面可以：
- ✅ 查看所有已创建的项目（管理员可看到所有人的项目）
- ✅ 查看每个项目的**创建者信息**（头像、用户名、角色）
- ✅ **重新生成**代码（基于最新配置）
- ✅ **下载**代码包
- ✅ **编辑**项目名称和描述
- ✅ **删除**不需要的项目

### 3️⃣ 用户管理（仅管理员）

管理员可以在用户管理页面：
- ➕ 创建新用户（分配角色）
- ✏️ 编辑用户信息
- 🗑️ 删除用户
- 👁️ 查看所有用户列表

### 4️⃣ 运维监控

实时监控系统运行状况：
- 📊 **总请求数**: 所有 API 调用次数（动态增长）
- 👥 **注册用户数**: 当前系统用户总数
- 💻 **代码生成次数**: 累计生成次数
- 📈 **最近事件**: 实时操作日志流
- ❤️ **健康状态**: 各服务组件状态

### 5️⃣ 文档中心

为选中的项目自动生成：
- 📘 **API 接口文档**: 包含所有 CRUD 接口说明
- ⚙️ **配置文档**: 环境变量、YAML 配置模板
- 🔧 **二次开发指南**: 架构说明、代码规范、调试技巧

---

## 🔐 安全机制

### JWT 认证流程

```
客户端                     Backend Service
  │                              │
  │  POST /auth/login            │
  │  {username, password}        │
  │ ──────────────────────────►  │
  │                              │ 验证凭据
  │                              │ 生成 JWT Token
  │  {token, user} ◄────────────│
  │                              │
  │  后续请求携带 Header:        │
  │  Authorization: Bearer <jwt> │
  │ ──────────────────────────►  │
  │                              │ 中间件验证 Token
  │  返回数据 ◄─────────────────│
```

### 权限控制

```go
// backend-service/middleware/auth_middleware.go

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        
        // 1. 提取 Token
        // 2. 解析验证 JWT
        // 3. 提取 user_id, role
        // 4. 存入 Context
        
        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}

// 使用示例
projects.GET("", func(c *gin.Context) {
    role := c.Get("role")
    
    if role != "admin" {
        // 普通用户只能查询自己的项目
        query = query.Where("user_id = ?", userID)
    } else {
        // 管理员可以看到所有项目
    }
})
```

### 密码安全

- 使用 **bcrypt** 加密存储密码
- 默认加密强度：**cost = 10**
- 密码明文**永远不会**记录到日志或返回给前端

---

## 📊 性能优化策略

### 后端优化

1. **数据库连接池**
   ```go
   sqlDB.SetMaxOpenConns(100)    // 最大连接数
   sqlDB.SetMaxIdleConns(10)     // 最大空闲连接
   sqlDB.SetConnMaxLifetime(time.Hour)  // 连接最大生命周期
   ```

2. **原子计数器**（用于统计数据）
   ```go
   var requestCount atomic.Int64
   
   func incrementRequestCounter() {
       requestCount.Add(1)  // 无锁原子操作
   }
   ```

3. **内存缓存**（Generator Service）
   - 最近生成的代码缓存在内存中
   - 数据库作为持久化备份
   - LRU 淘汰策略

### 前端优化

1. **代码分割**: React.lazy() + Suspense
2. **虚拟滚动**: 大列表性能优化
3. **防抖节流**: 搜索输入优化
4. **缓存策略**: Axios 拦截器统一处理

---

## 🧪 测试方法

### 单元测试（Go）

```bash
# 运行 Generator Service 的测试
cd generator-service
go test ./... -v

# 运行特定测试
go test -run TestGenerateCode -v
```

### API 测试（PowerShell 示例）

```powershell
# 1. 登录获取 Token
$body = @{username="admin"; password="admin123"} | ConvertTo-Json
$login = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method POST -Body $body -ContentType "application/json"
$token = $login.data.token

# 2. 生成代码
$headers = @{Authorization="Bearer $token"}
$genBody = @{
  project_name="TestProject"
  tables=@(@{name="test"; fields=@(@{name="id"; type="int"; primary=$true})})
} | ConvertTo-Json -Depth 10

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/generator/generate" `
  -Method POST -Body $genBody -ContentType "application/json" -Headers $headers

Write-Host "Project ID: $($response.data.project_id)"

# 3. 查询项目列表
$projects = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects" -Headers $headers
$projects.data | ForEach-Object { Write-Host $_.name }
```

---

## 🐳 Docker 部署（生产环境）

### docker-compose.yml 示例

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: generator-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD:-123456}
      POSTGRES_DB: generator_platform
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build: ./backend-service
    container_name: generator-backend
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD:-123456}
      DB_NAME: generator_platform
      GENERATOR_SERVICE_URL: http://generator-service:8084
    depends_on:
      postgres:
        condition: service_healthy

  generator:
    build: ./generator-service
    container_name: generator-service
    ports:
      - "8084:8084"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD:-123456}
      DB_NAME: generator_platform
    depends_on:
      postgres:
        condition: service_healthy

  frontend:
    build: ./web-admin
    container_name: generator-frontend
    ports:
      - "3000:80"
    depends_on:
      - backend

volumes:
  pgdata:
```

### 启动生产环境

```bash
# 设置环境变量
export DB_PASSWORD="your_secure_password"

# 构建并启动所有服务
docker-compose up -d --build

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

---

## ❓ 常见问题 FAQ

### Q1: 启动时提示 "port already in use"
**A**: 端口被占用，修改 `.env` 文件中的端口配置，或终止占用进程：
```bash
# Windows PowerShell
Get-NetTCPConnection -LocalPort 8080 | ForEach-Object { Stop-Process -Id $_.OwningProcess }

# Linux/macOS
lsof -ti:8080 | xargs kill -9
```

### Q2: 数据库连接失败 "password authentication failed"
**A**: 检查以下几点：
1. Docker PostgreSQL 是否正在运行：`docker ps`
2. `.env` 文件中的 `DB_PASSWORD` 是否正确（默认：`123456`）
3. 数据库名称是否为 `generator_platform`

### Q3: 代码生成返回 404 错误
**A**: 确保 Generator Service 正在运行：
```bash
# 检查进程
tasklist | findstr generator-service

# 或手动重启
cd generator-service
.\generator-service.exe
```

### Q4: 前端无法连接后端 API
**A**: 检查 Vite 代理配置 (`vite.config.js`)：
```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

### Q5: 如何重置数据库？
**A**:
```bash
# 删除容器和数据卷（会丢失所有数据！）
docker stop local-postgres
docker rm local-postgres
docker volume rm pgdata

# 重新创建
docker run -d --name local-postgres \
  -e POSTGRES_PASSWORD=123456 \
  -e POSTGRES_DB=generator_platform \
  -p 5432:5432 \
  postgres:15
```

---

## 📈 项目特色与亮点

### 1. 元数据驱动的代码生成
- 通过 JSON 配置描述数据表结构
- 模板引擎自动生成符合规范的 Go 代码
- 支持 PostgreSQL、MySQL 等多种数据库

### 2. 企业级架构设计
- 微服务拆分：Backend + Generator 独立部署
- API Gateway 统一入口
- 完善的认证授权体系

### 3. 全面的运维能力
- 实时操作日志记录
- 动态统计分析
- 服务健康检查
- 用户行为追踪

### 4. 开发者友好
- 一键启动脚本
- 详细的错误提示
- 完善的文档生成
- 代码预览和下载

### 5. 生产就绪
- Docker 容器化部署
- 环境变量配置
- 数据库迁移支持
- 日志分级输出

---

## 🎓 答辩常见问题及参考答案

### 🔥 高频问题 TOP 20

---

#### **Q1: 请介绍一下你的项目？**

**参考答案**:
> 我开发的这个项目叫做 **Generator Platform（低代码代码生成平台）**，是一个基于 Go + React 的全栈 Web 应用。
>
> **核心功能**是：用户通过可视化界面配置数据库表结构，系统能够自动生成完整的、可直接运行的 Go 后端代码。
>
> **技术架构**上采用了前后端分离的微服务思想：
> - **前端**使用 React 19 + TypeScript + Tailwind CSS，提供现代化的管理界面
> - **后端**分为两个服务：**Backend Service** 负责 API 网关、用户认证、项目管理；**Generator Service** 专注于代码生成引擎
> - **数据库**使用 PostgreSQL，通过 Docker 容器化部署
>
> **创新点**在于：
> 1. 元数据驱动的代码生成，无需手写重复性代码
> 2. 完善的多用户权限系统
> 3. 实时运维监控和动态统计
> 4. 自动生成项目文档（API文档、配置指南等）
>
> 这个项目适合快速搭建企业级 CRUD 应用的后端服务，能够显著提升开发效率。

---

#### **Q2: 为什么选择 Go 语言而不是 Java/Python？**

**参考答案**:
> 选择 Go 主要基于以下考虑：
>
> **1. 性能优势**
> - Go 是编译型语言，执行速度接近 C/C++
> - 天然支持高并发（goroutine），适合处理大量代码生成请求
> - 内存占用低，单机可支撑更多并发
>
> **2. 开发效率**
> - 语法简洁，学习曲线平缓
> - 强类型但编译速度快（秒级编译）
> - 内置工具链完善（go fmt, go test, go vet）
>
> **3. 生态匹配度**
> - Gin 框架轻量高性能，适合 RESTful API
> - GORM 是最流行的 Go ORM，支持多种数据库
> - 标准库丰富（HTTP、JSON、加密等开箱即用）
>
> **4. 部署便利**
> - 编译成单一二进制文件，无运行时依赖
> - 跨平台编译，轻松部署到 Linux 服务器
> - Docker 镜像极小（Alpine 基础镜像仅 5-10MB）
>
> 相比之下：
> - Java：虽然生态成熟，但启动慢、内存占用大（JVM）
> - Python：开发快但性能差，不适合 CPU 密集型的代码生成任务
>
> 所以对于这个**代码生成工具**的场景，Go 是最佳选择。

---

#### **Q3: 你的系统架构是怎样的？为什么要分成两个服务？**

**参考答案**:
> 系统采用 **Backend + Generator 双服务架构**：
>
> **架构图**:
> ```
> 用户 → Frontend (:3000) → Backend (:8080) → Generator (:8084) → DB
>                              ↓
>                         (API Gateway)
>                              ↓
>                    /api/v1/generator/*
> ```
>
> **拆分原因**:
>
> **1. 职责分离（单一职责原则）**
> - Backend：专注业务逻辑（用户、项目、权限）
> - Generator：专注代码生成算法（模板渲染、ZIP打包）
> - 两者耦合度低，可独立开发和测试
>
> **2. 扩展性需求**
> - 代码生成是 **CPU 密集型**操作，可能需要更多计算资源
> - 业务 API 是 **IO 密集型**，需要更多连接处理
> - 拆分后可根据负载情况**独立扩容**
>   - 例如：3个 Backend 实例 + 2个 Generator 实例
>
> **3. 故障隔离**
> - 如果 Generator 服务崩溃（比如内存溢出），不影响用户登录和项目管理
> - Backend 作为 API Gateway，可以优雅降级："代码生成暂时不可用"
>
> **4. 技术栈差异**
> - Generator 需要引入大量模板引擎、代码解析库
> - 保持 Backend 轻量，减少不必要的依赖
>
> **实现方式**:
> Backend 使用 Go 标准库的 `httputil.ReverseProxy` 实现**反向代理**：
> ```go
> // 当收到 /api/v1/generator/* 请求时
> // 自动转发到 http://localhost:8084
> proxy := httputil.NewSingleHostReverseProxy(targetURL)
> ```
>
> 对前端来说，这是**透明的**，只需要访问一个统一的 API 地址。

---

#### **Q4: 数据库为什么选择 PostgreSQL 而不是 MySQL？**

**参考答案**:
> 选择 PostgreSQL 的原因：
>
> **1. JSONB 类型支持**
> - 我们的项目需要存储**非结构化数据**：
>   - `table_config`: 表结构定义（JSON 数组）
>   - `generated_code`: 生成的代码内容（JSON 对象，key 是文件名）
>   - `db_config`: 数据库连接参数
> - PostgreSQL 的 **JSONB** 类型提供：
>   - 二进制存储效率高
>   - 支持 GIN 索引加速查询
>   - 支持丰富的 JSON 函数（查询、更新内部字段）
>
> **2. 复杂查询能力**
> - 未来可能需要的分析查询：
>   - "统计所有项目中包含 `users` 表的数量"
>   - SQL: `SELECT COUNT(*) FROM projects WHERE table_config @> '[{"name":"users"}]'`
> - PostgreSQL 的**高级特性**更强大：
>   - 窗口函数、CTE（公用表表达式）
>   - 全文搜索（FTS）
>   - 地理空间数据（PostGIS）
>
> **3. 开源协议友好**
> - PostgreSQL: **MIT-like** 开源协议，可自由商用
> - MySQL: Oracle 控制，GPL 协议有传染性风险
>
> **4. 可靠性和稳定性**
> - ACID 事务支持更完善
> - MVCC 并发控制更成熟
> - 备份恢复工具完善（pg_dump, WAL 归档）
>
> **实际应用**:
> 我们用 JSONB 存储生成的代码：
> ```sql
> -- 存储结构
> generated_code: {
>   "main.go": "package main...",
>   "models/user.go": "type User struct {...}",
>   "handlers/user.go": "func CreateUser() {...}"
> }
>
> -- 查询某个文件
> SELECT generated_code->>'main.go' FROM projects WHERE id = 15;
> ```
> 这种灵活的数据模型，MySQL 的 JSON 类型虽然也支持，但性能和功能都不如 PostgreSQL 的 JSONB。

---

#### **Q5: JWT 认证的原理是什么？你是如何实现的？**

**参考答案**:
> **JWT (JSON Web Token)** 是一种无状态的认证方案：
>
> **工作原理**:
> ```
> 1. 用户登录成功
>    → 服务器生成一个包含用户信息的 Token
>    → 格式: Header.Payload.Signature
>
> 2. 后续每次请求
>    → 客户端在 Header 中携带: Authorization: Bearer <token>
>    → 服务器验证签名，提取用户信息
>    → 无需查询数据库（无状态）
> ```
>
> **我的实现细节**:
>
> **Token 结构**:
> ```go
> type Claims struct {
>     UserID   uint   `json:"user_id"`
>     Username string `json:"username"`
>     Role     string `json:"role"`  // "admin" | "user"
>     jwt.RegisteredClaims
> }
> ```
>
> **生成 Token**:
> ```go
> token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
>     UserID:   user.ID,
>     Username: user.Username,
>     Role:     user.Role,
>     RegisteredClaims: jwt.RegisteredClaims{
>         ExpiresAt: jwt.New(time.Now().Add(24 * time.Hour)),  // 24小时过期
>         IssuedAt:  jwt.Now(),
>         NotBefore: jwt.Now(),
>         Issuer:    "generator-platform",
>     },
> })
>
> tokenString, _ := token.SignedString([]byte(jwtSecret))
> ```
>
> **验证中间件**:
> ```go
> func AuthMiddleware() gin.HandlerFunc {
>     return func(c *gin.Context) {
>         // 1. 从 Header 提取 Token
>         tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
>         
>         // 2. 解析并验证
>         token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
>             return []byte(jwtSecret), nil
>         })
>         
>         if err != nil || !token.Valid {
>             c.JSON(401, gin.H{"error": "invalid or expired token"})
>             c.Abort()
>             return
>         }
>         
>         // 3. 提取用户信息存入 Context
>         claims := token.Claims.(*Claims)
>         c.Set("user_id", claims.UserID)
>         c.Set("username", claims.Username)
>         c.Set("role", claims.Role)
>         
>         c.Next()
>     }
> }
> ```
>
> **安全性措施**:
> 1. **密码加密**: bcrypt（不可逆哈希，cost=10）
> 2. **Token 过期**: 24小时有效期
> 3. **HTTPS**: 生产环境强制使用 HTTPS 防止中间人攻击
> 4. **CORS 白名单**: 限制跨域来源
>
> **优势对比 Session**:
> | 特性 | JWT | Session |
> |------|-----|---------|
> | 存储位置 | 客户端 | 服务器（Redis/内存） |
> | 扩展性 | 天然支持分布式 | 需要 Session 共享 |
> | 性能 | 无需查库 | 每次需查询 |
> | 安全性 | 注意防 XSS | 相对安全 |

---

#### **Q6: 代码生成的具体流程是什么？模板引擎是怎么设计的？**

**参考答案**:
> **代码生成流程分为 4 个阶段**:
>
> ### 阶段 1: 输入解析
> ```go
> type GenerateRequest struct {
>     ProjectName string     `json:"project_name"`
>     DBConfig     DBConfig  `json:"db_config"`
>     Tables       []Table   `json:"tables"`  // 用户定义的表数组
> }
>
> type Table struct {
>     Name   string  `json:"name"`
>     Fields []Field `json:"fields"`
> }
>
> type Field struct {
>     Name     string `json:"name"`
>     Type     string `json:"type"`      // "int", "varchar(100)"
>     Primary  bool   `json:"primary"`
>     Nullable bool   `json:"nullable"`
> }
> ```
>
> ### 阶段 2: 类型映射
> 将通用类型映射到 Go 类型：
> ```go
> func mapToGoType(dbType string) string {
>     switch {
>     case contains(dbType, "int"):
>         return "uint"           // 主键用 uint
>     case contains(dbType, "varchar"), contains(dbType, "text"):
>         return "string"
>     case contains(dbType, "timestamp"), contains(dbType, "datetime"):
>         return "time.Time"
>     case contains(dbType, "decimal"), contains(dbType, "numeric"):
>         return "float64"
>     case contains(dbType, "bool"):
>         return "bool"
>     default:
>         return "string"
>     }
> }
> ```
>
> ### 阶段 3: 模板渲染
> 使用 **字符串拼接 + 模板变量** 生成代码：
> ```go
> func generateModelFile(table Table) string {
>     var builder strings.Builder
>     
>     builder.WriteString(fmt.Sprintf("package models\n\n"))
>     builder.WriteString(fmt.Sprintf("import \"time\"\n\n"))
>     
>     // 类名转换: users -> User
>     modelName := toPascalCase(table.Name)
>     
>     builder.WriteString(fmt.Sprintf("type %s struct {\n", modelName))
>     
>     for _, field := range table.Fields {
>         goType := mapToGoType(field.Type)
>         fieldName := toCamelCase(field.Name)
>         gormTag := ""
>         
>         if field.Primary {
>             gormTag = "primaryKey"
>         } else if !field.Nullable {
>             gormTag = "not null"
>         }
>         
>         builder.WriteString(fmt.Sprintf(
>             "    %s %s `gorm:\"%s\" json:\"%s\"`\n",
>             fieldName, goType, gormTag, field.Name,
>         ))
>     }
>     
>     builder.WriteString("}\n")
>     
>     // 添加 TableName 方法
>     builder.WriteString(fmt.Sprintf(
>         "func (%s) TableName() string { return \"%s\" }\n",
>         modelName, table.Name,
>     ))
>     
>     return builder.String()
> }
> ```
>
> ### 阶段 4: 文件组装
> ```go
> type GeneratedCode struct {
>     Language string            // "go"
>     Files    map[string]string // {"main.go": "...", "models/user.go": "..."}
> }
> 
> func doGenerate(req GenerateRequest) (*GeneratedCode, error) {
>     code := &GeneratedCode{
>         Language: "go",
>         Files:    make(map[string]string),
>     }
>     
>     // 生成各个文件
>     code.Files["main.go"] = generateMainFile(req.ProjectName, req.Tables)
>     code.Files["config/config.go"] = generateConfigFile(req.DBConfig)
>     
>     for _, table := range req.Tables {
>         modelName := toLowerCamel(table.Name)
>         code.Files[fmt.Sprintf("models/%s.go", modelName)] = generateModelFile(table)
>         code.Files[fmt.Sprintf("handler/%s.go", modelName)] = generateHandlerFile(table)
>         code.Files[fmt.Sprintf("router/%s.go", modelName)] = generateRouterFile(table)
>     }
>     
>     return code, nil
> }
> ```
>
> **最终输出结构**:
> ```
> generated-project/
> ├── main.go                 # 入口文件
> ├── config/
> │   └── config.go           # 数据库配置
> ├── internal/
> │   ├── models/
> │   │   ├── user.go         # User 模型
> │   │   └── product.go      # Product 模型
> │   ├── handler/
> │   │   ├── user.go         # User Handler
> │   │   └── product.go      # Product Handler
> │   └── router/
> │       └── router.go       # 路由注册
> ├── go.mod
> └── .env
> ```

---

#### **Q7: 如何保证普通用户只能看到自己的项目？（权限控制）**

**参考答案**:
> 采用 **RBAC（基于角色的访问控制）** + **数据隔离** 策略：
>
> **1. 角色定义**
> ```sql
> -- 用户表中的 role 字段
> ENUM('admin', 'user')
> ```
>
> **2. JWT Token 中携带角色**
> ```go
> // 登录时生成 Token
> token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
>     UserID:   user.ID,
>     Role:     user.Role,  // "admin" 或 "user"
>     // ...
> })
> ```
>
> **3. 查询时的权限过滤**
> ```go
> func listProjectsHandler(c *gin.Context) {
>     // 从 Context 获取当前用户信息
>     userID, _ := c.Get("user_id")      // uint 类型
>     role, _ := c.Get("role")            // "admin" 或 "user"
>     
>     var projects []models.Project
>     query := database.DB.Preload("User")  // 预加载关联的用户信息
>     
>     // 关键：根据角色过滤！
>     if role != models.RoleAdmin {
>         // 普通用户：只查询自己的项目
>         query = query.Where("user_id = ?", userID)
>     }
>     // 管理员：不添加条件，查询所有项目
>     
>     query.Find(&projects)
>     
>     c.JSON(200, gin.H{"data": projects})
> }
> ```
>
> **4. 代码生成时的用户绑定**
> ```go
> func generateCode(c *gin.Context) {
>     // 从 JWT 提取真实用户 ID（不是硬编码！）
>     currentUserID, _ := c.Get("user_id")
>     currentUsername, _ := c.Get("username")
>     
>     // 保存项目时关联当前用户
>     dbProject := models.Project{
>         UserID: currentUserID.(uint),  // ← 关键！
>         Name:   req.ProjectName,
>         // ...
>     }
>     db.Create(&dbProject)
> }
> ```
>
> **5. 前端展示创建者信息**
> ```jsx
> // Projects.jsx
> <td>
>   <div className="flex items-center gap-2">
>     <div className="w-8 h-8 rounded-full bg-blue-100">
>       {/* 显示首字母 */}
>       {(project.user?.username || '?')[0].toUpperCase()}
>     </div>
>     <div>
>       <span>{project.user?.username}</span>
>       <span>{project.user?.role === 'admin' ? '管理员' : '用户'}</span>
>     </div>
>   </div>
> </td>
> ```
>
> **效果**:
> - 普通用户登录后，只能看到自己创建的项目
> - 管理员可以看到**所有用户的项目**，并且能看到每个项目的**创建者是谁**
> - 数据库层面通过 `WHERE user_id = ?` 保证安全性

---

#### **Q8: 运维监控的统计数据是如何实现的？是实时的吗？**

**参考答案**:
> 采用 **内存原子计数器 + 数据库实时查询** 的混合策略：
>
> **1. 总请求数（内存原子计数器）**
> ```go
> // 定义全局原子变量
> var requestCount atomic.Int64
> 
> // 每次请求到达时递增
> func incrementRequestCounter() {
>     requestCount.Add(1)  // 原子操作，无需加锁
> }
> 
> // 在每个 Handler 开始处调用
> func someHandler(c *gin.Context) {
>     incrementRequestCounter()  // 先计数
>     // ... 处理业务逻辑
> }
> 
> // 统计接口返回当前值
> func getStatsHandler(c *gin.Context) {
>     stats := gin.H{
>         "total_requests": requestCount.Load(),  // 原子读取
>     }
>     c.JSON(200, stats)
> }
> ```
> **特点**:
> - ✅ 实时更新（每次请求立即反映）
> - ✅ 无锁高性能（atomic 包底层使用 CPU CAS 指令）
> - ⚠️ 服务重启后归零（可用 Redis 持久化解决）
>
> **2. 代码生成次数（数据库聚合查询）**
> ```go
> func getStatsHandler(c *gin.Context) {
>     var generationCount int64
     
>     // 查询 operation_logs 表中 generate/regenerate 的记录数
>     database.DB.Model(&models.OperationLog{}).
>         Where("action IN ?", []string{"generate", "regenerate"}).
>         Count(&generationCount)
>     
>     stats["generated_codes"] = generationCount
> }
> ```
> **特点**:
> - ✅ 持久化（重启不丢失）
> - ✅ 精确（来自实际操作记录）
> - ⚠️ 有微小延迟（毫秒级，可忽略）
>
> **3. 注册用户数（数据库 COUNT）**
> ```go
> var userCount int64
> database.DB.Model(&models.User{}).Count(&userCount)
> ```
>
> **4. 操作日志（实时写入）**
> ```go
> func recordOperationLog(c *gin.Context, action, resource string, ...) {
>     log := models.OperationLog{
>         UserID:     currentUserID,
>         Action:     action,       // "generate", "download", "login"
>         Resource:   resource,     // "project", "code"
>         Status:     status,       // "success", "failed"
>         Duration:   time.Since(startTime).Milliseconds(),  // 耗时
>         IPAddress: c.ClientIP(),
>         UserAgent: c.Request.UserAgent(),
>     }
>     database.DB.Create(&log)  // 异步写入（未来可优化）
> }
> ```
>
> **总结**:
> | 指标 | 数据来源 | 实时性 | 持久性 |
> |------|---------|--------|--------|
> | 总请求数 | 内存原子计数器 | ✅ 实时 | ❌ 重启清零 |
> | 生成次数 | 数据库 COUNT | ⚠️ 近实时 | ✅ 持久 |
> | 用户数 | 数据库 COUNT | ⚠️ 近实时 | ✅ 持久 |
> | 操作日志 | 数据库 INSERT | ✅ 实时 | ✅ 持久 |

---

#### **Q9: 前端用了哪些技术？为什么选择 React 而不是 Vue？**

**参考答案**:
> **前端技术栈**:
> - **框架**: React 19（最新版本）
> - **构建工具**: Vite 8（极速 HMR）
> - **样式**: Tailwind CSS 4（原子化 CSS）
> - **路由**: React Router 6
> - **状态管理**: React Context + useReducer
> - **HTTP**: Axios
> - **图表**: Recharts
> - **UI 库**: MUI (Material-UI) + Ant Design
> - **国际化**: i18next
>
> **选择 React 的原因**:
>
> **1. 生态系统成熟**
> - npm 上 React 相关包数量最多
> - 遇到问题容易找到解决方案
> - 组件库选择丰富（MUI、Ant Design、NextUI 都支持 React）
>
> **2. Hooks 编程范式**
> ```jsx
> // React 19 的 Hooks 让代码更简洁
> function ProjectList() {
>   const [projects, setProjects] = useState([])
>   const [loading, setLoading] = useState(true)
>   
>   useEffect(() => {
>     fetchProjects()
>   }, [])  // 依赖数组自动管理副作用
>   
>   return (
>     <Table data={projects} loading={loading} />
>   )
> }
> ```
>
> **3. TypeScript 支持优秀**
> - React + TypeScript 是黄金组合
> - 类型安全减少 Bug
> - IDE 智能提示完善
>
> **4. 团队/社区因素**
> - 个人熟悉程度
> - 招聘市场 React 开发者更多
>
> **Vue 也很好**，只是个人选择。如果时间充裕，可以用 Vue 重写前端做对比。

---

#### **Q10: 项目中遇到的最大困难是什么？怎么解决的？**

**参考答案**:
> **最大的困难**: **普通用户无法看到自己生成的项目**
>
> **问题描述**:
> - 用户反馈：普通用户生成代码后，在"项目管理"页面看不到项目
> - 管理员可以看到所有项目，但都显示为 admin 创建
>
> **排查过程**（耗时约 2 小时）:
>
> **第一层：发现 UserID 硬编码**
> ```go
> // handlers.go 第 148 行（原始代码）
> dbProject := models.Project{
>     UserID: 1,  // ❌ 硬编码为 admin 的 ID！
>     Name:   req.ProjectName,
> }
> ```
> **修复**: 从 JWT Token 提取真实用户 ID
> ```go
> currentUserID, _ := c.Get("user_id")
> dbProject.UserID = currentUserID.(uint)  // ✅
> ```
>
> **第二层：发现数据库连接失败**
> - 修复 UserID 后，仍然无法保存到数据库
> - 错误日志：`password authentication failed for user "postgres"`
>
> **根因定位**:
> Generator Service 使用的是公共配置包 `go-common/config`，其中默认密码是 `"postgres"`，而 Docker PostgreSQL 的密码是 `"123456"`
>
> ```go
> // go-common/config/config.go（原始代码）
> DBPassword: getEnv("DB_PASSWORD", "postgres")  // ❌ 错误默认值
> 
> // 修复后
> DBPassword: getEnv("DB_PASSWORD", "123456")    // ✅ 正确密码
> ```
>
> **第三层：数据库名称错误**
> ```go
> // generator-service/config.go（原始代码）
> DBName: getEnv("DB_NAME", "mydb")  // ❌ 不存在的数据库
> 
> // 修复后
> DBName: getEnv("DB_NAME", "generator_platform")  // ✅
> ```
>
> **最终结果**:
> - ✅ Generator Service 成功连接 PostgreSQL
> - ✅ 普通用户生成的项目正确关联到该用户
> - ✅ 项目管理页面正常显示
> - ✅ 文档中心也能看到（共用同一 API）
>
> **经验教训**:
> 1. **不要硬编码敏感值**（用户 ID、密码）
> 2. **统一配置管理**（避免多个 config.go 默认值不一致）
> 3. **充分测试**（不仅要测 admin，还要测普通用户）
> 4. **查看启动日志**（数据库连接失败会有明确报错）

---

#### **Q11: 如果用户量增大，系统如何扩展？**

**参考答案**:
> **水平扩展方案**:
>
> **1. 无状态化设计**
> - JWT Token 存储在客户端，服务器不保存会话状态
> - Backend 和 Generator 都是无状态的，可以任意增加实例
>
> **2. 负载均衡**
> ```nginx
> upstream backend_cluster {
>     server backend-1:8080;
>     server backend-2:8080;
>     server backend-3:8080;
> }
> 
> upstream generator_cluster {
>     server generator-1:8084;
>     server generator-2:8084;
> }
> 
> server {
>     listen 80;
>     
>     location /api/v1/auth/     { proxy_pass http://backend_cluster; }
>     location /api/v1/users/    { proxy_pass http://backend_cluster; }
>     location /api/v1/projects/  { proxy_pass http://backend_cluster; }
>     location /api/v1/generator/ { proxy_pass http://generator_cluster; }
> }
> ```
>
> **3. 数据库优化**
> - **读写分离**: 主库写，从库读
> - **分库分表**: 按用户 ID 取模分片
> - **缓存层**: Redis 缓存热点数据（用户信息、项目列表）
>
> **4. 消息队列（异步化）**
> - 代码生成任务放入队列（RabbitMQ/Kafka）
> - Generator Worker 异步消费任务
> - 用户轮询或 WebSocket 推送结果
>
> **5. CDN + 静态化**
> - 前端静态资源上传 CDN
> - 生成的代码包存储到对象存储（S3/OSS）
>
> **预估容量**（单机）:
> - QPS: ~1000-5000（取决于代码生成复杂度）
> - 并发用户: ~500-1000
> - 数据量: 百万级项目记录（PostgreSQL 轻松应对）

---

#### **Q12: 如何保证数据一致性？事务是怎么处理的？**

**参考答案**:
> **本项目的事务场景相对简单**，主要涉及：
>
> **1. 用户注册（双表插入）**
> ```go
> func registerHandler(c *gin.Context) {
>     // 开启事务
>     tx := database.DB.Begin()
>     
>     // 1. 创建用户
>     if err := tx.Create(&user).Error; err != nil {
>         tx.Rollback()
>         return
>     }
>     
>     // 2. 记录操作日志
>     if err := tx.Create(&log).Error; err != nil {
>         tx.Rollback()
>         return
>     }
>     
>     // 提交事务
>     tx.Commit()
> }
> ```
>
> **2. 代码生成（内存 + 数据库双写）**
> ```go
> func generateCode(c *gin.Context) {
>     // 1. 先生成代码到内存（快速响应用户）
>     storeMutex.Lock()
>     memoryStore[projectID] = generatedCode
>     storeMutex.Unlock()
>     
>     // 2. 异步持久化到数据库（最终一致性）
>     go func() {
>         db.Create(&dbProject)  // 失败不影响用户体验
>     }()
> }
> ```
>
> **3. GORM 事务封装**
> ```go
> // DAO 层提供事务支持
> func (d *DAO) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
>     return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
>         ctx = context.WithValue(ctx, "tx", tx)
>         return fn(ctx)
>     })
> }
> 
> // 使用示例
> dao.Transaction(ctx, func(txCtx context.Context) error {
>     txDAO := dao.WithTx(txCtx)
>     txDAO.CreateUser(user)
>     txDAO.CreateLog(log)
>     return nil  // 自动提交
> })
> ```
>
> **ACID 特性保证**:
> - **Atomicity（原子性）**: GORM Transaction 要么全成功，要么全回滚
> - **Consistency（一致性）**: 外键约束、唯一索引保证数据合法
> - **Isolation（隔离性）**: PostgreSQL 默认 Read Committed 隔离级别
> - **Durability（持久性）**: WAL (Write-Ahead Log) 保证提交不丢失

---

#### **Q13: 前端的国际化（i18n）是如何实现的？**

**参考答案**:
> 使用 **react-i18next** 库实现完整的中英文切换：
>
> **1. 配置初始化**
> ```jsx
> // i18n/index.js
> import i18n from 'i18next';
> import { initReactI18next } from 'react-i18next';
> import zh from './locales/zh.json';
> import en from './locales/en.json';
> 
> i18n.use(initReactI18next).init({
>   resources: {
>     zh: { translation: zh },
>     en: { translation: en },
>   },
>   lng: 'zh',           // 默认中文
>   fallbackLng: 'en',   // 回退英文
>   interpolation: {
>     escapeValue: false,  // React 已经处理 XSS
>   },
> });
> ```
>
> **2. 翻译文件结构**
> ```json
> // zh.json
> {
>   "common": {
>     "save": "保存",
>     "cancel": "取消",
>     "creator": "创建者"
>   },
>   "projects": {
>     "title": "项目管理",
>     "noProjects": "暂无项目"
>   }
> }
> ```
>
> **3. 组件中使用**
> ```jsx
> import { useTranslation } from 'react-i18next';
> 
> function Projects() {
>   const { t } = useTranslation();
>   
>   return (
>     <h1>{t('projects.title')}</h1>  // → "项目管理" or "Project Management"
>     <button>{t('common.save')}</button>
>   );
> }
> ```
>
> **4. 语言切换组件**
> ```jsx
> function LanguageToggle() {
>   const { i18n } = useTranslation();
>   
>   return (
>     <select value={i18n.language} onChange={(e) => i18n.changeLanguage(e.target.value)}>
>       <option value="zh">中文</option>
>       <option value="en">English</option>
>     </select>
>   );
> }
> ```
>
> **5. 动态插值**
> ```json
> // 翻译文件
> "welcome": "欢迎回来，{{name}}！"
> 
> // 使用
> t('welcome', { name: 'Admin' })  // → "欢迎回来，Admin！"
> ```

---

#### **Q14: 项目的不足之处是什么？未来如何改进？**

**参考答案**:
> **当前不足**:
>
> **1. 功能层面**
> - ❌ 不支持**数据库反向工程**（从已有数据库导入表结构）
> - ❌ 代码模板**不够灵活**（目前固定生成 Go/Gin/GORM）
> - ❌ 没有**在线编辑器**（生成的代码只能下载，不能在线修改）
> - ❌ 没有**版本控制**（无法对比不同版本的生成结果）
> - ❌ 不支持**协作开发**（多人同时编辑同一项目）
>
> **2. 技术层面**
> - ❌ **没有单元测试覆盖**（只有 Generator Service 有少量测试）
> - ❌ **没有 CI/CD 流水线**（GitHub Actions 自动构建）
> - ❌ **缺少监控告警**（Prometheus + Grafana）
> - ❌ **日志系统简陋**（只有标准输出，没有 ELK/Splunk）
> - ❌ **没有 API 文档自动生成**（Swagger/OpenAPI）
>
> **3. 性能层面**
> - ❌ 代码生成是**同步阻塞**的（应该异步化）
> - ❌ 没有**缓存层**（Redis 缓存热点数据）
> - ❌ 数据库**没有索引优化**（大数据量下性能差）
>
> **改进路线图**:
>
> **短期（1-2 周）**:
> - ✅ 添加 Swagger 文档
> - ✅ 补充单元测试（目标覆盖率 > 70%）
> - ✅ 代码生成异步化（消息队列）
> - ✅ 添加 Redis 缓存
>
> **中期（1-2 月）**:
> - 🔧 支持多种技术栈（Java Spring Boot, Python FastAPI）
> - 🔧 数据库逆向工程（连接数据库自动读取表结构）
> - 🔧 在线 IDE（Monaco Editor 预览/编辑代码）
> - 🔧 GitHub/GitLab 集成（直接推送到仓库）
>
> **长期（3-6 月）**:
> - 🚀 多租户 SaaS 化
> - 🚀 插件市场（自定义代码模板）
> - 🚀 AI 辅助生成（自然语言描述需求）
> - 🚀 移动端 App（React Native/Flutter）

---

#### **Q15: 如何保证系统的安全性？**

**参考答案**:
> **多层防御策略**:
>
> **1. 认证安全**
> - ✅ **bcrypt** 密码哈希（不可逆，抗彩虹表）
> - ✅ **JWT Token** 过期机制（24小时）
> - ✅ **HTTPS** 强制加密传输（生产环境）
> - ✅ **防暴力破解**（可添加登录次数限制）
>
> **2. 授权安全**
> - ✅ **RBAC** 角色权限（admin/user）
> - ✅ **数据隔离**（普通用户只能访问自己的数据）
> - ✅ **中间件统一校验**（每个受保护路由都要经过 AuthMiddleware）
>
> **3. 输入验证**
> ```go
> // 后端严格验证
> type GenerateRequest struct {
>     ProjectName string  `json:"project_name" binding:"required,min=3,max=100"`
>     Tables      []Table `json:"tables" binding:"required,min=1,dive"`
> }
> 
> type Table struct {
>     Name   string  `json:"name" binding:"required,alpha"`
>     Fields []Field `json:"fields" binding:"required,min=1,dive"`
> }
> ```
>
> **4. SQL 注入防护**
> - ✅ 使用 **GORM 参数化查询**（永远不拼接 SQL 字符串）
> ```go
> // ✅ 安全（参数化）
> db.Where("user_id = ?", userID)
> 
> // ❌ 危险（SQL 注入）
> db.Where("user_id = " + userID)  // 禁止！
> ```
>
> **5. XSS 防护**
> - ✅ React **自动转义** JSX 中的变量
> - ✅ CSP (Content Security Policy) 头部
> - ✅ Cookie 的 HttpOnly、Secure 标志
>
> **6. CORS 策略**
> ```go
> func CORSMiddleware() gin.HandlerFunc {
>     return func(c *gin.Context) {
>         c.Writer.Header().Set("Access-Control-Allow-Origin", "*")  // 生产环境应限定域名
>         c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
>         c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
>         
>         if c.Request.Method == "OPTIONS" {
>             c.AbortWithStatus(204)
>             return
>         }
>         c.Next()
>     }
> }
> ```
>
> **7. 其他安全措施**
> - ✅ **速率限制**（Rate Limiting，防止 DDoS）
> - ✅ **请求体大小限制**（防止大文件攻击）
> - ✅ **安全头设置**（X-Frame-Options, X-Content-Type-Options）
> - ✅ **依赖定期更新**（`go get -u`, `npm audit fix`）

---

#### **Q16: 代码生成器的核心技术难点在哪里？**

**参考答案**:
> **三大技术挑战**:
>
> **挑战 1: 类型系统映射**
>
> **问题**: 不同数据库的类型系统差异巨大
> ```
> PostgreSQL          Go 类型          GORM Tag
> ----------          -------          --------
> SERIAL              uint             primaryKey
> VARCHAR(100)        string           type:varchar(100)
> INTEGER             int              
> BIGINT              int64            
> TIMESTAMP           time.Time       
> BOOLEAN             bool            
> DECIMAL(10,2)       float64          
> TEXT                string           
> BYTEA               []byte          
> UUID                string           type:uuid
> ARRAY[]             pq.StringArray   需要特殊处理
> JSONB               json.RawMessage  需要特殊处理
> ```
>
> **解决方案**: 设计**可扩展的类型映射表**
> ```go
> var typeMapping = map[string]string{
>     "int": "uint", "integer": "uint", "serial": "uint",
>     "varchar": "string", "text": "string", "char": "string",
>     "timestamp": "time.Time", "date": "time.Time",
>     "boolean": "bool", "bool": "bool",
>     "decimal": "float64", "numeric": "float64",
>     // ...
> }
> 
> func mapToGoType(dbType string) string {
>     lower := strings.ToLower(strings.Split(dbType, "(")[0])
>     if goType, ok := typeMapping[lower]; ok {
>         return goType
>     }
>     return "string"  // 默认回退
> }
> ```
>
> **挑战 2: 命名规范转换**
>
> **问题**: 数据库命名风格 vs Go 命名风格
> ```
> 数据库: user_name, order_items, created_at
> Go:     UserName, OrderItems, CreatedAt
> 
> 数据库: users, products, order_details
> Go:     User, Product, OrderDetail (单数 + PascalCase)
> ```
>
> **解决方案**: 实现**命名转换工具函数**
> ```go
> // Snake Case → Camel Case
> func toCamelCase(s string) string {
>     parts := strings.Split(s, "_")
>     for i := 1; i < len(parts); i++ {
>         parts[i] = strings.Title(parts[i])
>     }
>     return strings.Join(parts, "")
> }
> 
> // Plural → Singular (简单规则)
> func toSingular(s string) string {
>     if strings.HasSuffix(s, "ies") {
>         return s[:len(s)-3] + "y"  // entities → entity
>     } else if strings.HasSuffix(s, "ses") {
>         return s[:len(s)-2]         // addresses → address
>     } else if strings.HasSuffix(s, "s") {
>         return s[:len(s)-1]         // users → user
>     }
>     return s
> }
> 
> // 组合使用
> func toModelName(tableName string) string {
>     singular := toSingular(tableName)     // users → user
>     pascal := toCamelCase(singular)       // user → User
>     return pascal
> }
> ```
>
> **挑战 3: 代码模板的可维护性**
>
> **问题**: 硬编码字符串拼接难以维护
>
> **解决方案演进**:
>
> **阶段 1: 字符串拼接（当前实现）**
> ```go
> builder.WriteString(fmt.Sprintf(
>     "func (ctrl *%sController) Create(c *gin.Context) {\n", modelName))
> // ...
> ```
> **缺点**: 代码冗长、易出错、难维护
>
> **阶段 2: Go Template（推荐改进）**
> ```go
> // templates/handler.tmpl
> func (ctrl *{{.ModelName}}Controller) Create(c *gin.Context) {
>     var req {{.ModelName}}CreateRequest
>     if err := c.ShouldBindJSON(&req); err != nil {
>         response.BadRequest(c, err.Error())
>         return
>     }
>     
>     result, err := ctrl.service.Create(c.Request.Context(), &req)
>     if err != nil {
>         response.Error(c, http.StatusInternalServerError, err.Error())
>         return
>     }
>     
>     response.Created(c, result)
> }
> 
> // 渲染
> tmpl := template.Must(template.ParseFiles("templates/handler.tmpl"))
> tmpl.Execute(buf, map[string]interface{}{
>     "ModelName": "User",
>     "Fields":    fields,
> })
> ```
>
> **阶段 3: 外部模板引擎（终极方案）**
> - 使用 **Twig / Jinja2 / Handlebars** 等专业模板引擎
> - 支持继承、宏、过滤器
> - 非技术人员也可修改模板

---

#### **Q17: 前后端是如何联调的？开发流程是怎样的？**

**参考答案**:
> **开发工作流**:
>
> **1. 接口定义先行**
> ```yaml
> # API 设计文档（YAML 或 OpenAPI）
> POST /api/v1/generator/generate
> Request:
>   project_name: string (required)
>   tables: array (required, min 1)
>     - name: string (required)
>       fields: array (required, min 1)
>         - name: string (required)
>           type: string (required)
>           primary: boolean
> Response:
>   code: 200
>   data:
>     project_id: integer
>     code:
>       language: "go"
>       files: object
> ```
>
> **2. 后端开发（并行）**
> ```bash
# 1. 定义数据结构
# types.go

# 2. 实现 Handler
# handlers.go

# 3. 编写单元测试
# types_test.go

# 4. 启动服务测试
go run .
curl -X POST http://localhost:8084/api/v1/generator/generate \
  -H "Content-Type: application/json" \
  -d '{"project_name":"Test","tables":[...]}' | jq
> ```
>
> **3. 前端开发（并行）**
> ```jsx
> // api.jsx - 封装 API 调用
> export const generatorAPI = {
>   generate: (data) => 
>     axios.post('/api/v1/generator/generate', data),
>   download: (projectId) =>
>     axios.get(`/api/v1/generator/download/${projectId}`, { responseType: 'blob' }),
> };
> 
> // Generator.jsx - 页面组件
> const handleGenerate = async () => {
>   setLoading(true);
>   try {
>     const response = await generatorAPI.generate(formData);
>     setGeneratedCode(response.data.data.code);
>   } catch (error) {
>     setError(error.message);
>   } finally {
>     setLoading(false);
>   }
> };
> ```
>
> **4. Vite 代理解决跨域**
> ```javascript
> // vite.config.js
> export default defineConfig({
>   server: {
>     proxy: {
>       '/api': {
>         target: 'http://localhost:8080',  // Backend Service
>         changeOrigin: true,
>       },
>     },
>   },
> });
> ```
>
> **5. 联调测试**
> ```bash
> # Terminal 1: 启动 Backend
> cd backend-service && go run .
> 
> # Terminal 2: 启动 Generator
> cd generator-service && go run .
> 
> # Terminal 3: 启动前端
> cd web-admin && npm run dev
> 
> # Browser: 打开 http://localhost:3000
> # F12: Network 面板观察 API 请求
> ```
>
> **调试技巧**:
> - **后端**: `log.Printf("[DEBUG] ...")` + 查看终端输出
> - **前端**: Chrome DevTools → Network → 查看 Request/Response
> - **数据库**: `docker exec -it local-postgres psql -U postgres -d generator_platform`
> - **接口测试**: Postman 或 curl 命令

---

#### **Q18: Docker 部署的优势是什么？容器化需要注意什么？**

**参考答案**:
> **Docker 优势**:
>
> **1. 环境一致性**
> ```
> 问题: "在我电脑上能跑，在服务器上不行"
> 解决: Docker 镜像打包了所有依赖（OS、Runtime、Libraries）
>       本地、测试、生产环境完全一致
> ```
>
> **2. 快速部署**
> ```bash
> # 传统方式（30分钟）
> 1. 安装 Go 环境
> 2. 安装 Node.js
> 3. 安装 PostgreSQL
> 4. 配置环境变量
> 5. 编译后端
> 6. 安装前端依赖
> 7. 构建前端
> 8. 配置 Nginx...
> 
> # Docker 方式（3分钟）
> docker-compose up -d  # 一条命令搞定！
> ```
>
> **3. 资源隔离**
> - 每个容器独立的文件系统、网络、进程空间
> - 不会污染宿主机环境
> - 可以运行多个版本的服务（PostgreSQL 12, 13, 14, 15 并存）
>
> **4. 易于扩展**
> ```bash
> # 水平扩展：3 个 Backend 实例
> docker-compose up -d --scale backend=3
> 
> # 滚动更新：零停机部署
> docker-compose up -d --no-deps --build backend
> ```
>
> **容器化注意事项**:
>
> **1. 镜像大小优化**
> ```dockerfile
> # ❌ 差：基础镜像太大（1GB+）
> FROM golang:1.21
> 
> # ✅ 好：多阶段构建（最终镜像 10-20MB）
> # 阶段 1: 编译
> FROM golang:1.21-alpine AS builder
> WORKDIR /app
> COPY . .
> RUN CGO_ENABLED=0 GOOS=linux go build -o server .
> 
> # 阶段 2: 运行
> FROM alpine:3.19
> COPY --from=builder /app/server /usr/local/bin/server
> CMD ["server"]
> ```
>
> **2. 数据持久化**
> ```yaml
> services:
>   postgres:
>     volumes:
>       - pgdata:/var/lib/postgresql/data  # ⚠️ 必须挂载卷！否则容器删除数据丢失
 
> volumes:
>   pgdata:  # 命名卷，即使容器删除数据仍在
> ```
>
> **3. 安全配置**
> ```yaml
> services:
>   backend:
>     # 不要以 root 运行
>     user: "1000:1000"
>     # 只读根文件系统
>     read_only: true
>     # 限制能力
>     cap_drop:
>       - ALL
>     cap_add:
>       - NET_BIND_SERVICE  # 只保留绑定端口的权限
> ```
>
> **4. 健康检查**
> ```yaml
> services:
>   backend:
>     healthcheck:
>       test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
>       interval: 30s
>       timeout: 5s
>       retries: 3
>       start_period: 10s
> ```

---

#### **Q19: 如果让你重新设计这个项目，你会怎么做？**

**参考答案**:
> **理想架构（基于现有经验的重构版）**:
>
> **1. 微服务化改造**
> ```
> 当前: Monolith (Backend + Generator)
> 理想: 独立微服务
> 
> ├─ auth-service        # 认证服务（OAuth2.0 / OIDC）
> ├─ user-service        # 用户管理
> ├─ project-service      # 项目 CRUD
> ├─ generator-service    # 代码生成（CPU 密集，独立扩展）
> ├─ file-service         # 文件存储（S3/MinIO）
> ├─ notification-service # 邮件/消息通知
> └─ gateway-service      # API 网关（鉴权、限流、路由）
> ```
>
> **2. 引入消息队列**
> ```
> 当前: 同步生成（用户等待 5-10 秒）
> 理想: 异步生成
> 
> 用户点击"生成" → 发送到 RabbitMQ/Kafka
>                  → 立即返回 "任务已提交"
>                  → Generator Worker 消费任务
>                  → 生成完成后推送 WebSocket 通知
>                  → 用户刷新页面查看结果
> ```
>
> **3. 数据库升级**
> ```
> 当前: 单 PostgreSQL 实例
> 理想: 分库分表 + 缓存
> 
> ├─ 主库 (PostgreSQL)  写操作
> ├─ 从库 (PostgreSQL × 2) 读操作
> ├─ Redis Cluster      缓存热点数据
> ├─ Elasticsearch      全文搜索（项目、日志）
> └─ MinIO/S3           文件对象存储
> ```
>
> **4. 前端重构**
> ```
> 当前: React 19 + JavaScript
> 理想: Next.js + TypeScript + SSR
> 
> ├─ Next.js 14 (App Router)
> ├─ TypeScript (类型安全)
> ├─ Server Components (SEO 友好)
> ├─ Zustand/Jotai (轻量状态管理)
> ├─ TanStack Query (数据获取/缓存)
> └─ shadcn/ui (现代组件库)
> ```
>
> **5. DevOps 完善**
> ```
> 当前: 手动部署
> 理想: GitOps 自动化流水线
> 
> ├─ GitHub Actions (CI/CD)
> │   ├─ PR 自动运行测试
> │   ├─ 合并后自动构建 Docker 镜像
> │   └─ 自动部署到 Staging/Production
> ├─ ArgoCD (GitOps)
> │   └─ 声明式 K8s 部署
> ├─ Prometheus + Grafana
> │   └─ 监控告警
> └─ ELK Stack
>     └─ 日志聚合分析
> ```
>
> **6. 新增功能**
> - 🤖 **AI 辅助**: "帮我生成一个电商后台，包含商品、订单、用户三个表"
> - 🔄 **数据库逆向**: 连接已有数据库，自动导入表结构
> - 🎨 **可视化建模**: ER 图拖拽式设计数据模型
> - 📦 **插件市场**: 社区贡献代码模板（Spring Boot, FastAPI, Laravel...）
> - 👥 **团队协作**: 多人实时编辑、评论、版本对比
>
> **但是！** 对于**课程设计/毕业设计**而言，当前的架构已经足够：
> - ✅ 功能完整
> - ✅ 技术栈主流
> - ✅ 代码质量良好
> - ✅ 文档齐全
>
> **过度设计**反而是扣分项！

---

#### **Q20: 你在这个项目中学到了什么？最大的收获是什么？**

**参考答案**:
> **技术层面的收获**:
>
> **1. 全栈开发能力**
> - 从前到后都能独立完成（前端 React + 后端 Go + 数据库 SQL）
> - 理解了前后端交互的完整链路（HTTP → Middleware → Handler → DB → Response）
> - 掌握了 RESTful API 设计规范
>
> **2. 系统设计思维**
> - 学会如何**拆分服务**（职责分离、关注点分离）
> - 学会**权衡取舍**（同步 vs 异步、强一致 vs 最终一致）
> - 学会**从用户角度思考**（权限、体验、容错）
>
> **3. 工程化实践**
> - 版本控制（Git 分支管理）
> - 环境配置（.env 文件、Docker 容器化）
> - 日志和调试技巧（结构化日志、错误追踪）
> - 代码组织（分层架构、公共库抽取）
>
> **4. 问题解决能力**
> - 本次遇到的 **UserID 硬编码问题**让我深刻理解了：
>   - **不要相信"应该没问题"**，要实际测试
>   - **配置管理的重要性**（多个 config.go 导致的不一致）
>   - **日志的价值**（数据库连接失败的错误信息帮助快速定位）
>
> **软技能层面的收获**:
>
> **1. 时间管理**
> - 大项目需要**任务分解**（Todo List、优先级排序）
> - **迭代开发**（先实现核心功能，再逐步完善）
> - **风险控制**（预留缓冲时间应对意外问题）
>
> **2. 文档和沟通**
> - 好的代码需要**好的文档**（注释、README、API 文档）
> - **答辩准备**要站在**评审老师的角度**思考问题
> - 清晰表达**技术决策的原因**（为什么选 A 而不是 B）
>
> **3. 持续学习**
> - 实践中发现不足 → 查阅官方文档 → 学习新技术
> - 例如：为了解决跨域问题，学习了 CORS 和 Proxy
> - 为了优化性能，学习了 atomic 包和连接池
>
> **最大的收获**:
> > **"纸上得来终觉浅，绝知此事要躬行"**
>
> 看再多的教程视频，不如亲手做一个完整项目。在这个过程中遇到的问题、踩过的坑、最终的解决方案，才是真正属于自己的知识财富。
>
> 这个项目让我从一个"会写代码的人"，成长为一个**"能交付产品的工程师"**。

---

## 📚 参考资料

### 官方文档
- [Go 语言官方文档](https://go.dev/doc/)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/docs/)
- [React 官方文档](https://react.dev/)
- [Vite 中文文档](https://cn.vitejs.dev/)
- [Tailwind CSS 文档](https://tailwindcss.com/docs)

### 推荐阅读
- 《Go 语言编程》- 许式伟
- 《Designing Data-Intensive Applications》- Martin Kleppmann
- 《Clean Architecture》- Robert C. Martin

### 在线工具
- [Draw.io](https://app.diagrams.net/) - 架构图绘制
- [DBDiagram](https://dbdiagram.io/) - ER 图设计
- [Swagger Editor](https://editor.swagger.io/) - API 文档

---

## 📄 许可证

MIT License

Copyright (c) 2026 Generator Platform

---

## 👨‍💻 作者信息

- **项目名称**: Generator Platform - 低代码代码生成平台
- **开发周期**: 2026年5月
- **技术栈**: Go 1.21 + React 19 + PostgreSQL 15 + Docker
- **代码行数**: 约 8000+ 行（后端 5000+，前端 3000+）

---

> 🎯 **祝您答辩顺利！如有任何问题，欢迎随时交流讨论！**

---

## 🔗 项目仓库

- **GitHub**: https://github.com/LeoNorth2004/server-generator-local
- **使用说明**: [USER_GUIDE.md](./USER_GUIDE.md) - 快速上手指南
- **Star ⭐**: 如果这个项目对您有帮助，欢迎给个 Star 支持一下！
