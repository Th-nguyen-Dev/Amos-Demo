package models

import "fmt"

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port        int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Host        string `mapstructure:"host"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development staging production"`
}

// DatabaseConfig represents PostgreSQL configuration
type DatabaseConfig struct {
	Host         string `mapstructure:"host" validate:"required"`
	Port         int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	User         string `mapstructure:"user" validate:"required"`
	Password     string `mapstructure:"password" validate:"required"`
	DBName       string `mapstructure:"dbname" validate:"required"`
	SSLMode      string `mapstructure:"sslmode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns int    `mapstructure:"max_open_conns" validate:"min=1"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" validate:"min=1"`
}

// ConnectionString builds PostgreSQL connection string
func (c DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

