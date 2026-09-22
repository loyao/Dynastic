package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"dynastic/internal/config"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open 建立数据库连接。若目标库不存在会先创建，随后按配置自动执行迁移。
func Open(cfg *config.Config) (*sql.DB, error) {
	// 先连接到 server（不指定库），确保目标数据库存在。
	serverDB, err := sql.Open("mysql", cfg.ServerDSN())
	if err != nil {
		return nil, fmt.Errorf("打开 MySQL server 连接失败: %w", err)
	}
	defer serverDB.Close()

	if err := serverDB.Ping(); err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败（请确认服务已启动、账号密码正确）: %w", err)
	}

	if _, err := serverDB.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName,
	)); err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 Ping 失败: %w", err)
	}

	if cfg.AutoMigrate {
		if err := migrate(db); err != nil {
			return nil, err
		}
	}
	return db, nil
}

// migrate 依次执行内嵌的建表与种子 SQL。脚本本身是幂等的。
func migrate(db *sql.DB) error {
	files := []string{"migrations/001_schema.sql", "migrations/002_seed.sql"}
	for _, f := range files {
		data, err := migrationsFS.ReadFile(f)
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s 失败: %w", f, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("执行迁移 %s 失败: %w", f, err)
		}
		log.Printf("已执行迁移: %s", f)
	}
	return nil
}
