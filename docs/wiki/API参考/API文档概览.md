---
title: API文档概览
tags: [API参考, REST, 认证, 接口]
aliases: [API概览, API文档, API参考]
---

# API 文档概览

WeKnora 提供了一系列 RESTful API，用于创建和管理知识库、检索知识，以及进行基于知识的问答。本文档详细描述了这些 API 的使用方式。

## 最权威参考：Swagger UI

WeKnora 同时提供基于 OpenAPI 的 Swagger 文档。**启动服务后访问 `http://localhost:8080/swagger/index.html`**，可看到所有端点的完整参数、请求/响应 schema，并可直接在浏览器内试调--它随代码自动更新，是最准确的接口参考。

本目录下的 markdown 文档提供更易读的示例与场景说明，与 swagger 同步维护；当二者出现差异时，以 swagger 为准。

> Swagger UI 仅在非 release 模式（`GIN_MODE != release`）下挂载；生产部署默认关闭。

## 基础信息

- **基础 URL**: `/api/v1`
- **响应格式**: JSON
- **认证方式**: API Key

> API 认证使用 WeKnora 本地 JWT，OIDC 认证流程参见 [OIDC认证调用流程](../安全认证/OIDC认证调用流程.md)

## 认证机制

所有 API 请求需要在 HTTP 请求头中包含 `X-API-Key` 进行身份认证：

```
X-API-Key: your_api_key
```

为便于问题追踪和调试，建议每个请求的 HTTP 请求头中添加 `X-Request-ID`：

```
X-Request-ID: unique_request_id
```

### 获取 API Key

在 web 页面完成账户注册后，请前往账户信息页面获取您的 API Key。

请妥善保管您的 API Key，避免泄露。API Key 代表您的账户身份，拥有完整的 API 访问权限。

## 错误处理

所有 API 使用标准的 HTTP 状态码表示请求状态，并返回统一的错误响应格式：

```json
{
  "success": false,
  "error": {
    "code": "错误代码",
    "message": "错误信息",
    "details": "错误详情"
  }
}
```

## API 分类

WeKnora API 按功能分为以下几类：

| 分类 | 描述 | 详细文档 |
|------|------|----------|
| 认证管理 | 用户注册、登录、令牌管理；OIDC 流程 | [auth.md](auth.md) · [OIDC认证调用流程.md](../安全认证/OIDC认证调用流程.md) |
| 空间管理 | 创建和管理空间账户 | [tenant.md](tenant.md) |
| 知识库管理 | 创建、查询和管理知识库 | [knowledge-base.md](knowledge-base.md) |
| 知识管理 | 上传、检索和管理知识内容 | [knowledge.md](knowledge.md) |
| 模型管理 | 配置和管理各种AI模型 | [model.md](model.md) |
| 分块管理 | 管理知识的分块内容 | [chunk.md](chunk.md) |
| 标签管理 | 管理知识库的标签分类 | [tag.md](tag.md) |
| FAQ管理 | 管理FAQ问答对 | [faq.md](faq.md) |
| 智能体管理 | 创建和管理自定义智能体 | [agent.md](agent.md) |
| 会话管理 | 创建和管理对话会话 | [session.md](session.md) |
| 知识搜索 | 在知识库中搜索内容 | [knowledge-search.md](knowledge-search.md) |
| 聊天功能 | 基于知识库和 Agent 进行问答 | [chat.md](chat.md) |
| 消息管理 | 获取和管理对话消息 | [message.md](message.md) |
| 评估功能 | 评估模型性能 | [evaluation.md](evaluation.md) |
| 初始化管理 | 知识库模型配置与 Ollama 管理 | [initialization.md](initialization.md) |
| 系统管理 | 系统信息、解析引擎、存储引擎 | [system.md](system.md) |
| MCP 服务 | MCP 工具服务管理 | [mcp-service.md](mcp-service.md) |
| 组织管理 | 组织、成员、知识库/智能体共享 | [organization.md](organization.md) |
| Skills | 预装智能体技能 | [skill.md](skill.md) |
| 网络搜索 | 网络搜索服务商 | [web-search.md](web-search.md) |
| 向量存储 | 向量数据库连接管理 | [vector-store.md](vector-store.md) |
| 存储后端 | 对象/文件存储实例（多实例）管理 | [storage-backend.md](storage-backend.md) |

## 相关主题

- [OIDC认证调用流程](../安全认证/OIDC认证调用流程.md) - API 认证的 OIDC 流程
- [内置模型管理](../核心功能/内置模型管理.md) - 模型管理 API 的配置参考
- [MCP功能使用说明](../核心功能/MCP功能使用说明.md) - MCP 服务管理 API 的使用
- [共享空间说明](../安全认证/共享空间说明.md) - 组织管理 API 的业务逻辑
- [IM集成开发](../集成扩展/IM集成开发.md) - IM 渠道管理 API
- [数据源导入开发](../集成扩展/数据源导入开发.md) - 数据源管理 API

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
- [OIDC认证调用流程](../安全认证/OIDC认证调用流程.md) - API 认证机制与 OIDC 相关
- [内置模型管理](../核心功能/内置模型管理.md) - 模型管理 API 的底层配置
- [MCP功能使用说明](../核心功能/MCP功能使用说明.md) - MCP 服务 API 的使用场景
- [共享空间说明](../安全认证/共享空间说明.md) - 组织管理 API 的业务逻辑
- [IM集成开发](../集成扩展/IM集成开发.md) - IM 渠道 API 的使用场景
- [数据源导入开发](../集成扩展/数据源导入开发.md) - 数据源 API 的使用场景
