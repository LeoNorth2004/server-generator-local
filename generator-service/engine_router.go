package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type RouteRegistry struct {
	mu      sync.RWMutex
	routes  map[string]RouteConfig
	groups  map[string][]RouteConfig
}

type RouteConfig struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Handler     string                 `json:"handler"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ModelName   string                 `json:"model_name"`
	Operation   string                 `json:"operation"`
	Middlewares []string               `json:"middlewares"`
	RateLimit   *RateLimitConfig       `json:"rate_limit,omitempty"`
	Cache       *CacheConfig           `json:"cache,omitempty"`
	Auth        *AuthConfig            `json:"auth,omitempty"`
	Validation  *ValidationConfig      `json:"validation,omitempty"`
	Hooks       []HookConfig           `json:"hooks,omitempty"`
	Extends     map[string]interface{} `json:"extends,omitempty"`
}

type RateLimitConfig struct {
	Requests int `json:"requests"`
	Window   int `json:"window"` // seconds
}

type CacheConfig struct {
	Enabled bool          `json:"enabled"`
	TTL     time.Duration `json:"ttl"`
	KeyFunc string        `json:"key_func,omitempty"`
}

type AuthConfig struct {
	Required    bool     `json:"required"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type ValidationConfig struct {
	Enabled  bool              `json:"enabled"`
	Rules    map[string]string  `json:"rules,omitempty"`
	Custom   string            `json:"custom_validator,omitempty"`
}

type HookConfig struct {
	When   string `json:"when"` // before, after
	Action string `json:"action"`
	Order  int    `json:"order"`
}

func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make(map[string]RouteConfig),
		groups: make(map[string][]RouteConfig),
	}
}

func (r *RouteRegistry) Register(route RouteConfig) error {
 r.mu.Lock()
 defer r.mu.Unlock()

 key := fmt.Sprintf("%s:%s", route.Method, route.Path)
 if _, exists := r.routes[key]; exists {
  return fmt.Errorf("route %s already registered", key)
 }

 r.routes[key] = route
 log.Printf("[RouteRegistry] Registered: %s %s -> %s", route.Method, route.Path, route.Handler)

 return nil
}

func (r *RouteRegistry) RegisterGroup(prefix string, routes []RouteConfig) error {
 r.mu.Lock()
 defer r.mu.Unlock()

 for _, route := range routes {
  route.Path = prefix + route.Path
  key := fmt.Sprintf("%s:%s", route.Method, route.Path)
  if _, exists := r.routes[key]; exists {
   return fmt.Errorf("route %s already registered in group %s", key, prefix)
  }
  r.routes[key] = route
 }

 r.groups[prefix] = append(r.groups[prefix], routes...)
 log.Printf("[RouteRegistry] Registered group: %s (%d routes)", prefix, len(routes))

 return nil
}

func (r *RouteRegistry) GetRoutes() []RouteConfig {
 r.mu.RLock()
 defer r.mu.RUnlock()

 routes := make([]RouteConfig, 0, len(r.routes))
 for _, route := range r.routes {
  routes = append(routes, route)
 }
 return routes
}

func (r *RouteRegistry) GetGroup(prefix string) ([]RouteConfig, bool) {
 r.mu.RLock()
 defer r.mu.RUnlock()

 routes, ok := r.groups[prefix]
 return routes, ok
}

func (r *RouteRegistry) GenerateRouterCode(projectName string) string {
 var sb strings.Builder

 sb.WriteString(fmt.Sprintf(`package router

import (
 "%s/internal/controller"
 "%s/internal/middleware"
 "%s/pkg/utils"

 "github.com/gin-gonic/gin"
 "gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
 // Global middleware
 r.Use(middleware.CORS())
 r.Use(middleware.Logger())
 r.Use(middleware.Recovery())
 r.Use(middleware.RequestID())

 // Health check endpoint
 r.GET("/health", func(c *gin.Context) {
  utils.Success(c, gin.H{
   "status":    "healthy",
   "timestamp": time.Now().Unix(),
   "version":   "1.0.0",
  })
 })

`, projectName, projectName, projectName))

 sb.WriteString("\n\t// API Routes\n")
 sb.WriteString("\tapi := r.Group(\"/api/v1\")\n")

 groups := make(map[string][]string)
 r.mu.RLock()
 for _, route := range r.routes {
  prefix := extractGroupPrefix(route.Path)
  groups[prefix] = append(groups[prefix], route.Path)
 }
 r.mu.RUnlock()

 for prefix := range groups {
  sb.WriteString(fmt.Sprintf("\n\t// %s\n", strings.TrimPrefix(prefix, "/api/v1/")))
  sb.WriteString(fmt.Sprintf("\t%s := api.Group(\"%s\")\n", sanitizeVarName(prefix), strings.TrimPrefix(prefix, "/api/v1")))
  sb.WriteString("\t{\n")

  r.mu.RLock()
  for _, route := range r.routes {
   if strings.HasPrefix(route.Path, prefix) {
    methodLower := strings.ToLower(route.Method)
    handlerFunc := formatHandlerName(route.Handler)
    relativePath := strings.TrimPrefix(route.Path, prefix)

    middlewareStr := ""
    if len(route.Middlewares) > 0 {
     middlewareStr = ", " + strings.Join(route.Middlewares, ", ")
    }

    sb.WriteString(fmt.Sprintf("\t\t%s.%s(\"%s\", ctrl.%s%s)\n",
     sanitizeVarName(prefix), methodLower, relativePath, handlerFunc, middlewareStr))
   }
  }
  r.mu.RUnlock()

  sb.WriteString("\t}\n")
 }

 sb.WriteString("}\n\n")

 return sb.String()
}

func extractGroupPrefix(path string) string {
 parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
 if len(parts) >= 3 {
  return "/" + parts[0] + "/" + parts[1]
 }
 return path
}

func sanitizeVarName(s string) string {
 s = strings.ReplaceAll(s, "/", "_")
 s = strings.ReplaceAll(s, "-", "_")
 s = strings.ReplaceAll(s, ":", "")
 return "group" + s
}

func formatHandlerName(handler string) string {
 parts := strings.Split(handler, ".")
 if len(parts) > 0 {
  return parts[len(parts)-1]
 }
 return handler
}
