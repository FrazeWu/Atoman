# Editor SV 统一 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 确保 AEditor 编辑器在所有内容输入场景按规格使用正确的 mode，并清理任何 Tiptap/WYSIWYG 残留依赖。

**Architecture:**
- `AEditor.vue` 已支持 `mode: 'sv' | 'plain'`（Tiptap 已移除，无 wysiwyg 模式）
- 规格规定：博客/视频/播客内容 → `sv`；Forum 发帖 → `sv`；**Forum 回复 → `plain`**；Debate 论点 → `plain`
- 当前已合规：Blog PostEditorView `sv` ✅，Forum NewTopicView `sv` ✅，Debate arguments `plain` ✅
- **当前不合规：** Forum TopicView 回复区用 `mode="sv"` ❌（规格要求 `plain`）

**已完整的部分（无需修改）：**
- `AEditor.vue`：支持 sv/plain 两种 mode，Tiptap 已清理
- `PostEditorView.vue`：`mode="sv"` ✅
- `ForumNewTopicView.vue`：`mode="sv"` ✅
- `DebateTopicView.vue`：`mode="plain"` ✅

---

## 文件清单

### 修改
| 文件 | 改动 |
|------|------|
| `web/src/views/forum/ForumTopicView.vue` | 将回复区 AEditor 的 `mode="sv"` 改为 `mode="plain"` |

---

## Task 1：审计 — 确认无 Tiptap 残留

**Files:** 只读，无修改

- [ ] **Step 1: 全局搜索 Tiptap 残留**

```bash
grep -rn "tiptap\|@tiptap\|BubbleMenu\|StarterKit\|NodeViewWrapper\|EditorContent" web/src/ 2>/dev/null | grep -v ".bak" | head -20
```

期望：无输出。若有残留则在 Task 1 末尾额外步骤中删除对应 import 和代码。

- [ ] **Step 2: 确认 AEditor mode 枚举**

```bash
grep -n "mode.*plain\|mode.*sv\|mode.*wysiwyg" web/src/components/shared/AEditor.vue | head -10
```

期望：只有 `'sv' | 'plain'`，无 `'wysiwyg'`。

- [ ] **Step 3: 全局扫描所有 AEditor 使用点的 mode 值**

```bash
grep -rn "mode=\"\|:mode=\"\|mode='" web/src/views/ web/src/components/ 2>/dev/null | grep -i "editor"
```

期望：mode 只有 `sv` 或 `plain`。

- [ ] **Step 4: Commit（若无 Tiptap 残留则只记录审计结论）**

```bash
git commit --allow-empty -m "chore(editor): audit confirms no Tiptap remnants, SV mode applied to all required scenes"
```

---

## Task 2：修复 — Forum 回复改为 plain 模式

**Files:**
- Modify: `web/src/views/forum/ForumTopicView.vue`

**背景：** 规格要求 Forum 回复（ForumReply）使用 `plain` 模式（纯文本 + 图片），当前代码使用 `mode="sv"`。

- [ ] **Step 1: 找到回复区的 AEditor 调用**

```bash
grep -n "AEditor\|mode=" web/src/views/forum/ForumTopicView.vue
```

期望：找到约第 160 行的 `<AEditor ... mode="sv" ...>`（回复输入框）。

- [ ] **Step 2: 将 mode="sv" 改为 mode="plain"**

找到回复输入框区域：

```vue
<!-- AEditor for reply -->
<AEditor
  ...
  mode="sv"
  ...
/>
```

改为：

```vue
<!-- AEditor for reply - plain mode per spec: reply uses plain text + image -->
<AEditor
  ...
  mode="plain"
  ...
/>
```

- [ ] **Step 3: 确认图片上传仍可用**

`AEditor` 在 `plain` 模式下应保留图片上传按钮。若无，检查 `AEditor.vue` 中 `mode === 'plain'` 分支是否有图片插入支持：

```bash
grep -n "image\|图片\|upload" web/src/components/shared/AEditor.vue | head -10
```

若 plain 模式不支持图片插入，在 `AEditor.vue` 的 plain 分支下方添加一个简单的图片上传 `<input type="file">` 按钮，点击后调用现有的图片上传逻辑（参照 sv 模式中的上传 handler）。

- [ ] **Step 4: 类型检查**

```bash
cd web && bun run type-check 2>&1 | tail -5
```

期望：无新增错误。

- [ ] **Step 5: Commit**

```bash
git add web/src/views/forum/ForumTopicView.vue
git commit -m "fix(editor): change Forum reply editor from sv to plain mode per spec"
```

---

## Task 3：验收

- [ ] **Step 1: 前端构建**

```bash
cd web && bun run type-check && bun run build
```

期望：0 type errors，build 成功。

- [ ] **Step 2: 手动冒烟测试**

1. `/blog/new` → 确认 PostEditorView 显示分栏 SV 模式（左 Markdown + 右预览）
2. `/forum/new-topic` → 确认 ForumNewTopicView 显示 SV 模式
3. `/forum/topic/:id` → 回复框为 plain 文本区（无预览栏），有图片插入按钮
4. `/debate/:id` → 论点输入框为 plain 文本区
5. 以上所有场景无 Tiptap 相关 JS 错误

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "feat(editor): complete Editor SV unification - all scenes use correct mode"
```
