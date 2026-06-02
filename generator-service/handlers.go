package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/generator-platform/go-common/models"
	"github.com/gin-gonic/gin"
)

func recordOperationLog(c *gin.Context, action, resource string, resourceID uint, details interface{}, status, errorMsg string, duration int64) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	uid, _ := userID.(uint)
	uname, _ := username.(string)
	if uname == "" {
		uname = "admin"
	}

	detailsJSON, _ := json.Marshal(details)

	recordLog(uid, uname, action, resource, fmt.Sprintf("%d", resourceID), string(detailsJSON), c.ClientIP(), c.Request.UserAgent(), status, duration)

	log.Printf("[OPERATION] User=%s Action=%s Resource=%s Status=%s Duration=%dms",
		username, action, resource, status, duration)
}

func generateCode(c *gin.Context) {
	startTime := time.Now()

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	userIDValue, _ := c.Get("user_id")
	usernameValue, _ := c.Get("username")

	currentUserID := uint(0)
	currentUsername := "anonymous"

	if uid, ok := userIDValue.(uint); ok {
		currentUserID = uid
	}
	if uname, ok := usernameValue.(string); ok {
		currentUsername = uname
	}

	log.Printf("[DEBUG] User: %s (ID: %d)", currentUsername, currentUserID)

	log.Printf("[DEBUG] Received request: %+v", req)
	log.Printf("[DEBUG] Number of tables: %d", len(req.Tables))
	for i, table := range req.Tables {
		log.Printf("[DEBUG] Table %d: %s, Number of fields: %d", i, table.Name, len(table.Fields))
		for j, field := range table.Fields {
			log.Printf("[DEBUG]   Field %d: %s, Type: %s", j, field.Name, field.Type)
		}
	}

	if req.ProjectName == "" {
		badRequest(c, "Project name is required")
		return
	}
	if len(req.Tables) == 0 {
		badRequest(c, "At least one table is required")
		return
	}
	for i, table := range req.Tables {
		if table.Name == "" {
			badRequest(c, fmt.Sprintf("Table %d: Name is required", i+1))
			return
		}
		if len(table.Fields) == 0 {
			badRequest(c, fmt.Sprintf("Table %s: At least one field is required", table.Name))
			return
		}
		for j, field := range table.Fields {
			if field.Name == "" {
				badRequest(c, fmt.Sprintf("Table %s, Field %d: Name is required", table.Name, j+1))
				return
			}
			if field.Type == "" {
				badRequest(c, fmt.Sprintf("Table %s, Field %d: Type is required", table.Name, j+1))
				return
			}
		}
	}

	generated, err := doGenerate(req)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "generate", "project", 0, gin.H{
			"project_name": req.ProjectName,
			"tables_count": len(req.Tables),
			"error":        err.Error(),
		}, "failed", err.Error(), duration)
		internalServerError(c, err.Error())
		return
	}

	storeMutex.Lock()
	projectIDCounter++
	codeIDCounter++
	projectID := projectIDCounter
	codeID := codeIDCounter

	tableConfigJSON, _ := json.Marshal(req.Tables)
	dbConfigJSON, _ := json.Marshal(req.DBConfig)

	project := &Project{
		ID:           projectID,
		Name:         req.ProjectName,
		Description:  fmt.Sprintf("自动生成 - %d 个表", len(req.Tables)),
		Technology:   "Go/Gin/GORM",
		Status:       "generated",
		TableConfig:  string(tableConfigJSON),
		DBConfig:     string(dbConfigJSON),
		CreatedBy:    1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	projectStore[project.ID] = project

	codeEntry := &CodeEntry{
		ID:          codeID,
		ProjectID:   projectID,
		ProjectName: req.ProjectName,
		Code:        marshalGeneratedCode(generated),
		Language:    "go",
		FileType:    "zip",
		CreatedAt:   time.Now(),
	}
	codeStore[codeID] = codeEntry
	storeMutex.Unlock()

	generatedCodeJSON, _ := json.Marshal(generated)
	if db != nil {
		dbProject := models.Project{
			UserID:        currentUserID,
			Name:          req.ProjectName,
			Description:   fmt.Sprintf("自动生成 - %d 个表", len(req.Tables)),
			DBConfig:      string(dbConfigJSON),
			TableConfig:   string(tableConfigJSON),
			GeneratedCode: string(generatedCodeJSON),
			Status:        "generated",
		}

		if result := db.Create(&dbProject); result.Error != nil {
			log.Printf("[WARNING] Failed to save project to database: %v", result.Error)
		} else {
			log.Printf("[Database] Project saved to PostgreSQL - User: %s (ID: %d), Project ID: %d", currentUsername, currentUserID, dbProject.ID)
			project.DBID = dbProject.ID
		}
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "generate", "project", projectID, gin.H{
		"project_name":    req.ProjectName,
		"tables_count":    len(req.Tables),
		"files_generated": len(generated.Files),
	}, "success", "", duration)

	successResponse(c, gin.H{
		"code":       generated,
		"project_id": projectID,
	})
}

func generateFromProject(c *gin.Context) {
	startTime := time.Now()
	projectID := c.Param("project_id")
	pid, _ := strconv.ParseUint(projectID, 10, 32)

	storeMutex.RLock()
	codeEntry, exists := codeStore[uint(pid)]
	storeMutex.RUnlock()

	if !exists {
		if db != nil {
			var dbProject models.Project
			result := db.First(&dbProject, pid)
			if result.Error == nil && dbProject.GeneratedCode != "" {
				log.Printf("[Database] Loading project from DB: %s (ID: %d)", dbProject.Name, pid)

				newCodeEntry := &CodeEntry{
					ID:          nextCodeID(),
					ProjectID:   uint(pid),
					ProjectName: dbProject.Name,
					Code:        dbProject.GeneratedCode,
					Language:    "go",
					FileType:    "zip",
					CreatedAt:   time.Now(),
				}

				storeMutex.Lock()
				codeStore[uint(pid)] = newCodeEntry
				storeMutex.Unlock()

				codeEntry = newCodeEntry
				exists = true
				log.Printf("[Database] Project loaded into memory cache")
			}
		}

		if !exists {
			duration := time.Since(startTime).Milliseconds()
			recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
				"project_id": projectID,
				"error":      "Generated code not found",
			}, "failed", "Generated code not found", duration)
			notFound(c, "Generated code not found. Please generate first.")
			return
		}
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "regenerate", "project", uint(pid), gin.H{
		"project_id":   projectID,
		"project_name": codeEntry.ProjectName,
	}, "success", "", duration)

	var generated GeneratedCode
	if err := json.Unmarshal([]byte(codeEntry.Code), &generated); err != nil {
		internalServerError(c, "Failed to parse generated code")
		return
	}

	successResponse(c, gin.H{
		"project_id": pid,
		"code":       generated,
	})
}

func downloadZip(c *gin.Context) {
	startTime := time.Now()
	projectID := c.Param("project_id")
	pid, _ := strconv.ParseUint(projectID, 10, 32)

	storeMutex.RLock()
	codeEntry, exists := codeStore[uint(pid)]
	storeMutex.RUnlock()

	if !exists {
		if db != nil {
			var dbProject models.Project
			result := db.First(&dbProject, pid)
			if result.Error == nil && dbProject.GeneratedCode != "" {
				log.Printf("[Database] Loading project from DB for download: %s (ID: %d)", dbProject.Name, pid)

				newCodeEntry := &CodeEntry{
					ID:          nextCodeID(),
					ProjectID:   uint(pid),
					ProjectName: dbProject.Name,
					Code:        dbProject.GeneratedCode,
					Language:    "go",
					FileType:    "zip",
					CreatedAt:   time.Now(),
				}

				storeMutex.Lock()
				codeStore[uint(pid)] = newCodeEntry
				storeMutex.Unlock()

				codeEntry = newCodeEntry
				exists = true
				log.Printf("[Database] Project loaded into memory cache for download")
			}
		}

		if !exists {
			duration := time.Since(startTime).Milliseconds()
			recordOperationLog(c, "download", "project", uint(pid), gin.H{
				"project_id": projectID,
				"error":      "Generated code not found",
			}, "failed", "Generated code not found", duration)
			notFound(c, "Generated code not found. Please generate the code first.")
			return
		}
	}

	log.Printf("[DEBUG] Download request for project: %s (ID: %s)", codeEntry.ProjectName, projectID)

	var generated GeneratedCode
	if err := json.Unmarshal([]byte(codeEntry.Code), &generated); err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_name": codeEntry.ProjectName,
			"error":        "Invalid generated code",
		}, "failed", "Invalid generated code", duration)
		badRequest(c, "Invalid generated code data")
		return
	}

	log.Printf("[DEBUG] Generated files count: %d", len(generated.Files))

	if len(generated.Files) == 0 {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_name": codeEntry.ProjectName,
			"error":        "No files to download",
		}, "failed", "No files to download", duration)
		badRequest(c, "No generated files available for download")
		return
	}

	zipBytes, err := generateZip(&generated)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "download", "project", uint(pid), gin.H{
			"project_name": codeEntry.ProjectName,
			"error":        err.Error(),
		}, "failed", err.Error(), duration)
		internalServerError(c, "Failed to generate zip file")
		return
	}

	log.Printf("[DEBUG] ZIP file size: %d bytes", len(zipBytes))

	if len(zipBytes) == 0 {
		internalServerError(c, "Generated zip file is empty")
		return
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "download", "project", uint(pid), gin.H{
		"project_name": codeEntry.ProjectName,
		"files_count":  len(generated.Files),
		"zip_size":     len(zipBytes),
	}, "success", "", duration)

	safeName := sanitizeFilename(codeEntry.ProjectName)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", safeName))
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(200, "application/zip", zipBytes)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	re := regexp.MustCompile(`[^\w\-.]`)
	name = re.ReplaceAllString(name, "")
	if len(name) > 100 {
		name = name[:100]
	}
	if name == "" {
		name = "project"
	}
	return name
}

func previewCode(c *gin.Context) {
	startTime := time.Now()
	projectID := c.Param("project_id")
	filePath := c.Query("file")
	pid, _ := strconv.ParseUint(projectID, 10, 32)

	storeMutex.RLock()
	codeEntry, exists := codeStore[uint(pid)]
	storeMutex.RUnlock()

	if !exists {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_id": projectID,
			"error":      "Generated code not found",
		}, "failed", "Generated code not found", duration)
		notFound(c, "Generated code not found. Please generate the code first.")
		return
	}

	var generated GeneratedCode
	if err := json.Unmarshal([]byte(codeEntry.Code), &generated); err != nil {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_name": codeEntry.ProjectName,
			"error":        "Invalid generated code",
		}, "failed", "Invalid generated code", duration)
		badRequest(c, "Invalid generated code")
		return
	}

	if filePath == "" {
		files := make([]string, 0, len(generated.Files))
		for path := range generated.Files {
			files = append(files, path)
		}

		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_name": codeEntry.ProjectName,
			"action":       "list_files",
			"files_count":  len(files),
		}, "success", "", duration)

		successResponse(c, gin.H{
			"files": files,
		})
		return
	}

	content, exists := generated.Files[filePath]
	if !exists {
		duration := time.Since(startTime).Milliseconds()
		recordOperationLog(c, "preview", "project", uint(pid), gin.H{
			"project_name": codeEntry.ProjectName,
			"file_path":    filePath,
			"error":        "File not found",
		}, "failed", "File not found", duration)
		notFound(c, "File not found")
		return
	}

	duration := time.Since(startTime).Milliseconds()
	recordOperationLog(c, "preview", "project", uint(pid), gin.H{
		"project_name":   codeEntry.ProjectName,
		"file_path":      filePath,
		"content_length": len(content),
	}, "success", "", duration)

	successResponse(c, gin.H{
		"path":    filePath,
		"content": content,
	})
}

func generateDocumentation(c *gin.Context) {
	var req GenerateDocsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err.Error())
		return
	}

	docs, err := generateDocs(req)
	if err != nil {
		internalServerError(c, err.Error())
		return
	}

	successResponse(c, gin.H{
		"docs":   docs,
		"format": req.Format,
	})
}

func generateDocs(req GenerateDocsRequest) (map[string]string, error) {
	docs := make(map[string]string)

	switch req.DocType {
	case "api":
		docs["api_documentation.md"] = generateAPIDocs(req.ProjectName, req.IncludeExamples, req.IncludeComments)
	case "config":
		docs["config_guide.md"] = generateConfigGuide()
	case "dev":
		docs["development_guide.md"] = generateDevelopmentGuide(req.ProjectName, req.Tables)
	default:
		docs["api_documentation.md"] = generateAPIDocs(req.ProjectName, req.IncludeExamples, req.IncludeComments)
		docs["config_guide.md"] = generateConfigGuide()
		docs["development_guide.md"] = generateDevelopmentGuide(req.ProjectName, req.Tables)
	}

	return docs, nil
}

func generateZip(code *GeneratedCode) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for path, content := range code.Files {
		f, err := w.Create(path)
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func badRequest(c *gin.Context, msg string) {
	c.JSON(400, gin.H{"code": 400, "message": msg, "data": nil})
}

func internalServerError(c *gin.Context, msg string) {
	c.JSON(500, gin.H{"code": 500, "message": msg, "data": nil})
}

func notFound(c *gin.Context, msg string) {
	c.JSON(404, gin.H{"code": 404, "message": msg, "data": nil})
}

func successResponse(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"code": 200, "message": "success", "data": data})
}
