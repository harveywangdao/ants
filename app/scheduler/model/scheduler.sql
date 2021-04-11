CREATE DATABASE skydb;
USE skydb;

CREATE TABLE IF NOT EXISTS `user_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `user_id` VARCHAR(50) NOT NULL COMMENT '唯一标识',
   `open_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '小程序openid',
   `union_id` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '小程序unionid',
   `session_key` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '小程序session_key',
   `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '姓名',
   `identity_no` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '身份证号码',
   `age` INT(11) NOT NULL DEFAULT 0 COMMENT '年龄',
   `gender` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '性别 0:male 1:female',
   `phone_number` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '手机号',
   `email` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '邮箱',
   `password` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '密码',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_email` (`email`),
   INDEX `index_phone_number` (`phone_number`),
   INDEX `index_open_id` (`open_id`),
   INDEX `index_union_id` (`union_id`),
   UNIQUE INDEX `unique_user_id` (`user_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '用户表';

CREATE TABLE IF NOT EXISTS `apikey_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `user_id` VARCHAR(50) NOT NULL COMMENT '用户唯一标识',
   `api_key` VARCHAR(500) NOT NULL COMMENT '交易所账号',
   `secret_key` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '交易所密钥',
   `exchange` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '交易所',
   `strategy` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '策略',
   `rate` DECIMAL(40,20) NOT NULL DEFAULT 0 COMMENT '费率',
   `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否删除 1:run 2:pause 3:stop',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_user_id` (`user_id`),
   UNIQUE INDEX `unique_api_key` (`api_key`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT 'apikey表';

CREATE TABLE IF NOT EXISTS `strategy_tb`(
   `id` INT(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `strategy` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '策略',
   `param` DECIMAL(40,20) NOT NULL DEFAULT 0 COMMENT '参数',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   UNIQUE INDEX `unique_strategy` (`strategy`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '策略表';
