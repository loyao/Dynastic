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
		if err := migrate(db, cfg.DBName); err != nil {
			return nil, err
		}
	}
	return db, nil
}

// migrate 执行建表结构，补齐历史库缺失的列，并在数据为空时注入种子数据。
// 结构脚本幂等；种子数据仅在 dynasty 表为空时写入一次，避免覆盖用户后续的增改。
func migrate(db *sql.DB, dbName string) error {
	schema, err := migrationsFS.ReadFile("migrations/001_schema.sql")
	if err != nil {
		return fmt.Errorf("读取建表脚本失败: %w", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("执行建表脚本失败: %w", err)
	}
	log.Printf("已执行迁移: migrations/001_schema.sql")

	// 兼容早期已建库但缺少 lineage_id 列的情况。
	if err := ensureColumn(db, dbName, "emperor", "lineage_id",
		"BIGINT NULL COMMENT '世系上游帝王 id' AFTER father_id"); err != nil {
		return err
	}

	empty, err := isTableEmpty(db, "dynasty")
	if err != nil {
		return err
	}
	if !empty {
		log.Println("检测到已有数据，跳过种子数据注入（保留用户的修改）")
		return nil
	}

	seed, err := migrationsFS.ReadFile("migrations/002_seed.sql")
	if err != nil {
		return fmt.Errorf("读取种子脚本失败: %w", err)
	}
	if _, err := db.Exec(string(seed)); err != nil {
		return fmt.Errorf("执行种子脚本失败: %w", err)
	}
	log.Printf("已执行迁移: migrations/002_seed.sql（首次初始化种子数据）")
	return nil
}

// ensureColumn 在指定表缺少某列时补充该列，实现幂等的结构演进。
func ensureColumn(db *sql.DB, dbName, table, column, definition string) error {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ? AND column_name = ?`,
		dbName, table, column).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查列 %s.%s 失败: %w", table, column, err)
	}
	if count > 0 {
		return nil
	}
	stmt := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", table, column, definition)
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("补充列 %s.%s 失败: %w", table, column, err)
	}
	log.Printf("已为表 %s 补充列 %s", table, column)
	return nil
}

// isTableEmpty 判断表是否无数据。
func isTableEmpty(db *sql.DB, table string) (bool, error) {
	var n int
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)).Scan(&n); err != nil {
		return false, fmt.Errorf("统计表 %s 失败: %w", table, err)
	}
	return n == 0, nil
}
