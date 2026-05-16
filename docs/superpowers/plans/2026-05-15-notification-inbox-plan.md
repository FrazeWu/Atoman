# 通知与私信（Inbox）实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现全平台统一收件箱，包含通知系统（forum_reply / forum_mention / forum_like / forum_solved）和一对一私信（DM），共用 `/ws/user` WebSocket 连接，前端统一入口 `/inbox`。

**Architecture:** 后端新增 `NotificationService`（通知创建/去重/赞合并）和 DM handler，扩展现有 collab hub 支持按 `userID` 路由的私有频道；前端新增 `notificationStore`、`dmStore`、`inboxStore`（协调层），`/inbox` 页面使用左右分栏 + 4 Tab 布局。

**Tech Stack:** Go / Gin / GORM / PostgreSQL / gorilla-websocket（已有）；Vue 3.5 / Pinia / TypeScript / Tailwind CSS v4（已有）

**规格文档：**
- `docs/superpowers/specs/2026-05-15-notification-design.md`
- `docs/superpowers/specs/2026-05-15-dm-design.md`

---

## Task 1：后端数据模型与迁移

**Files:**
- Create: `server/internal/model/notification.go`
- Create: `server/internal/model/dm.go`
- Modify: `server/internal/model/user.go`（UserSettings 新增 `DMPermission`）
- Modify: `server/cmd/start_server/main.go`（AutoMigrate 注册新模型）

- [ ] **Step 1: 创建 notification 模型**

```go
// server/internal/model/notification.go
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NotificationMeta is the JSONB payload stored per notification type.
type NotificationMeta map[string]interface{}

func (m NotificationMeta) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *NotificationMeta) Scan(value interface{}) error {
	if value == nil {
		*m = NotificationMeta{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return fmt.Errorf("NotificationMeta: unsupported scan type %T", value)
	}
	if len(raw) == 0 || string(raw) == "null" {
		*m = NotificationMeta{}
		return nil
	}
	return json.Unmarshal(raw, m)
}

// Notification represents a single notification for a recipient.
type Notification struct {
	Base
	RecipientID uuid.UUID         `json:"recipient_id" gorm:"type:uuid;not null;index"`
	Recipient   *User             `json:"recipient,omitempty" gorm:"foreignKey:RecipientID;references:UUID"`
	ActorID     *uuid.UUID        `json:"actor_id" gorm:"type:uuid;index"`
	Actor       *User             `json:"actor,omitempty" gorm:"foreignKey:ActorID;references:UUID"`
	Type        string            `json:"type" gorm:"not null"`        // "forum_reply" | "forum_mention" | "forum_solved" | "forum_like"
	SourceType  string            `json:"source_type" gorm:"not null"` // "forum_reply" | "forum_topic"
	SourceID    uuid.UUID         `json:"source_id" gorm:"type:uuid;not null"`
	Meta        NotificationMeta  `json:"meta" gorm:"type:jsonb;default:'{}'"`
	ReadAt      *time.Time        `json:"read_at"`
}

func (Notification) TableName() string { return "notifications" }
```

- [ ] **Step 2: 创建 DM 模型**

```go
// server/internal/model/dm.go
package model

import (
	"time"

	"github.com/google/uuid"
)

// DMConversation is a unique 1-to-1 conversation between two users.
// ParticipantA < ParticipantB (UUID byte order) — enforces the UNIQUE constraint.
type DMConversation struct {
	Base
	ParticipantA       uuid.UUID  `json:"participant_a" gorm:"type:uuid;not null;index"`
	ParticipantB       uuid.UUID  `json:"participant_b" gorm:"type:uuid;not null;index"`
	LastMessageAt      *time.Time `json:"last_message_at"`
	LastMessagePreview string     `json:"last_message_preview" gorm:"size:100"`
}

func (DMConversation) TableName() string { return "dm_conversations" }

// DMMessage is a single message in a DMConversation.
type DMMessage struct {
	Base
	ConversationID uuid.UUID  `json:"conversation_id" gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID  `json:"sender_id" gorm:"type:uuid;not null;index"`
	Sender         *User      `json:"sender,omitempty" gorm:"foreignKey:SenderID;references:UUID"`
	Content        string     `json:"content" gorm:"type:text"`  // may be empty for image-only messages
	ImageURL       string     `json:"image_url" gorm:"column:image_url"`
	ReadAt         *time.Time `json:"read_at"`
}

func (DMMessage) TableName() string { return "dm_messages" }
```

- [ ] **Step 3: UserSettings 新增 DMPermission**

在 `server/internal/model/user.go` 的 `UserSettings` struct 末尾加一个字段：

```go
// before (existing):
type UserSettings struct {
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	PrivateProfile bool      `json:"private_profile" gorm:"default:false"`
}

// after:
type UserSettings struct {
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	PrivateProfile bool      `json:"private_profile" gorm:"default:false"`
	DMPermission   string    `json:"dm_permission" gorm:"default:'anyone'"` // "anyone" | "following_only" | "one_before_reply"
}
```

- [ ] **Step 4: 注册新模型到 AutoMigrate**

在 `server/cmd/start_server/main.go` 的 `db.AutoMigrate(...)` 调用中，在 `&model.SiteSetting{}` 之前追加：

```go
&model.Notification{},
&model.DMConversation{},
&model.DMMessage{},
```

- [ ] **Step 5: 创建数据库唯一索引迁移文件**

```go
// server/internal/migrations/notification_dm_indexes.go
package migrations

import "gorm.io/gorm"

// RunNotificationDMIndexes 创建通知去重唯一索引和 DM 会话唯一索引
func RunNotificationDMIndexes(db *gorm.DB) error {
	// 通知去重索引
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_notification_dedup
		ON notifications (recipient_id, source_type, source_id)`)

	// 通知未读查询索引
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notification_recipient_read
		ON notifications (recipient_id, read_at)`)

	// DM 会话唯一索引（保证两人之间只有一条记录）
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_dm_conversation
		ON dm_conversations (participant_a, participant_b)`)

	// DM 消息分页索引
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_dm_message_conv_created
		ON dm_messages (conversation_id, created_at)`)

	return nil
}
```

在 `main.go` 的 AutoMigrate 之后调用：

```go
if err := migrations.RunNotificationDMIndexes(db); err != nil {
    log.Printf("WARN: notification/dm index migration: %v", err)
}
```

- [ ] **Step 6: 构建验证**

```bash
cd server && go build ./...
```

Expected: 无编译错误

- [ ] **Step 7: Commit**

```bash
git add server/internal/model/notification.go \
        server/internal/model/dm.go \
        server/internal/model/user.go \
        server/internal/migrations/notification_dm_indexes.go \
        server/cmd/start_server/main.go
git commit -m "feat: add Notification, DMConversation, DMMessage models and migrations"
```

---

## Task 2：NotificationService

**Files:**
- Create: `server/internal/service/notification_service.go`

- [ ] **Step 1: 创建 NotificationService**

```go
// server/internal/service/notification_service.go
package service

import (
	"bytes"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"atoman/internal/model"
)

// notificationPriority maps notification type to priority for dedup.
var notificationPriority = map[string]int{
	"forum_mention": 3,
	"forum_reply":   2,
	"forum_solved":  1,
	"forum_like":    0,
}

// NotificationService handles creation, dedup, and like-aggregation.
type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification creates or updates a notification with priority-based dedup.
// For forum_like it aggregates (actor_count, recent_actors) instead of inserting a new row.
// Returns the saved notification (nil if skipped — e.g. self-notification).
func (s *NotificationService) CreateNotification(
	recipientID uuid.UUID,
	actorID *uuid.UUID,
	notifType string,
	sourceType string,
	sourceID uuid.UUID,
	meta model.NotificationMeta,
) (*model.Notification, error) {
	// Self-notification guard
	if actorID != nil && *actorID == recipientID {
		return nil, nil
	}

	if notifType == "forum_like" {
		return s.upsertLikeNotification(recipientID, actorID, sourceType, sourceID, meta)
	}

	// Check for existing notification on same (recipient, source_type, source_id)
	var existing model.Notification
	err := s.db.Where(
		"recipient_id = ? AND source_type = ? AND source_id = ?",
		recipientID, sourceType, sourceID,
	).First(&existing).Error

	if err == nil {
		// Existing found — only upgrade if new type has higher priority
		existingPrio := notificationPriority[existing.Type]
		newPrio := notificationPriority[notifType]
		if newPrio <= existingPrio {
			return nil, nil // skip lower/equal priority
		}
		existing.Type = notifType
		existing.ActorID = actorID
		existing.Meta = meta
		existing.ReadAt = nil
		if saveErr := s.db.Save(&existing).Error; saveErr != nil {
			return nil, saveErr
		}
		return &existing, nil
	}

	// No existing — insert new
	notif := &model.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        notifType,
		SourceType:  sourceType,
		SourceID:    sourceID,
		Meta:        meta,
	}
	if err := s.db.Create(notif).Error; err != nil {
		return nil, err
	}
	return notif, nil
}

// upsertLikeNotification aggregates likes: increments actor_count, updates recent_actors.
func (s *NotificationService) upsertLikeNotification(
	recipientID uuid.UUID,
	actorID *uuid.UUID,
	sourceType string,
	sourceID uuid.UUID,
	meta model.NotificationMeta,
) (*model.Notification, error) {
	var existing model.Notification
	err := s.db.Where(
		"recipient_id = ? AND source_type = ? AND source_id = ? AND type = 'forum_like'",
		recipientID, sourceType, sourceID,
	).First(&existing).Error

	actorUsername := ""
	if v, ok := meta["actor_username"].(string); ok {
		actorUsername = v
	}

	if err == nil {
		// Aggregate
		count := 1
		if v, ok := existing.Meta["actor_count"].(float64); ok {
			count = int(v) + 1
		}
		recentActors := []string{}
		if ra, ok := existing.Meta["recent_actors"].([]interface{}); ok {
			for _, a := range ra {
				if s, ok := a.(string); ok {
					recentActors = append(recentActors, s)
				}
			}
		}
		if actorUsername != "" {
			recentActors = prependUnique(recentActors, actorUsername, 3)
		}
		existing.Meta["actor_count"] = count
		existing.Meta["recent_actors"] = recentActors
		existing.ReadAt = nil
		if saveErr := s.db.Save(&existing).Error; saveErr != nil {
			return nil, saveErr
		}
		return &existing, nil
	}

	// First like — insert
	if actorUsername != "" {
		meta["recent_actors"] = []string{actorUsername}
	}
	meta["actor_count"] = 1
	notif := &model.Notification{
		RecipientID: recipientID,
		ActorID:     actorID,
		Type:        "forum_like",
		SourceType:  sourceType,
		SourceID:    sourceID,
		Meta:        meta,
	}
	if err := s.db.Create(notif).Error; err != nil {
		return nil, err
	}
	return notif, nil
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(notifID uuid.UUID, recipientID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&model.Notification{}).
		Where("id = ? AND recipient_id = ?", notifID, recipientID).
		Update("read_at", now).Error
}

// MarkAllRead marks all unread notifications of a given type for a recipient.
// notifType "" means all types.
func (s *NotificationService) MarkAllRead(recipientID uuid.UUID, notifType string) error {
	now := time.Now()
	q := s.db.Model(&model.Notification{}).
		Where("recipient_id = ? AND read_at IS NULL", recipientID)
	if notifType != "" {
		q = q.Where("type = ?", notifType)
	}
	return q.Update("read_at", now).Error
}

// UnreadCount returns the total unread notification count for a recipient.
func (s *NotificationService) UnreadCount(recipientID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&model.Notification{}).
		Where("recipient_id = ? AND read_at IS NULL", recipientID).
		Count(&count).Error
	return count, err
}

// List returns paginated notifications for a recipient, optionally filtered by type.
func (s *NotificationService) List(recipientID uuid.UUID, notifType string, page, pageSize int) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	q := s.db.Model(&model.Notification{}).
		Where("recipient_id = ?", recipientID)
	if notifType != "" {
		q = q.Where("type = ?", notifType)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Preload("Actor").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// NotifyForumReply creates forum_reply and forum_mention notifications
// for a new ForumReply. Mentions are parsed from content via ParseMentions.
func (s *NotificationService) NotifyForumReply(
	db *gorm.DB,
	reply *model.ForumReply,
	topic *model.ForumTopic,
	wsPush func(recipientID uuid.UUID, notif *model.Notification),
) error {
	actorID := reply.UserID

	// 1. Notify topic author of reply (if not same as actor)
	if topic.UserID != actorID {
		meta := model.NotificationMeta{
			"topic_id":      topic.ID.String(),
			"topic_title":   topic.Title,
			"reply_excerpt": truncate(reply.Content, 100),
		}
		notif, err := s.CreateNotification(topic.UserID, &actorID, "forum_reply", "forum_reply", reply.ID, meta)
		if err == nil && notif != nil && wsPush != nil {
			wsPush(topic.UserID, notif)
		}
	}

	// 2. If this is a child reply, notify parent reply author
	if reply.ParentReplyID != nil {
		var parentReply model.ForumReply
		if err := db.First(&parentReply, "id = ?", reply.ParentReplyID).Error; err == nil {
			if parentReply.UserID != actorID && parentReply.UserID != topic.UserID {
				meta := model.NotificationMeta{
					"topic_id":      topic.ID.String(),
					"topic_title":   topic.Title,
					"reply_excerpt": truncate(reply.Content, 100),
				}
				notif, err := s.CreateNotification(parentReply.UserID, &actorID, "forum_reply", "forum_reply", reply.ID, meta)
				if err == nil && notif != nil && wsPush != nil {
					wsPush(parentReply.UserID, notif)
				}
			}
		}
	}

	// 3. Parse @mentions and notify each mentioned user
	mentionedUsers, err := ParseMentions(db, reply.Content)
	if err != nil {
		return err
	}
	for _, user := range mentionedUsers {
		if user.UUID == actorID {
			continue
		}
		meta := model.NotificationMeta{
			"topic_id":      topic.ID.String(),
			"topic_title":   topic.Title,
			"reply_excerpt": truncate(reply.Content, 100),
		}
		notif, err := s.CreateNotification(user.UUID, &actorID, "forum_mention", "forum_reply", reply.ID, meta)
		if err == nil && notif != nil && wsPush != nil {
			wsPush(user.UUID, notif)
		}
	}

	return nil
}

// NotifyForumSolved notifies the reply author that their reply was marked solved.
func (s *NotificationService) NotifyForumSolved(
	reply *model.ForumReply,
	topic *model.ForumTopic,
	markedByID uuid.UUID,
	wsPush func(recipientID uuid.UUID, notif *model.Notification),
) error {
	if reply.UserID == markedByID {
		return nil
	}
	meta := model.NotificationMeta{
		"topic_id":    topic.ID.String(),
		"topic_title": topic.Title,
	}
	notif, err := s.CreateNotification(reply.UserID, &markedByID, "forum_solved", "forum_reply", reply.ID, meta)
	if err == nil && notif != nil && wsPush != nil {
		wsPush(reply.UserID, notif)
	}
	return err
}

// NotifyForumLike notifies the content owner of a new like.
func (s *NotificationService) NotifyForumLike(
	ownerID uuid.UUID,
	actorID uuid.UUID,
	actorUsername string,
	sourceType string, // "forum_topic" or "forum_reply"
	sourceID uuid.UUID,
	topicID uuid.UUID,
	topicTitle string,
	wsPush func(recipientID uuid.UUID, notif *model.Notification),
) error {
	if ownerID == actorID {
		return nil
	}
	meta := model.NotificationMeta{
		"topic_id":       topicID.String(),
		"topic_title":    topicTitle,
		"actor_username": actorUsername,
	}
	notif, err := s.CreateNotification(ownerID, &actorID, "forum_like", sourceType, sourceID, meta)
	if err == nil && notif != nil && wsPush != nil {
		wsPush(ownerID, notif)
	}
	return err
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}

func prependUnique(slice []string, item string, maxLen int) []string {
	// Remove existing occurrence
	filtered := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			filtered = append(filtered, s)
		}
	}
	result := append([]string{item}, filtered...)
	if len(result) > maxLen {
		result = result[:maxLen]
	}
	return result
}

// Ensure bytes import is used (used by bytes.Equal in future extensions)
var _ = bytes.Equal
var _ = clause.OnConflict{}
```

- [ ] **Step 2: 构建验证**

```bash
cd server && go build ./...
```

Expected: 无编译错误

- [ ] **Step 3: Commit**

```bash
git add server/internal/service/notification_service.go
git commit -m "feat: add NotificationService with dedup, mention parsing, like aggregation"
```

---

## Task 3：User Hub（WebSocket 按 userID 路由）

**Files:**
- Create: `server/internal/collab/user_hub.go`
- Modify: `server/cmd/start_server/main.go`（注册 `/ws/user` 路由）

- [ ] **Step 1: 创建 UserHub**

```go
// server/internal/collab/user_hub.go
package collab

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// UserMessage is the envelope pushed to a user's WebSocket connection.
type UserMessage struct {
	Event string      `json:"event"` // "notification" | "dm"
	Data  interface{} `json:"data"`
}

// userClient represents one WebSocket connection for a single user.
type userClient struct {
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
	hub    *UserHub
}

// UserHub routes push messages to individual users (by userID).
// Multiple connections per user are supported (e.g. two browser tabs).
type UserHub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*userClient]struct{} // userID → set of clients
	join    chan *userClient
	leave   chan *userClient
}

func NewUserHub() *UserHub {
	h := &UserHub{
		clients: make(map[uuid.UUID]map[*userClient]struct{}),
		join:    make(chan *userClient, 64),
		leave:   make(chan *userClient, 64),
	}
	go h.run()
	return h
}

func (h *UserHub) run() {
	for {
		select {
		case c := <-h.join:
			h.mu.Lock()
			if h.clients[c.userID] == nil {
				h.clients[c.userID] = make(map[*userClient]struct{})
			}
			h.clients[c.userID][c] = struct{}{}
			h.mu.Unlock()

		case c := <-h.leave:
			h.mu.Lock()
			if conns, ok := h.clients[c.userID]; ok {
				delete(conns, c)
				if len(conns) == 0 {
					delete(h.clients, c.userID)
				}
			}
			h.mu.Unlock()
			close(c.send)
		}
	}
}

// Push sends a message to all connections of a specific user.
func (h *UserHub) Push(userID uuid.UUID, event string, data interface{}) {
	msg := UserMessage{Event: event, Data: data}
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()

	for c := range conns {
		select {
		case c.send <- b:
		default:
			// Slow client — skip (don't block the hub)
		}
	}
}

// ServeWS upgrades the HTTP connection to WebSocket.
// JWT token must be provided as ?token=<jwt> query param or Authorization header.
func (h *UserHub) ServeWS(c *gin.Context, jwtSecret string) {
	userID, err := extractUserIDFromRequest(c, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &userClient{
		conn:   conn,
		send:   make(chan []byte, 64),
		userID: userID,
		hub:    h,
	}
	h.join <- client

	go client.writePump()
	go client.readPump()
}

func (c *userClient) writePump() {
	defer func() {
		c.hub.leave <- c
		c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *userClient) readPump() {
	defer func() {
		c.hub.leave <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	for {
		// We don't expect messages from client — just drain to detect disconnect
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

// extractUserIDFromRequest extracts the user UUID from the JWT token.
// Accepts: Authorization: Bearer <token>  OR  ?token=<jwt>
func extractUserIDFromRequest(c *gin.Context, jwtSecret string) (uuid.UUID, error) {
	tokenStr := ""
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	} else if q := c.Query("token"); q != "" {
		tokenStr = q
	}

	if tokenStr == "" {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}

	return uuid.Parse(sub)
}
```

- [ ] **Step 2: 注册 /ws/user 路由到 main.go**

在 `main.go` 的 `collabGroup` 块之后添加：

```go
// User-level WebSocket hub (notifications + DM)
userHub := collab.NewUserHub()
jwtSecret := os.Getenv("JWT_SECRET")
r.GET("/ws/user", func(c *gin.Context) {
    userHub.ServeWS(c, jwtSecret)
})
```

同时在 `main.go` 中将 `userHub` 传给 handler（下面 Task 4 用到），通过 dependency injection：

```go
// Pass userHub to notification and DM handlers
handlers.SetupNotificationRoutes(r, db, userHub)
handlers.SetupDMRoutes(r, db, userHub, s3Client)
```

- [ ] **Step 3: 确认 go.mod 中有 golang-jwt/jwt**

```bash
cd server && grep "golang-jwt" go.mod
```

如果不存在，运行：
```bash
go get github.com/golang-jwt/jwt/v4
```

（查看现有 auth_handler.go 里用的 JWT 库，统一使用相同版本）

- [ ] **Step 4: 构建验证**

```bash
cd server && go build ./...
```

Expected: 无编译错误

- [ ] **Step 5: Commit**

```bash
git add server/internal/collab/user_hub.go server/cmd/start_server/main.go
git commit -m "feat: add UserHub (per-user WebSocket routing) and /ws/user endpoint"
```

---

## Task 4：通知 HTTP Handler

**Files:**
- Create: `server/internal/handlers/notification_handler.go`

- [ ] **Step 1: 创建 notification handler**

```go
// server/internal/handlers/notification_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/collab"
	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/service"
)

type notificationHandler struct {
	db      *gorm.DB
	svc     *service.NotificationService
	userHub *collab.UserHub
}

func SetupNotificationRoutes(r *gin.Engine, db *gorm.DB, userHub *collab.UserHub) {
	svc := service.NewNotificationService(db)
	h := &notificationHandler{db: db, svc: svc, userHub: userHub}

	auth := r.Group("/api/notifications")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("", h.list)
		auth.GET("/unread-count", h.unreadCount)
		auth.PUT("/:id/read", h.markRead)
		auth.PUT("/read-all", h.markAllRead)
	}
}

// GET /api/notifications?page=1&type=forum_reply|forum_mention|forum_like
func (h *notificationHandler) list(c *gin.Context) {
	claims := middleware.GetClaims(c)
	recipientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	notifType := c.Query("type") // empty = all
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	const pageSize = 20

	notifications, total, err := h.svc.List(recipientID, notifType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  notifications,
		"total": total,
		"page":  page,
	})
}

// GET /api/notifications/unread-count
func (h *notificationHandler) unreadCount(c *gin.Context) {
	claims := middleware.GetClaims(c)
	recipientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	count, err := h.svc.UnreadCount(recipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// PUT /api/notifications/:id/read
func (h *notificationHandler) markRead(c *gin.Context) {
	claims := middleware.GetClaims(c)
	recipientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.MarkRead(notifID, recipientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PUT /api/notifications/read-all?type=forum_reply
func (h *notificationHandler) markAllRead(c *gin.Context) {
	claims := middleware.GetClaims(c)
	recipientID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	notifType := c.Query("type") // empty = all types

	if err := h.svc.MarkAllRead(recipientID, notifType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// WsPushNotif is a helper used by Forum handlers to push real-time notifications.
func WsPushNotif(userHub *collab.UserHub) func(recipientID uuid.UUID, notif *model.Notification) {
	return func(recipientID uuid.UUID, notif *model.Notification) {
		userHub.Push(recipientID, "notification", notif)
	}
}
```

- [ ] **Step 2: 检查 middleware.GetClaims 和 middleware.AuthRequired 的实际签名**

```bash
grep -n "GetClaims\|AuthRequired\|UserID" server/internal/middleware/*.go | head -20
```

根据实际签名调整 `claims.UserID`（可能是 `claims["sub"]` 或其他字段）。

- [ ] **Step 3: 构建验证**

```bash
cd server && go build ./...
```

Expected: 无编译错误

- [ ] **Step 4: Commit**

```bash
git add server/internal/handlers/notification_handler.go
git commit -m "feat: add notification HTTP handler (list, unread-count, mark-read)"
```

---

## Task 5：Forum Handler 接入通知

**Files:**
- Modify: `server/internal/handlers/forum_handler.go`（在回复创建、点赞、Solved 后触发通知）

- [ ] **Step 1: 阅读现有 forum_handler.go 的 reply create 和 like 处理**

```bash
grep -n "createReply\|PostReply\|LikeReply\|SolveTopic\|markSolved\|forum_like\|ForumLike" \
  server/internal/handlers/forum_handler.go | head -30
```

- [ ] **Step 2: SetupForumRoutes 改为接受 NotificationService + UserHub**

修改 `forum_handler.go` 中的 `SetupForumRoutes` 签名（在此文件顶部的 handler struct 中增加字段）：

```go
type forumHandler struct {
	db      *gorm.DB
	notifSvc *service.NotificationService  // 新增
	userHub  *collab.UserHub               // 新增
}

func SetupForumRoutes(r *gin.Engine, db *gorm.DB, notifSvc *service.NotificationService, userHub *collab.UserHub) {
	h := &forumHandler{db: db, notifSvc: notifSvc, userHub: userHub}
	// ... 路由注册不变
}
```

在 `main.go` 中相应修改调用：

```go
notifSvc := service.NewNotificationService(db)
handlers.SetupForumRoutes(r, db, notifSvc, userHub)
handlers.SetupNotificationRoutes(r, db, userHub)
```

- [ ] **Step 3: 在回复创建成功后调用 NotifyForumReply**

找到 `forum_handler.go` 中创建 `ForumReply` 成功后的位置（`db.Create(&reply)` 之后），追加：

```go
// 加载 topic（若 handler 中还没有）
var topic model.ForumTopic
if err := h.db.First(&topic, "id = ?", reply.TopicID).Error; err == nil {
    _ = h.notifSvc.NotifyForumReply(
        h.db,
        &reply,
        &topic,
        WsPushNotif(h.userHub),
    )
}
```

- [ ] **Step 4: 在点赞成功后调用 NotifyForumLike**

在 `forum_handler.go` 的点赞处理（ForumLike 创建成功后），追加：

```go
// 获取内容 owner
var ownerID uuid.UUID
var topicID uuid.UUID
var topicTitle string
switch like.TargetType {
case "topic":
    var t model.ForumTopic
    if err := h.db.Select("user_id, id, title").First(&t, "id = ?", like.TargetID).Error; err == nil {
        ownerID = t.UserID
        topicID = t.ID
        topicTitle = t.Title
    }
case "reply":
    var rp model.ForumReply
    if err := h.db.Select("user_id, topic_id").First(&rp, "id = ?", like.TargetID).Error; err == nil {
        ownerID = rp.UserID
        var t model.ForumTopic
        if err2 := h.db.Select("id, title").First(&t, "id = ?", rp.TopicID).Error; err2 == nil {
            topicID = t.ID
            topicTitle = t.Title
        }
    }
}
if ownerID != uuid.Nil {
    actor := authStore /* 从 claims 获取 */
    _ = h.notifSvc.NotifyForumLike(
        ownerID,
        actorID,
        actorUsername,
        "forum_"+like.TargetType,
        like.TargetID,
        topicID,
        topicTitle,
        WsPushNotif(h.userHub),
    )
}
```

> 注意：根据 `forum_handler.go` 中实际的 like handler 结构调整变量名，保持一致。

- [ ] **Step 5: 在 Solved 标记成功后调用 NotifyForumSolved**

在 `solved` 标记 handler（`POST /api/forum/replies/:id/solve`）中，`db.Save` 成功后：

```go
var topic model.ForumTopic
if err := h.db.First(&topic, "id = ?", reply.TopicID).Error; err == nil {
    _ = h.notifSvc.NotifyForumSolved(
        &reply,
        &topic,
        actorID,
        WsPushNotif(h.userHub),
    )
}
```

- [ ] **Step 6: 构建验证**

```bash
cd server && go build ./...
```

Expected: 无编译错误

- [ ] **Step 7: Commit**

```bash
git add server/internal/handlers/forum_handler.go server/cmd/start_server/main.go
git commit -m "feat: integrate notification triggers into forum reply/like/solve handlers"
```

---

## Task 6：DM Handler

**Files:**
- Create: `server/internal/handlers/dm_handler.go`

- [ ] **Step 1: 创建 DM handler**

```go
// server/internal/handlers/dm_handler.go
package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/collab"
	"atoman/internal/middleware"
	"atoman/internal/model"
	"atoman/internal/storage"

	"github.com/aws/aws-sdk-go/service/s3"
)

type dmHandler struct {
	db      *gorm.DB
	userHub *collab.UserHub
	s3      *s3.S3
}

func SetupDMRoutes(r *gin.Engine, db *gorm.DB, userHub *collab.UserHub, s3Client *s3.S3) {
	h := &dmHandler{db: db, userHub: userHub, s3: s3Client}

	auth := r.Group("/api/dm")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/conversations", h.listConversations)
		auth.GET("/conversations/:username", h.getMessages)
		auth.POST("/conversations/:username", h.sendMessage)
		auth.PUT("/conversations/:username/read", h.markRead)
		auth.GET("/unread-count", h.unreadCount)
		auth.POST("/upload", h.uploadImage)
	}
}

// GET /api/dm/conversations
func (h *dmHandler) listConversations(c *gin.Context) {
	senderID := mustGetUserUUID(c)
	if senderID == uuid.Nil {
		return
	}

	var convs []model.DMConversation
	if err := h.db.Where("participant_a = ? OR participant_b = ?", senderID, senderID).
		Order("last_message_at DESC NULLS LAST").
		Find(&convs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch conversations"})
		return
	}

	// For each conversation, resolve the other participant's username
	type ConvItem struct {
		ConversationID string     `json:"conversation_id"`
		OtherUsername  string     `json:"other_username"`
		OtherUserID    string     `json:"other_user_id"`
		LastMessageAt  *time.Time `json:"last_message_at"`
		Preview        string     `json:"preview"`
		UnreadCount    int64      `json:"unread_count"`
	}
	result := make([]ConvItem, 0, len(convs))
	for _, conv := range convs {
		otherID := conv.ParticipantA
		if conv.ParticipantA == senderID {
			otherID = conv.ParticipantB
		}
		var other model.User
		if err := h.db.Select("uuid, username").First(&other, "uuid = ?", otherID).Error; err != nil {
			continue
		}
		var unread int64
		h.db.Model(&model.DMMessage{}).
			Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", conv.ID, senderID).
			Count(&unread)
		result = append(result, ConvItem{
			ConversationID: conv.ID.String(),
			OtherUsername:  other.Username,
			OtherUserID:    otherID.String(),
			LastMessageAt:  conv.LastMessageAt,
			Preview:        conv.LastMessagePreview,
			UnreadCount:    unread,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GET /api/dm/conversations/:username?page=1
func (h *dmHandler) getMessages(c *gin.Context) {
	senderID := mustGetUserUUID(c)
	if senderID == uuid.Nil {
		return
	}

	var other model.User
	if err := h.db.First(&other, "username = ?", c.Param("username")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	conv, err := h.getOrCreateConversation(senderID, other.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	const pageSize = 30

	var messages []model.DMMessage
	var total int64
	q := h.db.Model(&model.DMMessage{}).Where("conversation_id = ?", conv.ID)
	q.Count(&total)
	q.Preload("Sender").
		Order("created_at ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&messages)

	c.JSON(http.StatusOK, gin.H{"data": messages, "total": total, "page": page})
}

// POST /api/dm/conversations/:username  body: {content, image_url}
func (h *dmHandler) sendMessage(c *gin.Context) {
	senderID := mustGetUserUUID(c)
	if senderID == uuid.Nil {
		return
	}

	var other model.User
	if err := h.db.Preload("Settings").First(&other, "username = ?", c.Param("username")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if other.UUID == senderID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot message yourself"})
		return
	}

	// Permission check
	if err := h.checkDMPermission(senderID, other); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	var body struct {
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if strings.TrimSpace(body.Content) == "" && strings.TrimSpace(body.ImageURL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content or image_url required"})
		return
	}

	conv, err := h.getOrCreateConversation(senderID, other.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}

	msg := model.DMMessage{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Content:        body.Content,
		ImageURL:       body.ImageURL,
	}
	if err := h.db.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send message"})
		return
	}

	// Update conversation preview
	preview := body.Content
	if preview == "" {
		preview = "[图片]"
	}
	if len([]rune(preview)) > 100 {
		preview = string([]rune(preview)[:100])
	}
	now := time.Now()
	h.db.Model(&conv).Updates(map[string]interface{}{
		"last_message_at":      now,
		"last_message_preview": preview,
	})

	// Load sender for WS push
	if err := h.db.Preload("Sender").First(&msg, "id = ?", msg.ID).Error; err == nil {
		h.userHub.Push(other.UUID, "dm", map[string]interface{}{
			"conversation_id":  conv.ID.String(),
			"message_id":       msg.ID.String(),
			"sender_id":        senderID.String(),
			"sender_username":  msg.Sender.Username,
			"content":          msg.Content,
			"image_url":        msg.ImageURL,
			"created_at":       msg.CreatedAt,
		})
	}

	c.JSON(http.StatusCreated, gin.H{"data": msg})
}

// PUT /api/dm/conversations/:username/read
func (h *dmHandler) markRead(c *gin.Context) {
	senderID := mustGetUserUUID(c)
	if senderID == uuid.Nil {
		return
	}

	var other model.User
	if err := h.db.First(&other, "username = ?", c.Param("username")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	conv, err := h.getOrCreateConversation(senderID, other.UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}

	now := time.Now()
	h.db.Model(&model.DMMessage{}).
		Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", conv.ID, senderID).
		Update("read_at", now)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/dm/unread-count
func (h *dmHandler) unreadCount(c *gin.Context) {
	userID := mustGetUserUUID(c)
	if userID == uuid.Nil {
		return
	}

	// Find conversations where user is a participant
	var convIDs []uuid.UUID
	h.db.Model(&model.DMConversation{}).
		Where("participant_a = ? OR participant_b = ?", userID, userID).
		Pluck("id", &convIDs)

	var count int64
	if len(convIDs) > 0 {
		h.db.Model(&model.DMMessage{}).
			Where("conversation_id IN ? AND sender_id != ? AND read_at IS NULL", convIDs, userID).
			Count(&count)
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// POST /api/dm/upload
func (h *dmHandler) uploadImage(c *gin.Context) {
	userID := mustGetUserUUID(c)
	if userID == uuid.Nil {
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}
	defer file.Close()

	if header.Size > 10<<20 { // 10MB
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10MB)"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type"})
		return
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(file)

	key := fmt.Sprintf("dm/%s/%d%s", userID.String(), time.Now().UnixNano(), ext)
	url, err := storage.UploadToS3OrLocal(h.s3, key, buf.Bytes(), header.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_url": url})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func mustGetUserUUID(c *gin.Context) uuid.UUID {
	claims := middleware.GetClaims(c)
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return uuid.Nil
	}
	return id
}

// getOrCreateConversation returns the DMConversation between two users, creating it if needed.
// participant_a is always the UUID with the smaller byte value.
func (h *dmHandler) getOrCreateConversation(userA, userB uuid.UUID) (*model.DMConversation, error) {
	pa, pb := userA, userB
	if bytes.Compare(pa[:], pb[:]) > 0 {
		pa, pb = pb, pa
	}

	var conv model.DMConversation
	err := h.db.Where("participant_a = ? AND participant_b = ?", pa, pb).First(&conv).Error
	if err == nil {
		return &conv, nil
	}

	conv = model.DMConversation{ParticipantA: pa, ParticipantB: pb}
	if err := h.db.Create(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

func (h *dmHandler) checkDMPermission(senderID uuid.UUID, recipient model.User) error {
	// Load settings if not preloaded
	var settings model.UserSettings
	if err := h.db.First(&settings, "user_id = ?", recipient.UUID).Error; err != nil {
		return nil // default: anyone
	}

	switch settings.DMPermission {
	case "following_only":
		var count int64
		h.db.Model(&model.Follow{}).
			Where("follower_id = ? AND following_id = ?", senderID, recipient.UUID).
			Count(&count)
		if count == 0 {
			return fmt.Errorf("dm_permission_denied")
		}
	case "one_before_reply":
		var conv model.DMConversation
		pa, pb := senderID, recipient.UUID
		if bytes.Compare(pa[:], pb[:]) > 0 {
			pa, pb = pb, pa
		}
		if err := h.db.Where("participant_a = ? AND participant_b = ?", pa, pb).First(&conv).Error; err != nil {
			return nil // no conversation yet — first message allowed
		}
		// Check sender has sent at least one message AND recipient has not replied
		var senderCount, recipientCount int64
		h.db.Model(&model.DMMessage{}).
			Where("conversation_id = ? AND sender_id = ?", conv.ID, senderID).
			Count(&senderCount)
		h.db.Model(&model.DMMessage{}).
			Where("conversation_id = ? AND sender_id = ?", conv.ID, recipient.UUID).
			Count(&recipientCount)
		if senderCount >= 1 && recipientCount == 0 {
			return fmt.Errorf("dm_waiting_reply")
		}
	}
	return nil
}
```

- [ ] **Step 2: 检查 storage.UploadToS3OrLocal 的实际函数签名**

```bash
grep -n "func Upload\|func.*S3\|UploadToS3" server/internal/storage/*.go | head -10
```

根据实际签名调整 `dm_handler.go` 中的上传调用。

- [ ] **Step 3: 构建验证**

```bash
cd server && go build ./...
```

Expected: 无编译错误

- [ ] **Step 4: Commit**

```bash
git add server/internal/handlers/dm_handler.go
git commit -m "feat: add DM handler (conversations, messages, send, upload, permission check)"
```

---

## Task 7：前端类型定义

**Files:**
- Modify: `web/src/types.ts`

- [ ] **Step 1: 添加 Notification 和 DM 类型**

在 `web/src/types.ts` 末尾追加：

```typescript
// ─── Notification ────────────────────────────────────────────────────────────

export interface NotificationMeta {
  topic_id?: string
  topic_title?: string
  reply_excerpt?: string
  actor_count?: number
  recent_actors?: string[]
  [key: string]: unknown
}

export interface Notification {
  id: string
  recipient_id: string
  actor_id: string | null
  actor?: {
    uuid: string
    username: string
    display_name: string
    avatar_url: string
  }
  type: 'forum_reply' | 'forum_mention' | 'forum_solved' | 'forum_like'
  source_type: string
  source_id: string
  meta: NotificationMeta
  read_at: string | null
  created_at: string
}

// ─── DM ──────────────────────────────────────────────────────────────────────

export interface DMConversation {
  conversation_id: string
  other_username: string
  other_user_id: string
  last_message_at: string | null
  preview: string
  unread_count: number
}

export interface DMMessage {
  id: string
  conversation_id: string
  sender_id: string
  sender?: {
    uuid: string
    username: string
    display_name: string
    avatar_url: string
  }
  content: string
  image_url: string
  read_at: string | null
  created_at: string
}
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check
```

Expected: 无新增类型错误

- [ ] **Step 3: Commit**

```bash
git add web/src/types.ts
git commit -m "feat: add Notification and DM TypeScript types"
```

---

## Task 8：前端 Store（notificationStore + dmStore + inboxStore）

**Files:**
- Create: `web/src/stores/notification.ts`
- Create: `web/src/stores/dm.ts`
- Create: `web/src/stores/inbox.ts`

- [ ] **Step 1: 创建 notificationStore**

```typescript
// web/src/stores/notification.ts
import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { Notification } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'

export const useNotificationStore = defineStore('notification', () => {
  const unreadCount = ref(0)
  const notifications = ref<Notification[]>([])
  const total = ref(0)
  const loading = ref(false)
  const currentType = ref<string>('')

  const authHeaders = () => {
    const token = localStorage.getItem('token') || ''
    return { Authorization: `Bearer ${token}` }
  }

  async function fetchUnreadCount() {
    try {
      const res = await fetch(`${API_URL}/notifications/unread-count`, {
        headers: authHeaders(),
      })
      if (res.ok) {
        const data = await res.json()
        unreadCount.value = data.count ?? 0
      }
    } catch (e) {
      console.error('Failed to fetch notification unread count', e)
    }
  }

  async function fetchNotifications(type = '', page = 1) {
    loading.value = true
    currentType.value = type
    try {
      const params = new URLSearchParams({ page: String(page) })
      if (type) params.set('type', type)
      const res = await fetch(`${API_URL}/notifications?${params}`, {
        headers: authHeaders(),
      })
      if (res.ok) {
        const data = await res.json()
        notifications.value = data.data ?? []
        total.value = data.total ?? 0
      }
    } finally {
      loading.value = false
    }
  }

  async function markRead(id: string) {
    await fetch(`${API_URL}/notifications/${id}/read`, {
      method: 'PUT',
      headers: authHeaders(),
    })
    const n = notifications.value.find(n => n.id === id)
    if (n) {
      n.read_at = new Date().toISOString()
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
  }

  async function markAllRead() {
    const params = currentType.value ? `?type=${currentType.value}` : ''
    await fetch(`${API_URL}/notifications/read-all${params}`, {
      method: 'PUT',
      headers: authHeaders(),
    })
    notifications.value.forEach(n => { n.read_at = n.read_at ?? new Date().toISOString() })
    await fetchUnreadCount()
  }

  // Called by inboxStore when WS delivers a notification event
  function onWsNotification(notif: Notification) {
    unreadCount.value++
    // Prepend to list if currently viewing this type or all
    if (!currentType.value || currentType.value === notif.type) {
      notifications.value.unshift(notif)
      total.value++
    }
  }

  return { unreadCount, notifications, total, loading, currentType, fetchUnreadCount, fetchNotifications, markRead, markAllRead, onWsNotification }
})
```

- [ ] **Step 2: 创建 dmStore**

```typescript
// web/src/stores/dm.ts
import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { DMConversation, DMMessage } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'

export const useDMStore = defineStore('dm', () => {
  const unreadCount = ref(0)
  const conversations = ref<DMConversation[]>([])
  const activeUsername = ref<string | null>(null)
  const messages = ref<DMMessage[]>([])
  const messagesTotal = ref(0)
  const loading = ref(false)

  const authHeaders = () => {
    const token = localStorage.getItem('token') || ''
    return { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }
  }

  async function fetchUnreadCount() {
    try {
      const res = await fetch(`${API_URL}/dm/unread-count`, { headers: authHeaders() })
      if (res.ok) {
        const data = await res.json()
        unreadCount.value = data.count ?? 0
      }
    } catch (e) {
      console.error('Failed to fetch DM unread count', e)
    }
  }

  async function fetchConversations() {
    loading.value = true
    try {
      const res = await fetch(`${API_URL}/dm/conversations`, { headers: authHeaders() })
      if (res.ok) {
        const data = await res.json()
        conversations.value = data.data ?? []
      }
    } finally {
      loading.value = false
    }
  }

  async function openConversation(username: string, page = 1) {
    activeUsername.value = username
    loading.value = true
    try {
      const res = await fetch(`${API_URL}/dm/conversations/${username}?page=${page}`, {
        headers: authHeaders(),
      })
      if (res.ok) {
        const data = await res.json()
        messages.value = data.data ?? []
        messagesTotal.value = data.total ?? 0
      }
      // Mark as read
      await fetch(`${API_URL}/dm/conversations/${username}/read`, {
        method: 'PUT',
        headers: authHeaders(),
      })
      // Update local unread count for this conversation
      const conv = conversations.value.find(c => c.other_username === username)
      if (conv) {
        unreadCount.value = Math.max(0, unreadCount.value - conv.unread_count)
        conv.unread_count = 0
      }
    } finally {
      loading.value = false
    }
  }

  async function sendMessage(username: string, content: string, imageUrl = '') {
    const res = await fetch(`${API_URL}/dm/conversations/${username}`, {
      method: 'POST',
      headers: authHeaders(),
      body: JSON.stringify({ content, image_url: imageUrl }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'send failed')
    }
    const data = await res.json()
    messages.value.push(data.data)
    return data.data as DMMessage
  }

  // Called by inboxStore when WS delivers a dm event
  function onWsDM(payload: {
    conversation_id: string
    message_id: string
    sender_id: string
    sender_username: string
    content: string
    image_url: string
    created_at: string
  }) {
    if (activeUsername.value === payload.sender_username) {
      // Currently in this conversation — append and auto-mark read
      messages.value.push({
        id: payload.message_id,
        conversation_id: payload.conversation_id,
        sender_id: payload.sender_id,
        sender: { uuid: payload.sender_id, username: payload.sender_username, display_name: payload.sender_username, avatar_url: '' },
        content: payload.content,
        image_url: payload.image_url,
        read_at: new Date().toISOString(),
        created_at: payload.created_at,
      })
      // Auto mark read via API (fire and forget)
      fetch(`${API_URL}/dm/conversations/${payload.sender_username}/read`, {
        method: 'PUT',
        headers: authHeaders(),
      })
    } else {
      unreadCount.value++
      const conv = conversations.value.find(c => c.other_username === payload.sender_username)
      if (conv) {
        conv.unread_count++
        conv.preview = payload.content || '[图片]'
        conv.last_message_at = payload.created_at
      }
    }
  }

  return { unreadCount, conversations, activeUsername, messages, messagesTotal, loading, fetchUnreadCount, fetchConversations, openConversation, sendMessage, onWsDM }
})
```

- [ ] **Step 3: 创建 inboxStore（WS 协调层）**

```typescript
// web/src/stores/inbox.ts
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { Notification } from '@/types'
import { useNotificationStore } from '@/stores/notification'
import { useDMStore } from '@/stores/dm'

const WS_URL = import.meta.env.VITE_WS_URL || (
  window.location.protocol === 'https:' ? 'wss' : 'ws'
) + `://${window.location.host}/ws/user`

export const useInboxStore = defineStore('inbox', () => {
  const notifStore = useNotificationStore()
  const dmStore = useDMStore()
  const totalUnread = computed(() => notifStore.unreadCount + dmStore.unreadCount)

  let ws: WebSocket | null = null
  let pollInterval: ReturnType<typeof setInterval> | null = null
  const wsConnected = ref(false)

  function connect(token: string) {
    if (ws && ws.readyState === WebSocket.OPEN) return

    ws = new WebSocket(`${WS_URL}?token=${encodeURIComponent(token)}`)

    ws.onopen = () => {
      wsConnected.value = true
      if (pollInterval) {
        clearInterval(pollInterval)
        pollInterval = null
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as { event: string; data: unknown }
        if (msg.event === 'notification') {
          notifStore.onWsNotification(msg.data as Notification)
        } else if (msg.event === 'dm') {
          dmStore.onWsDM(msg.data as Parameters<typeof dmStore.onWsDM>[0])
        }
      } catch (e) {
        console.error('Failed to parse WS message', e)
      }
    }

    ws.onclose = () => {
      wsConnected.value = false
      ws = null
      // Fallback polling every 60s
      if (!pollInterval) {
        pollInterval = setInterval(() => {
          notifStore.fetchUnreadCount()
          dmStore.fetchUnreadCount()
        }, 60_000)
      }
      // Reconnect after 5s
      setTimeout(() => {
        const token = localStorage.getItem('token')
        if (token) connect(token)
      }, 5_000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function disconnect() {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
    ws?.close()
    ws = null
  }

  async function init(token: string) {
    await Promise.all([
      notifStore.fetchUnreadCount(),
      dmStore.fetchUnreadCount(),
    ])
    connect(token)
  }

  return { totalUnread, wsConnected, init, disconnect }
})
```

- [ ] **Step 4: 在 App.vue 启动时初始化 inboxStore**

在 `web/src/App.vue` 的 `<script setup>` 中：

```typescript
import { watch } from 'vue'
import { useInboxStore } from '@/stores/inbox'

const inboxStore = useInboxStore()

watch(
  () => authStore.token,
  (token) => {
    if (token) {
      inboxStore.init(token)
    } else {
      inboxStore.disconnect()
    }
  },
  { immediate: true }
)
```

- [ ] **Step 5: 类型检查**

```bash
cd web && bun run type-check
```

Expected: 无类型错误

- [ ] **Step 6: Commit**

```bash
git add web/src/stores/notification.ts web/src/stores/dm.ts web/src/stores/inbox.ts web/src/App.vue
git commit -m "feat: add notificationStore, dmStore, inboxStore with WebSocket connection"
```

---

## Task 9：导航栏「收件箱 N」按钮

**Files:**
- Modify: `web/src/components/AppTopbar.vue`

- [ ] **Step 1: 在 AppTopbar.vue 添加收件箱入口**

在 `<script setup>` 中引入 inboxStore：

```typescript
import { useInboxStore } from '@/stores/inbox'
const inboxStore = useInboxStore()
```

在 `<template>` 的 `<!-- Right side -->` 区域，在登录按钮之前（已认证时）添加：

```html
<RouterLink
  v-if="authStore.isAuthenticated"
  to="/inbox"
  class="nav-link"
  :class="{ active: $route.path.startsWith('/inbox') }"
>
  收件箱<span v-if="inboxStore.totalUnread > 0" class="inbox-badge">{{ inboxStore.totalUnread }}</span>
</RouterLink>
```

在 `<style scoped>` 中添加角标样式：

```css
.inbox-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  margin-left: 4px;
  background: #ef4444;
  color: #fff;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}
```

- [ ] **Step 2: 类型检查**

```bash
cd web && bun run type-check
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/AppTopbar.vue
git commit -m "feat: add inbox button with unread badge to AppTopbar"
```

---

## Task 10：/inbox 页面（通知 + 私信统一收件箱）

**Files:**
- Create: `web/src/views/inbox/InboxPage.vue`
- Modify: `web/src/router.ts`

- [ ] **Step 1: 创建 InboxPage.vue**

```vue
<!-- web/src/views/inbox/InboxPage.vue -->
<template>
  <div class="inbox-layout">
    <!-- 左侧列表 -->
    <aside class="inbox-sidebar">
      <div class="inbox-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="inbox-tab"
          :class="{ active: activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
          <span v-if="tab.unread > 0" class="tab-badge">{{ tab.unread }}</span>
        </button>
      </div>

      <!-- 通知列表 -->
      <div v-if="activeTab !== 'dm'" class="inbox-list">
        <div
          v-for="notif in notifStore.notifications"
          :key="notif.id"
          class="inbox-item"
          :class="{ unread: !notif.read_at, active: selectedNotifId === notif.id }"
          @click="selectNotif(notif)"
        >
          <div class="item-actor">{{ notif.actor?.username ?? '系统' }}</div>
          <div class="item-desc">{{ notifDescription(notif) }}</div>
          <div class="item-meta">
            <span class="item-excerpt">{{ notif.meta.topic_title }}</span>
            <span class="item-time">{{ timeAgo(notif.created_at) }}</span>
          </div>
        </div>
        <div v-if="notifStore.notifications.length === 0 && !notifStore.loading" class="inbox-empty">
          暂无通知
        </div>
        <button
          v-if="notifStore.notifications.length > 0"
          class="mark-all-btn"
          @click="notifStore.markAllRead()"
        >
          全部已读
        </button>
      </div>

      <!-- 私信会话列表 -->
      <div v-else class="inbox-list">
        <div
          v-for="conv in dmStore.conversations"
          :key="conv.conversation_id"
          class="inbox-item"
          :class="{ unread: conv.unread_count > 0, active: dmStore.activeUsername === conv.other_username }"
          @click="openConv(conv.other_username)"
        >
          <div class="item-actor">{{ conv.other_username }}</div>
          <div class="item-desc">{{ conv.preview || '（暂无消息）' }}</div>
          <div class="item-meta">
            <span class="item-time">{{ conv.last_message_at ? timeAgo(conv.last_message_at) : '' }}</span>
            <span v-if="conv.unread_count > 0" class="tab-badge">{{ conv.unread_count }}</span>
          </div>
        </div>
        <div v-if="dmStore.conversations.length === 0 && !dmStore.loading" class="inbox-empty">
          暂无私信
        </div>
      </div>
    </aside>

    <!-- 右侧详情 / 对话区 -->
    <main class="inbox-main">
      <!-- 通知详情 -->
      <template v-if="activeTab !== 'dm' && selectedNotif">
        <div class="detail-header">
          <span class="detail-actor">{{ selectedNotif.actor?.username ?? '系统' }}</span>
          {{ notifDescription(selectedNotif) }}
        </div>
        <div class="detail-body">
          <p v-if="selectedNotif.meta.topic_title" class="detail-topic">
            话题：{{ selectedNotif.meta.topic_title }}
          </p>
          <p v-if="selectedNotif.meta.reply_excerpt" class="detail-excerpt">
            {{ selectedNotif.meta.reply_excerpt }}
          </p>
          <p v-if="selectedNotif.meta.actor_count && selectedNotif.meta.actor_count > 1" class="detail-like">
            共 {{ selectedNotif.meta.actor_count }} 人点赞，包括 {{ (selectedNotif.meta.recent_actors ?? []).join('、') }}
          </p>
          <RouterLink :to="notifTargetUrl(selectedNotif)" class="detail-link">
            跳转到内容 →
          </RouterLink>
        </div>
      </template>

      <!-- 私信对话 -->
      <template v-else-if="activeTab === 'dm' && dmStore.activeUsername">
        <div class="dm-header">与 {{ dmStore.activeUsername }} 的对话</div>
        <div class="dm-messages" ref="messagesEl">
          <div
            v-for="msg in dmStore.messages"
            :key="msg.id"
            class="dm-msg"
            :class="{ 'dm-msg-self': msg.sender_id === authStore.user?.uuid }"
          >
            <div class="dm-bubble">
              <img v-if="msg.image_url" :src="msg.image_url" class="dm-img" @click="openImage(msg.image_url)" />
              <span v-if="msg.content">{{ msg.content }}</span>
            </div>
            <div class="dm-msg-time">{{ timeAgo(msg.created_at) }}</div>
          </div>
        </div>
        <div class="dm-input-row">
          <textarea
            v-model="msgInput"
            class="dm-input"
            placeholder="输入消息..."
            rows="2"
            @keydown.enter.ctrl="doSend"
          />
          <label class="dm-upload-btn">
            上传图片
            <input type="file" accept="image/*" style="display:none" @change="handleImageUpload" />
          </label>
          <button class="dm-send-btn" :disabled="sending" @click="doSend">发送</button>
        </div>
        <div v-if="sendError" class="dm-error">{{ sendError }}</div>
      </template>

      <!-- 空状态 -->
      <div v-else class="inbox-empty-main">
        从左侧选择一条通知或对话
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { RouterLink } from 'vue-router'
import { useNotificationStore } from '@/stores/notification'
import { useDMStore } from '@/stores/dm'
import { useAuthStore } from '@/stores/auth'
import type { Notification } from '@/types'

const API_URL = import.meta.env.VITE_API_URL || '/api'

const notifStore = useNotificationStore()
const dmStore = useDMStore()
const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const activeTab = ref<'reply' | 'like' | 'mention' | 'dm'>('reply')
const selectedNotifId = ref<string | null>(null)
const selectedNotif = computed(() => notifStore.notifications.find(n => n.id === selectedNotifId.value) ?? null)
const msgInput = ref('')
const sending = ref(false)
const sendError = ref('')
const messagesEl = ref<HTMLElement | null>(null)

const tabs = computed(() => [
  { key: 'reply' as const,   label: '回复我的', unread: 0 },
  { key: 'like' as const,    label: '给我的赞', unread: 0 },
  { key: 'mention' as const, label: '@我的',    unread: 0 },
  { key: 'dm' as const,      label: '私信',     unread: dmStore.unreadCount },
])

const tabToType: Record<string, string> = {
  reply:   'forum_reply',
  like:    'forum_like',
  mention: 'forum_mention',
}

async function switchTab(tab: typeof activeTab.value) {
  activeTab.value = tab
  selectedNotifId.value = null
  if (tab === 'dm') {
    await dmStore.fetchConversations()
  } else {
    await notifStore.fetchNotifications(tabToType[tab])
  }
}

async function selectNotif(notif: Notification) {
  selectedNotifId.value = notif.id
  if (!notif.read_at) {
    await notifStore.markRead(notif.id)
  }
}

async function openConv(username: string) {
  await dmStore.openConversation(username)
  await nextTick()
  scrollToBottom()
}

function scrollToBottom() {
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

async function doSend() {
  if (!dmStore.activeUsername || sending.value) return
  const content = msgInput.value.trim()
  if (!content) return
  sending.value = true
  sendError.value = ''
  try {
    await dmStore.sendMessage(dmStore.activeUsername, content)
    msgInput.value = ''
    await nextTick()
    scrollToBottom()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'send failed'
    if (msg === 'dm_waiting_reply') {
      sendError.value = '对方设置了回复前仅可发送一条消息'
    } else if (msg === 'dm_permission_denied') {
      sendError.value = '仅对方关注的用户可发送私信'
    } else {
      sendError.value = '发送失败，请重试'
    }
  } finally {
    sending.value = false
  }
}

async function handleImageUpload(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file || !dmStore.activeUsername) return
  const form = new FormData()
  form.append('image', file)
  const token = localStorage.getItem('token') || ''
  const res = await fetch(`${API_URL}/dm/upload`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  })
  if (res.ok) {
    const data = await res.json()
    await dmStore.sendMessage(dmStore.activeUsername, '', data.image_url)
    await nextTick()
    scrollToBottom()
  }
}

function openImage(url: string) {
  window.open(url, '_blank')
}

function notifDescription(notif: Notification): string {
  switch (notif.type) {
    case 'forum_reply':   return '回复了你的帖子/回复'
    case 'forum_mention': return '@提及了你'
    case 'forum_solved':  return '将你的回复标为解决方案'
    case 'forum_like': {
      const count = notif.meta.actor_count ?? 1
      return count > 1 ? `等 ${count} 人赞了你的内容` : '赞了你的内容'
    }
    default: return '有新通知'
  }
}

function notifTargetUrl(notif: Notification): string {
  switch (notif.source_type) {
    case 'forum_reply':
      return `/topic/${notif.meta.topic_id}#reply-${notif.source_id}`
    case 'forum_topic':
      return `/topic/${notif.source_id}`
    default:
      return '/inbox'
  }
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

// Handle ?tab=dm&user=xxx query params (from user profile "发私信" button)
onMounted(async () => {
  const tabParam = route.query.tab as string | undefined
  const userParam = route.query.user as string | undefined

  if (tabParam === 'dm') {
    await switchTab('dm')
    if (userParam) {
      await openConv(userParam)
    }
  } else {
    await switchTab('reply')
  }
})

// Auto scroll when new DM arrives
watch(() => dmStore.messages.length, () => {
  nextTick(() => scrollToBottom())
})
</script>

<style scoped>
.inbox-layout {
  display: flex;
  max-width: 1152px;
  margin: 0 auto;
  padding: 2rem;
  gap: 0;
  height: calc(100vh - 64px - 128px);
  min-height: 500px;
}

.inbox-sidebar {
  width: 320px;
  flex-shrink: 0;
  border: 2px solid #000;
  border-right: none;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.inbox-main {
  flex: 1;
  border: 2px solid #000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.inbox-tabs {
  display: flex;
  border-bottom: 2px solid #000;
}

.inbox-tab {
  flex: 1;
  padding: 8px 4px;
  font-size: 13px;
  font-weight: 600;
  background: none;
  border: none;
  border-right: 1px solid #e5e7eb;
  cursor: pointer;
  transition: background 0.1s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}
.inbox-tab:last-child { border-right: none; }
.inbox-tab.active { background: #000; color: #fff; }
.inbox-tab:hover:not(.active) { background: #f3f4f6; }

.tab-badge {
  background: #ef4444;
  color: #fff;
  border-radius: 9999px;
  font-size: 10px;
  font-weight: 700;
  padding: 0 5px;
  min-width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.inbox-list {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.inbox-item {
  padding: 12px 14px;
  border-bottom: 1px solid #f1f5f9;
  cursor: pointer;
  transition: background 0.1s;
}
.inbox-item:hover { background: #f8fafc; }
.inbox-item.active { background: #eff6ff; }
.inbox-item.unread .item-actor { font-weight: 700; }

.item-actor { font-size: 14px; font-weight: 500; color: #111; }
.item-desc { font-size: 12px; color: #475569; margin-top: 2px; }
.item-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
}
.item-excerpt { font-size: 11px; color: #94a3b8; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.item-time { font-size: 11px; color: #94a3b8; flex-shrink: 0; margin-left: 8px; }

.mark-all-btn {
  width: 100%;
  padding: 8px;
  border: none;
  border-top: 1px solid #e5e7eb;
  background: #f8fafc;
  font-size: 12px;
  color: #64748b;
  cursor: pointer;
}
.mark-all-btn:hover { background: #f1f5f9; }

.inbox-empty { padding: 32px; text-align: center; color: #94a3b8; font-size: 14px; }

/* 右侧详情 */
.detail-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  font-size: 15px;
}
.detail-actor { font-weight: 700; margin-right: 4px; }
.detail-body { padding: 20px; }
.detail-topic { font-weight: 600; margin-bottom: 8px; }
.detail-excerpt { color: #475569; line-height: 1.6; margin-bottom: 12px; }
.detail-like { color: #f59e0b; margin-bottom: 12px; }
.detail-link { color: #3b82f6; font-size: 14px; }

/* DM */
.dm-header {
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
  font-weight: 600;
  font-size: 14px;
}
.dm-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.dm-msg { display: flex; flex-direction: column; align-items: flex-start; }
.dm-msg.dm-msg-self { align-items: flex-end; }
.dm-bubble {
  max-width: 70%;
  background: #f1f5f9;
  padding: 8px 12px;
  border-radius: 12px 12px 12px 2px;
  font-size: 14px;
  line-height: 1.5;
  word-break: break-word;
}
.dm-msg-self .dm-bubble {
  background: #000;
  color: #fff;
  border-radius: 12px 12px 2px 12px;
}
.dm-img { max-width: 200px; max-height: 200px; border-radius: 4px; cursor: pointer; display: block; }
.dm-msg-time { font-size: 11px; color: #94a3b8; margin-top: 2px; }
.dm-input-row {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #e5e7eb;
  align-items: flex-end;
}
.dm-input {
  flex: 1;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  padding: 8px;
  font-size: 14px;
  resize: none;
  font-family: inherit;
}
.dm-upload-btn {
  font-size: 13px;
  padding: 6px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
  user-select: none;
}
.dm-send-btn {
  font-size: 13px;
  font-weight: 600;
  padding: 6px 16px;
  background: #000;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.dm-send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.dm-error { padding: 4px 16px; color: #ef4444; font-size: 13px; }
.inbox-empty-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-size: 14px;
}
</style>
```

- [ ] **Step 2: 注册路由**

在 `web/src/router.ts` 中，在 podcast 路由之前追加（需要 auth）：

```typescript
{
  path: '/inbox',
  component: () => import('@/views/inbox/InboxPage.vue'),
  meta: { requiresAuth: true },
},
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check
```

Expected: 无类型错误

- [ ] **Step 4: Commit**

```bash
git add web/src/views/inbox/InboxPage.vue web/src/router.ts
git commit -m "feat: add /inbox page with 4-tab layout (notification + DM)"
```

---

## Task 11：用户主页「发私信」按钮

**Files:**
- Modify: `web/src/views/blog/ProfileView.vue`（或现有用户主页文件）

- [ ] **Step 1: 找到用户主页文件中显示用户操作按钮的区域**

```bash
grep -n "关注\|follow\|profile-action\|user-action" web/src/views/blog/ProfileView.vue | head -10
```

- [ ] **Step 2: 在非本人主页时添加「发私信」按钮**

在关注按钮附近，已认证且不是本人时添加：

```html
<RouterLink
  v-if="authStore.isAuthenticated && authStore.user?.username !== profileUser.username"
  :to="`/inbox?tab=dm&user=${profileUser.username}`"
  class="profile-action-btn"
>
  发私信
</RouterLink>
```

- [ ] **Step 3: 类型检查**

```bash
cd web && bun run type-check
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/blog/ProfileView.vue
git commit -m "feat: add send DM button on user profile page"
```

---

## Task 12：用户设置页私信权限

**Files:**
- Modify: `web/src/views/blog/BlogSettingsView.vue`（用户设置页）

- [ ] **Step 1: 找到现有设置页的结构**

```bash
grep -n "settings\|profile\|privacy" web/src/views/blog/BlogSettingsView.vue | head -20
```

- [ ] **Step 2: 添加私信权限设置项**

在设置表单中添加：

```html
<div class="settings-section">
  <h3>私信权限</h3>
  <div class="settings-field">
    <label class="settings-label">谁可以给我发私信</label>
    <div class="settings-radio-group">
      <label class="settings-radio">
        <input type="radio" v-model="dmPermission" value="anyone" />
        任意人可私信
      </label>
      <label class="settings-radio">
        <input type="radio" v-model="dmPermission" value="following_only" />
        仅我关注的人
      </label>
      <label class="settings-radio">
        <input type="radio" v-model="dmPermission" value="one_before_reply" />
        陌生人仅可发一条（回复前）
      </label>
    </div>
  </div>
</div>
```

在现有 `saveSettings` 函数中包含 `dm_permission: dmPermission.value`，或单独提交。

- [ ] **Step 3: 类型检查 + 构建验证**

```bash
cd web && bun run type-check
cd server && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add web/src/views/blog/BlogSettingsView.vue
git commit -m "feat: add DM permission setting on user settings page"
```

---

## Task 13：端到端验证

- [ ] **Step 1: 启动开发服务器**

```bash
# 终端 1 - 后端
cd server && go run cmd/start_server/main.go

# 终端 2 - 前端
cd web && bun run dev
```

- [ ] **Step 2: 验证通知流程**

1. 以用户 A 登录，在 Forum 发一个帖子
2. 以用户 B 登录，回复该帖子，并在回复中 `@用户A的用户名`
3. 切回用户 A，检查：
   - 导航栏「收件箱 N」数字应 = 1（收到 mention，优先级高于 reply，去重生效）
   - 打开 `/inbox`，「回复我的」Tab 有一条，「@我的」Tab 也有一条
   - 点击通知后「跳转到内容 →」链接指向正确帖子

- [ ] **Step 3: 验证点赞合并**

1. 用户 B、C 都点赞用户 A 的回复
2. 用户 A 收件箱「给我的赞」Tab：应显示一条（合并），`actor_count = 2`，`recent_actors` 包含 B 和 C

- [ ] **Step 4: 验证私信流程**

1. 以用户 A 访问用户 B 的主页，点「发私信」
2. 应跳转到 `/inbox?tab=dm&user=B`，自动打开对话
3. 发送消息，用户 B 登录后查看「私信」Tab，应看到消息（WS 实时推送）
4. 验证 `one_before_reply` 权限：将用户 B 权限设为此值，用户 C（陌生人）发一条消息成功，发第二条时返回 403

- [ ] **Step 5: 最终构建检查**

```bash
cd server && go build ./...
cd web && bun run type-check && bun run build
```

Expected: 均无错误

- [ ] **Step 6: 最终 Commit**

```bash
git add -A
git commit -m "feat: complete notification + inbox system (Tasks 1-13)"
```
