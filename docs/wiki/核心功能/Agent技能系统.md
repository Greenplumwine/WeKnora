---
title: Agent技能系统
tags: [核心功能, Agent, Skills, 技能, 沙箱]
aliases: [Agent Skills, 技能系统, agent-skills]
---

# Agent 技能系统

## 概述

Agent Skills 是一种让 Agent 通过阅读"使用说明书"来学习新能力的扩展机制。与传统的硬编码工具不同，Skills 通过注入到 System Prompt 来扩展 Agent 的能力，遵循 **Progressive Disclosure（渐进式披露）** 的设计理念。目前仅支持带**智能推理**能力的智能体使用。前端可在智能体的编辑页面找到相关配置。

### 核心特性

- **非侵入式扩展**：不影响原有 Agent ReAct 流程
- **按需加载**：三级渐进式加载，优化 Token 使用
- **沙箱执行**：脚本在隔离环境中安全执行
- **灵活配置**：支持多目录、白名单过滤

> Skills 与 [MCP](MCP功能使用说明.md) 是两种不同的 Agent 扩展机制：Skills 通过 Prompt 注入，MCP 通过协议调用外部工具。

## 设计理念

### Progressive Disclosure（渐进式披露）

Skills 采用三级加载机制，确保只在需要时才向 LLM 提供详细信息：

```
┌─────────────────────────────────────────────────────────────────┐
│ Level 1: 元数据 (Metadata)                                      │
│ • 始终加载到 System Prompt                                       │
│ • 约 100 tokens/skill                                           │
│ • 包含：技能名称 + 简短描述                                       │
└─────────────────────────────────────────────────────────────────┘
                              ↓ 用户请求匹配时
┌─────────────────────────────────────────────────────────────────┐
│ Level 2: 指令 (Instructions)                                    │
│ • 通过 read_skill 工具按需加载                                   │
│ • SKILL.md 的指令内容                                           │
│ • 包含：详细指令、代码示例、使用方法                               │
└─────────────────────────────────────────────────────────────────┘
                              ↓ 需要更多信息时
┌─────────────────────────────────────────────────────────────────┐
│ Level 3: 附加资源 (Resources)                                   │
│ • 通过 read_skill 工具加载特定文件                               │
│ • 补充文档、配置模板、脚本文件                                    │
│ • 通过 execute_skill_script 执行脚本                            │
└─────────────────────────────────────────────────────────────────┘
```

## Skill 目录结构

每个 Skill 是一个目录，包含 `SKILL.md` 主文件和可选的附加资源：

```
my-skill/
├── SKILL.md           # 必需：主文件（含 YAML frontmatter）
├── REFERENCE.md       # 可选：补充文档
├── templates/         # 可选：模板文件
│   └── config.yaml
└── scripts/           # 可选：可执行脚本
    ├── analyze.py
    └── generate.sh
```

## SKILL.md 格式

### YAML Frontmatter

每个 `SKILL.md` 必须以 YAML frontmatter 开头，定义元数据：

```markdown
---
name: pdf-processing
description: Extract text and tables from PDF files, fill forms, merge documents. Use when working with PDF files or when the user mentions PDFs, forms, or document extraction.
---

# PDF Processing

This skill provides utilities for working with PDF documents.

## Quick Start

Use pdfplumber to extract text from PDFs:

​```python
import pdfplumber

with pdfplumber.open("document.pdf") as pdf:
    text = pdf.pages[0].extract_text()
    print(text)
​```
```

### 元数据验证规则

| 字段 | 要求 |
|------|------|
| `name` | 1-50 字符，仅允许汉字、英文字母、数字，不能是保留词 |
| `description` | 1-500 字符，描述技能用途和触发条件 |

**保留词**：`system`, `default`, `internal`, `core`, `base`, `root`, `admin`

## 配置

### AgentConfig 配置项

```go
type AgentConfig struct {
    // ... 其他配置 ...

    // Skills 相关配置
    SkillsEnabled  bool     `json:"skills_enabled"`   // 是否启用 Skills
    SkillDirs      []string `json:"skill_dirs"`       // Skill 目录列表
    AllowedSkills  []string `json:"allowed_skills"`   // 白名单（空=全部允许）
}
```

### 配置示例

```json
{
  "skills_enabled": true,
  "skill_dirs": [
    "/path/to/project/skills",
    "/home/user/.agent-skills"
  ],
  "allowed_skills": ["pdf-processing", "code-review"]
}
```

### Sandbox 配置（环境变量）

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `WEKNORA_SANDBOX_MODE` | sandbox 模式: `docker`, `local`, `disabled` | `disabled` |
| `WEKNORA_SANDBOX_TIMEOUT` | 脚本执行超时（秒） | `60` |
| `WEKNORA_SANDBOX_DOCKER_IMAGE` | 自定义 Docker 镜像 | `wechatopenai/weknora-sandbox:latest` |

### Sandbox 模式

| 模式 | 说明 |
|------|------|
| `docker` | 使用 Docker 容器隔离（推荐） |
| `local` | 本地进程执行（基础安全限制） |
| `disabled` | 禁用脚本执行 |

## Agent 工具

Skills 功能通过两个工具与 Agent 交互：

### read_skill

读取技能内容或特定文件。

**参数**：
```json
{
  "skill_name": "pdf-processing",      // 必需：技能名称
  "file_path": "FORMS.md"              // 可选：相对路径
}
```

**使用场景**：
1. 加载 Level 2 内容：仅传 `skill_name`
2. 加载 Level 3 资源：同时传 `skill_name` 和 `file_path`

**示例调用**：
```json
// 加载技能主内容
{"skill_name": "pdf-processing"}

// 加载补充文档
{"skill_name": "pdf-processing", "file_path": "FORMS.md"}

// 查看脚本内容
{"skill_name": "pdf-processing", "file_path": "scripts/analyze.py"}
```

### execute_skill_script

在沙箱中执行技能脚本。

**参数**：
```json
{
  "skill_name": "pdf-processing",           // 必需：技能名称
  "script_path": "scripts/analyze.py",      // 必需：脚本相对路径
  "args": ["input.pdf", "--format", "json"] // 可选：命令行参数
}
```

**支持的脚本类型**：Python (`.py`)、Shell (`.sh`)、JavaScript/Node.js (`.js`)、Ruby (`.rb`)、Go (`.go`)

## 预加载技能（Preloaded Skills）

系统内置了以下 5 个预加载技能，用于增强知识库问答和文档处理能力：

### 1. citation-generator - 引用生成器

**用途**：自动生成规范引用格式

**触发场景**：需要生成参考文献、标注知识库内容出处、要求提供引用信息

| 功能 | 说明 |
|------|------|
| 来源标注 | 为回答中使用的每个知识点标注来源 |
| 格式化引用 | 支持 APA、MLA、Chicago、简化格式 |
| 参考文献列表 | 在回答末尾生成完整的参考文献列表 |

**简化引用格式示例**：
```
根据公司政策[员工手册2024.pdf, 第15页]，年假申请需提前...
```

### 2. data-processor - 数据处理器

**用途**：数据处理与分析

**触发场景**："分析这些数据"、"统计一下"、"计算总数/平均值"、"转换为 JSON/CSV 格式"、"提取关键信息"、"生成报告"

| 功能 | 说明 |
|------|------|
| 数据分析 | 对检索到的文档数据进行统计分析 |
| 格式转换 | JSON/CSV/Markdown 等格式相互转换 |
| 数据提取 | 从非结构化文本中提取结构化信息 |
| 报告生成 | 生成数据分析报告和摘要 |

**可用脚本**：`scripts/analyze.py`（数据分析）、`scripts/format_converter.py`（格式转换）、`scripts/extract_info.py`（信息提取）

```bash
# 数据分析
echo '{"items": [1, 2, 3, 4, 5]}' | python scripts/analyze.py

# 格式转换（JSON 转 CSV）
echo '[{"name": "A", "value": 1}]' | python scripts/format_converter.py --to csv
```

### 3. doc-coauthoring - 文档协作（源于 Claude 官方 Skill）

**用途**：引导用户完成结构化文档创作

**触发场景**：编写文档（"write a doc"、"draft a proposal"、"create a spec"），文档类型如 PRD、设计文档、决策文档、RFC

**三阶段工作流**：

| 阶段 | 目标 | 关键活动 |
|------|------|----------|
| Stage 1 | 缩小用户与 Claude 之间的信息差 | 元信息提问、上下文收集、澄清问题 |
| Stage 2 | 逐节构建文档 | 头脑风暴、筛选整理、迭代修改 |
| Stage 3 | 测试文档对读者的效果 | 预测读者问题、子代理测试、修复盲点 |

### 4. document-analyzer - 文档分析器

**用途**：深度分析文档结构和内容

**触发场景**：分析文档结构、提取关键信息、识别文档类型、进行内容质量评估

| 功能 | 说明 |
|------|------|
| 结构分析 | 识别文档的章节层级、组织架构 |
| 关键信息提取 | 提取核心论点、关键数据、重要结论 |
| 文档类型识别 | 判断文档类型（报告、手册、论文、合同等） |
| 内容质量评估 | 评估文档的完整性、一致性、可读性 |

**分析流程**：文档概览 -> 结构分析 -> 内容提取 -> 质量评估

### 5. summary-generator - 内容摘要生成

预加载技能位于 `skills/preloaded/` 目录下：

```
skills/preloaded/
├── citation-generator/
│   └── SKILL.md
├── data-processor/
│   ├── SKILL.md
│   └── scripts/
│       ├── analyze.py
│       ├── format_converter.py
│       └── extract_info.py
├── doc-coauthoring/
│   └── SKILL.md
├── document-analyzer/
│   └── SKILL.md
└── summary-generator/
    └── SKILL.md
```

## 创建自定义 Skill

暂时不支持用户自主创建自定义 Skill。

## 沙箱安全机制

### 脚本安全校验（Script Validator）

在脚本执行前，系统会进行多层安全校验，拦截潜在的恶意操作：

| 类型 | 说明 | 示例 |
|------|------|------|
| **危险命令检测** | 检测可能破坏系统的命令 | `rm -rf /`, `mkfs`, `shutdown`, fork bombs |
| **危险模式匹配** | 正则匹配高危操作模式 | `curl \| bash`, `base64 -d`, `eval()` |
| **网络访问检测** | 检测网络请求尝试 | `curl`, `wget`, `socket.connect`, `requests.get` |
| **反向 Shell 检测** | 检测远程控制后门 | `/dev/tcp/`, `bash -i`, `nc -e` |
| **参数注入检测** | 检测命令行参数中的注入 | `&&`, `\|`, `$()`, 反引号 |
| **Stdin 注入检测** | 检测标准输入中的嵌入命令 | 嵌入的命令替换语法 |

**拦截的危险命令**：

- **系统破坏类**：`rm -rf /`、`rm -rf /*`、`mkfs`、`dd if=/dev/zero`、Fork bombs (`:(){ :|:& };:`)
- **系统控制类**：`shutdown`、`reboot`、`halt`、`poweroff`、`killall`、`pkill`、`systemctl`、`service`
- **权限提升类**：`chmod 777 /`、`chown root`、`setuid`、`setgid`、`passwd`、访问 `/etc/passwd`、`/etc/shadow`、`/etc/sudoers`
- **凭证窃取类**：访问 `.ssh/`、`id_rsa`、`id_ed25519`、读取敏感配置文件
- **容器逃逸类**：`docker`、`kubectl`、`nsenter`、`unshare`、`capsh`

**拦截的危险模式**：

- **代码注入**：`curl ... | bash`、`wget ... | sh`、`eval()`、`exec()`、`os.system()`、`subprocess.Popen(shell=True)`
- **编码绕过**：`base64 -d`、`echo ... | base64 -d`、`xxd -r`
- **Python 特有风险**：`__import__()`、`pickle.load()`、`yaml.load()`、`yaml.unsafe_load()`

**Shell 操作符拦截**：`&&`、`||`、`;`、`|`、`$()`、反引号、`>`、`>>`、`<`、`2>`、`&>`、`\n`、`\r`

校验失败时返回详细的错误信息（类型 / 匹配模式 / 上下文 / 可读描述），例如：
```
security validation failed [dangerous_command]: Script contains dangerous command: rm -rf / (pattern: rm -rf /, context: ...cleanup && rm -rf / && echo done...)
```

### Docker 沙箱

Docker 模式提供最强的隔离：

- **非 root 用户**：容器内以普通用户运行
- **Capability 限制**：移除所有 Linux capabilities
- **只读文件系统**：根文件系统只读
- **资源限制**：内存 256MB，CPU 限制
- **网络隔离**：默认无网络访问
- **临时挂载**：Skill 目录只读挂载
- **脚本预校验**：执行前进行安全校验

系统使用专用沙箱镜像 `wechatopenai/weknora-sandbox`，预装 Python 3.11、Node.js 20、常用 CLI 工具和 Python 库。推荐首次部署时预拉取，避免首次执行脚本时等待下载：

```bash
# 方式一：直接拉取
docker pull wechatopenai/weknora-sandbox:latest

# 方式二：本地构建
sh scripts/build_images.sh -s
```

> 如果未预拉取，应用启动时会自动异步拉取镜像（`EnsureImage`），但首次执行可能需要等待下载完成。

### Local 沙箱

Local 模式提供基础保护：

- **命令白名单**：仅允许 `python`/`python3`、`node`/`nodejs`、`bash`/`sh`、`ruby`、`go run`
- **工作目录限制**：限定在 Skill 目录
- **环境变量过滤**：仅传递安全变量
- **超时控制**：默认 30 秒超时
- **路径遍历防护**：防止访问 Skill 目录外文件
- **脚本预校验**：执行前进行安全校验

## API 参考

### SkillManager

```go
type Manager interface {
    Initialize(ctx context.Context) error
    GetAllMetadata() []*SkillMetadata
    LoadSkill(ctx context.Context, skillName string) (*Skill, error)
    ReadSkillFile(ctx context.Context, skillName, filePath string) (string, error)
    ListSkillFiles(ctx context.Context, skillName string) ([]string, error)
    ExecuteScript(ctx context.Context, skillName, scriptPath string, args []string) (*sandbox.ExecuteResult, error)
    IsEnabled() bool
}
```

### Skill / ExecuteResult 结构

```go
type Skill struct {
    Name         string // 技能名称
    Description  string // 技能描述
    BasePath     string // 目录绝对路径
    FilePath     string // SKILL.md 绝对路径
    Instructions string // SKILL.md 主体指令内容
    Loaded       bool   // 是否已加载 Level 2
}

type ExecuteResult struct {
    ExitCode int           // 退出码
    Stdout   string        // 标准输出
    Stderr   string        // 标准错误
    Duration time.Duration // 执行时长
    Error    error         // 执行错误
}
```

## 完整工作流示例

```
用户: "帮我从 report.pdf 提取表格数据"

Agent 思考:
  -> 查看 System Prompt 中的 Skills 列表
  -> 发现 "pdf-processing" 技能匹配

Agent 行动 1: 调用 read_skill
  -> {"skill_name": "pdf-processing"}
  -> 获取 SKILL.md 指令内容
  -> 学习如何使用 pdfplumber

Agent 行动 2: 调用 execute_skill_script
  -> {"skill_name": "pdf-processing",
     "script_path": "scripts/extract_text.py",
     "args": ["report.pdf"]}
  -> 脚本在沙箱中执行，返回提取的表格数据

Agent 回复:
  -> 向用户展示提取的表格数据
  -> 提供数据使用建议
```

## 故障排查

### Skill 未被发现

1. 检查 `skill_dirs` 配置是否正确
2. 确认目录中存在 `SKILL.md` 文件
3. 验证 YAML frontmatter 格式

```bash
# 运行 demo 验证
go run ./cmd/skills-demo/main.go
```

### 脚本执行失败

1. 检查 `sandbox_mode` 配置
2. Docker 模式：确认 Docker 服务运行中
3. Local 模式：确认解释器已安装
4. 检查脚本权限和语法

### 元数据验证错误

- `skill name too long`：名称超过 50 字符
- `skill name contains invalid characters`：包含非法字符
- `skill name is reserved`：使用了保留词
- `skill description too long`：描述超过 500 字符

## 相关主题

- [MCP功能使用说明](MCP功能使用说明.md) - 另一种 Agent 扩展机制
- [IM集成开发](../集成扩展/IM集成开发.md) - Agent 可通过 IM 渠道使用技能
- [开发指南](../开发部署/开发指南.md) - 沙箱镜像的构建
- [内置智能体管理](内置智能体管理.md) - 内置智能体的 smart-reasoning 模式可使用技能

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
- [MCP功能使用说明](MCP功能使用说明.md) - 与 Skills 并列的 Agent 扩展机制
- [IM集成开发](../集成扩展/IM集成开发.md) - Agent 在 IM 渠道中可使用技能
- [版本路线图](../项目概述/版本路线图.md) - 路线图中的 Skills 社区扩展方向
