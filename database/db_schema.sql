-- MySQL schema for 教师/项目/邮件/附件
SET
    NAMES utf8mb4;

SET
    CHARACTER
SET
    utf8mb4;

SET
    FOREIGN_KEY_CHECKS = 0;

-- Departments
DROP TABLE IF EXISTS departments;

CREATE TABLE
    departments (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        code VARCHAR(50),
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- Teachers basic info
DROP TABLE IF EXISTS teachers;

CREATE TABLE
    teachers (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(255) NOT NULL UNIQUE,
        department_id INT,
        phone VARCHAR(50),
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (department_id) REFERENCES departments (id) ON DELETE SET NULL
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- Projects (每个汇总为一个 project)
DROP TABLE IF EXISTS projects;

CREATE TABLE
    projects (
        id INT AUTO_INCREMENT PRIMARY KEY,
        code VARCHAR(100) NOT NULL UNIQUE, -- 例如: 2025_WORKLOAD
        name VARCHAR(255) NOT NULL, -- 人可读名称: 2025年度工作量汇总
        status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, archived
        email_subject_template VARCHAR(255),
        email_body_template TEXT,
        excel_template_filename VARCHAR(255), -- 存储在 file storage 下的模板文件名
        created_by INT, -- 管理员 user id, 可为空
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- 项目成员/收件人列表（指定哪些教师属于该项目）
DROP TABLE IF EXISTS project_members;

CREATE TABLE
    project_members (
        id INT AUTO_INCREMENT PRIMARY KEY,
        project_id INT NOT NULL,
        teacher_id INT NOT NULL,
        role VARCHAR(50) DEFAULT 'recipient', -- recipient | manager
        invited_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        sent_at DATETIME, -- 邮件发送时间
        current_status VARCHAR(50) DEFAULT 'pending', -- pending | replied | ignored
        last_reply_at DATETIME,
        UNIQUE KEY uq_project_teacher (project_id, teacher_id),
        FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
        FOREIGN KEY (teacher_id) REFERENCES teachers (id) ON DELETE CASCADE
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- Dispatch 记录（每次将邮件发送到一组人都记录一条 dispatch）
DROP TABLE IF EXISTS dispatches;

CREATE TABLE
    dispatches (
        id INT AUTO_INCREMENT PRIMARY KEY,
        project_id INT NOT NULL,
        dispatched_by INT,
        target_type VARCHAR(50) NOT NULL, -- all | department | selected
        target_detail JSON, -- 例如 {"department":"计算机系"} 或 {"teacher_ids":[1,2,3]}
        sent_count INT,
        dispatched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- Replies: 从邮件服务器解析到的一封回复（按邮件 message-id）
DROP TABLE IF EXISTS replies;

CREATE TABLE
    replies (
        id INT AUTO_INCREMENT PRIMARY KEY,
        project_id INT NOT NULL,
        teacher_id INT, -- 若无法匹配到教师，可为空并把 from_email 存下
        from_email VARCHAR(255) NOT NULL,
        subject VARCHAR(255),
        message_id VARCHAR(255) UNIQUE, -- 邮件服务器的 message-id
        in_reply_to VARCHAR(255),
        received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        raw_headers JSON,
        raw_body LONGTEXT,
        FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
        FOREIGN KEY (teacher_id) REFERENCES teachers (id)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- Attachments metadata: 存放在本地/云存储的文件索引
DROP TABLE IF EXISTS attachments;

CREATE TABLE
    attachments (
        id INT AUTO_INCREMENT PRIMARY KEY,
        reply_id INT,
        project_id INT NOT NULL,
        teacher_id INT,
        original_filename VARCHAR(255),
        stored_path VARCHAR(500) NOT NULL, -- 本地或云存储路径
        content_type VARCHAR(100),
        file_size INT,
        excel_type VARCHAR(50), -- type_a | type_b | unknown
        parsed BOOLEAN DEFAULT FALSE,
        parsed_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (reply_id) REFERENCES replies (id) ON DELETE CASCADE,
        FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
        FOREIGN KEY (teacher_id) REFERENCES teachers (id)
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

-- 索引建议
CREATE INDEX idx_teachers_email ON teachers (email);

CREATE INDEX idx_project_members_project ON project_members (project_id);

CREATE INDEX idx_replies_project ON replies (project_id);

CREATE INDEX idx_attachments_project ON attachments (project_id);

SET
    FOREIGN_KEY_CHECKS = 1;