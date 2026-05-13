# 研究发现：统一风格组件系统

## 2026-05-12
- 用户希望用“逐步确认”的方式设计统一组件系统，而不是直接重做全站 UI。
- 风格方向已明确为 Archive / Brutalist：黑白、高对比、强结构、0 圆角、2px 黑边。
- 按钮与弹层保留“悬浮感”，输入类组件则采用更扁平、更克制的设计。
- Select / Dropdown / Popover / Modal 是最容易风格漂移的核心对象，应优先纳入统一系统。
- Popover 不单独发展为更轻的子风格，而是保持与 Dropdown 同一套视觉语言，避免再次分叉。
- 统一组件系统的最佳推进方式是：先定 tokens，再定基础组件 API，再逐步替换页面，而不是直接在页面上散写 Tailwind 类。
- tokens 采用更完整的拆分方案，而不是只保留极少量变量；这样更适合后续做完整基础组件体系。
- tokens 的最佳落点是全局 CSS variables（建议放在 `web/src/style.css`），而不是以 Tailwind theme 为主。
- 第一版 token 已确认应覆盖：颜色、边框、圆角、阴影、间距、字号、字重、字距、层级。
- `surface` 需要在第一版 token 中保留，用于 modal header/footer、popover、dropdown 等轻层级区块，但应保持非常克制，不能破坏黑白主基调。
- 第一版 AButton props 已收敛为：`variant`、`size`、`disabled`、`loading`、`block`；不应一开始就扩展出过多变体。
- 第一版 AInput / ATextarea / ASelect / AModal / ADropdown / APopover 也都已收敛到最小必要 API，目标是先建立稳定的底座，再谈扩展。