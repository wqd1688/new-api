# 如何在 IntelliJ IDEA 中使用 Centoken API

## 一句话说明

如果你在 IntelliJ IDEA 中使用的是支持 OpenAI 兼容接口的 AI 插件，或者你想直接在 IDEA 里调试 HTTP 请求、编写项目代码调用 AI 接口，那么都可以接入 Centoken。

Centoken 线上地址：<http://centoken.cn>

## 适合哪些人看这篇文档

这篇文档适合以下用户：

- 想在 IDEA 里配置 AI 编程助手的开发者
- 想在 IDEA 中直接测试 Centoken API 的用户
- 想在 Java、Kotlin、Spring Boot 等项目里接入 Centoken 的开发者

## 在 IDEA 中使用 Centoken 的几种方式

在 IntelliJ IDEA 中，通常有三种常见用法：

### 1. 配置 AI 插件使用

如果你安装的 IDEA 插件支持自定义 OpenAI Base URL、API Key 和模型名称，那么通常可以直接接入 Centoken。

### 2. 使用 IDEA 自带的 HTTP Client 调试接口

如果你只是想验证接口、测试模型是否可用，IDEA 自带的 HTTP 请求调试功能就足够了。

### 3. 在项目代码中调用 Centoken API

如果你是在 IDEA 中开发 Java、Kotlin 或其他后端项目，可以直接在代码里调用 Centoken API，把它接入你的业务系统。

## 使用前准备

在开始之前，请先准备好以下信息：

- Centoken 地址：`http://centoken.cn`
- 你的 API Key：在控制台创建或查看
- 你要使用的模型名称：例如 `deepseek-v4-flash`

相关入口：

- 首页：<http://centoken.cn>
- 登录页：<http://centoken.cn/login>
- 控制台：<http://centoken.cn/console>
- 模型广场：<http://centoken.cn/pricing>

## 方式一：在 IDEA 中通过 AI 插件接入 Centoken

这是最适合日常编码使用的一种方式。

### 适合哪些插件

只要插件支持以下配置项，通常就可以尝试接入：

- OpenAI API Key
- OpenAI Base URL
- API Host
- Custom Provider
- Model Name

如果插件只能填写 OpenAI Key，但不能修改接口地址，那通常无法直接使用 Centoken。

### 通用配置步骤

不同插件界面不完全一样，但配置思路基本一致：

1. 在 IDEA 中打开插件设置页
2. 找到 AI Provider、Model Provider、OpenAI 或自定义模型相关配置
3. 将接口地址改为 Centoken 地址
4. 填写 API Key
5. 选择或填写模型名称
6. 保存后进行一次测试对话或代码生成

### 通常这样填写

- Base URL：`http://centoken.cn`
- API Key：你的 Centoken Key
- Model：`deepseek-v4-flash`

如果插件要求填写完整聊天接口地址，可以尝试：

```text
http://centoken.cn/v1/chat/completions
```

如果插件支持 Responses 风格接口，也可能会用到：

```text
http://centoken.cn/v1/responses
```

### 在 IDEA 中可以实现哪些能力

成功接入后，通常可以在 IDEA 里使用这些 AI 能力：

- 代码解释
- 函数生成
- 单元测试生成
- 注释补全
- Bug 分析
- SQL 或正则生成
- 重构建议
- 文档草拟

### 适合优先尝试的插件类型

- 支持 OpenAI 兼容接口的聊天插件
- 支持自定义 Provider 的代码助手插件
- 支持填写 Base URL 的开发类 AI 插件

### 常见问题

#### 插件里只有 API Key，没有 Base URL

这种情况通常不能直接接入 Centoken，因为插件把服务地址固定死了。

#### 插件保存后能连上，但回答失败

通常优先检查以下几项：

- 模型名称是否填写正确
- 插件是否要求完整接口路径
- 插件是否默认使用了某个 Centoken 当前未启用的高级能力
- API Key 是否有效

#### 插件支持流式输出，但一直报错

有些插件会强依赖流式响应格式或工具调用格式。如果普通对话能通、插件功能异常，通常是插件和目标模型能力没有完全对齐，建议先换一个更通用的聊天模型测试。

## 方式二：使用 IDEA 自带 HTTP Client 调试 Centoken API

如果你不想先折腾插件，可以先在 IDEA 里直接发送 HTTP 请求测试 API。

### 第一步：新建一个 HTTP 请求文件

在项目中创建一个文件，例如：

```text
centoken.http
```

然后写入下面的测试请求：

```http
POST http://centoken.cn/v1/chat/completions
Content-Type: application/json
Authorization: Bearer 你的密钥

{
  "model": "deepseek-v4-flash",
  "messages": [
    {
      "role": "user",
      "content": "请解释一下这段代码的作用，并给出优化建议。"
    }
  ]
}
```

### 第二步：点击运行请求

IDEA 会直接展示响应结果。你可以通过这种方式快速验证：

- API Key 是否可用
- 模型是否可用
- 返回格式是否符合预期
- 当前场景适合哪个模型

### 这种方式适合什么场景

- 调试接口
- 联调插件前先测通 API
- 验证不同模型返回结果
- 给后端开发写接口测试样例

## 方式三：在 IDEA 中开发项目并接入 Centoken API

如果你正在 IDEA 中开发 Java 或 Kotlin 项目，可以直接在业务代码中调用 Centoken API。

这也是最适合正式项目接入的方式，因为你可以自己控制：

- 超时
- 重试
- 日志
- 权限
- 缓存
- 错误处理

### Java 示例

下面是一个使用 Java 原生 HttpClient 的简单示例：

```java
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class CentokenExample {
    public static void main(String[] args) throws Exception {
        String requestBody = """
                {
                  \"model\": \"deepseek-v4-flash\",
                  \"messages\": [
                    {\"role\": \"user\", \"content\": \"请帮我生成一个 Spring Boot Controller 示例\"}
                  ]
                }
                """;

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create("http://centoken.cn/v1/chat/completions"))
                .header("Content-Type", "application/json")
                .header("Authorization", "Bearer 你的密钥")
                .POST(HttpRequest.BodyPublishers.ofString(requestBody))
                .build();

        HttpClient client = HttpClient.newHttpClient();
        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

        System.out.println(response.body());
    }
}
```

### 适合接入到哪些项目里

- Spring Boot 项目
- 企业内部后台系统
- SaaS 平台后端
- 智能客服系统
- 内容生成服务
- 文档问答系统

### 推荐做法

如果是正式业务，不建议在前端页面里直接暴露 API Key。更推荐由服务端统一调用 Centoken，再把结果返回给前端。

## IDEA 中最常见的使用场景

把 Centoken 接入 IntelliJ IDEA 后，最常见的用法包括：

### 代码编写辅助

- 根据注释生成代码
- 解释复杂方法
- 生成 DTO、Service、Controller 样板代码
- 生成测试用例

### 代码阅读和重构

- 分析旧代码逻辑
- 给出重构建议
- 找潜在 Bug
- 补充注释和文档

### 后端开发提效

- 生成 SQL
- 生成接口文档
- 设计参数结构
- 输出异常处理方案

### 文档和沟通辅助

- 编写 README
- 生成技术方案草稿
- 总结会议内容
- 整理接口说明

## 如何判断一个 IDEA 插件能不能接 Centoken

最简单的判断方式是看插件设置中有没有下面三类信息：

- 可以自定义服务地址
- 可以填写 API Key
- 可以指定模型名称

如果这三项都支持，通常就可以尝试接入。

如果只支持官方 OpenAI 地址，不能修改 Base URL，那通常不能直接使用 Centoken。

## 接入建议

- 第一次接入时，先用 HTTP Client 测通接口，再配置插件
- 优先选择通用聊天模型做联调，不要一开始就上复杂能力
- 插件报错时，先排查模型名、接口地址和 Key 是否正确
- 正式项目建议由后端统一封装调用逻辑
- 如果后续启用了 HTTPS 域名，生产环境优先使用 HTTPS 地址

## 总结

在 IntelliJ IDEA 中使用 Centoken，最实用的路径有三种：通过支持 OpenAI 兼容接口的插件接入、通过 IDEA 自带 HTTP Client 调试接口，以及在项目代码中直接调用 API。

如果你的目标是提升日常编码效率，优先走插件接入；如果你的目标是验证接口，优先用 HTTP Client；如果你的目标是正式业务接入，优先走服务端代码调用。