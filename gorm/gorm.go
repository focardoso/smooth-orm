package smoothgorm

import (
	"context"
	"errors"
	"fmt"

	"github.com/focardoso/smooth-orm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormEngine struct {
	DB     *gorm.DB
	Error  error
	Status bool
}

type txKeyType struct{}

var txKey = txKeyType{}

func Open(config Config) smooth.Engine {
	var eng GormEngine = GormEngine{
		DB:     nil,
		Error:  nil,
		Status: false,
	}
	var dialector gorm.Dialector
	switch config.Driver {
	case "postgres":
		dialector = postgres.Open(fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			config.Host, config.User, config.Password, config.DataBase, config.Port,
		))
	default:
		eng.Error = errDriverNotSupported
	}

	if eng.Error == nil {
		db, err := gorm.Open(dialector, &gorm.Config{
			Logger: customLogger{logger.Default.LogMode(logger.Error)},
		})
		if err != nil {
			eng.Error = err
		} else {
			eng.DB = db
			eng.Status = true
		}
	}
	return &eng
}

func (e *GormEngine) First(ctx context.Context, i interface{}, query smooth.Query) error {
	db := e.QueryConstructor(query, nil)
	result := db.First(i)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return smooth.ErrRecordNotFound
		} else {
			return result.Error
		}
	}
	return nil
}

func (e *GormEngine) Get(ctx context.Context, i interface{}, query smooth.Query) error {
	db := e.QueryConstructor(query, nil)
	result := db.Find(i)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return smooth.ErrRecordNotFound
		} else {
			return result.Error
		}
	}
	return nil
}

func (e *GormEngine) Create(ctx context.Context, i interface{}) error {
	value := ctx.Value(txKey)
	var result *gorm.DB
	if value != nil {
		tx, ok := value.(*gorm.DB)
		if !ok {
			return errors.New("failed to get transaction from context")
		}
		result = tx.Create(i)
	} else {
		result = e.DB.Create(i)
	}

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (e *GormEngine) Update(ctx context.Context, i interface{}) error {
	value := ctx.Value(txKey)
	var result *gorm.DB
	if value != nil {
		tx, ok := value.(*gorm.DB)
		if !ok {
			return errors.New("failed to get transaction from context")
		}
		result = tx.Save(i)
	} else {
		result = e.DB.Save(i)
	}

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (e *GormEngine) Delete(ctx context.Context, i interface{}) error {
	value := ctx.Value(txKey)
	var result *gorm.DB
	if value != nil {
		tx, ok := value.(*gorm.DB)
		if !ok {
			return errors.New("failed to get transaction from context")
		}
		result = tx.Delete(i)
	} else {
		result = e.DB.Delete(i)
	}

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (e *GormEngine) Raw(ctx context.Context, i interface{}, query smooth.Query) error {
	var gDB *gorm.DB
	value := ctx.Value(txKey)
	if value != nil {
		tx, ok := value.(*gorm.DB)
		if !ok {
			return errors.New("failed to get transaction from context")
		}
		gDB = tx
	} else {
		gDB = e.DB
	}

	gormDB := e.QueryConstructor(query, gDB)
	result := gormDB.Scan(i)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return result.Error
		}
	}
	return nil
}

func (e *GormEngine) Health() (bool, error) {
	if !e.Status {
		return e.Status, e.Error
	}

	db, err := e.DB.DB()
	if err != nil {
		return false, err
	}

	err = db.Ping()
	if err != nil {
		return false, err
	}

	return e.Status, err

}

func (e *GormEngine) Exec(ctx context.Context, sql string) error {
	return e.DB.Exec(sql).Error
}

func (e *GormEngine) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return e.DB.Transaction(func(tx *gorm.DB) error {
		// Cria um novo contexto com o transaction do Gorm
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}
