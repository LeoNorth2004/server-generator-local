package main

import (
	"github.com/generator-platform/go-common/database"
	"github.com/generator-platform/go-common/models"
	"github.com/generator-platform/go-common/response"
	"github.com/gin-gonic/gin"
)

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	DBConfig    string `json:"db_config"`
	TableConfig string `json:"table_config"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DBConfig    string `json:"db_config"`
	TableConfig string `json:"table_config"`
}

func createProjectHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project := models.Project{
		UserID:      userID.(uint),
		Name:        req.Name,
		Description: req.Description,
		DBConfig:    req.DBConfig,
		TableConfig: req.TableConfig,
	}

	if err := database.DB.Create(&project).Error; err != nil {
		response.InternalServerError(c, "Failed to create project")
		return
	}

	response.Success(c, project)
}

func getProjectHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := c.Param("id")

	var project models.Project
	query := database.DB.Where("id = ?", id)

	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&project).Error; err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	response.Success(c, project)
}

func updateProjectHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := c.Param("id")

	var project models.Project
	query := database.DB.Where("id = ?", id)

	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&project).Error; err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.DBConfig != "" {
		project.DBConfig = req.DBConfig
	}
	if req.TableConfig != "" {
		project.TableConfig = req.TableConfig
	}

	if err := database.DB.Save(&project).Error; err != nil {
		response.InternalServerError(c, "Failed to update project")
		return
	}

	response.Success(c, project)
}

func deleteProjectHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := c.Param("id")

	query := database.DB.Where("id = ?", id)

	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Delete(&models.Project{}).Error; err != nil {
		response.InternalServerError(c, "Failed to delete project")
		return
	}

	response.Success(c, nil)
}

func listProjectsHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var projects []models.Project
	query := database.DB.Preload("User")

	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&projects).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch projects")
		return
	}

	response.Success(c, projects)
}
