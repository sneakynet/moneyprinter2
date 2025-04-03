package db

import (
	"gorm.io/gorm"
)

// DB binds the convenience methods that build on top of Gorm to fetch
// records and do things.
type DB struct {
	d *gorm.DB
}
