SET
    NAMES utf8mb4;

SET
    CHARACTER
SET
    utf8mb4;

-- Insert sample departments
INSERT INTO
    departments (name, code)
VALUES
    ('计算机系', 'CS'),
    ('数学系', 'MATH'),
    ('物理系', 'PHY'),
    ('化学系', 'CHEM');

-- Insert sample teachers
INSERT INTO
    teachers (name, email, department_id, phone)
VALUES
    ('张伟', 'zhangwei@univ.edu.cn', 1, '13800138001'),
    ('李娜', 'lina@univ.edu.cn', 2, '13800138002'),
    ('王强', 'wangqiang@univ.edu.cn', 3, '13800138003'),
    ('赵敏', 'zhaomin@univ.edu.cn', 1, '13800138004'),
    ('刘洋', 'liuyang@univ.edu.cn', 4, '13800138005');

-- Insert sample users
INSERT INTO
    users (
        username, password, 
        smtp_host, smtp_port, smtp_username, smtp_password, 
        imap_host, imap_port, imap_username, imap_password, email_address
    )
VALUES
    (
        'admin', '$2a$10$l570xRhUYWPIHUShwUup5.Wqfkgs6NawDzn34zA3eRwCZlVP8uhpC', 
        'smtp.163.com', '587', '19857338587@163.com', 'WUb4GAbRrYp25tK6', 
        'imap.163.com', '993', '19857338587@163.com', 'WUb4GAbRrYp25tK6', '19857338587@163.com'
    );
