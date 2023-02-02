CREATE DATABASE `mtop` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

use `mtop`;

CREATE TABLE `mtop_user`(
    `username` VARCHAR(128) NOT NULL,
    `salt` VARCHAR(32) NOT NULL,
    `password` VARCHAR(1024) NOT NULL DEFAULT '',
    `created_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(`username`)
);