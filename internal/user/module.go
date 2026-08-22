package user

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/artsgoz/artsgoz-backend/internal/user/controller"
	"github.com/artsgoz/artsgoz-backend/internal/user/repository"
	"github.com/artsgoz/artsgoz-backend/internal/user/usecase"
)

type Module struct {
	handler *controller.Handler
}

func New(database *gorm.DB) *Module {
	users := repository.NewPostgresUserRepository(database)
	hasher := usecase.PasswordHasherFunc(func(password string) (string, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		return string(hash), err
	})
	registerUser := usecase.NewRegisterUser(
		users,
		hasher,
		usecase.IDGeneratorFunc(uuid.NewString),
		usecase.ClockFunc(func() time.Time { return time.Now().UTC() }),
	)
	return &Module{handler: controller.NewHandler(registerUser)}
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/users")
	group.POST("/register", m.handler.RegisterUser)
}
