package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseAdapter interface {
 Connect(ctx context.Context, config *DatabaseConfig) (*gorm.DB, error)
 Disconnect() error
 Ping() error
 GetDialect() string
 IsConnected() bool
 HealthCheck() (*HealthStatus, error)
 ExecuteRaw(sql string, args ...interface{}) (interface{}, error)
 Migrate(models ...interface{}) error
 Transaction(fn func(tx *gorm.DB) error) error
}

type DatabaseConfig struct {
 Driver       string            `json:"driver"` // postgres, mysql, sqlite, mssql
 Host         string            `json:"host"`
 Port         int               `json:"port"`
 User         string            `json:"user"`
 Password     string            `json:"password"`
 DBName       string            `json:"db_name"`
 SSLMode      string            `json:"ssl_mode"`
 MaxOpenConns int               `json:"max_open_conns"` // default: 100
 MaxIdleConns int               `json:"max_idle_conns"` // default: 10
 MaxLifetime  time.Duration     `json:"max_lifetime"`  // default: 1h
 LoggerLevel logger.LogLevel   `json:"logger_level"`
 EnableLog   bool              `json:"enable_log"`
 Parameters   map[string]string `json:"parameters,omitempty"`
}

type HealthStatus struct {
 Status    string            `json:"status"` // healthy, degraded, unhealthy
 Latency   time.Duration     `json:"latency_ms"`
 OpenConns int               `json:"open_connections"`
 IdleConns int               `json:"idle_connections"`
 InUse     int               `json:"in_use"`
 WaitCount int64             `json:"wait_count"`
 WaitDuration time.Duration  `json:"wait_duration_ms"`
 Errors    []string          `json:"errors,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
}

type PostgreSQLAdapter struct {
 db  *gorm.DB
 cfg *DatabaseConfig
}

func NewPostgreSQLAdapter(cfg *DatabaseConfig) *PostgreSQLAdapter {
 return &PostgreSQLAdapter{cfg: cfg}
}

func (a *PostgreSQLAdapter) Connect(ctx context.Context, config *DatabaseConfig) (*gorm.DB, error) {
 a.cfg = config

 dsn := fmt.Sprintf(
  "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
  config.Host,
  config.Port,
  config.User,
  config.Password,
  config.DBName,
  config.SSLMode,
 )

 if len(config.Parameters) > 0 {
  for k, v := range config.Parameters {
   dsn += fmt.Sprintf(" %s=%s", k, v)
  }
 }

 gormConfig := &gorm.Config{
  Logger: logger.Default.LogMode(config.LoggerLevel),
 }

 if !config.EnableLog {
  gormConfig.Logger = logger.Default.LogMode(logger.Silent)
 }

 db, err := gorm.Open(postgres.Open(dsn), gormConfig)
 if err != nil {
  return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
 }

 sqlDB, err := db.DB()
 if err != nil {
  return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
 }

 maxOpen := config.MaxOpenConns
 if maxOpen <= 0 {
  maxOpen = 100
 }
 maxIdle := config.MaxIdleConns
 if maxIdle <= 0 {
  maxIdle = 10
 }
 lifetime := config.MaxLifetime
 if lifetime <= 0 {
  lifetime = time.Hour
 }

 sqlDB.SetMaxOpenConns(maxOpen)
 sqlDB.SetMaxIdleConns(maxIdle)
 sqlDB.SetConnMaxLifetime(lifetime)

 a.db = db

 log.Printf("[PostgreSQLAdapter] Connected to %s:%d/%s", config.Host, config.Port, config.DBName)

 return db, nil
}

func (a *PostgreSQLAdapter) Disconnect() error {
 if a.db == nil {
  return nil
 }
 sqlDB, err := a.db.DB()
 if err != nil {
  return err
 }
 return sqlDB.Close()
}

func (a *PostgreSQLAdapter) Ping() error {
 if a.db == nil {
  return fmt.Errorf("database not connected")
 }
 sqlDB, err := a.db.DB()
 if err != nil {
  return err
 }
 return sqlDB.Ping()
}

func (a *PostgreSQLAdapter) GetDialect() string { return "postgres" }
func (a *PostgreSQLAdapter) IsConnected() bool { return a.db != nil }

func (a *PostgreSQLAdapter) HealthCheck() (*HealthStatus, error) {
 status := &HealthStatus{
  Status:    "healthy",
  Timestamp: time.Now(),
  Version:   "postgresql",
 }

 start := time.Now()

 if err := a.Ping(); err != nil {
  status.Status = "unhealthy"
  status.Errors = append(status.Errors, err.Error())
  status.Latency = time.Since(start)
  return status, nil
 }

 status.Latency = time.Since(start)

 if a.db != nil {
  sqlDB, _ := a.db.DB()
  stats := sqlDB.Stats()
  status.OpenConns = stats.OpenConnections
  status.IdleConns = stats.Idle
  status.InUse = stats.InUse
  status.WaitCount = stats.WaitCount
  status.WaitDuration = stats.WaitDuration

  if stats.OpenConnections >= stats.MaxOpenConnections*9/10 {
   status.Status = "degraded"
  }
 }

 return status, nil
}

func (a *PostgreSQLAdapter) ExecuteRaw(sql string, args ...interface{}) (interface{}, error) {
 if a.db == nil {
  return nil, fmt.Errorf("database not connected")
 }

 var result []map[string]interface{}
 rows, err := a.db.Raw(sql, args...).Rows()
 if err != nil {
  return nil, fmt.Errorf("failed to execute raw SQL: %w", err)
 }
 defer rows.Close()

 columns, err := rows.Columns()
 if err != nil {
  return nil, fmt.Errorf("failed to get columns: %w", err)
 }

 for rows.Next() {
  values := make([]interface{}, len(columns))
  valuePtrs := make([]interface{}, len(columns))
  for i := range values {
   valuePtrs[i] = &values[i]
  }

  if err := rows.Scan(valuePtrs...); err != nil {
   continue
  }

  row := make(map[string]interface{})
  for i, col := range columns {
   row[col] = values[i]
  }
  result = append(result, row)
 }

 return result, nil
}

func (a *PostgreSQLAdapter) Migrate(models ...interface{}) error {
 if a.db == nil {
  return fmt.Errorf("database not connected")
 }
 return a.db.AutoMigrate(models...)
}

func (a *PostgreSQLAdapter) Transaction(fn func(tx *gorm.DB) error) error {
 if a.db == nil {
  return fmt.Errorf("database not connected")
 }
 return a.db.Transaction(fn)
}

type MySQLAdapter struct {
 db  *gorm.DB
 cfg *DatabaseConfig
}

func NewMySQLAdapter(cfg *DatabaseConfig) *MySQLAdapter {
 return &MySQLAdapter{cfg: cfg}
}

func (a *MySQLAdapter) Connect(ctx context.Context, config *DatabaseConfig) (*gorm.DB, error) {
 a.cfg = config

 dsn := fmt.Sprintf(
  "%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
  config.User,
  config.Password,
  config.Host,
  config.Port,
  config.DBName,
 )

 if len(config.Parameters) > 0 {
  for k, v := range config.Parameters {
   dsn += fmt.Sprintf("&%s=%s", k, v)
  }
 }

 gormConfig := &gorm.Config{
  Logger: logger.Default.LogMode(config.LoggerLevel),
 }

 if !config.EnableLog {
  gormConfig.Logger = logger.Default.LogMode(logger.Silent)
 }

 db, err := gorm.Open(mysql.Open(dsn), gormConfig)
 if err != nil {
  return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
 }

 sqlDB, err := db.DB()
 if err != nil {
  return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
 }

 maxOpen := config.MaxOpenConns
 if maxOpen <= 0 {
  maxOpen = 100
 }
 maxIdle := config.MaxIdleConns
 if maxIdle <= 0 {
  maxIdle = 10
 }
 lifetime := config.MaxLifetime
 if lifetime <= 0 {
  lifetime = time.Hour
 }

 sqlDB.SetMaxOpenConns(maxOpen)
 sqlDB.SetMaxIdleConns(maxIdle)
 sqlDB.SetConnMaxLifetime(lifetime)

 a.db = db

 log.Printf("[MySQLAdapter] Connected to %s:%d/%s", config.Host, config.Port, config.DBName)

 return db, nil
}

func (a *MySQLAdapter) Disconnect() error { return disconnectDB(a.db) }
func (a *MySQLAdapter) Ping() error        { return pingDB(a.db) }
func (a *MySQLAdapter) GetDialect() string  { return "mysql" }
func (a *MySQLAdapter) IsConnected() bool   { return a.db != nil }

func (a *MySQLAdapter) HealthCheck() (*HealthStatus, error) {
 status := &HealthStatus{
  Status:    "healthy",
  Timestamp: time.Now(),
  Version:   "mysql",
 }

 start := time.Now()
 if err := a.Ping(); err != nil {
  status.Status = "unhealthy"
  status.Errors = append(status.Errors, err.Error())
  status.Latency = time.Since(start)
  return status, nil
 }
 status.Latency = time.Since(start)

 if a.db != nil {
  fillStats(a.db, status)
 }

 return status, nil
}

func (a *MySQLAdapter) ExecuteRaw(sql string, args ...interface{}) (interface{}, error) {
 return executeRawQuery(a.db, sql, args...)
}

func (a *MySQLAdapter) Migrate(models ...interface{}) error { return migrateModels(a.db, models...) }
func (a *MySQLAdapter) Transaction(fn func(tx *gorm.DB) error) error {
 if a.db == nil {
  return fmt.Errorf("database not connected")
 }
 return a.db.Transaction(fn)
}

type AdapterManager struct {
 adapters map[string]DatabaseAdapter
 primary  string
 mu       sync.RWMutex
}

func NewAdapterManager() *AdapterManager {
 return &AdapterManager{
  adapters: make(map[string]DatabaseAdapter),
 }
}

func (m *AdapterManager) Register(name string, adapter DatabaseAdapter) error {
 m.mu.Lock()
 defer m.mu.Unlock()

 if _, exists := m.adapters[name]; exists {
  return fmt.Errorf("adapter %s already registered", name)
 }

 m.adapters[name] = adapter

 if m.primary == "" {
  m.primary = name
 }

 log.Printf("[AdapterManager] Registered adapter: %s (%s)", name, adapter.GetDialect())

 return nil
}

func (m *AdapterManager) SetPrimary(name string) error {
 m.mu.Lock()
 defer m.mu.Unlock()

 if _, exists := m.adapters[name]; !exists {
  return fmt.Errorf("adapter %s not found", name)
 }

 m.primary = name
 return nil
}

func (m *AdapterManager) GetPrimary() (DatabaseAdapter, bool) {
 m.mu.RLock()
 defer m.mu.RUnlock()

 adapter, ok := m.adapters[m.primary]
 return adapter, ok
}

func (m *AdapterManager) Get(name string) (DatabaseAdapter, bool) {
 m.mu.RLock()
 defer m.mu.RUnlock()

 adapter, ok := m.adapters[name]
 return adapter, ok
}

func (m *AdapterManager) GetAllHealth() map[string]*HealthStatus {
 m.mu.RLock()
 defer m.mu.RUnlock()

 results := make(map[string]*HealthStatus)

 for name, adapter := range m.adapters {
  health, err := adapter.HealthCheck()
  if err != nil {
   hr := &HealthResult{
    Status:  "unhealthy",
    Errors: []string{err.Error()},
   }
   results[name] = hr.ToHealthStatus()
   continue
  }
  results[name] = health
 }

 return results
}

func (m *AdapterManager) DisconnectAll() {
 m.mu.Lock()
 defer m.mu.Unlock()

 for name, adapter := range m.adapters {
  if err := adapter.Disconnect(); err != nil {
   log.Printf("[AdapterManager] Error disconnecting %s: %v", name, err)
  } else {
   log.Printf("[AdapterManager] Disconnected: %s", name)
  }
 }
}

func CreateAdapter(driver string, config *DatabaseConfig) (DatabaseAdapter, error) {
 switch driver {
 case "postgres", "postgresql":
  return NewPostgreSQLAdapter(config), nil
 case "mysql", "mariadb":
  return NewMySQLAdapter(config), nil
 default:
  return nil, fmt.Errorf("unsupported database driver: %s", driver)
 }
}

func disconnectDB(db *gorm.DB) error {
 if db == nil {
  return nil
 }
 sqlDB, err := db.DB()
 if err != nil {
  return err
 }
 return sqlDB.Close()
}

func pingDB(db *gorm.DB) error {
 if db == nil {
  return fmt.Errorf("database not connected")
 }
 sqlDB, err := db.DB()
 if err != nil {
  return err
 }
 return sqlDB.Ping()
}

func fillStats(db *gorm.DB, status *HealthStatus) {
 sqlDB, err := db.DB()
 if err != nil {
  return
 }
 stats := sqlDB.Stats()
 status.OpenConns = stats.OpenConnections
 status.IdleConns = stats.Idle
 status.InUse = stats.InUse
 status.WaitCount = stats.WaitCount
 status.WaitDuration = stats.WaitDuration

 if stats.OpenConnections >= stats.MaxOpenConnections*9/10 {
  status.Status = "degraded"
 }
}

func executeRawQuery(db *gorm.DB, sql string, args ...interface{}) (interface{}, error) {
 if db == nil {
  return nil, fmt.Errorf("database not connected")
 }

 var result []map[string]interface{}
 rows, err := db.Raw(sql, args...).Rows()
 if err != nil {
  return nil, fmt.Errorf("failed to execute raw SQL: %w", err)
 }
 defer rows.Close()

 columns, err := rows.Columns()
 if err != nil {
  return nil, fmt.Errorf("failed to get columns: %w", err)
 }

 for rows.Next() {
  values := make([]interface{}, len(columns))
  valuePtrs := make([]interface{}, len(columns))
  for i := range values {
   valuePtrs[i] = &values[i]
  }

  if err := rows.Scan(valuePtrs...); err != nil {
   continue
  }

  row := make(map[string]interface{})
  for i, col := range columns {
   row[col] = values[i]
  }
  result = append(result, row)
 }

 return result, nil
}

func migrateModels(db *gorm.DB, models ...interface{}) error {
 if db == nil {
  return fmt.Errorf("database not connected")
 }
 return db.AutoMigrate(models...)
}

type HealthResult struct {
 Status  string   `json:"status"`
 Latency int64   `json:"latency_ms"`
 Errors  []string `json:"errors,omitempty"`
}

func (r *HealthResult) ToHealthStatus() *HealthStatus {
 return &HealthStatus{
  Status: r.Status,
  Latency: time.Duration(r.Latency) * time.Millisecond,
  Errors: r.Errors,
  Timestamp: time.Now(),
 }
}

func CheckDatabaseConnection(db *sql.DB, dialect string) (*HealthStatus, error) {
 start := time.Now()

 if err := db.Ping(); err != nil {
  hr := &HealthResult{
   Status: "unhealthy",
   Latency: time.Since(start).Milliseconds(),
   Errors: []string{err.Error()},
  }
  return hr.ToHealthStatus(), nil
 }

 stats := db.Stats()
 latency := time.Since(start)

 status := "healthy"
 if stats.OpenConnections >= stats.MaxOpenConnections*9/10 {
  status = "degraded"
 }

 return &HealthStatus{
  Status:      status,
  Latency:     latency,
  OpenConns:   stats.OpenConnections,
  IdleConns:   stats.Idle,
  InUse:       stats.InUse,
  WaitCount:   stats.WaitCount,
  WaitDuration: stats.WaitDuration,
  Timestamp:   time.Now(),
  Version:     dialect,
 }, nil
}
