package community

import (
	"net/http"
	"strings"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/article"
	"github.com/artsgoz/artsgoz-backend/internal/comment"
	"github.com/artsgoz/artsgoz-backend/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FAQ struct {
	ID        string   `json:"id" gorm:"type:uuid;primaryKey"`
	Question  string   `json:"question" binding:"required"`
	Answer    string   `json:"answer" binding:"required"`
	Tags      []string `json:"tags" gorm:"serializer:json"`
	SortOrder int      `json:"sortOrder"`
}

func (FAQ) TableName() string { return "faqs" }

type Contact struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name" binding:"required"`
	StudentID string    `json:"studentId"`
	Email     string    `json:"email" binding:"required,email"`
	Category  string    `json:"category" binding:"required"`
	Message   string    `json:"message" binding:"required"`
	Status    string    `json:"status"`
	Reply     string    `json:"reply"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Contact) TableName() string { return "contact_submissions" }

type Review struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	Company   string    `json:"company" binding:"required"`
	Role      string    `json:"role" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	Author    string    `json:"author"`
	Date      string    `json:"date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Review) TableName() string { return "reviews" }

func failure(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "error": gin.H{"message": message}, "message": message})
}
func success(c *gin.Context, status int, value any) {
	c.JSON(status, gin.H{"success": true, "data": value})
}
func result(c *gin.Context, tx *gorm.DB) bool {
	if tx.Error != nil {
		failure(c, 500, "database operation failed")
		return false
	}
	if tx.RowsAffected == 0 {
		failure(c, 404, "record not found")
		return false
	}
	return true
}
func validID(c *gin.Context) bool {
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		failure(c, 400, "invalid ID")
		return false
	}
	return true
}

func Register(api *gin.RouterGroup, db *gorm.DB, requireAuth gin.HandlerFunc) {
	admin := api.Group("/admin", requireAuth, server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin, server.Role("admin")))
	api.GET("/faqs", func(c *gin.Context) {
		items := []FAQ{}
		query := db.WithContext(c.Request.Context()).Order("sort_order, id")
		if q := c.Query("q"); q != "" {
			query = query.Where("question ILIKE ? OR answer ILIKE ?", "%"+q+"%", "%"+q+"%")
		}
		if err := query.Find(&items).Error; err != nil {
			failure(c, 500, "failed to list FAQs")
			return
		}
		if tag := c.Query("tag"); tag != "" {
			filtered := []FAQ{}
			for _, item := range items {
				for _, t := range item.Tags {
					if t == tag {
						filtered = append(filtered, item)
						break
					}
				}
			}
			items = filtered
		}
		success(c, 200, items)
	})
	api.POST("/faqs", requireAuth, server.RequireRoles(server.RoleEditor, server.RoleSuperAdmin, server.Role("admin")), func(c *gin.Context) {
		var input FAQ
		if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Question) == "" || strings.TrimSpace(input.Answer) == "" {
			failure(c, 400, "question and answer are required")
			return
		}
		input.ID = uuid.NewString()
		if result(c, db.WithContext(c.Request.Context()).Create(&input)) {
			success(c, 201, gin.H{"id": input.ID})
		}
	})
	api.POST("/contacts", func(c *gin.Context) {
		var input Contact
		if c.ShouldBindJSON(&input) != nil || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Message) == "" {
			failure(c, 400, "name, email, category, and message are required")
			return
		}
		input.ID = uuid.NewString()
		input.Status = "pending"
		input.Reply = ""
		input.CreatedAt = time.Now().UTC()
		input.UpdatedAt = input.CreatedAt
		if result(c, db.WithContext(c.Request.Context()).Create(&input)) {
			success(c, 201, gin.H{"id": input.ID})
		}
	})
	admin.GET("/contacts", func(c *gin.Context) {
		items := []Contact{}
		if err := db.WithContext(c.Request.Context()).Order("created_at DESC").Find(&items).Error; err != nil {
			failure(c, 500, "failed to list contacts")
			return
		}
		success(c, 200, items)
	})
	admin.PUT("/contacts/:id", func(c *gin.Context) {
		if !validID(c) {
			return
		}
		var input struct {
			Status string `json:"status" binding:"required,oneof=pending replied resolved closed"`
			Reply  string `json:"reply"`
		}
		if c.ShouldBindJSON(&input) != nil {
			failure(c, 400, "invalid status")
			return
		}
		if result(c, db.WithContext(c.Request.Context()).Model(&Contact{}).Where("id = ?", c.Param("id")).Updates(map[string]any{"status": input.Status, "reply": input.Reply, "updated_at": time.Now().UTC()})) {
			success(c, 200, nil)
		}
	})
	api.POST("/reviews", requireAuth, func(c *gin.Context) {
		var input Review
		if c.ShouldBindJSON(&input) != nil {
			failure(c, 400, "company, role, and content are required")
			return
		}
		input.ID = uuid.NewString()
		input.Status = "pending"
		input.Author = c.GetString("username")
		input.Date = time.Now().UTC().Format("2006-01-02")
		input.CreatedAt = time.Now().UTC()
		input.UpdatedAt = input.CreatedAt
		if result(c, db.WithContext(c.Request.Context()).Create(&input)) {
			success(c, 201, gin.H{"id": input.ID})
		}
	})
	api.GET("/reviews", func(c *gin.Context) {
		items := []Review{}
		if err := db.WithContext(c.Request.Context()).Where("status = ?", "approved").Find(&items).Error; err != nil {
			failure(c, 500, "failed to list reviews")
			return
		}
		success(c, 200, items)
	})
	admin.GET("/reviews", func(c *gin.Context) {
		items := []Review{}
		if err := db.WithContext(c.Request.Context()).Order("created_at DESC").Find(&items).Error; err != nil {
			failure(c, 500, "failed to list reviews")
			return
		}
		success(c, 200, items)
	})
	admin.PUT("/reviews/:id/approve", func(c *gin.Context) {
		if !validID(c) {
			return
		}
		if result(c, db.WithContext(c.Request.Context()).Model(&Review{}).Where("id = ?", c.Param("id")).Update("status", "approved")) {
			success(c, 200, nil)
		}
	})
	admin.DELETE("/reviews/:id", func(c *gin.Context) {
		if !validID(c) {
			return
		}
		if result(c, db.WithContext(c.Request.Context()).Where("id = ?", c.Param("id")).Delete(&Review{})) {
			success(c, 200, nil)
		}
	})
	admin.GET("/tags", func(c *gin.Context) {
		items := []article.Tag{}
		query := db.WithContext(c.Request.Context())
		if q := c.Query("search"); q != "" {
			query = query.Where("name ILIKE ?", "%"+q+"%")
		}
		if err := query.Order("name").Find(&items).Error; err != nil {
			failure(c, 500, "failed to list tags")
			return
		}
		success(c, 200, items)
	})
	admin.GET("/comments", func(c *gin.Context) {
		items := []comment.Comment{}
		query := db.WithContext(c.Request.Context())
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}
		if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
			failure(c, 500, "failed to list comments")
			return
		}
		success(c, 200, items)
	})
	admin.PATCH("/comments/:id", func(c *gin.Context) {
		if !validID(c) {
			return
		}
		var input struct {
			Status string `json:"status" binding:"required,oneof=approved rejected spam pending"`
		}
		if c.ShouldBindJSON(&input) != nil {
			failure(c, 400, "invalid comment status")
			return
		}
		if result(c, db.WithContext(c.Request.Context()).Model(&comment.Comment{}).Where("id = ?", c.Param("id")).Update("status", input.Status)) {
			success(c, 200, nil)
		}
	})
	admin.DELETE("/comments/:id", func(c *gin.Context) {
		if !validID(c) {
			return
		}
		if result(c, db.WithContext(c.Request.Context()).Where("id = ?", c.Param("id")).Delete(&comment.Comment{})) {
			success(c, http.StatusOK, nil)
		}
	})
}
