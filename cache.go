package cachestore

import (
	"time"
)

// Cache type
type Cache struct {
	ID        string     `db:"id"`
	Key       string     `db:"cache_key"`
	Value     string     `db:"cache_value"`
	ExpiresAt *time.Time `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// BeforeCreate adds UID to model
// func (c *Cache) BeforeCreate(tx *gorm.DB) (err error) {
// 	uuid := uid.HumanUid()
// 	c.ID = uuid
// 	return nil
// }
