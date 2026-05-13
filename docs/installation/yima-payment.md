# 倚码支付配置与联调指南

本文档说明如何在 new-api 中启用倚码支付，并完成充值与订阅的联调验证。

当前接入范围包括：

1. 普通充值
2. 订阅购买
3. 支付方式 `yima_alipay`
4. 支付方式 `yima_wechat`

***

## 一、接入前提

启用前请先确认以下条件：

1. 你的 new-api 已经可以正常访问，且公网可以访问它的业务域名
2. 后台“系统设置”中的 `ServerAddress` 已配置为真实外网地址，而不是 `http://localhost:3000`
3. 你已经拿到倚码支付商户参数：
   - `mch_id`
   - `mch_key`
   - 支付网关基础地址，例如 `https://zf.rx.sc.cn`
4. 如果站点前面有 Nginx、CDN 或反向代理，回调路径没有被拦截、重写或限制方法

如果你的站点只能在内网访问，或者 `ServerAddress` 仍然是本地地址，支付页面可能可以拉起，但异步回调和支付完成跳转通常会失败。

***

## 二、后台需要配置的项目

默认前端后台中已经提供了倚码支付设置项，经典前端后台也同步支持。

需要填写的配置如下。

### 1. 必填项

#### `YimaEnabled`

是否启用倚码支付。

只有它为开启状态，并且下面三个关键字段都不为空时，前台才会出现倚码支付按钮。

#### `YimaPayAddress`

倚码支付基础地址。

示例：

```text
https://zf.rx.sc.cn
```

说明：

1. 不要拼 `/v1/merchant-openapi/pay/create`
2. 只填写域名或基础路径即可
3. 系统会自动去掉结尾多余的 `/`

#### `YimaMerchantId`

倚码商户 ID，对应上游文档中的 `mch_id`。

#### `YimaMerchantKey`

倚码商户密钥，对应上游文档中的 `mch_key`。

系统会使用它对请求参数做 MD5 签名，并对倚码回调进行验签。

### 2. 建议配置项

#### `YimaMinTopUp`

倚码支付按钮对应的最小充值数量。

说明：

1. 前台展示时会带出这个下限
2. 后端下单时也会校验这个值
3. 如果充值数量小于该值，接口会直接返回错误

### 3. 可选覆盖项

以下字段可以留空。留空时，系统会按默认规则自动生成地址。

#### `YimaNotifyUrl`

充值异步回调地址覆盖值。

留空时默认使用：

```text
{CustomCallbackAddress 或 ServerAddress}/api/user/yima/notify
```

其中：

1. 如果配置了 `CustomCallbackAddress`，优先使用它
2. 否则使用 `ServerAddress`

#### `YimaReturnUrl`

充值完成后的浏览器返回地址覆盖值。

留空时默认使用：

```text
{ServerAddress}/api/user/yima/return
```

#### `YimaSubscriptionReturnUrl`

订阅购买完成后的浏览器返回地址覆盖值。

留空时默认按以下优先级选择：

1. `YimaSubscriptionReturnUrl`
2. `YimaReturnUrl`
3. `{ServerAddress}/api/subscription/yima/return`

***

## 三、系统实际使用到的回调路径

当前后端已经暴露以下路径：

### 充值相关

```text
POST /api/user/yima/notify
GET  /api/user/yima/return
```

### 订阅相关

```text
POST /api/subscription/yima/notify
GET  /api/subscription/yima/return
```

### 通用 webhook 入口

```text
POST /api/yima/webhook
```

说明：

1. 实际下单时，充值默认会把 `notify_url` 传成 `/api/user/yima/notify`
2. 订阅默认会把 `notify_url` 传成 `/api/subscription/yima/notify`
3. `/api/yima/webhook` 目前兼容指向充值处理逻辑，如果你已经按默认下单流程接入，一般不需要单独填写这个地址

***

## 四、推荐配置流程

### 第一步：确认系统基础地址

先在后台确认：

1. `ServerAddress` 是公网可访问地址
2. 如果你的回调需要走另一个专用域名，再配置 `CustomCallbackAddress`

推荐示例：

```text
ServerAddress=https://api.example.com
CustomCallbackAddress=https://api.example.com
```

如果你把 `ServerAddress` 留成下面这种值：

```text
http://localhost:3000
```

那么支付完成后浏览器回跳通常会被带到本机地址，线上用户无法使用。

### 第二步：在后台填写倚码参数

建议至少填写：

```text
YimaEnabled=true
YimaPayAddress=https://zf.rx.sc.cn
YimaMerchantId=你的商户ID
YimaMerchantKey=你的商户密钥
YimaMinTopUp=1
```

如果你的部署结构比较标准，三个 URL 覆盖项通常可以先留空。

### 第三步：检查反向代理

如果你前面有 Nginx，请确认以下路径能正常转发到 new-api：

```text
/api/user/yima/notify
/api/user/yima/return
/api/subscription/yima/notify
/api/subscription/yima/return
```

同时确认：

1. `POST` 方法没有被拦截
2. 回调请求体没有被代理层改写
3. HTTPS 证书正常，外部可以访问

### 第四步：保存配置后检查前台按钮

保存成功后，前台充值区域应该会出现：

1. 支付宝（倚码）
2. 微信（倚码）

如果没有出现，优先检查：

1. `YimaEnabled` 是否已开启
2. `YimaPayAddress`、`YimaMerchantId`、`YimaMerchantKey` 是否为空
3. 前端是否已经更新到包含倚码支付的版本

***

## 五、联调验证步骤

推荐按下面顺序验证。

### 场景一：充值支付宝

1. 登录普通用户账号
2. 进入充值页面
3. 选择 `支付宝（倚码）`
4. 输入不小于 `YimaMinTopUp` 的充值数量
5. 提交支付
6. 浏览器应跳转到倚码返回的 `payment_url`
7. 完成支付后，等待页面回跳
8. 检查用户额度是否增加
9. 检查充值记录是否变为成功

### 场景二：充值微信

步骤与支付宝一致，只是支付方式改为 `微信（倚码）`。

### 场景三：订阅购买

1. 进入订阅购买弹窗
2. 选择 `yima_alipay` 或 `yima_wechat`
3. 发起支付
4. 浏览器应跳转到倚码支付页面
5. 完成支付后，检查订阅是否已生效

### 场景四：异步回调验证

支付成功后，重点确认以下结果：

1. 充值订单状态从 pending 变为 success
2. 订阅订单状态正确完成
3. 用户额度或订阅权益已经落库
4. 后端日志中没有验签失败或金额不匹配错误

***

## 六、排错建议

### 1. 前台看不到倚码支付按钮

优先排查：

1. `YimaEnabled` 未开启
2. `YimaPayAddress`、`YimaMerchantId`、`YimaMerchantKey` 其中有空值
3. 保存配置后没有刷新页面

### 2. 点击支付后提示“倚码支付未启用”

通常说明后端判定配置不完整。

检查：

1. `YimaEnabled=true`
2. `YimaPayAddress` 不为空
3. `YimaMerchantId` 不为空
4. `YimaMerchantKey` 不为空

### 3. 点击支付后提示“充值数量不能小于 X”

说明当前充值数量小于 `YimaMinTopUp`。

处理方式：

1. 提高充值数量
2. 或在后台调低 `YimaMinTopUp`

### 4. 支付页能打开，但支付成功后没有到账

重点检查：

1. 异步回调地址是否能被公网访问
2. 代理层是否放通 `POST /api/user/yima/notify`
3. 商户密钥是否填写错误，导致验签失败
4. 上游回调中的金额是否与本地下单金额一致

### 5. 支付完成后跳回了 `localhost`

说明 `ServerAddress` 仍然配置成了本地地址。

应改为：

```text
https://你的真实域名
```

### 6. 订阅成功但回跳地址不正确

如果你希望订阅和充值跳回不同页面，可以单独填写：

```text
YimaSubscriptionReturnUrl
```

否则系统会优先复用 `YimaReturnUrl`。

### 7. 日志里出现验签失败

优先检查：

1. `YimaMerchantKey` 是否与倚码后台完全一致
2. 代理层是否改写了表单参数
3. 上游是否把请求发到了错误的商户环境

***

## 七、建议的上线前检查清单

正式上线前，至少完成以下检查：

1. 充值支付宝成功一次
2. 充值微信成功一次
3. 订阅支付成功一次
4. 充值回调成功一次
5. 订阅回调成功一次
6. `ServerAddress` 已经改成正式域名
7. Nginx 或 CDN 没有拦截 `/api/.../yima/...` 路径
8. 商户密钥没有误填到测试环境或旧环境

***

## 八、补充说明

当前实现中：

1. 充值和订阅都会走倚码 OpenAPI 下单
2. 前端收到 `payment_url` 后会直接新开页面跳转
3. 支付方式标识固定为：

```text
yima_alipay
yima_wechat
```

如果后续还要接倚码的更多支付类型，建议继续沿用同一套实现方式扩展 `payment_method`。