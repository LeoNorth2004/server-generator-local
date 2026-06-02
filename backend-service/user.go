package main

import (
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/models"
	"github.com/generator-platform/go-common/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func createUserHandler(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.InternalServerError(c, "Failed to hash password")
		return
	}

	role := models.RoleUser
	if req.Role == "admin" {
		role = models.RoleAdmin
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Role:     role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		response.BadRequest(c, "Username already exists")
		return
	}

	response.Success(c, user)
}

func getUserHandler(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

func updateUserHandler(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := database.DB.First(&user, id).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Role != "" {
		if req.Role == "admin" {
			user.Role = models.RoleAdmin
		} else {
			user.Role = models.RoleUser
		}
	}

	if err := database.DB.Save(&user).Error; err != nil {
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.Success(c, user)
}

func deleteUserHandler(c *gin.Context) {
	id := c.Param("id")

	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		response.InternalServerError(c, "Failed to delete user")
		return
	}

	response.Success(c, nil)
}

func listUsersHandler(c *gin.Context) {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch users")
		return
	}

	response.Success(c, users)
}
