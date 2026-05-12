package validator

import (
	"fmt"
	"sync"

	v10 "github.com/go-playground/validator/v10"

	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
)

var (
	once     sync.Once
	instance *v10.Validate
)

func get() *v10.Validate {
	once.Do(func() {
		instance = v10.New(v10.WithRequiredStructEnabled())
	})
	return instance
}

// Struct validates v against its `validate:` tags. On failure it returns an
// *apperr.Error of kind validation with per-field details so the HTTP layer
// can render a structured 400.
func Struct(v any) error {
	if err := get().Struct(v); err != nil {
		if ve, ok := err.(v10.ValidationErrors); ok {
			details := make(map[string]string, len(ve))
			for _, fe := range ve {
				details[fe.Field()] = fmt.Sprintf("failed on '%s'", fe.Tag())
			}
			return apperr.Validation("validation failed", details)
		}
		return apperr.Validation(err.Error(), nil)
	}
	return nil
}
