CREATE TABLE IF NOT EXISTS `user_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `user_id` VARCHAR(50) NOT NULL COMMENT '唯一标识',
   `name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '姓名',
   `identity_no` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '身份证号码',
   `age` INT(11) NOT NULL DEFAULT 0 COMMENT '年龄',
   `gender` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '性别 0:male 1:female',
   `phone_number` VARCHAR(50) NOT NULL COMMENT '手机号',
   `email` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '邮箱',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_email` (`email`),
   UNIQUE INDEX `unique_user_id` (`user_id`),
   UNIQUE INDEX `unique_phone_number` (`phone_number`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '用户表';

INSERT INTO `ant_test`.`user_tb`(`user_id`, `name`, `phone_number`) VALUES ('fsadfgsdfsadf', 'dfd', '342342');
