# Voyara C2C 跨国电商平台 — 完整功能设计文档

> 基于现有 Voyara Marketplace 模块的渐进式增强方案
> 参考：Amazon, Alibaba International, AliExpress

---

# Part A: 业务概述

## 1. 业务模式

| 维度 | 定义 |
|------|------|
| 模式 | **C2C** — 个人对个人，卖家自行发布商品、自行发货 |
| 定位 | 二手电器/大件商品跨境交易平台 |
| 目标市场 | 非洲、东南亚、中亚（后续扩展全球） |
| 盈利模式 | 平台交易佣金（5-10%）、广告位、卖家增值服务 |
| 多语言 | 英语（默认）、法语、阿拉伯语、俄语、中文、后续扩展 |
| 多币种 | 美元（基准）+ 目标市场本币显示 |

---

# Part B: 功能模块设计

## 2. 用户系统

### 2.1 注册 / 登录

**注册方式：**
- 邮箱注册（首选）：填写邮箱 → 发送验证码 → 验证 → 设置密码 → 完成
- 手机号注册（二期）：填写手机号 → SMS 验证码 → 完成

**登录方式：**
- 邮箱 + 密码
- 手机号 + 密码
- 第三方 OAuth 登录（Google, Facebook, Apple ID）

### 2.2 认证流程

```
┌─ 注册 ─────────────────────────────────┐
│  1. 用户填写邮箱 / 手机号                │
│  2. 前端格式校验（正则 + 实时反馈）        │
│  3. 后端发送验证码（AWS SES / SMS）       │
│  4. 用户输入验证码                        │
│  5. 后端验证 → 创建账户 → 返回 JWT       │
│  6. 引导完善个人资料                      │
└──────────────────────────────────────────┘
┌─ 登录 ─────────────────────────────────┐
│  1. 用户输入邮箱/手机号 + 密码            │
│  2. 后端 bcrypt 验证                     │
│  3. 返回 Access Token (24h) + Refresh    │
│     Token (7天)                          │
│  4. 旧密码哈希检测 → 自动升级 bcrypt      │
└──────────────────────────────────────────┘
```

### 2.3 密码安全

```sql
-- 用户表新增字段
ALTER TABLE voyara_users
  ADD COLUMN password_hash_method ENUM('sha256_legacy','bcrypt') NOT NULL DEFAULT 'bcrypt',
  ADD COLUMN email_verified_at DATETIME,
  ADD COLUMN phone_verified_at DATETIME,
  ADD COLUMN login_attempts INT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN locked_until DATETIME,
  ADD COLUMN last_login_at DATETIME,
  ADD COLUMN last_login_ip VARCHAR(45),
  ADD COLUMN role ENUM('user','seller','admin') NOT NULL DEFAULT 'user',
  ADD INDEX idx_email (email),
  ADD INDEX idx_phone (phone);
```

**密码规则：**
- 最少 8 位，含大写 + 小写 + 数字
- 后端 bcrypt cost=12
- 连续 5 次登录失败 → 锁定 30 分钟
- 旧 SHA-256 哈希在登录时自动升级为 bcrypt

### 2.4 第三方登录

```
API: POST /voyara/auth/oauth/{provider}
provider: google, facebook, apple

流程:
  1. 前端使用对应 SDK 获取 OAuth Access Token
  2. 后端用 Token 换取用户信息（邮箱、姓名、头像）
  3. 新用户 → 自动创建账户（无需额外注册）
  4. 老用户 → 返回 JWT
```

**数据库表 `voyara_oauth_accounts`：**

```sql
CREATE TABLE voyara_oauth_accounts (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id       BIGINT UNSIGNED NOT NULL,
  provider      VARCHAR(20) NOT NULL,     -- google, facebook, apple
  provider_uid  VARCHAR(255) NOT NULL,    -- 第三方平台 UID
  access_token  TEXT,
  refresh_token TEXT,
  expires_at    DATETIME,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  UNIQUE KEY uk_provider_uid (provider, provider_uid)
);
```

### 2.5 用户地址管理

**数据库表 `voyara_addresses`：**

```sql
CREATE TABLE voyara_addresses (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id       BIGINT UNSIGNED NOT NULL,
  label         VARCHAR(50) COMMENT '标签：家/公司/学校',
  recipient_name  VARCHAR(100) NOT NULL,
  phone         VARCHAR(50) NOT NULL,
  country       VARCHAR(100) NOT NULL,
  state         VARCHAR(100),
  city          VARCHAR(100) NOT NULL,
  district      VARCHAR(100),
  street        VARCHAR(500) NOT NULL,
  zip_code      VARCHAR(20),
  is_default    TINYINT(1) NOT NULL DEFAULT 0,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  INDEX idx_user (user_id)
);
```

**API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/addresses` | 地址列表 |
| POST | `/voyara/addresses` | 新增地址 |
| PUT | `/voyara/addresses/:id` | 编辑地址 |
| DELETE | `/voyara/addresses/:id` | 删除地址 |
| PUT | `/voyara/addresses/:id/default` | 设为默认 |

### 2.6 用户收藏 / 心愿单

**数据库表 `voyara_favorites`：**

```sql
CREATE TABLE voyara_favorites (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL,
  product_id  BIGINT UNSIGNED NOT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  FOREIGN KEY (product_id) REFERENCES voyara_products(id),
  UNIQUE KEY uk_user_product (user_id, product_id)
);
```

**API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/favorites` | 收藏列表（分页） |
| POST | `/voyara/favorites/:productId` | 添加收藏 |
| DELETE | `/voyara/favorites/:productId` | 取消收藏 |
| GET | `/voyara/favorites/check?productIds=1,2,3` | 批量检查是否已收藏 |

### 2.7 会员等级

```sql
ALTER TABLE voyara_users
  ADD COLUMN membership_tier ENUM('bronze','silver','gold','platinum') NOT NULL DEFAULT 'bronze',
  ADD COLUMN points INT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN total_spent DECIMAL(14,2) NOT NULL DEFAULT 0;
```

**等级规则：**

| 等级 | 累计消费门槛 | 权益 |
|------|-------------|------|
| Bronze | $0 | 基础功能 |
| Silver | $500 | 免运费、优先客服 |
| Gold | $2,000 | 额外 5% 折扣、专属优惠 |
| Platinum | $10,000 | 额外 10% 折扣、VIP 客服、生日礼包 |

### 2.8 忘记 / 重置密码

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/voyara/auth/forgot-password` | 发送重置密码邮件（含 token 链接） |
| POST | `/voyara/auth/reset-password` | 验证 token + 设置新密码 |
| POST | `/voyara/auth/change-password` | 修改密码（需旧密码） |

### 2.9 账号安全

- 登录设备管理（查看已登录设备、远程登出）
- 登录历史记录（IP、地点、时间）
- 双重认证 (2FA) — 二期

---

## 3. 商品系统

### 3.1 核心概念

| 概念 | 说明 |
|------|------|
| **SPU** (Standard Product Unit) | 商品标准化单元，如 "iPhone 14" |
| **SKU** (Stock Keeping Unit) | 库存量单位，如 "iPhone 14 黑色 128G" |
| **规格属性** | 颜色、尺寸、容量等可组合的变体维度 |

### 3.2 数据库设计

**`voyara_spus` — SPU 表（商品模板）：**

```sql
CREATE TABLE voyara_spus (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  seller_id       BIGINT UNSIGNED NOT NULL,
  title           VARCHAR(500) NOT NULL,
  description     TEXT,
  brand_id        BIGINT UNSIGNED,
  category_id     BIGINT UNSIGNED NOT NULL,
  images          JSON COMMENT '最多9张图片URL',
  videos          JSON COMMENT '商品视频URL',
  tags            JSON COMMENT '标签数组',
  search_keywords VARCHAR(500) COMMENT '搜索关键词，逗号分隔',
  status          ENUM('draft','pending','active','rejected','inactive') NOT NULL DEFAULT 'draft',
  reject_reason   VARCHAR(500),
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (seller_id) REFERENCES voyara_sellers(id),
  FOREIGN KEY (brand_id) REFERENCES voyara_brands(id),
  FOREIGN KEY (category_id) REFERENCES voyara_categories(id),
  FULLTEXT INDEX ft_search (title, description, search_keywords)
);
```

**`voyara_skus` — SKU 表（实际可售单元，取代原有 voyara_products）：**

```sql
CREATE TABLE voyara_skus (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  spu_id          BIGINT UNSIGNED NOT NULL,
  seller_id       BIGINT UNSIGNED NOT NULL,
  spec_names      JSON COMMENT '{ "颜色": "黑色", "存储": "128G" }',
  price           DECIMAL(12,2) NOT NULL,
  currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
  stock           INT UNSIGNED NOT NULL DEFAULT 0,
  reserved_stock  INT UNSIGNED NOT NULL DEFAULT 0,
  sold_count      INT UNSIGNED NOT NULL DEFAULT 0,
  status          ENUM('active','inactive') NOT NULL DEFAULT 'active',
  images          JSON COMMENT 'SKU专属图片，覆盖SPU图片',
  sort_order      INT NOT NULL DEFAULT 0,
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (spu_id) REFERENCES voyara_spus(id),
  FOREIGN KEY (seller_id) REFERENCES voyara_sellers(id),
  INDEX idx_spu (spu_id),
  INDEX idx_seller_status (seller_id, status)
);
```

**`voyara_brands` — 品牌表：**

```sql
CREATE TABLE voyara_brands (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(200) NOT NULL,
  name_zh     VARCHAR(200) COMMENT '中文名',
  logo        VARCHAR(500),
  description TEXT,
  sort_order  INT NOT NULL DEFAULT 0,
  status      ENUM('active','inactive') NOT NULL DEFAULT 'active',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**`voyara_spec_templates` — 规格模板：**

```sql
CREATE TABLE voyara_spec_templates (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  category_id   BIGINT UNSIGNED NOT NULL,
  spec_name     VARCHAR(100) NOT NULL COMMENT '规格名：颜色、尺寸等',
  sort_order    INT NOT NULL DEFAULT 0,
  FOREIGN KEY (category_id) REFERENCES voyara_categories(id)
);

CREATE TABLE voyara_spec_values (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  spec_template_id BIGINT UNSIGNED NOT NULL,
  spec_value      VARCHAR(100) NOT NULL COMMENT '规格值：黑色、红色等',
  sort_order      INT NOT NULL DEFAULT 0,
  FOREIGN KEY (spec_template_id) REFERENCES voyara_spec_templates(id)
);
```

### 3.3 商品发布流程

```
卖家填写 → 基本信息（标题、描述、品牌、分类）
         → 上传图片（最多9张，首图为封面）
         → 上传视频（可选）
         → 填写规格（依据分类模板）
         → 填写每个SKU的价格、库存
         → 填写标签和搜索关键词
         → 提交审核
```

**API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/spus` | 商品列表（分页、筛选） |
| GET | `/voyara/spus/:id` | 商品详情（含 SKU 列表） |
| POST | `/voyara/spus` | 创建商品 |
| PUT | `/voyara/spus/:id` | 编辑商品 |
| PUT | `/voyara/spus/:id/status` | 更新状态（上架/下架） |
| POST | `/voyara/spus/:id/submit-review` | 提交审核 |
| POST | `/voyara/upload/presigned` | 获取图片上传预签名 URL |

**SKU 管理 API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/spus/:spuId/skus` | SKU 列表 |
| POST | `/voyara/spus/:spuId/skus` | 新增 SKU |
| PUT | `/voyara/spus/:spuId/skus/:id` | 编辑 SKU |
| DELETE | `/voyara/spus/:spuId/skus/:id` | 删除 SKU |

**品牌 API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/brands` | 品牌列表 |

### 3.4 商品状态流转

```
draft ───→ pending ───→ active ───→ inactive
  ↑            ↓                      ↑
  └──────── rejected                  └── (卖家重新上架)

active → sold (全部 SKU 售罄时自动)
active → inactive (卖家手动下架 / 管理员强制下架)
```

### 3.5 商品上下架状态

- **SPU 级别**：控制整个商品可见性
- **SKU 级别**：控制特定规格是否可售
- 全部 SKU inactive → SPU 自动变为 inactive
- 至少一个 SKU active → SPU 可设为 active

---

## 4. 购物车与下单流程

### 4.1 购物车

**数据库 `voyara_cart_items`：**

```sql
CREATE TABLE voyara_cart_items (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL,
  sku_id      BIGINT UNSIGNED NOT NULL,
  spu_id      BIGINT UNSIGNED NOT NULL,
  quantity    INT UNSIGNED NOT NULL DEFAULT 1,
  selected    TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否选中用于结算',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  FOREIGN KEY (sku_id) REFERENCES voyara_skus(id),
  FOREIGN KEY (spu_id) REFERENCES voyara_spus(id),
  UNIQUE KEY uk_user_sku (user_id, sku_id)
);
```

**购物车规则：**
- 同一 SKU 合并数量，最多 99 件
- 跨卖家商品可在购物车共存（C2C 平台特性）
- 已下架/库存不足的商品显示为灰色并提示
- 购物车总计最多 50 个商品项
- 前端缓存购物车数量徽标，每次打开时刷新

**API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/cart` | 购物车列表（含 SKU/SPU 详情、卖家分组） |
| POST | `/voyara/cart` | 添加 `{skuId, quantity}` |
| PUT | `/voyara/cart/:itemId` | 更新数量 |
| PUT | `/voyara/cart/select` | 批量选中/取消 `{itemIds, selected}` |
| DELETE | `/voyara/cart/:itemId` | 删除商品 |
| DELETE | `/voyara/cart` | 清空购物车 |

**购物车前端数据结构：**

```typescript
// 按卖家分组的购物车
interface CartGrouped {
  sellerId: number;
  shopName: string;
  items: CartItem[];
  checked: boolean;       // 全选/取消
}

interface CartItem {
  id: number;
  skuId: number;
  spuId: number;
  title: string;
  specNames: Record<string, string>;  // { "颜色": "黑色" }
  image: string;
  price: number;
  quantity: number;
  stock: number;
  selected: boolean;
  isAvailable: boolean;   // 可购买状态
  errorMessage?: string;  // 不可购买原因
}
```

### 4.2 下单流程

```
购物车 → 选中商品 → 点击"结算"
  ↓
选择/填写收货地址
  ↓
选择优惠券（自动计算最优）
  ↓
计算运费（按运费模板）
  ↓
确认订单页（商品、数量、价格、运费、优惠、总计）
  ↓
提交订单 → 幂等键防重 → 锁库存 → 创建订单
  ↓
跳转支付页
```

### 4.3 选择规格

- 商品详情页通过规格选择器切换不同 SKU
- 每个 SKU 组合对应唯一价格和库存
- 规格选择器 UI：按钮组（颜色→容量→尺寸按序选择）
- 所选规格实时更新价格、库存、图片
- 无库存的 SKU 选项灰化

### 4.4 填写地址

- 地址选择器：从已保存地址中选择，或新建
- 跨境地址需填写：国家、城市、详细地址、邮编、收件人、电话
- 地址格式按国家动态调整（如美国需州，中国需省市区）

### 4.5 选择优惠券

- 自动展示可用优惠券列表
- 系统自动选择最优优惠券（用户可手动切换）
- 不可用的优惠券灰色显示并标注原因

### 4.6 计算运费

- 根据卖家的运费模板计算
- 多商品同卖家合并运费
- 具体见第 8 节物流与配送

### 4.7 确认订单页面数据结构

```typescript
interface CheckoutConfirm {
  groups: CheckoutGroup[];    // 按卖家分组
  address: Address;
  coupon?: Coupon;
  summary: {
    subtotal: number;          // 商品总额
    shippingTotal: number;     // 运费总额
    discountTotal: number;     // 优惠总额
    grandTotal: number;        // 应付总额
  };
}

interface CheckoutGroup {
  sellerId: number;
  shopName: string;
  items: CheckoutItem[];
  shippingFee: number;
  shippingMethod: string;
  subtotal: number;
}

interface CheckoutItem {
  skuId: number;
  spuId: number;
  title: string;
  specNames: Record<string,string>;
  price: number;
  quantity: number;
  image: string;
}
```

---

## 5. 支付系统

### 5.1 支持的支付方式

| 方式 | 适用区域 | 接入方式 | 优先级 |
|------|----------|----------|--------|
| Stripe | 全球 | API + Webhook | P0 |
| PayPal | 全球 | API + Webhook | P0 |
| 支付宝 | 中国 / 东南亚 | 官方 API | P1 |
| 微信支付 | 中国 | 官方 API | P1 |
| 信用卡直付 | 全球 | Stripe 封装 | P0 (通过 Stripe) |
| 货到付款 | 特定市场 | 线下 | P2 |

**首期实现（P0）：Stripe + PayPal**，覆盖全球多数用户。

### 5.2 支付流程

```
┌─────────────────────────────────────────────────────┐
│ 用户确认订单 → 选择支付方式                           │
│                                                     │
│ ┌─ Stripe ────────────────────────────────────────┐ │
│ │ 1. 后端创建 PaymentIntent → 返回 clientSecret   │ │
│ │ 2. 前端 Stripe Elements 渲染支付表单             │ │
│ │ 3. 用户输入卡信息 → Stripe.js 处理 → 返回 PI    │ │
│ │ 4. 前端 confirmPayment → 支付                    │ │
│ │ 5. Stripe Webhook → payment_intent.succeeded     │ │
│ │ 6. 后端验签 → 更新订单 → 扣库存                  │ │
│ └─────────────────────────────────────────────────┘ │
│                                                     │
│ ┌─ PayPal ────────────────────────────────────────┐ │
│ │ 1. 后端创建 PayPal Order → 返回 approval URL    │ │
│ │ 2. 前端跳转 PayPal / 打开弹窗                    │ │
│ │ 3. 用户登录 PayPal → 确认支付                    │ │
│ │ 4. PayPal 回调 → 前端调用后端 capture API        │ │
│ │ 5. 后端 capture → 更新订单 → 扣库存              │ │
│ └─────────────────────────────────────────────────┘ │
│                                                     │
│ ┌─ 支付宝 / 微信支付 ───────────────────────────┐ │
│ │ 1. 后端生成支付二维码 / 支付链接                  │ │
│ │ 2. 前端展示二维码                                │ │
│ │ 3. 用户扫码支付                                  │ │
│ │ 4. 异步回调 → 更新订单                           │ │
│ └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

### 5.3 数据库 `voyara_payments`

```sql
CREATE TABLE voyara_payments (
  id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id                BIGINT UNSIGNED NOT NULL,
  buyer_id                BIGINT UNSIGNED NOT NULL,
  amount                  DECIMAL(12,2) NOT NULL,
  currency                VARCHAR(3) NOT NULL DEFAULT 'USD',
  payment_method          ENUM('stripe','paypal','alipay','wechat','cod') NOT NULL,
  payment_status          ENUM('pending','processing','succeeded','failed','refunded','partial_refunded') NOT NULL DEFAULT 'pending',
  
  -- Stripe
  stripe_payment_intent_id VARCHAR(255),
  stripe_charge_id        VARCHAR(255),
  
  -- PayPal
  paypal_order_id         VARCHAR(255),
  paypal_capture_id       VARCHAR(255),
  
  -- 支付宝 / 微信
  alipay_trade_no         VARCHAR(255),
  wechat_transaction_id   VARCHAR(255),
  
  -- 通用
  gateway_response        JSON COMMENT '原始回调数据',
  paid_at                 DATETIME,
  refunded_at             DATETIME,
  refund_amount           DECIMAL(12,2) DEFAULT 0,
  
  -- 对账
  reconciliation_status   ENUM('unmatched','matched','disputed') NOT NULL DEFAULT 'unmatched',
  reconciled_at           DATETIME,
  
  created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  FOREIGN KEY (buyer_id) REFERENCES voyara_users(id),
  INDEX idx_order (order_id),
  INDEX idx_stripe_intent (stripe_payment_intent_id),
  INDEX idx_paypal_order (paypal_order_id)
);
```

### 5.4 支付 Webhook 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/voyara/payment/stripe-webhook` | Stripe 支付回调（公开，签名验证） |
| POST | `/voyara/payment/paypal-webhook` | PayPal 支付回调（公开，签名验证） |
| POST | `/voyara/payment/alipay-notify` | 支付宝异步通知（公开，签名验证） |
| POST | `/voyara/payment/wechat-notify` | 微信支付通知（公开） |

### 5.5 支付验签

- **Stripe**：验证 `stripe-signature` header，使用 Webhook Secret
- **PayPal**：验证 `PAYPAL-AUTH-ALGO` 等 headers，使用 PayPal REST API 验证
- **支付宝**：验证签名（MD5/RSA2）+ 校验 notify_id
- **微信支付**：验证签名（HMAC-SHA256）

### 5.6 支付失败处理

| 失败场景 | 处理方式 |
|----------|----------|
| 卡余额不足 | 提示用户换卡或换支付方式 |
| 支付超时（用户未完成） | 订单保持不变，库存锁定 30 分钟 |
| 网络错误 | 提供"重新支付"按钮 |
| 风控拒绝 | 提示联系银行或换支付方式 |
| Webhook 未收到 | 前端轮询订单状态（最多 10 次） |
| Webhook 重复 | 幂等键处理（以 gateway_transaction_id 去重） |

### 5.7 支付超时处理

```
定时任务（每 5 分钟）:
  SELECT orders WHERE payment_status = 'pending' AND created_at < NOW() - 30min
  FOR EACH expired order:
    BEGIN TX
      UPDATE orders SET status = 'cancelled', cancelled_at = NOW()
      UPDATE skus SET reserved_stock = reserved_stock - quantity
      WHERE id = ? AND reserved_stock >= quantity
    COMMIT
```

### 5.8 退款

**API：**

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/voyara/refunds` | 买家 | 申请退款 |
| PUT | `/voyara/refunds/:id/approve` | 管理员/卖家 | 同意退款 |
| PUT | `/voyara/refunds/:id/reject` | 管理员/卖家 | 驳回退款 |

**退款处理流程：**

```
买家申请 → 卖家审核 → 同意 → 平台执行退款 → 资金原路返回
                                   → 恢复库存（如果未发货）
                                   → 更新商品销量
```

**退款方式与对应处理：**

| 支付方式 | 退款接口 | 到账时间 |
|----------|----------|----------|
| Stripe | Refund API | 5-10 工作日 |
| PayPal | Refund API | 即时-7天 |
| 支付宝 | 退款 API | 即时 |
| 微信支付 | 退款 API | 1-3 工作日 |

### 5.9 部分退款

- 卖家可指定部分退款金额
- 订单中部分商品退货 → 按比例退款
- 部分退款后订单标记为 `partial_refunded`
- 全部退款完成后标记为 `refunded`

### 5.10 对账

```
每日自动对账:
  1. 从 Stripe/PayPal 拉取前一日交易记录
  2. 与本地 voyara_payments 比对
  3. 匹配 → reconciliation_status = 'matched'
  4. 不匹配 → reconciliation_status = 'disputed' → 告警

差异场景:
  - 本地有、网关无：可能是测试数据，需人工确认
  - 网关有、本地无：可能是漏处理 Webhook，需补单
  - 金额不一致：需人工核查
```

### 5.11 金额精度

| 规则 | 说明 |
|------|------|
| 存储精度 | `DECIMAL(12,4)` 内部存储，4位小数 |
| 显示精度 | 前端按币种自动保留合适小数位（USD=2, JPY=0） |
| 计算精度 | 数据库层面使用 DECIMAL，Go 层面使用 `int64`（分/最小单位） |
| 避免浮点 | 任何金额计算禁用 float/double |

---

## 6. 订单系统

### 6.1 订单编号

```sql
-- 格式: V + 日期(YYMMDD) + 序列号(6位)
-- 示例: V260513000001
-- 由数据库自增ID + Redis 计数器生成，保证唯一且有时序
```

### 6.2 订单状态机

```
                      ┌───────────┐
                      │ 待支付     │
                      │ (pending) │
                      └─────┬─────┘
                            │
               ┌────────────┴────────────┐
               │ 支付成功                 │ 支付超时/取消
               ▼                         ▼
          ┌─────────┐             ┌───────────┐
          │ 已支付   │             │ 已取消     │
          │ (paid)  │             │ (cancelled)│
          └────┬────┘             └───────────┘
               │ 卖家发货                  ↑
               ▼                          │
          ┌─────────┐              ┌──────┴──────┐
          │ 已发货   │              │ 退款中/已退款│
          │(shipped) │              │refunding/   │
          └────┬────┘              │ refunded    │
               │ 买家确认收货       └─────────────┘
               ▼
          ┌─────────┐
          │ 已完成   │
          │(delivered)│
          └────┬────┘
               │ 申请售后
               ▼
          ┌─────────┐
          │ 售后中   │
          │(after_sale)│
          └─────────┘
```

### 6.3 数据库 `voyara_orders`

```sql
CREATE TABLE voyara_orders (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_no          VARCHAR(20) NOT NULL UNIQUE COMMENT '订单编号',
  buyer_id          BIGINT UNSIGNED NOT NULL,
  seller_id         BIGINT UNSIGNED NOT NULL,
  
  -- 订单状态
  payment_status    ENUM('pending','paid','refunded','partial_refunded','cancelled') NOT NULL DEFAULT 'pending',
  shipping_status   ENUM('pending','shipped','delivered') NOT NULL DEFAULT 'pending',
  after_sale_status ENUM('none','requested','approved','rejected','completed') NOT NULL DEFAULT 'none',
  
  -- 金额
  subtotal          DECIMAL(12,2) NOT NULL COMMENT '商品总价',
  shipping_fee      DECIMAL(12,2) NOT NULL DEFAULT 0,
  discount_amount   DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '优惠金额',
  coupon_id         BIGINT UNSIGNED,
  grand_total       DECIMAL(12,2) NOT NULL COMMENT '应付总额',
  currency          VARCHAR(3) NOT NULL DEFAULT 'USD',
  
  -- 快照（下单时锁定，防止数据变更影响订单）
  snapshot_product  JSON COMMENT '商品信息快照',
  snapshot_price    JSON COMMENT '价格快照',
  snapshot_address  JSON COMMENT '收货地址快照',
  snapshot_coupon   JSON COMMENT '优惠券快照',
  
  -- 时间戳
  paid_at           DATETIME,
  shipped_at        DATETIME,
  delivered_at      DATETIME,
  cancelled_at      DATETIME,
  cancel_reason     VARCHAR(500),
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (buyer_id) REFERENCES voyara_users(id),
  FOREIGN KEY (seller_id) REFERENCES voyara_sellers(id),
  INDEX idx_buyer (buyer_id),
  INDEX idx_seller (seller_id),
  INDEX idx_order_no (order_no),
  INDEX idx_payment_status (payment_status),
  INDEX idx_created (created_at)
);
```

### 6.4 订单明细表 `voyara_order_items`

```sql
CREATE TABLE voyara_order_items (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id      BIGINT UNSIGNED NOT NULL,
  sku_id        BIGINT UNSIGNED NOT NULL,
  spu_id        BIGINT UNSIGNED NOT NULL,
  title         VARCHAR(500) NOT NULL COMMENT '下单时商品名（快照）',
  spec_names    JSON COMMENT '下单时规格（快照）',
  image         VARCHAR(500) COMMENT '商品图片（快照）',
  price         DECIMAL(12,2) NOT NULL COMMENT '下单时单价（快照）',
  quantity      INT UNSIGNED NOT NULL,
  subtotal      DECIMAL(12,2) NOT NULL,
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  INDEX idx_order (order_id)
);
```

### 6.5 订单快照

下单时冻结以下数据到订单，防止卖家修改商品后影响已下单订单：

```json
{
  "snapshot_product": {
    "title": "iPhone 14",
    "spec_names": {"颜色": "黑色", "存储": "128G"},
    "image": "https://s3...",
    "seller_name": "TechShop",
    "seller_id": 5
  },
  "snapshot_price": {
    "unit_price": 699.00,
    "quantity": 1,
    "shipping_fee": 15.00,
    "discount": -50.00,
    "total": 664.00
  },
  "snapshot_address": {
    "recipient_name": "John Doe",
    "phone": "+1234567890",
    "country": "United States",
    "state": "CA",
    "city": "San Francisco",
    "street": "123 Main St",
    "zip_code": "94102"
  },
  "snapshot_coupon": {
    "code": "NEW50",
    "type": "fixed",
    "value": 50.00,
    "description": "新人优惠券满500减50"
  }
}
```

### 6.6 发票信息

```sql
ALTER TABLE voyara_orders
  ADD COLUMN invoice_type ENUM('none','personal','business') NOT NULL DEFAULT 'none',
  ADD COLUMN invoice_title VARCHAR(200),
  ADD COLUMN invoice_tax_id VARCHAR(50),
  ADD COLUMN invoice_email VARCHAR(200),
  ADD COLUMN invoice_status ENUM('pending','issued','failed') NOT NULL DEFAULT 'pending';
```

**API：**

| 方法 | 路径 | 说明 |
|------|------|------|
| PUT | `/voyara/orders/:id/invoice` | 填写/修改发票信息 |
| GET | `/voyara/orders/:id/invoice` | 获取发票信息 |

### 6.7 退款记录表 `voyara_refunds`

```sql
CREATE TABLE voyara_refunds (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id          BIGINT UNSIGNED NOT NULL,
  order_item_id     BIGINT UNSIGNED COMMENT '部分退款对应具体商品项',
  buyer_id          BIGINT UNSIGNED NOT NULL,
  seller_id         BIGINT UNSIGNED NOT NULL,
  refund_no         VARCHAR(20) NOT NULL UNIQUE COMMENT '退款编号 RF+日期+序列',
  refund_type       ENUM('full','partial') NOT NULL DEFAULT 'full',
  reason            VARCHAR(500) NOT NULL,
  amount            DECIMAL(12,2) NOT NULL COMMENT '退款金额',
  status            ENUM('pending','approved','rejected','completed','failed') NOT NULL DEFAULT 'pending',
  reviewer_id       BIGINT UNSIGNED COMMENT '审核人（管理员/卖家）',
  review_note       VARCHAR(500),
  gateway_refund_id VARCHAR(255) COMMENT '支付网关退款ID',
  completed_at      DATETIME,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  FOREIGN KEY (buyer_id) REFERENCES voyara_users(id),
  INDEX idx_order (order_id)
);
```

### 6.8 售后记录表 `voyara_after_sales`

```sql
CREATE TABLE voyara_after_sales (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id          BIGINT UNSIGNED NOT NULL,
  order_item_id     BIGINT UNSIGNED,
  buyer_id          BIGINT UNSIGNED NOT NULL,
  seller_id         BIGINT UNSIGNED NOT NULL,
  type              ENUM('return','exchange','repair','refund_only') NOT NULL,
  reason            VARCHAR(500) NOT NULL,
  description       TEXT,
  images            JSON COMMENT '凭证图片',
  status            ENUM('pending','approved','rejected','completed','cancelled') NOT NULL DEFAULT 'pending',
  shipping_status   ENUM('waiting_return','returned','waiting_send','resent','none') NOT NULL DEFAULT 'none',
  tracking_number   VARCHAR(200) COMMENT '退货物流单号',
  result_note       VARCHAR(500) COMMENT '处理结果说明',
  completed_at      DATETIME,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  INDEX idx_order (order_id)
);
```

### 6.9 物流信息

集成在订单详情中（详见第 8 节），在 `voyara_orders` 基础上可拆分：

```sql
-- 拆单发货表
CREATE TABLE voyara_order_shipments (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id          BIGINT UNSIGNED NOT NULL,
  seller_id         BIGINT UNSIGNED NOT NULL,
  shipment_no       VARCHAR(30) NOT NULL UNIQUE COMMENT '发货编号',
  carrier           VARCHAR(100) NOT NULL COMMENT '物流公司',
  tracking_number   VARCHAR(200) NOT NULL COMMENT '物流单号',
  tracking_url      VARCHAR(500),
  shipped_at        DATETIME NOT NULL,
  estimated_delivery DATETIME,
  FOREIGN KEY (order_id) REFERENCES voyara_orders(id),
  INDEX idx_order (order_id)
);
```

### 6.10 订单 API

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/voyara/orders` | 买家 | 订单列表（按状态筛选、分页） |
| GET | `/voyara/orders/:id` | 买家/卖家 | 订单详情 |
| POST | `/voyara/orders` | 买家 | 创建订单（幂等） |
| POST | `/voyara/orders/:id/cancel` | 买家 | 取消订单（未支付前） |
| POST | `/voyara/orders/:id/pay` | 买家 | 发起支付 |
| PUT | `/voyara/orders/:id/ship` | 卖家 | 发货 |
| POST | `/voyara/orders/:id/confirm` | 买家 | 确认收货 |
| GET | `/voyara/seller/orders` | 卖家 | 卖家订单列表 |
| GET | `/voyara/orders/:id/tracking` | 买家/卖家 | 物流轨迹 |

---

## 7. 库存系统

### 7.1 库存模型

```
                   ┌──────────────┐
                   │  真实库存     │
                   │ physical_stk │
                   └──────┬───────┘
                          │
          ┌───────────────┴───────────────┐
          │ 下单锁定                       │
          ▼                               ▼
   ┌──────────────┐             ┌────────────────┐
   │ 锁定库存      │             │ 可售库存        │
   │ reserved_stk │             │ available_stk   │
   │ (买家已下单   │             │ = physical -    │
   │  未支付)     │             │   reserved      │
   └──────┬───────┘             └────────────────┘
          │
          ├── 支付成功 ──→ 从 physical 扣减（库存真正减少）
          ├── 支付超时 ──→ reserved_stock 回滚
          └── 取消订单 ──→ reserved_stock 回滚
```

### 7.2 库存字段

```sql
-- SKU 表库存字段 (已在 voyara_skus 中定义)
-- stock              INT UNSIGNED 实际库存
-- reserved_stock     INT UNSIGNED 已预占（锁定）库存
-- sold_count         INT UNSIGNED 已售数量

-- 计算可售库存: stock - reserved_stock
-- 约束: reserved_stock <= stock
```

### 7.3 库存操作

| 操作 | SQL | 说明 |
|------|-----|------|
| 下单锁库存 | `UPDATE skus SET reserved_stock=reserved_stock+? WHERE id=? AND (stock - reserved_stock) >= ?` | 原子操作，自检库存 |
| 支付成功扣库存 | `UPDATE skus SET stock=stock-?, reserved_stock=reserved_stock-?, sold_count=sold_count+? WHERE id=?` | 支付确认后执行 |
| 取消/超时释放 | `UPDATE skus SET reserved_stock=reserved_stock-? WHERE id=? AND reserved_stock >= ?` | 释放锁定 |
| 退款恢复库存 | `UPDATE skus SET stock=stock+? WHERE id=?` | 仅未发货退款恢复 |

### 7.4 秒杀场景

```
针对特价/秒杀活动的库存隔离方案：

voyara_flash_sale_stock:
  sku_id            BIGINT UNSIGNED
  flash_stock       INT UNSIGNED      -- 秒杀专用库存
  flash_reserved    INT UNSIGNED      -- 秒杀已锁定
  flash_sold        INT UNSIGNED      -- 秒杀已售

秒杀库存与普通库存隔离，互不影响
```

### 7.5 多仓库（二期）

```sql
CREATE TABLE voyara_warehouses (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  seller_id   BIGINT UNSIGNED NOT NULL,
  name        VARCHAR(200) NOT NULL,
  country     VARCHAR(100) NOT NULL,
  city        VARCHAR(100),
  address     TEXT,
  status      ENUM('active','inactive') NOT NULL DEFAULT 'active',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- SKU 单仓库存 → 改为多仓
CREATE TABLE voyara_warehouse_stock (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  sku_id        BIGINT UNSIGNED NOT NULL,
  warehouse_id  BIGINT UNSIGNED NOT NULL,
  stock         INT UNSIGNED NOT NULL DEFAULT 0,
  reserved      INT UNSIGNED NOT NULL DEFAULT 0,
  FOREIGN KEY (sku_id) REFERENCES voyara_skus(id),
  FOREIGN KEY (warehouse_id) REFERENCES voyara_warehouses(id),
  UNIQUE KEY uk_sku_warehouse (sku_id, warehouse_id)
);
```

### 7.6 超卖防护三层保障

| 层级 | 措施 |
|------|------|
| 数据库层 | 原子 UPDATE 自检 `(stock - reserved_stock) >= ?` |
| 应用层 | 下单事务中 `SELECT ... FOR UPDATE` 锁行 |
| 缓存层（可选） | Redis 原子 DECR 预检 + 兜底数据库 |

---

## 8. 物流与配送

### 8.1 运费模板

**数据库 `voyara_shipping_templates`：**

```sql
CREATE TABLE voyara_shipping_templates (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  seller_id         BIGINT UNSIGNED NOT NULL,
  name              VARCHAR(200) NOT NULL COMMENT '模板名称',
  template_type     ENUM('fixed','weight','piece','free') NOT NULL DEFAULT 'fixed',
  
  -- 固定运费
  fixed_fee         DECIMAL(10,2) DEFAULT 0,
  
  -- 按件计费
  first_piece       INT UNSIGNED DEFAULT 0 COMMENT '首件数',
  first_piece_fee   DECIMAL(10,2) DEFAULT 0,
  extra_piece_fee   DECIMAL(10,2) DEFAULT 0 COMMENT '续件费用',
  
  -- 包邮规则
  free_shipping_threshold DECIMAL(12,2) DEFAULT NULL COMMENT '满额包邮',
  
  -- 适用范围
  shipping_regions  JSON COMMENT '适用区域，格式：[{country, state?}]',
  estimated_days    VARCHAR(50) COMMENT '预计配送天数',
  
  status            ENUM('active','inactive') NOT NULL DEFAULT 'active',
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  FOREIGN KEY (seller_id) REFERENCES voyara_sellers(id),
  INDEX idx_seller (seller_id)
);
```

**每个商品关联运费模板：**

```sql
ALTER TABLE voyara_spus
  ADD COLUMN shipping_template_id BIGINT UNSIGNED,
  ADD COLUMN weight_kg DECIMAL(8,2) COMMENT '重量（kg），用于按重计费',
  ADD FOREIGN KEY (shipping_template_id) REFERENCES voyara_shipping_templates(id);
```

### 8.2 运费计算逻辑

```
下单时计算运费:
  FOR EACH seller_group:
    IF 商品总额 >= 运费模板.free_shipping_threshold:
      shipping_fee = 0
    ELSE:
      IF template_type = 'fixed':
        shipping_fee = fixed_fee
      ELSE IF template_type = 'piece':
        total_pieces = SUM(order_items.quantity)
        IF total_pieces <= first_piece:
          shipping_fee = first_piece_fee
        ELSE:
          shipping_fee = first_piece_fee + (total_pieces - first_piece) * extra_piece_fee
      ELSE IF template_type = 'weight':
        total_weight = SUM(item.weight * item.quantity)
        shipping_fee = first_weight_fee + max(0, total_weight - first_weight) * extra_weight_fee
      ELSE IF template_type = 'free':
        shipping_fee = 0

跨卖家不合并运费: 每个卖家单独计算
```

### 8.3 物流公司

```sql
CREATE TABLE voyara_carriers (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name        VARCHAR(200) NOT NULL COMMENT '物流公司名',
  code        VARCHAR(50) NOT NULL UNIQUE COMMENT '编码（用于对接API）',
  tracking_url_template VARCHAR(500) COMMENT '物流追踪URL模板 {tracking_no}',
  status      ENUM('active','inactive') NOT NULL DEFAULT 'active'
);

INSERT INTO voyara_carriers (name, code, tracking_url_template) VALUES
  ('DHL', 'dhl', 'https://www.dhl.com/track/{tracking_no}'),
  ('FedEx', 'fedex', 'https://www.fedex.com/track/{tracking_no}'),
  ('UPS', 'ups', 'https://www.ups.com/track/{tracking_no}'),
  ('USPS', 'usps', 'https://tools.usps.com/go/TrackConfirmAction?tLabels={tracking_no}'),
  ('其他物流', 'other', NULL);
```

### 8.4 物流轨迹

```sql
CREATE TABLE voyara_tracking_events (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  shipment_id   BIGINT UNSIGNED NOT NULL,
  status        VARCHAR(100) COMMENT '状态: picked_up, in_transit, customs, delivered',
  location      VARCHAR(200),
  description   TEXT,
  event_time    DATETIME NOT NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (shipment_id) REFERENCES voyara_order_shipments(id),
  INDEX idx_shipment (shipment_id)
);
```

**物流追踪实现方式：**
- 初期：卖家手动填写物流单号 + 提供追踪链接
- 后期：集成第三方物流 API（AfterShip、17Track）自动拉取轨迹

### 8.5 发货后台

- 卖家在订单详情页点击"发货"
- 选择物流公司、填写物流单号
- 支持拆单发货（部分商品先发、多个包裹）
- 发货后订单状态变更为 `shipped`
- 系统推送发货通知给买家

### 8.6 拆单发货

```
一个订单可能包含多个商品项：
  - 场景1：部分商品先备好，部分还需时间 → 拆单
  - 场景2：商品存放在不同仓库 → 多仓发货
  - 场景3：大件商品分多个包裹 → 多包裹
  
voyara_order_shipments 支持一订单多条发货记录
```

### 8.7 自提 / 同城配送（二期）

```sql
ALTER TABLE voyara_spus
  ADD COLUMN pickup_available TINYINT(1) NOT NULL DEFAULT 0,
  ADD COLUMN local_delivery_available TINYINT(1) NOT NULL DEFAULT 0,
  ADD COLUMN local_delivery_fee DECIMAL(10,2) DEFAULT 0,
  ADD COLUMN local_delivery_radius_km INT UNSIGNED DEFAULT 0;
```

- 自提：卖家设置自提地址，买家下单后自行取货
- 同城配送：基于买家地址与卖家地址的距离计算是否在配送范围内

---

## 9. 促销与营销

### 9.1 优惠券系统

**数据库 `voyara_coupons`：**

```sql
CREATE TABLE voyara_coupons (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code              VARCHAR(50) NOT NULL UNIQUE COMMENT '优惠码',
  type              ENUM('fixed','percentage','free_shipping') NOT NULL,
  value             DECIMAL(12,2) NOT NULL COMMENT '固定金额或折扣百分比',
  min_purchase      DECIMAL(12,2) DEFAULT 0 COMMENT '最低消费',
  max_discount      DECIMAL(12,2) DEFAULT NULL COMMENT '最大折扣（百分比券用）',
  
  -- 适用范围
  scope             ENUM('global','category','seller','product') NOT NULL DEFAULT 'global',
  scope_ids         JSON COMMENT '适用范围ID列表',
  
  -- 使用限制
  usage_limit       INT UNSIGNED DEFAULT NULL COMMENT '总发放数量',
  usage_per_user    INT UNSIGNED DEFAULT 1 COMMENT '每人限用次数',
  used_count        INT UNSIGNED NOT NULL DEFAULT 0,
  
  -- 有效期
  valid_from        DATETIME NOT NULL,
  valid_until       DATETIME NOT NULL,
  
  status            ENUM('active','expired','disabled') NOT NULL DEFAULT 'active',
  created_by        BIGINT UNSIGNED COMMENT '创建人（管理员）',
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**用户领券表 `voyara_user_coupons`：**

```sql
CREATE TABLE voyara_user_coupons (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL,
  coupon_id   BIGINT UNSIGNED NOT NULL,
  used_at     DATETIME COMMENT '使用时间',
  order_id    BIGINT UNSIGNED COMMENT '使用的订单',
  status      ENUM('unused','used','expired') NOT NULL DEFAULT 'unused',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  FOREIGN KEY (coupon_id) REFERENCES voyara_coupons(id),
  INDEX idx_user (user_id)
);
```

### 9.2 满减 / 折扣

- 在优惠券系统基础上增加活动维度
- 满减：满 $100 减 $10（`fixed` 类型 + `min_purchase=100`）
- 折扣：全场 8 折（`percentage` 类型 + `value=20`）
- 可与优惠券叠加使用（需配置叠加规则）

### 9.3 秒杀活动

```sql
CREATE TABLE voyara_flash_sales (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title           VARCHAR(200) NOT NULL,
  description     TEXT,
  
  -- 秒杀时间
  start_time      DATETIME NOT NULL,
  end_time        DATETIME NOT NULL,
  
  status          ENUM('pending','active','ended','cancelled') NOT NULL DEFAULT 'pending',
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE voyara_flash_sale_items (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  flash_sale_id   BIGINT UNSIGNED NOT NULL,
  sku_id          BIGINT UNSIGNED NOT NULL,
  flash_price     DECIMAL(12,2) NOT NULL COMMENT '秒杀价',
  flash_stock     INT UNSIGNED NOT NULL COMMENT '秒杀专用库存',
  flash_sold      INT UNSIGNED NOT NULL DEFAULT 0,
  per_user_limit  INT UNSIGNED NOT NULL DEFAULT 1,
  FOREIGN KEY (flash_sale_id) REFERENCES voyara_flash_sales(id),
  FOREIGN KEY (sku_id) REFERENCES voyara_skus(id)
);
```

**秒杀库存逻辑：** 秒杀使用独立库存池，不影响普通库存

### 9.4 拼团（二期）

- 多人成团享折扣价
- 发起拼团 → 分享链接 → 邀请好友参团
- 规定时间内未成团 → 自动退款

### 9.5 会员价

```
基于用户会员等级的自动折扣:
  Silver:   原价 - 5%
  Gold:     原价 - 10%
  Platinum: 原价 - 15%
  
SKU 可单独设置会员价覆盖默认折扣
```

### 9.6 积分系统

```sql
ALTER TABLE voyara_users
  ADD COLUMN points        INT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN total_earned  INT UNSIGNED NOT NULL DEFAULT 0;

CREATE TABLE voyara_points_log (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL,
  points      INT NOT NULL COMMENT '正=收入，负=支出',
  reason      VARCHAR(200) NOT NULL COMMENT '原因: purchase, sign_in, review, etc.',
  reference_id VARCHAR(50) COMMENT '关联ID（订单号等）',
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  INDEX idx_user (user_id)
);
```

**积分规则：**
- 消费 $1 = 1 积分
- 签到 = 5 积分/天
- 写评价 = 20 积分
- 100 积分 = $1（下单时抵扣）

### 9.7 赠品

- 满额赠：订单满一定金额自动赠送
- 买赠：购买指定商品赠送关联商品
- 赠品在订单中单独展示，价格显示为 $0

### 9.8 推荐码

```sql
ALTER TABLE voyara_users
  ADD COLUMN referral_code VARCHAR(20) UNIQUE,
  ADD COLUMN referred_by BIGINT UNSIGNED;
```

- 老用户生成推荐码分享给新用户
- 新用户注册时填写推荐码
- 新用户首单后推荐人获得奖励（优惠券/积分）

### 9.9 营销 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/voyara/coupons/available` | 可用优惠券列表 |
| POST | `/voyara/coupons/claim` | 领取优惠券 `{code}` |
| GET | `/voyara/flash-sales` | 秒杀活动列表 |
| GET | `/voyara/flash-sales/:id/items` | 秒杀商品列表 |
| POST | `/voyara/flash-sales/:id/buy` | 秒杀下单（与普通下单不同接口） |

---

## 10. 后台管理系统

### 10.1 平台管理员系统

**路由：`/voyara/admin/*`，需要 `role = 'admin'` 权限**

```sql
-- 管理员表（与用户表关联，role = 'admin' 标识）
ALTER TABLE voyara_users
  ADD COLUMN role ENUM('user','seller','admin') NOT NULL DEFAULT 'user';
```

### 10.2 后台功能模块

| 路由 | 功能 | 说明 |
|------|------|------|
| `/voyara/admin` | Dashboard | 数据概览：今日订单数、销售额、新用户数、活跃卖家数 |
| `/voyara/admin/users` | 用户管理 | 列表、搜索、禁用/解禁、查看详情 |
| `/voyara/admin/sellers` | 卖家管理 | 审核申请、查看店铺、下架商品、封禁 |
| `/voyara/admin/products` | 商品管理 | 列表、审核上架、下架违规商品、编辑 |
| `/voyara/admin/categories` | 分类管理 | 增删改分类、排序、关联规格模板 |
| `/voyara/admin/brands` | 品牌管理 | 增删改品牌 |
| `/voyara/admin/orders` | 订单管理 | 查看所有订单、处理退款 |
| `/voyara/admin/refunds` | 退款管理 | 待审核退款列表、处理退款 |
| `/voyara/admin/coupons` | 优惠券管理 | 创建、发放、停用优惠券 |
| `/voyara/admin/shipping` | 物流管理 | 物流公司配置 |
| `/voyara/admin/reviews` | 评价管理 | 审核商品评价（通过/驳回） |
| `/voyara/admin/content` | 内容管理 | 首页 Banner、活动页、公告 |
| `/voyara/admin/stats` | 数据统计 | 图表：销售额趋势、用户增长、商品排行 |
| `/voyara/admin/admins` | 权限管理 | 管理员账号管理、角色分配 |
| `/voyara/admin/logs` | 操作日志 | 所有敏感操作审计日志 |

### 10.3 操作审计日志

```sql
CREATE TABLE voyara_audit_logs (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  admin_id      BIGINT UNSIGNED NOT NULL,
  action        VARCHAR(100) NOT NULL COMMENT '操作: approve_product, disable_user 等',
  target_type   VARCHAR(50) NOT NULL COMMENT '对象类型: product, user, order',
  target_id     BIGINT UNSIGNED NOT NULL,
  detail        JSON COMMENT '操作详情（变更前后对比）',
  ip_address    VARCHAR(45),
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (admin_id) REFERENCES voyara_users(id),
  INDEX idx_target (target_type, target_id),
  INDEX idx_admin (admin_id),
  INDEX idx_created (created_at)
);
```

### 10.4 商家后台

| 路由 | 功能 |
|------|------|
| `/voyara/seller/dashboard` | 卖家数据概览（商品数、订单数、收入） |
| `/voyara/seller/products` | 商品管理（列表、编辑、上下架） |
| `/voyara/seller/products/new` | 发布商品 |
| `/voyara/seller/orders` | 订单管理（待发货/已发货/已完成） |
| `/voyara/seller/refunds` | 退款管理（处理退款申请） |
| `/voyara/seller/coupons` | 店铺优惠券（创建、管理） |
| `/voyara/seller/shipping` | 运费模板管理 |
| `/voyara/seller/settlements` | 结算管理（查看收入、提现） |
| `/voyara/seller/profile` | 店铺信息编辑 |

### 10.5 商家结算

```sql
CREATE TABLE voyara_settlements (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  seller_id         BIGINT UNSIGNED NOT NULL,
  period_start      DATE NOT NULL COMMENT '结算周期开始',
  period_end        DATE NOT NULL COMMENT '结算周期结束',
  total_sales       DECIMAL(14,2) NOT NULL DEFAULT 0,
  commission_rate   DECIMAL(5,2) NOT NULL COMMENT '佣金比例 %',
  commission_amount DECIMAL(14,2) NOT NULL COMMENT '佣金金额',
  net_amount        DECIMAL(14,2) NOT NULL COMMENT '应结算金额 = 销售额 - 佣金',
  status            ENUM('pending','approved','paid','disputed') NOT NULL DEFAULT 'pending',
  paid_at           DATETIME,
  notes             TEXT,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (seller_id) REFERENCES voyara_sellers(id),
  INDEX idx_seller (seller_id),
  INDEX idx_period (period_start, period_end)
);
```

### 10.6 平台审核流程

```
商品审核:
  spu.status = 'pending' → 管理员审核
    → 通过 (status = 'active')
    → 驳回 (status = 'rejected', reject_reason)

卖家审核:
  seller.status = 'pending' → 管理员审核
    → 通过 (status = 'approved')
    → 驳回 (status = 'rejected', reject_reason)

评价审核:
  review.status = 'pending' → 管理员审核
    → 通过 (status = 'approved')
    → 驳回 (status = 'rejected')
```

---

## 11. 搜索与推荐

### 11.1 关键词搜索

**当前实现：** MySQL LIKE 查询（性能随数据增长急剧下降）

**升级路径：**

| 阶段 | 方案 | 适用阶段 |
|------|------|----------|
| Phase 1 | MySQL FULLTEXT 索引 + LIKE 回退 | 初期数据量 < 10 万 |
| Phase 2 | Elasticsearch / OpenSearch | 数据量增长后 |

**Phase 1 实现：**

```sql
ALTER TABLE voyara_spus ADD FULLTEXT INDEX ft_search (title, description, search_keywords);

-- 搜索查询
SELECT * FROM voyara_spus
WHERE status = 'active'
  AND MATCH(title, description, search_keywords) AGAINST(? IN BOOLEAN MODE)
ORDER BY sold_count DESC
LIMIT 20 OFFSET ?;
```

### 11.2 分类筛选

```sql
GET /voyara/spus?categoryId=5&brandId=3&minPrice=100&maxPrice=500
                     &sortBy=price_asc&page=1&pageSize=20
```

**支持筛选维度：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `categoryId` | int | 分类ID |
| `brandId` | int | 品牌ID |
| `sellerId` | int | 卖家ID |
| `minPrice` | decimal | 最低价 |
| `maxPrice` | decimal | 最高价 |
| `condition` | string | 成色：new/like_new/used/refurbished |
| `search` | string | 关键词 |
| `sortBy` | string | `price_asc`, `price_desc`, `newest`, `sold` |
| `page` | int | 页码，默认 1 |
| `pageSize` | int | 每页数，默认 20，最大 50 |

**响应：**

```json
{
  "data": [ /* SPU 数组 */ ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "totalItems": 156,
    "totalPages": 8
  },
  "filters": {
    "availableBrands": [ /* 当前搜索结果下的品牌聚合 */ ],
    "priceRange": { "min": 10, "max": 5000 }
  }
}
```

### 11.3 销量排序 / 新品排序

```sql
-- 销量排序
ORDER BY sold_count DESC, created_at DESC

-- 新品排序
ORDER BY created_at DESC

-- 价格排序
ORDER BY min_sku_price ASC/DESC
```

### 11.4 热门商品

```sql
-- 按近期销量 + 浏览量加权计算
SELECT spu_id, SUM(sold_count) * 0.7 + SUM(view_count) * 0.3 AS hot_score
FROM voyara_skus
GROUP BY spu_id
ORDER BY hot_score DESC
LIMIT 20;
```

### 11.5 相关商品

```
规则:
  - 同分类下价格相近的商品
  - 同品牌推荐
  - 买了该商品的用户也买了（协同过滤，二期）

API: GET /voyara/spus/:id/related → 返回 6-12 个相关商品
```

### 11.6 个性化推荐（二期）

- 基于用户浏览历史 + 购买历史
- 基于协同过滤算法
- 基于 Elasticsearch More Like This 查询

---

## 13. 合规与隐私

### 13.1 法律文档

| 文档 | 位置 | 内容 |
|------|------|------|
| 用户协议 (Terms of Service) | `/voyara/terms` | 平台使用条款、责任限制、争议解决 |
| 隐私政策 (Privacy Policy) | `/voyara/privacy` | 数据收集、使用、共享说明 |
| Cookie 政策 | `/voyara/cookies` | Cookie 使用说明 + 偏好设置 |
| 退款政策 | `/voyara/refund-policy` | 退款条件、流程、时限 |
| 发票政策 | `/voyara/invoice-policy` | 发票开具规则 |

**实现方式：**
- 注册时必须勾选同意用户协议 + 隐私政策
- 法律文档版本管理（`voyara_legal_docs` 表记录每次更新）
- 政策更新时通知用户重新确认

```sql
CREATE TABLE voyara_legal_docs (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  doc_type    ENUM('terms','privacy','cookies','refund','invoice') NOT NULL,
  version     INT UNSIGNED NOT NULL,
  content     TEXT NOT NULL,
  language    VARCHAR(10) NOT NULL DEFAULT 'en',
  effective_at DATETIME NOT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_version (doc_type, version, language)
);

CREATE TABLE voyara_legal_consents (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id     BIGINT UNSIGNED NOT NULL,
  doc_type    ENUM('terms','privacy','cookies') NOT NULL,
  version     INT UNSIGNED NOT NULL,
  consented_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ip_address  VARCHAR(45),
  FOREIGN KEY (user_id) REFERENCES voyara_users(id),
  INDEX idx_user (user_id)
);
```

### 13.2 数据保护

| 类别 | 措施 |
|------|------|
| 密码 | bcrypt 哈希，不存储明文 |
| 支付信息 | 不存储完整卡号，通过 Stripe/PayPal 处理 |
| 个人数据 | 用户可导出/删除自己的数据 (GDPR) |
| 日志脱敏 | 审计日志中隐藏邮箱、IP 部分内容 |
| 数据传输 | 全站 HTTPS，支付数据直连网关 |

### 13.3 支付合规

- Stripe/PayPal 均符合 PCI DSS Level 1
- 平台不接触原始卡号（通过 Stripe Elements / PayPal 按钮）
- 每笔交易保留完整的网关记录用于审计

### 13.4 税务合规

| 税务场景 | 处理方式 |
|----------|----------|
| 跨境 VAT/GST | 订单金额不含税，由买家自行承担进口税费 |
| 平台佣金 | 平台向卖家收取佣金，开具服务发票 |
| 交易记录 | 所有订单记录保留至少 7 年 |

### 13.5 跨境合规

- 商品发布时选择目标市场，自动标注该市场的进口限制
- 禁售品类（武器、毒品、活体动物等）在前端和后端双重校验
- 卖家需确认商品符合目标市场的法规要求
- 高风险品类（电子产品等）需额外认证信息

---

# Part C: 技术架构设计

## 14. 技术架构

### 12.1 整体架构

```
┌──────────────────────────────────────────────────────────┐
│                       前端 (React 19)                      │
│  VoyaraApp.tsx ── React Router ── Pages ── Components     │
│        │                                                    │
│        ▼                                                    │
│  API Client (fetch + Bearer Token)                         │
└──────────────────────────┬───────────────────────────────┘
                           │ HTTPS
                           ▼
┌──────────────────────────────────────────────────────────┐
│                   Nginx (反向代理)                         │
│  /voyara/api/*  → proxy_pass :8000/voyara/*               │
│  /voyara/static/* → 静态文件                               │
└──────────────────────────┬───────────────────────────────┘
                           │
                           ▼
┌──────────────────────────────────────────────────────────┐
│                GoFrame HTTP 服务器 (:8000)                 │
│                                                           │
│  ┌──────┬────────┬────────┬────────┬────────┬────────┐   │
│  │ Auth │ Product│ Cart   │ Order  │Payment │ Admin  │   │
│  │ Ctrl │  Ctrl  │  Ctrl  │  Ctrl  │  Ctrl  │  Ctrl  │   │
│  └──┬───┴───┬────┴───┬────┴───┬────┴───┬────┴───┬────┘   │
│     │       │        │        │        │        │         │
│  ┌──┴───┐ ┌┴────┐ ┌┴────┐ ┌┴────┐ ┌┴────┐ ┌┴────┐      │
│  │Auth  │ │Prod │ │Cart │ │Order│ │Paymt│ │Admin│      │
│  │Service│ │Svc  │ │Svc  │ │Svc  │ │Svc  │ │Svc  │      │
│  └──┬───┘ └┬────┘ └┬────┘ └┬────┘ └┬────┘ └┬────┘      │
│     │      │       │       │       │       │            │
│  ┌──┴──────┴───────┴───────┴───────┴───────┴────┐       │
│  │              Database (MySQL)                 │       │
│  └──────────────────────┬───────────────────────┘       │
│                         │                                │
│  ┌──────────────────────┴───────────────────────┐       │
│  │         External Services                     │       │
│  │  Stripe │ PayPal │ AWS SES │ AWS S3           │       │
│  └──────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────────┘
```

### 12.2 技术选型

| 组件 | 技术 | 说明 |
|------|------|------|
| 后端框架 | GoFrame v2.9.0 | 保持项目一致 |
| 语言 | Go 1.24 | |
| 前端框架 | React 19 + TypeScript | |
| 构建工具 | Vite 6 | |
| 数据库 | MySQL 8.0+ (InnoDB, utf8mb4) | |
| 缓存 (可选) | Redis 7+ | 会话缓存、速率限制、秒杀 |
| 文件存储 | AWS S3 / 兼容对象存储 | 商品图片 |
| 搜索 (二期) | Elasticsearch / OpenSearch | 全文搜索 |
| 消息队列 (可选) | RabbitMQ / AWS SQS | 异步任务、订单超时 |
| 支付 | Stripe + PayPal (P0) | 后续扩展 |
| 邮件 | AWS SES | 验证码、通知 |
| 短信 (二期) | AWS SNS / Twilio | 手机验证码 |
| 部署 | Docker + Nginx | 保持现有方式 |
| CI/CD (二期) | GitHub Actions | 自动化测试部署 |

### 12.3 后端目录结构

```
GoServer/Voyara/
├── api/v1/                      # API 请求/响应结构定义
│   ├── auth.go
│   ├── product.go
│   ├── cart.go                  # (新)
│   ├── order.go
│   ├── payment.go               # (新)
│   ├── category.go
│   ├── admin.go                 # (新)
│   ├── review.go                # (新)
│   ├── upload.go                # (新)
│   ├── coupon.go                # (新)
│   └── shipping.go              # (新)
│
├── core/
│   ├── controller/              # HTTP 控制器
│   │   ├── auth.go
│   │   ├── product.go
│   │   ├── cart.go              # (新)
│   │   ├── order.go
│   │   ├── payment.go           # (新)
│   │   ├── category.go
│   │   ├── admin.go             # (新)
│   │   ├── review.go            # (新)
│   │   ├── upload.go            # (新)
│   │   ├── coupon.go            # (新)
│   │   └── shipping.go          # (新)
│   │
│   ├── service/                 # 业务逻辑
│   │   ├── auth.go              # (重写 - JWT + bcrypt)
│   │   ├── product.go           # (重写 - SPU/SKU)
│   │   ├── cart.go              # (新)
│   │   ├── order.go             # (重写 - 完整流程)
│   │   ├── payment.go           # (新)
│   │   ├── payment_stripe.go    # (新 - Stripe 具体实现)
│   │   ├── payment_paypal.go    # (新 - PayPal 具体实现)
│   │   ├── category.go
│   │   ├── inventory.go         # (新 - 库存管理)
│   │   ├── shipping.go          # (新 - 运费模板)
│   │   ├── coupon.go            # (新)
│   │   ├── admin.go             # (新)
│   │   ├── review.go            # (新)
│   │   ├── upload.go            # (新 - S3 预签名)
│   │   ├── points.go            # (新 - 积分)
│   │   ├── audit.go             # (新 - 审计日志)
│   │   ├── idempotency.go       # (新 - 幂等键)
│   │   └── scheduler.go         # (新 - 定时任务)
│   │
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # (重写 - JWT)
│   │   ├── ratelimit.go         # (新 - 速率限制)
│   │   ├── csrf.go              # (新 - CSRF)
│   │   └── admin_required.go    # (新 - 管理员权限)
│   │
│   └── model/
│       └── voyara.go            # (重写 - 完整模型)
│
├── manifest/sql/
│   └── voyara.sql               # (重写 - 完整 DDL)
│
└── core/service/db.go           # 数据库连接
```

### 12.4 前端目录结构

```
Front/myrepapp-vite/src/Voyara/
├── VoyaraApp.tsx                # (更新 - 新增路由)
├── Voyara.css                   # (扩展)
│
├── api/
│   ├── client.ts                # (更新 - 拦截器增强)
│   ├── types.ts                 # (重写 - 完整类型)
│   ├── auth.ts
│   ├── product.ts
│   ├── cart.ts                  # (新)
│   ├── order.ts
│   ├── payment.ts               # (新)
│   ├── category.ts
│   ├── admin.ts                 # (新)
│   ├── review.ts                # (新)
│   └── coupon.ts                # (新)
│
├── contexts/
│   ├── LanguageContext.tsx       # (新 - 语言即时刷新)
│   ├── AuthContext.tsx           # (新 - 认证状态管理)
│   └── CartContext.tsx           # (新 - 购物车数量徽标)
│
├── pages/
│   ├── HomePage.tsx             # (重写 - 搜索/筛选/分页)
│   ├── ProductDetail.tsx        # (重写 - SKU选择器)
│   ├── CartPage.tsx             # (新)
│   ├── Checkout.tsx             # (重写 - 完整结算流程)
│   ├── PaymentPage.tsx          # (新 - 支付页)
│   ├── Orders.tsx               # (重写)
│   ├── OrderDetail.tsx          # (新)
│   ├── Login.tsx                # (更新)
│   ├── Register.tsx             # (更新 - 邮箱验证)
│   ├── Favorites.tsx            # (新)
│   ├── Addresses.tsx            # (新)
│   ├── Profile.tsx              # (新)
│   │
│   ├── seller/                  # 卖家端
│   │   ├── Dashboard.tsx
│   │   ├── NewProduct.tsx       # (重写 - SPU/SKU)
│   │   ├── MyProducts.tsx
│   │   ├── SellerOrders.tsx     # (新)
│   │   ├── SellerRefunds.tsx    # (新)
│   │   ├── ShippingTemplates.tsx# (新)
│   │   ├── Settlement.tsx       # (新)
│   │   └── ShopProfile.tsx      # (新)
│   │
│   └── admin/                   # 管理后台
│       ├── Dashboard.tsx        # (新)
│       ├── Users.tsx            # (新)
│       ├── Sellers.tsx          # (新)
│       ├── Products.tsx         # (新)
│       ├── Orders.tsx           # (新)
│       ├── Refunds.tsx          # (新)
│       ├── Coupons.tsx          # (新)
│       ├── Categories.tsx       # (新)
│       ├── Reviews.tsx          # (新)
│       └── AuditLogs.tsx        # (新)
│
├── components/
│   ├── Navbar.tsx               # (重写 - 语言切换+搜索+购物车徽标)
│   ├── Footer.tsx
│   ├── ProductCard.tsx          # (更新)
│   ├── SpecSelector.tsx         # (新 - 规格选择器)
│   ├── ImageUploader.tsx        # (新 - 图片上传组件)
│   ├── AddressSelector.tsx      # (新 - 地址选择器)
│   ├── CouponSelector.tsx       # (新 - 优惠券选择器)
│   └── Pagination.tsx           # (新 - 分页)
│
└── i18n/
    ├── index.ts                 # (重写 - Context 驱动)
    ├── en.json                  # (扩展)
    ├── fr.json
    ├── ar.json
    ├── ru.json
    └── zh.json                  # (扩展)
```

### 12.5 数据库部署

```yaml
# 在现有 MySQL 中创建独立数据库
CREATE DATABASE IF NOT EXISTS Voyara CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

**现有 docker-compose.yml 调整：**

```yaml
# 数据库连接配置移入容器环境变量
# GoServer/Voyara/core/service/db.go 支持环境变量
VOYARA_DB_HOST=${VOYARA_DB_HOST:-127.0.0.1}
VOYARA_DB_PORT=${VOYARA_DB_PORT:-13306}
VOYARA_DB_USER=${VOYARA_DB_USER:-repgameadmin}
VOYARA_DB_PASS=${VOYARA_DB_PASS:-repgameadmin}
VOYARA_DB_NAME=${VOYARA_DB_NAME:-Voyara}
```

### 12.6 缓存策略

```
适用场景:
  商品分类列表 → 缓存 5 分钟（变更不频繁）
  商品详情     → 缓存 2 分钟（价格/库存变更敏感）
  首页推荐     → 缓存 10 分钟
  
不使用缓存:
  购物车数据   → 强一致，每次从数据库拉取
  订单数据     → 强一致
  库存数据     → 使用数据库原子操作
```

### 12.7 定时任务

| 任务 | 间隔 | 说明 |
|------|------|------|
| 支付超时取消 | 每 5 分钟 | 取消 30 分钟未支付订单 |
| 自动确认收货 | 每 1 小时 | 发货后 30 天未确认 → 自动确认 |
| 清理过期优惠券 | 每日 | 标记过期优惠券 |
| 对账任务 | 每日 | 拉取支付网关交易比对 |
| 清理幂等键 | 每日 | 删除超过 24h 的幂等记录 |

### 12.8 消息队列（二期）

```
适用场景:
  发送邮件/短信通知 → 异步，不阻塞主流程
  订单超时取消      → 延迟消息
  支付成功通知      → 事件广播
  
可选技术:
  Go channel + 后台 worker（轻量，适合初期）
  RabbitMQ / AWS SQS（生产环境）
```

---

# Part D: 数据分析

## 15. 数据分析指标

### 13.1 核心指标

| 指标 | 定义 | 计算方式 |
|------|------|----------|
| 访问量 (PV) | 页面浏览次数 | 前端埋点统计 |
| 独立访客 (UV) | 独立用户数 | 按用户ID + IP 去重 |
| 转化率 | 下单用户 / 访客数 | 订单数 / UV × 100% |
| 加购率 | 加购用户 / 访客数 | 加购用户 / UV × 100% |
| 支付成功率 | 支付成功 / 支付发起 | paid_orders / total_orders × 100% |
| 退款率 | 退款订单 / 总订单 | refunded / total × 100% |
| 客单价 | 平均每单金额 | 总销售额 / 总订单数 |
| 复购率 | 再次购买用户比例 | 多次购买用户 / 总购买用户 × 100% |

### 13.2 数据统计表

```sql
CREATE TABLE voyara_stats_daily (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  stat_date       DATE NOT NULL UNIQUE,
  
  -- 用户
  new_users       INT UNSIGNED NOT NULL DEFAULT 0,
  new_sellers     INT UNSIGNED NOT NULL DEFAULT 0,
  total_users     INT UNSIGNED NOT NULL DEFAULT 0,
  
  -- 商品
  new_products    INT UNSIGNED NOT NULL DEFAULT 0,
  active_products INT UNSIGNED NOT NULL DEFAULT 0,
  
  -- 订单
  total_orders    INT UNSIGNED NOT NULL DEFAULT 0,
  paid_orders     INT UNSIGNED NOT NULL DEFAULT 0,
  refunded_orders INT UNSIGNED NOT NULL DEFAULT 0,
  
  -- 金额
  total_revenue   DECIMAL(14,2) NOT NULL DEFAULT 0,
  platform_fee    DECIMAL(14,2) NOT NULL DEFAULT 0,
  
  -- 流量
  page_views      INT UNSIGNED NOT NULL DEFAULT 0,
  unique_visitors INT UNSIGNED NOT NULL DEFAULT 0,
  
  created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 13.3 前端埋点

```typescript
// 初期简单埋点（通过 API 记录）
API: POST /voyara/track
{
  event: "page_view" | "product_view" | "add_to_cart" | "purchase",
  data: { productId?, category?, source? },
  timestamp: 2026-05-13T10:00:00Z
}
```

---

# Part E: 并发与数据一致性

## 16. 并发处理设计

### 14.1 三大并发冲突场景

```
场景1：超卖 ──── 同一商品只剩1件，两个买家同时结算
场景2：重复下单 ── 用户快速连点"提交订单"
场景3：Webhook 重试 ── Stripe/PayPal 因网络发送多次通知
```

### 14.2 下单库存扣减 — 乐观锁 + 原子 SQL

```sql
-- 下单时预占库存：原子 UPDATE + 自检
UPDATE voyara_skus 
SET reserved_stock = reserved_stock + ?
WHERE id = ? 
  AND (stock - reserved_stock) >= ?
  AND status = 'active';

-- 检查 RowsAffected
-- 0 → 库存不足，拒绝订单（返回具体商品名）
-- 1 → 成功
```

### 14.3 订单创建事务

```
下单事务 (isolation: REPEATABLE READ):
  BEGIN TX
    1. 幂等检查（Idempotency-Key）
    2. SELECT ... FOR UPDATE 锁住 SKU 行
    3. 验证每个 SKU 的 stock - reserved_stock >= 需求数量
    4. UPDATE skus SET reserved_stock += 需求数量 WHERE id IN (...)
    5. INSERT INTO orders
    6. INSERT INTO order_items
    7. 从购物车删除已购商品
    8. 记录幂等键
  COMMIT
```

### 14.4 支付完成 — 二阶段确认

```
Webhook 处理事务:
  BEGIN TX
    1. 幂等检查（gateway_transaction_id）
    2. UPDATE orders SET payment_status = 'paid', paid_at = NOW()
    3. UPDATE skus SET 
         stock = stock - quantity,
         reserved_stock = reserved_stock - quantity,
         sold_count = sold_count + quantity
       WHERE id IN (...)
    4. INSERT INTO payments (status = 'succeeded')
    5. 给卖家发通知（消息队列 / 邮件）
  COMMIT
```

### 14.5 幂等键机制

```go
// 幂等键存储
type IdempotencyStore struct {
    db *sql.DB
}

func (s *IdempotencyStore) CheckAndSet(key string, response interface{}) (exists bool, err error) {
    // INSERT IGNORE INTO voyara_idempotency_keys (idempotent_key, response, created_at)
    // 如果受影响行数 = 0 → 键已存在，返回已有 response
    // 如果受影响行数 = 1 → 首次请求，继续处理
}

// 幂等键清理（每日定时任务）
// DELETE FROM voyara_idempotency_keys WHERE created_at < NOW() - INTERVAL 24 HOUR
```

### 14.6 Webhook 幂等处理

```go
func HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    // 1. 验签
    payload, err := ioutil.ReadAll(r.Body)
    event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
    
    // 2. 使用 event.ID 作为幂等键
    exists, _ := idempotencyStore.CheckAndSet("stripe:"+event.ID, nil)
    if exists {
        w.WriteHeader(200) // 已处理，返回成功
        return
    }
    
    // 3. 处理支付成功逻辑
    switch event.Type {
    case "payment_intent.succeeded":
        // ... 更新订单
    }
}
```

### 14.7 支付超时释放

```go
func CancelExpiredOrders() {
    // 每 5 分钟运行一次
    rows, _ := db.Query(`
        SELECT id FROM voyara_orders 
        WHERE payment_status = 'pending' 
        AND created_at < NOW() - INTERVAL 30 MINUTE
        FOR UPDATE
    `)
    
    for rows.Next() {
        // 对每个过期订单启动事务
        tx, _ := db.Begin()
        // 1. UPDATE orders SET status = 'cancelled'
        // 2. UPDATE skus SET reserved_stock = reserved_stock - quantity
        // 3. COMMIT
    }
}
```

---

# Part F: 实施路线图

## 17. 分阶段实施计划

### Phase 1：基础加固（预计 2-3 周）

| 任务 | 文件涉及 | 工作量 |
|------|----------|--------|
| 1.1 JWT 替换自定义 token | `service/auth.go`, `middleware/auth.go` | 2 天 |
| 1.2 bcrypt 密码 + 旧哈希迁移 | `service/auth.go`, `model/voyara.go` | 1 天 |
| 1.3 邮箱格式验证 + AWS SES 集成 | `service/auth.go`, `service/email.go` | 2 天 |
| 1.4 注册验证码流程 | `api/v1/auth.go`, `controller/auth.go` | 1 天 |
| 1.5 速率限制中间件 | `middleware/ratelimit.go` | 1 天 |
| 1.6 CSRF 防护中间件 | `middleware/csrf.go` | 1 天 |
| 1.7 语言切换 Context 重写 | `contexts/LanguageContext.tsx`, `i18n/index.ts` | 1 天 |
| 1.8 数据库迁移脚本 | `manifest/sql/` | 1 天 |

### Phase 2：购物车 + 支付（预计 3-4 周）

| 任务 | 文件涉及 | 工作量 |
|------|----------|--------|
| 2.1 购物车后端 CRUD | `cart.go` (api/controller/service) | 2 天 |
| 2.2 购物车前页面 | `CartPage.tsx` | 2 天 |
| 2.3 Stripe 接入 | `payment_stripe.go` | 3 天 |
| 2.4 PayPal 接入 | `payment_paypal.go` | 3 天 |
| 2.5 Webhook 处理 + 幂等 | `payment.go` | 2 天 |
| 2.6 支付超时定时任务 | `scheduler.go` | 1 天 |
| 2.7 前端支付页面 | `PaymentPage.tsx` | 2 天 |
| 2.8 幂等键机制 | `idempotency.go` | 1 天 |

### Phase 3：SPU/SKU + 商品系统重构（预计 2-3 周）

| 任务 | 文件涉及 | 工作量 |
|------|----------|--------|
| 3.1 SPU/SKU 数据库建表 | `manifest/sql/` | 1 天 |
| 3.2 SPU CRUD 后端 | `product.go` (重写) | 3 天 |
| 3.3 SKU 管理后端 | `product.go` | 2 天 |
| 3.4 规格模板系统 | `category.go` | 2 天 |
| 3.5 品牌管理 | `brand.go` | 1 天 |
| 3.6 前端规格选择器 | `SpecSelector.tsx` | 2 天 |
| 3.7 前端发布商品页重写 | `NewProduct.tsx` | 2 天 |

### Phase 4：卖家/买家分离 + 订单管理（预计 2-3 周）

| 任务 | 文件涉及 | 工作量 |
|------|----------|--------|
| 4.1 卖家审核流程 | `admin.go`, `service/seller.go` | 2 天 |
| 4.2 订单状态机完善 | `order.go` (重写) | 3 天 |
| 4.3 订单快照系统 | `order.go` | 2 天 |
| 4.4 卖家不可自购逻辑 | `order.go`, `cart.go` | 1 天 |
| 4.5 退款/售后系统 | `refund.go` | 3 天 |
| 4.6 库存管理 | `inventory.go` | 2 天 |

### Phase 5：后台管理 + 物流（预计 3-4 周）

| 任务 | 文件涉及 | 工作量 |
|------|----------|--------|
| 5.1 管理员权限中间件 | `middleware/admin_required.go` | 1 天 |
| 5.2 管理后台各页面 | `pages/admin/*` | 5 天 |
| 5.3 运费模板系统 | `shipping.go` | 3 天 |
| 5.4 物流追踪 | `shipping.go`, `tracking.go` | 2 天 |
| 5.5 审计日志 | `audit.go` | 1 天 |
| 5.6 商家结算 | `settlement.go` | 2 天 |

### Phase 6：营销 + 增强功能（预计 3-4 周）

| 任务 | 文件涉及 | 工作量 |
|------|----------|--------|
| 6.1 优惠券系统 | `coupon.go` | 3 天 |
| 6.2 积分系统 | `points.go` | 2 天 |
| 6.3 秒杀活动 | `flash_sale.go` | 3 天 |
| 6.4 图片上传 (AWS S3) | `upload.go` | 2 天 |
| 6.5 评价系统 | `review.go` | 2 天 |
| 6.6 搜索升级 (FULLTEXT) | `product.go` | 2 天 |
| 6.7 第三方登录 | `auth.go` | 2 天 |
| 6.8 数据分析 | `stats.go` | 2 天 |

---

## 18. 项目估算

| 阶段 | 内容 | 预计时间 | 新增后端文件 | 新增前端文件 |
|------|------|----------|-------------|-------------|
| Phase 1 | 基础加固 | 2-3 周 | 6 | 3 |
| Phase 2 | 购物车+支付 | 3-4 周 | 10 | 4 |
| Phase 3 | 商品系统重构 | 2-3 周 | 6 | 5 |
| Phase 4 | 卖家/订单 | 2-3 周 | 5 | 4 |
| Phase 5 | 后台+物流 | 3-4 周 | 8 | 12 |
| Phase 6 | 营销+增强 | 3-4 周 | 10 | 6 |
| **合计** | | **15-21 周** | **~45** | **~34** |

---

## 附录 A：全局配置

```go
// GoServer/Voyara/core/config.go
type VoyaraConfig struct {
    JWTSecret         string `env:"VOYARA_JWT_SECRET"`
    JWTExpiry         time.Duration // 24h
    RefreshExpiry     time.Duration // 7d
    
    StripeSecretKey   string `env:"STRIPE_SECRET_KEY"`
    StripeWebhookSec  string `env:"STRIPE_WEBHOOK_SECRET"`
    
    PayPalClientID    string `env:"PAYPAL_CLIENT_ID"`
    PayPalSecretKey   string `env:"PAYPAL_SECRET_KEY"`
    PayPalWebhookID   string `env:"PAYPAL_WEBHOOK_ID"`
    
    AWSS3Bucket       string `env:"AWS_S3_BUCKET"`
    AWSRegion         string `env:"AWS_REGION"`
    AWSAccessKey      string `env:"AWS_ACCESS_KEY"`
    AWSSecretKey      string `env:"AWS_SECRET_KEY"`
    
    AWSSESFrom        string `env:"AWS_SES_FROM_EMAIL"`
    
    DBHost            string `env:"VOYARA_DB_HOST"`
    DBPort            string `env:"VOYARA_DB_PORT"`
    DBUser            string `env:"VOYARA_DB_USER"`
    DBPass            string `env:"VOYARA_DB_PASS"`
    DBName            string `env:"VOYARA_DB_NAME"`
}
```
