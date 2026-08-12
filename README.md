# ABSC - Automatic Bangumi Subscribe Center
> 基于 Go + React + SQLite 构建的全自动化动漫订阅、智能重命名与跨季路径整理系统。

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://react.dev/)
[![Docker Build](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 核心特性

* **当季新番动态发现**：自动聚合 Mikan 蜜柑计划 与 TMDB 元数据，提供精致的海报墙、剧情大纲与周播周期展示。
* **字幕组懒加载与一键订阅**：实时抓取 Mikan 各字幕组 RSS 资源与历史文件列表，支持包含/排除关键词过滤，一键直接下发规则至 qBittorrent。
* **智能改名与跨季自动路径纠偏**：
  * 后台 Cron 守护进程轮询 qBittorrent 下载任务。
  * 自动基于多季集数偏移量（Offset）将绝对集数换算为标准相对集数（公式：$Episode_{relative} = Episode_{absolute} - Offset$）。
  * 识别跨季（如 Season 2）后，自动调用 qB API 将物理文件夹从 `Season 1` 迁移至 `Season 2` 并重新命名。
* **改名实时效果预览**：前端支持点击字幕组历史文件，贴身悬浮即时显示改名后效果（如 `[Group] Title - 13.mp4` $\rightarrow$ `E01.mp4`）。
* **开箱即用单容器部署**：采用 Docker 多阶段构建，Go 原生托管编译后的 React 静态应用，配合 `sqlite-web` 实现数据图形化维护。

---

## 技术栈

* **后端 (Backend)**：Go 1.22 / Gin Framework / GORM (SQLite3) / GoQuery (HTML Scraping) / qBittorrent Web API / TMDB API Client
* **前端 (Frontend)**：React 18 / TypeScript / Vite 5 / Tailwind CSS / Lucide React / Axios
* **部署 (Deployment)**：Docker (Multi-Stage Build) / Docker Compose / Alpine Linux

---

## 快速部署 (Docker Compose)

推荐使用 Docker Compose 进行轻量一键部署。

### 1. 编写 `docker-compose.yml`

在服务器或虚拟机上创建项目目录并编写 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  ABSC:
    build:
      context: .
      dockerfile: Dockerfile
    image: absc
    container_name: ABSC
    restart: always
    ports:
      - "8899:8899"
    environment:
      - PORT=8899
      - ABSC_DB_PATH=/app/data/absc.db
      - QB_URL=http://192.168.1.1:8080        # 替换为你的 qBittorrent 地址
      - QB_USER=admin                         # 替换为你的 qB 账号
      - QB_PASS=adminadmin                    # 替换为你的 qB 密码
      - SERIES_DIR=/downloads/Series          # 替换为你在qbittorrent中的剧集存放目录
      - INCOMPLETE_DIR=/downloads/incomplete  # 替换为你在qbittorrent中的剧集存放目录
      - TMDB_API_KEY=your_tmdb_api_key_here   # 填入你的 TMDB API Key
      #- HTTP_PROXY=                           # 软路由代理地址 (选填)
    volumes:
      # 将虚拟机的本地目录挂载进容器，确保数据库永久持久化！
      - ./data:/app/data

  # SQLite 网页版可视控制台
  sqlite-web:
    image: coleifer/sqlite-web
    container_name: absc-sqlite-web
    restart: always
    ports:
      - "8089:8080"
    volumes:
      - ./data:/data
    environment:
      - SQLITE_DATABASE=absc.db
```

### 2. 启动容器

```bash
# 启动服务
docker compose up -d

# 查看运行日志
docker compose logs -f anime-manager

```

启动完成后，访问 `http://<服务器IP>:8899` 即可进入控制中心；访问 `http://<服务器IP>:8089` 可打开数据库管理后台。

---

## 后续计划：
1. 后端：重命名种子文件功能，在数据库中没有对应番剧时，自动按默认集数偏移（0）进行重命名
2. 后端：蜜柑计划首页的剧场版内容，实现识别
3. 前端：页面美化

## ⚙️ 环境变量说明

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | Web 服务监听端口 |
| `ABSC_DB_PATH` | `/app/data/absc.db` | SQLite 数据库存储绝对路径 |
| `QB_URL` | `http://localhost:8080` | qBittorrent 地址 (包含端口) |
| `QB_USER` | `admin` | qBittorrent 用户名 |
| `QB_PASS` | `admin` | qBittorrent 密码 |
| `SERIES_DIR` | `/downloads/Series` | 整理完成后多季存储根目录 |
| `INCOMPLETE_DIR` | `/downloads/incomplete` | 下载未完成临时目录 |
| `TMDB_API_KEY` | - | TMDB API 密钥 (用于抓取海报及元数据) |
| `HTTP_PROXY` | - | HTTP/HTTPS 代理地址 (选填，加速 TMDB / Mikan 抓取) |

---

## 本地开发环境配置

如果你需要对本项目进行二次开发或调试，请按以下步骤启动：

### 前置条件

* Go 1.22+
* Node.js 18+ & npm
* GCC 编译器 (Windows 需安装 MinGW-w64，用于 CGO 编译 SQLite3)

### 1. 后端启动 (Backend)

```bash
cd backend

# 安装依赖
go mod download

# 运行主进程 (运行前可在 main.go 中修改默认配置)
go run .

```

### 2. 前端启动 (Frontend)

```bash
cd frontend

# 安装依赖
npm install

# 启动 Vite 本地热更新服务
npm run dev

```

打开浏览器访问 `http://localhost:5173` 即可进行前端实时预览。

---

## 📁 项目结构概要

```text
.
├── backend/                  # Go 后端工程
│   ├── internal/
│   │   ├── client/           # qBittorrent Web API 客户端
│   │   ├── database/         # SQLite 连接与初始 Seeds
│   │   ├── handler/          # Gin RESTful API 控制层
│   │   ├── model/            # GORM 数据库结构体定义
│   │   ├── router/           # 路由挂载与静态资源托管
│   │   ├── scraper/          # Mikan / TMDB 爬虫与刮削器
│   │   └── service/          # 订阅、改名、跨季迁移核心业务逻辑
│   └── main.go               # 主程序入口与 Cron 守护协程
├── frontend/                 # React + TypeScript 前端工程
│   ├── src/
│   │   ├── api/              # Axios 统一 API 模组
│   │   ├── components/       # UI 组件 (海报墙/弹窗/导航栏)
│   │   ├── types/            # TypeScript 接口类型定义
│   │   ├── App.tsx           # 主界面状态响应
│   │   └── index.css         # Tailwind CSS 全局样式
│   ├── vite.config.ts        # Vite 反向代理配置
│   └── tailwind.config.js    # Tailwind 配置文件
├── Dockerfile                # 多阶段打包镜像构建文件
├── docker-compose.yml        # Docker 服务编排文件
└── .dockerignore             # Docker 构建忽略规则

```

---

## 📄 开源许可证

本项目基于 [MIT License](https://www.google.com/search?q=LICENSE) 许可证开源，欢迎随意 fork 和贡献代码！
