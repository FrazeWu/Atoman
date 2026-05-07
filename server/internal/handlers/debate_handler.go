package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"atoman/internal/middleware"
	"atoman/internal/model"
)

func SetupDebateRoutes(router *gin.Engine, db *gorm.DB) {
	debate := router.Group("/api/debate")
	{
		debate.GET("/topics", GetDebateTopics(db))
		debate.GET("/topics/:id", GetDebateTopic(db))
		debate.GET("/topics/:id/arguments", middleware.OptionalAuthMiddleware(), GetDebateArguments(db))
		debate.GET("/topics/search", SearchDebateTopics(db))

		protected := debate.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Topic CRUD
			protected.POST("/topics", CreateDebateTopic(db))
			protected.PUT("/topics/:id", UpdateDebateTopic(db))
			protected.DELETE("/topics/:id", DeleteDebateTopic(db))
			protected.POST("/topics/:id/conclude", ConcludeDebateTopic(db))
			protected.POST("/topics/:id/reopen", ReopenDebateTopic(db))
			protected.POST("/topics/:id/conclude-vote", VoteToConclude(db))

			// Argument CRUD
			protected.POST("/topics/:id/arguments", CreateArgument(db))
			protected.PUT("/arguments/:id", UpdateArgument(db))
			protected.DELETE("/arguments/:id", DeleteArgument(db))

			// Argument references (argument-to-argument)
			protected.POST("/arguments/:id/reference", AddArgumentReference(db))
			protected.DELETE("/arguments/:id/reference/:ref_id", RemoveArgumentReference(db))

			// Argument debate references (argument-to-debate-topic)
			protected.POST("/arguments/:id/debate-reference", AddDebateReference(db))
			protected.DELETE("/arguments/:id/debate-reference/:debate_id", RemoveDebateReference(db))

			// Voting
			protected.POST("/arguments/:id/vote", VoteArgument(db))
			protected.DELETE("/arguments/:id/vote", RemoveVote(db))
			protected.GET("/arguments/:id/votes", GetArgumentVotes(db)) // admin only
		}
	}
}

// ====== Topic Handlers ======

func GetDebateTopics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}
		offset := (page - 1) * limit

		status := c.Query("status")
		tag := c.Query("tag")

		query := db.Model(&model.Debate{}).Preload("User")

		if status != "" {
			query = query.Where("status = ?", status)
		}
		if tag != "" {
			query = query.Where("tags @> ?", []string{tag})
		}

		var total int64
		query.Count(&total)

		var debates []model.Debate
		if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&debates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch debates"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  debates,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func GetDebateTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var debate model.Debate

		if err := db.Preload("User").First(&debate, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		// Increment view count
		db.Model(&debate).Update("view_count", debate.ViewCount+1)

		c.JSON(http.StatusOK, gin.H{"data": debate})
	}
}

type CreateDebateInput struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
}

func CreateDebateTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input CreateDebateInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, _ := c.Get("user_id")
		debate := model.Debate{
			UserID:      userID.(uuid.UUID),
			Title:       input.Title,
			Description: input.Description,
			Content:     input.Content,
			Tags:        input.Tags,
			Status:      "open",
		}

		if err := db.Create(&debate).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create debate"})
			return
		}

		// Preload user for response
		db.Preload("User").First(&debate, debate.ID)

		c.JSON(http.StatusCreated, gin.H{"data": debate})
	}
}

func UpdateDebateTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var debate model.Debate

		if err := db.First(&debate, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		// Check ownership or admin
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if debate.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		var input CreateDebateInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := map[string]interface{}{
			"title":       input.Title,
			"description": input.Description,
			"content":     input.Content,
			"tags":        input.Tags,
		}

		if err := db.Model(&debate).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update debate"})
			return
		}

		db.Preload("User").First(&debate, debate.ID)
		c.JSON(http.StatusOK, gin.H{"data": debate})
	}
}

func DeleteDebateTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var debate model.Debate

		if err := db.First(&debate, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		// Check ownership or admin
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if debate.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		// Cascade delete arguments and votes
		db.Where("debate_id = ?", id).Delete(&model.Argument{})
		db.Delete(&debate)

		c.JSON(http.StatusOK, gin.H{"message": "Debate deleted"})
	}
}

func ConcludeDebateTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var debate model.Debate

		if err := db.First(&debate, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		// Check ownership or admin
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if debate.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		type ConcludeInput struct {
			ConclusionType    string `json:"conclusion_type" binding:"required,oneof=yes no inconclusive"`
			ConclusionSummary string `json:"conclusion_summary"`
		}
		var input ConcludeInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		if err := db.Model(&debate).Updates(map[string]interface{}{
			"status":             "concluded",
			"concluded_at":       now,
			"conclusion_type":    input.ConclusionType,
			"conclusion_summary": input.ConclusionSummary,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to conclude debate"})
			return
		}

		db.Preload("User").First(&debate, debate.ID)
		c.JSON(http.StatusOK, gin.H{"data": debate})
	}
}

func ReopenDebateTopic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var debate model.Debate

		if err := db.First(&debate, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		// Admin only
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}

		if err := db.Model(&debate).Updates(map[string]interface{}{
			"status":             "open",
			"concluded_at":       nil,
			"conclusion_type":    "",
			"conclusion_summary": "",
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reopen debate"})
			return
		}

		db.Preload("User").First(&debate, debate.ID)
		c.JSON(http.StatusOK, gin.H{"data": debate})
	}
}

func VoteToConclude(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var debate model.Debate

		if err := db.First(&debate, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		if debate.Status != "open" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Debate is not open"})
			return
		}

		userID, _ := c.Get("user_id")
		userUUID := userID.(uuid.UUID)

		// Check if user already voted to conclude
		var existing model.DebateConcludeVote
		if err := db.Where("debate_id = ? AND user_id = ?", debate.ID, userUUID).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Already voted to conclude"})
			return
		}

		vote := model.DebateConcludeVote{
			DebateID: debate.ID,
			UserID:   userUUID,
		}
		if err := db.Create(&vote).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
			return
		}

		newCount := debate.ConcludeVoteCount + 1
		db.Model(&debate).Update("conclude_vote_count", newCount)

		// Auto-conclude if threshold reached
		if newCount >= debate.ConcludeThreshold {
			now := time.Now()
			db.Model(&debate).Updates(map[string]interface{}{
				"status":          "concluded",
				"concluded_at":    now,
				"conclusion_type": "inconclusive",
			})
		}

		db.First(&debate, debate.ID)
		c.JSON(http.StatusOK, gin.H{
			"conclude_vote_count": debate.ConcludeVoteCount,
			"conclude_threshold":  debate.ConcludeThreshold,
			"auto_concluded":      debate.Status == "concluded",
		})
	}
}

func SearchDebateTopics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if limit < 1 || limit > 50 {
			limit = 10
		}

		var debates []model.Debate
		query := db.Model(&model.Debate{}).Preload("User")
		if q != "" {
			query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+q+"%", "%"+q+"%")
		}
		if err := query.Order("created_at DESC").Limit(limit).Find(&debates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search debates"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": debates})
	}
}

// ====== Argument Handlers ======

func GetDebateArguments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		debateID := c.Param("id")

		var arguments []model.Argument
		if err := db.Where("debate_id = ?", debateID).
			Preload("User").
			Preload("References").
			Preload("ReferencedDebates").
			Order("created_at ASC").
			Find(&arguments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch arguments"})
			return
		}

		// Inject user votes if authenticated
		userVotes := make(map[string]int)
		if userID, exists := c.Get("user_id"); exists {
			var votes []model.DebateVote
			db.Where("user_id = ? AND argument_id IN (?)",
				userID,
				db.Model(&model.Argument{}).Select("id").Where("debate_id = ?", debateID),
			).Find(&votes)
			for _, v := range votes {
				userVotes[v.ArgumentID.String()] = v.VoteType
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": arguments, "user_votes": userVotes})
	}
}

type CreateArgumentInput struct {
	Content      string             `json:"content" binding:"required"`
	ArgumentType model.ArgumentType `json:"argument_type" binding:"required"`
	ParentID     *uuid.UUID         `json:"parent_id"`
}

func CreateArgument(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		debateID := c.Param("id")
		var input CreateArgumentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify debate exists and is open
		var debate model.Debate
		if err := db.Where("id = ?", debateID).First(&debate).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}
		if debate.Status != "open" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Debate is closed"})
			return
		}

		userID, _ := c.Get("user_id")
		if input.ParentID != nil {
			var quoted model.Argument
			if err := db.Select("id").First(&quoted, "id = ? AND debate_id = ?", *input.ParentID, debate.ID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Quoted argument not found"})
				return
			}
		}

		argument := model.Argument{
			DebateID:     debate.ID,
			ParentID:     input.ParentID,
			UserID:       userID.(uuid.UUID),
			Content:      input.Content,
			ArgumentType: input.ArgumentType,
		}

		if err := db.Create(&argument).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create argument"})
			return
		}

		// Update debate argument count
		db.Model(&debate).Update("argument_count", debate.ArgumentCount+1)

		db.Preload("User").Preload("References").Where("id = ?", argument.ID).First(&argument)
		c.JSON(http.StatusCreated, gin.H{"data": argument})
	}
}

func UpdateArgument(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var argument model.Argument

		if err := db.First(&argument, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		// Check ownership
		userID, _ := c.Get("user_id")
		if argument.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		var input CreateArgumentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := map[string]interface{}{
			"content":       input.Content,
			"argument_type": input.ArgumentType,
		}

		if err := db.Model(&argument).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update argument"})
			return
		}

		db.Preload("User").Preload("References").First(&argument, argument.ID)
		c.JSON(http.StatusOK, gin.H{"data": argument})
	}
}

func DeleteArgument(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var argument model.Argument

		if err := db.First(&argument, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		// Check ownership or admin
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if argument.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		db.Model(&model.Argument{}).Where("parent_id = ?", argument.ID).Update("parent_id", nil)
		db.Where("argument_id = ?", argument.ID).Delete(&model.DebateVote{})
		_ = db.Model(&argument).Association("References").Clear()
		_ = db.Model(&argument).Association("ReferencedDebates").Clear()
		db.Delete(&argument)
		db.Model(&model.Debate{}).Where("id = ?", argument.DebateID).
			UpdateColumn("argument_count", gorm.Expr("CASE WHEN argument_count > 0 THEN argument_count - 1 ELSE 0 END"))

		c.JSON(http.StatusOK, gin.H{"message": "Argument deleted"})
	}
}

// ====== Reference Handlers ======

type ReferenceInput struct {
	ReferenceID uuid.UUID `json:"reference_id" binding:"required"`
}

func AddArgumentReference(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		argumentID := c.Param("id")
		var input ReferenceInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var argument model.Argument
		if err := db.First(&argument, "id = ?", argumentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		var refArgument model.Argument
		if err := db.First(&refArgument, "id = ?", input.ReferenceID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reference argument not found"})
			return
		}

		if err := db.Model(&argument).Association("References").Append(&refArgument); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reference"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Reference added"})
	}
}

func RemoveArgumentReference(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		argumentID := c.Param("id")
		refID := c.Param("ref_id")

		var argument model.Argument
		if err := db.First(&argument, "id = ?", argumentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		var refArgument model.Argument
		if err := db.First(&refArgument, "id = ?", refID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reference not found"})
			return
		}

		if err := db.Model(&argument).Association("References").Delete(&refArgument); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reference"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Reference removed"})
	}
}

// ====== Debate Reference Handlers ======

type DebateReferenceInput struct {
	DebateID uuid.UUID `json:"debate_id" binding:"required"`
}

func AddDebateReference(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		argumentID := c.Param("id")
		var input DebateReferenceInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var argument model.Argument
		if err := db.First(&argument, "id = ?", argumentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		var debate model.Debate
		if err := db.First(&debate, "id = ?", input.DebateID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		if err := db.Model(&argument).Association("ReferencedDebates").Append(&debate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add debate reference"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Debate reference added"})
	}
}

func RemoveDebateReference(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		argumentID := c.Param("id")
		debateID := c.Param("debate_id")

		var argument model.Argument
		if err := db.First(&argument, "id = ?", argumentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		var debate model.Debate
		if err := db.First(&debate, "id = ?", debateID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Debate not found"})
			return
		}

		if err := db.Model(&argument).Association("ReferencedDebates").Delete(&debate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove debate reference"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Debate reference removed"})
	}
}

// ====== Voting Handlers ======

type VoteInput struct {
	VoteType int `json:"vote_type" binding:"required,oneof=1 -1"`
}

func VoteArgument(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		argumentID := c.Param("id")
		var input VoteInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var argument model.Argument
		if err := db.First(&argument, "id = ?", argumentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Argument not found"})
			return
		}

		userID, _ := c.Get("user_id")
		userUUID := userID.(uuid.UUID)

		// Check if user already voted
		var existingVote model.DebateVote
		result := db.Where("argument_id = ? AND user_id = ?", argumentID, userUUID).First(&existingVote)

		if result.Error == nil {
			// Update existing vote
			if existingVote.VoteType == input.VoteType {
				// Same vote, remove it
				db.Delete(&existingVote)
				db.Model(&argument).Update("vote_count", argument.VoteCount-existingVote.VoteType)
			} else {
				// Different vote, update and record history
				oldVoteType := existingVote.VoteType
				db.Model(&existingVote).Update("vote_type", input.VoteType)
				db.Model(&argument).Update("vote_count", argument.VoteCount-oldVoteType+input.VoteType)

				// Record history
				history := model.VoteHistory{
					ArgumentID:  argument.ID,
					UserID:      userUUID,
					OldVoteType: oldVoteType,
					NewVoteType: input.VoteType,
				}
				db.Create(&history)
			}
		} else {
			// Create new vote
			vote := model.DebateVote{
				ArgumentID: argument.ID,
				UserID:     userUUID,
				VoteType:   input.VoteType,
			}
			db.Create(&vote)
			db.Model(&argument).Update("vote_count", argument.VoteCount+input.VoteType)
		}

		db.Preload("User").First(&argument, argument.ID)
		c.JSON(http.StatusOK, gin.H{"data": argument})
	}
}

func RemoveVote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		argumentID := c.Param("id")
		userID, _ := c.Get("user_id")
		userUUID := userID.(uuid.UUID)

		var vote model.DebateVote
		if err := db.Where("argument_id = ? AND user_id = ?", argumentID, userUUID).First(&vote).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vote not found"})
			return
		}

		var argument model.Argument
		db.First(&argument, argumentID)

		db.Delete(&vote)
		db.Model(&argument).Update("vote_count", argument.VoteCount-vote.VoteType)

		c.JSON(http.StatusOK, gin.H{"message": "Vote removed"})
	}
}

func GetArgumentVotes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Admin only
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}

		argumentID := c.Param("id")
		var votes []model.DebateVote
		if err := db.Where("argument_id = ?", argumentID).
			Preload("User").
			Order("created_at DESC").
			Find(&votes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch votes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": votes})
	}
}
