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