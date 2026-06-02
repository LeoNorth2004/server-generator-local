# 📘 Generator Platform 使用说明书

> **版本**: 1.0.0  
> **更新日期**: 2026年6月  
> **适用系统**: Windows 10/11 (64位)

---

## 📖 目录

1. [简介](#简介)
2. [环境要求](#环境要求)
3. [获取项目](#快速开始) - GitHub 克隆 / ZIP 下载
4. [功能详解](#功能详解)
5. [常见问题](#常见问题)
6. [技术支持](#技术支持)

---

## 🎯 简介

**Generator Platform（低代码代码生成平台）** 是一个基于 **Go + React** 的全栈 Web 应用，旨在帮助开发者快速生成后端应用程序代码。

### 核心价值

- ⚡ **效率提升**: 通过可视化配置，10分钟内生成完整后端代码
- 🔥 **开箱即用**: 一键启动，无需复杂配置
- 👥 **多用户支持**: 团队协作，权限管理
- 📊 **运维监控**: 实时掌握系统运行状态

---

## 💻 环境要求

### 必须安装的软件

| 软件 | 版本要求 | 用途 | 下载地址 |
|------|---------|------|---------|
| **Docker Desktop** | 4.0+ | 运行 PostgreSQL 数据库 | https://www.docker.com/products/docker-desktop |
| **Go (Golang)** | 1.21+ | 编译运行后端服务 | https://go.dev/dl/ |
| **Node.js** | 18+ 或 20+ | 运行前端开发服务器 | https://nodejs.org/ |

### 可选软件（推荐）

| 软件 | 用途 |
|------|------|
| VS Code | 代码编辑 |
| Postman | API 测试 |
| Git | 版本控制 |

---

## 🚀 快速开始

### ⬇️ 第一步：获取项目代码

#### 方式 A：从 GitHub 克隆（推荐）

打开终端（PowerShell / CMD / Git Bash），执行：

```bash
# 克隆仓库
git clone https://github.com/LeoNorth2004/server-generator-local.git

# 进入项目目录
cd server-generator-local
```

**如果没有安装 Git**：
1. 下载并安装 Git: https://git-scm.com/downloads
2. 安装时选择 **"Git from the command line and also from 3rd-party software"**
3. 重启终端后再次执行上述命令

#### 方式 B：下载 ZIP 包

1. 打开浏览器访问: **https://github.com/LeoNorth2004/server-generator-local**
2. 点击绿色的 **"Code"** 按钮 → 选择 **"Download ZIP"**
3. 解压到任意目录（例如 `C:\Projects\generator`）
4. 进入解压后的文件夹

---

### 方式一：一键启动（推荐）

#### 第二步：双击启动

```
📁 打开项目文件夹（克隆或解压后的目录）
🖱️ 双击文件: start-all.bat
```

#### 第二步：等待启动完成

启动脚本会自动执行以下操作：

```
[1/5] ✅ 检查/创建 PostgreSQL 数据库容器
[2/5] ✅ 编译并启动 Backend Service (:8080)
[3/5] ✅ 编译并启动 Generator Service (:8084)
[4/5] ✅ 安装依赖并启动前端 (:3000)
[5/5] ✅ 自动打开浏览器
```

**首次启动需要时间**：
- Go 编译：约 1-2 分钟
- npm install：约 1-3 分钟
- 后续启动：约 15-30 秒

#### 第三步：登录系统

浏览器会自动打开 `http://localhost:3000`，使用默认账户登录：

| 用户名 | 密码 | 角色 | 权限 |
|--------|------|------|------|
| `admin` | `admin123` | 管理员 | 全部功能 |
| 自定义 | 自定义 | 普通用户 | 只能查看自己的项目 |

---

### 方式二：手动启动（调试用）

如果需要单独启动某个服务进行调试：

```powershell
# 1️⃣ 启动数据库（如未运行）
docker start local-postgres

# 2️⃣ 启动 Backend Service（终端 1）
cd backend-service
$env:DB_PASSWORD="123456"
go build -o backend-service.exe .
.\backend-service.exe

# 3️⃣ 启动 Generator Service（终端 2）
cd generator-service
$env:DB_PASSWORD="123456"
go build -o generator-service.exe .
.\generator-service.exe

# 4️⃣ 启动前端（终端 3）
cd web-admin
npm run dev
```

---

### 如何停止服务

#### 一键停止

```
🖱️ 双击文件: stop.bat
```

停止脚本会安全关闭所有服务（PostgreSQL 保持运行以保留数据）。

#### 手动停止

```powershell
# 关闭对应的服务窗口即可
# 或者使用命令：
Stop-Process -Name backend-service -Force
Stop-Process -Name generator-service -Force
Stop-Process -Name node -Force
```

---

## 🎨 功能详解

### 1️⃣ 仪表盘首页

登录后首先看到的是仪表盘，展示系统概览：

| 卡片 | 说明 |
|------|------|
| **总项目数** | 当前系统中所有项目的总数 |
| **已生成** | 成功生成代码的项目数 |
| **待处理** | 还未生成的项目数 |

**操作入口**：
- 点击 **"+ 新建项目"** → 进入代码生成页面
- 左侧菜单可切换到其他功能模块

---

### 2️⃣ 代码生成（核心功能）

这是本系统的**核心功能**，用于根据数据库表结构自动生成 Go 后端代码。

#### 操作步骤

**Step 1: 进入代码生成器**

点击侧边栏 **"代码生成"** 或顶部 **"+ 新建项目"**

**Step 2: 配置基本信息**

```
┌─────────────────────────────────────┐
│ 项目名称: MyECommerce              │ ← 必填，例如"电商后台"
│                                     │
│ 数据库配置（可选）:                 │
│   Host: localhost                  │
│   Port: 5432                       │
│   User: postgres                   │
│   Password: 123456                 │
│   DB Name: myapp                   │
└─────────────────────────────────────┘
```

**Step 3: 添加数据表**

点击 **"+ 添加表"** 按钮，为每个表配置字段：

```
示例：users 表
┌───────────────────────────────────────────────────┐
│ 表名: users                                      │
├───────────────────────────────────────────────────┤
│ 字段名        类型          主键    允许空        │
│ ─────────────────────────────────────────────── │
│ id            int           ✓       ✗            │
│ username      varchar(100)  ✗       ✗            │
│ email         varchar(200)  ✗       ✓            │
│ created_at    timestamp     ✗       ✓            │
└───────────────────────────────────────────────────┘
```

**支持的类型**:
- `int`, `bigint` - 整数
- `varchar(n)` - 字符串
- `text` - 长文本
- `decimal(10,2)` - 小数
- `timestamp` / `datetime` - 时间戳
- `boolean` / `bool` - 布尔值
- `json` / `jsonb` - JSON 数据

**Step 4: 生成代码**

点击底部 **"🚀 生成代码"** 按钮，等待几秒后显示结果。

**Step 5: 查看和下载**

生成完成后可以：
- ✅ **预览代码**：点击左侧文件名查看具体内容
- ✅ **下载 ZIP 包**：点击 **"📥 下载代码"** 获取完整项目
- ✅ **保存到项目管理**：进入项目管理页面查看历史记录

**生成的代码结构**:

```
generated-project/
├── main.go                    # 入口文件
├── config/
│   └── config.go             # 数据库配置
├── internal/
│   ├── models/
│   │   ├── user.go           # User 模型定义
│   │   └── product.go        # Product 模型定义
│   ├── handler/
│   │   ├── user.go           # User CRUD Handler
│   │   └── product.go        # Product CRUD Handler
│   └── router/
│       └── router.go         # 路由注册
├── go.mod                    # Go 模块文件
└── .env                      # 环境变量
```

---

### 3️⃣ 项目管理

在项目管理页面可以：

#### 功能列表

| 功能 | 说明 |
|------|------|
| **查看列表** | 显示所有项目（管理员可见全部，普通用户只能看到自己的） |
| **查看创建者** | 每个项目显示是谁创建的（头像、用户名、角色） |
| **编辑项目** | 修改项目名称和描述 |
| **下载代码** | 下载已生成的代码包（ZIP 格式） |
| **删除项目** | 删除不需要的项目（仅管理员） |

#### 权限说明

- **管理员 (admin)**: 可以看到所有用户创建的项目，可以删除任何项目
- **普通用户 (user)**: 只能看到自己创建的项目，不能删除其他人的项目

---

### 4️⃣ 用户管理（仅管理员）

管理员可以在用户管理页面进行以下操作：

| 操作 | 说明 |
|------|------|
| **查看用户列表** | 显示所有注册用户 |
| **创建新用户** | 设置用户名、密码、角色（admin/user） |
| **编辑用户** | 修改用户信息或角色 |
| **删除用户** | 删除用户账户（不能删除自己） |

**角色说明**:

| 角色 | 权限 |
|------|------|
| **admin (管理员)** | 全部功能：用户管理、查看所有项目、删除项目 |
| **user (普通用户)** | 基础功能：代码生成、只看自己的项目、不能删除 |

---

### 5️⃣ 运维监控

实时监控系统运行状态：

#### 统计指标

| 指标 | 说明 | 更新方式 |
|------|------|---------|
| **总请求数** | 所有 API 调用次数 | 实时递增 |
| **注册用户数** | 当前系统用户总数 | 数据库实时查询 |
| **代码生成次数** | 累计生成次数 | 数据库实时查询 |

#### 最近事件

显示最近的操作日志，包括：
- 登录/登出
- 代码生成
- 文件下载
- 用户操作

每条记录包含：操作人、操作类型、资源、耗时、IP 地址等。

---

### 6️⃣ 文档中心

为选中的项目自动生成技术文档：

| 文档类型 | 内容 |
|---------|------|
| **API 接口文档** | 所有 CRUD 接口的详细说明（请求/响应格式） |
| **配置文档** | 环境变量说明、YAML 配置模板 |
| **二次开发指南** | 架构说明、代码规范、调试技巧 |

---

## ❓ 常见问题

### Q1: 启动时提示 "Docker 未运行"

**原因**: Docker Desktop 没有启动  
**解决**: 
1. 打开 Docker Desktop 应用
2. 等待托盘图标变为稳定状态（鲸鱼图标）
3. 重新运行 `start-all.bat`

---

### Q2: 提示 "端口已被占用"

**原因**: 该端口号被其他程序占用  
**解决**:
- **3000 端口**: 前端会自动切换到 3001
- **8080 端口**: 可能是之前的 Backend 没有完全关闭，运行 `stop.bat`
- **8084 端口**: 同上，运行 `stop.bat`

```powershell
# 手动查找占用端口的进程
netstat -ano | findstr ":8080"
# 记下最后一列的 PID，然后结束它
taskkill /PID <PID号> /F
```

---

### Q3: 代码生成失败

**可能的原因和解决方案**:

1. **Generator Service 未运行**
   ```
   解决: 运行 start-all.bat 重启服务
   ```

2. **数据库连接失败**
   ```
   检查 PostgreSQL 是否在运行:
   docker ps --filter "name=local-postgres"
   
   如果没有运行:
   docker start local-postgres
   ```

3. **表结构配置错误**
   ```
   确保:
   - 至少添加了 1 个表
   - 每个表至少有 1 个字段
   - 每个表有且仅有 1 个主键字段
   - 字段名不包含特殊字符
   ```

---

### Q4: 生成的代码在哪里？

**答案**: 
- **立即下载**: 在代码生成页面点击 **"📥 下载代码"**
- **稍后访问**: 进入 **"项目管理"** 页面，找到对应项目，点击下载按钮
- **数据库存储**: 所有项目信息都保存在 PostgreSQL 中，不会丢失

---

### Q5: 忘记密码怎么办？

**解决方案**:

1. **如果是 admin 账户**:
   ```sql
   -- 直接在数据库中重置密码
   docker exec -it local-postgres psql -U postgres -d generator_platform -c "UPDATE users SET password_hash='$2a$10$N9qo8LOig2JZSkR4A/ou3LlGmEKEJ7xOY' WHERE username='admin';"
   -- 新密码重置为: admin123
   ```

2. **如果是其他账户**: 让管理员在用户管理中重置密码，或者直接删除该账户重新注册

---

### Q6: 如何切换中文/英文界面？

**方法**: 在导航栏右上角找到语言切换下拉框，选择 **"中文"** 或 **"English"**

---

### Q7: 数据保存在哪里？

**答案**:
- **PostgreSQL 数据库**: 所有用户数据、项目数据、操作日志
- **Docker 卷**: 即使删除容器重建，数据也会保留（除非显式删除卷）
- **备份建议**: 定期导出数据库
   ```bash
   docker exec local-postgres pg_dump -U postgres generator_platform > backup.sql
   ```

---

### Q8: 如何部署到生产环境？

**基本步骤**:

1. **修改配置**:
   - 修改 `.env` 文件中的密码（不要使用默认的 `123456`）
   - 修改 JWT Secret 为随机字符串

2. **构建前端生产版本**:
   ```bash
   cd web-admin
   npm run build
   ```

3. **使用 Nginx 托管静态文件 + 反向代理**:
   ```nginx
   server {
       listen 80;
       
       location / {
           root /path/to/web-admin/dist;
           try_files $uri $uri/ /index.html;
       }
       
       location /api/ {
           proxy_pass http://localhost:8080;
       }
   }
   ```

4. **使用进程管理工具** (PM2/supervisor/systemd) 管理 Go 服务

详细部署文档请参考 `PROJECT_DOCUMENTATION.md`

---

## 🔧 高级用法

### 自定义代码模板

当前系统使用内置的 Go/Gin/GORM 模板。如果您想自定义输出格式：

1. 编辑 `generator-service/core.go` 中的生成函数
2. 重新编译: `go build -o generator-service.exe .`
3. 重启服务

### 多环境配置

创建不同的 `.env` 文件：

```bash
# .env.development (开发环境)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123456

# .env.production (生产环境)
DB_HOST=prod-db.example.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=secure_password_here
```

---

## 📊 系统架构图

```
┌─────────────────────────────────────────────────────┐
│                     浏览器                           │
│               React SPA (:3000)                      │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP/AJAX
                       ▼
┌─────────────────────────────────────────────────────┐
│              Backend Service (:8080)                 │
│  ┌──────────┬──────────┬──────────┬───────────────┐ │
│  │ 用户认证 │ 项目管理 │ 用户管理 │ 运维监控     │ │
│  └──────────┴──────────┴──────────┴───────────────┘ │
│                      │                               │
│              API Gateway (反向代理)                │
│         /api/v1/generator/* → :8084               │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│            Generator Service (:8084)                │
│  ┌──────────────┬──────────────┬─────────────────┐  │
│  │ 代码生成引擎 │ 元数据引擎   │ 文档生成器      │  │
│  └──────────────┴──────────────┴─────────────────┘  │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│            PostgreSQL Database (:5432)              │
│  ┌──────────┬──────────┬────────────────────────┐  │
│  │ users    │ projects │ operation_logs         │  │
│  └──────────┴──────────┴────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 🎓 学习资源

如果您想深入了解系统原理，请参考：

- **[PROJECT_DOCUMENTATION.md](./PROJECT_DOCUMENTATION.md)** - 技术文档（含答辩 Q&A）
- **Go 官方文档**: https://go.dev/doc/
- **React 官方文档**: https://react.dev/learn
- **GORM 文档**: https://gorm.io/docs/

---

## 📞 技术支持

遇到问题时，请按以下顺序排查：

### 1. 查看日志

**Backend Service 日志**（启动的终端窗口）
```
[GIN-debug] POST /api/v1/auth/login --> main.loginHandler
[Database] Connected to PostgreSQL at localhost:5432/generator_platform
```

**Generator Service 日志**
```
[Database] AutoMigration completed successfully
[Info] Metadata Engine initialized
```

**前端控制台**（浏览器 F12 → Console）
```
API Error: 401 Unauthorized
Failed to fetch projects
```

### 2. 常用诊断命令

```powershell
# 检查端口占用
netstat -ano | findstr ":8080 :8084 :3000 :5432"

# 检查进程
tasklist | findstr "backend generator node"

# 测试 API 连通性
curl http://localhost:8080/api/v1/operations/health
curl http://localhost:8084/health

# 查看 Docker 状态
docker ps
docker logs local-postgres --tail 20
```

### 3. 重启大法

90% 的问题可以通过重启解决：

```bash
# 1. 停止所有服务
stop.bat

# 2. 等待 5 秒

# 3. 重新启动
start-all.bat
```

---

## 📝 更新日志

### v1.0.0 (2026-06-02)

- ✅ 核心功能：代码生成、项目管理、用户管理
- ✅ 运维监控：实时统计、操作日志
- ✅ 多用户支持：RBAC 权限控制
- ✅ 国际化：中英文界面切换
- ✅ 一键启动/停止脚本
- ✅ 文档自动生成

---

## 📄 许可证

MIT License

Copyright (c) 2026 Generator Platform

---

## 🔗 项目仓库

- **GitHub 地址**: https://github.com/LeoNorth2004/server-generator-local
- **在线文档**: [USER_GUIDE.md](./USER_GUIDE.md) - 使用说明书
- **技术文档**: [PROJECT_DOCUMENTATION.md](./PROJECT_DOCUMENTATION.md) - 架构与答辩 Q&A
- **问题反馈**: 在 GitHub Issues 提交问题

---

> 💡 **提示**: 如需更多技术细节，请参阅同目录下的 **PROJECT_DOCUMENTATION.md** 文件。
