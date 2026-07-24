# 内置模型管理指南

## 概述

内置模型是系统级别的模型配置，对所有空间可见。普通空间用户（包括空间管理员）看到的敏感信息会被隐藏，且不能修改；系统管理员可以在「模型管理」界面编辑配置和凭据。内置模型通常用于提供系统默认的模型配置，确保所有空间都能使用统一的模型服务。

## 内置模型特性

- **所有空间可见**：内置模型对所有空间都可见，无需单独配置
- **安全保护**：API Key 等凭据永远不会明文返回；只有系统管理员能看到 Base URL 等管理信息和凭据是否已配置
- **权限保护**：普通空间用户只能查看；系统管理员可以编辑配置和凭据
- **统一管理**：内置模型对所有空间生效；删除仍通过 YAML 或 SQL 管理，避免部署配置与运行时状态冲突

## 在管理界面编辑

系统管理员可直接在「设置 → 模型管理」中点击内置模型，修改模型参数或更新凭据。普通空间管理员仍然只能查看。

首次通过管理界面保存 YAML 托管的内置模型后，该行会转为运行时托管（`managed_by` 清空），后续应用启动不会再用同 ID 的 YAML 条目覆盖它。这样可以保证界面保存的结果在重启后仍然有效。若希望重新交给 YAML 管理，需要由部署管理员在数据库中把该行的 `managed_by` 恢复为 `yaml`，或移除运行时覆盖后重新启动。

## 如何添加内置模型

WeKnora 支持两种方式添加内置模型：**推荐**使用 YAML 声明式配置（自动幂等下发），保留 SQL 直插作为兼容路径。

### 方式一（推荐）：YAML 配置文件

#### 文件位置

默认路径是 `config/builtin_models.yaml`（与 `config.yaml`、`builtin_agents.yaml` 同目录）。如需挂载到其他位置，设置环境变量 `BUILTIN_MODELS_CONFIG=/absolute/path/builtin_models.yaml` 覆盖。

文件不存在时启动期会跳过、不报错；解析失败仅记录 Warning、不影响主流程。每次应用启动会重新读取并按 `id` 字段 UPSERT 到 `models` 表（保留 `created_at`，刷新其他字段）。如果同 ID 模型已经由系统管理员在界面保存并转为运行时托管，则保留运行时配置，不再由 YAML 覆盖。

#### Schema

```yaml
builtin_models:
  - id: <required, stable, UPSERT key, 最长 64 字符>
    tenant_id: <int, default 10000>          # 与 tenants_id_seq 起点对齐
    name: <string>                            # 运行时实际调用的模型名
    display_name: <string, optional>          # 界面展示名，缺省回退 name
    type: KnowledgeQA | Embedding | Rerank | VLLM | ASR
    source: <string, default "remote">        # 18 种取值，见下文
    description: <string, optional>
    is_default: <bool, default false>         # 每 (tenant_id, type) 仅一个
    status: <active | downloading | download_failed, default "active">
    parameters:
      base_url: <string>
      api_key: <string, supports ${ENV_VAR}>
      provider: <string>                      # 26 种取值，见下文
      interface_type: <string, optional>      # 接口协议类型
      parameter_size: <string, optional>      # Ollama 参数量级，如 "7B"
      supports_vision: <bool, optional>       # 是否接受图像/多模态输入
      max_concurrency: <int, optional>        # 后台调用并发上限，0=走全局
      custom_headers: <map, optional>         # 自定义 HTTP 头（保留头会被忽略）
      embedding_parameters:                   # 仅 Embedding 类型
        dimension: <int>
        truncate_prompt_tokens: <int>
        supports_dimension_override: <bool>
      extra_config: <map, optional>           # 厂商专属配置，见下文 5 个已知 key
      app_id: <string, optional>              # WeKnoraCloud 厂商专用
      app_secret: <string, optional>          # WeKnoraCloud 厂商专用（AES 加密存储）
```

##### 模型类型（type）

共 5 种（`internal/types/model.go:18-22`）：`KnowledgeQA`（对话问答）、`Embedding`（向量嵌入）、`Rerank`（重排序）、`VLLM`（vLLM 私有化部署）、`ASR`（语音识别）。loader 会强校验，未知值被拒绝。

##### 模型来源（source）

共 18 种（`internal/types/model.go:38-55`），`source` 不做枚举强校验但应使用标准值：`local`、`remote`、`aliyun`、`zhipu`、`volcengine`、`deepseek`、`hunyuan`、`minimax`、`openai`、`gemini`、`mimo`、`siliconflow`、`jina`、`openrouter`、`requesty`、`nvidia`、`novita`、`azure_openai`。

##### 服务商（provider）

共 26 种（`internal/models/provider/provider.go`，`AllProviders()` 返回值）。YAML 未填时由 `DetectProvider` 按 `base_url` 自动探测：`generic`、`weknoracloud`、`aliyun`、`zhipu`、`volcengine`、`hunyuan`、`siliconflow`、`deepseek`、`minimax`、`moonshot`、`modelscope`、`qianfan`、`qiniu`、`openai`、`anthropic`、`gemini`、`openrouter`、`requesty`、`jina`、`mimo`、`longcat`、`lkeap`、`gpustack`、`nvidia`、`novita`、`azure_openai`。`generic` 为兜底通用适配（OpenAI 兼容协议）。

##### extra_config 已知 key

`extra_config` 是 `map[string]string`，承载厂商专属配置。已知的 5 个 key：

| key | 用途 | 消费位置 |
|-----|------|----------|
| `api_version` | 指定 API 版本（Azure 等） | `internal/models/chat/remote_api.go:62` |
| `remote_model_name` | 覆盖请求上游时的模型名（与 `name` 解耦） | `internal/models/chat/remote_api.go:86` |
| `thinking_control` | 思考模式：`enabled` / `disabled` / `auto` | `internal/models/chat/thinking.go:128` |
| `secret_key` | LKEAP rerank 的 SecretKey | `internal/models/rerank/lkeap_reranker.go:36` |
| `region` | LKEAP 服务地域，缺省走 `LKEAPDefaultRegion` | `internal/models/rerank/lkeap_reranker.go:44` |

##### SYSTEM_AES_KEY 安全要求

API Key / AppSecret 在 `parameters` JSONB 中以 AES-256-GCM 加密存储，密钥取自环境变量 `SYSTEM_AES_KEY`。**该变量必须恰好 32 字节**--长度不对时 `utils.GetAESKey()` 静默返回 nil，加密功能被**静默禁用**（降级为明文兼容路径），仅在启动期打一条 `[startup-env]` 警告（`internal/runtime/startup.go:137`）。生产环境务必确认日志中无此告警。密钥轮换/丢失后历史密文解密失败，`ModelParameters.Scan` 会把对应字段置空并打 `[crypto]` 日志，保证 `ListModels` 不被单行坏数据拖垮。

#### 完整示例

```yaml
builtin_models:
  - id: builtin-openai-chat
    name: gpt-4o-mini
    type: KnowledgeQA
    source: remote
    description: OpenAI 默认对话模型
    is_default: false
    parameters:
      base_url: https://api.openai.com/v1
      api_key: ${OPENAI_API_KEY}
      provider: openai

  - id: builtin-openai-embeddings
    name: text-embedding-3-small
    type: Embedding
    source: remote
    parameters:
      base_url: https://api.openai.com/v1
      api_key: ${OPENAI_API_KEY}
      provider: openai
      embedding_parameters:
        dimension: 1536
        truncate_prompt_tokens: 0

  - id: builtin-rerank
    name: bge-reranker-v2-m3
    type: Rerank
    source: remote
    parameters:
      base_url: ${RERANK_BASE_URL}
      api_key: ${RERANK_API_KEY}
      provider: generic
```

#### `${ENV}` 插值

`api_key` / `base_url` / `name` 等任何**字符串**字段都可以引用环境变量：`${OPENAI_API_KEY}` 会在启动时被对应的 `os.Getenv("OPENAI_API_KEY")` 替换。

- 环境变量存在 → 替换为实际值
- 环境变量不存在 → **保留字面 `${OPENAI_API_KEY}`** 字符串（让 401 错误能直接看出来 env 没设，便于排查）
- 不支持 `${VAR:-default}` 这种 shell 扩展，行为与现有 `config.yaml` 的插值实现一致
- **非字符串字段不能 env 化**（如 `type`、`dimension`、`is_default`），因为它们必须按 YAML 的目标类型解析

#### env 变量怎么进入容器

`docker-compose.yml` 的 `app` 服务已经预置了：

```yaml
env_file:
  - path: .env
    required: false
```

意味着把变量值写到项目根目录的 `.env` 文件里，启动时自动透传到容器。**无需**在 `environment:` 块里逐个透传。`required: false` 保证 `.env` 不存在时容器仍可启动（适配上游 fresh clone 场景）。

仓库的 `.env.example` 顶部预留了 **Built-in Models** 注释段，列出 LLM / Embedding / Rerank 的参考变量名作为起点；复制 `.env.example` 为 `.env` 后取消注释并填值即可。变量名由 YAML 自行决定，参考段只是常见样板，不是保留字。

完整端到端示例：

`.env`
```bash
LLM_MODEL_NAME=gpt-4o-mini
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-...
LLM_PROVIDER=openai
```

`config/builtin_models.yaml`
```yaml
builtin_models:
  - id: builtin-llm-default
    type: KnowledgeQA
    is_default: true
    name: ${LLM_MODEL_NAME}
    parameters:
      base_url: ${LLM_BASE_URL}
      api_key: ${LLM_API_KEY}
      provider: ${LLM_PROVIDER}
```

启动：
```bash
docker compose up -d
```

#### 启动后验证

```bash
docker compose logs app | grep -E 'Built-in models? config'
```

会看到类似：

```
Built-in model upserted: id=builtin-openai-chat name=gpt-4o-mini type=KnowledgeQA
Built-in model upserted: id=builtin-openai-embeddings name=text-embedding-3-small type=Embedding
Built-in models config applied: 2 entries from /app/config/builtin_models.yaml.
```

#### Docker 部署

在 `docker-compose.yml` 的 `app` 服务 `volumes` 块挂载文件：

```yaml
services:
  app:
    volumes:
      - ./config/builtin_models.yaml:/app/config/builtin_models.yaml:ro
```

仓库提供了 `config/builtin_models.yaml.example` 作为起点，复制为 `config/builtin_models.yaml` 后按需修改。

### 方式二：直接 SQL 插入

支持的 provider（共 26 个，与上方 YAML 路径一致）：`generic`（自定义）、`openai`、`anthropic`、`azure_openai`、`aliyun`、`zhipu`、`volcengine`、`hunyuan`、`deepseek`、`minimax`、`mimo`、`siliconflow`、`jina`、`openrouter`、`requesty`、`gemini`、`modelscope`、`moonshot`、`qianfan`、`qiniu`、`longcat`、`lkeap`、`gpustack`、`nvidia`、`novita`、`weknoracloud`

```sql
-- 示例：LLM 内置模型
INSERT INTO models (
    id, tenant_id, name, type, source, description,
    parameters, is_default, status, is_builtin
) VALUES (
    'builtin-llm-001',
    10000,
    'gpt-4o-mini',
    'KnowledgeQA',
    'remote',
    '系统内置 LLM 模型',
    '{"base_url": "https://api.openai.com/v1", "api_key": "sk-xxx", "provider": "openai"}'::jsonb,
    false,
    'active',
    true
) ON CONFLICT (id) DO NOTHING;

-- Embedding
INSERT INTO models (
    id, tenant_id, name, type, source, description,
    parameters, is_default, status, is_builtin
) VALUES (
    'builtin-embedding-001',
    10000,
    'text-embedding-3-small',
    'Embedding',
    'remote',
    '系统内置 Embedding 模型',
    '{"base_url": "https://api.openai.com/v1", "api_key": "sk-xxx", "provider": "openai", "embedding_parameters": {"dimension": 1536, "truncate_prompt_tokens": 0}}'::jsonb,
    false,
    'active',
    true
) ON CONFLICT (id) DO NOTHING;

-- Rerank
INSERT INTO models (
    id, tenant_id, name, type, source, description,
    parameters, is_default, status, is_builtin
) VALUES (
    'builtin-rerank-001',
    10000,
    'bge-reranker-v2-m3',
    'Rerank',
    'remote',
    '系统内置 Rerank 模型',
    '{"base_url": "https://api.jina.ai/v1", "api_key": "jina-xxx", "provider": "jina"}'::jsonb,
    false,
    'active',
    true
) ON CONFLICT (id) DO NOTHING;
```

### 验证插入结果

```sql
SELECT id, name, type, is_builtin, status
FROM models
WHERE is_builtin = true
ORDER BY type, created_at;
```

## 将现有模型设置为内置模型

如果你已经手工创建了一个普通模型，想把它升级为内置模型：

```sql
UPDATE models
SET is_builtin = true
WHERE id = '模型ID';
```

## 移除内置模型

**从 YAML 删除条目即可。** 应用启动时会自动软删除 `models` 表中 YAML 不再声明的 YAML 托管行 —— 你不再需要手工跑 SQL。

工作原理：每条由 YAML 写入的行会被打上 `managed_by = 'yaml'` 标记。重启时 loader 走两步：

1. UPSERT 当前 YAML 中的所有条目（按 `id` 幂等，包含把之前软删过的 `deleted_at` 重置为 NULL —— 也就是说从 YAML 拿掉再加回来等于"复活"）
2. 软删除 `is_builtin = true AND managed_by = 'yaml' AND id NOT IN (当前 YAML 中的 id 集合)` 的行

**手工通过 SQL 插入的 builtin 行（`managed_by = ''`）永远不会被 loader 触碰**，与 YAML 完全隔离。

### 手工路径补充

如果你是走 SQL 路径管理的（`managed_by = ''`），删除仍然走老方法：

```sql
-- 取消 builtin 标记，恢复为普通模型
UPDATE models SET is_builtin = false WHERE id = '模型ID';

-- 或直接删除
DELETE FROM models WHERE id = '模型ID';
```

### 紧急关闭 YAML 接管

如果误改了 YAML 想立刻停用接管又不想清空文件，最快的方法是：把环境变量 `BUILTIN_MODELS_CONFIG` 指向一个不存在的路径并重启 —— loader 看到文件缺失会直接 no-op，**包括跳过 drift sweep**，已经写入的 YAML 托管行保留原状。

## 注意事项

1. **ID 命名规范**：建议使用 `builtin-{type}-{slug}` 的格式，例如 `builtin-openai-chat`、`builtin-rerank`
2. **空间ID**：内置模型可以属于任意空间，默认 `10000`（与 `tenants_id_seq` 起点一致）
3. **YAML 与 SQL 并存**：两种方式可以同时使用，loader 只动 `managed_by='yaml'` 的行；通过 SQL 插入的 builtin 行对 loader 完全不可见
4. **`is_default` 单一保证**：YAML 中将某条 entry 标记 `is_default: true` 时，loader 会先把同 `(tenant_id, type)` 下的其它默认模型置为 `false`，避免 API 路径维护的"每类型一个默认模型"语义被破坏
5. **重启即生效**：修改 YAML 后 `docker compose restart app` 即可让新配置生效
6. **加密**：API Key 在 `parameters` JSONB 中以加密形式存储（若 `SYSTEM_AES_KEY` 已配置），未配置时降级为明文兼容路径
7. **安全性**：前端会自动隐藏内置模型的 API Key 和 Base URL，但数据库中的原始数据仍然存在，请妥善保管数据库访问权限
8. **解析错误自我保护**：YAML 解析失败时 loader 仅打 warning 并跳过 reconcile，**不会**执行 drift sweep，确保一个手抖的 YAML 改动不会大规模软删既有内置模型
