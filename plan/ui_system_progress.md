# 进度日志：统一风格组件系统

## 2026-05-12
- 与用户逐步确认了第一批基础组件系统的视觉方向。
- 已确认整体风格：黑白极简、档案馆 / Brutalist、高对比、0 圆角、2px 黑边。
- 已确认 AButton 的四类语义及 primary / secondary / danger / ghost 的交互规则。
- 已确认 AInput / ATextarea 采用扁平化方案，focus 不悬浮、不加阴影，仅靠边框细节变化表达状态。
- 已确认 ASelect 的 trigger 与 input 同体系，dropdown panel 使用硬阴影，selected 采用左侧标记方案。
- 已确认 AModal 需要独立标题栏和固定 footer。
- 已确认 ADropdown / APopover 共用同一套硬风格弹层语言。
- 已整理出后续实施路线：tokens → 组件 API → 实现顺序 → 页面迁移顺序。
- 继续与用户确认了 design tokens 的拆分深度与落点：采用更完整的 token 方案，并统一放入 `web/src/style.css` 的全局 CSS variables。
- 已完成第一版 token 规划，覆盖颜色、边框、圆角、阴影、间距、字号、字重、字距、层级。
- 已完成第一版组件 API 规划，覆盖 `AButton`、`AInput`、`ATextarea`、`ASelect`、`AModal`、`ADropdown`、`APopover`。
- 已将第一版 tokens 与基础通用类落入 `web/src/style.css`。
- 已升级/实现基础组件：
  - `src/components/ui/ABtn.vue`
  - `src/components/ui/AInput.vue`
  - `src/components/ui/ATextarea.vue`
  - `src/components/ui/AModal.vue`
  - `src/components/ui/ASelect.vue`
  - `src/components/ui/ADropdown.vue`
  - `src/components/ui/APopover.vue`
- 兼容性策略：`ABtn` 同时兼容旧写法（如 `outline` / `danger`）和新写法（`variant`），以避免一次性修改大量页面。
- 已执行前端构建验证：`npm run build` 通过。