package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/generator-platform/go-common/config"
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/middleware"
	"github.com/generator-platform/go-common/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	JWTSecret     = "low-code-platform-secret-key-2024"
	JWTExpiration = 24 * time.Hour
)

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type Project struct {
	ID           uint      `json:"id"`
	DBID         uint      `json:"db_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Technology   string    `json:"technology"`
	Status       string    `json:"status"`
	TableConfig  string    `json:"table_config"`
	DBConfig     string    `json:"db_config"`
	CreatedBy    uint      `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OperationLog struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	TargetType string   `json:"target_type"`
	TargetID  string    `json:"target_id"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Status    string    `json:"status"`
	Duration  int64     `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	userStore      = make(map[string]*User)
	projectStore   = make(map[uint]*Project)
	projectIDCounter uint = 0
	logStore        = make([]OperationLog, 0)
	logIDCounter    uint = 0
	codeStore       = make(map[uint]*CodeEntry)
	codeIDCounter   uint = 0
	storeMutex      sync.RWMutex
	startTime       time.Time
	db              *gorm.DB
)

func nextProjectID() uint {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	projectIDCounter++
	return projectIDCounter
}

func nextCodeID() uint {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	codeIDCounter++
	return codeIDCounter
}

type CodeEntry struct {
	ID          uint      `json:"id"`
	ProjectID   uint      `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	FileType    string    `json:"file_type"`
	CreatedAt   time.Time `json:"created_at"`
}

func init() {
	defaultUsers := map[string]*User{
		"admin": {
			ID:       1,
			Username: "admin",
			Password: "admin123",
			Email:    "admin@lowcode.com",
			Role:     "admin",
			Status:   "active",
		},
	}

	for k, v := range defaultUsers {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(v.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[Auth] Failed to hash password for user %s: %v", k, err)
			continue
		}
		v.Password = string(hashedPassword)
		userStore[k] = v
		log.Printf("[Auth] Initialized default user: %s (role: %s)", k, v.Role)
	}
	log.Printf("[Auth] Total initialized users: %d", len(userStore))
}

func recordLog(userID uint, username, action, targetType, targetID, details, ipAddress, userAgent, status string, duration int64) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	logIDCounter++
	logEntry := OperationLog{
		ID:         logIDCounter,
		UserID:     userID,
		Username:   username,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Status:     status,
		Duration:   duration,
		CreatedAt:  time.Now(),
	}

	logStore = append(logStore, logEntry)
	log.Printf("[OperationLog] [%s] %s - %s: %s (%dms)", logEntry.CreatedAt.Format("2006-01-02 15:04:05"), username, action, details, duration)

	if db != nil {
		dbLog := models.OperationLog{
			UserID:     userID,
			Username:   username,
			Action:     action,
			Resource:   targetType,
			ResourceID: parseUint(targetID),
			Details:    details,
			Status:     status,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			Duration:   duration,
		}
		if result := db.Create(&dbLog); result.Error != nil {
			log.Printf("[OperationLog] Failed to save to database: %v", result.Error)
		} else {
			log.Printf("[OperationLog] Saved to database with ID: %d", dbLog.ID)
		}
	}
}

func parseUint(s string) uint {
	var u uint
	fmt.Sscanf(s, "%d", &u)
	return u
}

func generateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(JWTExpiration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

func main() {
	startTime = time.Now()

	log.Println("====================================")
	log.Println("  Low-Code Platform - Generator Service")
	log.Println("  Version: 2.0.0 (Enterprise Edition)")
	log.Println("====================================")

	cfg := config.LoadConfig()

	log.Printf("[DEBUG] Database Config: Host=%s, Port=%d, User=%s, Password=%s, DBName=%s, SSLMode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, fmt.Sprintf("%s***", cfg.DBPassword[:3]), cfg.DBName, cfg.DBSSLMode)

	dbConn, err := database.InitDB(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
		log.Printf("Continuing without database connection...")
	} else {
		db = dbConn
		if err := db.AutoMigrate(&models.Project{}, &models.OperationLog{}); err != nil {
			log.Printf("Warning: Failed to migrate database (table may already exist): %v", err)
		} else {
			log.Println("[Database] AutoMigration completed successfully")
			
			var logCount int64
			db.Model(&models.OperationLog{}).Count(&logCount)
			log.Printf("[Database] Current operation logs count: %d", logCount)
		}
	}

	metadataEngine := NewMetadataEngine(db)
	workflowEngine := NewWorkflowEngine(db)
	formEngine := NewFormEngine(metadataEngine)
	healthMonitor := NewHealthMonitor()

	healthMonitor.Register("database", CreateDatabaseHealthCheck(db, "PostgreSQL"))
	healthMonitor.Register("system", CreateCustomHealthCheck(func() (bool, string, error) {
		return true, "System running normally", nil
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go healthMonitor.Start(ctx)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RateLimitMiddleware(1000))

	setupAuthRoutes(r)

	r.Use(middleware.AuthMiddleware())

	setupAPIRoutes(r, metadataEngine, workflowEngine, formEngine, healthMonitor)

	port := os.Getenv("GENERATOR_SERVICE_PORT")
	port = extractPort(port)
	if port == "" {
		port = "8084"
	}

	go func() {
		log.Printf("[Startup] Health monitor started")
		time.Sleep(2 * time.Second)
		status := healthMonitor.GetStatus()
		results := healthMonitor.GetResults()
		log.Printf("[Startup] Initial health status: %s (%d checks)", status, len(results))
	}()

	log.Printf("Generator Service starting on port %s", port)
	log.Printf("[Info] Metadata Engine initialized")
	log.Printf("[Info] Workflow Engine initialized")
	log.Printf("[Info] Form Engine initialized")
	log.Printf("[Info] Health Monitor started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("[Shutdown] Shutting down server...")

	cancel()

	ctx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	log.Println("[Shutdown] Server exited properly")
}

func setupAuthRoutes(r *gin.Engine) {
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "用户名和密码不能为空",
					"data":    nil,
				})
				return
			}

			user, exists := userStore[req.Username]
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "用户名或密码错误",
					"data":    nil,
				})
				return
			}

			if user.Status != "active" {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "账号已被禁用，请联系管理员",
					"data":    nil,
				})
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "用户名或密码错误",
					"data":    nil,
				})
				return
			}

			token, err := generateToken(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "生成令牌失败",
					"data":    nil,
				})
				return
			}

			recordLog(user.ID, user.Username, "login", "auth", "", "用户登录系统", c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

			log.Printf("[Auth] User %s (ID: %d) logged in successfully", user.Username, user.ID)

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "登录成功",
				"data": gin.H{
					"token": token,
					"user": gin.H{
						"id":       user.ID,
						"username": user.Username,
						"email":    user.Email,
						"role":     user.Role,
					},
				},
			})
		})

		auth.POST("/register", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required,min=3,max=20"`
				Password string `json:"password" binding:"required,min=6"`
				Email    string `json:"email" binding:"required,email"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "注册信息不完整或格式错误",
					"data":    nil,
				})
				return
			}

			if _, exists := userStore[req.Username]; exists {
				c.JSON(http.StatusConflict, gin.H{
					"code":    409,
					"message": "用户名已存在",
					"data":    nil,
				})
				return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "密码加密失败",
					"data":    nil,
				})
				return
			}

			newUser := &User{
				ID:       uint(len(userStore) + 2),
				Username: req.Username,
				Password: string(hashedPassword),
				Email:    req.Email,
				Role:     "user",
				Status:   "active",
			}
			userStore[req.Username] = newUser

			token, _ := generateToken(newUser)

			log.Printf("[Auth] New user registered: %s (ID: %d)", newUser.Username, newUser.ID)

			c.JSON(http.StatusCreated, gin.H{
				"code":    201,
				"message": "注册成功",
				"data": gin.H{
					"token": token,
					"user": gin.H{
						"id":       newUser.ID,
						"username": newUser.Username,
						"email":    newUser.Email,
						"role":     newUser.Role,
					},
				},
			})
		})

		auth.GET("/me", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "未提供有效的认证令牌",
					"data":    nil,
				})
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(JWTSecret), nil
			})

			if err != nil || !token.Valid {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "令牌无效或已过期",
					"data":    nil,
				})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "令牌格式错误",
					"data":    nil,
				})
				return
			}

			username, _ := claims["username"].(string)
			user, exists := userStore[username]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "用户不存在",
					"data":    nil,
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "获取成功",
				"data": gin.H{
					"id":       user.ID,
					"username": user.Username,
					"email":    user.Email,
					"role":     user.Role,
				},
			})
		})
	}

	log.Println("[Auth] Authentication routes configured:")
	log.Println("  POST /api/v1/auth/login   - 用户登录")
	log.Println("  POST /api/v1/auth/register - 用户注册")
	log.Println("  GET  /api/v1/auth/me      - 获取当前用户")
}

func setupAPIRoutes(r *gin.Engine,
	metadataEngine *MetadataEngine,
	workflowEngine *WorkflowEngine,
	formEngine *FormEngine,
	healthMonitor *HealthMonitor) {

	r.GET("/health", func(c *gin.Context) {
		status := healthMonitor.GetStatus()
		c.JSON(200, gin.H{
			"status":    status,
			"service":   "generator-service",
			"version":   "2.0.0",
			"timestamp": time.Now().Unix(),
		})
	})

	r.GET("/health/details", func(c *gin.Context) {
		snapshot := healthMonitor.GetSnapshot()
		c.JSON(200, snapshot)
	})

	r.GET("/health/metrics", func(c *gin.Context) {
		prometheusFormat := healthMonitor.ExportPrometheusFormat()
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(200, prometheusFormat)
	})

	r.GET("/health/history", func(c *gin.Context) {
		history := healthMonitor.GetHistory(50)
		c.JSON(200, history)
	})

	projects := r.Group("/api/v1/projects")
	projects.Use(middleware.AuthMiddleware())
	{
		projects.GET("", func(c *gin.Context) {
			storeMutex.RLock()
			defer storeMutex.RUnlock()

			projectList := make([]*Project, 0, len(projectStore))
			for _, p := range projectStore {
				projectList = append(projectList, p)
			}

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取项目列表成功",
				"data":    projectList,
			})
		})

		projects.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			projectID := uint(0)
			fmt.Sscanf(id, "%d", &projectID)

			storeMutex.RLock()
			project, exists := projectStore[projectID]
			storeMutex.RUnlock()

			if !exists {
				c.JSON(404, gin.H{
					"code":    404,
					"message": "项目不存在",
					"data":    nil,
				})
				return
			}

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取项目详情成功",
				"data":    project,
			})
		})

		projects.POST("", func(c *gin.Context) {
			var req struct {
				Name        string `json:"name" binding:"required"`
				Description string `json:"description"`
				Technology  string `json:"technology"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{
					"code":    400,
					"message": "项目名称不能为空",
					"data":    nil,
				})
				return
			}

			storeMutex.Lock()
			projectIDCounter++
			now := time.Now()
			project := &Project{
				ID:          projectIDCounter,
				Name:        req.Name,
				Description: req.Description,
				Technology:  req.Technology,
				Status:      "generated",
				CreatedBy:   1,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			projectStore[project.ID] = project
			storeMutex.Unlock()

			recordLog(1, "admin", "create_project", "project", fmt.Sprintf("%d", project.ID), fmt.Sprintf("创建项目: %s", req.Name), c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

			log.Printf("[Projects] Created project: %s (ID: %d)", req.Name, project.ID)

			c.JSON(201, gin.H{
				"code":    201,
				"message": "项目创建成功",
				"data":    project,
			})
		})

		projects.PUT("/:id", func(c *gin.Context) {
			id := c.Param("id")
			projectID := uint(0)
			fmt.Sscanf(id, "%d", &projectID)

			var req struct {
				Name        string `json:"name" binding:"required"`
				Description string `json:"description"`
				Technology  string `json:"technology"`
				Status      string `json:"status"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{
					"code":    400,
					"message": "请求数据无效",
					"data":    nil,
				})
				return
			}

			storeMutex.Lock()
			if project, exists := projectStore[projectID]; exists {
				project.Name = req.Name
				project.Description = req.Description
				project.Technology = req.Technology
				if req.Status != "" {
					project.Status = req.Status
				}
				project.UpdatedAt = time.Now()
				storeMutex.Unlock()

				recordLog(1, "admin", "update_project", "project", id, fmt.Sprintf("更新项目: %s", req.Name), c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

				c.JSON(200, gin.H{
					"code":    200,
					"message": "项目更新成功",
					"data":    project,
				})
			} else {
				storeMutex.Unlock()
				c.JSON(404, gin.H{
					"code":    404,
					"message": "项目不存在",
					"data":    nil,
				})
			}
		})

		projects.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")
			projectID := uint(0)
			fmt.Sscanf(id, "%d", &projectID)

			storeMutex.Lock()
			if _, exists := projectStore[projectID]; exists {
				delete(projectStore, projectID)
				storeMutex.Unlock()

				recordLog(1, "admin", "delete_project", "project", id, fmt.Sprintf("删除项目 ID: %s", id), c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

				c.JSON(200, gin.H{
					"code":    200,
					"message": "项目删除成功",
					"data":    nil,
				})
			} else {
				storeMutex.Unlock()
				c.JSON(404, gin.H{
					"code":    404,
					"message": "项目不存在",
					"data":    nil,
				})
			}
		})
	}

	users := r.Group("/api/v1/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("", func(c *gin.Context) {
			storeMutex.RLock()
			defer storeMutex.RUnlock()

			userList := make([]*User, 0, len(userStore))
			for _, u := range userStore {
				userList = append(userList, u)
			}

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取用户列表成功",
				"data":    userList,
			})
		})

		users.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			userID := uint(0)
			fmt.Sscanf(id, "%d", &userID)

			storeMutex.RLock()
			defer storeMutex.RUnlock()

			for _, u := range userStore {
				if u.ID == userID {
					c.JSON(200, gin.H{
						"code":    200,
						"message": "获取用户信息成功",
						"data":    u,
					})
					return
				}
			}

			c.JSON(404, gin.H{
				"code":    404,
				"message": "用户不存在",
				"data":    nil,
			})
		})

		users.POST("", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required,min=2,max=20"`
				Password string `json:"password" binding:"required,min=4"`
				Email    string `json:"email"`
				Role     string `json:"role"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "Username") {
					errMsg = "用户名不能为空，且需要 2-20 个字符"
				} else if strings.Contains(errMsg, "Password") {
					errMsg = "密码不能为空，且至少需要 4 个字符"
				}
				c.JSON(400, gin.H{
					"code":    400,
					"message": errMsg,
					"data":    nil,
				})
				return
			}

			if req.Email == "" {
				req.Email = fmt.Sprintf("%s@localhost", req.Username)
			}

			storeMutex.Lock()
			if _, exists := userStore[req.Username]; exists {
				storeMutex.Unlock()
				c.JSON(409, gin.H{
					"code":    409,
					"message": "用户名已存在",
					"data":    nil,
				})
				return
			}

			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			newUserID := uint(len(userStore) + 2)
			if req.Role == "" {
				req.Role = "user"
			}

			newUser := &User{
				ID:       newUserID,
				Username: req.Username,
				Password: string(hashedPassword),
				Email:    req.Email,
				Role:     req.Role,
				Status:   "active",
			}
			userStore[req.Username] = newUser
			storeMutex.Unlock()

			recordLog(1, "admin", "create_user", "user", fmt.Sprintf("%d", newUserID), fmt.Sprintf("创建用户: %s", req.Username), c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

			c.JSON(201, gin.H{
				"code":    201,
				"message": "用户创建成功",
				"data":    newUser,
			})
		})

		users.PUT("/:id", func(c *gin.Context) {
			id := c.Param("id")
			userID := uint(0)
			fmt.Sscanf(id, "%d", &userID)

			var req struct {
				Email  string `json:"email"`
				Role   string `json:"role"`
				Status string `json:"status"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{
					"code":    400,
					"message": "请求数据无效",
					"data":    nil,
				})
				return
			}

			storeMutex.Lock()
			for username, u := range userStore {
				if u.ID == userID {
					if req.Email != "" {
						u.Email = req.Email
					}
					if req.Role != "" {
						u.Role = req.Role
					}
					if req.Status != "" {
						u.Status = req.Status
					}
					userStore[username] = u
					storeMutex.Unlock()

					recordLog(1, "admin", "update_user", "user", id, fmt.Sprintf("更新用户: %s", username), c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

					c.JSON(200, gin.H{
						"code":    200,
						"message": "用户更新成功",
						"data":    u,
					})
					return
				}
			}
			storeMutex.Unlock()

			c.JSON(404, gin.H{
				"code":    404,
				"message": "用户不存在",
				"data":    nil,
			})
		})

		users.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")
			userID := uint(0)
			fmt.Sscanf(id, "%d", &userID)

			storeMutex.Lock()
			for username, u := range userStore {
				if u.ID == userID {
					delete(userStore, username)
					storeMutex.Unlock()

					recordLog(1, "admin", "delete_user", "user", id, fmt.Sprintf("删除用户: %s", username), c.ClientIP(), c.GetHeader("User-Agent"), "success", 0)

					c.JSON(200, gin.H{
						"code":    200,
						"message": "用户删除成功",
						"data":    nil,
					})
					return
				}
			}
			storeMutex.Unlock()

			c.JSON(404, gin.H{
				"code":    404,
				"message": "用户不存在",
				"data":    nil,
			})
		})
	}

	operations := r.Group("/api/v1/operations")
	operations.Use(middleware.AuthMiddleware())
	{
		operations.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "healthy",
				"checks": []gin.H{
					{"name": "generator-service", "port": 8084, "status": "healthy"},
				},
			})
		})

		operations.GET("/stats", func(c *gin.Context) {
			metrics := healthMonitor.GetMetrics()

			storeMutex.RLock()
			activeProjects := len(projectStore)
			totalLogs := len(logStore)
			totalUsers := len(userStore)
			totalCodes := len(codeStore)
			storeMutex.RUnlock()

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取统计信息成功",
				"data": gin.H{
					"total_requests":     totalLogs,
					"success_rate":       99.8,
					"avg_response_time":  "12ms",
					"active_projects":    activeProjects,
					"total_logs":         totalLogs,
					"total_users":        totalUsers,
					"generated_codes":    totalCodes,
					"system_metrics":     metrics,
					"runtime": gin.H{
						"goroutines":        runtime.NumGoroutine(),
						"memory_alloc_mb":   float64(m.Alloc) / 1024 / 1024,
						"memory_sys_mb":     float64(m.Sys) / 1024 / 1024,
						"gc_collections":    m.NumGC,
						"uptime_seconds":    time.Since(startTime).Seconds(),
					},
				},
			})
		})

		operations.GET("/metrics", func(c *gin.Context) {
			prometheusFormat := healthMonitor.ExportPrometheusFormat()
			c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
			c.String(200, prometheusFormat)
		})

		operations.GET("/services", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取服务列表成功",
				"data": []gin.H{
					{"name": "generator-service", "port": 8084, "status": "running", "health": "healthy", "uptime": "active"},
				},
			})
		})

		operations.GET("/events", func(c *gin.Context) {
			storeMutex.RLock()
			recentLogs := make([]OperationLog, 0)
			if len(logStore) > 10 {
				recentLogs = logStore[len(logStore)-10:]
			} else {
				recentLogs = logStore
			}
			storeMutex.RUnlock()

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取事件列表成功",
				"data":    recentLogs,
			})
		})

		operations.GET("/operation-logs", func(c *gin.Context) {
			pageStr := c.DefaultQuery("page", "1")
			pageSizeStr := c.DefaultQuery("page_size", "20")

			actionFilter := c.Query("action")
			resourceFilter := c.Query("resource")
			statusFilter := c.Query("status")

			page := 1
			pageSize := 20
			fmt.Sscanf(pageStr, "%d", &page)
			fmt.Sscanf(pageSizeStr, "%d", &pageSize)

			if page < 1 {
				page = 1
			}
			if pageSize < 1 || pageSize > 100 {
				pageSize = 20
			}

			if db != nil {
				query := db.Model(&models.OperationLog{})

				if actionFilter != "" {
					query = query.Where("action = ?", actionFilter)
				}
				if resourceFilter != "" {
					query = query.Where("resource = ?", resourceFilter)
				}
				if statusFilter != "" {
					query = query.Where("status = ?", statusFilter)
				}

				var totalFiltered int64
				query.Count(&totalFiltered)

				var dbLogs []models.OperationLog
				offset := (page - 1) * pageSize
				result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&dbLogs)

				if result.Error != nil {
					log.Printf("[OperationLog] Database query error: %v", result.Error)
					c.JSON(500, gin.H{
						"code":    500,
						"message": "查询操作日志失败",
						"data":    nil,
					})
					return
				}

				logs := make([]OperationLog, len(dbLogs))
				for i, dbLog := range dbLogs {
					logs[i] = OperationLog{
						ID:        dbLog.ID,
						UserID:    dbLog.UserID,
						Username:  dbLog.Username,
						Action:    dbLog.Action,
						TargetType: dbLog.Resource,
						TargetID:  fmt.Sprintf("%d", dbLog.ResourceID),
						Details:   dbLog.Details,
						IPAddress: dbLog.IPAddress,
						UserAgent: dbLog.UserAgent,
						Status:    dbLog.Status,
						Duration:  dbLog.Duration,
						CreatedAt: dbLog.CreatedAt,
					}
				}

				totalPages := int((totalFiltered + int64(pageSize) - 1) / int64(pageSize))

				c.JSON(200, gin.H{
					"code":    200,
					"message": "获取操作日志成功",
					"data": gin.H{
						"items":       logs,
						"total":       totalFiltered,
						"page":        page,
						"page_size":   pageSize,
						"total_pages": totalPages,
						"filters": gin.H{
							"action":   actionFilter,
							"resource": resourceFilter,
							"status":   statusFilter,
						},
						"source": "database",
					},
				})
				return
			}

			storeMutex.RLock()

			var filteredLogs []OperationLog
			for _, log := range logStore {
				if actionFilter != "" && !strings.Contains(log.Action, actionFilter) {
					continue
				}
				if resourceFilter != "" && !strings.Contains(log.TargetType, resourceFilter) {
					continue
				}
				if statusFilter != "" && log.Status != statusFilter {
					continue
				}
				filteredLogs = append(filteredLogs, log)
			}

			for i, j := 0, len(filteredLogs)-1; i < j; i, j = i+1, j-1 {
				filteredLogs[i], filteredLogs[j] = filteredLogs[j], filteredLogs[i]
			}

			totalFiltered := len(filteredLogs)
			startIdx := (page - 1) * pageSize
			endIdx := startIdx + pageSize

			var paginatedLogs []OperationLog
			if startIdx < totalFiltered {
				if endIdx > totalFiltered {
					endIdx = totalFiltered
				}
				paginatedLogs = make([]OperationLog, endIdx-startIdx)
				copy(paginatedLogs, filteredLogs[startIdx:endIdx])
			} else {
				paginatedLogs = []OperationLog{}
			}
			totalPages := (totalFiltered + pageSize - 1) / pageSize
			storeMutex.RUnlock()

			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取操作日志成功",
				"data": gin.H{
					"items":       paginatedLogs,
					"total":       totalFiltered,
					"page":        page,
					"page_size":   pageSize,
					"total_pages": totalPages,
					"filters": gin.H{
						"action":   actionFilter,
						"resource": resourceFilter,
						"status":   statusFilter,
					},
					"source": "memory",
				},
			})
		})

		operations.POST("/operation-logs/record", func(c *gin.Context) {
			var req struct {
				Action     string `json:"action"`
				TargetType string `json:"target_type"`
				TargetID   string `json:"target_id"`
				Details    string `json:"details"`
				Status     string `json:"status"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{
					"code":    400,
					"message": "请求数据无效",
					"data":    nil,
				})
				return
			}

			if req.Status == "" {
				req.Status = "success"
			}

			recordLog(1, "admin", req.Action, req.TargetType, req.TargetID, req.Details, c.ClientIP(), c.GetHeader("User-Agent"), req.Status, 0)

			c.JSON(201, gin.H{
				"code":    201,
				"message": "操作日志记录成功",
				"data":    nil,
			})
		})
	}

	api := r.Group("/api/v1/generator")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/generate", generateCode)
		api.POST("/generate/:project_id", generateFromProject)
		api.GET("/download/:project_id", downloadZip)
		api.GET("/preview/:project_id", previewCode)
		api.POST("/docs/generate", generateDocumentation)

		api.POST("/metadata/register", func(c *gin.Context) {
			var req EntityMetadata
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			errors := metadataEngine.ValidateMetadata(&req)
			if len(errors) > 0 {
				c.JSON(400, gin.H{"errors": errors})
				return
			}

			if err := metadataEngine.RegisterModel(nil, &req); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(201, gin.H{"message": "Metadata registered successfully", "model_name": req.ModelName})
		})

		api.GET("/metadata/:model_name", func(c *gin.Context) {
			modelName := c.Param("model_name")
			metadata, exists := metadataEngine.GetMetadata(modelName)
			if !exists {
				c.JSON(404, gin.H{"error": "Model not found"})
				return
			}
			c.JSON(200, metadata)
		})

		api.GET("/metadata", func(c *gin.Context) {
			metadataList := metadataEngine.GetAllMetadata()
			c.JSON(200, metadataList)
		})

		api.GET("/metadata/:model_name/schema", func(c *gin.Context) {
			modelName := c.Param("model_name")
			schema, err := metadataEngine.GenerateDynamicFormSchema(modelName)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, schema)
		})

		api.GET("/metadata/:model_name/handlers", func(c *gin.Context) {
			modelName := c.Param("model_name")
			handlers, err := metadataEngine.GenerateCRUDHandlers(context.Background(), modelName)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, handlers)
		})

		api.GET("/metadata/:model_name/export", func(c *gin.Context) {
			modelName := c.Param("model_name")
			data, err := metadataEngine.ExportMetadataJSON(modelName)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.Data(200, "application/json", data)
		})

		api.POST("/metadata/import", func(c *gin.Context) {
			data, err := c.GetRawData()
			if err != nil {
				c.JSON(400, gin.H{"error": "Invalid JSON data"})
				return
			}
			metadata, err := metadataEngine.ImportMetadataJSON(data)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(201, metadata)
		})

		api.POST("/form/generate/:model_name", func(c *gin.Context) {
			modelName := c.Param("model_name")
			template, err := formEngine.GenerateFormFromMetadata(modelName)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, template)
		})

		api.POST("/form/react-code/:model_name", func(c *gin.Context) {
			modelName := c.Param("model_name")
			template, err := formEngine.GenerateFormFromMetadata(modelName)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			code, err := formEngine.GenerateReactCode(template)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"code": code, "language": "typescript"})
		})

		api.GET("/form/components", func(c *gin.Context) {
			components := formEngine.GetAllComponents()
			c.JSON(200, components)
		})

		api.GET("/form/components/categories", func(c *gin.Context) {
			categories := make(map[string][]ComponentDefinition)
			for _, comp := range formEngine.GetAllComponents() {
				categories[comp.Category] = append(categories[comp.Category], comp)
			}
			c.JSON(200, categories)
		})

		api.POST("/workflow/definitions", func(c *gin.Context) {
			var def WorkflowDefinition
			if err := c.ShouldBindJSON(&def); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			def.CreatedBy = 1
			if err := workflowEngine.RegisterDefinition(context.Background(), &def); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(201, gin.H{"message": "Workflow definition registered", "id": def.ID})
		})

		api.GET("/workflow/definitions", func(c *gin.Context) {
			entity := c.Query("entity")
			var defs []*WorkflowDefinition
			var err error

			if entity != "" {
				defs, err = workflowEngine.GetDefinitionsByEntity(entity)
			} else {
				err = workflowEngine.db.Where("status = ?", "active").Find(&defs).Error
			}

			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, defs)
		})

		api.GET("/workflow/definitions/:id", func(c *gin.Context) {
			id := c.Param("id")
			def, exists := workflowEngine.GetDefinition(id)
			if !exists {
				c.JSON(404, gin.H{"error": "Workflow definition not found"})
				return
			}
			c.JSON(200, def)
		})

		api.POST("/workflow/instances", func(c *gin.Context) {
			var req struct {
				DefinitionID string                 `json:"definition_id"`
				EntityID     string                 `json:"entity_id"`
				EntityType   string                 `json:"entity_type"`
				Initiator     uint                   `json:"initiator"`
				Variables    map[string]interface{} `json:"variables,omitempty"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			instance, err := workflowEngine.StartInstance(
				context.Background(),
				req.DefinitionID,
				req.EntityID,
				req.EntityType,
				req.Initiator,
				req.Variables,
			)

			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(201, instance)
		})

		api.GET("/workflow/instances/:id", func(c *gin.Context) {
			id := c.Param("id")
			instance, exists := workflowEngine.GetInstance(id)
			if !exists {
				c.JSON(404, gin.H{"error": "Workflow instance not found"})
				return
			}
			c.JSON(200, instance)
		})

		api.GET("/workflow/tasks/my", func(c *gin.Context) {
			userID := uint(1)
			status := c.Query("status")

			tasks, err := workflowEngine.GetTasksForUser(userID, status)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, tasks)
		})

		api.POST("/workflow/tasks/:task_id/complete", func(c *gin.Context) {
			taskID := c.Param("task_id")

			var req struct {
				Action  string                 `json:"action"`
				Comment string                 `json:"comment,omitempty"`
				Data    map[string]interface{} `json:"data,omitempty"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			instance, err := workflowEngine.CompleteTask(
				context.Background(),
				taskID,
				1,
				req.Action,
				req.Comment,
				req.Data,
			)

			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, instance)
		})

		api.GET("/workflow/export/:id", func(c *gin.Context) {
			id := c.Param("id")
			data, err := workflowEngine.ExportDefinitionJSON(id)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.Data(200, "application/json", data)
		})

		api.POST("/engine/stats", func(c *gin.Context) {
			metrics := healthMonitor.GetMetrics()
			metadataCount := len(metadataEngine.GetAllMetadata())

			c.JSON(200, gin.H{
				"metadata_models": metadataCount,
				"system_metrics": metrics,
				"uptime_seconds": metrics.Uptime,
				"goroutines":      metrics.Goroutines,
			})
		})
	}
}
