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
   `email` VARCHAR(100) NOT NULL COMMENT '邮箱',
   `password` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '密码',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   UNIQUE INDEX `unique_email` (`email`),
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
   `passphrase` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '交易所密码',
   `exchange` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '交易所',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_user_id` (`user_id`),
   UNIQUE INDEX `unique_api_key` (`api_key`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT 'apikey表';

CREATE TABLE IF NOT EXISTS `strategy_tb`(
   `id` INT(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `strategy` VARCHAR(100) NOT NULL COMMENT '策略',
   `desc` VARCHAR(4096) NOT NULL DEFAULT '' COMMENT '策略描述',
   `param` VARCHAR(4096) NOT NULL DEFAULT '' COMMENT '策略字段/参数(前端来读写)',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   UNIQUE INDEX `unique_strategy` (`strategy`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '策略表';

CREATE TABLE IF NOT EXISTS `strategy_template_tb`(
   `id` INT(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `strategy_id` INT(11) NOT NULL COMMENT '策略id',
   `template_name` VARCHAR(128) NOT NULL COMMENT '策略模板名称',
   `param` VARCHAR(4096) NOT NULL DEFAULT '' COMMENT '策略参数值(前端来读写)',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `strategy_id` (`strategy_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '策略模板表';

CREATE TABLE `asset_tb` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键', 
  `exchange` varchar(32) NOT NULL COMMENT '交易所名称',
  `api_key` varchar(128) NOT NULL COMMENT '交易所账号',
  `initial_rights` double NOT NULL COMMENT '初始权益',
  `account_rights` double NOT NULL COMMENT '账户权益',
  `frozen_rights` double DEFAULT NULL COMMENT '冻结权益',
  `drawdown_ratio` double DEFAULT NULL COMMENT '最大回撤',
  `insert_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `unique_apikey` (`apikey`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='资金表';

CREATE TABLE `asset_snapshot_tb` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键', 
  `exchange` varchar(32) NOT NULL COMMENT '交易所名称',
  `api_key` varchar(128) NOT NULL COMMENT '交易所账号',
  `initial_rights` double NOT NULL COMMENT '初始权益',
  `account_rights` double NOT NULL COMMENT '账户权益',
  `frozen_rights` double DEFAULT NULL COMMENT '冻结权益',
  `drawdown_ratio` double DEFAULT NULL COMMENT '最大回撤',
  `insert_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='资金快照表';
