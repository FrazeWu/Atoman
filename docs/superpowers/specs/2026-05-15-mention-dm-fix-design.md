# @ 提及与私信链路修复设计

**日期：** 2026-05-15  
**状态：** 已确认设计，待实现  
**模块：** Forum mention / DM / Inbox

---

## 一、背景

当前实现中，`@` 提及与私信链路都存在实际可用性问题：

1. **@ 提及收不到通知**
   - 前端编辑器已有 `@` 自动补全，但选中候选后插入的是 Markdown 链接格式：`[@显示名](/user/username)`
   - 后端提及解析当前只识别纯文本 `@username`
   - 结果是：用户看起来完成了 @ 操作，但后端没有命中被提及用户，也不会创建通知

2. **私信提示 `user not found`**
   - 私信当前的正确使用链路应是：从用户主页点击“发私信” → 跳转 `/inbox?tab=dm&user=:username` → 打开会话并发送消息
   - 本次不扩展“全局搜人发私信”，只要求这条 username 深链稳定可用

本次修复目标是：**统一 username 为唯一目标标识，修复 @ 通知与私信发送链路，并让前端交互与后端解析保持一致。**

---

## 二、行为定义

### 2.1 @ 提及

- 当用户在编辑器中输入 `@` 时，前端显示候选下拉框
- 候选来源限定为：**关注我的用户（followers）**
- 下拉展示可以包含显示名、头像等辅助信息，但最终插入正文时统一写入：

```text
@username
```

- 用户不必从下拉中选择；如果最终正文中存在有效的 `@username`，后端也应识别并发通知
- 历史内容中已经存在的 Markdown 链接 mention：

```markdown
[@显示名](/user/username)
```

  也必须继续兼容解析

### 2.2 私信

- 私信目标统一只用 `username` 标识
- 用户从某个用户主页点击“发私信”时，跳转：

```text
/inbox?tab=dm&user=:username
```

- Inbox 读取 query 中的 `user`，并打开/创建与该 username 对应的会话
- 用户发送消息时，后端按 username 定位目标用户
- 本次不提供额外的“搜人发私信”入口

---

## 三、范围

### 本次纳入

1. 修复 `@` 自动补全候选来源为 followers
2. 修复 `@` 选中插入格式，统一插入 `@username`
3. 后端 mention 解析兼容：
   - `@username`
   - `[@显示名](/user/username)`
4. 修复用户主页发私信链路，确保 `/inbox?tab=dm&user=:username` 能正确打开目标会话并发送
5. 修复 DM 后端按 username 查人时的稳健性问题
6. 保证消息发送后，对方能收到 DM 推送或未读变化

### 本次不纳入

1. 私信全局搜人入口
2. `@` 候选范围配置化
3. mention 交互扩展到其他非当前编辑器场景
4. 已读回执、撤回、屏蔽、搜索等 DM 二期能力

---

## 四、推荐方案

采用 **username 单一真源 + 前后端兼容层**。

### 核心原则

- 前端 `@` 与 DM 深链统一只传 `username`
- 后端解析统一只以 `username` 为命中目标
- 旧正文格式通过后端兼容层承接，不继续作为新的输出格式

### 选择理由

该方案同时满足：
- 用户体验一致
- 新旧内容兼容
- 风险低于仅靠前端或仅靠后端兜底
- 后续维护成本最低

---

## 五、结构设计

### 5.1 前端

#### `web/src/components/shared/AEditor.vue`

职责调整：
- mention 候选接口改为 followers 范围
- 用户选中候选时，插入 `@username`
- 不再插入 Markdown 链接格式 mention

#### `web/src/views/blog/ProfileView.vue`

职责保持不变：
- “发私信”按钮继续跳到 `/inbox?tab=dm&user=:username`
- 这里的 `username` 是私信唯一目标标识

#### `web/src/views/feed/InboxPage.vue`

职责调整：
- 读取 query 中的 `user`
- 将其作为唯一会话目标传给 `dmStore.openConversation(username)`
- 如果该 username 无效，页面应显示明确错误，而不是静默失败

#### `web/src/stores/dm.ts`

职责保持并收紧：
- `openConversation(username)`
- `sendMessage(username, content, imageUrl?)`
- `markRead(username)`

这些 action 都以 `username` 作为唯一入参，不引入其他目标标识

### 5.2 后端

#### `server/internal/handlers/user_handler.go`

新增或调整用户搜索接口能力：
- 现有搜索接口当前是通用用户搜索
- 本次需要支持 **followers 范围** 的候选查询，供 mention 下拉使用

推荐行为：
- 在保留通用搜索能力的前提下，为当前登录用户增加 followers 过滤逻辑
- 候选结果字段仍返回：`uuid / username / display_name / avatar_url`

#### `server/internal/service/forum_mention_parser.go`

职责调整：
- 统一提取正文中的 username
- 同时兼容两种格式：
  1. `@username`
  2. `[@显示名](/user/username)`

解析输出应为一组去重后的 username，再统一查询用户并创建通知

#### `server/internal/handlers/dm_handler.go`

职责调整：
- 所有 username 入参在查询前先做稳健处理：例如去前后空格
- 用户查找统一以 username 为准
- 当 username 无效时，返回明确错误
- 从 `/inbox?tab=dm&user=:username` 进入时，必须能稳定命中目标用户并打开/创建会话

---

## 六、数据流

### 6.1 @ 提及

1. 用户在编辑器输入 `@`
2. 前端请求 followers 范围候选
3. 用户可从下拉中选择，插入 `@username`
4. 用户也可以纯手打 `@username`
5. 提交内容到后端
6. 后端统一解析正文中的 username（兼容新旧格式）
7. 若命中有效用户，为其创建 `forum_mention` 通知
8. 如果对方在线，通过 `/ws/user` 推送；否则通过未读数与收件箱查看

### 6.2 私信

1. 用户在用户主页点击“发私信”
2. 前端跳转 `/inbox?tab=dm&user=:username`
3. Inbox 读取 `user` query
4. `dmStore.openConversation(username)` 调用后端
5. 后端按 username 找到目标用户，打开或创建会话
6. 用户发送消息
7. 后端落库并向接收方推送 `dm` 事件
8. 接收方看到未读变化或实时消息

---

## 七、失败处理

### @ 提及

- 候选为空：不显示下拉，不阻止继续输入
- 手打了不存在的 `@username`：不报错，但不发送 mention 通知
- 历史 Markdown mention：继续兼容，不要求用户手动修改旧内容

### 私信

- query 中 username 无效：Inbox 显示明确错误
- 发送失败：前端保留输入内容并展示后端错误
- 权限拒绝：继续使用现有错误码与错误信息

---

## 八、验证标准

1. 输入 `@` 时，只出现 followers 候选
2. 选择候选后，正文插入的是 `@username`
3. 纯手打 `@username` 也能触发 mention 通知
4. 历史 `[@显示名](/user/username)` 内容仍能触发 mention 通知
5. 从用户主页点击“发私信”后，能够稳定打开目标会话
6. 发送私信后，对方能收到 DM 推送或至少看到未读数变化
7. 无效 username 能返回明确失败，而不是静默错误或错误命中其他用户

---

## 九、实施注意点

1. 不要继续生成 Markdown 链接 mention 作为新格式
2. 后端兼容旧格式仅作为过渡能力，新的输入与存储语义统一按 username 处理
3. 本次不要引入新的全局用户搜索式私信入口，避免扩大范围
4. 前端与后端对 username 的处理必须保持一致，否则会继续出现“前端看似成功、后端未命中”的问题
