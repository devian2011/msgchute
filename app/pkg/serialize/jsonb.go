package serialize

import (
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
)

type JSONB[T any] struct {
	Data T
}

// Value реализует driver.Valuer для автоматического маршалинга в JSON.
func (j JSONB[T]) Value() (driver.Value, error) {
	return sonic.Marshal(j.Data)
}

func (j *JSONB[T]) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("JSONB Scan: type assertion to []byte failed, got %T", value)
	}

	return sonic.Unmarshal(bytes, &j.Data)
}
