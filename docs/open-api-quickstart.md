# Open API 快速接入指南

这是一份给接入方的一页式说明。

如果你的目标只是尽快把系统接进客户端、插件、脚本、后端服务或 AI 应用，通常按本文操作即可。

更完整说明见：

- `docs/open-api-client.md`
- `docs/open-api.md`

## 1. 你只需要准备 3 样东西

1. 服务地址
2. API Key
3. 模型名称

假设你的服务地址是：

```text
https://your-domain
```

那么最常用的 Base URL 是：

```text
https://your-domain/v1
```

## 2. 最推荐的接法

绝大多数应用优先使用 OpenAI 兼容方式：

```http
Authorization: Bearer sk-your-token
```

最常用接口：

- `GET /v1/models`
- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/embeddings`
- `POST /v1/images/generations`
- `POST /v1/audio/transcriptions`
- `POST /v1/audio/speech`

## 3. 5 分钟接通流程

### 第一步：探测服务是否正常

```bash
curl 'https://your-domain/v1/models' \
  -H 'Authorization: Bearer sk-your-token'
```

如果能正常返回模型列表，说明：

- 服务地址没问题
- Key 基本可用
- 当前 token 有权限访问模型接口

### 第二步：发起一次聊天请求

```bash
curl 'https://your-domain/v1/chat/completions' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "你好，请介绍一下你支持哪些能力。"}
    ],
    "stream": false
  }'
```

如果这一步成功，大多数 OpenAI 兼容客户端就已经可以接入了。

### 第三步：按需求扩展其他接口

如果你不是只做聊天，可以继续接：

- 向量化：`POST /v1/embeddings`
- 图像生成：`POST /v1/images/generations`
- 语音转文字：`POST /v1/audio/transcriptions`
- 文本转语音：`POST /v1/audio/speech`
- Responses：`POST /v1/responses`

## 4. 最小可用示例

### 4.1 Chat Completions

```bash
curl 'https://your-domain/v1/chat/completions' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "写一个 Go 的 hello world"}
    ]
  }'
```

### 4.2 Responses

```bash
curl 'https://your-domain/v1/responses' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gpt-4.1-mini",
    "input": "总结一下 REST API 设计原则"
  }'
```

### 4.3 Embeddings

```bash
curl 'https://your-domain/v1/embeddings' \
  -H 'Authorization: Bearer sk-your-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "text-embedding-3-small",
    "input": "hello world"
  }'
```

## 5. 如果你不是 OpenAI 风格客户端

### Claude 风格

```http
POST /v1/messages
x-api-key: sk-your-token
anthropic-version: 2023-06-01
```

### Gemini 风格

```http
POST /v1beta/models/{model}:{action}
x-goog-api-key: sk-your-token
```

或：

```text
POST /v1beta/models/{model}:{action}?key=sk-your-token
```

## 6. 常见场景对应接口

| 场景 | 推荐接口 |
| --- | --- |
| 聊天、问答、写作 | `/v1/chat/completions` |
| 新版 OpenAI SDK | `/v1/responses` |
| 知识库、RAG、向量检索 | `/v1/embeddings` |
| 文生图 | `/v1/images/generations` |
| 录音转文字 | `/v1/audio/transcriptions` |
| 文本转语音 | `/v1/audio/speech` |
| Claude SDK | `/v1/messages` |
| Gemini SDK | `/v1beta/models/{model}:{action}` |

## 7. 任务型接口怎么接

如果你接的是视频、Midjourney、Suno 这类异步任务接口，流程通常不是一次完成，而是：

1. 提交任务
2. 得到 `task_id`
3. 轮询查询结果

常见路径：

- 视频：`POST /v1/video/generations` + `GET /v1/video/generations/:task_id`
- Midjourney：`POST /mj/submit/imagine` + `GET /mj/task/:id/fetch`
- Suno：`POST /suno/submit/:action` + `GET /suno/fetch/:id`

## 8. 遇到错误先看这 5 项

1. Base URL 是否写成了 `/v1`
2. Key 是否有效
3. 模型名是否真实存在且当前可用
4. 当前 token 是否允许访问这个模型或分组
5. 请求格式是否和目标接口兼容

常见错误码：

- `model_not_found`
- `invalid_request`
- `insufficient_user_quota`
- `access_denied`
- `model_price_error`
- `channel:no_available_key`

## 9. 哪些客户端最容易直接接入

下面这些通常可以直接对接：

- 支持自定义 OpenAI Base URL 的聊天客户端
- 支持自定义 Provider 的 IDE 插件
- 使用 OpenAI SDK 的后端服务
- 知识库 / RAG 系统
- 自动化脚本和工作流平台

## 10. 一句话建议

第一次对接时，不要一上来就接复杂能力。

先用：

1. `GET /v1/models`
2. `POST /v1/chat/completions`

这两步跑通之后，再继续接图片、音频、向量、视频或任务接口，效率最高。