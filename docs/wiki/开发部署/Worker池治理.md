---
title: Worker池治理
tags: [开发部署, Worker, 任务队列, Asynq, 并发治理]
aliases: [worker-pool-governance, WorkerPool, 任务池治理]
---

# Worker 池治理

WeKnora 为文档摄入流水线采用按阶段保底的 worker 池加一个弹性池。Worker 并发是一种调度预算,并不能替代模型配额、DocReader 容量、向量存储限制或数据库连接限制。

> 运行时队列面板与故障排查参见 [常见问题](../运维排障/常见问题.md) 第 30 条。

## 拓扑

| 池 | 每实例默认 | 队列 | 用途 |
| --- | ---: | --- | --- |
| Core | 8 | `default`, `chat_attachment` | 文档解析与手工重解析保底,以及会话级聊天附件解析 |
| Post-process | 2 | `postprocess` | 解析收尾与富化扇出 |
| Enrichment | 12 | `summary`, `multimodal`, `graph`, `question` | 模型密集型富化保底 |
| Maintenance | 4 | `sync`, `low` | 数据源同步、批处理、移动/删除/清理 |
| Shared | 6 | Core 与 Enrichment 队列 | 由有积压的一侧借用的弹性容量 |
| Wiki | 8 | `wiki` | 独立治理的 Wiki 生成 |

上游总量默认仍为每服务实例 32 个 worker。Wiki 独立运行,不计入该总量。

Post-process 拥有独立的物理队列,这样轻量的扇出工作不会排在长时间的 DocReader 调用之后。Maintenance 被有意排除在共享池之外,因为其长时任务可能占住用户面流水线所需的弹性容量。

Asynq 出队是原子的。专用服务器与共享服务器可以安全订阅同一批 core/enrichment 队列;一个任务仍只由一个 worker 处理。

## 配置

所有设置都位于「系统设置」下,且修改后需要重启服务:

- `asynq.core_concurrency` / `WEKNORA_ASYNQ_CORE_CONCURRENCY`
- `asynq.postprocess_concurrency` / `WEKNORA_ASYNQ_POSTPROCESS_CONCURRENCY`
- `asynq.enrichment_concurrency` / `WEKNORA_ASYNQ_ENRICHMENT_CONCURRENCY`
- `asynq.maintenance_concurrency` / `WEKNORA_ASYNQ_MAINTENANCE_CONCURRENCY`
- `asynq.shared_concurrency` / `WEKNORA_ASYNQ_SHARED_CONCURRENCY`
- `asynq.wiki_concurrency` / `WEKNORA_WIKI_ASYNQ_CONCURRENCY`

旧的聚合配置 `asynq.concurrency` / `WEKNORA_ASYNQ_CONCURRENCY` 已废弃。设置了该项的部署必须迁移到显式的分池配置。已持久化的旧行会被忽略,并从「系统设置」页面隐藏,以免被误当作有效的运行时控制。

## 容量层级

以下层级可独立调整:

1. Worker 并发控制每个服务实例准入的任务处理器数量。
2. 模型配额治理控制跨副本和配额组的供应商并发、RPM 与 TPM。
3. DocReader、向量存储、对象存储、Postgres 以及本地 CPU/RAM 保留各自的资源上限。

若 worker 队列繁忙的同时模型限流器等待时长增长,仅增加 worker 只会产生更多等待者。此时应提升供应商配额,或减少 worker 准入。

## 容量规划

对每个池,使用下式估算峰值所需 worker:

```text
required workers = ceil(peak task arrival rate * mean task runtime / 0.70)
```

使用扇出之后的任务到达率。一个已解析文档可能产生一次摘要、多个问题批次、每个 chunk 一个图谱任务,以及多个图片任务。队列数量并不是容量信号。

运行时面板上报:

- 每实例已配置并发;
- 存活的服务器实例数;
- 由活动 asynq 心跳聚合得到的集群容量;
- 活动 worker 数与利用率;
- 队列积压、重试、死信以及最早待处理任务年龄;
- 模型并发与限流器等待。

仅当某池的下游依赖尚有余量且其积压年龄增长时,才增加该池容量。当 core 长期未充分利用而扇出队列增长时,将容量重分配给 enrichment。当 DocReader 的 CPU、内存或延迟成为瓶颈时,降低 core 准入。

## 相关主题

- [Langfuse集成](../集成扩展/Langfuse集成.md) - 任务执行链路的可观测性追踪
- [开发指南](开发指南.md) - 后台任务的服务启动
- [常见问题](../运维排障/常见问题.md) - 任务积压排查（第 30 条）

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
- [开发指南](开发指南.md) - 服务启动与并发配置
- [常见问题](../运维排障/常见问题.md) - 后台任务积压与失败排查
