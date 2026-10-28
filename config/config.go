package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func Load() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("config: load env: %w", err)
	}
	return nil
}
