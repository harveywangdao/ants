CREATE TABLE IF NOT EXISTS `order_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `order_id` VARCHAR(50) NOT NULL COMMENT '订单唯一标识',
   `seller_id` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '商家ID',
   `buyer_id` VARCHAR(50) NOT NULL COMMENT '买家userID',
   `goods_id` VARCHAR(50) NOT NULL COMMENT '商品唯一标识',
   `goods_name` VARCHAR(100) NOT NULL COMMENT '商品名称',
   `count` INT(11) NOT NULL COMMENT '购买数量',
   `price` DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '价格,单位为分',
   `pay` DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '已支付金额,单位为分',
   `status` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '状态 0:未完成 1:完成',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_seller_id` (`seller_id`),
   INDEX `index_buyer_id` (`buyer_id`),
   INDEX `index_goods_id` (`goods_id`),
   UNIQUE INDEX `unique_order_id` (`order_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '订单表';

ALTER TABLE order_tb ADD COLUMN `count` INT(11) NOT NULL COMMENT '购买数量' AFTER `goods_name`;
ALTER TABLE order_tb MODIFY COLUMN `count` INT(11) NOT NULL COMMENT '购买数量';
ALTER TABLE order_tb DROP COLUMN `count`;
ALTER TABLE order_tb RENAME TO order_tb2;
ALTER TABLE order_tb COMMENT '修改表注释';
ALTER TABLE order_tb CHANGE `count` `count` INT(11) NOT NULL COMMENT '购买数量' AFTER `goods_name`;

ALTER TABLE order_tb ADD INDEX `index_goods_name` (`goods_name`);
