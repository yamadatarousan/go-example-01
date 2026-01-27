package config

import (
	"os"
)

// Config はアプリケーションの設定を保持
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
}

// DatabaseConfig はデータベース接続の設定
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	AutoMigrate bool
}

// ServerConfig はサーバーの設定
type ServerConfig struct {
	Port         string
	AllowOrigins []string
}

// JWTConfig はJWT認証の設定
type JWTConfig struct {
	Secret []byte
}

// Load は環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5435"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "todo_db"),
			AutoMigrate: getEnvBool("AUTO_MIGRATE", false),
		},
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			AllowOrigins: []string{getEnv("ALLOW_ORIGIN", "http://localhost:3000")},
		},
		JWT: JWTConfig{
			Secret: []byte(getEnv("JWT_SECRET", "a-very-secret-key")),
		},
	}
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool は環境変数を取得し、boolに変換する
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "TRUE"
}

// DSN はデータベース接続文字列を生成
func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=disable"
}

// MigrationURL はマイグレーション用の接続文字列を生成
func (c *DatabaseConfig) MigrationURL() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=disable"
}
