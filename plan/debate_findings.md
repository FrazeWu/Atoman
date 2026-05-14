# 发现与决策

## 需求
辩论模块应服务于“结构化表达”和“高质量争议讨论”，而不是普通灌水式回帖。

## 产品目标
- 让用户围绕议题发起有组织的辩论。
- 让每个立场下的论点可以清晰分叉、引用证据、持续补充。
- 让读者能快速理解各方主要观点、证据链和结论演变。
- 让治理机制能控制低质量、人身攻击、刷屏和跑题内容。

## 推荐对象模型

### 1. DebateTopic
建议字段：
- `id`
- `title`
- `slug`
- `description`
- `category`
- `status`（open / locked / archived）
- `created_by`
- `cover_image`
- `summary`
- `conclusion_status`

### 2. DebatePosition
建议字段：
- `id`
- `topic_id`
- `title`
- `description`
- `sort_order`
- `created_by`

### 3. DebateArgument
建议字段：
- `id`
- `topic_id`
- `position_id`
- `parent_id`
- `title`
- `body`
- `stance_type`（support / rebuttal / question / clarification）
- `created_by`
- `updated_by`
- `is_folded`
- `vote_score`

### 4. DebateEvidence
建议字段：
- `id`
- `argument_id`
- `source_type`（url / quote / internal_post / internal_music / file）
- `source_ref`
- `title`
- `excerpt`
- `credibility_note`

### 5. Vote / Conclusion / Moderation
- Vote：用户对议题或立场的表态
- Conclusion：阶段性总结、当前共识、未决问题
- Moderation：折叠、锁定、置顶、标注争议等管理操作记录

## 关键交互模式

### 1. 首页
- 热门议题
- 最新议题
- 高争议议题
- 已形成结论的议题

### 2. 议题详情页
建议结构：
- 顶部：标题、简介、分类、状态、参与人数
- 立场栏：不同立场切换
- 中央：论点树
- 侧栏：证据、投票、总结、相关议题

### 3. 论点树
要求：
- 可以区分“支持”“反驳”“追问”“澄清”
- 每个节点可挂多个证据
- 节点支持折叠展开
- 深层结构仍需保持可读性

### 4. 总结系统
- 允许作者或主持人写阶段性总结
- 总结要指出：主要论点、关键证据、分歧点、未解决问题
- 总结应可更新，有版本感

## API 规划建议
- `POST /api/debate/topics`
- `GET /api/debate/topics`
- `GET /api/debate/topics/:id`
- `PUT /api/debate/topics/:id`
- `POST /api/debate/topics/:id/positions`
- `POST /api/debate/topics/:id/arguments`
- `POST /api/debate/arguments/:id/reply`
- `POST /api/debate/arguments/:id/evidences`
- `POST /api/debate/topics/:id/vote`
- `POST /api/debate/topics/:id/conclusions`
- `POST /api/debate/moderation/:id/action`

## 分阶段开发建议

### Phase A
- 议题 CRUD
- 多立场
- 论点树
- 投票表态

### Phase B
- 证据引用
- 总结区
- 折叠/置顶/锁定

### Phase C
- 高级筛选排序
- 跨模块引用
- 用户声誉/贡献统计

## 风险与取舍
- **论点树复杂度**：结构太深会降低可读性。
- **治理成本**：争议内容需要更强 moderation 工具。
- **证据质量问题**：不能默认外部链接就是高质量来源。
- **情绪化表达风险**：需要靠交互和规则引导回到论证。 

## 最推荐的首期范围
1. 议题创建
2. 多立场配置
3. 论点树回复
4. 投票表态
5. 简单总结区
6. 管理员折叠/锁定