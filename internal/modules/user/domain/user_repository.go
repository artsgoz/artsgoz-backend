package domain

type UserRepository interface {
	GetByEmail(email string) (*User, error)
	GetByUID(uid string) (*User, error)
	CreateUser(user *User) error
	GetAllUsers() ([]*User, error)
	UpdateRole(uid string, role string) error
}
