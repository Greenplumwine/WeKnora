---
title: WeKnora Wiki 首页
tags: [首页, 导航, MOC]
aliases: [Home, Index, wiki首页]
---

# WeKnora Wiki

欢迎使用 WeKnora 知识库 wiki！这里是 WeKnora 项目文档的互联知识网络，所有页面通过双向链接关联，帮助你从任意入口探索整个知识体系。

---

## 项目概述

| 页面 | 简介 |
|------|------|
| [版本路线图](项目概述/版本路线图.md) | 产品规划与计划方向 |
| [Lite与标准版区别](项目概述/Lite与标准版区别.md) | 轻量版与标准版的功能对比 |

## 核心功能

| 页面 | 简介 |
|------|------|
| [知识图谱](核心功能/知识图谱.md) | Neo4j 知识图谱的启用流程与使用 |
| [文档分块策略](核心功能/文档分块策略.md) | 自适应三级分块策略与设置参考 |
| [MCP功能使用说明](核心功能/MCP功能使用说明.md) | MCP 服务的用户操作指南 |
| [MCP工具人工审核](核心功能/MCP工具人工审核.md) | MCP 危险工具的人工审核机制 |
| [内置MCP服务管理](核心功能/内置MCP服务管理.md) | 内置 MCP 服务的系统级管理 |
| [内置模型管理](核心功能/内置模型管理.md) | 内置模型的系统级管理 |
| [内置智能体管理](核心功能/内置智能体管理.md) | 内置智能体的 YAML 配置与模型字段 |
| [Agent技能系统](核心功能/Agent技能系统.md) | Agent Skills 扩展机制与预加载技能 |

## 集成与扩展

| 页面 | 简介 |
|------|------|
| [IM集成开发](集成扩展/IM集成开发.md) | 企业即时通讯平台接入开发 |
| [数据源导入开发](集成扩展/数据源导入开发.md) | 外部平台数据自动同步与导入 |
| [添加网络搜索引擎](集成扩展/添加网络搜索引擎.md) | 扩展新的网络搜索 Provider |
| [集成向量数据库](集成扩展/集成向量数据库.md) | 集成新的向量数据库检索引擎 |
| [网页嵌入集成](集成扩展/网页嵌入集成.md) | 网站嵌入 WeKnora 智能体（iframe/浮窗/安全模式/子域部署） |
| [Langfuse集成](集成扩展/Langfuse集成.md) | Langfuse 全链路可观测性追踪 |

## 安全与认证

| 页面 | 简介 |
|------|------|
| [OIDC认证调用流程](安全认证/OIDC认证调用流程.md) | OIDC 第三方登录的完整调用链路 |
| [空间RBAC说明](安全认证/RBAC说明.md) | 空间内角色矩阵、资源归属与审计 |
| [共享空间说明](安全认证/共享空间说明.md) | 跨空间协作与知识库/智能体共享 |

## 开发与部署

| 页面 | 简介 |
|------|------|
| [开发指南](开发部署/开发指南.md) | 本地开发环境搭建与工作流 |
| [快速开发模式](开发部署/快速开发模式.md) | 后端/前端热更新开发模式 |
| [Worker池治理](开发部署/Worker池治理.md) | 后台任务分阶段 Worker 池与并发治理 |
| [云镜像打包](开发部署/云镜像打包.md) | 可分发云镜像的制作与上架 |
| [腾讯云轻量镜像](开发部署/腾讯云轻量镜像.md) | 腾讯云 Lighthouse/CVM 镜像制作 |
| [OpenSearch集成测试](开发部署/OpenSearch集成测试.md) | OpenSearch k-NN 驱动本地集成测试 |

## 运维与排障

| 页面 | 简介 |
|------|------|
| [常见问题](运维排障/常见问题.md) | 部署与使用中的 FAQ |
| [日志配置](运维排障/日志配置.md) | 环境变量驱动的日志级别/格式/落盘 |
| [迁移排障](运维排障/迁移排障.md) | 数据库迁移失败的诊断与恢复 |

## API 参考

| 页面 | 简介 |
|------|------|
| [API文档概览](API参考/API文档概览.md) | RESTful API 基础信息与分类索引 |

> 各 API 模块的详细说明见 [API文档概览](API参考/API文档概览.md) 中的分类表，包括 auth / knowledge / model / agent / session / chat / mcp-service / organization / vector-store 等 20 余个模块文档。

---

## 知识图谱

```mermaid
graph TB
    Home[WeKnora Wiki]
    Home --> 项目概述
    Home --> 核心功能
    Home --> 集成扩展
    Home --> 安全认证
    Home --> 开发部署
    Home --> 运维排障
    Home --> API参考

    项目概述 --> ROADMAP[版本路线图]
    项目概述 --> LITE[Lite与标准版区别]

    核心功能 --> KG[知识图谱]
    核心功能 --> Chunking[文档分块策略]
    核心功能 --> MCP[MCP功能使用说明]
    核心功能 --> MCPApproval[MCP工具人工审核]
    核心功能 --> BuiltinMCP[内置MCP服务管理]
    核心功能 --> BuiltinModel[内置模型管理]
    核心功能 --> BuiltinAgent[内置智能体管理]
    核心功能 --> Skills[Agent技能系统]

    集成扩展 --> IM[IM集成开发]
    集成扩展 --> DS[数据源导入开发]
    集成扩展 --> WebSearch[添加网络搜索引擎]
    集成扩展 --> VecDB[集成向量数据库]
    集成扩展 --> EmbedInt[网页嵌入集成]
    集成扩展 --> Langfuse[Langfuse集成]

    安全认证 --> OIDC[OIDC认证调用流程]
    安全认证 --> RBAC[空间RBAC说明]
    安全认证 --> SharedSpace[共享空间说明]

    开发部署 --> DevGuide[开发指南]
    开发部署 --> QuickDev[快速开发模式]
    开发部署 --> WorkerPool[Worker池治理]
    开发部署 --> CloudImg[云镜像打包]
    开发部署 --> TencentLH[腾讯云轻量镜像]
    开发部署 --> OSImg[OpenSearch集成测试]

    运维排障 --> FAQ[常见问题]
    运维排障 --> LogConfig[日志配置]
    运维排障 --> Migrate[迁移排障]

    API参考 --> APIOverview[API文档概览]

    KG -.-> Chunking
    MCP -.-> BuiltinMCP
    MCP -.-> MCPApproval
    BuiltinMCP -.-> BuiltinModel
    BuiltinModel -.-> BuiltinAgent
    Skills -.-> IM
    IM -.-> DS
    DS -.-> SharedSpace
    OIDC -.-> RBAC
    RBAC -.-> SharedSpace
    OIDC -.-> SharedSpace
    LITE -.-> SharedSpace
    WebSearch -.-> VecDB
    DevGuide -.-> QuickDev
    FAQ -.-> DevGuide
    FAQ -.-> LogConfig
    FAQ -.-> Migrate
    EmbedInt -.-> Langfuse
    Langfuse -.-> WorkerPool
    CloudImg -.-> TencentLH
    DevGuide -.-> OSImg
```

---

## 反向链接

本页为 wiki 首页，所有页面均链接回此处。
