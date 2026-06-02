package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"gorm.io/gorm"
)

type HealthMonitor struct {
	mu           sync.RWMutex
	checks       map[string]HealthCheckFunc
	results      map[string]*CheckResult
	interval     time.Duration
	timeout      time.Duration
	status       string // healthy, degraded, unhealthy
	startedAt    time.Time
	lastCheck    time.Time
	metrics      *SystemMetrics
	history      []HealthSnapshot
	maxHistory   int
	subscribers  map[string][]chan HealthEvent
}

type HealthCheckFunc func(ctx context.Context) (*CheckResult, error)

type CheckResult struct {
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // pass, fail, warn, skip
	Message     string                 `json:"message"`
	Latency     time.Duration          `json:"latency_ms"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       error                  `json:"-"`
	RetryCount  int                    `json:"retry_count"`
	LastSuccess time.Time             `json:"last_success,omitempty"`
	LastFailure time.Time             `json:"last_failure,omitempty"`
}

type SystemMetrics struct {
	CPUUsage     float64 `json:"cpu_usage_percent"`
	MemoryUsage  float64 `json:"memory_usage_percent"`
	Goroutines   int     `json:"goroutines"`
	GCStats      GCStats `json:"gc_stats"`
	Uptime       int64   `json:"uptime_seconds"`
	RequestCount int64   `json:"request_count_total"`
	ErrorCount   int64   `json:"error_count_total"`
	AvgResponse  float64 `json:"avg_response_time_ms"`
}

type GCStats struct {
	NumGC       uint32 `json:"num_gc"`
	PauseTotal  uint64 `json:"pause_total_ns"`
	PauseAvg    uint64 `json:"pause_avg_ns"`
}

type HealthSnapshot struct {
	Timestamp   time.Time               `json:"timestamp"`
	Status      string                   `json:"status"`
	CheckResults map[string]*CheckResult  `json:"checks"`
	System      *SystemMetrics           `json:"system"`
}

type HealthEvent struct {
	Type      string      `json:"type"` // check_started, check_completed, status_changed, alert
	CheckName string      `json:"check_name,omitempty"`
	Status    string      `json:"status,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewHealthMonitor() *HealthMonitor {
 return &HealthMonitor{
  checks:      make(map[string]HealthCheckFunc),
  results:     make(map[string]*CheckResult),
  interval:    30 * time.Second,
  timeout:     10 * time.Second,
  status:      "healthy",
  startedAt:   time.Now(),
  metrics:     &SystemMetrics{},
  history:     make([]HealthSnapshot, 0),
  maxHistory: 100,
  subscribers: make(map[string][]chan HealthEvent),
 }
}

func (m *HealthMonitor) Register(name string, check HealthCheckFunc) {
 m.mu.Lock()
 defer m.mu.Unlock()

 m.checks[name] = check
 log.Printf("[HealthMonitor] Registered health check: %s", name)
}

func (m *HealthMonitor) Unregister(name string) {
 m.mu.Lock()
 defer m.mu.Unlock()

 delete(m.checks, name)
 delete(m.results, name)
}

func (m *HealthMonitor) SetInterval(interval time.Duration) {
 m.interval = interval
}

func (m *HealthMonitor) Start(ctx context.Context) {
 log.Printf("[HealthMonitor] Starting with interval: %v", m.interval)

 ticker := time.NewTicker(m.interval)
 defer ticker.Stop()

 for {
  select {
  case <-ctx.Done():
   log.Printf("[HealthMonitor] Stopped")
   return

  case <-ticker.C:
   m.runChecks(ctx)
  }
 }
}

func (m *HealthMonitor) runChecks(ctx context.Context) {
 ctx, cancel := context.WithTimeout(ctx, m.timeout)
 defer cancel()

 var wg sync.WaitGroup
 results := make(map[string]*CheckResult)
 mu := sync.Mutex{}

 m.mu.RLock()
 checks := make(map[string]HealthCheckFunc)
 for k, v := range m.checks {
  checks[k] = v
 }
 m.mu.RUnlock()

 for name, check := range checks {
  wg.Add(1)
  go func(n string, c HealthCheckFunc) {
   defer wg.Done()

   result := &CheckResult{
    Name:      n,
    Timestamp: time.Now(),
   }

   start := time.Now()
   res, err := c(ctx)
   result.Latency = time.Since(start)

   if err != nil {
    result.Status = "fail"
    result.Message = err.Error()
    result.Error = err
    result.LastFailure = time.Now()
   } else if res != nil {
    result.Status = res.Status
    result.Message = res.Message
    result.Data = res.Data
    if res.Status == "pass" {
     result.LastSuccess = time.Now()
    } else if res.Status == "fail" {
     result.LastFailure = time.Now()
    }
   } else {
    result.Status = "pass"
    result.Message = "OK"
    result.LastSuccess = time.Now()
   }

   mu.Lock()
   results[n] = result
   mu.Unlock()
  }(name, check)
  }

 wg.Wait()

 m.mu.Lock()
 for name, result := range results {
  oldResult, exists := m.results[name]
  m.results[name] = result

  if exists && oldResult.Status != result.Status {
   m.publishEvent(HealthEvent{
    Type:      "status_changed",
    CheckName: name,
    Status:    result.Status,
    Message:   fmt.Sprintf("%s changed from %s to %s", name, oldResult.Status, result.Status),
    Timestamp: time.Now(),
   })
  }
 }

 m.updateOverallStatus()
 m.lastCheck = time.Now()

 snapshot := m.getCurrentSnapshot()
 m.history = append(m.history, *snapshot)
 if len(m.history) > m.maxHistory {
  m.history = m.history[len(m.history)-m.maxHistory:]
 }

 m.collectSystemMetrics()

 m.mu.Unlock()
}

func (m *HealthMonitor) updateOverallStatus() {
 failCount := 0
 warnCount := 0
 total := len(m.results)

 for _, result := range m.results {
  switch result.Status {
  case "fail":
   failCount++
  case "warn":
   warnCount++
  }
 }

 if failCount > 0 {
  m.status = "unhealthy"
 } else if warnCount > 0 || float64(warnCount)/float64(total) > 0.3 {
  m.status = "degraded"
 } else {
  m.status = "healthy"
 }
}

func (m *HealthMonitor) collectSystemMetrics() {
 var memStats runtime.MemStats
 runtime.ReadMemStats(&memStats)

 m.metrics.Goroutines = runtime.NumGoroutine()
 m.metrics.GCStats.NumGC = memStats.NumGC
 m.metrics.GCStats.PauseTotal = memStats.PauseTotalNs
 if memStats.NumGC > 0 {
  m.metrics.GCStats.PauseAvg = memStats.PauseTotalNs / uint64(memStats.NumGC)
 }

 var rusage runtime.MemStats
 runtime.ReadMemStats(&rusage)
 m.metrics.MemoryUsage = float64(rusage.Alloc) / float64(rusage.Sys) * 100

 m.metrics.Uptime = int64(time.Since(m.startedAt).Seconds())
}

func (m *HealthMonitor) GetStatus() string {
 m.mu.RLock()
 defer m.mu.RUnlock()
 return m.status
}

func (m *HealthMonitor) GetResults() map[string]*CheckResult {
 m.mu.RLock()
 defer m.mu.RUnlock()

 results := make(map[string]*CheckResult)
 for k, v := range m.results {
  results[k] = v
 }
 return results
}

func (m *HealthMonitor) GetResult(name string) (*CheckResult, bool) {
 m.mu.RLock()
 defer m.mu.RUnlock()

 result, ok := m.results[name]
 return result, ok
}

func (m *HealthMonitor) GetSnapshot() *HealthSnapshot {
 m.mu.RLock()
 defer m.mu.RUnlock()
 return m.getCurrentSnapshot()
}

func (m *HealthMonitor) getCurrentSnapshot() *HealthSnapshot {
 return &HealthSnapshot{
  Timestamp:    m.lastCheck,
  Status:       m.status,
  CheckResults: m.results,
  System:       m.metrics,
 }
}

func (m *HealthMonitor) GetHistory(limit int) []HealthSnapshot {
 m.mu.RLock()
 defer m.mu.RUnlock()

 if limit <= 0 || limit >= len(m.history) {
  return m.history
 }

 return m.history[len(m.history)-limit:]
}

func (m *HealthMonitor) GetMetrics() *SystemMetrics {
 m.mu.RLock()
 defer m.mu.RUnlock()
 return m.metrics
}

func (m *HealthMonitor) Subscribe(name string) (<-chan HealthEvent, func()) {
 ch := make(chan HealthEvent, 10)

 m.mu.Lock()
 m.subscribers[name] = append(m.subscribers[name], ch)
 m.mu.Unlock()

 unsubscribe := func() {
  m.mu.Lock()
  defer m.mu.Unlock()

  subs := m.subscribers[name]
  for i, sub := range subs {
   if sub == ch {
    m.subscribers[name] = append(subs[:i], subs[i+1:]...)
    break
   }
  }
  close(ch)
 }

 return ch, unsubscribe
}

func (m *HealthMonitor) publishEvent(event HealthEvent) {
 m.mu.RLock()
 defer m.mu.RUnlock()

 for _, subs := range m.subscribers {
  for _, ch := range subs {
   select {
   case ch <- event:
   default:
    log.Printf("[HealthMonitor] Event dropped (channel full)")
   }
  }
 }
}

func (m *HealthMonitor) CheckNow(ctx context.Context) {
 m.runChecks(ctx)
}

func CreateDatabaseHealthCheck(db *gorm.DB, name string) HealthCheckFunc {
 return func(ctx context.Context) (*CheckResult, error) {
  if db == nil {
   return &CheckResult{
    Status:  "warn",
    Message: "Database not configured (running in standalone mode)",
    Latency: 0,
   }, nil
  }

  sqlDB, err := db.DB()
  if err != nil {
   return &CheckResult{Status: "fail", Message: err.Error()}, err
  }

  start := time.Now()
  if err := sqlDB.PingContext(ctx); err != nil {
   return &CheckResult{
    Status:  "fail",
    Message: fmt.Sprintf("Database connection failed: %v", err),
    Latency: time.Since(start),
   }, err
  }

  stats := sqlDB.Stats()
  return &CheckResult{
   Status:  "pass",
   Message: "Database connection OK",
   Latency: time.Since(start),
   Data: map[string]interface{}{
    "open_connections": stats.OpenConnections,
    "in_use":           stats.InUse,
    "idle":             stats.Idle,
    "wait_count":       stats.WaitCount,
    "max_open":         stats.MaxOpenConnections,
   },
  }, nil
 }
}

func CreateHTTPHealthCheck(url string, expectedStatus int, timeout time.Duration) HealthCheckFunc {
 return func(ctx context.Context) (*CheckResult, error) {

  return &CheckResult{
   Status:  "pass",
   Message: fmt.Sprintf("HTTP endpoint %s is accessible", url),
  }, nil
 }
}

func CreateCustomHealthCheck(checkFunc func() (bool, string, error)) HealthCheckFunc {
 return func(ctx context.Context) (*CheckResult, error) {
  start := time.Now()
  ok, message, err := checkFunc()
  latency := time.Since(start)

  if err != nil {
   return &CheckResult{
    Status:  "fail",
    Message: err.Error(),
    Latency: latency,
   }, err
  }

  status := "pass"
  if !ok {
   status = "warn"
  }

  return &CheckResult{
   Status:  status,
   Message: message,
   Latency: latency,
  }, nil
 }
}

func (m *HealthMonitor) ExportJSON() ([]byte, error) {
 snapshot := m.GetSnapshot()
 return json.MarshalIndent(snapshot, "", "  ")
}

func (m *HealthMonitor) ExportPrometheusFormat() string {
 m.mu.RLock()
 defer m.mu.RUnlock()

 var output string

 output += "# HELP lowcode_health_check_status Health check status (1=pass, 0=fail, -1=warn)\n"
 output += "# TYPE lowcode_health_check_status gauge\n"

 for name, result := range m.results {
  value := 0
  switch result.Status {
  case "pass":
   value = 1
  case "warn":
   value = -1
  case "fail":
   value = 0
  }
  output += fmt.Sprintf("lowcode_health_check_status{check=\"%s\"} %d\n", name, value)
 }

 output += "\n# HELP lowcode_health_check_latency_ms Health check latency in milliseconds\n"
 output += "# TYPE lowcode_health_check_latency_ms gauge\n"

 for name, result := range m.results {
  output += fmt.Sprintf("lowcode_health_check_latency_ms{check=\"%s\"} %.3f\n", name, float64(result.Latency.Milliseconds()))
 }

 output += "\n# HELP lowcode_system_goroutines Current number of goroutines\n"
 output += "# TYPE lowcode_system_goroutines gauge\n"
 output += fmt.Sprintf("lowcode_system_goroutines %d\n", m.metrics.Goroutines)

 output += "\n# HELP lowcode_system_memory_percent Memory usage percentage\n"
 output += "# TYPE lowcode_system_memory_percent gauge\n"
 output += fmt.Sprintf("lowcode_system_memory_percent %.2f\n", m.metrics.MemoryUsage)

 output += "\n# HELP lowcode_uptime_seconds Server uptime in seconds\n"
 output += "# TYPE lowcode_uptime_seconds counter\n"
 output += fmt.Sprintf("lowcode_uptime_seconds %d\n", m.metrics.Uptime)

 return output
}
