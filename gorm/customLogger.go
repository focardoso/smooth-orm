package smoothgorm

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type customLogger struct {
	logger.Interface
}

func (l customLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Ignore ErrRecordNotFound
		return
	}
	l.Interface.Trace(ctx, begin, fc, err)
}
