package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"gorm.io/gorm"
)

type MetadataEngine struct {
	db            *gorm.DB
	models        map[string]interface{}
	metadataCache sync.Map
	routeRegistry *RouteRegistry
}

func NewMetadataEngine(db *gorm.DB) *MetadataEngine {
	return &MetadataEngine{
		db:            db,
		models:        make(map[string]interface{}),
		routeRegistry: NewRouteRegistry(),
	}
}

type EntityMetadata struct {
	TableName    string                 `json:"table_name"`
	ModelName    string                 `json:"model_name"`
	Fields       []FieldMetadata        `json:"fields"`
	Relations    []RelationMetadata     `json:"relations,omitempty"`
	Indexes      []IndexMetadata        `json:"indexes,omitempty"`
	Validation   ValidationRules       `json:"validation"`
	FormConfig   FormFieldConfig        `json:"form_config"`
	Workflow     *WorkflowConfig        `json:"workflow,omitempty"`
	CustomAPIs   []CustomAPIConfig      `json:"custom_apis,omitempty"`
	Extends      map[string]interface{} `json:"extends,omitempty"`
}

type FieldMetadata struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	GoType       string      `json:"go_type"`
	DBType       string      `json:"db_type"`
	IsPrimaryKey bool        `json:"is_primary_key"`
	IsNullable   bool        `json:"is_nullable"`
	IsUnique     bool        `json:"is_unique"`
	DefaultValue interface{} `json:"default_value"`
	Comment      string      `json:"comment"`
	Tags         map[string]string `json:"tags"`
	FormUI       FieldUIConfig `json:"form_ui"`
}

type FieldUIConfig struct {
	Component  string                 `json:"component"`
	Label      string                 `json:"label"`
	Placeholder string                `json:"placeholder,omitempty"`
	Options    []SelectOption         `json:"options,omitempty"`
	Rules      []FieldRule            `json:"rules,omitempty"`
	Props      map[string]interface{} `json:"props,omitempty"`
	Grid       GridConfig             `json:"grid"`
	Visible    bool                   `json:"visible"`
	Editable   bool                   `json:"editable"`
	Sortable   bool                   `json:"sortable"`
}

type SelectOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type FieldRule struct {
	Type    string `json:"type"` // required, min, max, pattern, custom
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

type GridConfig struct {
	Xs int `json:"xs"` // extra-small
	Sm int `json:"sm"` // small
	Md int `json:"md"` // medium
	Lg int `json:"lg"` // large
}

type RelationMetadata struct {
	Type        string `json:"type"` // has_many, belongs_to, many_to_many
	ForeignTable string `json:"foreign_table"`
	ForeignKey   string `json:"foreign_key"`
	References  string `json:"references"`
}

type IndexMetadata struct {
	Name    string   `json:"name"`
	Fields  []string `json:"fields"`
	Unique  bool     `json:"unique"`
}

type ValidationRules struct {
	GlobalRules []RuleConfig `json:"global_rules"`
	FieldRules  map[string][]RuleConfig `json:"field_rules"`
}

type RuleConfig struct {
	When   string      `json:"when,omitempty"` // create, update, delete
	Action string      `json:"action"`          // required, unique, custom
	Params interface{} `json:"params,omitempty"`
}

type FormFieldConfig struct {
	Layout       string          `json:"layout"` // vertical, horizontal, inline
	LabelWidth   int             `json:"label_width"`
	LabelPosition string         `json:"label_position"`
	Size         string          `json:"size"` // small, medium, large
	Groups       []FormGroup     `json:"groups,omitempty"`
	Tabs         []FormTab       `json:"tabs,omitempty"`
}

type FormGroup struct {
	Title  string   `json:"title"`
	Fields []string `json:"fields"`
	Collapsible bool `json:"collapsible"`
}

type FormTab struct {
	Title  string   `json:"title"`
	Fields []string `json:"fields"`
	Icon   string   `json:"icon,omitempty"`
}

type WorkflowConfig struct {
	Enabled    bool           `json:"enabled"`
	Nodes      []WorkflowNode `json:"nodes"`
	Edges      []WorkflowEdge `json:"edges"`
	Version    int            `json:"version"`
}

type WorkflowNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // start, end, approval, condition, action, parallel
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	Position    Position               `json:"position"`
}

type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
	Condition string `json:"condition,omitempty"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type CustomAPIConfig struct {
	Method      string                 `json:"method"` // GET, POST, PUT, DELETE
	Path        string                 `json:"path"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Handler     string                 `json:"handler"`
	Params      []APIParam             `json:"params"`
	Middlewares []string               `json:"middlewares,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type APIParam struct {
	Name     string `json:"name"`
	In       string `json:"in"` // query, path, body, header
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func (e *MetadataEngine) RegisterModel(model interface{}, metadata *EntityMetadata) error {
 modelName := metadata.ModelName

 e.models[modelName] = model
 e.metadataCache.Store(modelName, metadata)

 log.Printf("[MetadataEngine] Registered model: %s (table: %s)", modelName, metadata.TableName)

 return nil
}

func (e *MetadataEngine) GetMetadata(modelName string) (*EntityMetadata, bool) {
 if val, ok := e.metadataCache.Load(modelName); ok {
  return val.(*EntityMetadata), true
 }
 return nil, false
}

func (e *MetadataEngine) GetAllMetadata() []*EntityMetadata {
 var result []*EntityMetadata
 e.metadataCache.Range(func(key, value interface{}) bool {
  result = append(result, value.(*EntityMetadata))
  return true
 })
 return result
}

func (e *MetadataEngine) GenerateCRUDHandlers(ctx context.Context, modelName string) (*GeneratedHandlers, error) {
 metadata, exists := e.GetMetadata(modelName)
 if !exists {
  return nil, fmt.Errorf("model %s not found", modelName)
 }

 handlers := &GeneratedHandlers{
  ModelName: modelName,
  Metadata:  metadata,
  Operations: make(map[string]HandlerInfo),
 }

 for _, op := range []string{"Create", "Read", "Update", "Delete", "List", "Search", "BatchDelete"} {
  handler := e.generateHandler(ctx, metadata, op)
  if handler != nil {
   handler.Name = fmt.Sprintf("%s%s", modelName, op)
   handler.Path = e.buildRoutePath(metadata.TableName, strings.ToLower(op))
   handler.Method = e.getHTTPMethod(op)
   handlers.Operations[op] = *handler
  }
 }

 return handlers, nil
}

type GeneratedHandlers struct {
 ModelName  string                 `json:"model_name"`
 Metadata   *EntityMetadata        `json:"metadata"`
 Operations map[string]HandlerInfo `json:"operations"`
}

type HandlerInfo struct {
 Name        string                 `json:"name"`
 Path        string                 `json:"path"`
 Method      string                 `json:"method"`
	Description string                `json:"description"`
	HandlerFunc string                 `json:"handler_func"`
	RequestType string                 `json:"request_type"`
	ResponseType string                `json:"response_type"`
	Middlewares []string               `json:"middlewares"`
	Params      []HandlerParam         `json:"params"`
	Config      map[string]interface{} `json:"config"`
}

type HandlerParam struct {
 Name     string `json:"name"`
 Type     string `json:"name"`
 In       string `json:"in"` // path, query, body
 Required bool   `json:"required"`
}

func (e *MetadataEngine) generateHandler(ctx context.Context, metadata *EntityMetadata, operation string) *HandlerInfo {
 switch operation {
 case "Create":
  return e.generateCreateHandler(metadata)
 case "Read":
  return e.generateReadHandler(metadata)
 case "Update":
  return e.generateUpdateHandler(metadata)
 case "Delete":
  return e.generateDeleteHandler(metadata)
 case "List":
  return e.generateListHandler(metadata)
 case "Search":
  return e.generateSearchHandler(metadata)
 case "BatchDelete":
  return e.generateBatchDeleteHandler(metadata)
 default:
  return nil
 }
}

func (e *MetadataEngine) generateCreateHandler(metadata *EntityMetadata) *HandlerInfo {
 fields := make([]HandlerParam, 0)
 for _, field := range metadata.Fields {
  if !field.IsPrimaryKey && field.FormUI.Editable {
   fields = append(fields, HandlerParam{
    Name:     field.Name,
    Type:     field.GoType,
    In:       "body",
    Required: !field.IsNullable,
   })
  }
 }

 return &HandlerInfo{
  Description: fmt.Sprintf("Create new %s", metadata.ModelName),
  RequestType:  fmt.Sprintf("Create%sRequest", metadata.ModelName),
  ResponseType: fmt.Sprintf("*models.%s", metadata.ModelName),
  Params:      fields,
  Config: map[string]interface{}{
   "validate": true,
   "hooks":    []string{"before_create", "after_create"},
  },
 }
}

func (e *MetadataEngine) generateReadHandler(metadata *EntityMetadata) *HandlerInfo {
 return &HandlerInfo{
  Description: fmt.Sprintf("Get %s by ID", metadata.ModelName),
  RequestType:  "nil",
  ResponseType: fmt.Sprintf("*models.%s", metadata.ModelName),
  Params: []HandlerParam{
   {Name: "id", Type: "uint", In: "path", Required: true},
  },
  Config: map[string]interface{}{
   "cache": map[string]interface{}{
    "enabled": true,
    "ttl":    300,
   },
  },
 }
}

func (e *MetadataEngine) generateUpdateHandler(metadata *EntityMetadata) *HandlerInfo {
 fields := make([]HandlerParam, 0)
 for _, field := range metadata.Fields {
  if !field.IsPrimaryKey && field.FormUI.Editable {
   fields = append(fields, HandlerParam{
    Name:     field.Name,
    Type:     field.GoType,
    In:       "body",
    Required: false,
   })
  }
 }

 return &HandlerInfo{
  Description: fmt.Sprintf("Update %s by ID", metadata.ModelName),
  RequestType:  fmt.Sprintf("Update%sRequest", metadata.ModelName),
  ResponseType: fmt.Sprintf("*models.%s", metadata.ModelName),
  Params: append([]HandlerParam{
   {Name: "id", Type: "uint", In: "path", Required: true},
  }, fields...),
  Config: map[string]interface{}{
   "validate": true,
   "hooks":    []string{"before_update", "after_update"},
   "optimistic_lock": true,
  },
 }
}

func (e *MetadataEngine) generateDeleteHandler(metadata *EntityMetadata) *HandlerInfo {
 return &HandlerInfo{
  Description: fmt.Sprintf("Delete %s by ID", metadata.ModelName),
  RequestType:  "nil",
  ResponseType: "map[string]interface{}",
  Params: []HandlerParam{
   {Name: "id", Type: "uint", In: "path", Required: true},
  },
  Config: map[string]interface{}{
   "soft_delete": true,
   "hooks":      []string{"before_delete", "after_delete"},
  },
 }
}

func (e *MetadataEngine) generateListHandler(metadata *EntityMetadata) *HandlerInfo {
 return &HandlerInfo{
  Description: fmt.Sprintf("List all %s with pagination", metadata.ModelName),
  RequestType:  "ListRequest",
  ResponseType: fmt.Sprintf("*%sListResponse", metadata.ModelName),
  Params: []HandlerParam{
   {Name: "page", Type: "int", In: "query", Required: false},
   {Name: "page_size", Type: "int", In: "query", Required: false},
   {Name: "sort", Type: "string", In: "query", Required: false},
   {Name: "order", Type: "string", In: "query", Required: false},
  },
  Config: map[string]interface{}{
   "pagination": map[string]interface{}{
    "default_page":     1,
    "defaultPageSize": 10,
    "maxPageSize":     100,
   },
   "cache": map[string]interface{}{
    "enabled": true,
    "ttl":    60,
   },
  },
 }
}

func (e *MetadataEngine) generateSearchHandler(metadata *EntityMetadata) *HandlerInfo {
 searchableFields := make([]HandlerParam, 0)
 for _, field := range metadata.Fields {
  if field.FormUI.Sortable {
   searchableFields = append(searchableFields, HandlerParam{
    Name: field.Name,
    Type: field.GoType,
    In:   "query",
   })
  }
 }

 return &HandlerInfo{
  Description: fmt.Sprintf("Search %s with filters", metadata.ModelName),
  RequestType:  "SearchRequest",
  ResponseType: fmt.Sprintf("*%sListResponse", metadata.ModelName),
  Params: append([]HandlerParam{
   {Name: "keyword", Type: "string", In: "query", Required: false},
   {Name: "page", Type: "int", In: "query", Required: false},
   {Name: "page_size", Type: "int", In: "query", Required: false},
  }, searchableFields...),
  Config: map[string]interface{}{
   "fulltext_search": true,
   "filters":         true,
  },
 }
}

func (e *MetadataEngine) generateBatchDeleteHandler(metadata *EntityMetadata) *HandlerInfo {
 return &HandlerInfo{
  Description: fmt.Sprintf("Batch delete %s by IDs", metadata.ModelName),
  RequestType:  "BatchDeleteRequest",
  ResponseType: "map[string]interface{}",
  Params: []HandlerParam{
   {Name: "ids", Type: "[]uint", In: "body", Required: true},
  },
  Config: map[string]interface{}{
   "soft_delete": true,
   "batch_size": 100,
  },
 }
}

func (e *MetadataEngine) buildRoutePath(tableName, operation string) string {
 base := fmt.Sprintf("/api/v1/%s", tableName)
 switch operation {
 case "create":
  return base
 case "read":
  return base + "/:id"
 case "update":
  return base + "/:id"
 case "delete":
  return base + "/:id"
 case "list":
  return base
 case "search":
  return base + "/search"
 case "batch_delete":
  return base + "/batch"
 default:
  return base
 }
}

func (e *MetadataEngine) getHTTPMethod(operation string) string {
 switch operation {
 case "create":
  return "POST"
 case "read":
  return "GET"
 case "update":
  return "PUT"
 case "delete", "batch_delete":
  return "DELETE"
 case "list", "search":
  return "GET"
 default:
  return "GET"
 }
}

func (e *MetadataEngine) GenerateDynamicFormSchema(modelName string) (map[string]interface{}, error) {
 metadata, exists := e.GetMetadata(modelName)
 if !exists {
  return nil, fmt.Errorf("model %s not found", modelName)
 }

 schema := map[string]interface{}{
  "schema_version": "1.0",
  "form_id":        fmt.Sprintf("%s_form", strings.ToLower(modelName)),
  "model_name":     modelName,
  "table_name":     metadata.TableName,
  "title":          fmt.Sprintf("%s Form", modelName),
  "layout":         metadata.FormConfig.Layout,
  "fields":         e.generateFormFieldSchemas(metadata.Fields),
  "groups":         metadata.FormConfig.Groups,
  "tabs":           metadata.FormConfig.Tabs,
  "validation":     metadata.Validation,
  "workflow":       metadata.Workflow,
  "actions": []map[string]interface{}{
   {"type": "submit", "label": "Save", "primary": true},
   {"type": "reset", "label": "Reset"},
   {"type": "cancel", "label": "Cancel"},
  },
  "i18n": map[string]string{
   "save":              "保存",
   "reset":            "重置",
   "cancel":           "取消",
   "required_field":   "此字段为必填项",
   "invalid_format":   "格式不正确",
   "success_message":  "保存成功",
   "error_message":    "保存失败",
  },
 }

 return schema, nil
}

func (e *MetadataEngine) generateFormFieldSchemas(fields []FieldMetadata) []map[string]interface{} {
 schemas := make([]map[string]interface{}, len(fields))

 for i, field := range fields {
  schemas[i] = map[string]interface{}{
   "name":        field.Name,
   "type":        field.Type,
   "go_type":     field.GoType,
   "label":       field.FormUI.Label,
   "placeholder": field.FormUI.Placeholder,
   "required":    !field.IsNullable,
   "default_value": field.DefaultValue,
   "component":   e.resolveComponent(field.Type, field.FormUI.Component),
   "grid":        field.FormUI.Grid,
   "visible":     field.FormUI.Visible,
   "editable":    field.FormUI.Editable,
   "sortable":    field.FormUI.Sortable,
   "options":     field.FormUI.Options,
   "rules":       field.FormUI.Rules,
   "props":       field.FormUI.Props,
   "comment":     field.Comment,
  }
 }

 return schemas
}

func (e *MetadataEngine) resolveComponent(fieldType, preferredComponent string) string {
 if preferredComponent != "" {
  return preferredComponent
 }

 componentMap := map[string]string{
  "string":      "Input",
  "text":        "Textarea",
  "int":         "InputNumber",
  "integer":     "InputNumber",
  "float":       "InputNumber",
  "double":      "InputNumber",
  "decimal":     "InputNumber",
  "bool":        "Switch",
  "boolean":     "Switch",
  "date":        "DatePicker",
  "datetime":    "DateTimePicker",
  "time":        "TimePicker",
  "enum":        "Select",
  "select":      "Select",
  "json":        "JSONEditor",
  "jsonb":       "JSONEditor",
  "uuid":        "Input",
  "email":       "Input",
  "phone":       "Input",
  "url":         "Input",
  "password":    "Password",
  "image":       "ImageUpload",
  "file":        "FileUpload",
  "rich_text":   "RichTextEditor",
  "color":       "ColorPicker",
  "rate":        "Rate",
  "slider":      "Slider",
  "tags":       "TagInput",
  "array":      "ArrayInput",
 }

 if comp, ok := componentMap[strings.ToLower(fieldType)]; ok {
  return comp
 }

 return "Input"
}

func (e *MetadataEngine) ExportMetadataJSON(modelName string) ([]byte, error) {
 metadata, exists := e.GetMetadata(modelName)
 if !exists {
  return nil, fmt.Errorf("model %s not found", modelName)
 }

 data := map[string]interface{}{
  "version":      "1.0.0",
  "generated_at": "auto",
  "metadata":     metadata,
  "handlers": func() interface{} {
   handlers, _ := e.GenerateCRUDHandlers(context.Background(), modelName)
   return handlers
  }(),
  "form_schema": func() interface{} {
   schema, _ := e.GenerateDynamicFormSchema(modelName)
   return schema
  }(),
 }

 return json.MarshalIndent(data, "", "  ")
}

func (e *MetadataEngine) ImportMetadataJSON(data []byte) (*EntityMetadata, error) {
 var wrapper struct {
  Metadata *EntityMetadata `json:"metadata"`
 }

 if err := json.Unmarshal(data, &wrapper); err != nil {
  return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
 }

 if wrapper.Metadata == nil {
  return nil, fmt.Errorf("metadata is required in JSON")
 }

 err := e.RegisterModel(nil, wrapper.Metadata)
 if err != nil {
  return nil, fmt.Errorf("failed to register imported metadata: %w", err)
 }

 log.Printf("[MetadataEngine] Imported metadata for model: %s", wrapper.Metadata.ModelName)

 return wrapper.Metadata, nil
}

func (e *MetadataEngine) ValidateMetadata(metadata *EntityMetadata) []string {
 var errors []string

 if metadata.ModelName == "" {
  errors = append(errors, "model_name is required")
 }

 if metadata.TableName == "" {
  errors = append(errors, "table_name is required")
 }

 if len(metadata.Fields) == 0 {
  errors = append(errors, "at least one field is required")
 } else {
  hasPrimaryKey := false
  for i, field := range metadata.Fields {
   if field.Name == "" {
    errors = append(errors, fmt.Sprintf("field[%d]: name is required", i))
   }
   if field.Type == "" {
    errors = append(errors, fmt.Sprintf("field[%d]: type is required", i))
   }
   if field.IsPrimaryKey {
    hasPrimaryKey = true
   }
  }

  if !hasPrimaryKey {
   errors = append(errors, "at least one primary key field is required")
  }
 }

 return errors
}

func (e *MetadataEngine) ReflectModel(model interface{}) (*EntityMetadata, error) {
 t := reflect.TypeOf(model)
 if t.Kind() == reflect.Ptr {
  t = t.Elem()
 }

 if t.Kind() != reflect.Struct {
  return nil, fmt.Errorf("model must be a struct or pointer to struct")
 }

 metadata := &EntityMetadata{
  ModelName: t.Name(),
  Fields:   make([]FieldMetadata, 0),
  Validation: ValidationRules{
   GlobalRules: make([]RuleConfig, 0),
   FieldRules:  make(map[string][]RuleConfig),
  },
  FormConfig: FormFieldConfig{
   Layout: "vertical",
  },
 }

 for i := 0; i < t.NumField(); i++ {
  field := t.Field(i)

  if field.Anonymous {
   continue
  }

  fieldMeta := FieldMetadata{
   Name:       field.Name,
   Type:       field.Type.Name(),
   GoType:     field.Type.String(),
   Comment:    field.Tag.Get("comment"),
   Tags:       parseGormTags(field.Tag.Get("gorm")),
   FormUI: FieldUIConfig{
    Component: resolveDefaultComponent(field.Type.Name()),
    Label:     field.Name,
    Visible:   true,
    Editable:  field.Name != "id" && field.Name != "created_at" && field.Name != "updated_at",
    Sortable:  true,
    Grid: GridConfig{Xs: 24, Sm: 12, Md: 8, Lg: 6},
   },
  }

  jsonTag := field.Tag.Get("json")
  if jsonTag != "" && jsonTag != "-" {
   parts := strings.Split(jsonTag, ",")
   if len(parts) > 0 && parts[0] != "" {
    fieldMeta.Name = parts[0]
   }
  }

  gormTag := field.Tag.Get("gorm")
  if gormTag != "" {
   tags := parseGormTags(gormTag)
   if v, ok := tags["primarykey"]; ok && v != "" {
    fieldMeta.IsPrimaryKey = true
   }
   if v, ok := tags["unique"]; ok && v != "" {
    fieldMeta.IsUnique = true
   }
   if v, ok := tags["not null"]; ok && v != "" {
    fieldMeta.IsNullable = false
   } else {
    fieldMeta.IsNullable = true
   }
   if v, ok := tags["column"]; ok && v != "" {
    fieldMeta.Name = v
   }
   if v, ok := tags["default"]; ok && v != "" {
    fieldMeta.DefaultValue = v
   }
   if v, ok := tags["type"]; ok && v != "" {
    fieldMeta.DBType = v
   }
  }

  metadata.Fields = append(metadata.Fields, fieldMeta)
 }

 return metadata, nil
}

func parseGormTags(tag string) map[string]string {
 result := make(map[string]string)
 parts := strings.Split(tag, ";")

 for _, part := range parts {
  part = strings.TrimSpace(part)
  kv := strings.SplitN(part, ":", 2)
  if len(kv) == 2 {
   result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
  } else if len(kv) == 1 && kv[0] != "" {
   result[kv[0]] = ""
  }
 }

 return result
}

func resolveDefaultComponent(goType string) string {
 typeMap := map[string]string{
  "string":  "Input",
  "int":     "InputNumber",
  "int64":   "InputNumber",
  "float64": "InputNumber",
  "bool":    "Switch",
  "time.Time": "DatePicker",
 }

 if comp, ok := typeMap[goType]; ok {
  return comp
 }

 return "Input"
}
