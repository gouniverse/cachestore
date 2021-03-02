package cachestore

import (
	"time"

	"github.com/gouniverse/uid"
	"gorm.io/gorm"
)

// Cache type
type Cache struct {
	ID        string     `gorm:"type:varchar(40);column:id;primary_key;"`
	Key       string     `gorm:"type:varchar(255);column:cache_key;"`
	Value     string     `gorm:"type:longtext;column:cache_value;"`
	ExpiresAt *time.Time `gorm:"type:datetime;column:expires_at;DEFAULT NULL;"`
	CreatedAt time.Time  `gorm:"type:datetime;column:created_at;DEFAULT NULL;"`
	UpdatedAt time.Time  `gorm:"type:datetime;column:updated_at;DEFAULT NULL;"`
	DeletedAt *time.Time `gorm:"type:datetime;olumn:deleted_at;DEFAULT NULL;"`
}

// BeforeCreate adds UID to model
func (c *Cache) BeforeCreate(tx *gorm.DB) (err error) {
	uuid := uid.HumanUid()
	c.ID = uuid
	return nil
}
