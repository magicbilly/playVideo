create database if not exists video;
use video;
CREATE TABLE if not exists `video` (
                         `Id`           INT AUTO_INCREMENT PRIMARY KEY,
    -- 核心字段：存储文件名的 MD5 或缩减后的唯一字符串
    -- 用于：硬盘文件夹名、URL 路由、防止路径超长
                         `FileHash`    VARCHAR(64) NOT NULL UNIQUE,

    -- 逻辑字段：用户看到的原始长文件名（带特殊字符）
    -- 用于：页面标题显示
                         `Title`        TEXT NOT NULL,
                         `Path`     TEXT NOT NULL,
    -- 海报图地址
                         `Poster`       VARCHAR(255) DEFAULT '',

    -- 视频状态：0-未处理, 1-切片中, 2-已完成, 3-失败
                         `status`       TINYINT DEFAULT 0,

                         `created_at`   DATETIME DEFAULT CURRENT_TIMESTAMP,
                         `updated_at`   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- 为 Hash 建立索引，极速查询
                         INDEX `idx_hash` (`FileHash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
drop table user
use video_test;
create table if not exists `user`(
    `id` char(10) primary key,
    `name` varchar(20) not null,
    `role` varchar(10) default 'visitor',
    `password` char(32) not null ,

    `created_at`   DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)