---
title: Worker池治理
tags: [开发部署, Worker, 任务队列, Asynq, 并发治理]
aliases: [worker-pool-governance, WorkerPool, 任务池治理]
---

# Worker Pool Governance

WeKnora uses guaranteed per-stage worker pools plus an elastic pool for the
document ingestion pipeline. Worker concurrency is a scheduling budget, not a
replacement for model quotas, DocReader capacity, vector-store limits, or
database connection limits.

> 运行时队列面板与故障排查参见 [常见问题](../运维排障/常见问题.md) 第 30 条。

## Topology

| Pool | Default per instance | Queues | Purpose |
| --- | ---: | --- | --- |
| Core | 8 | `default` | Document parsing and manual reparse guarantee |
| Post-process | 2 | `postprocess` | Parse finalization and enrichment fan-out |
| Enrichment | 12 | `summary`, `multimodal`, `graph`, `question` | Model-heavy enrichment guarantee |
| Maintenance | 4 | `sync`, `low` | Source sync, batch work, move/delete/cleanup |
| Shared | 6 | Core and enrichment queues | Elastic capacity borrowed by the side with backlog |
| Wiki | 8 | `wiki` | Independently governed Wiki generation |

The upstream total remains 32 workers per service instance by default. Wiki is
separate and is not included in that total.

Post-process has its own physical queue so lightweight fan-out work cannot sit
behind long DocReader calls. Maintenance is intentionally excluded from the
shared pool because its long-running tasks could pin elastic capacity needed by
the user-facing pipeline.

Asynq dequeue is atomic. Dedicated and shared servers may safely subscribe to
the same core/enrichment queues; one task is still processed by one worker.

## Configuration

All settings are available under System settings and require a service restart:

- `asynq.core_concurrency` / `WEKNORA_ASYNQ_CORE_CONCURRENCY`
- `asynq.postprocess_concurrency` / `WEKNORA_ASYNQ_POSTPROCESS_CONCURRENCY`
- `asynq.enrichment_concurrency` / `WEKNORA_ASYNQ_ENRICHMENT_CONCURRENCY`
- `asynq.maintenance_concurrency` / `WEKNORA_ASYNQ_MAINTENANCE_CONCURRENCY`
- `asynq.shared_concurrency` / `WEKNORA_ASYNQ_SHARED_CONCURRENCY`
- `asynq.wiki_concurrency` / `WEKNORA_WIKI_ASYNQ_CONCURRENCY`

The old aggregate `asynq.concurrency` / `WEKNORA_ASYNQ_CONCURRENCY` setting is
retired. Deployments that set it must migrate to the explicit pool settings.
Persisted old rows are ignored and hidden from the System settings page so they
cannot be mistaken for an effective runtime control.

## Capacity layers

Tune these layers independently:

1. Worker concurrency controls the number of task handlers admitted per
   service instance.
2. Model quota governance controls provider concurrency, RPM, and TPM across
   replicas and quota groups.
3. DocReader, vector stores, object storage, Postgres, and local CPU/RAM retain
   their own resource limits.

If model limiter waiting grows while worker queues are busy, adding workers
only creates more waiters. Increase the provider quota or reduce worker
admission instead.

## Sizing

For each pool, estimate peak required workers using:

```text
required workers = ceil(peak task arrival rate * mean task runtime / 0.70)
```

Use task arrival rate after fan-out. One parsed document can create one summary,
multiple question batches, one graph task per chunk, and multiple image tasks.
Queue count is not a capacity signal.

The runtime dashboard reports:

- configured concurrency per instance;
- live server instance count;
- cluster capacity aggregated from active asynq heartbeats;
- active workers and utilization;
- queue backlog, retries, dead letters, and oldest pending age;
- model concurrency and limiter waiting.

Increase a pool only when backlog age grows while its downstream dependency has
headroom. Reallocate to enrichment when core stays underutilized and fan-out
queues grow. Reduce core admission when DocReader CPU, memory, or latency is the
bottleneck.

## 相关主题

- [Langfuse集成](../集成扩展/Langfuse集成.md) - 任务执行链路的可观测性追踪
- [开发指南](开发指南.md) - 后台任务的服务启动
- [常见问题](../运维排障/常见问题.md) - 任务积压排查（第 30 条）

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
- [开发指南](开发指南.md) - 服务启动与并发配置
- [常见问题](../运维排障/常见问题.md) - 后台任务积压与失败排查
