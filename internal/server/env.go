package server

import (
	"github.com/joho/godotenv"
)

func LoadEnvs() {
	godotenv.Load("dns-server.env", "postgres.env")
}
