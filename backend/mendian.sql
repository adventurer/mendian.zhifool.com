/*
 Navicat Premium Data Transfer

 Source Server         : lcoalhost
 Source Server Type    : MySQL
 Source Server Version : 80046
 Source Host           : localhost:3306
 Source Schema         : mendian

 Target Server Type    : MySQL
 Target Server Version : 80046
 File Encoding         : 65001

 Date: 28/09/2026 11:03:33
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for cart_items
-- ----------------------------
DROP TABLE IF EXISTS `cart_items`;
CREATE TABLE `cart_items`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `product_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `quantity` bigint NOT NULL,
  `unit_price` double NOT NULL,
  `options` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `cart_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'default',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_cart_items_merchant_id`(`merchant_id` ASC) USING BTREE,
  INDEX `idx_cart_items_product_id`(`product_id` ASC) USING BTREE,
  INDEX `idx_cart_items_scope`(`merchant_id` ASC, `cart_id` ASC) USING BTREE,
  INDEX `idx_cart_items_store_cart`(`merchant_id` ASC, `store_id` ASC, `cart_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of cart_items
-- ----------------------------
INSERT INTO `cart_items` VALUES (1, 'demo-merchant', 'rose-rice-latte', '玫瑰米酿拿铁', 2, 28, '超大热杯 473ml，燕麦奶，不另外加糖', 'split-test', 'default');
INSERT INTO `cart_items` VALUES (35, 'demo-merchant', 'one-fen-payment-test', '支付测试商品（1分）', 1, 0.01, '支付测试商品', 'cart-mukh7h91-z0qvt730zji', 'default');
INSERT INTO `cart_items` VALUES (36, 'demo-merchant', 'one-fen-payment-test', '支付测试商品（1分）', 1, 0.01, '支付测试商品', 'cart-mukh7h91-z0qvt730zji', 'default');

-- ----------------------------
-- Table structure for merchant_stores
-- ----------------------------
DROP TABLE IF EXISTS `merchant_stores`;
CREATE TABLE `merchant_stores`  (
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `latitude` double NOT NULL DEFAULT 0,
  `longitude` double NOT NULL DEFAULT 0,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `is_test` tinyint(1) NOT NULL DEFAULT 0,
  `phone` varchar(24) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`merchant_id`, `store_id`) USING BTREE,
  INDEX `idx_merchant_stores_is_active`(`is_active` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of merchant_stores
-- ----------------------------
INSERT INTO `merchant_stores` VALUES ('demo-merchant', 'default', '贵阳荟华里店', '贵州省贵阳市', 26.6477, 106.6302, 1, 1, '2026-09-28 07:23:00.452', '2026-09-28 07:23:00.452', 0, '16620835367');
INSERT INTO `merchant_stores` VALUES ('demo-merchant', 'demo-store', '多门店演示店（测试）', '演示数据，实际地址待配置', 24.6477, 88.6302, 1, 0, '2026-09-28 07:34:42.636', '2026-09-28 07:34:42.636', 0, '16620835367');

-- ----------------------------
-- Table structure for merchants
-- ----------------------------
DROP TABLE IF EXISTS `merchants`;
CREATE TABLE `merchants`  (
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `we_chat_pay_enabled` tinyint(1) NOT NULL DEFAULT 0,
  `we_chat_app_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `we_chat_mch_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `we_chat_certificate_serial_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `we_chat_apiv3_key_ciphertext` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `we_chat_private_key_ciphertext` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `we_chat_notify_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`merchant_id`) USING BTREE,
  INDEX `idx_merchants_we_chat_mch_id`(`we_chat_mch_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of merchants
-- ----------------------------
INSERT INTO `merchants` VALUES ('demo-merchant', '知甜蛋糕', 0, '', '', '', '', '', '', '2026-09-28 02:57:02.689', '2026-09-28 02:57:02.689');

-- ----------------------------
-- Table structure for payment_orders
-- ----------------------------
DROP TABLE IF EXISTS `payment_orders`;
CREATE TABLE `payment_orders`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `order_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending',
  `total_amount` bigint NOT NULL,
  `currency` varchar(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'CNY',
  `items` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `we_chat_transaction_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `we_chat_prepay_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `payer_open_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `expires_at` datetime(3) NULL DEFAULT NULL,
  `paid_at` datetime(3) NULL DEFAULT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `cart_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'default',
  `store_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `fulfillment_type` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'dine_in',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_payment_orders_merchant_order_no`(`merchant_id` ASC, `order_no` ASC) USING BTREE,
  UNIQUE INDEX `idx_payment_orders_we_chat_transaction_id`(`we_chat_transaction_id` ASC) USING BTREE,
  INDEX `idx_payment_orders_merchant_status`(`merchant_id` ASC, `status` ASC) USING BTREE,
  INDEX `idx_payment_orders_payer_open_id`(`payer_open_id` ASC) USING BTREE,
  INDEX `idx_payment_orders_cart_id`(`cart_id` ASC) USING BTREE,
  INDEX `idx_payment_orders_merchant_store`(`merchant_id` ASC, `store_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of payment_orders
-- ----------------------------
INSERT INTO `payment_orders` VALUES (1, 'demo-merchant', 'MD20260927214101153d175de69dde1b6338d838', 'closed', 2500, 'CNY', '[{\"cartItemId\":15,\"productId\":\"citrus-americano\",\"name\":\"柑橘美式\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, '', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 05:56:01.660', NULL, '2026-09-28 05:41:01.660', '2026-09-28 05:41:01.890', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (2, 'demo-merchant', 'MD20260927214107233b417c6fff8962a5195604', 'closed', 2500, 'CNY', '[{\"cartItemId\":15,\"productId\":\"citrus-americano\",\"name\":\"柑橘美式\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, '', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 05:56:07.577', NULL, '2026-09-28 05:41:07.577', '2026-09-28 05:41:07.692', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (3, 'demo-merchant', 'MD202609272142160ffa6d7e78b67d91a1937961', 'closed', 2500, 'CNY', '[{\"cartItemId\":15,\"productId\":\"citrus-americano\",\"name\":\"柑橘美式\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, '', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 05:57:16.015', NULL, '2026-09-28 05:42:16.015', '2026-09-28 05:42:16.236', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (4, 'demo-merchant', 'MD2026092721430192e811ec1636f0c5b6150d09', 'closed', 2500, 'CNY', '[{\"cartItemId\":15,\"productId\":\"citrus-americano\",\"name\":\"柑橘美式\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, '', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 05:58:01.024', NULL, '2026-09-28 05:43:01.024', '2026-09-28 05:43:01.223', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (5, 'demo-merchant', 'MD2609272144279b6dad7e86252749fd', 'pending', 2500, 'CNY', '[{\"cartItemId\":15,\"productId\":\"citrus-americano\",\"name\":\"柑橘美式\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, '', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 05:59:27.345', NULL, '2026-09-28 05:44:27.345', '2026-09-28 05:44:27.345', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (6, 'demo-merchant', 'MD2609272144425bc6c377f219848cfd', 'pending', 2500, 'CNY', '[{\"cartItemId\":16,\"productId\":\"rose-rice-latte\",\"name\":\"玫瑰米酿拿铁\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, '', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 05:59:42.519', NULL, '2026-09-28 05:44:42.519', '2026-09-28 05:44:42.519', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (7, 'demo-merchant', 'MD26092721470426c6a63e2508788bb2', 'pending', 2500, 'CNY', '[{\"cartItemId\":16,\"productId\":\"rose-rice-latte\",\"name\":\"玫瑰米酿拿铁\",\"quantity\":1,\"unitAmount\":2500,\"lineAmount\":2500,\"options\":\"超大冰杯 473ml（正常冰），牛奶，标准糖\"}]', NULL, 'wx3517566815829046450000045220263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:02:04.731', NULL, '2026-09-28 05:47:04.731', '2026-09-28 05:47:05.336', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (8, 'demo-merchant', 'MD26092721484855a82eb072c3753d81', 'paid', 1, 'CNY', '[{\"cartItemId\":17,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000470202609281738208122', 'wx2218028371829093450000047020263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:03:48.910', '2026-09-28 05:49:09.248', '2026-09-28 05:48:48.910', '2026-09-28 05:49:09.249', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (9, 'demo-merchant', 'MD260927215210bed2ee77ba41ba4e3a', 'pending', 1, 'CNY', '[{\"cartItemId\":18,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx73341903328290fe450000045020263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:07:10.490', NULL, '2026-09-28 05:52:10.490', '2026-09-28 05:52:11.056', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (10, 'demo-merchant', 'MD260927215319d1198d0943e835438e', 'pending', 1, 'CNY', '[{\"cartItemId\":19,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx63197102828290d2450000048520263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:08:19.804', NULL, '2026-09-28 05:53:19.804', '2026-09-28 05:53:20.333', 'cart-mukcugxn-ucv38s42mq', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (11, 'demo-merchant', 'MD260927215426efaec7c0a722a17602', 'pending', 1, 'CNY', '[{\"cartItemId\":18,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx1294692473829020450000045020263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:09:26.728', NULL, '2026-09-28 05:54:26.728', '2026-09-28 05:54:27.295', 'cart-muk6yel9-wzf1bpmj7sd', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (12, 'demo-merchant', 'MD260927215456779133b059122f515d', 'paid', 1, 'CNY', '[{\"cartItemId\":19,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000450202609281024111364', 'wx46311142018290ba450000045020263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:09:56.133', '2026-09-28 05:55:01.677', '2026-09-28 05:54:56.133', '2026-09-28 05:55:01.677', 'cart-mukcugxn-ucv38s42mq', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (13, 'demo-merchant', 'MD26092722510063ea870044525fdd52', 'pending', 1, 'CNY', '[{\"cartItemId\":21,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx1450279001829003450000045720263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 07:06:00.074', NULL, '2026-09-28 06:51:00.074', '2026-09-28 06:51:00.646', 'cart-mukd4bhs-bxjjx4jd2t', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (14, 'demo-merchant', 'MD260927225150ce63798ef91261918b', 'paid', 1, 'CNY', '[{\"cartItemId\":22,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000481202609281052856976', 'wx67965825018290ca450000048120263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 07:06:50.291', '2026-09-28 06:51:56.350', '2026-09-28 06:51:50.291', '2026-09-28 06:51:56.350', 'cart-mukcugxn-ucv38s42mq', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (15, 'demo-merchant', 'MD260927225705c3ceb1ada876c20322', 'paid', 3, 'CNY', '[{\"cartItemId\":23,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":2,\"unitAmount\":1,\"lineAmount\":2,\"options\":\"支付测试商品\"},{\"cartItemId\":25,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000477202609284441291140', 'wx0411921444829052450000047720263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 07:12:05.994', '2026-09-28 06:57:21.727', '2026-09-28 06:57:05.994', '2026-09-28 06:57:21.727', 'cart-mukd4bhs-bxjjx4jd2t', 'default', '', 'dine_in');
INSERT INTO `payment_orders` VALUES (16, 'demo-merchant', 'MD260927232431f1930c51e88382d444', 'paid', 1, 'CNY', '[{\"cartItemId\":27,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000446202609289221668667', 'wx766866122982901e450000044620263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 07:39:31.674', '2026-09-28 07:24:47.497', '2026-09-28 07:24:31.674', '2026-09-28 07:24:47.497', 'cart-mukd4bhs-bxjjx4jd2t', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (17, 'demo-merchant', 'MD2609272354014f6bfb416be9f839b6', 'paid', 2, 'CNY', '[{\"cartItemId\":28,\"productId\":\"demo-latte\",\"name\":\"演示拿铁（测试）\",\"quantity\":1,\"unitAmount\":2,\"lineAmount\":2,\"options\":\"\"}]', '4500000474202609286104212593', 'wx3952124016829098450000047420263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 08:09:01.306', '2026-09-28 07:54:07.599', '2026-09-28 07:54:01.306', '2026-09-28 07:54:07.600', 'cart-mukh3mob-hn0bleznfm', 'demo-store', '多门店演示店（测试）', 'dine_in');
INSERT INTO `payment_orders` VALUES (18, 'demo-merchant', 'MD260928000403a8f5be1f211de4dcba', 'pending', 1, 'CNY', '[{\"cartItemId\":29,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx15861618968290af450000049020263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 08:19:03.833', NULL, '2026-09-28 08:04:03.833', '2026-09-28 08:04:04.355', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (19, 'demo-merchant', 'MD260928003343832473b21992508164', 'pending', 2, 'CNY', '[{\"cartItemId\":29,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"},{\"cartItemId\":30,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx681442707482907b450000049320263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 08:48:43.301', NULL, '2026-09-28 08:33:43.301', '2026-09-28 08:33:43.798', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (20, 'demo-merchant', 'MD26092800364374f6a9d87ec4017bee', 'pending', 2, 'CNY', '[{\"cartItemId\":29,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"},{\"cartItemId\":30,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx55854582098290d2450000044420263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 08:51:43.805', NULL, '2026-09-28 08:36:43.805', '2026-09-28 08:36:44.341', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (21, 'demo-merchant', 'MD26092800395745f2c9454d3d900630', 'pending', 2, 'CNY', '[{\"cartItemId\":29,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"},{\"cartItemId\":30,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', NULL, 'wx557988506282906a450000044020263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 08:54:57.808', NULL, '2026-09-28 08:39:57.808', '2026-09-28 08:39:58.353', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (22, 'demo-merchant', 'MD26092800460952da38f3b55d04cf49', 'paid', 1, 'CNY', '[{\"cartItemId\":32,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000459202609286856331572', 'wx2751336586829026450000045920263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 09:01:09.916', '2026-09-28 08:46:29.506', '2026-09-28 08:46:09.916', '2026-09-28 08:46:29.506', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (23, 'demo-merchant', 'MD26092800471388584808268aa74c4f', 'paid', 1, 'CNY', '[{\"cartItemId\":33,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000492202609288395117628', 'wx8267115938829095450000049220263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 09:02:13.182', '2026-09-28 08:47:28.281', '2026-09-28 08:47:13.182', '2026-09-28 08:47:28.281', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'dine_in');
INSERT INTO `payment_orders` VALUES (24, 'demo-merchant', 'MD260928005240da62890644535c02f2', 'paid', 1, 'CNY', '[{\"cartItemId\":34,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000477202609283154195594', 'wx495591451382900b450000047720263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 09:07:40.055', '2026-09-28 08:52:52.902', '2026-09-28 08:52:40.055', '2026-09-28 08:52:52.902', 'cart-mukh7h91-z0qvt730zji', 'default', '贵阳荟华里店', 'delivery');
INSERT INTO `payment_orders` VALUES (25, 'demo-merchant', 'MD260928021104e53aec6bcc3842f9ce', 'paid', 1, 'CNY', '[{\"cartItemId\":37,\"productId\":\"one-fen-payment-test\",\"name\":\"支付测试商品（1分）\",\"quantity\":1,\"unitAmount\":1,\"lineAmount\":1,\"options\":\"支付测试商品\"}]', '4500000451202609285933893760', 'wx067398339582905f450000045120263000', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 10:26:04.005', '2026-09-28 10:11:09.241', '2026-09-28 10:11:04.005', '2026-09-28 10:11:09.241', 'cart-mukh38b5-5htlxtralbi', 'default', '贵阳荟华里店', 'dine_in');

-- ----------------------------
-- Table structure for store_menu_series
-- ----------------------------
DROP TABLE IF EXISTS `store_menu_series`;
CREATE TABLE `store_menu_series`  (
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `icon` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `badge` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `sort_order` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`merchant_id`, `store_id`, `id`) USING BTREE,
  CONSTRAINT `fk_store_menus_series` FOREIGN KEY (`merchant_id`, `store_id`) REFERENCES `store_menus` (`merchant_id`, `store_id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of store_menu_series
-- ----------------------------
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'classic-americano', '经典美式', 'icons/coffee.svg', '', 8);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'classic-espresso', '经典意式', 'icons/coffee.svg', '', 4);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'espresso-special', '浓缩特调', 'icons/bean.svg', '新', 11);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'flavored-latte', '风味拿铁', 'icons/milk.svg', '新', 9);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'fruit-americano', '果味美式', 'icons/citrus.svg', '爆', 2);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'large-cup', '超大杯系列', 'icons/cup-soda.svg', '', 3);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'milk-coffee', '热卖奶咖', 'icons/milk.svg', '', 6);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'oat', '燕麦系列', 'icons/wheat.svg', '', 7);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'payment-test', '支付测试', 'icons/coffee.svg', '', 12);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'reserve', '珍藏系列', 'icons/coffee.svg', '', 1);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'seasonal', '季节新品', 'icons/sparkles.svg', '', 0);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'soe', '单品豆SOE', 'icons/bean.svg', '新', 5);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'default', 'tea-coffee', '茶咖特调', 'icons/citrus.svg', '', 10);
INSERT INTO `store_menu_series` VALUES ('demo-merchant', 'demo-store', 'demo-products', '演示商品', 'icons/sparkles.svg', '', 0);

-- ----------------------------
-- Table structure for store_menus
-- ----------------------------
DROP TABLE IF EXISTS `store_menus`;
CREATE TABLE `store_menus`  (
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `brand_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  PRIMARY KEY (`merchant_id`, `store_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of store_menus
-- ----------------------------
INSERT INTO `store_menus` VALUES ('demo-merchant', 'default', '知甜蛋糕');
INSERT INTO `store_menus` VALUES ('demo-merchant', 'demo-store', '知甜蛋糕1');

-- ----------------------------
-- Table structure for store_product_option_values
-- ----------------------------
DROP TABLE IF EXISTS `store_product_option_values`;
CREATE TABLE `store_product_option_values`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `product_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `option_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `sort_order` bigint NOT NULL,
  `label` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `extra_price` double NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_store_product_option_value_order`(`merchant_id` ASC, `store_id` ASC, `product_id` ASC, `option_id` ASC, `sort_order` ASC) USING BTREE,
  CONSTRAINT `fk_store_product_options_values` FOREIGN KEY (`merchant_id`, `store_id`, `product_id`, `option_id`) REFERENCES `store_product_options` (`merchant_id`, `store_id`, `product_id`, `id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 164 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of store_product_option_values
-- ----------------------------
INSERT INTO `store_product_option_values` VALUES (164, 'demo-merchant', 'default', 'rose-rice-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (165, 'demo-merchant', 'default', 'rose-rice-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (166, 'demo-merchant', 'default', 'rose-rice-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (167, 'demo-merchant', 'default', 'rose-rice-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (168, 'demo-merchant', 'default', 'rose-rice-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (169, 'demo-merchant', 'default', 'rose-rice-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (170, 'demo-merchant', 'default', 'rose-rice-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (171, 'demo-merchant', 'default', 'rose-rice-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (172, 'demo-merchant', 'default', 'rose-rice-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (173, 'demo-merchant', 'default', 'butter-soe-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (174, 'demo-merchant', 'default', 'butter-soe-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (175, 'demo-merchant', 'default', 'butter-soe-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (176, 'demo-merchant', 'default', 'butter-soe-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (177, 'demo-merchant', 'default', 'butter-soe-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (178, 'demo-merchant', 'default', 'butter-soe-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (179, 'demo-merchant', 'default', 'butter-soe-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (180, 'demo-merchant', 'default', 'butter-soe-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (181, 'demo-merchant', 'default', 'butter-soe-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (182, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (183, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (184, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (185, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (186, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (187, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (188, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (189, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (190, 'demo-merchant', 'default', 'osmanthus-iced-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (191, 'demo-merchant', 'default', 'reserve-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (192, 'demo-merchant', 'default', 'reserve-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (193, 'demo-merchant', 'default', 'reserve-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (194, 'demo-merchant', 'default', 'reserve-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (195, 'demo-merchant', 'default', 'reserve-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (196, 'demo-merchant', 'default', 'reserve-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (197, 'demo-merchant', 'default', 'reserve-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (198, 'demo-merchant', 'default', 'reserve-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (199, 'demo-merchant', 'default', 'reserve-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (200, 'demo-merchant', 'default', 'citrus-americano', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (201, 'demo-merchant', 'default', 'citrus-americano', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (202, 'demo-merchant', 'default', 'citrus-americano', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (203, 'demo-merchant', 'default', 'citrus-americano', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (204, 'demo-merchant', 'default', 'citrus-americano', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (205, 'demo-merchant', 'default', 'citrus-americano', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (206, 'demo-merchant', 'default', 'citrus-americano', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (207, 'demo-merchant', 'default', 'citrus-americano', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (208, 'demo-merchant', 'default', 'citrus-americano', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (209, 'demo-merchant', 'default', 'large-americano', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (210, 'demo-merchant', 'default', 'large-americano', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (211, 'demo-merchant', 'default', 'large-americano', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (212, 'demo-merchant', 'default', 'large-americano', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (213, 'demo-merchant', 'default', 'large-americano', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (214, 'demo-merchant', 'default', 'large-americano', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (215, 'demo-merchant', 'default', 'large-americano', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (216, 'demo-merchant', 'default', 'large-americano', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (217, 'demo-merchant', 'default', 'large-americano', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (218, 'demo-merchant', 'default', 'classic-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (219, 'demo-merchant', 'default', 'classic-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (220, 'demo-merchant', 'default', 'classic-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (221, 'demo-merchant', 'default', 'classic-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (222, 'demo-merchant', 'default', 'classic-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (223, 'demo-merchant', 'default', 'classic-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (224, 'demo-merchant', 'default', 'classic-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (225, 'demo-merchant', 'default', 'classic-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (226, 'demo-merchant', 'default', 'classic-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (227, 'demo-merchant', 'default', 'soe-americano', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (228, 'demo-merchant', 'default', 'soe-americano', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (229, 'demo-merchant', 'default', 'soe-americano', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (230, 'demo-merchant', 'default', 'soe-americano', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (231, 'demo-merchant', 'default', 'soe-americano', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (232, 'demo-merchant', 'default', 'soe-americano', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (233, 'demo-merchant', 'default', 'soe-americano', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (234, 'demo-merchant', 'default', 'soe-americano', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (235, 'demo-merchant', 'default', 'soe-americano', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (236, 'demo-merchant', 'default', 'hot-milk-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (237, 'demo-merchant', 'default', 'hot-milk-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (238, 'demo-merchant', 'default', 'hot-milk-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (239, 'demo-merchant', 'default', 'hot-milk-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (240, 'demo-merchant', 'default', 'hot-milk-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (241, 'demo-merchant', 'default', 'hot-milk-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (242, 'demo-merchant', 'default', 'hot-milk-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (243, 'demo-merchant', 'default', 'hot-milk-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (244, 'demo-merchant', 'default', 'hot-milk-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (245, 'demo-merchant', 'default', 'oat-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (246, 'demo-merchant', 'default', 'oat-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (247, 'demo-merchant', 'default', 'oat-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (248, 'demo-merchant', 'default', 'oat-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (249, 'demo-merchant', 'default', 'oat-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (250, 'demo-merchant', 'default', 'oat-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (251, 'demo-merchant', 'default', 'oat-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (252, 'demo-merchant', 'default', 'oat-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (253, 'demo-merchant', 'default', 'oat-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (254, 'demo-merchant', 'default', 'classic-americano-ice', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (255, 'demo-merchant', 'default', 'classic-americano-ice', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (256, 'demo-merchant', 'default', 'classic-americano-ice', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (257, 'demo-merchant', 'default', 'classic-americano-ice', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (258, 'demo-merchant', 'default', 'classic-americano-ice', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (259, 'demo-merchant', 'default', 'classic-americano-ice', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (260, 'demo-merchant', 'default', 'classic-americano-ice', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (261, 'demo-merchant', 'default', 'classic-americano-ice', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (262, 'demo-merchant', 'default', 'classic-americano-ice', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (263, 'demo-merchant', 'default', 'classic-americano-hot', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (264, 'demo-merchant', 'default', 'classic-americano-hot', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (265, 'demo-merchant', 'default', 'classic-americano-hot', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (266, 'demo-merchant', 'default', 'classic-americano-hot', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (267, 'demo-merchant', 'default', 'classic-americano-hot', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (268, 'demo-merchant', 'default', 'classic-americano-hot', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (269, 'demo-merchant', 'default', 'classic-americano-hot', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (270, 'demo-merchant', 'default', 'classic-americano-hot', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (271, 'demo-merchant', 'default', 'classic-americano-hot', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (272, 'demo-merchant', 'default', 'vanilla-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (273, 'demo-merchant', 'default', 'vanilla-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (274, 'demo-merchant', 'default', 'vanilla-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (275, 'demo-merchant', 'default', 'vanilla-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (276, 'demo-merchant', 'default', 'vanilla-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (277, 'demo-merchant', 'default', 'vanilla-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (278, 'demo-merchant', 'default', 'vanilla-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (279, 'demo-merchant', 'default', 'vanilla-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (280, 'demo-merchant', 'default', 'vanilla-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (281, 'demo-merchant', 'default', 'caramel-latte', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (282, 'demo-merchant', 'default', 'caramel-latte', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (283, 'demo-merchant', 'default', 'caramel-latte', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (284, 'demo-merchant', 'default', 'caramel-latte', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (285, 'demo-merchant', 'default', 'caramel-latte', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (286, 'demo-merchant', 'default', 'caramel-latte', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (287, 'demo-merchant', 'default', 'caramel-latte', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (288, 'demo-merchant', 'default', 'caramel-latte', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (289, 'demo-merchant', 'default', 'caramel-latte', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (290, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (291, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (292, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (293, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (294, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (295, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (296, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (297, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (298, 'demo-merchant', 'default', 'osmanthus-tea-coffee', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (299, 'demo-merchant', 'default', 'citrus-tea-coffee', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (300, 'demo-merchant', 'default', 'citrus-tea-coffee', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (301, 'demo-merchant', 'default', 'citrus-tea-coffee', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (302, 'demo-merchant', 'default', 'citrus-tea-coffee', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (303, 'demo-merchant', 'default', 'citrus-tea-coffee', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (304, 'demo-merchant', 'default', 'citrus-tea-coffee', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (305, 'demo-merchant', 'default', 'citrus-tea-coffee', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (306, 'demo-merchant', 'default', 'citrus-tea-coffee', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (307, 'demo-merchant', 'default', 'citrus-tea-coffee', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (308, 'demo-merchant', 'default', 'espresso-tonic', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (309, 'demo-merchant', 'default', 'espresso-tonic', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (310, 'demo-merchant', 'default', 'espresso-tonic', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (311, 'demo-merchant', 'default', 'espresso-tonic', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (312, 'demo-merchant', 'default', 'espresso-tonic', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (313, 'demo-merchant', 'default', 'espresso-tonic', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (314, 'demo-merchant', 'default', 'espresso-tonic', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (315, 'demo-merchant', 'default', 'espresso-tonic', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (316, 'demo-merchant', 'default', 'espresso-tonic', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (317, 'demo-merchant', 'default', 'dirty-coffee', 'size', 0, '超大冰杯 473ml（正常冰）', 5);
INSERT INTO `store_product_option_values` VALUES (318, 'demo-merchant', 'default', 'dirty-coffee', 'size', 1, '超大热杯 473ml', 5);
INSERT INTO `store_product_option_values` VALUES (319, 'demo-merchant', 'default', 'dirty-coffee', 'size', 2, '大冰杯 355ml（正常冰）', 0);
INSERT INTO `store_product_option_values` VALUES (320, 'demo-merchant', 'default', 'dirty-coffee', 'size', 3, '大热杯 355ml', 0);
INSERT INTO `store_product_option_values` VALUES (321, 'demo-merchant', 'default', 'dirty-coffee', 'milk', 0, '牛奶', 0);
INSERT INTO `store_product_option_values` VALUES (322, 'demo-merchant', 'default', 'dirty-coffee', 'milk', 1, '燕麦奶', 3);
INSERT INTO `store_product_option_values` VALUES (323, 'demo-merchant', 'default', 'dirty-coffee', 'sweetness', 0, '标准糖', 0);
INSERT INTO `store_product_option_values` VALUES (324, 'demo-merchant', 'default', 'dirty-coffee', 'sweetness', 1, '少糖', 0);
INSERT INTO `store_product_option_values` VALUES (325, 'demo-merchant', 'default', 'dirty-coffee', 'sweetness', 2, '不另外加糖', 0);
INSERT INTO `store_product_option_values` VALUES (326, 'demo-merchant', 'default', 'one-fen-payment-test', 'test-option', 0, '支付测试商品', 0);

-- ----------------------------
-- Table structure for store_product_options
-- ----------------------------
DROP TABLE IF EXISTS `store_product_options`;
CREATE TABLE `store_product_options`  (
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `product_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `title` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `selected_index` bigint NOT NULL DEFAULT 0,
  `sort_order` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`merchant_id`, `store_id`, `product_id`, `id`) USING BTREE,
  CONSTRAINT `fk_store_products_options` FOREIGN KEY (`merchant_id`, `store_id`, `product_id`) REFERENCES `store_products` (`merchant_id`, `store_id`, `id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of store_product_options
-- ----------------------------
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'butter-soe-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'butter-soe-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'butter-soe-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'caramel-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'caramel-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'caramel-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'citrus-americano', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'citrus-americano', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'citrus-americano', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'citrus-tea-coffee', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'citrus-tea-coffee', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'citrus-tea-coffee', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-americano-hot', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-americano-hot', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-americano-hot', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-americano-ice', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-americano-ice', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-americano-ice', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'classic-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'dirty-coffee', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'dirty-coffee', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'dirty-coffee', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'espresso-tonic', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'espresso-tonic', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'espresso-tonic', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'hot-milk-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'hot-milk-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'hot-milk-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'large-americano', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'large-americano', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'large-americano', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'oat-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'oat-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'oat-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'one-fen-payment-test', 'test-option', '商品', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'osmanthus-iced-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'osmanthus-iced-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'osmanthus-iced-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'osmanthus-tea-coffee', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'osmanthus-tea-coffee', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'osmanthus-tea-coffee', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'reserve-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'reserve-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'reserve-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'rose-rice-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'rose-rice-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'rose-rice-latte', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'soe-americano', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'soe-americano', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'soe-americano', 'sweetness', '甜度', 0, 2);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'vanilla-latte', 'milk', '奶搭配', 0, 1);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'vanilla-latte', 'size', '杯型及温度', 0, 0);
INSERT INTO `store_product_options` VALUES ('demo-merchant', 'default', 'vanilla-latte', 'sweetness', '甜度', 0, 2);

-- ----------------------------
-- Table structure for store_products
-- ----------------------------
DROP TABLE IF EXISTS `store_products`;
CREATE TABLE `store_products`  (
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `store_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `series_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `price` double NOT NULL,
  `image` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `sort_order` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`merchant_id`, `store_id`, `id`) USING BTREE,
  INDEX `idx_store_products_series_id`(`series_id` ASC) USING BTREE,
  INDEX `fk_store_menu_series_products`(`merchant_id` ASC, `store_id` ASC, `series_id` ASC) USING BTREE,
  CONSTRAINT `fk_store_menu_series_products` FOREIGN KEY (`merchant_id`, `store_id`, `series_id`) REFERENCES `store_menu_series` (`merchant_id`, `store_id`, `id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of store_products
-- ----------------------------
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'butter-soe-latte', 'seasonal', '黄油SOE拿铁', '精选黄油棕果SOE咖啡豆，不额外加糖，烘烤榛果温润扎实。', 20, 'coffee/butter-latte.jpg', 1);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'caramel-latte', 'flavored-latte', '焦糖拿铁', '焦糖香气融入绵密牛奶，甜润而不腻。', 24, 'coffee/butter-latte.jpg', 1);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'citrus-americano', 'fruit-americano', '柑橘美式', '清新柑橘香气融入醇正美式，酸甜清爽。', 20, 'coffee/osmanthus-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'citrus-tea-coffee', 'tea-coffee', '柑橘冷萃', '柑橘果香融入冷萃咖啡，酸甜爽口。', 22, 'coffee/osmanthus-latte.jpg', 1);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'classic-americano-hot', 'classic-americano', '经典热美式', '醇厚浓缩与热水相融，带来平衡顺口的风味。', 16, 'coffee/rose-latte.jpg', 1);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'classic-americano-ice', 'classic-americano', '经典冰美式', '双份浓缩融合清冽冰水，口感干净明亮。', 16, 'coffee/butter-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'classic-latte', 'classic-espresso', '经典拿铁', '浓缩咖啡与绵密牛奶融合，口感平衡柔和。', 20, 'coffee/rose-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'dirty-coffee', 'espresso-special', '脏脏咖啡', '热浓缩缓缓注入冰牛奶，呈现浓郁冷热交融。', 24, 'coffee/butter-latte.jpg', 1);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'espresso-tonic', 'espresso-special', '浓缩汤力', '浓缩咖啡搭配气泡汤力水，清爽带有柑橘香。', 22, 'coffee/osmanthus-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'hot-milk-latte', 'milk-coffee', '热卖奶咖', '现萃浓缩与热牛奶相融，口感醇厚细腻。', 20, 'coffee/butter-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'large-americano', 'large-cup', '超大杯美式', '双份浓缩搭配清冽冷水，满足大杯畅饮。', 18, 'coffee/butter-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'oat-latte', 'oat', '燕麦拿铁', '植物燕麦奶带来谷物香气，轻盈顺口。', 23, 'coffee/rose-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'one-fen-payment-test', 'payment-test', '支付测试商品（1分）', '用于微信支付流程验证，单价为人民币0.01元。', 0.01, 'coffee/rose-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'osmanthus-iced-latte', 'seasonal', '桂花冰拿铁', '浓郁咖啡与清甜桂花，入口清爽柔和。', 22, 'coffee/osmanthus-latte.jpg', 2);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'osmanthus-tea-coffee', 'tea-coffee', '桂花茶咖', '桂花清香与咖啡醇香相遇，层次清新。', 23, 'coffee/osmanthus-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'reserve-latte', 'reserve', '珍藏拿铁', '甄选咖啡豆搭配醇厚牛奶，香气饱满顺滑。', 22, 'coffee/butter-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'rose-rice-latte', 'seasonal', '玫瑰米酿拿铁', '玫瑰芬芳融入温润发酵米香，伴随淡淡桂花清香与苹果甜香。含少量酒精（低于0.5%vol），孕妇、驾驶人士及未成年人请谨慎选择。', 20, 'coffee/rose-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'soe-americano', 'soe', '单品豆SOE美式', '单一产地咖啡豆呈现明亮果香与干净回甘。', 18, 'coffee/osmanthus-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'default', 'vanilla-latte', 'flavored-latte', '香草拿铁', '香草甜香与浓缩咖啡交织，口感柔和醇厚。', 24, 'coffee/rose-latte.jpg', 0);
INSERT INTO `store_products` VALUES ('demo-merchant', 'demo-store', 'demo-americano', 'demo-products', '演示美式（测试）', '测试商品：零元演示数据，不可支付。', 0, 'coffee/butter-latte.jpg', 1);
INSERT INTO `store_products` VALUES ('demo-merchant', 'demo-store', 'demo-citrus', 'demo-products', '演示柑橘饮（测试）', '测试商品：零元演示数据，不可支付。', 0, 'coffee/osmanthus-latte.jpg', 2);
INSERT INTO `store_products` VALUES ('demo-merchant', 'demo-store', 'demo-latte', 'demo-products', '演示拿铁（测试）', '测试商品：零元演示数据，不可支付。', 0.02, 'coffee/rose-latte.jpg', 0);

-- ----------------------------
-- Table structure for user_addresses
-- ----------------------------
DROP TABLE IF EXISTS `user_addresses`;
CREATE TABLE `user_addresses`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `merchant_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `app_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `open_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `recipient` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `phone` varchar(24) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `province` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `city` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `district` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  `detail` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_addresses_owner`(`merchant_id` ASC, `app_id` ASC, `open_id` ASC, `is_default` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_addresses
-- ----------------------------
INSERT INTO `user_addresses` VALUES (5, 'demo-merchant', 'wxb2d783f0bfbff242', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '让外人', '13254685975', '北京市', '北京市', '东城区', '丰富43', 0, '2026-09-28 09:13:15.994', '2026-09-28 09:13:48.772');
INSERT INTO `user_addresses` VALUES (6, 'demo-merchant', 'wxb2d783f0bfbff242', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2423', '13254695874', '天津市', '天津市', '和平区', '234', 1, '2026-09-28 09:13:38.270', '2026-09-28 09:13:48.773');

-- ----------------------------
-- Table structure for wechat_users
-- ----------------------------
DROP TABLE IF EXISTS `wechat_users`;
CREATE TABLE `wechat_users`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `app_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `openid` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_wechat_users_app_openid`(`app_id` ASC, `openid` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 48 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of wechat_users
-- ----------------------------
INSERT INTO `wechat_users` VALUES (1, 'wxb2d783f0bfbff242', 'oEuFe15f6ClqUekFNWQwS23Zx6F8', '2026-09-28 06:08:48.535', '2026-09-28 06:08:48.535');

SET FOREIGN_KEY_CHECKS = 1;
