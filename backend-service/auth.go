package main

import (
	"fmt"
	"time"

	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/models"
	"github.com/generator-platform/go-common/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("low-code-platform-secret-key-2024")

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func generateToken(userID uint, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "generator-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func loginHandler(c *gin.Context) {
	startTime := time.Now()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		recordAuthLog(c, "login", req.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.BadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		recordAuthLog(c, "login", req.Username, "failed", "Invalid username or password", time.Since(startTime).Milliseconds())
		response.Unauthorized(c, "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		recordAuthLog(c, "login", req.Username, "failed", "Invalid password", time.Since(startTime).Milliseconds())
		response.Unauthorized(c, "Invalid username or password")
		return
	}

	token, err := generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		recordAuthLog(c, "login", user.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	recordAuthLog(c, "login", user.Username, "success", "", time.Since(startTime).Milliseconds())

	response.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

func registerHandler(c *gin.Context) {
	startTime := time.Now()
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		recordAuthLog(c, "register", req.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.BadRequest(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		recordAuthLog(c, "register", req.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.InternalServerError(c, "Failed to hash password")
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Role:     models.RoleUser,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		recordAuthLog(c, "register", req.Username, "failed", "Username already exists", time.Since(startTime).Milliseconds())
		response.BadRequest(c, "Username already exists")
		return
	}

	token, err := generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		recordAuthLog(c, "register", user.Username, "failed", err.Error(), time.Since(startTime).Milliseconds())
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	recordAuthLog(c, "register", user.Username, "success", fmt.Sprintf("User ID: %d, Email: %s", user.ID, user.Email), time.Since(startTime).Milliseconds())

	response.Success(c, LoginResponse{
		Token: token,
		User:  user,
	})
}

func getCurrentUserHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}
