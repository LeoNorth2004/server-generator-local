package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email     string         `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      string         `gorm:"size:20;default:user" json:"role"`
	Status    string         `gorm:"size:20;default:active" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Project struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`
	User          *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	DBConfig      string         `gorm:"type:text" json:"db_config"`
	TableConfig   string         `gorm:"type:text" json:"table_config"`
	GeneratedCode string         `gorm:"type:text" json:"generated_code"`
	Status        string         `gorm:"size:20;default:pending" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type OperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Username   string    `gorm:"size:50" json:"username"`
	Action     string    `gorm:"size:50" json:"action"`
	Resource   string    `gorm:"size:50" json:"resource"`
	ResourceID uint      `json:"resource_id"`
	Details    string    `gorm:"type:text" json:"details"`
	Status     string    `gorm:"size:20" json:"status"`
	Error      string    `gorm:"type:text" json:"error"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	Duration   int64     `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}
