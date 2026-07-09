package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"zcyp-im/internal/config"
)

func New(cfg config.MySQLConfig) (*sql.DB, error) {
	log.Printf("mysql: opening connection host=%s port=%d database=%s user=%s", cfg.Host, cfg.Port, cfg.Database, cfg.Username)

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		cfg.ParseTime,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	log.Printf("mysql: ping ok host=%s port=%d database=%s", cfg.Host, cfg.Port, cfg.Database)

	return db, nil
}
