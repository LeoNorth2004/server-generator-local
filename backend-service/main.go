package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/middleware"
	"github.com/generator-platform/go-common/models"
	"github.com/gin-gonic/gin"
)

const generatorServiceURL = "http://localhost:8084"

func extractPort(portStr string) string {
	if portStr == "" {
		return ""
	}
	if strings.Contains(portStr, ":") {
		parts := strings.Split(portStr, ":")
		lastPart := parts[len(parts)-1]
		if _, err := strconv.Atoi(lastPart); err == nil {
			return lastPart
		}
	}
	if _, err := strconv.Atoi(portStr); err == nil {
		return portStr
	}
	return ""
}

func proxyRequestToGenerator(c *gin.Context) {
	targetURL, _ := url.Parse(generatorServiceURL)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host

		requestPath := c.Param("proxyPath")
		if requestPath != "" {
			req.URL.Path = "/api/v1/" + requestPath
		}

		queryString := c.Request.URL.RawQuery
		if queryString != "" {
			req.URL.RawQuery = queryString
		}
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		log.Printf("[PROXY] %s %s → %d", c.Request.Method, c.Request.URL.Path, resp.StatusCode)
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[PROXY] Error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": "Generator service unavailable",
			"error":   err.Error(),
		})
	}

	proxy.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

func main() {
	cfg := config.LoadConfig()

	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	_, err := database.InitDB(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.OperationLog{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	initDefaultUser()
	initSampleData()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	setupRoutes(r)

	port := os.Getenv("BACKEND_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8080"
	}

	log.Printf("Backend Service starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  - Auth:      /api/v1/auth/*")
	log.Printf("  - Users:     /api/v1/users/*")
	log.Printf("  - Projects:  /api/v1/projects/*")
	log.Printf("  - Ops:       /api/v1/operations/*")
	log.Printf("  - Generator: /api/v1/generator/* (→ :8084)")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/login", loginHandler)
		auth.POST("/register", registerHandler)
		auth.GET("/me", middleware.AuthMiddleware(), getCurrentUserHandler)
	}

	users := api.Group("/users")
	{
		users.POST("", middleware.AuthMiddleware(), createUserHandler)
		users.GET("/:id", middleware.AuthMiddleware(), getUserHandler)
		users.PUT("/:id", middleware.AuthMiddleware(), updateUserHandler)
		users.DELETE("/:id", middleware.AuthMiddleware(), deleteUserHandler)
		users.GET("", middleware.AuthMiddleware(), listUsersHandler)
	}

	projects := api.Group("/projects")
	projects.Use(middleware.AuthMiddleware())
	{
		projects.POST("", createProjectHandler)
		projects.GET("/:id", getProjectHandler)
		projects.PUT("/:id", updateProjectHandler)
		projects.DELETE("/:id", deleteProjectHandler)
		projects.GET("", listProjectsHandler)
	}

	operations := api.Group("/operations")
	operations.Use(middleware.AuthMiddleware())
	{
		operations.GET("/health", healthCheckHandler)
		operations.GET("/stats", getStatsHandler)
		operations.GET("/metrics", getSystemMetricsHandler)
		operations.GET("/services", getServicesStatusHandler)
		operations.GET("/events", getRecentEventsHandler)
		operations.GET("/overview", getOverviewHandler)
		operations.GET("/operation-logs", getOperationLogsHandler)
		operations.POST("/operation-logs/record", recordOperationLogHandler)
	}

	generator := api.Group("/generator")
	generator.Use(middleware.AuthMiddleware())
	{
		generator.Any("/*path", proxyRequestToGenerator)
	}
}
