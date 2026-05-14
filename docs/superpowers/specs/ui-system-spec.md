# UI System Spec

## Purpose
UI System 的目标是为整个网站建立统一的基础视觉组件系统，减少按钮、输入框、下拉框、弹窗等交互元素的风格漂移，并让新旧页面逐步收口到 Archive / Brutalist 风格语言。

## Visual Language
- 整体视觉：黑白极简、档案馆风格、强结构感。
- 组件气质：高对比、硬核、克制。
- 圆角统一为 `0`。
- 边框统一为 `2px solid #000`。
- 主色为黑 / 白，危险态允许红色，成功/正常态允许绿色。
- 悬浮对象使用硬阴影。
- 输入类组件默认扁平化，focus 不浮起。

## Base Component Set
- `AButton`
- `AInput`
- `ATextarea`
- `ASelect`
- `AModal`
- `ADropdown`
- `APopover`

## Component Rules

### Button
- 统一 variants：`primary | secondary | danger | ghost`
- 统一 sizes：`sm | md | lg`
- 强字重、uppercase、拉开字距

### Input / Textarea
- 白底黑字、2px 黑边、0 圆角
- label 固定在控件上方
- error 态为红边 + 红色说明文字
- textarea 不允许 resize

### Select
- trigger 与输入框同体系
- dropdown panel 使用白底黑边 + 硬阴影
- option hover 黑白反转
- selected 使用左侧标记，不做整行永久反黑

### Modal / Dropdown / Popover
- 白底黑边、0 圆角、硬阴影
- modal 有独立标题区与 footer 区
- 不引入毛玻璃或柔和 SaaS 风格

## Token Rules
设计 tokens 统一落在 `web/src/style.css` 的全局 CSS variables，覆盖：
- 颜色
- 边框
- 阴影
- 间距
- 字号
- 字重
- 字距
- z-index

## Migration Rules
- 新页面优先直接使用基础组件。
- 历史页面逐步替换原生 `<select>`、手写 modal 壳层与高频 inline layout style。
- “UI 结构色”必须优先走 token；“内容语义色”可保留少量特例。

## Durable Decisions
- 整个系统不追求多风格兼容，而是统一收口到 Archive / Brutalist 语言。
- ghost 按钮也保留边框，不退化成裸文本链接。
- 模态与弹层都应复用同一视觉外观，而不是分裂成两套语言。
- 页面迁移应优先清理原生控件、重复 modal 壳层与高频 inline 布局样式。

## Remaining Convergence Scope
- 剩余原生 `<select>` 替换
- modal 壳层统一
- 高频 inline layout style 抽离
- 旧 `outline` / 红色按钮写法进一步收口到 `variant`
- badge / chip 统一到 token 体系
