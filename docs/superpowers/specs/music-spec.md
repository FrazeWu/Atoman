# Music Spec

## Purpose
音乐模块不是单纯上传与播放系统，而是一个以专辑为中心条目的维基式音乐资料库。用户可以自由上传、编辑、替换资源、查看历史，并通过讨论、回滚与保护机制处理争议。

## Core Model

### Album-centered model
- 专辑是基础对象。
- 所有编辑、历史、争议、回滚都围绕专辑发生。
- 歌曲是专辑内部结构，不是首期独立治理中心。
- 单曲视作只包含一首歌的 `single` 类型专辑。

### Main entities
- `Album`
- `Song`
- `Artist`
- `Revision`
- `Discussion`
- `Protection`
- `LyricAnnotation`
- `ArtistAlias`
- `ArtistMerge`

### Status and governance
- 状态最少包括 `Open`、`Confirmed`、`Disputed`。
- 普通用户可在 Open / Disputed 条目上编辑。
- 管理员负责确认、重新开放、解决争议、回滚、保护与合并条目。
- 维基模式遵循“先改，事后纠正”，所有修改必须保留历史。

## Product Rules
- 所有修改都以专辑为基础，而不是按歌曲独立治理。
- 普通登录用户可创建和编辑音乐条目。
- 所有编辑必须写入 revision 历史。
- 回滚通过“基于旧版本创建新版本”实现，不直接覆盖历史。
- 不同用户意见冲突通过编辑说明、讨论、回滚与保护解决，而不是“谁上传谁做主”。
- 删除应弱化，优先使用保护、隐藏、回滚。
- 艺术家条目与专辑条目共享同一套维基化逻辑。
- 艺术家支持主名 + 别名，别名参与搜索。
- 重复条目合并由管理员执行，被合并条目永久重定向到主条目。

## Playback and annotation rules
- 播放队列以专辑为单位。
- 点击任意曲目会从该曲开始播放整张专辑。
- 底部播放条与音乐大页面共用统一播放体验。
- 右侧面板默认显示歌词解析，播放列表可覆盖其上。
- 歌词解析采用 Genius 风格按行注释，多用户解析互不覆盖。
- 歌词解析编辑器复用博客 Tiptap 编辑器，但仅保留 WYSIWYG 模式。

## Page Structure

### Music home
- 搜索与筛选
- 专辑列表
- 创建专辑入口

### Album detail
- 头部信息区
- 曲目列表
- 专辑简介
- 状态与管理操作

### Album edit
- 专辑元数据编辑
- 曲目列表编辑
- 封面与标签编辑
- 编辑说明

### Album history
- revision 时间线
- 版本快照
- 差异展示
- 回滚入口

### Discussion
- 条目状态提示
- 讨论线程
- 管理员争议处理入口

### Artist surfaces
- Artist 详情
- Artist 编辑
- Artist 历史
- Artist 讨论

## Durable Decisions
- 专辑高于歌曲，是首期统一治理中心。
- single / EP / album 走同一套逻辑。
- 音乐资料事实空间有限，不需要复杂信任层级，只需普通用户 / 管理员两层权限。
- 维基治理的第一工具是历史，而不是预审。
- 艺术家与专辑应共享一致的数据治理与操作模型。

## UX Structure Reference
- `/music` 为专辑发现页。
- `/music/artists/:id` 为艺人详情页。
- `/music/albums/:id` 为专辑详情页。
- `/music/albums/:id/history` 为历史页。
- `/music/albums/:id/discussion` 为讨论页。
- `/music/albums/:id/proposals` 仅在 Confirmed 条目上有意义。
