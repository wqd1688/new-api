# new-api Open API 文档

本文档基于当前仓库中的路由实现整理，覆盖项目当前对外暴露的主要 HTTP API、OpenAI 兼容中继接口、用户接口、管理接口与运维接口。

适用代码范围：

- `router/relay-router.go`
- `router/video-router.go`
- `router/api-router.go`
- `router/dashboard.go`

## 1. 总览

这个项目同时提供两类 API：

1. AI 中继 API
   - 面向模型调用方
   - 主要兼容 OpenAI、Claude、Gemini、Midjourney、Suno、视频生成等接口风格
   - 主要路径前缀为 `/v1`、`/v1beta`、`/mj`、`/suno`、`/kling/v1`、`/jimeng`

2. 站点业务 API
   - 面向控制台、用户中心、后台管理、计费、渠道管理、日志查询等场景
   - 主要路径前缀为 `/api`

此外还有一组 OpenAI Dashboard 兼容账单接口：

- `/dashboard/billing/subscription`
- `/v1/dashboard/billing/subscription`
- `/dashboard/billing/usage`
- `/v1/dashboard/billing/usage`

## 2. 基础地址

示例：

- 站点业务 API：`https://your-domain/api`
- OpenAI 兼容 API：`https://your-domain/v1`
- Gemini 兼容 API：`https://your-domain/v1beta`
- Midjourney 中继：`https://your-domain/mj`

## 3. 认证方式

### 3.1 AI 中继接口认证

AI 中继接口主要使用令牌认证，兼容多种上游客户端习惯：

1. OpenAI 兼容
   - Header: `Authorization: Bearer sk-xxx`

2. Claude 兼容
   - Header: `x-api-key: sk-xxx`
   - 主要用于 `/v1/messages` 和 Claude 风格的 `/v1/models`

3. Gemini 兼容
   - Header: `x-goog-api-key: sk-xxx`
   - 或 query 参数：`?key=sk-xxx`
   - 主要用于 `/v1beta/models`、`/v1beta/openai/models`、`/v1beta/models/*path`

4. Realtime WebSocket
   - 通过 `Sec-WebSocket-Protocol` 传递 key
   - 示例：`openai-insecure-api-key.sk-xxx`

### 3.2 站点业务 API 认证

`/api` 下的控制台接口主要基于登录会话，部分接口也支持 Access Token。

注意：`UserAuth`、`AdminAuth`、`RootAuth` 类型接口除认证信息外，通常还要求请求头中带上：

- `New-Api-User: <当前用户 ID>`

### 3.3 只读 Token 查询接口

部分只读接口支持只校验令牌存在，不校验额度、过期与启停状态：

- `Authorization: Bearer sk-xxx`
- 代表接口：`/api/usage/token`、`/api/log/token`

## 4. 返回格式约定

### 4.1 AI 中继接口

AI 中继接口通常直接返回上游兼容格式：

- OpenAI 兼容 JSON
- Claude Messages JSON
- Gemini JSON
- Midjourney / Suno / 视频任务 JSON

错误时通常返回 OpenAI 风格错误对象：

```json
{
  "error": {
    "message": "模型价格未配置",
    "type": "new_api_error",
    "param": "",
    "code": "model_price_error"
  }
}
```

常见错误码包括：

- `invalid_request`
- `model_price_error`
- `access_denied`
- `model_not_found`
- `insufficient_user_quota`
- `pre_consume_token_quota_failed`
- `channel:no_available_key`
- `channel:invalid_key`

### 4.2 `/api` 业务接口

大多数业务接口使用如下包装结构：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

但也存在部分历史接口使用以下结构：

```json
{
  "message": "success",
  "data": {}
}
```

因此客户端对 `/api` 的调用不要只根据 HTTP 200 判断成功，建议同时检查：

- `success === true`
- 或 `message !== "error"`

## 5. OpenAI 兼容 AI 中继 API

### 5.1 模型查询

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/v1/models` | TokenAuth | 列出模型。会根据请求头自动适配 OpenAI、Claude、Gemini 风格 |
| GET | `/v1/models/:model` | TokenAuth | 获取单个模型信息 |
| GET | `/v1beta/models` | TokenAuth | Gemini 风格模型列表 |
| GET | `/v1beta/openai/models` | TokenAuth | Gemini OpenAI 兼容模型列表 |

### 5.2 文本与对话

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/v1/completions` | TokenAuth | OpenAI Completion 兼容接口 |
| POST | `/v1/chat/completions` | TokenAuth | OpenAI Chat Completions 兼容接口 |
| POST | `/pg/chat/completions` | UserAuth + Distribute | Playground 专用聊天接口 |
| POST | `/v1/messages` | TokenAuth | Claude Messages 兼容接口 |
| POST | `/v1/responses` | TokenAuth | OpenAI Responses 接口 |
| POST | `/v1/responses/compact` | TokenAuth | 紧凑响应版 Responses 接口 |
| POST | `/v1/moderations` | TokenAuth | 内容审核接口 |
| GET | `/v1/realtime` | TokenAuth | OpenAI Realtime WebSocket 接口 |

### 5.3 图像

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/v1/images/generations` | TokenAuth | 文生图接口 |
| POST | `/v1/images/edits` | TokenAuth | 图像编辑接口 |
| POST | `/v1/edits` | TokenAuth | 图像/编辑兼容入口，内部按 OpenAIImage 处理 |
| POST | `/v1/images/variations` | TokenAuth | 已注册，但当前返回未实现 |

### 5.4 向量与排序

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/v1/embeddings` | TokenAuth | 向量嵌入接口 |
| POST | `/v1/engines/:model/embeddings` | TokenAuth | 旧式/兼容嵌入接口 |
| POST | `/v1/rerank` | TokenAuth | 重排接口 |

### 5.5 音频

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/v1/audio/transcriptions` | TokenAuth | 语音转文字 |
| POST | `/v1/audio/translations` | TokenAuth | 音频翻译 |
| POST | `/v1/audio/speech` | TokenAuth | 文本转语音 |

### 5.6 Gemini 原生/兼容转发

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/v1/models/*path` | TokenAuth | Gemini 转发兼容入口 |
| POST | `/v1beta/models/*path` | TokenAuth | Gemini 原生路径转发 |

### 5.7 未实现但保留的兼容路径

以下接口已注册但当前统一返回未实现：

| 方法 | 路径 |
| --- | --- |
| GET | `/v1/files` |
| POST | `/v1/files` |
| DELETE | `/v1/files/:id` |
| GET | `/v1/files/:id` |
| GET | `/v1/files/:id/content` |
| POST | `/v1/fine-tunes` |
| GET | `/v1/fine-tunes` |
| GET | `/v1/fine-tunes/:id` |
| POST | `/v1/fine-tunes/:id/cancel` |
| GET | `/v1/fine-tunes/:id/events` |
| DELETE | `/v1/models/:model` |

## 6. 视频生成 API

### 6.1 OpenAI 兼容视频接口

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/v1/video/generations` | TokenAuth | 创建视频生成任务 |
| GET | `/v1/video/generations/:task_id` | TokenAuth | 查询视频任务结果 |
| POST | `/v1/videos` | TokenAuth | OpenAI 风格视频创建接口 |
| GET | `/v1/videos/:task_id` | TokenAuth | OpenAI 风格视频查询接口 |
| POST | `/v1/videos/:video_id/remix` | TokenAuth | 视频 remix |
| GET | `/v1/videos/:task_id/content` | TokenOrUserAuth | 代理获取视频内容 |

### 6.2 Kling 接口适配

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/kling/v1/videos/text2video` | TokenAuth | Kling 文生视频，内部转换为统一任务格式 |
| POST | `/kling/v1/videos/image2video` | TokenAuth | Kling 图生视频 |
| GET | `/kling/v1/videos/text2video/:task_id` | TokenAuth | 查询文生视频任务 |
| GET | `/kling/v1/videos/image2video/:task_id` | TokenAuth | 查询图生视频任务 |

### 6.3 即梦接口适配

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/jimeng/` | TokenAuth | 即梦官方风格入口，内部通过中间件转换为统一任务接口 |

## 7. Midjourney API

### 7.1 Midjourney 中继路径

支持两种前缀：

- `/mj/...`
- `/:mode/mj/...`

### 7.2 接口列表

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/mj/image/:id` | 无 | 获取 Midjourney 图片代理内容 |
| POST | `/mj/submit/action` | TokenAuth | 提交动作任务 |
| POST | `/mj/submit/shorten` | TokenAuth | 提交 shorten |
| POST | `/mj/submit/modal` | TokenAuth | 提交 modal |
| POST | `/mj/submit/imagine` | TokenAuth | 提交 imagine |
| POST | `/mj/submit/change` | TokenAuth | 提交 change |
| POST | `/mj/submit/simple-change` | TokenAuth | 提交 simple-change |
| POST | `/mj/submit/describe` | TokenAuth | 提交 describe |
| POST | `/mj/submit/blend` | TokenAuth | 提交 blend |
| POST | `/mj/submit/edits` | TokenAuth | 提交 edits |
| POST | `/mj/submit/video` | TokenAuth | 提交视频任务 |
| GET | `/mj/task/:id/fetch` | TokenAuth | 查询任务 |
| GET | `/mj/task/:id/image-seed` | TokenAuth | 查询图片 seed |
| POST | `/mj/task/list-by-condition` | TokenAuth | 条件查询任务列表 |
| POST | `/mj/insight-face/swap` | TokenAuth | 人脸替换 |
| POST | `/mj/submit/upload-discord-images` | TokenAuth | 上传 Discord 图片 |

`/:mode/mj/...` 下提供与上表同名的全部接口。

## 8. Suno API

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/suno/submit/:action` | TokenAuth | 提交 Suno 任务 |
| POST | `/suno/fetch` | TokenAuth | 按 body 查询任务 |
| GET | `/suno/fetch/:id` | TokenAuth | 按任务 ID 查询 |

## 9. OpenAI Dashboard 兼容账单接口

这些接口启用 `TokenAuth`，返回 OpenAI Dashboard 常见账单结构：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/dashboard/billing/subscription` | 查询配额/订阅信息 |
| GET | `/v1/dashboard/billing/subscription` | 同上 |
| GET | `/dashboard/billing/usage` | 查询用量 |
| GET | `/v1/dashboard/billing/usage` | 同上 |

## 10. 站点公开 API `/api`

以下接口无需管理员权限，适合前台页面、注册登录、展示页和支付回调使用。

### 10.1 系统与展示信息

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/setup` | 无 | 获取初始化状态 |
| POST | `/api/setup` | 无 | 执行初始化 |
| GET | `/api/status` | 无 | 获取站点运行状态、开关、OAuth、主题等信息 |
| GET | `/api/uptime/status` | 无 | 获取 Uptime 状态 |
| GET | `/api/notice` | 无 | 公告 |
| GET | `/api/user-agreement` | 无 | 用户协议 |
| GET | `/api/privacy-policy` | 无 | 隐私政策 |
| GET | `/api/about` | 无 | 关于页面内容 |
| GET | `/api/home_page_content` | 无 | 首页内容 |
| GET | `/api/rankings` | 无 | 排行榜 |

### 10.2 定价与模型展示

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/pricing` | 可匿名，支持 TryUserAuth | 获取定价、供应商、分组倍率、可用分组、支持端点等 |
| GET | `/api/models` | UserAuth | 控制台模型列表 |
| GET | `/api/ratio_config` | 无 | 获取倍率相关配置 |

### 10.3 性能指标展示

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/perf-metrics/summary` | TryUserAuth | 性能指标摘要 |
| GET | `/api/perf-metrics` | TryUserAuth | 性能指标详情 |

### 10.4 验证、密码与 OAuth

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/verification` | 无 | 发送邮箱验证码 |
| GET | `/api/reset_password` | 无 | 发送重置密码邮件 |
| POST | `/api/user/reset` | 无 | 重置密码 |
| GET | `/api/oauth/state` | 无 | 生成 OAuth state/code |
| POST | `/api/oauth/email/bind` | 无 | 邮箱绑定 |
| GET | `/api/oauth/wechat` | 无 | 微信 OAuth |
| POST | `/api/oauth/wechat/bind` | 无 | 微信绑定 |
| GET | `/api/oauth/telegram/login` | 无 | Telegram 登录 |
| GET | `/api/oauth/telegram/bind` | 无 | Telegram 绑定 |
| GET | `/api/oauth/:provider` | 无 | 标准 OAuth 提供商登录入口 |
| POST | `/api/verify` | UserAuth | 通用安全验证 |

### 10.5 支付回调

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/stripe/webhook` | 无 | Stripe 回调 |
| POST | `/api/creem/webhook` | 无 | Creem 回调 |
| POST | `/api/waffo/webhook` | 无 | Waffo 回调 |
| POST | `/api/yima/webhook` | 无 | 易码回调 |
| POST | `/api/subscription/epay/notify` | 无 | 订阅 Epay 回调 |
| GET | `/api/subscription/epay/notify` | 无 | 订阅 Epay 回调 |
| GET | `/api/subscription/epay/return` | 无 | 订阅 Epay 返回 |
| POST | `/api/subscription/epay/return` | 无 | 订阅 Epay 返回 |
| POST | `/api/subscription/yima/notify` | 无 | 订阅易码回调 |
| GET | `/api/subscription/yima/return` | 无 | 订阅易码返回 |

## 11. 用户账户 API `/api/user`

### 11.1 登录注册

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/user/register` | 无 | 注册 |
| POST | `/api/user/login` | 无 | 登录 |
| POST | `/api/user/login/2fa` | 无 | 二步验证登录 |
| POST | `/api/user/passkey/login/begin` | 无 | Passkey 登录开始 |
| POST | `/api/user/passkey/login/finish` | 无 | Passkey 登录完成 |
| GET | `/api/user/logout` | 无 | 登出 |

### 11.2 用户公开分组与支付回调

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| POST | `/api/user/epay/notify` | 无 | Epay 回调 |
| GET | `/api/user/epay/notify` | 无 | Epay 回调 |
| POST | `/api/user/yima/notify` | 无 | 易码充值回调 |
| GET | `/api/user/yima/return` | 无 | 易码充值返回 |
| GET | `/api/user/groups` | 无 | 获取用户可见分组 |

### 11.3 当前用户自助接口

以下接口要求 `UserAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/user/self/groups` | 当前用户可用分组 |
| GET | `/api/user/self` | 当前用户资料 |
| GET | `/api/user/models` | 当前用户可用模型 |
| PUT | `/api/user/self` | 更新个人资料 |
| DELETE | `/api/user/self` | 删除当前账户 |
| GET | `/api/user/token` | 生成 Access Token |
| GET | `/api/user/passkey` | Passkey 状态 |
| POST | `/api/user/passkey/register/begin` | Passkey 注册开始 |
| POST | `/api/user/passkey/register/finish` | Passkey 注册完成 |
| POST | `/api/user/passkey/verify/begin` | Passkey 验证开始 |
| POST | `/api/user/passkey/verify/finish` | Passkey 验证完成 |
| DELETE | `/api/user/passkey` | 删除 Passkey |
| GET | `/api/user/aff` | 获取邀请信息 |
| GET | `/api/user/topup/info` | 获取充值信息 |
| GET | `/api/user/topup/self` | 获取当前用户充值记录 |
| POST | `/api/user/topup` | 充值 |
| POST | `/api/user/pay` | 发起 Epay 支付 |
| POST | `/api/user/amount` | 计算支付金额 |
| POST | `/api/user/stripe/pay` | 发起 Stripe 支付 |
| POST | `/api/user/stripe/amount` | 计算 Stripe 金额 |
| POST | `/api/user/creem/pay` | 发起 Creem 支付 |
| POST | `/api/user/waffo/amount` | 计算 Waffo 金额 |
| POST | `/api/user/waffo/pay` | 发起 Waffo 支付 |
| POST | `/api/user/aff_transfer` | 邀请额度转移 |
| PUT | `/api/user/setting` | 更新用户设置 |
| GET | `/api/user/2fa/status` | 2FA 状态 |
| POST | `/api/user/2fa/setup` | 初始化 2FA |
| POST | `/api/user/2fa/enable` | 启用 2FA |
| POST | `/api/user/2fa/disable` | 关闭 2FA |
| POST | `/api/user/2fa/backup_codes` | 重新生成备用码 |
| GET | `/api/user/checkin` | 获取签到状态 |
| POST | `/api/user/checkin` | 执行签到 |
| GET | `/api/user/oauth/bindings` | 获取已绑定 OAuth |
| DELETE | `/api/user/oauth/bindings/:provider_id` | 解绑自定义 OAuth |

### 11.4 用户管理接口（管理员）

以下接口要求 `AdminAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/user/` | 所有用户列表 |
| GET | `/api/user/topup` | 所有充值记录 |
| POST | `/api/user/topup/complete` | 管理员确认充值 |
| GET | `/api/user/search` | 搜索用户 |
| GET | `/api/user/:id/oauth/bindings` | 查询指定用户 OAuth 绑定 |
| DELETE | `/api/user/:id/oauth/bindings/:provider_id` | 管理员解绑自定义 OAuth |
| DELETE | `/api/user/:id/bindings/:binding_type` | 清理用户绑定 |
| GET | `/api/user/:id` | 获取用户详情 |
| POST | `/api/user/` | 创建用户 |
| POST | `/api/user/manage` | 管理用户状态/额度等 |
| PUT | `/api/user/` | 更新用户 |
| DELETE | `/api/user/:id` | 删除用户 |
| DELETE | `/api/user/:id/reset_passkey` | 重置用户 Passkey |
| GET | `/api/user/2fa/stats` | 2FA 统计 |
| DELETE | `/api/user/:id/2fa` | 管理员关闭用户 2FA |

## 12. 订阅 API `/api/subscription`

### 12.1 用户订阅接口

要求 `UserAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/subscription/plans` | 订阅套餐列表 |
| GET | `/api/subscription/self` | 当前用户订阅信息 |
| PUT | `/api/subscription/self/preference` | 更新订阅偏好 |
| POST | `/api/subscription/epay/pay` | 发起订阅 Epay 支付 |
| POST | `/api/subscription/stripe/pay` | 发起订阅 Stripe 支付 |
| POST | `/api/subscription/creem/pay` | 发起订阅 Creem 支付 |

### 12.2 订阅管理接口

要求 `AdminAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/subscription/admin/plans` | 查询订阅计划 |
| POST | `/api/subscription/admin/plans` | 创建订阅计划 |
| PUT | `/api/subscription/admin/plans/:id` | 更新订阅计划 |
| PATCH | `/api/subscription/admin/plans/:id` | 修改订阅计划状态 |
| POST | `/api/subscription/admin/bind` | 绑定订阅 |
| GET | `/api/subscription/admin/users/:id/subscriptions` | 查询用户订阅 |
| POST | `/api/subscription/admin/users/:id/subscriptions` | 创建用户订阅 |
| POST | `/api/subscription/admin/user_subscriptions/:id/invalidate` | 失效用户订阅 |
| DELETE | `/api/subscription/admin/user_subscriptions/:id` | 删除用户订阅 |

## 13. Token 与用量 API

### 13.1 令牌管理 `/api/token`

要求 `UserAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/token/` | 当前用户全部令牌 |
| GET | `/api/token/search` | 搜索令牌 |
| GET | `/api/token/:id` | 查询令牌详情 |
| POST | `/api/token/:id/key` | 获取令牌明文 key |
| POST | `/api/token/` | 创建令牌 |
| PUT | `/api/token/` | 更新令牌 |
| DELETE | `/api/token/:id` | 删除令牌 |
| POST | `/api/token/batch` | 批量删除令牌 |
| POST | `/api/token/batch/keys` | 批量导出令牌 key |

### 13.2 令牌只读用量接口 `/api/usage`

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/usage/token/` | TokenAuthReadOnly | 通过 token 查询其用量 |

## 14. 渠道、模型、供应商与部署管理 API

### 14.1 渠道管理 `/api/channel`

要求 `AdminAuth`，个别接口还额外要求 `RootAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/channel/` | 全部渠道 |
| GET | `/api/channel/search` | 搜索渠道 |
| GET | `/api/channel/models` | 渠道模型列表 |
| GET | `/api/channel/models_enabled` | 已启用模型列表 |
| GET | `/api/channel/:id` | 渠道详情 |
| POST | `/api/channel/:id/key` | 获取渠道密钥，需 RootAuth |
| GET | `/api/channel/test` | 测试全部渠道 |
| GET | `/api/channel/test/:id` | 测试指定渠道 |
| GET | `/api/channel/update_balance` | 更新全部渠道余额 |
| GET | `/api/channel/update_balance/:id` | 更新指定渠道余额 |
| POST | `/api/channel/` | 新增渠道 |
| PUT | `/api/channel/` | 更新渠道 |
| DELETE | `/api/channel/disabled` | 删除已禁用渠道 |
| POST | `/api/channel/tag/disabled` | 按标签禁用 |
| POST | `/api/channel/tag/enabled` | 按标签启用 |
| PUT | `/api/channel/tag` | 批量修改标签 |
| DELETE | `/api/channel/:id` | 删除渠道 |
| POST | `/api/channel/batch` | 批量删除渠道 |
| POST | `/api/channel/fix` | 修复渠道能力 |
| GET | `/api/channel/fetch_models/:id` | 抓取上游模型 |
| POST | `/api/channel/fetch_models` | 批量抓取模型，需 RootAuth |
| POST | `/api/channel/codex/oauth/start` | 启动 Codex OAuth |
| POST | `/api/channel/codex/oauth/complete` | 完成 Codex OAuth |
| POST | `/api/channel/:id/codex/oauth/start` | 指定渠道启动 Codex OAuth |
| POST | `/api/channel/:id/codex/oauth/complete` | 指定渠道完成 Codex OAuth |
| POST | `/api/channel/:id/codex/refresh` | 刷新 Codex 凭证 |
| GET | `/api/channel/:id/codex/usage` | 查询 Codex 用量 |
| POST | `/api/channel/ollama/pull` | 拉取 Ollama 模型 |
| POST | `/api/channel/ollama/pull/stream` | 流式拉取 Ollama 模型 |
| DELETE | `/api/channel/ollama/delete` | 删除 Ollama 模型 |
| GET | `/api/channel/ollama/version/:id` | 查询 Ollama 版本 |
| POST | `/api/channel/batch/tag` | 批量设置标签 |
| GET | `/api/channel/tag/models` | 查询标签模型 |
| POST | `/api/channel/copy/:id` | 复制渠道 |
| POST | `/api/channel/multi_key/manage` | 管理多 key 渠道 |
| POST | `/api/channel/upstream_updates/apply` | 应用上游模型变更 |
| POST | `/api/channel/upstream_updates/apply_all` | 应用全部上游变更 |
| POST | `/api/channel/upstream_updates/detect` | 检测上游模型变更 |
| POST | `/api/channel/upstream_updates/detect_all` | 检测全部上游模型变更 |

### 14.2 供应商元数据 `/api/vendors`

要求 `AdminAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/vendors/` | 供应商列表 |
| GET | `/api/vendors/search` | 搜索供应商 |
| GET | `/api/vendors/:id` | 供应商详情 |
| POST | `/api/vendors/` | 创建供应商 |
| PUT | `/api/vendors/` | 更新供应商 |
| DELETE | `/api/vendors/:id` | 删除供应商 |

### 14.3 模型元数据 `/api/models`

要求 `AdminAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/models/sync_upstream/preview` | 预览同步上游模型 |
| POST | `/api/models/sync_upstream` | 同步上游模型 |
| GET | `/api/models/missing` | 查询缺失模型 |
| GET | `/api/models/` | 查询模型元数据 |
| GET | `/api/models/search` | 搜索模型元数据 |
| GET | `/api/models/:id` | 模型详情 |
| POST | `/api/models/` | 创建模型元数据 |
| PUT | `/api/models/` | 更新模型元数据 |
| DELETE | `/api/models/:id` | 删除模型元数据 |

### 14.4 部署管理 `/api/deployments`

要求 `AdminAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/deployments/settings` | 获取部署配置 |
| POST | `/api/deployments/settings/test-connection` | 测试 io.net 连接 |
| GET | `/api/deployments/` | 部署列表 |
| GET | `/api/deployments/search` | 搜索部署 |
| POST | `/api/deployments/test-connection` | 测试连接 |
| GET | `/api/deployments/hardware-types` | 硬件类型列表 |
| GET | `/api/deployments/locations` | 区域列表 |
| GET | `/api/deployments/available-replicas` | 可用副本数 |
| POST | `/api/deployments/price-estimation` | 价格估算 |
| GET | `/api/deployments/check-name` | 检查集群名可用性 |
| POST | `/api/deployments/` | 创建部署 |
| GET | `/api/deployments/:id` | 部署详情 |
| GET | `/api/deployments/:id/logs` | 部署日志 |
| GET | `/api/deployments/:id/containers` | 容器列表 |
| GET | `/api/deployments/:id/containers/:container_id` | 容器详情 |
| PUT | `/api/deployments/:id` | 更新部署 |
| PUT | `/api/deployments/:id/name` | 修改部署名 |
| POST | `/api/deployments/:id/extend` | 续期/扩容 |
| DELETE | `/api/deployments/:id` | 删除部署 |

## 15. 配置、性能与倍率同步 API

### 15.1 系统配置 `/api/option`

要求 `RootAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/option/` | 获取全部配置 |
| PUT | `/api/option/` | 更新配置 |
| GET | `/api/option/channel_affinity_cache` | 渠道亲和缓存统计 |
| DELETE | `/api/option/channel_affinity_cache` | 清空渠道亲和缓存 |
| POST | `/api/option/rest_model_ratio` | 重置模型倍率 |
| POST | `/api/option/migrate_console_setting` | 控制台配置迁移 |

### 15.2 自定义 OAuth 提供商 `/api/custom-oauth-provider`

要求 `RootAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/custom-oauth-provider/discovery` | 发现 OIDC / OAuth 配置 |
| GET | `/api/custom-oauth-provider/` | 列表 |
| GET | `/api/custom-oauth-provider/:id` | 详情 |
| POST | `/api/custom-oauth-provider/` | 创建 |
| PUT | `/api/custom-oauth-provider/:id` | 更新 |
| DELETE | `/api/custom-oauth-provider/:id` | 删除 |

### 15.3 性能与维护 `/api/performance`

要求 `RootAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/performance/stats` | 性能统计 |
| DELETE | `/api/performance/disk_cache` | 清空磁盘缓存 |
| POST | `/api/performance/reset_stats` | 重置统计 |
| POST | `/api/performance/gc` | 触发 GC |
| GET | `/api/performance/logs` | 获取日志文件列表 |
| DELETE | `/api/performance/logs` | 清理日志文件 |

### 15.4 上游倍率同步 `/api/ratio_sync`

要求 `RootAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/ratio_sync/channels` | 获取可同步渠道 |
| POST | `/api/ratio_sync/fetch` | 拉取上游倍率数据 |

## 16. 日志、兑换码、任务与统计 API

### 16.1 日志 `/api/log`

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/log/` | AdminAuth | 所有日志 |
| DELETE | `/api/log/` | AdminAuth | 删除历史日志 |
| GET | `/api/log/stat` | AdminAuth | 日志统计 |
| GET | `/api/log/self/stat` | UserAuth | 当前用户日志统计 |
| GET | `/api/log/channel_affinity_usage_cache` | AdminAuth | 渠道亲和使用缓存统计 |
| GET | `/api/log/search` | AdminAuth | 搜索全站日志 |
| GET | `/api/log/self` | UserAuth | 当前用户日志 |
| GET | `/api/log/self/search` | UserAuth | 搜索当前用户日志 |
| GET | `/api/log/token` | TokenAuthReadOnly | 根据 token 查询日志 |

### 16.2 数据统计 `/api/data`

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/data/` | AdminAuth | 全站额度日期统计 |
| GET | `/api/data/users` | AdminAuth | 用户维度额度统计 |
| GET | `/api/data/self` | UserAuth | 当前用户额度统计 |

### 16.3 兑换码 `/api/redemption`

要求 `AdminAuth`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/redemption/` | 兑换码列表 |
| GET | `/api/redemption/search` | 搜索兑换码 |
| GET | `/api/redemption/:id` | 兑换码详情 |
| POST | `/api/redemption/` | 创建兑换码 |
| PUT | `/api/redemption/` | 更新兑换码 |
| DELETE | `/api/redemption/invalid` | 删除无效兑换码 |
| DELETE | `/api/redemption/:id` | 删除兑换码 |

### 16.4 分组与预填分组

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/group/` | AdminAuth | 分组列表 |
| GET | `/api/prefill_group/` | AdminAuth | 预填分组列表 |
| POST | `/api/prefill_group/` | AdminAuth | 创建预填分组 |
| PUT | `/api/prefill_group/` | AdminAuth | 更新预填分组 |
| DELETE | `/api/prefill_group/:id` | AdminAuth | 删除预填分组 |

### 16.5 Midjourney 与任务查询

| 方法 | 路径 | 认证 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/mj/self` | UserAuth | 当前用户 Midjourney 任务 |
| GET | `/api/mj/` | AdminAuth | 所有 Midjourney 任务 |
| GET | `/api/task/self` | UserAuth | 当前用户任务 |
| GET | `/api/task/` | AdminAuth | 所有任务 |

## 17. 示例

### 17.1 Chat Completions

```bash
curl --request POST 'https://your-domain/v1/chat/completions' \
  --header 'Authorization: Bearer sk-your-token' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "你好，介绍一下这个项目。"}
    ],
    "stream": false
  }'
```

### 17.2 Claude Messages

```bash
curl --request POST 'https://your-domain/v1/messages' \
  --header 'x-api-key: sk-your-token' \
  --header 'anthropic-version: 2023-06-01' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "model": "claude-3-5-sonnet",
    "max_tokens": 512,
    "messages": [
      {"role": "user", "content": "请总结一下本文档。"}
    ]
  }'
```

### 17.3 Responses API

```bash
curl --request POST 'https://your-domain/v1/responses' \
  --header 'Authorization: Bearer sk-your-token' \
  --header 'Content-Type: application/json' \
  --data-raw '{
    "model": "gpt-4.1-mini",
    "input": "生成一段 Go 代码，读取 JSON 配置文件。"
  }'
```

### 17.4 获取站点状态

```bash
curl 'https://your-domain/api/status'
```

### 17.5 获取定价

```bash
curl 'https://your-domain/api/pricing'
```

## 18. 实现说明与注意事项

1. 路由完整性
   - 本文档按当前仓库路由实现整理，不包含前端页面路由。

2. 兼容层行为
   - `/v1/models` 会根据请求头自动判断返回 OpenAI、Claude 或 Gemini 风格。
   - `/v1/messages` 使用 Claude Messages 格式。
   - `/v1beta/models/*path` 与 `/v1/models/*path` 会进入 Gemini 相关转发逻辑。

3. 成功判定
   - `/v1` 系列请按上游兼容规范判断。
   - `/api` 系列建议同时检查 HTTP 状态、`success` 字段和 `message` 字段。

4. 权限分层
   - `UserAuth`：普通登录用户
   - `AdminAuth`：管理员
   - `RootAuth`：最高权限
   - `TokenAuth`：模型调用 token

5. 限流与验证码
   - 部分注册、登录、重置密码、签到、支付接口带有限流与 Turnstile 校验。

6. 文档维护建议
   - 新增路由时同步更新本文档。
   - 若未来需要机器可读文档，建议继续生成 `openapi.yaml`，按本文档分类补齐 schema。