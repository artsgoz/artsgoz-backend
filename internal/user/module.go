package user

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/artsgoz/artsgoz-backend/internal/user/controller"
	"github.com/artsgoz/artsgoz-backend/internal/user/repository"
	"github.com/artsgoz/artsgoz-backend/internal/user/usecase"
)

type Module struct {
	handler *controller.Handler
}

func New(pool *pgxpool.Pool) *Module {
	users := repository.NewPostgresUserRepository(pool)
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

func (m *Module) RegisterRoutes(router fiber.Router) {
	group := router.Group("/users")
	group.Post("/register", m.handler.RegisterUser)
}
