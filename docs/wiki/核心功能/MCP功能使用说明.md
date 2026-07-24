---
title: MCP功能使用说明
tags: [核心功能, MCP, 工具集成, 人工审核, 安全]
aliases: [MCP使用, MCP功能, MCP工具人工审核, mcp-approval, MCP审核, 工具审核, tool-approval]
---

# MCP 功能使用说明

## 功能概述

- MCP（Model Context Protocol）让 WeKnora 可以安全地连接外部工具或数据源，扩展 Agent 在推理时可调用的能力。
- 在前端 `设置 > MCP 服务`（`frontend/src/views/settings/McpSettings.vue`）中集中管理所有服务，无需手动改配置文件。
- 每个服务都包含名称、传输方式（SSE / HTTP Streamable）、连接地址、认证信息以及高级超时与重试策略。出于安全考虑，`stdio` 传输方式在服务端已被禁用，仅支持 SSE 与 HTTP Streamable。

> 关于系统级的内置 MCP 服务管理，参见 [内置MCP服务管理](内置MCP服务管理.md)

## 入口与界面

- 打开控制台左侧菜单 `设置 -> MCP 服务`，即可看到当前空间下的所有 MCP 服务列表。
- 列表中可快速启停服务、查看描述，并通过右侧菜单执行"测试 / 编辑 / 删除"。
- "添加服务"按钮会弹出 `McpServiceDialog`，用于创建或修改服务。

## 常用操作流程

### 1. 新建服务

- 点击"添加服务"，填写名称与描述，选择传输方式（SSE 或 HTTP Streamable）。
- SSE / HTTP Streamable 需提供可访问的服务 URL；可附加自定义 HTTP 头与环境变量。
- 根据需要填写 API Key、Bearer Token、超时与重试策略，保存后服务会出现在列表中。

### 2. 启停服务

- 在列表开关中切换启用状态，系统会即时调用后端 `updateMCPService`，失败时会自动回滚状态并弹出提示。

### 3. 连接测试

- 通过更多菜单选择"测试"，前端会调用 `/api/v1/mcp-services/{id}/test` 并弹出 `McpTestResult`。
- 成功时会展示服务可用的工具清单（含输入 schema）和资源列表；失败时会显示错误信息，方便排查网络或鉴权问题。

### 4. 编辑 / 删除

- "编辑"会带出原有配置，修改后保存即可。
- "删除"需要在弹窗中确认，完成后列表自动刷新。

## 工具人工审核（危险调用）

对应需求：智能体调用 MCP 工具前可中断，待人工确认后再执行（GitHub #1173）。

### 行为说明

1. 在 **设置 -> MCP** 中连接测试成功后，在工具列表上打开 **「需人工审核」** 开关，即可为该工具打标。
2. Agent 运行时若即将调用已打标的工具，会推送 `tool_approval_required` 事件，对话界面展示审批卡片（可编辑 JSON 参数）。
3. 用户 **通过** 或 **拒绝** 后，后端恢复执行；拒绝时工具返回错误信息给模型，不会调用远端 MCP。
4. 若超时未处理（默认 10 分钟，可通过配置 `agent.tool_approval_timeout_seconds` 调整），视为拒绝。

### 配置示例

任选一种：

**1. config.yaml**

```yaml
agent:
  tool_approval_timeout_seconds: 600  # 可选，默认 600（秒）
```

**2. 环境变量**（优先级高于 yaml）

```bash
# 支持纯秒数或 Go duration（30s / 5m / 1h）
WEKNORA_AGENT_TOOL_APPROVAL_TIMEOUT=600
```

### 审核 API

- `GET /api/v1/mcp-services/:id/tool-approvals` - 列出已保存的审核配置
- `PUT /api/v1/mcp-services/:id/tool-approvals/:tool_name` - 设置某工具是否需审核（`{"require_approval": true}`）
- `POST /api/v1/agent/tool-approvals/:pending_id` - 在审批卡片中提交结果
  - body: `{"decision":"approve"|"reject","modified_args":{...}可选,"reason":"..."可选}`

### 部署与限制

- **审批等待状态保存在进程内存** 中：`pending_id` 仅对当前实例有效；进程重启后进行中的等待会失败（表现为拒绝/取消）。
- **多副本部署**：当配置了 `REDIS_ADDR` 时，`Resolve` 会通过 Redis Pub/Sub（频道 `weknora:mcp_approval:resolve`）跨实例转发，因此 SSE 与提交审批的 HTTP 请求落到不同实例也能正确唤醒等待者；未配置 Redis 时退化为单机模式，需要使用会话粘滞（sticky session）。
- **审批等待不会被工具默认 60s 超时取消**：审批阶段使用 round 级别的 ctx（不带 `defaultToolExecTimeout`），仅受 `agent.tool_approval_timeout_seconds` 与请求级取消控制。
- 安全边界：审核通过后的参数仍由当前登录空间提交；请仅在可信环境下授予「通过」权限。

## 使用建议

- **传输方式选择**：优先使用 SSE 获取流式体验；需要标准 HTTP Streamable 兼容时再切换。注意 `stdio` 已被禁用，本地调试也需以 SSE 或 HTTP Streamable 方式启动 MCP Server。
- **鉴权管理**：将 API Key / Token 保存在"认证配置"中，生产环境建议单独创建最小权限 Key，并定期轮换。
- **重试策略**：对公网或第三方服务适当提高 `retry_count` 与 `retry_delay`，避免间歇性超时导致 Agent 中断。
- **危险工具审核**：对有写副作用或不可逆操作的工具（删数据、发消息、调外部 API）建议开启「需人工审核」，避免 Agent 误操作。

## 相关主题

- [内置MCP服务管理](内置MCP服务管理.md) - 系统管理员视角的内置 MCP 服务配置
- [Agent技能系统](Agent技能系统.md) - 另一种 Agent 扩展机制
- [IM集成开发](../集成扩展/IM集成开发.md) - IM 渠道中 Agent 使用 MCP 工具
- [添加网络搜索引擎](../集成扩展/添加网络搜索引擎.md) - 扩展搜索能力的另一种方式

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
- [内置MCP服务管理](内置MCP服务管理.md) - MCP 的系统级管理（管理员视角）
- [Agent技能系统](Agent技能系统.md) - 与 MCP 并列的 Agent 扩展机制
- [IM集成开发](../集成扩展/IM集成开发.md) - Agent 在 IM 渠道中可调用 MCP 工具
