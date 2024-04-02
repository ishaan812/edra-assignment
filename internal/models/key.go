package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Key struct {
	gorm.Model
	KeyID          string     `gorm:"primarykey" json:"key_id,omitempty"`
	LastAccessed   time.Time  `json:"last_accessed"`
	UserAccessTime *time.Time `json:"user_access_time"`
	Busy           bool       `json:"busy,omitempty" gorm:"default:false;index"`
}

func (k *Key) Init() {
	k.Busy = false
	k.KeyID = uuid.New().String()
	k.LastAccessed = time.Now()
	now := time.Now()
	k.UserAccessTime = &now
}
