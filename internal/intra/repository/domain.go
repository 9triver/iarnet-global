package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type DomainDAO struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type DomainRepo interface {
	CreateDomain(ctx context.Context, dao *DomainDAO) error
	UpdateDomain(ctx context.Context, dao *DomainDAO) error
	DeleteDomain(ctx context.Context, id string) error
	GetDomain(ctx context.Context, id string) (*DomainDAO, error)
	GetAllDomains(ctx context.Context) ([]*DomainDAO, error)
	Close() error
}

func NewDomainRepo(dbPath string, maxOpenConns int, maxIdleConns int, connMaxLifetimeSeconds int) (DomainRepo, error) {

	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	if connMaxLifetimeSeconds > 0 {
		db.SetConnMaxLifetime(time.Duration(connMaxLifetimeSeconds) * time.Second)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &domainRepoSQLite{
		db: db,
	}

	// 初始化表结构
	if err := repo.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logrus.Infof("Domain repository initialized with SQLite at %s", dbPath)
	return repo, nil
}

type domainRepoSQLite struct {
	db *sql.DB
}

// initSchema 初始化数据库表结构
func (r *domainRepoSQLite) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS domains (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);
	CREATE INDEX IF NOT EXISTS idx_domains_created_at ON domains(created_at);
	`

	if _, err := r.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (r *domainRepoSQLite) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func (r *domainRepoSQLite) CreateDomain(ctx context.Context, dao *DomainDAO) error {
	return nil
}

func (r *domainRepoSQLite) UpdateDomain(ctx context.Context, dao *DomainDAO) error {
	return nil
}

func (r *domainRepoSQLite) DeleteDomain(ctx context.Context, id string) error {
	return nil
}

func (r *domainRepoSQLite) GetDomain(ctx context.Context, id string) (*DomainDAO, error) {
	return nil, nil
}

func (r *domainRepoSQLite) GetAllDomains(ctx context.Context) ([]*DomainDAO, error) {
	return nil, nil
}
