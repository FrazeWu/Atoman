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

// parseDateTime 尝试多种格式解析时间，支持精确到小时分钟
func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Value: s}
}

func SetupTimelineRoutes(router *gin.Engine, db *gorm.DB) {
	tl := router.Group("/api/timeline")
	{
		// Public routes
		tl.GET("/events", GetTimelineEvents(db))
		tl.GET("/events/:id", GetTimelineEvent(db))
		tl.GET("/persons", GetTimelinePersons(db))
		tl.GET("/persons/:id", GetTimelinePerson(db))
		tl.GET("/persons/:id/locations", GetPersonLocations(db))

		// Protected routes
		protected := tl.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/events", CreateTimelineEvent(db))
			protected.PUT("/events/:id", UpdateTimelineEvent(db))
			protected.DELETE("/events/:id", DeleteTimelineEvent(db))

			protected.POST("/persons", CreateTimelinePerson(db))
			protected.PUT("/persons/:id", UpdateTimelinePerson(db))
			protected.DELETE("/persons/:id", DeleteTimelinePerson(db))

			protected.POST("/persons/:id/locations", AddPersonLocation(db))
			protected.PUT("/locations/:id", UpdatePersonLocation(db))
			protected.DELETE("/locations/:id", DeletePersonLocation(db))
		}
	}
}

// ====== Event Handlers ======

func GetTimelineEvents(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 200 {
			limit = 50
		}
		offset := (page - 1) * limit

		category := c.Query("category")
		yearStart := c.Query("year_start")
		yearEnd := c.Query("year_end")

		query := db.Model(&model.TimelineEvent{}).Preload("User").Where("is_public = ?", true)

		if category != "" {
			query = query.Where("category = ?", category)
		}
		if yearStart != "" {
			if y, err := strconv.Atoi(yearStart); err == nil {
				query = query.Where("event_date >= ?", time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC))
			}
		}
		if yearEnd != "" {
			if y, err := strconv.Atoi(yearEnd); err == nil {
				query = query.Where("event_date <= ?", time.Date(y, 12, 31, 23, 59, 59, 0, time.UTC))
			}
		}

		var total int64
		query.Count(&total)

		var events []model.TimelineEvent
		if err := query.Order("event_date ASC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  events,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func GetTimelineEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var event model.TimelineEvent
		if err := db.Preload("User").First(&event, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": event})
	}
}

type CreateEventInput struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	EventDate   string   `json:"event_date" binding:"required"`
	EndDate     string   `json:"end_date"`
	Location    string   `json:"location" binding:"required"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Source      string   `json:"source" binding:"required"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	IsPublic    *bool    `json:"is_public"`
}

func CreateTimelineEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input CreateEventInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		eventDate, err := parseDateTime(input.EventDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_date format"})
			return
		}

		isPublic := true
		if input.IsPublic != nil {
			isPublic = *input.IsPublic
		}

		userID, _ := c.Get("user_id")
		event := model.TimelineEvent{
			UserID:      userID.(uuid.UUID),
			Title:       input.Title,
			Description: input.Description,
			Content:     input.Content,
			EventDate:   eventDate,
			Location:    input.Location,
			Latitude:    input.Latitude,
			Longitude:   input.Longitude,
			Source:      input.Source,
			Category:    input.Category,
			Tags:        input.Tags,
			IsPublic:    isPublic,
		}

		if input.EndDate != "" {
			if endDate, err := parseDateTime(input.EndDate); err == nil {
				event.EndDate = &endDate
			}
		}

		if err := db.Create(&event).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
			return
		}

		db.Preload("User").First(&event, event.ID)
		c.JSON(http.StatusCreated, gin.H{"data": event})
	}
}

func UpdateTimelineEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var event model.TimelineEvent

		if err := db.First(&event, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if event.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		var input CreateEventInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		eventDate, err := parseDateTime(input.EventDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event_date format"})
			return
		}

		updates := map[string]interface{}{
			"title":       input.Title,
			"description": input.Description,
			"content":     input.Content,
			"event_date":  eventDate,
			"location":    input.Location,
			"latitude":    input.Latitude,
			"longitude":   input.Longitude,
			"source":      input.Source,
			"category":    input.Category,
			"tags":        input.Tags,
		}

		if input.IsPublic != nil {
			updates["is_public"] = *input.IsPublic
		}

		if input.EndDate != "" {
			if endDate, err := parseDateTime(input.EndDate); err == nil {
				updates["end_date"] = endDate
			}
		} else {
			updates["end_date"] = nil
		}

		if err := db.Model(&event).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
			return
		}

		db.Preload("User").First(&event, event.ID)
		c.JSON(http.StatusOK, gin.H{"data": event})
	}
}

func DeleteTimelineEvent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var event model.TimelineEvent

		if err := db.First(&event, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if event.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		db.Delete(&event)
		c.JSON(http.StatusOK, gin.H{"message": "Event deleted"})
	}
}

// ====== Person Handlers ======

func GetTimelinePersons(db *gorm.DB) gin.HandlerFunc {
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

		search := c.Query("search")
		query := db.Model(&model.TimelinePerson{}).Preload("User").Where("is_public = ?", true)

		if search != "" {
			query = query.Where("name ILIKE ?", "%"+search+"%")
		}

		var total int64
		query.Count(&total)

		var persons []model.TimelinePerson
		if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&persons).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch persons"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  persons,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func GetTimelinePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var person model.TimelinePerson

		if err := db.Preload("User").First(&person, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		var locations []model.PersonLocation
		db.Where("person_id = ?", person.ID).Order("date ASC").Find(&locations)
		person.Locations = locations

		c.JSON(http.StatusOK, gin.H{"data": person})
	}
}

func GetPersonLocations(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var person model.TimelinePerson
		if err := db.First(&person, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		var locations []model.PersonLocation
		if err := db.Where("person_id = ?", id).Order("date ASC").Find(&locations).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": locations})
	}
}

type CreatePersonInput struct {
	Name      string   `json:"name" binding:"required"`
	Bio       string   `json:"bio"`
	BirthDate string   `json:"birth_date"`
	DeathDate string   `json:"death_date"`
	Tags      []string `json:"tags"`
	IsPublic  *bool    `json:"is_public"`
}

func CreateTimelinePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input CreatePersonInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		isPublic := true
		if input.IsPublic != nil {
			isPublic = *input.IsPublic
		}

		userID, _ := c.Get("user_id")
		person := model.TimelinePerson{
			UserID:   userID.(uuid.UUID),
			Name:     input.Name,
			Bio:      input.Bio,
			Tags:     input.Tags,
			IsPublic: isPublic,
		}

		if input.BirthDate != "" {
			if d, err := parseDateTime(input.BirthDate); err == nil {
				person.BirthDate = &d
			}
		}
		if input.DeathDate != "" {
			if d, err := parseDateTime(input.DeathDate); err == nil {
				person.DeathDate = &d
			}
		}

		if err := db.Create(&person).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create person"})
			return
		}

		db.Preload("User").First(&person, person.ID)
		c.JSON(http.StatusCreated, gin.H{"data": person})
	}
}

func UpdateTimelinePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var person model.TimelinePerson

		if err := db.First(&person, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if person.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		var input CreatePersonInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := map[string]interface{}{
			"name": input.Name,
			"bio":  input.Bio,
			"tags": input.Tags,
		}

		if input.IsPublic != nil {
			updates["is_public"] = *input.IsPublic
		}

		if input.BirthDate != "" {
			if d, err := parseDateTime(input.BirthDate); err == nil {
				updates["birth_date"] = d
			}
		} else {
			updates["birth_date"] = nil
		}

		if input.DeathDate != "" {
			if d, err := parseDateTime(input.DeathDate); err == nil {
				updates["death_date"] = d
			}
		} else {
			updates["death_date"] = nil
		}

		if err := db.Model(&person).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update person"})
			return
		}

		db.Preload("User").First(&person, person.ID)
		c.JSON(http.StatusOK, gin.H{"data": person})
	}
}

func DeleteTimelinePerson(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var person model.TimelinePerson

		if err := db.First(&person, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if person.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		db.Where("person_id = ?", id).Delete(&model.PersonLocation{})
		db.Delete(&person)
		c.JSON(http.StatusOK, gin.H{"message": "Person deleted"})
	}
}

// ====== Location Handlers ======

type CreateLocationInput struct {
	Date      string  `json:"date" binding:"required"`
	EndDate   string  `json:"end_date"`
	PlaceName string  `json:"place_name" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Source    string  `json:"source" binding:"required"`
	Note      string  `json:"note"`
}

func AddPersonLocation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		personID := c.Param("id")
		var person model.TimelinePerson

		if err := db.First(&person, "id = ?", personID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if person.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		var input CreateLocationInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		date, err := parseDateTime(input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		pid, _ := uuid.Parse(personID)
		location := model.PersonLocation{
			PersonID:  pid,
			Date:      date,
			PlaceName: input.PlaceName,
			Latitude:  input.Latitude,
			Longitude: input.Longitude,
			Source:    input.Source,
			Note:      input.Note,
		}

		if input.EndDate != "" {
			if endDate, err := parseDateTime(input.EndDate); err == nil {
				location.EndDate = &endDate
			}
		}

		if err := db.Create(&location).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create location"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": location})
	}
}

func UpdatePersonLocation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var location model.PersonLocation

		if err := db.First(&location, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
			return
		}

		var person model.TimelinePerson
		db.First(&person, "id = ?", location.PersonID)

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if person.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		var input CreateLocationInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		date, err := parseDateTime(input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		updates := map[string]interface{}{
			"date":       date,
			"place_name": input.PlaceName,
			"latitude":   input.Latitude,
			"longitude":  input.Longitude,
			"source":     input.Source,
			"note":       input.Note,
		}

		if input.EndDate != "" {
			if endDate, err := parseDateTime(input.EndDate); err == nil {
				updates["end_date"] = endDate
			}
		} else {
			updates["end_date"] = nil
		}

		if err := db.Model(&location).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": location})
	}
}

func DeletePersonLocation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var location model.PersonLocation

		if err := db.First(&location, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
			return
		}

		var person model.TimelinePerson
		db.First(&person, "id = ?", location.PersonID)

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if person.UserID != userID.(uuid.UUID) && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}

		db.Delete(&location)
		c.JSON(http.StatusOK, gin.H{"message": "Location deleted"})
	}
}
