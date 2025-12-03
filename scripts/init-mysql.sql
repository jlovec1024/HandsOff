-- HandsOff MySQL 初始化脚本
-- 该脚本在 MySQL 容器首次启动时自动执行

-- 设置字符集
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS handsoff
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE handsoff;

-- 授予权限
GRANT ALL PRIVILEGES ON handsoff.* TO 'handsoff'@'%';
FLUSH PRIVILEGES;

-- 输出初始化信息
SELECT 'HandsOff database initialized successfully!' AS message;
