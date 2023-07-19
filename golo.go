package bolo

import (
	"os"

	"github.com/joho/godotenv"
)

func init() {
	initDotEnvConfigSupport()
}

// Init doc env config with default development.env configuration file
// The configuration file pattern is: [environment].env
func initDotEnvConfigSupport() {
	env, _ := os.LookupEnv("GO_ENV")

	if env == "" {
		env = "development"
	}

	if _, err := os.Stat(env + ".env"); err == nil {
		godotenv.Load(env + ".env")
	}
}
