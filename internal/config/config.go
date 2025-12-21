package config

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func NewConfig() *Config {
	return &Config{
		RunAddress:           ":8081",
		DatabaseURI:          "postgres://localhost/gophermart?sslmode=disable",
		AccrualSystemAddress: "http://localhost:8080",
	}
}

func (c *Config) Parse() {
	flag.StringVar(&c.RunAddress, "a", c.RunAddress, "Address and port to run the server on")
	flag.StringVar(&c.DatabaseURI, "d", c.DatabaseURI, "Database connection URI")
	flag.StringVar(&c.AccrualSystemAddress, "r", c.AccrualSystemAddress, "Address of the accrual system")

	flag.Parse()

	// перезатрем если указаны переменные окружения
	if addr := os.Getenv("RUN_ADDRESS"); addr != "" {
		c.RunAddress = addr
	}
	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		c.DatabaseURI = dbURI
	}
	if accrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddr != "" {
		c.AccrualSystemAddress = strings.TrimRight(accrualAddr, "/")
	}
}
