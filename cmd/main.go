package main

import (
	"github.com/artsgoz/artsgoz-backend/internal/app"
	"log"
	"os"
)

func main() {
	if err := app.Run(); err != nil {
		log.Printf("application stopped: %v", err)
		os.Exit(1)
	}
}
