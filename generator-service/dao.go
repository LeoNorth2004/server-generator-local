package main

import (
	"fmt"
	"strings"
)

func generateDAO(projectName string, tables []TableConfig) string {
	var sb strings.Builder
	
	// 基础DAO代码
	sb.WriteString(fmt.Sprintf(`package dao

import (
	"context"
	"%s/internal/models"
	"gorm.io/gorm"
)

type DAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) *DAO {
	return &DAO{db: db}
}

func (d *DAO) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "tx", tx))
	})
}

func (d *DAO) DB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return d.db.WithContext(ctx)
}

`, projectName))

	// 为每个表生成CRUD方法
	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		varName := toLowerCamelCase(table.Name)

		// 使用显式的字符串拼接而不是复杂的fmt.Sprintf
		sb.WriteString("\n\ntype " + modelName + "DAO struct {\n")
		sb.WriteString("\t*DAO\n}\n\n")
		
		// New function
		sb.WriteString(fmt.Sprintf("func (d *DAO) New%sDAO() *%sDAO {\n", modelName, modelName))
		sb.WriteString(fmt.Sprintf("\treturn &%sDAO{DAO: d}\n}\n", modelName))
		
		// Create function
		sb.WriteString(fmt.Sprintf("\nfunc (dao *%sDAO) Create(ctx context.Context, %s *models.%s) error {\n", modelName, varName, modelName))
		sb.WriteString(fmt.Sprintf("\treturn dao.DB(ctx).Create(%s).Error\n}\n", varName))
		
		// GetByID function
		sb.WriteString(fmt.Sprintf("\nfunc (dao *%sDAO) GetByID(ctx context.Context, id uint) (*models.%s, error) {\n", modelName, modelName))
		sb.WriteString(fmt.Sprintf("\tvar %s models.%s\n", varName, modelName))
		sb.WriteString(fmt.Sprintf("\terr := dao.DB(ctx).First(&%s, id).Error\n", varName))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, err\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\treturn &%s, nil\n}\n", varName))
		
		// List function
		sb.WriteString(fmt.Sprintf("\nfunc (dao *%sDAO) List(ctx context.Context, page, pageSize int) ([]*models.%s, int64, error) {\n", modelName, modelName))
		sb.WriteString(fmt.Sprintf("\tvar %sList []*models.%s\n", varName, modelName))
		sb.WriteString("\tvar total int64\n\n")
		sb.WriteString(fmt.Sprintf("\tdb := dao.DB(ctx).Model(&models.%s{})\n", modelName))
		sb.WriteString("\tdb.Count(&total)\n\n")
		sb.WriteString("\toffset := (page - 1) * pageSize\n")
		sb.WriteString(fmt.Sprintf("\terr := db.Offset(offset).Limit(pageSize).Find(&%sList).Error\n", varName))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString("\t\treturn nil, 0, err\n")
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\treturn %sList, total, nil\n}\n", varName))
		
		// Update function
		sb.WriteString(fmt.Sprintf("\nfunc (dao *%sDAO) Update(ctx context.Context, %s *models.%s) error {\n", modelName, varName, modelName))
		sb.WriteString(fmt.Sprintf("\treturn dao.DB(ctx).Save(%s).Error\n}\n", varName))
		
		// Delete function
		sb.WriteString(fmt.Sprintf("\nfunc (dao *%sDAO) Delete(ctx context.Context, id uint) error {\n", modelName))
		sb.WriteString(fmt.Sprintf("\treturn dao.DB(ctx).Delete(&models.%s{}, id).Error\n}\n", modelName))
	}

	return sb.String()
}

func generateDAOGen(tables []TableConfig) string {
	var sb strings.Builder
	sb.WriteString(`package dao

import (
	"gorm.io/gen"
	"gorm.io/gorm"
)

type GenQuerier interface {
	GetByID(id int64) (gen.T, error)
	
	FindAll() ([]gen.T, error)
}

func GenerateDAO(db *gorm.DB) error {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./internal/dao/query",
		Mode:    gen.WithDefaultQuery,
	})
	
	g.UseDB(db)
	
	g.ApplyBasic(
`)

	for _, table := range tables {
		modelName := toCamelCase(table.Name)
		sb.WriteString(fmt.Sprintf("\t\tg.GenerateModelAs(\"%s\", \"%s\"),\n", table.Name, modelName))
	}

	sb.WriteString(`	)
	
	g.Execute()
	return nil
}
`)
	return sb.String()
}
