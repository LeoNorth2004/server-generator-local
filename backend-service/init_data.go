package main

import (
	"log"
	"time"

	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/models"
	"golang.org/x/crypto/bcrypt"
)

func initDefaultUser() {
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Users already exist, skipping default user creation")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash default password: %v", err)
		return
	}

	defaultUser := models.User{
		Username: "admin",
		Password: string(hashedPassword),
		Email:    "admin@example.com",
		Role:     models.RoleAdmin,
	}

	if err := database.DB.Create(&defaultUser).Error; err != nil {
		log.Printf("Failed to create default user: %v", err)
		return
	}

	log.Println("✓ Default admin user created successfully")
	log.Println("  Username: admin")
	log.Println("  Password: admin123")
}

func initSampleData() {
	initSampleProjects()
	initSampleOperationLogs()
}

func initSampleProjects() {
	var count int64
	database.DB.Model(&models.Project{}).Count(&count)
	if count > 0 {
		return
	}

	sampleProjects := []models.Project{
		{
			UserID:        1,
			Name:          "示例电商系统",
			Description:   "一个完整的电商平台，包含商品管理、订单处理、用户系统等模块",
			GeneratedCode: `{"files": {}}`,
		},
		{
			UserID:        1,
			Name:          "博客管理系统",
			Description:   "支持多用户博客发布、评论、标签管理等功能的CMS系统",
			GeneratedCode: `{"files": {}}`,
		},
		{
			UserID:        1,
			Name:          "任务管理系统",
			Description:   "企业级任务跟踪和项目管理工具",
			GeneratedCode: "",
		},
	}

	for _, project := range sampleProjects {
		if err := database.DB.Create(&project).Error; err != nil {
			log.Printf("Failed to create sample project: %v", err)
		}
	}
	log.Println("✓ Sample projects created")
}

func initSampleOperationLogs() {
	var count int64
	database.DB.Model(&models.OperationLog{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now()
	sampleLogs := []models.OperationLog{
		{
			UserID:    1,
			Username:  "admin",
			Action:    "system_start",
			Resource:  "system",
			Details:   `{"event": "系统初始化完成", "services": 10}`,
			Status:    "success",
			Duration:  1250,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-120 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "login",
			Resource:  "auth",
			Details:   `{"method": "密码登录", "role": "admin"}`,
			Status:    "success",
			Duration:  85,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-95 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "refresh",
			Resource:  "operations",
			Details:   `{"action": "refresh_monitor"}`,
			Status:    "success",
			Duration:  230,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-60 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "login",
			Resource:  "auth",
			Details:   `{"method": "密码登录", "error": "用户名或密码错误"}`,
			Status:    "failed",
			Duration:  120,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-30 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "login",
			Resource:  "auth",
			Details:   `{"method": "密码登录", "role": "admin"}`,
			Status:    "success",
			Duration:  78,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-29 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "keep_alive",
			Resource:  "system",
			Details:   `{"enabled": true}`,
			Status:    "success",
			Duration:  15,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-15 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "restart_node",
			Resource:  "cluster",
			Details:   `{"node_name": "agent-0", "action": "restart"}`,
			Status:    "success",
			Duration:  3200,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-5 * time.Minute),
		},
		{
			UserID:    1,
			Username:  "admin",
			Action:    "refresh",
			Resource:  "operations",
			Details:   `{"action": "refresh_monitor", "page": "operations"}`,
			Status:    "success",
			Duration:  195,
			IPAddress: "127.0.0.1",
			CreatedAt: now.Add(-2 * time.Minute),
		},
		{
			UserID:    0,
			Username:  "anonymous",
			Action:    "login",
			Resource:  "auth",
			Details:   `{"method": "密码登录", "error": "认证失败: invalid credentials"}`,
			Status:    "failed",
			Duration:  95,
			IPAddress: "192.168.1.100",
			CreatedAt: now.Add(-45 * time.Second),
		},
	}

	for _, logEntry := range sampleLogs {
		if err := database.DB.Create(&logEntry).Error; err != nil {
			log.Printf("Failed to create sample log: %v", err)
		}
	}
	log.Println("✓ Sample operation logs created")
}
