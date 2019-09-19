CREATE TABLE IF NOT EXISTS `student`(
   `id` BIGINT(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
   `student_no` VARCHAR(100) NOT NULL COMMENT '唯一标识',
   `name` VARCHAR(100) NOT NULL COMMENT '姓名',
   `age` INT(11) NOT NULL COMMENT '年龄',
   `gender` VARCHAR(16) NOT NULL COMMENT '性别:male/female',
   `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
   `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
   `is_delete` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否删除 0:未删除 1:已删除',
   PRIMARY KEY (`id`),
   INDEX `index_name` (`name`),
   UNIQUE INDEX `unique_student_no` (`student_no`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT '学生表';

/*select * from student where age='12' group by name;*/

/*
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1002', 'aaa', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1003', 'sss', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1004', 'xiaoming', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1005', 'ddd', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1006', 'fff', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1007', 'xiaoming', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1008', 'xiaoming', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1009', 'xiaoming', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1010', 'aaa', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1011', 'fff', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1012', 'fff', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1013', 'aaa', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1014', 'sss', 13, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1015', 'aaa', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1016', 'sss', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1017', 'aaa', 12, 'male');
INSERT INTO `ant_test`.`student`(`student_no`, `name`, `age`, `gender`) VALUES ('1018', 'aaa', 12, 'male');
*/

SELECT * FROM student a WHERE id = (SELECT MIN(id) FROM student b WHERE b.age = 12 GROUP BY b.name);
SELECT * FROM student a WHERE id = (SELECT MIN(id) FROM student b WHERE b.name = a.name AND b.age=12 GROUP BY b.name);