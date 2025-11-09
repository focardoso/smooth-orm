package smoothgorm

import (
	"fmt"

	"github.com/focardoso/smooth-orm"
	"gorm.io/gorm"
)

func (e *GormEngine) QueryConstructor(query smooth.Query, d *gorm.DB) *gorm.DB {
	var db *gorm.DB
	if d == nil {
		db = e.DB
	} else {
		db = d
	}

	if query.Where != nil {
		for _, v := range *query.Where {
			db = Where(db, v)
		}
	}

	if query.With != nil {
		for _, p := range *query.With {
			db = db.Preload(p.Field)
		}
	}

	if query.InnerJoins != nil {
		for _, ij := range *query.InnerJoins {
			if ij.Where != nil {
				dbs := []*gorm.DB{}
				for _, ijw := range *ij.Where {
					dbs = append(dbs, Where(db, ijw))
				}
				db = db.InnerJoins(ij.Field, dbs)
			} else {
				db = db.InnerJoins(ij.Field)
			}
		}
	}

	if query.Raw != nil {
		db = db.Raw(query.Raw.Query, query.Raw.Interfaces...)
	}

	if query.Limit != nil {
		db = db.Limit(*query.Limit)
	}

	if query.Offset != nil {
		db = db.Offset(*query.Offset)
	}

	if query.Unscoped {
		db = db.Unscoped()
	}

	if query.Debug {
		db = db.Debug()
	}

	return db
}

func Where(db *gorm.DB, query smooth.Where) *gorm.DB {
	if query.Condition == "" {
		return db.Where(fmt.Sprint(query.Column, " = ?"), query.Value)
	} else {
		return db.Where(fmt.Sprint(query.Column, " ", query.Condition, " ?"), query.Value)
	}
}
