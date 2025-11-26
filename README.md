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

## 项目结构

```
db_front/
├── front/              # React 前端
│   ├── src/
│   │   ├── App.jsx    # 主应用组件
│   │   ├── api.js     # API 调用封装
│   │   └── main.jsx   # 入口文件
│   └── package.json
├── back/               # Go 后端
│   ├── main.go        # 主服务器和API路由
│   ├── email.go       # 邮件发送功能
│   └── go.mod
├── database/           # MySQL 数据库
│   ├── db_schema.sql  # 数据库结构
│   └── init_data.sql  # 示例数据
├── uploads/            # 文件存储目录
└── docker-compose.yml  # Docker 编排配置
```

## 快速开始

### 前置要求
- Docker & Docker Compose
- Node.js 18+ (本地开发)
- Go 1.21+ (本地开发)

### 使用 Docker 运行（推荐）

1. 克隆项目并进入目录：
```bash
cd db_front
```

2. 启动所有服务：
```bash
docker-compose up -d
```

3. 访问应用：
   - 前端: http://localhost:3000
   - 后端API: http://localhost:8080/api

4. 查看日志：
```bash
docker-compose logs -f
```

5. 停止服务：
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

#### 后端开发
```bash
cd back
go mod tidy
go run .
```

#### 数据库
需要本地运行 MySQL 8.0，并执行 `database/db_schema.sql` 初始化数据库。

## API 接口

详细的API文档请参考 [API_DESIGN.md](./API_DESIGN.md)

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

### 环境变量

#### 后端 (back/)
- `DB_HOST` - MySQL 主机地址（默认: localhost）
- `DB_USER` - MySQL 用户名（默认: root）
- `DB_PASSWORD` - MySQL 密码（默认: root）
- `DB_NAME` - 数据库名（默认: db_front）
- `SMTP_HOST` - 邮件服务器地址
- `SMTP_PORT` - 邮件服务器端口（默认: 587）
- `SENDER_EMAIL` - 发件人邮箱
- `SENDER_PASS` - 发件人邮箱密码

#### 前端代理配置
前端通过 Vite 代理转发 `/api` 请求到后端（见 `front/vite.config.js`）

## 数据库设计

详细的数据库设计请参考 [DB_README.md](./DB_README.md)

### 核心表
- `departments` - 系别信息
- `teachers` - 教师信息
- `projects` - 项目信息
- `project_members` - 项目成员关系
- `replies` - 邮件回复记录
- `attachments` - 附件元数据
- `excel_a_rows` - Excel格式A数据
- `excel_b_rows` - Excel格式B数据

## 待完善功能

- [ ] 实际的SMTP邮件发送功能（目前仅为占位实现）
- [ ] IMAP邮件接收和解析功能
- [ ] Excel文件解析和合并功能
- [ ] 用户认证和权限管理
- [ ] 数据导出为多种格式（PDF、CSV等）
- [ ] 邮件发送队列和重试机制
- [ ] 数据统计和可视化报表

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！
