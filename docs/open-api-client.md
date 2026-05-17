# 面向调用方的 Open API 文档

这份文档是给实际接入方看的精简版说明。

如果你只是想把这个系统接到自己的客户端、插件、脚本、后端服务或 AI 应用里，通常只需要看这一份，不需要先读完整后台接口文档。

完整后台与管理接口清单见：`docs/open-api.md`

## 1. 一句话说明

这个项目对外主要提供的是一组兼容 OpenAI 的 AI 调用接口，同时兼容部分 Claude、Gemini、Midjourney、Suno 和视频生成调用方式。

大多数接入场景下，你只需要准备三样东西：

- Base URL
- API Key
- 模型名称

## 2. 基础信息

假设你的服务地址是：

```text
https://your-domain
```

那么常见基地址分别是：

- OpenAI 兼容：`https://your-domain/v1`
- Gemini 兼容：`https://your-domain/v1beta`
- Midjourney：`https://your-domain/mj`
- Suno：`https://your-domain/suno`

## 3. 认证方式

### 3.1 OpenAI 兼容方式

最常见的认证方式：

```http
Authorization: Bearer sk-your-token
```

### 3.2 Claude 兼容方式

如果你调用的是 Claude Messages 风格接口：

```http
x-api-key: sk-your-token
anthropic-version: 2023-06-01
```

### 3.3 Gemini 兼容方式

Gemini 风格支持以下两种写法之一：

```http
x-goog-api-key: sk-your-token
```

或：

```text
?key=sk-your-token
```

## 4. 最常用的接口

如果你是第一次接入，优先关注下面这些。

### 4.1 获取模型列表

```http
GET /v1/models
```

用途：

- 查看当前令牌可调用的模型
- 给客户端做模型下拉框
- 接入前探测服务是否正常

示例：

```bash
curl 'https://your-domain/v1/models' \
  -H 'Authorization: Bearer sk-your-token'
```

### 4.2 Chat Completions

```http
POST /v1/chat/completions
```

这是最通用、兼容性最好的文本对话接口。大部分聊天客户端、插件、IDE、脚本都优先支持这个接口。

示例：

```bash
curl 'https://your-domain/v1/chat/completions' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {
        "role": "user",
        "content": "请用三句话介绍这个系统的作用。"
      }
    ],
    "stream": false
  }'
```

典型请求字段：

- `model`: 模型名称
- `messages`: 对话消息数组
- `stream`: 是否流式输出
- `temperature`: 采样温度
- `max_tokens`: 最大输出 token

### 4.3 Responses API

```http
POST /v1/responses
```

如果你的客户端或 SDK 已经切到 OpenAI Responses 风格，可以用这个接口。

示例：

```bash
curl 'https://your-domain/v1/responses' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-4.1-mini",
    "input": "帮我写一个 Go 的 HTTP 请求示例"
  }'
```

### 4.4 Embeddings

```http
POST /v1/embeddings
```

适合：

- 知识库
- RAG 检索
- 语义搜索
- 向量化入库

示例：

```bash
curl 'https://your-domain/v1/embeddings' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "text-embedding-3-small",
    "input": "这是一段需要做向量化的文本"
  }'
```

### 4.5 图像生成

```http
POST /v1/images/generations
```

适合：

- 文生图
- 图片素材生成
- 海报草图生成

示例：

```bash
curl 'https://your-domain/v1/images/generations' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-image-1",
    "prompt": "一只站在雨夜霓虹街头的机械猫",
    "size": "1024x1024"
  }'
```

### 4.6 语音转文字

```http
POST /v1/audio/transcriptions
```

适合：

- 录音转写
- 会议纪要
- 语音内容提取

通常为 `multipart/form-data` 上传。

### 4.7 文本转语音

```http
POST /v1/audio/speech
```

适合：

- 配音
- 语音播报
- 语音助手

### 4.8 Claude Messages

```http
POST /v1/messages
```

如果你用的是 Claude 官方风格 SDK 或已经按 Claude 请求格式组织数据，可以走这个接口。

示例：

```bash
curl 'https://your-domain/v1/messages' \
  -H 'x-api-key: sk-your-token' \
  -H 'anthropic-version: 2023-06-01' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "claude-3-5-sonnet",
    "max_tokens": 512,
    "messages": [
      {
        "role": "user",
        "content": "请总结一下 API 接入步骤"
      }
    ]
  }'
```

### 4.9 Gemini 风格调用

```http
POST /v1beta/models/{model}:{action}
```

适合已经按 Gemini SDK 或 Gemini API 协议组织请求的场景。

例如：

```text
POST /v1beta/models/gemini-2.5-flash:generateContent
```

### 4.10 视频生成

最常见的视频接口有：

- `POST /v1/video/generations`
- `GET /v1/video/generations/:task_id`
- `POST /v1/videos`
- `GET /v1/videos/:task_id`

适合：

- 文生视频
- 图生视频
- 视频 remix

这类接口通常是异步任务模式：

1. 先创建任务
2. 拿到 `task_id`
3. 再轮询查询结果

## 5. 任务型接口

除了标准 OpenAI 兼容接口，这个项目还提供了一些任务型 API。

### 5.1 Midjourney

常见接口：

- `POST /mj/submit/imagine`
- `POST /mj/submit/blend`
- `POST /mj/submit/describe`
- `GET /mj/task/:id/fetch`

适合已经接了 Midjourney 工作流的应用。

### 5.2 Suno

常见接口：

- `POST /suno/submit/:action`
- `POST /suno/fetch`
- `GET /suno/fetch/:id`

适合做音乐生成或音频任务轮询。

### 5.3 Kling

常见接口：

- `POST /kling/v1/videos/text2video`
- `POST /kling/v1/videos/image2video`
- `GET /kling/v1/videos/text2video/:task_id`
- `GET /kling/v1/videos/image2video/:task_id`

## 6. 推荐接入顺序

如果你是第一次对接，建议按下面顺序来：

1. 先调用 `GET /v1/models`
2. 再调用 `POST /v1/chat/completions`
3. 确认模型名和 Key 可用后，再接 `responses`、`embeddings`、`images`、`audio` 等其他接口
4. 如果你做的是异步任务型场景，再接视频、Midjourney 或 Suno

## 7. 常见错误

错误通常是 OpenAI 风格：

```json
{
  "error": {
    "message": "insufficient quota",
    "type": "new_api_error",
    "param": "",
    "code": "insufficient_user_quota"
  }
}
```

常见错误码：

- `invalid_request`: 请求格式不合法
- `model_not_found`: 模型不存在或当前不可用
- `model_price_error`: 模型计费未配置
- `insufficient_user_quota`: 额度不足
- `pre_consume_token_quota_failed`: 预扣额度失败
- `access_denied`: 无权访问该接口、分组或资源
- `channel:no_available_key`: 上游渠道当前没有可用 key
- `channel:invalid_key`: 上游渠道 key 无效

## 8. 调用注意事项

### 8.1 不同接口的响应风格不同

对接方主要使用的是 `/v1`、`/v1beta`、`/mj`、`/suno` 等 AI 接口，这些接口通常直接返回兼容上游的响应格式。

如果你调用的是 `/api/...`，那是站点业务接口，返回格式和 AI 接口不是同一套，不建议普通 AI 客户端直接依赖它们。

### 8.2 模型名必须以服务端实际配置为准

虽然接口兼容 OpenAI/Claude/Gemini，但最终能不能调用成功，仍取决于：

- 后台是否启用了该模型
- 当前 token 是否允许访问该模型
- 当前分组是否允许访问该模型

因此第一步最好总是先请求 `/v1/models`。

### 8.3 流式输出是否可用取决于模型和渠道

即使接口本身支持 `stream: true`，具体是否稳定，还取决于上游渠道和模型实现。

### 8.4 任务型接口多为异步

视频、Suno、Midjourney 这类接口通常不是一次请求直接出结果，而是：

- 创建任务
- 返回任务 ID
- 轮询获取结果

## 9. 最小可用示例

如果你只想做最小接入，下面这一组通常已经够了：

### 9.1 探测服务

```bash
curl 'https://your-domain/v1/models' \
  -H 'Authorization: Bearer sk-your-token'
```

### 9.2 发起聊天

```bash
curl 'https://your-domain/v1/chat/completions' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

### 9.3 做向量化

```bash
curl 'https://your-domain/v1/embeddings' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "text-embedding-3-small",
    "input": "hello world"
  }'
```

## 10. 适合哪些客户端直接接入

通常下面这几类可以直接接：

- 支持自定义 OpenAI Base URL 的聊天客户端
- 支持自定义 OpenAI Provider 的 IDE 插件
- 支持 OpenAI API 的后端 SDK
- 支持 Responses 或 Chat Completions 的脚本与自动化程序
- 支持 Embeddings 的知识库/RAG 系统

## 11. 不建议普通接入方优先使用的接口

以下接口更偏后台管理，不适合作为普通模型调用入口：

- `/api/channel/...`
- `/api/option/...`
- `/api/performance/...`
- `/api/models/...`
- `/api/vendors/...`
- `/api/deployments/...`

这些接口主要给控制台后台和管理员使用。

## 12. 文档索引

如果你需要更完整的信息，可以继续看：

- 完整接口总表：`docs/open-api.md`
- IDEA 接入说明：`docs/centoken-idea-guide.md`
- 使用场景说明：`docs/centoken-api-use-cases.md`