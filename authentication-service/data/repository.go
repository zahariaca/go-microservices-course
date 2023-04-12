package data

type Repository interface {
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	GetOne(id int) (*User, error)
	DeleteById(id int) error
	Insert(user User) (int, error)
	ResetPassword(password string, user User) error
	PasswordMatches(plainText string, user User) (bool, error)
}