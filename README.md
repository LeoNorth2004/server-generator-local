# 🚀 Generator Platform - 低代码代码生成平台

<p align="center">
  <img src="https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=Modern%20web-based%20code%20generator%20platform%20dashboard%20with%20clean%20UI,%20dark%20theme,%20showing%20database%20tables,%20generated%20code%20preview,%20and%20project%20management%20interface&image_size=landscape_16_9" alt="Generator Platform" width="800"/>
</p>

<p align="center">
  <strong>基于 Go + React 的全栈低代码平台，通过可视化配置自动生成完整的后端应用程序代码</strong>
</p>

<p align="center">
  <a href="#-功能特性">功能特性</a> •
  <a href="#-技术栈">技术栈</a> •
  <a href="#-快速开始">快速开始</a> •
  <a href="#-项目结构">项目结构</a> •
  <a href="#-使用指南">使用指南</a> •
  <a href="#-文档">文档</a> •
  <a href="#-贡献">贡献</a>
</p>

---

## ✨ 功能特性

### 🔥 核心功能

- **📝 一键代码生成** - 通过可视化界面配置数据库表结构，自动生成 Go/Gin/GORM 后端代码
- **👥 多用户系统** - 支持管理员和普通用户角色，完善的 RBAC 权限控制
- **📊 实时运维监控** - 动态统计请求数、用户数、生成次数等关键指标
- **🌐 国际化支持** - 完整的中英文界面切换（i18next）
- **📄 自动文档生成** - 为项目生成 API 文档、配置指南、开发手册

### 🎯 适用场景

- 快速搭建企业级 CRUD 应用的后端服务
- 原型开发和 MVP 验证
- 学习 Go/Gin/GORM 全栈开发的示例项目
- 毕业设计 / 课程设计项目

---

## 🛠️ 技术栈

### 后端 (Backend)

| 技术 | 版本 | 说明 |
|------|------|------|
| **Go (Golang)** | 1.21+ | 高性能编程语言 |
| **Gin Web Framework** | v1.10.0 | 轻量级 HTTP 框架 |
| **GORM** | v1.25.x | ORM 数据库操作 |
| **PostgreSQL Driver** | v1.5.7 | 数据库驱动 |
| **JWT (golang-jwt)** | v5.2.1 | 用户认证 |

### 前端 (Frontend)

| 技术 | 版本 | 说明 |
|------|------|------|
| **React** | 19.x | UI 框架 |
| **Vite** | 8.x | 极速构建工具 |
| **Tailwind CSS** | 4.x | 原子化 CSS 框架 |
| **React Router** | 6.x | 路由管理 |
| **i18next** | 26.x | 国际化方案 |
| **Recharts** | 3.x | 数据可视化图表 |

### 基础设施

| 组件 | 说明 |
|------|------|
| **PostgreSQL 15+** | 关系型数据库（Docker 部署） |
| **Docker** | 容器化运行环境 |

---

## 🚀 快速开始

### ⬇️ 方式一：Git 克隆（推荐）

```bash
# 克隆仓库
git clone https://github.com/LeoNorth2004/server-generator-local.git

# 进入项目目录
cd server-generator-local

# 启动所有服务（Windows）
start-all.bat

# 或手动启动（Linux/Mac）
./start.sh
```

### 📦 方式二：下载 ZIP

1. 访问 [GitHub Releases](https://github.com/LeoNorth2004/server-generator-local/releases)
2. 下载最新版本的 ZIP 包
3. 解压后双击 `start-all.bat` 启动

### 🔧 环境要求

- ✅ **Docker Desktop** 4.0+ （用于 PostgreSQL）
- ✅ **Go** 1.21+ （编译后端）
- ✅ **Node.js** 18+ / 20+ （前端开发）

> 💡 首次启动会自动安装依赖并编译，请耐心等待 2-3 分钟。

---

## 🌐 访问地址

启动成功后，打开浏览器访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| **Web Admin** | http://localhost:3000 | 管理界面 |
| **Backend API** | http://localhost:8080 | 后端接口 |
| **Generator API** | http://localhost:8084 | 代码生成服务 |
| **PostgreSQL** | localhost:5432 | 数据库 |

### 👤 默认账户

| 角色 | 用户名 | 密码 |
|------|--------|------|
| 🔑 管理员 | `admin` | `admin123` |
| 👤 用户 | 自定义注册 | 自定义设置 |

---

## 📂 项目结构

```
server-generator-local/
│
├── 📄 start-all.bat              # Windows 一键启动脚本
├── 📄 stop.bat                   # Windows 一键停止脚本
├── 📄 README.md                  # 本文件 - 项目说明
├── 📄 USER_GUIDE.md             # 使用说明书（详细教程）
├── 📄 PROJECT_DOCUMENTATION.md   # 技术文档（架构 + 答辩 Q&A）
├── 📄 .gitignore                 # Git 忽略规则
│
├── 📂 backend-service/           # 后端主服务 (:8080)
│   ├── main.go                   # 入口 & 路由
│   ├── auth.go                   # 认证模块
│   ├── user.go                   # 用户管理
│   ├── project.go                # 项目管理
│   ├── operations.go             # 运维监控
│   └── .env                      # 环境变量
│
├── 📂 generator-service/         # 代码生成引擎 (:8084)
│   ├── main.go                   # 服务入口
│   ├── handlers.go               # HTTP 处理器
│   ├── core.go                   # 核心生成逻辑
│   ├── types.go                  # 类型定义
│   └── docs.go                   # 文档生成器
│
├── 📂 libs/go-common/            # 公共共享库
│   ├── config/config.go          # 配置加载
│   ├── database/database.go      # 数据库连接池
│   ├── middleware/middleware.go  # JWT 中间件
│   └── models/models.go          # 数据模型
│
└── 📂 web-admin/                 # React 前端 (:3000)
    ├── src/
    │   ├── pages/                # 页面组件
    │   │   ├── Home.jsx         # 仪表盘
    │   │   ├── Login.jsx        # 登录页
    │   │   ├── Generator.jsx    # 代码生成器 ⭐
    │   │   ├── Projects.jsx     # 项目管理
    │   │   ├── Users.jsx        # 用户管理
    │   │   ├── Operations.jsx   # 运维监控
    │   │   └── Docs.jsx         # 文档中心
    │   ├── components/           # 可复用组件
    │   ├── contexts/             # React Context
    │   └── i18n/locales/         # 中英文翻译
    ├── package.json
    └── vite.config.js            # Vite 配置
```

---

## 🎮 使用指南

### 1️⃣ 代码生成流程

```
登录 → 代码生成页面 → 配置表结构 → 点击"生成" → 预览代码 → 下载 ZIP
```

**简单三步走**：

```bash
# Step 1: 启动系统
start-all.bat

# Step 2: 打开浏览器访问 http://localhost:3000
# 使用 admin / admin123 登录

# Step 3: 进入"代码生成"，添加表，点击"生成"
```

### 2️⃣ 生成的代码示例

输入：
```json
{
  "tables": [
    {
      "name": "users",
      "fields": [
        { "name": "id", "type": "int", "primary": true },
        { "name": "username", "type": "varchar(100)" },
        { "name": "email", "type": "varchar(200)" }
      ]
    }
  ]
}
```

输出：完整的 Go/Gin/GORM 项目
```
generated-project/
├── main.go                    # 入口文件
├── config/config.go           # 数据库配置
├── internal/models/user.go    # User 模型
├── internal/handler/user.go   # CRUD Handler
└── internal/router/router.go  # 路由注册
```

### 3️⃣ 服务管理

```bash
# 启动所有服务
start-all.bat

# 停止所有服务（PostgreSQL 保持运行）
stop.bat

# 单独重启某个服务（调试用）
cd backend-service && go build -o backend-service.exe . && ./backend-service.exe
```

---

## 🏗️ 系统架构

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

## 📖 文档

| 文档 | 说明 | 适合人群 |
|------|------|---------|
| **[USER_GUIDE.md](./USER_GUIDE.md)** | 📘 详细使用说明书，包含安装、配置、常见问题 | 所有用户 |
| **[PROJECT_DOCUMENTATION.md](./PROJECT_DOCUMENTATION.md)** | 📗 技术架构文档 + 20道答辩 Q&A | 开发者 / 学生 |

---

## 🎓 核心亮点

### 1. 元数据驱动的代码生成
通过 JSON 配置描述数据表结构，模板引擎自动生成符合规范的 Go 代码，支持多种字段类型映射。

### 2. 企业级权限控制
完善的 JWT 认证 + RBAC 角色权限，管理员可查看所有用户的项目及创建者信息。

### 3. 实时运维监控
采用原子计数器 + 数据库实时查询的混合策略，提供动态统计数据。

### 4. 开发者友好
一键启动脚本、详细的错误提示、完善的文档体系、代码预览和下载。

### 5. 生产就绪
Docker 容器化部署、环境变量配置、数据持久化、日志分级输出。

---

## 🧪 测试验证

```powershell
# 测试代码生成 API
$token = (Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
  -Method POST -Body '{"username":"admin","password":"admin123"}' `
  -ContentType "application/json").data.token

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/generator/generate" `
  -Method POST -Body '{"project_name":"Test","tables":[{"name":"test","fields":[{"name":"id","type":"int","primary":true}]}]}' `
  -ContentType "application/json" `
  -Headers @{"Authorization"="Bearer $token"}
```

---

## 🐛 常见问题

<details>
<summary><b>❓ 启动失败？</b></summary>

1. 确保 Docker Desktop 已启动并运行
2. 检查端口 3000, 8080, 8084, 5432 是否被占用
3. 运行 `stop.bat` 清理残留进程后重试
</details>

<details>
<summary><b>❓ 代码生成失败？</b></summary>

1. 检查 Generator Service 是否在运行（端口 8084）
2. 确保至少添加了 1 个表且每表有 1 个主键字段
3. 查看 Backend 终端的错误日志
</details>

<details>
<summary><b>❓ 忘记密码？</b></summary>

在数据库中重置 admin 密码为 `admin123`:
```bash
docker exec local-postgres psql -U postgres -d generator_platform \
  -c "UPDATE users SET password_hash='\$2a\$10\$N9qo8LOig2JZSkR4A/ou3LlGmEKEJ7xOY' WHERE username='admin';"
```
</details>

更多问题请查阅 [USER_GUIDE.md](./USER_GUIDE.md) 的 FAQ 章节。

---

## 📈 路线图

- [x] v1.0.0 - 核心功能（代码生成、项目管理、用户管理、运维监控）
- [ ] v1.1.0 - 数据库逆向工程（从已有 DB 导入表结构）
- [ ] v1.2.0 - 多模板支持（Java Spring Boot, Python FastAPI）
- [ ] v1.3.0 - 在线 IDE 编辑器（Monaco Editor）
- [ ] v2.0.0 - AI 辅助生成（自然语言描述需求）

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

---

## 👨‍💻 作者

**LeoNorth2004**

- GitHub: [@LeoNorth2004](https://github.com/LeoNorth2004)

---

## ⭐ Star 支持

如果这个项目对您有帮助，欢迎给个 ⭐ 支持一下！

<p align="center">
  <strong>感谢使用 Generator Platform！如有问题欢迎提 Issue 🎉</strong>
</p>
