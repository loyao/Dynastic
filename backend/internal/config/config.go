package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 保存后端运行所需的配置项，全部可通过环境变量覆盖。
type Config struct {
	// HTTP 服务监听地址，例如 :8080
	Addr string
	// 允许跨域访问的前端源，多个以逗号分隔
	CORSOrigins []string

	// MySQL 连接参数
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// 是否在启动时自动执行建表与种子数据迁移
	AutoMigrate bool
}

// Load 从环境变量读取配置，未设置时使用合理的默认值（面向本地开发）。
// 会先尝试加载当前目录或 backend 目录下的 .env 文件。
func Load() *Config {
	loadDotEnv(".env", "backend/.env")
	return &Config{
		Addr:        getEnv("SERVER_ADDR", ":8080"),
		CORSOrigins: splitAndTrim(getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnvInt("DB_PORT", 3306),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "dynastic"),

		AutoMigrate: getEnvBool("AUTO_MIGRATE", true),
	}
}

// DSN 返回 go-sql-driver/mysql 使用的数据源名称。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// ServerDSN 返回不含数据库名的连接串，用于首次创建数据库。
func (c *Config) ServerDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort)
}

// loadDotEnv 依次尝试读取给定路径的 .env 文件，把其中的键值写入环境变量。
// 已存在的环境变量优先级更高，不会被文件覆盖。找不到文件时静默跳过。
func loadDotEnv(paths ...string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, val, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(strings.Trim(val, `"'`))
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
		return // 只加载第一个找到的文件
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
