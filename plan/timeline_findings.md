# 发现与决策

## 需求
时间线模块应服务于“历史事件整理、关系关联、可视化浏览和知识沉淀”，而不是普通文章列表。

## 产品目标
- 让用户可以录入事件并组织成可浏览的时间序列。
- 让事件与人物、地点、标签、来源形成知识关联。
- 让读者能按时间、主题、地理维度理解一组历史资料。
- 让系统具备档案感、可追溯性和长期扩展能力。

## 推荐对象模型

### 1. TimelineEvent
建议字段：
- `id`
- `title`
- `slug`
- `summary`
- `description`
- `start_date`
- `end_date`
- `date_precision`（day / month / year / range / uncertain）
- `category`
- `importance`
- `created_by`
- `status`

### 2. TimelinePerson
建议字段：
- `id`
- `name`
- `aliases`
- `bio`
- `birth_date`
- `death_date`
- `country`

### 3. TimelinePlace
建议字段：
- `id`
- `name`
- `aliases`
- `latitude`
- `longitude`
- `region`
- `description`

### 4. Relation / Tag / Source
- Relation：事件与人物、地点、其他事件之间的关联
- Tag：主题聚合，如“革命”“战争”“音乐史”
- Source：来源引用、文献、网页、书籍、档案编号

## 推荐页面结构

### 1. 时间线主视图
- 时间轴主画布
- 时间缩放
- 分类/标签筛选
- 搜索栏
- 事件预览卡片

### 2. 事件详情页
- 标题、时间、摘要
- 正文描述
- 相关人物
- 相关地点
- 来源列表
- 相关事件跳转

### 3. 人物页 / 地点页
- 基本资料
- 关联事件列表
- 时间分布概览

## API 规划建议
- `POST /api/timeline/events`
- `GET /api/timeline/events`
- `GET /api/timeline/events/:id`
- `PUT /api/timeline/events/:id`
- `POST /api/timeline/persons`
- `POST /api/timeline/places`
- `POST /api/timeline/events/:id/relations`
- `POST /api/timeline/events/:id/sources`
- `GET /api/timeline/search`
- `GET /api/timeline/visualization`

## 分阶段开发建议

### Phase A
- Event CRUD
- 基础时间轴
- 事件详情页

### Phase B
- 人物/地点关联
- 标签过滤
- 搜索

### Phase C
- 来源引用
- 关系图增强
- 数据质量治理

### Phase D
- 地图联动
- 高级可视化
- 知识图谱型探索

## 风险与取舍
- **时间表达复杂度**：历史资料常含模糊时间。
- **信息密度问题**：时间轴上事件过多时需要分层展示。
- **来源质量问题**：没有引用的事件会削弱可信度。
- **关系维护成本**：人物、地点、事件的多对多关系较复杂。

## 最推荐的首期范围
1. 事件 CRUD
2. 基础时间轴
3. 事件详情页
4. 人物/地点关联
5. 来源列表
6. 搜索与筛选