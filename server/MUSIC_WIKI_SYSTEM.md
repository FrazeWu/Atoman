# Music Module - Wiki-Style Collaboration System

## 概览

音乐模块已改造为维基百科式的协作编辑系统，支持：

- ✅ 任何登录用户可编辑（移除了 AdminMiddleware 限制）
- ✅ 完整的版本历史和回退功能
- ✅ 3-way merge 冲突检测
- ✅ 内容保护机制（none/semi/full）
- ✅ 审核工作流（pending/approved/rejected）

## 数据模型

### Revision（版本记录）
- 存储每次编辑的完整快照（JSON 格式）
- 形成版本链（PreviousRevisionID）
- 支持版本对比和回退

### EditConflict（编辑冲突）
- 自动检测并发编辑冲突
- 记录冲突字段和不同值
- 支持手动解决

### ContentProtection（保护机制）
- `none`: 任何人可编辑，自动批准
- `semi`: 任何人可编辑，需审核
- `full`: 只有管理员可编辑

### Discussion（讨论）
- 每个专辑/歌曲的讨论页面
- 支持嵌套回复（Markdown 格式）
- 软删除机制

## API 端点

### 版本历史

```bash
# 获取版本列表
GET /api/albums/:id/revisions?limit=50&offset=0

# 获取特定版本
GET /api/albums/:id/revisions/:version

# 对比两个版本
GET /api/albums/:id/revisions/diff?v1=1&v2=3
```

### 编辑操作

```bash
# 创建新修订（任何登录用户）
POST /api/albums/:id/edit
{
  "base_revision": 5,           # 基于的版本号
  "changes": {
    "title": "New Title",
    "year": 2023
  },
  "edit_summary": "Fixed typo"
}

# 响应：
{
  "data": {
    "id": "uuid",
    "version_number": 6,
    "status": "pending",        # or "approved" (admin)
    ...
  },
  "message": "Changes saved and pending approval"
}

# 如果有冲突：
{
  "error": "Edit conflicts detected",
  "conflicts": [
    {
      "field_name": "title",
      "base_value": "Old Title",
      "value1": "Your Title",
      "value2": "Someone Else's Title"
    }
  ]
}
```

### 版本回退

```bash
# 回退到指定版本（仅登录用户）
POST /api/albums/:id/revert/:version
{
  "edit_summary": "Reverting vandalism"
}
```

### 审核操作（管理员）

```bash
# 批准修订
POST /api/admin/revisions/:id/approve
{
  "review_notes": "Looks good"
}

# 驳回修订
POST /api/admin/revisions/:id/reject
{
  "review_notes": "Incorrect information"
}
```

## 数据迁移

### 运行迁移

```bash
cd server
go run cmd/migrate_revision_system/main.go
```

迁移会执行：
1. 创建新表：`revisions`, `edit_conflicts`, `content_protections`, `discussions`
2. 为每个 Album/Song 创建初始版本（version 1）
3. 将所有 AlbumCorrection/SongCorrection 转换为 Revision

### 回滚迁移（谨慎使用）

```bash
go run cmd/migrate_revision_system/main.go rollback
```

## 工作流示例

### 场景 1：协作编辑

```
1. 用户 A 和用户 B 同时编辑同一专辑
2. 用户 B 先提交 (基于版本 5) → 创建版本 6
3. 用户 A 后提交 (基于版本 5) → 检测到冲突
4. 系统返回冲突详情，用户 A 手动解决
5. 管理员审核批准 → 版本 7 生效
```

### 场景 2：保护机制

```
1. 高流量专辑设置为 semi-protected
2. 普通用户编辑 → status='pending'
3. 管理员批准后 → status='approved', is_current=true
4. 更改应用到实际 Album 记录
```

### 场景 3：版本回退

```
1. 发现版本 8 有错误数据
2. 管理员查看版本历史，发现版本 6 正确
3. 点击"回退到版本 6"
4. 系统创建版本 9（内容与版本 6 相同）
5. 版本历史清晰记录回退操作
```

## 权限说明

| 操作 | 普通用户 | 管理员 |
|-----|---------|--------|
| 查看版本历史 | ✓ | ✓ |
| 编辑无保护内容 | ✓ (需审核) | ✓ (自动批准) |
| 编辑半保护内容 | ✓ (需审核) | ✓ (自动批准) |
| 编辑全保护内容 | ✗ | ✓ |
| 批准修订 | ✗ | ✓ |
| 回退版本 | ✗ | ✓ |
| 设置保护 | ✗ | ✓ |

## 前端集成（待实现）

前端需要实现以下功能：
1. 版本历史页面（`AlbumHistoryView.vue`）
2. 版本对比工具（diff 可视化）
3. 编辑时的冲突解决界面
4. 保护状态显示
5. 讨论页面

详见计划文件：`.claude/plans/dynamic-seeking-raven.md`

## 测试

### 手动测试流程

```bash
# 1. 运行迁移
go run cmd/migrate_revision_system/main.go

# 2. 启动服务器
go run cmd/start_server/main.go

# 3. 测试 API（需要 token）
# 获取版本列表
curl http://localhost:8080/api/albums/{album_id}/revisions

# 创建编辑
curl -X POST http://localhost:8080/api/albums/{album_id}/edit \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "base_revision": 1,
    "changes": {"title": "Updated Title"},
    "edit_summary": "Test edit"
  }'

# 查看版本对比
curl "http://localhost:8080/api/albums/{album_id}/revisions/diff?v1=1&v2=2"
```

## 后续开发

阶段 2（讨论和保护功能）：
- [ ] Discussion API 实现
- [ ] ContentProtection API 实现
- [ ] 前端讨论组件

阶段 3（前端完整体验）：
- [ ] AlbumHistoryView.vue
- [ ] RevisionDiffView.vue
- [ ] ConflictResolver.vue

阶段 4（测试和优化）：
- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能优化

## 技术细节

### 3-Way Merge 算法

```
对于每个用户修改的字段：
  if 用户值 == 基础值:
    → 无修改，跳过
  else if 用户值 == 当前值:
    → 已被他人改为相同值，无冲突
  else if 基础值 != 当前值:
    → 冲突！需要人工解决
  else:
    → 自动合并
```

### 版本快照存储

- 使用 JSONB 存储完整对象（PostgreSQL）
- 或 JSON 字符串（SQLite）
- 优点：简单、可恢复任何历史状态
- 缺点：存储空间较大（后续可压缩旧版本）

## License

GPL（与主项目一致）
