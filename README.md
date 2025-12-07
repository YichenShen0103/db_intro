# 学院科研数据汇总系统

一个用于学院科研秘书自动化收集和汇总教师数据的系统。

## 功能特性

1. **基础信息管理**
   - 维护教师信息（姓名、系别、邮箱）
   - 管理系别信息

2. **项目制管理**
   - 创建多个独立的汇总项目（如"2025年度工作量"、"2024课题审批"）
   - 为每个项目配置独立的邮件模板和Excel模板
   - 按项目分别管理邮件收发

3. **邮件发送**
   - 支持向全体教师、指定系别或指定教师发送邮件
   - 自动附加Excel收集表格
   - 可配置邮件主题和正文模板

4. **回复监控**
   - 实时查看各教师的回复状态
   - 自动识别已回复和未回复的教师
   - 支持一键催办未回复教师

5. **数据汇总**
   - 自动从邮件中提取Excel附件
   - 合并多个Excel文件为一个总表
   - 支持两种不同格式的Excel模板（A格式：工作量类，B格式：项目申报类）

## 技术栈

### 前端
- React 18
- Tailwind CSS
- Axios
- Vite

### 后端
- Go 1.21
- Gin Web Framework
- MySQL 8.0

### 部署
- Docker & Docker Compose

## 架构说明

本项目采用前后端分离架构，容器化部署：

1.  **Nginx 容器 (`nginx`)**:
    *   基于 `nginx:stable-alpine`。
    *   采用多阶段构建：先在 Node.js 环境中构建 React 前端，生成静态文件，然后复制到 Nginx 容器中。
    *   负责提供前端静态页面服务。
    *   负责将 `/api` 开头的请求反向代理到后端容器。
    *   资源占用极低（仅需 ~10MB 内存）。

2.  **后端容器 (`backend`)**:
    *   基于 Go 1.21。
    *   提供 RESTful API。
    *   处理邮件收发、Excel 解析等业务逻辑。
    *   挂载宿主机 `uploads` 目录实现文件持久化。

3.  **数据库容器 (`database`)**:
    *   基于 MySQL 8.0。
    *   数据持久化存储。

## 快速开始

### 前置要求
- Docker & Docker Compose
- Node.js 18+ (仅本地开发需要)
- Go 1.21+ (仅本地开发需要)

### 使用 Docker 运行（推荐）

1. 克隆项目并进入目录：
```bash
git clone <repository-url>
cd db_intro
```

2. 配置环境变量：
   复制 `.env.example` (如果有) 或直接创建 `.env` 文件，填入数据库和邮件服务器配置。

3. 启动所有服务：
```bash
docker-compose up -d --build
```

4. 访问应用：
   - 前端页面: http://localhost
   - 后端API: http://localhost/api/ping

5. 查看日志：
```bash
docker-compose logs -f
```

6. 停止服务：
```bash
docker-compose down
```

### 本地开发

#### 前端开发
```bash
cd front
npm install
npm run dev
```
*注意：本地开发时，Vite 会代理 `/api` 请求到 `http://localhost:8080`。*

#### 后端开发
```bash
cd back
# 确保本地有运行中的 MySQL，并配置好环境变量
go mod tidy
go run .
```

#### 数据库
需要本地运行 MySQL 8.0，并执行 `database/db_schema.sql` 初始化数据库。

## API 接口

### 主要接口

- `GET /api/projects` - 获取项目列表
- `POST /api/projects` - 创建新项目
- `POST /api/projects/:id/dispatch` - 发送邮件
- `GET /api/projects/:id/tracking` - 获取回复状态
- `POST /api/projects/:id/remind` - 催办未回复
- `POST /api/projects/:id/aggregate` - 汇总数据
- `GET /api/teachers` - 获取教师列表
- `POST /api/teachers` - 添加教师

## 配置说明

### 环境变量 (.env)

```properties
# 数据库配置
DB_NAME=db_intro
DB_USER=root
DB_PASSWORD=root
DB_HOST=database  # Docker 内部服务名，本地开发请用 localhost

# 系统配置
TZ=Asia/Shanghai  # 时区设置
```

### 数据持久化

- **数据库数据**: 存储在 Docker Volume `db_data` 中。
- **上传文件**: 映射到宿主机的 `./uploads` 目录，容器重启不会丢失。
- **日志**: Nginx 日志存储在 Docker Volume `nginx-logs` 中。

## 数据库设计

### 核心表
- `departments` - 系别信息
- `teachers` - 教师信息
- `projects` - 项目信息
- `project_members` - 项目成员关系
- `dispatches` - 邮件发送记录
- `replies` - 邮件回复记录
- `attachments` - 附件元数据

## 待完善功能

Nothing to be done...
