---
title: Embed组件集成
tags: [集成扩展, Embed, 嵌入, Widget, 网站集成]
aliases: [embed-integration, Embed集成, 嵌入式组件]
---

# 网页嵌入集成指南

> **一句话**:把 WeKnora 智能体嵌进任意第三方网页,支持三种形态--**固定 iframe**、**悬浮浮窗 Widget**、**安全模式 Widget**。本文一次性讲清楚选型、接入、SDK 用法、上下文注入、限制与排错。

本篇是总览入口。安全模式细节见 [Embed安全模式](Embed安全模式.md),独立子域部署见 [Embed子域部署](Embed子域部署.md)。

---

## 1. 三种形态怎么选

| 形态 | 代码来源 | UI 外观 | 发布 Token 在哪 | 有 SDK API | 传上下文 |
|------|----------|---------|-----------------|-----------|---------|
| **iframe** | 面板「iframe」Tab | 页面内固定聊天框 | URL hash 明文 | ✗ | 需自己 postMessage |
| **浮窗 Widget** | 面板「Widget」Tab | 右下角气泡,点击展开 | `data-token` 明文 | ✓ | `WeKnora.setContext()` |
| **安全模式 Widget** | 面板「安全模式」Tab | 同上,右下角气泡 | **不进浏览器**(服务端换短期 token) | ✓ | `WeKnora.setContext()` |

**选型建议**:

- 快速验证 / 内部使用 -> **浮窗 Widget**
- 生产对外嵌入 -> **安全模式 Widget**(密钥不落地)
- 必须是页面内固定框 + 不需上下文 -> **iframe**
- 必须是固定框 + 需要上下文 -> 浮窗 Widget 用 CSS 改造,或 iframe 自写 postMessage(见 §6)

---

## 2. 前置准备

所有形态都要先在 WeKnora 后台「智能体 -> 嵌入渠道」完成:

1. **创建渠道**,绑定一个智能体(agent)。记下:
   - **渠道 ID**(`channel_id`)--URL 路径用
   - **发布 Token**(`em_` 开头)--管理端「渠道密钥」查看,仅服务端持有(安全模式)或仅测试用
2. **域名白名单**(`allowed_origins`)--见 §3,最常见踩坑点
3. **限流**:每 IP 每分钟、渠道每分钟、渠道每天三档
4. **页面标题**(`page_title`)--浮窗 / iframe 顶部显示的标题(见 §8)
5. 其他可选:欢迎语、主题色、推荐问题、联网搜索开关、文件上传开关、默认语言

---

## 3. 域名白名单(必读)

后端 `internal/middleware/embed_auth.go` 对**每一个** `/api/v1/embed/...` 请求校验 `Origin` 头,不在白名单内返回 `403 origin not allowed`。

**匹配规则**(`originAllowed`):

- 空白名单 = **全拒**(不是全放)
- 精确匹配(大小写不敏感)
- `*` = 全放(仅开发环境,生产禁止,见 `embed_channel.go` `validateAllowedOrigins`)
- `*.example.com` = 后缀匹配(注意:只支持 `*.` 开头,**不支持** `https://*.example.com` 这种 scheme+通配混写)

**Origin 格式严格为 `scheme://host[:port]`,无路径、无尾斜杠**:

| 填法 | 是否生效 |
|------|---------|
| `http://localhost:5173` | ✅ |
| `https://app.example.com` | ✅ |
| `*.example.com` | ✅(匹配所有子域) |
| `localhost:5173` | ❌ 缺 scheme |
| `http://localhost:5173/` | ❌ 多尾斜杠 |
| `http://localhost:5173/embed-test.html` | ❌ 带路径 |
| `https://*.example.com` | ❌ scheme+通配不支持 |

**到底填什么 origin**:浏览器 F12 -> Network -> 找任意 `/api/v1/embed/...` 请求 -> 看请求头 `Origin`,原样填进白名单。或看后端日志 `[embed_auth] origin "xxx" not allowed for channel <id>`,日志里的值就是后端收到的 origin。

详细见 [Embed安全模式](Embed安全模式.md) 与 [Embed子域部署](Embed子域部署.md)。

---

## 4. iframe 嵌入

### 4.1 声明式(最简,无上下文)

面板「iframe」Tab 复制的代码:

```html
<iframe
  src="https://<weknora-host>/embed/<渠道ID>#token=<em_发布Token>"
  style="width:400px;height:600px;border:none;border-radius:12px"
  allow="clipboard-write">
</iframe>
```

标题自动读渠道 `page_title`。**此方式无法传上下文**(没有 postMessage 通道)。

### 4.2 编程式(要传上下文时)

iframe 的 `src` 不带 token,通过 postMessage 把 token + 上下文喂进去。消息协议就是浮窗 SDK 内部用的那套:

```html
<iframe id="wk"
  src="https://<weknora-host>/embed/<渠道ID>"
  style="width:400px;height:600px;border:none;border-radius:12px"
  allow="clipboard-write"></iframe>

<script>
  const iframe = document.getElementById('wk');
  const EMBED_ORIGIN = 'https://<weknora-host>';   // postMessage targetOrigin,必须精确,禁止 '*'
  const PUBLISH_TOKEN = 'em_发布Token';            // 测试用;生产改走安全模式
  const CHANNEL_ID = '<渠道ID>';

  window.addEventListener('message', (e) => {
    const d = e.data;
    if (e.source !== iframe.contentWindow) return;
    if (!d || d.source !== 'weknora-embed') return;

    // embed 页就绪后主动要 token
    if (d.type === 'bootstrap_request') {
      iframe.contentWindow.postMessage({
        source: 'weknora-host',
        type: 'provide_token',
        token: PUBLISH_TOKEN,
        channel_id: CHANNEL_ID,
      }, EMBED_ORIGIN);
    }

    // embed 页 ready 后才能注入上下文(早发会丢)
    if (d.type === 'ready') {
      iframe.contentWindow.postMessage({
        source: 'weknora-host',
        type: 'set_context',
        payload: { 问题描述: '用户在结算页报错 ERR_500' },
      }, EMBED_ORIGIN);
    }
  });

  // 可选:直接发一条问题
  function ask(q) {
    iframe.contentWindow.postMessage({
      source: 'weknora-host',
      type: 'open_with_query',
      payload: { query: q },
    }, EMBED_ORIGIN);
  }
</script>
```

**消息协议**:

| 方向 | type | 作用 |
|------|------|------|
| embed -> 宿主 | `bootstrap_request` | 要 token |
| 宿主 -> embed | `provide_token` | 喂发布/会话 token |
| embed -> 宿主 | `ready` | iframe 就绪,可 setContext |
| 宿主 -> embed | `set_context` | 注入上下文(merge 语义) |
| 宿主 -> embed | `open_with_query` | 弹窗 + 发一条问题 |
| 宿主 -> embed | `set_locale` | 切语言(`zh-CN`/`en-US`/`ko-KR`/`ru-RU`) |
| embed -> 宿主 | `message_sent` | 用户发送了消息 |
| embed -> 宿主 | `message_received` | 智能体回复完成 |

所有消息都带 `source` 字段(`weknora-host` 或 `weknora-embed`),embed 页会校验来源 origin(`isTrustedParentMessage`),不可信来源直接忽略。

---

## 5. 浮窗 Widget

### 5.1 声明式自动初始化(出气泡,无 API 句柄)

```html
<script src="https://<weknora-host>/weknora-widget.js"
        data-channel="<渠道ID>"
        data-token="<em_发布Token>"
        data-position="bottom-right"
        data-primary-color="#07C05F"
        data-title="AI 助手"></script>
```

加载即出气泡。但**这种方式拿不到 `WeKnora` 句柄,无法 `setContext`**。要传上下文必须用 5.2。

### 5.2 编程式初始化(推荐,有完整 SDK)

```html
<!-- 不带 data-* 属性,否则会自动 init -->
<script src="https://<weknora-host>/weknora-widget.js"></script>
<script>
  const w = WeKnora.init({
    channel: '<渠道ID>',
    token: 'em_发布Token',          // 上线改 tokenEndpoint,见 §6
    baseUrl: 'https://<weknora-host>',
    position: 'bottom-right',
    primaryColor: '#07C05F',
    title: 'AI 助手',
  });

  // setContext 必须在 ready 之后(早发会丢)
  WeKnora.on('ready', () => {
    WeKnora.setContext({ 问题描述: '用户在结算页报错 ERR_500' });
  });

  // 可选:弹窗 + 直接发问题(openWithQuery 内部会等 ready,无需自己卡时机)
  // WeKnora.openWithQuery('根据上下文,我遇到了什么问题?');
</script>
```

---

## 6. 安全模式 Widget(生产推荐)

发布 Token `em_` 永不落地浏览器。页面只放 `data-token-endpoint`,指向你自己的后端;后端用 `em_` 调 `/exchange` 换 30 分钟有效的会话 Token `ems_`,返回给前端。Widget 在过期前自动刷新。

```html
<script src="https://<weknora-host>/weknora-widget.js"
        data-channel="<渠道ID>"
        data-token-endpoint="https://你的后端/weknora/embed-token"
        data-position="bottom-right"></script>
<script>
  WeKnora.init({
    channel: '<渠道ID>',
    tokenEndpoint: 'https://你的后端/weknora/embed-token',  // 注意是 tokenEndpoint,不是 token
    baseUrl: 'https://<weknora-host>',
  });
  WeKnora.on('ready', () => WeKnora.setContext({ 问题描述: 'ERR_500' }));
</script>
```

取令牌接口与服务端示例(Node/Go)、exchange 约定、白名单填法、上线检查、排错表,全部见 [Embed安全模式](Embed安全模式.md)。这里只强调三点:

1. **exchange 只接受发布 Token**:`Authorization: Embed em_…`,用 `ems_` 换会被拒(`publish token required`)。
2. **exchange 请求要手动带 `Origin` 头**,且该 Origin 必须在渠道白名单内,否则 `403 origin not allowed`。
3. **iframe 形态无法走安全模式**:iframe 必须把 token 喂进去,而安全模式的短期 token 又是 Widget 内部自动刷新的。要"固定框 + 隐藏密钥",只能用浮窗 Widget + CSS 改造(见 §9)。

---

## 7. WeKnora SDK 完整 API

`weknora-widget.js` 挂载的全局对象。所有方法都依赖先 `init`。

| 方法 | 作用 |
|------|------|
| `WeKnora.init(opts)` | 初始化,返回实例。`opts`: `channel` / `token` 或 `tokenEndpoint` / `baseUrl` / `position` / `primaryColor` / `title` / `width` / `height` |
| `WeKnora.setContext(obj)` | **注入上下文**,merge 语义,每轮对话带在 query 前 |
| `WeKnora.openWithQuery(str)` | 弹窗 + 发一条问题(内部等 ready) |
| `WeKnora.setLocale(str)` | 切语言 `zh-CN` / `en-US` / `ko-KR` / `ru-RU` |
| `WeKnora.open()` / `close()` / `toggle()` | 控制面板开关 |
| `WeKnora.on(event, fn)` | 监听:`ready` / `open` / `close` / `message_sent` / `message_received` |
| `WeKnora.off(event, fn)` | 取消监听 |
| `WeKnora.destroy()` | 卸载浮窗 |

实例方法(编程式 init 返回值)与上述同名,也可直接用返回值调用。

---

## 8. 标题控制

浮窗 / iframe **顶部 header 显示的标题来自后端渠道配置**,不是 SDK 的 `title` 参数。

**优先级**(`EmbedPage.vue` `channelDisplayTitle`):

```
page_title  >  渠道 name  >  智能体 name  >  "AI Assistant"
```

改标题:后台编辑渠道 -> 改「页面标题」-> 保存刷新即可,无需重启。

**`header_title_mode`** 字段控制标题行为:

- `channel`(默认):始终显示渠道标题
- `session`:首条消息后,标题切换成**会话自动生成的标题**(由首条用户消息触发后端生成),渠道标题降级为副标题

SDK 的 `title` 参数只影响 launcher 按钮的 `aria-label` 和 iframe 的 `title` 属性(无障碍),建议设成与渠道标题一致。

---

## 9. 限制与注意事项

### 9.1 上下文注入的机制与限制

`setContext` 注入的上下文**不是独立的 system/context 通道**,而是由前端 `buildQueryWithHostContext`(`frontend/src/utils/embedContext.ts`)拼到用户 query 文本前:

```
[Host context]
问题描述: 用户在结算页报错 ERR_500

{用户实际输入}
```

这个拼接后的字符串作为 `query` 字段发给后端,后端 `patchEmbedChatPayload` **不修改 query**,原样进入 RAG 检索与 LLM。

**由此带来的限制**:

1. **每轮对话都带**:`hostContext` 是响应式全局 merge,不是"仅首条带"。想只带一次,需 setContext 后再置空。
2. **merge 语义**:`setContext({a:1})` 后 `setContext({b:2})` 得到 `{a:1,b:2}`。清空某键:`setContext({a:''})`(空值会被 `buildQueryWithHostContext` 过滤,不再带入)。
3. **影响检索**:上下文拼进 query,会改变 RAG 向量检索的召回结果。上下文过长会污染召回。
4. **进入历史**:拼接后的内容作为 user message 存进会话历史,后续轮次可见。
5. **非结构化**:键值对字符串拼接,模型按 user 输入理解,无独立 role。想要"纯上下文不进检索"的语义,当前不支持。
6. **setContext 必须在 ready 之后**:SDK 的 `setContext` 没有等 iframe ready 的保护(对比 `openWithQuery` 有 `whenIframeReady`),早发会静默丢失。务必放进 `on('ready', ...)`。

### 9.2 iframe 与 SDK 的关系

**SDK 内部本身就是 iframe**(`weknora-widget.js` 的 `init` 即 `createElement('iframe')` 指向 `/embed/:channelId`)。所以:

- `WeKnora.setContext()` 管的是 SDK 自己 `init` 出来的气泡 iframe,**不认识**你在 HTML 里写的 `<iframe>`。
- SDK 没有提供"把 API 套到外部 iframe"的入口。
- **iframe 形态无法走安全模式**(见 §6.3)。

想要"固定框 + SDK API + 安全模式",唯一稳的方案:浮窗 Widget + CSS 把固定浮层拍进容器、隐藏 launcher(见 9.3)。或 iframe 自写 postMessage(§4.2,但不能安全模式)。

### 9.3 浮窗伪装成固定框(hack,慎用)

```html
<div id="wk-box" style="width:400px;height:600px;position:relative"></div>
<script src="https://<weknora-host>/weknora-widget.js"></script>
<script>
  WeKnora.init({ channel:'<渠道ID>', tokenEndpoint:'…', baseUrl:'…' });
  WeKnora.on('ready', () => { WeKnora.setContext({问题描述:'…'}); WeKnora.open(); });
</script>
<style>
  button[aria-label] { display: none !important; }            /* 隐藏气泡按钮 */
  body > div[style*="z-index:2147482999"] {                   /* 把浮层拍进容器 */
    position: absolute !important; inset: 0 !important;
    width: 100% !important; height: 100% !important; display: block !important;
  }
</style>
```

⚠️ SDK 的 DOM 结构与内联样式无稳定契约,版本升级可能失效。能用但不优雅,非必要不选。

### 9.4 安全限制

- **发布 Token `em_` 是长期密钥**:明文写进页面/URL hash = 等于公开。生产必须走安全模式。
- **生产禁止 `*` 白名单**:`validateAllowedOrigins` 在生产模式拒绝通配。
- **会话 Token `ems_` 约 30 分钟过期**:Widget 自动刷新;iframe 自写需自行处理过期换发。
- **postMessage 的 `targetOrigin` 必须精确**:禁止 `'*'`,否则 `provide_token` 会泄露密钥。embed 页 `isTrustedParentMessage` 会校验来源。
- **限流**:每 IP 每分钟 + 渠道每分钟 + 渠道每天三重;超额 `429`。
- **跨域自动 sandbox**:宿主与 embed 源站不同域时,Widget 自动给 iframe 加 `sandbox`(见 [Embed子域部署](Embed子域部署.md))。

### 9.5 浮窗行为

- 浮窗默认收起,但 iframe 一直在 DOM 加载,`ready` 照常触发。不 `open` 也能 `setContext`,只是用户看不见。
- 推荐问题、欢迎语在用户首条消息前显示,发送后隐藏。

---

## 10. 故障排查

| 现象 | 原因与处理 |
|------|------------|
| **403 `origin not allowed`** | 白名单未含当前 `Origin`。F12 看请求头 `Origin` 原样填入;注意无路径/无尾斜杠/有 scheme(见 §3) |
| **浮窗一直「等待 Token」** | 声明式缺 `data-token`/`data-token-endpoint`;安全模式接口未返回 `{token, expiresIn}` 或 CORS 未放行 |
| **`setContext` 没生效** | 在 `ready` 之前调用了。放进 `on('ready', ...)` |
| **exchange 401 `publish token required`** | 用了 `ems_` 会话 Token 换 token;exchange 只接受 `em_` 发布 Token |
| **exchange 403 `origin not allowed`** | 服务端 exchange 未手动带 `Origin` 头,或带的 Origin 不在白名单 |
| **标题改了没变** | 改的是 SDK `title` 参数而非渠道 `page_title`;或 `header_title_mode=session` 切成了会话标题 |
| **上下文每轮都带 / 想清空** | merge 语义;`setContext({字段:''})` 清空单键(空值被过滤) |
| **iframe 传上下文失败** | 声明式 iframe 无 postMessage 通道,改用 §4.2 编程式 |
| **生产用 `*` 白名单报错** | 生产模式禁止通配,填具体 origin 或 `*.example.com` |
| **429 `rate limit exceeded`** | 触发限流,调高渠道分钟/日限额 |

---

## 11. 本地调试

仓库内置 `frontend/public/embed-test.html`(浮窗上下文注入测试页)。用法:

1. `make dev-start` 起服务(前端 5173)
2. 浏览器开 `http://localhost:5173/embed-test.html`(**勿用 `file://`**,origin 为 null 会被拒)
3. 填渠道 ID / 发布 Token / baseUrl
4. 渠道白名单加 `http://localhost:5173`
5. 顺序点:初始化浮窗 -> 设置上下文 -> 打开浮窗 -> 提问

页面有黑底日志区实时打 `ready` / `message_sent` / `message_received`。

---

## 12. 源码参考

- Widget SDK 源码:`frontend/public/weknora-widget.js`
- 代码生成器:`frontend/src/api/embed/index.ts`
- embed 前端入口:`frontend/src/embed-main.ts`、`frontend/src/views/embed/`
- 后端 handler:`internal/handler/embed_channel.go`
- origin 校验:`internal/middleware/embed_auth.go`

---

## 相关主题

- [Embed安全模式](Embed安全模式.md) - 安全模式 Widget 详解:exchange 约定、服务端示例、上线检查
- [Embed子域部署](Embed子域部署.md) - 独立子域部署、跨域 sandbox、域名白名单填法
- [IM集成开发](IM集成开发.md) - 另一类对外接入渠道(IM 平台),共享渠道/会话管理底层

---

## 反向链接

- [Home](../Home.md) - Wiki 首页导航
