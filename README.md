# 帝王世系图谱 (Dynastic Genealogy)

一个基于 **Go + MySQL + React** 的帝王世系图谱应用，支持「家谱树」与「时间轴」两种可视化视图，
内置秦、西汉、东汉、唐、宋、明、清等主要朝代的帝王世系示例数据。

## 技术栈

| 层次 | 技术 |
| --- | --- |
| 后端 | Go 1.23（标准库 `net/http`，无重型框架） |
| 数据库 | MySQL 8（`go-sql-driver/mysql`） |
| 前端 | React 18 + Vite 5 + TypeScript |
| 图谱可视化 | React Flow（`@xyflow/react`）家谱树 + 自研横向时间轴 |

## 目录结构

```
Dynastic/
├─ backend/                     Go 后端
│  ├─ cmd/server/main.go        程序入口（优雅启停）
│  ├─ internal/
│  │  ├─ config/                配置加载 + 轻量 .env 解析
│  │  ├─ database/              连接、自动建库、内嵌迁移
│  │  │  └─ migrations/         001_schema.sql / 002_seed.sql
│  │  ├─ model/                 数据模型
│  │  ├─ repository/            数据访问 + 世系树构建
│  │  ├─ handler/               REST API 路由与处理
│  │  └─ middleware/            CORS
│  ├─ .env.example / .env       环境变量
│  └─ go.mod
└─ frontend/                    React 前端
   ├─ src/
   │  ├─ api/client.ts          后端接口封装
   │  ├─ components/            家谱树 / 时间轴 / 详情面板 / 朝代选择
   │  ├─ types.ts               类型定义与年份格式化
   │  ├─ App.tsx                主界面与状态管理
   │  └─ main.tsx
   ├─ vite.config.ts            开发服务器 + /api 代理
   └─ package.json
```

## 快速开始

### 0. 前置条件
- 已安装 Go 1.23+、Node 18+、MySQL 8，并已启动 MySQL 服务。

### 1. 配置数据库口令
编辑 `backend/.env`，把 `DB_PASSWORD=` 改成你的 MySQL root 密码
（若使用非 root 账号，一并修改 `DB_USER`）。程序首次启动会自动创建 `dynastic` 库、建表并写入种子数据。

### 2. 启动后端
```powershell
cd backend
go run ./cmd/server
```
成功后日志显示：`帝王世系图谱后端已启动，监听 http://localhost:8080`

### 3. 启动前端
```powershell
cd frontend
npm install
npm run dev
```
浏览器打开 http://localhost:5173 。Vite 已配置将 `/api` 代理到 `http://localhost:8080`，无需额外处理跨域。

## API 一览

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/health` | 健康检查 |
| GET | `/api/dynasties` | 朝代列表（含帝王数量） |
| GET | `/api/dynasties/{id}` | 朝代详情 + 全部帝王 |
| GET | `/api/dynasties/{id}/tree` | 该朝代的世系森林（父子树结构） |
| GET | `/api/emperors?dynastyId={id}` | 帝王列表（可按朝代过滤） |
| GET | `/api/emperors/{id}` | 单个帝王详情 |

## 数据模型

- **dynasty（朝代）**：名称、起止年份、都城、简介。
- **emperor（帝王）**：姓名、庙号、谥号、年号、`father_id`（自引用父帝）、在位顺序、
  继位关系说明、在位起止年、生卒年、简介。

年份统一用整数表示，**负数代表公元前**（如 `-221` 表示公元前 221 年）。

## 功能特性

- **家谱树视图**：按 `father_id` 自动构建世系森林并布局，支持缩放、拖拽、小地图，点击节点看详情。
  对于养子继位、隔代继承等「父节点不在同朝代数据中」的帝王，会作为独立的根节点展示，保证不遗漏。
- **时间轴视图**：按在位先后横向排列，卡片上下交错，直观呈现传承顺序与在位年代。
- **详情面板**：展示庙号、谥号、年号、在位与生卒年份、继位关系与简介。

## 生产构建

```powershell
cd frontend
npm run build   # 产物在 frontend/dist
```
可将 `dist` 交由任意静态服务器托管，并通过 `frontend/.env` 的 `VITE_API_BASE` 指向后端地址。
