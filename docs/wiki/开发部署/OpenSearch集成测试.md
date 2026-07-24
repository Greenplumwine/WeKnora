---
title: OpenSearch集成测试
tags: [开发部署, OpenSearch, 向量库, 集成测试, k-NN]
aliases: [opensearch-integration-test, OpenSearch测试, k-NN测试]
---

# OpenSearch k-NN 驱动 - 本地集成测试

本指南会拉起一个单节点 OpenSearch 集群,并对 OpenSearch 检索引擎做端到端验证。驱动代码位于 `internal/application/repository/retriever/opensearch/`。

> OpenSearch 作为向量库的启用方法参见 [常见问题](../运维排障/常见问题.md) 第 15 条。

## 1. 启动开发集群

```bash
docker compose -f docker-compose.dev.yml --profile opensearch up -d
```

这会启动:

- `opensearch`,监听 `http://localhost:9200` - 单节点,**安全插件已禁用**(纯 HTTP,无鉴权/TLS)。该镜像内置了 `opensearch-knn` 插件。

> **OpenSearch Dashboards 是可选的**,它位于独立的 `opensearch-ui` profile 下,因此 *不会* 被 `--profile opensearch` 启动。下面的整个集成测试都可以用 curl 对 `:9200` 完成验证。如果需要 Web UI(Dev Tools 控制台 / 可视化索引检视),按需启动:
>
> ```bash
> docker compose -f docker-compose.dev.yml --profile opensearch-ui up -d
> # opensearch-dashboards 监听 http://localhost:5601(depends_on 会自动拉起集群)
> ```

验证:

```bash
curl -s localhost:9200 | jq '.version.distribution, .version.number'
# "opensearch" "3.3.2"
curl -s 'localhost:9200/_cat/plugins?format=json' | jq -r '.[].component' | grep opensearch-knn
```

> 生产集群必须启用安全插件(TLS + 鉴权)。dev profile 之所以禁用它,只是为了让本地搭建尽量简单。连接受保护集群时,需设置 `username` / `password`,仅在 dev 环境遇到自签名证书时才设 `insecure_skip_verify=true`。

## 2. 注册存储

### 方案 A - DB 存储(UI / API)

> **SSRF 白名单(dev)。** `CreateStore` 和原始连接测试会用 SSRF 策略校验用户填入的 `addr`。`http://localhost:9200` 默认会被拒绝 - `localhost` 是受限主机名,`9200` 是被封锁的端口。当后端跑在宿主机上(`go run`)时,注册前需在 `.env` 里把 `localhost` 加入白名单:
>
> ```bash
> SSRF_WHITELIST=localhost
> ```
>
> 容器化的 compose 部署会自动把内置的向量库服务名加入白名单(`SSRF_WHITELIST_EXTRA`),所以这一步只用于 dev。env 存储路径(方案 B)不受影响。

`POST /api/v1/vector-stores`:

```json
{
  "name": "opensearch-local",
  "engine_type": "opensearch",
  "connection_config": { "addr": "http://localhost:9200" },
  "index_config": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "hnsw_m": 16,
    "hnsw_ef_construction": 100,
    "knn_engine": "lucene"
  }
}
```

`CreateStore` 在落库前会先跑连接探测(版本 + k-NN 插件);地址错误 / 版本不支持 / 缺插件都会以 `400` 拒绝。

### 方案 B - env 存储

```bash
export RETRIEVE_DRIVER=opensearch
export OPENSEARCH_ADDR=http://localhost:9200
# 受保护集群需 export OPENSEARCH_USERNAME / OPENSEARCH_PASSWORD
# export OPENSEARCH_INSECURE_SKIP_VERIFY=true   # 仅 dev 自签名 TLS
```

## 3. 单节点注意事项(重要)

在单节点集群上,任何用 `number_of_replicas >= 1` 创建的索引,其副本分片都会处于 **unassigned** 状态,索引健康度因此变成 **Yellow**。Yellow **不会**阻塞读写 - 本地测试是安全的 - 但若想让集群保持 Green,注册存储时请设 **`number_of_replicas: 0`**(如上方方案 A 示例)。驱动默认值是 `1`(假定集群至少 2 节点)。

## 4. 走通流程

1. 把一个知识库绑定到该存储,并灌入若干文档。
2. 确认按维度建好的索引已出现:
   `curl -s 'localhost:9200/_cat/indices?v' | grep weknora`
   (例如 `weknora_<storeprefix>_768` 及其别名,外加 `weknora_<storeprefix>_keywords`)。
3. 对绑定的 KB 发起一次检索查询,确认能召回结果。
4. 把该 KB 复制到另一个 KB,确认文档被重新索引
   (`opensearch.reindex_executed` 审计事件)。
5. 切换 chunk 的启用状态 / 标签,确认 `_update_by_query` 生效。

## 5. 销毁

```bash
docker compose -f docker-compose.dev.yml --profile opensearch down -v
```

## 范围说明

- 大批量异步 reindex / 删除(任务轮询)属于后续工作;同步路径足以处理常规 KB 规模(分页受 `max_result_window` 约束,默认 10000)。
- 原生 `hybrid` 查询 + search pipeline 不在范围内 - 融合仍留在服务层(RRF)。

## 相关主题

- [集成向量数据库](../集成扩展/集成向量数据库.md) - 向量数据库的扩展开发
- [常见问题](../运维排障/常见问题.md) - OpenSearch 启用（第 15 条）与 SSRF 白名单（第 8 条）
- [开发指南](开发指南.md) - 本地开发环境

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
- [集成向量数据库](../集成扩展/集成向量数据库.md) - 向量库扩展开发
- [开发指南](开发指南.md) - 开发环境搭建
