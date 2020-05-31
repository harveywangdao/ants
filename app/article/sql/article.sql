CREATE TABLE IF NOT EXISTS `article_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `article_id` VARCHAR(50) NOT NULL COMMENT '文章唯一标识',
   `user_id` VARCHAR(50) NOT NULL COMMENT '用户唯一标识',
   `title` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '标题',
   `content` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '内容',
   `tags` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '标签',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   UNIQUE INDEX `unique_article_id` (`article_id`),
   INDEX `index_user_id` (`user_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '文章表';

mysql -u root -p
use ant_test;
DROP TABLE article_tb;
TRUNCATE TABLE article_tb;
