# 任务计划：统一风格组件系统

## 目标
为网站建立一套统一的基础视觉组件系统，减少下拉框、弹窗、按钮、输入框等交互元素的风格漂移，保证新旧页面逐步收口到同一套 Archive / Brutalist 风格语言。

## 当前阶段
已完成第一轮规则确认，进入可实施的系统规划阶段。

## 已确认设计方向
- 整体视觉：黑白极简、档案馆风格、强结构感
- 组件气质：硬核、克制、高对比，不走现代柔和 SaaS 风格
- 圆角：统一 `0`
- 边框：统一 `2px solid #000`
- 主色：黑 / 白，危险态可用红色，成功/正常态可用绿色
- 阴影：仅用于按钮、弹层、模态等需要“悬浮感”的对象，且使用硬阴影
- 输入类组件：扁平化，不悬浮，focus 不加阴影

## 第一批组件范围
- [x] AButton
- [x] AInput
- [x] ATextarea
- [x] ASelect
- [x] AModal
- [x] ADropdown
- [x] APopover

## 已确认的组件规则

### AButton
- variants：`primary | secondary | danger | ghost`
- sizes：`sm | md | lg`
- props：`variant`、`size`、`disabled`、`loading`、`block`
- 统一：0 圆角、2px 边框、粗字重、uppercase、字距拉开
- primary：黑底白字，默认悬浮，hover 只加下划线，active 下沉
- secondary：白底黑字，默认悬浮，hover 只加下划线，active 下沉
- danger：白底红字红边，默认悬浮，hover 只加下划线，active 下沉
- ghost：轻量按钮，但保留边框，不退化为裸文字链接
- 第一版不引入 `outline`、`plain`、`iconOnly`、`elevated` 等扩展 props，避免系统重新发散

### AInput / ATextarea
- 白底黑字、2px 黑边、0 圆角、扁平
- 默认无阴影，focus 也不浮起
- focus 仅通过边框细节变化表达状态
- label 固定在输入框上方，不以 placeholder 替代 label
- error 态使用红边 + 红色说明文字
- textarea 不允许 resize
- AInput props：`modelValue`、`label`、`placeholder`、`disabled`、`error`、`hint`
- ATextarea props：`modelValue`、`label`、`placeholder`、`disabled`、`error`、`hint`、`rows`

### ASelect
- trigger 与 AInput 同体系：扁平、白底黑字、2px 黑边、0 圆角、无阴影
- dropdown panel 使用白底黑边 + 硬阴影
- option hover 黑白反转
- selected 采用左侧标记方案，不使用整行永久反黑
- props：`modelValue`、`options`、`placeholder`、`disabled`、`label`、`error`、`hint`
- trigger 右侧只使用简单展开符号（如 `▾`），不引入花哨图标语言

### AModal
- 遮罩：黑色半透明，不使用毛玻璃
- 本体：白底黑边、0 圆角、强硬阴影
- 标题区：独立标题栏 + 分隔线
- footer：固定存在，统一承载按钮区
- props：`open`、`title`、`size`、`closable`
- 通过 slots 承载正文与 footer 操作区

### ADropdown / APopover
- 共用外观：白底黑字、2px 黑边、0 圆角、硬阴影
- ADropdown 用于紧凑型菜单列表
- APopover 保持同风格，只是内容更自由，不额外派生一套视觉语言
- ADropdown API：`trigger` + 默认插槽菜单项，支持点击外部关闭
- APopover API：`trigger` + 默认插槽内容区，支持点击外部关闭

## 实施阶段规划

### 阶段 1：定义 design tokens
- [x] 整理颜色、边框、阴影、间距、字体、字号、字距、层级等 tokens
- [x] 决定 tokens 落点：采用 `web/src/style.css` 中的全局 CSS variables
- [x] 明确刚性统一项：圆角、边框宽度、阴影风格、主色基调、字距风格
- [x] 明确语义可变项：danger / success / disabled / surface / 不同层级阴影
- [x] 第一版 token 清单：
  - 颜色：`bg` / `fg` / `border` / `surface` / `muted` / `muted-soft` / `danger` / `success` / `disabled-*`
  - 边框：`border-width` / `border-style` / `border-color` / `border`
  - 圆角：`radius-none`
  - 阴影：`shadow-button` / `shadow-dropdown` / `shadow-modal` / `shadow-pressed`
  - 间距：`space-1` ~ `space-6`
  - 字体：`text-xs` / `text-sm` / `text-md` / `text-lg`
  - 字重：`font-weight-normal` / `font-weight-strong` / `font-weight-black`
  - 字距：`letter-spacing-tight` / `letter-spacing-wide` / `letter-spacing-widest`
  - 层级：`z-dropdown` / `z-popover` / `z-modal-backdrop` / `z-modal` / `z-toast`

### 阶段 2：设计基础组件 API
- [x] 为 AButton 设计 props（variant/size/loading/disabled/block）
- [x] 为 AInput / ATextarea 设计 props（modelValue/label/placeholder/disabled/error/hint/rows）
- [x] 为 ASelect 设计 props（modelValue/options/placeholder/disabled/label/error/hint）
- [x] 为 AModal 设计 props（open/title/size/closable + body/footer 插槽）
- [x] 为 ADropdown / APopover 设计触发方式、插槽和点击外部关闭规则

### 阶段 3：实现与收口
- [x] 先实现 AButton
- [x] 再实现 AInput / ATextarea
- [x] 再实现 ASelect
- [x] 再实现 AModal
- [x] 再实现 ADropdown / APopover

### 阶段 4：页面迁移顺序
- [ ] 登录 / 注册页
- [ ] 顶栏用户菜单 / 下拉菜单
- [ ] 各种确认弹窗
- [ ] 常见表单页
- [ ] 管理页筛选与操作区

## 当前推荐的下一步
1. 开始把现有页面里的内联样式和散落的原生 `<select>` / 下拉菜单逐步迁移到新基础组件
2. 优先替换：登录 / 注册 → 顶栏菜单 → 确认弹窗 → 表单页 → 管理页筛选区
3. 在页面迁移过程中，逐步把旧的 `outline` / 内联红色按钮写法收口到统一 `variant`

## 遇到的错误
| 错误 | 尝试次数 | 解决方案 |
|------|---------|---------|
| 暂无 | 0 | 暂无 |
