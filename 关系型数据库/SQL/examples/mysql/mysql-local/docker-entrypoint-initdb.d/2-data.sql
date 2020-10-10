SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
--  Records of `component`
-- ----------------------------
BEGIN;
INSERT INTO `component` VALUES ('1', 'c1', '2019-01-09 05:23:10', null, null);
COMMIT;

-- ----------------------------
--  Records of `product`
-- ----------------------------
BEGIN;
INSERT INTO `product` VALUES ('1', 'p1', 'p1', 'cat1', '0.10', '2019-01-09 05:23:26', null, null), ('2', 'p2', 'p2', 'cat1', '0.20', '2019-01-09 05:23:41', null, null), ('3', 'p3', 'p3', 'cat2', '0.30', '2019-01-09 05:23:51', null, null);
COMMIT;

-- ----------------------------
--  Records of `product_component`
-- ----------------------------
BEGIN;
INSERT INTO `product_component` VALUES ('1', '1', '1', '2019-01-09 05:24:18', null, null);
COMMIT;

-- ----------------------------
--  Records of `user`
-- ----------------------------
BEGIN;
INSERT INTO `user` VALUES ('1', 'user', 'pasword', 'salt', '2019-01-09 05:24:04', null, null);
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
