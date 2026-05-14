# Podcast Spec

## Purpose
Podcast 是 Studio 的播客子模块，用于让创作者通过频道发布播客节目（Show + Episode），复用统一创作身份与底部播放条，同时输出符合播客规范的 RSS feed。

## Core Model

### Identity and ownership
- Podcast 复用 Studio 频道体系。
- Show = 频道。
- Episode = 频道下发布的播客内容。
- 频道不受单一内容类型限制，播客只是频道内容类型之一。

### Content rules
- 音频文件必须上传，不支持外链音频。
- Episode 支持草稿 / 发布状态。
- 支持定时发布。
- 可见范围支持公开 / 仅关注者 / 私密。
- Shownotes 仅保留 URL 与加粗，不需要完整 Markdown 编辑器。
- 不做时间章节。

## Distribution Rules
- 频道主页需要按内容类型分 Tab 展示：全部 / 文章 / 视频 / 播客。
- 播客内容应有独立 RSS：`/channel/:slug/rss/podcast`。
- 播客 RSS 必须包含 `<enclosure>` 音频标签，以符合外部播客客户端规范。

## Listening Experience
- 收听复用底部播放条。
- 单集可加入队列播放。
- 评论复用论坛纯文本 + 图片格式。

## Durable Decisions
- Podcast 不是独立身份系统，而是 Studio 频道体系下的内容子模块。
- Show 与 Channel 合一，避免重复建模。
- 播客说明不需要复杂编辑能力，因此 shownotes 保持轻量。
- 外部客户端兼容性要求独立播客 RSS 成为首期基础能力。
- 收听体验与音乐模块保持统一，而不是重新设计独立播放器。

## Open Boundaries
- 是否允许 Feed 模块直接订阅播客 RSS 仍可后续扩展。
- 是否支持跨节目收听队列仍可后续决策。
- 节目封面与单集封面是否分离仍可增强。
