CREATE TABLE IF NOT EXISTS `goods_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `goods_id` VARCHAR(50) NOT NULL COMMENT '商品唯一标识',
   `seller_id` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '商家ID',
   `goods_name` VARCHAR(100) NOT NULL COMMENT '商品名称',
   `price` DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '价格,单位为分',
   `category` INT(11) NOT NULL DEFAULT 0 COMMENT '商品种类',
   `stock` INT(11) NOT NULL DEFAULT 0 COMMENT '商品库存',
   `brand` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '品牌',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_goods_name` (`goods_name`),
   UNIQUE INDEX `unique_goods_id` (`goods_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '商品表';

CREATE TABLE IF NOT EXISTS `purchase_record_tb`(
   `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `goods_id` VARCHAR(50) NOT NULL COMMENT '商品唯一标识',
   `order_id` VARCHAR(50) NOT NULL COMMENT '订单唯一标识',
   `pay_id` VARCHAR(50) NOT NULL COMMENT '支付号',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_goods_id` (`goods_id`),
   UNIQUE INDEX `unique_order_id` (`order_id`),
   UNIQUE INDEX `unique_pay_id` (`pay_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '购买记录表';