package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxNameLength     = 50
	maxUsernameLength = 50
)

type Role string

const (
	RoleUser  Role = "User"
	RoleAdmin Role = "Admin"
)

var (
	ErrInvalidID        = errors.New("El identificador del usuario es obligatorio")
	ErrInvalidName      = errors.New("El nombre es obligatorio y debe tener como máximo 50 caracteres")
	ErrInvalidUsername  = errors.New("El nombre de usuario es obligatorio y debe tener como máximo 50 caracteres")
	ErrInvalidPassword  = errors.New("La contraseña es obligatoria")
	ErrInvalidRole      = errors.New("El rol del usuario no es válido")
	ErrInvalidUpdate    = errors.New("Debe actualizar al menos un campo del usuario")
	ErrInvalidUpdatedBy = errors.New("El usuario que realiza la actualización es obligatorio")
)

type User struct {
	ID        uuid.UUID
	Name      string
	Username  string
	Password  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

func NewUser(
	id uuid.UUID,
	name string,
	username string,
	passwordHash string,
	role Role,
	createdBy string,
	now time.Time,
) (*User, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidID
	}

	name, err := validateName(name)
	if err != nil {
		return nil, err
	}

	username, err = validateUsername(username)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(passwordHash) == "" {
		return nil, ErrInvalidPassword
	}

	if role == "" {
		role = RoleUser
	}
	if !isValidRole(role) {
		return nil, ErrInvalidRole
	}

	return &User{
		ID:        id,
		Name:      name,
		Username:  username,
		Password:  passwordHash,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
	}, nil
}

func (u *User) UpdateOwnProfile(name *string, passwordHash *string, updatedBy string, now time.Time) error {
	if u == nil || u.ID == uuid.Nil {
		return ErrInvalidID
	}
	if strings.TrimSpace(updatedBy) == "" {
		return ErrInvalidUpdatedBy
	}
	if name == nil && passwordHash == nil {
		return ErrInvalidUpdate
	}

	if name != nil {
		validatedName, err := validateName(*name)
		if err != nil {
			return err
		}
		u.Name = validatedName
	}

	if passwordHash != nil {
		if strings.TrimSpace(*passwordHash) == "" {
			return ErrInvalidPassword
		}
		u.Password = *passwordHash
	}

	u.UpdatedAt = now
	u.UpdatedBy = updatedBy
	return nil
}

func (u *User) UpdateByAdmin(
	name *string,
	username *string,
	passwordHash *string,
	role *Role,
	updatedBy string,
	now time.Time,
) error {
	if u == nil || u.ID == uuid.Nil {
		return ErrInvalidID
	}
	if strings.TrimSpace(updatedBy) == "" {
		return ErrInvalidUpdatedBy
	}
	if name == nil && username == nil && passwordHash == nil && role == nil {
		return ErrInvalidUpdate
	}

	validatedName := u.Name
	if name != nil {
		var err error
		validatedName, err = validateName(*name)
		if err != nil {
			return err
		}
	}

	validatedUsername := u.Username
	if username != nil {
		var err error
		validatedUsername, err = validateUsername(*username)
		if err != nil {
			return err
		}
	}

	validatedPassword := u.Password
	if passwordHash != nil {
		if strings.TrimSpace(*passwordHash) == "" {
			return ErrInvalidPassword
		}
		validatedPassword = *passwordHash
	}

	validatedRole := u.Role
	if role != nil {
		if !isValidRole(*role) {
			return ErrInvalidRole
		}
		validatedRole = *role
	}

	u.Name = validatedName
	u.Username = validatedUsername
	u.Password = validatedPassword
	u.Role = validatedRole
	u.UpdatedAt = now
	u.UpdatedBy = updatedBy
	return nil
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > maxNameLength {
		return "", ErrInvalidName
	}
	return name, nil
}

func validateUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || len([]rune(username)) > maxUsernameLength {
		return "", ErrInvalidUsername
	}
	return username, nil
}

func isValidRole(role Role) bool {
	return role == RoleUser || role == RoleAdmin
}
