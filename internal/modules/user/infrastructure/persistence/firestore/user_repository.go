package firestore

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userRepository struct {
	db *firestore.Client
}

func NewUserRepository(db *firestore.Client) domain.UserRepository {
	return &userRepository{db}
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("กรุณาระบุ email")
	}

	ctx := context.Background()
	iter := r.db.Collection("users").Where("email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, errors.New("ไม่พบผู้ใช้")
	}
	if err != nil {
		return nil, err
	}

	var user domain.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUID — ดึงข้อมูล user จาก Firestore โดยใช้ Firebase UID (document ID)
func (r *userRepository) GetByUID(uid string) (*domain.User, error) {
	if strings.TrimSpace(uid) == "" {
		return nil, errors.New("กรุณาระบุ uid")
	}

	ctx := context.Background()
	doc, err := r.db.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errors.New("ไม่พบผู้ใช้")
		}
		return nil, err
	}

	var user domain.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(user *domain.User) error {
	if user == nil || strings.TrimSpace(user.FirebaseUID) == "" {
		return errors.New("ข้อมูลผู้ใช้ไม่ถูกต้อง")
	}

	ctx := context.Background()
	_, err := r.db.Collection("users").Doc(user.FirebaseUID).Set(ctx, user)
	return err
}

func (r *userRepository) GetAllUsers() ([]*domain.User, error) {
	ctx := context.Background()
	iter := r.db.Collection("users").Documents(ctx)
	defer iter.Stop()

	users := make([]*domain.User, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var user domain.User
		if err := doc.DataTo(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func (r *userRepository) UpdateRole(uid string, role string) error {
	if strings.TrimSpace(uid) == "" || strings.TrimSpace(role) == "" {
		return errors.New("กรุณาระบุ uid และ role")
	}

	ctx := context.Background()
	_, err := r.db.Collection("users").Doc(uid).Update(ctx, []firestore.Update{
		{Path: "role", Value: role},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}
