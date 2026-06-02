package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type FormEngine struct {
	metadataEngine    *MetadataEngine
	templates        map[string]FormTemplate
	componentRegistry *ComponentRegistry
}

type FormTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     int                    `json:"version"`
	Description string                 `json:"description"`
	Schema      FormSchema             `json:"schema"`
	Layout      FormLayout             `json:"layout"`
	Validation  FormValidation         `json:"validation"`
	I18n        map[string]string       `json:"i18n"`
	Extends     map[string]interface{} `json:"extends,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type FormSchema struct {
	Fields     []FormFieldDef  `json:"fields"`
	Groups     []FormGroupDef `json:"groups,omitempty"`
	Tabs       []FormTabDef   `json:"tabs,omitempty"`
	GridSystem GridSystem    `json:"grid_system"`
}

type FormFieldDef struct {
	Key          string          `json:"key"`
	Type         string          `json:"type"`
	Label        string          `json:"label"`
	Placeholder  string          `json:"placeholder,omitempty"`
	DefaultValue interface{}     `json:"default_value,omitempty"`
	Required     bool            `json:"required"`
	Readonly     bool            `json:"readonly"`
	Disabled     bool            `json:"disabled"`
	Hidden       bool            `json:"hidden"`
	Span         int             `json:"span"`
	Props        map[string]interface{} `json:"props,omitempty"`
	Rules        []FormRuleDef   `json:"rules,omitempty"`
	Options      []SelectOption  `json:"options,omitempty"`
	DependsOn    []string        `json:"depends_on,omitempty"`
	Conditional  *ConditionalConfig `json:"conditional,omitempty"`
	Slot         map[string]interface{} `json:"slot,omitempty"`
	CustomRender string          `json:"custom_render,omitempty"`
	Description  string          `json:"description,omitempty"`
	Prefix       string          `json:"prefix,omitempty"`
	Suffix       string          `json:"suffix,omitempty"`
}

type FormRuleDef struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Value     interface{} `json:"value,omitempty"`
	Trigger   string      `json:"trigger"`
	Validator string      `json:"validator,omitempty"`
}

type ConditionalConfig struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Action   string      `json:"action"`
	SetValue interface{} `json:"set_value,omitempty"`
}

type FormGroupDef struct {
	Title       string   `json:"title"`
	Key         string   `json:"key"`
	Fields      []string `json:"fields"`
	Collapsible bool     `json:"collapsible"`
	DefaultOpen bool     `json:"default_open"`
	Icon        string   `json:"icon,omitempty"`
}

type FormTabDef struct {
	Title  string       `json:"title"`
	Key    string       `json:"key"`
	Fields []string     `json:"fields"`
	Icon   string       `json:"icon,omitempty"`
	Badge  *BadgeConfig `json:"badge,omitempty"`
}

type BadgeConfig struct {
	Type   string      `json:"type"`
	Value  interface{} `json:"value"`
	Color  string      `json:"color"`
	Offset []int       `json:"offset"`
}

type FormLayout struct {
	Type            string          `json:"type"`
	LabelWidth      int             `json:"label_width"`
	LabelPosition   string          `json:"label_position"`
	Size            string          `json:"size"`
	Colon           bool            `json:"colon"`
	HideRequiredMark bool           `json:"hide_required_mark"`
	LabelAlign      string          `json:"label_align"`
	WrapperCol      WrapperColConfig `json:"wrapper_col"`
}

type WrapperColConfig struct {
	Span   int `json:"span"`
	Offset int `json:"offset"`
}

type FormValidation struct {
	Mode          string         `json:"mode"`
	ShowMessage   bool           `json:"show_message"`
	AutoValidate  bool           `json:"auto_validate"`
	ScrollToError bool           `json:"scroll_to_error"`
	GlobalRules   []FormRuleDef  `json:"global_rules,omitempty"`
}

type GridSystem struct {
	Columns int `json:"columns"`
	Gutter  int `json:"gutter"`
}

type ComponentRegistry struct {
	components map[string]ComponentDefinition
}

type ComponentDefinition struct {
	Name        string             `json:"name"`
	Type        string             `json:"type"`
	Category    string             `json:"category"`
	Icon        string             `json:"icon"`
	Description string             `json:"description"`
	Props       []ComponentPropDef `json:"props"`
	Events      []ComponentEventDef `json:"events,omitempty"`
	Slots       []ComponentSlotDef `json:"slots,omitempty"`
	DefaultProps map[string]interface{} `json:"default_props"`
	Example     string             `json:"example"`
	Version     string             `json:"version"`
}

type ComponentPropDef struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Description  string      `json:"description"`
	Options      []string    `json:"options,omitempty"`
}

type ComponentEventDef struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Params      []string `json:"params"`
}

type ComponentSlotDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       bool   `json:"scope"`
}

func NewFormEngine(metadataEngine *MetadataEngine) *FormEngine {
	engine := &FormEngine{
		metadataEngine:    metadataEngine,
		templates:        make(map[string]FormTemplate),
		componentRegistry: NewComponentRegistry(),
	}
	engine.registerDefaultComponents()
	return engine
}

func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]ComponentDefinition),
	}
}

func (f *FormEngine) registerDefaultComponents() {
	components := []ComponentDefinition{
		{
			Name: "Input", Type: "text", Category: "input", Icon: "edit",
			Description: "文本输入框",
			Props: []ComponentPropDef{
				{Name: "placeholder", Type: "string", DefaultValue: "请输入", Description: "占位文本"},
				{Name: "maxLength", Type: "number", DefaultValue: nil, Description: "最大长度"},
				{Name: "allowClear", Type: "boolean", DefaultValue: true, Description: "允许清除"},
			},
			Example: `<Input placeholder="请输入" />`,
		},
		{
			Name: "InputNumber", Type: "number", Category: "input", Icon: "number",
			Description: "数字输入框",
			Props: []ComponentPropDef{
				{Name: "min", Type: "number", DefaultValue: nil, Description: "最小值"},
				{Name: "max", Type: "number", DefaultValue: nil, Description: "最大值"},
				{Name: "step", Type: "number", DefaultValue: 1, Description: "步长"},
			},
			Example: `<InputNumber min={0} max={100} step={1} />`,
		},
		{
			Name: "Select", Type: "select", Category: "selection", Icon: "list",
			Description: "选择器",
			Props: []ComponentPropDef{
				{Name: "mode", Type: "string", DefaultValue: "", Description: "模式"},
				{Name: "options", Type: "array", DefaultValue: nil, Description: "选项列表"},
			},
			Example: `<Select options={options} />`,
		},
		{
			Name: "Switch", Type: "switch", Category: "selection", Icon: "swap",
			Description: "开关",
			Example: `<Switch checkedChildren="开" unCheckedChildren="关" />`,
		},
		{
			Name: "DatePicker", Type: "date", Category: "input", Icon: "calendar",
			Description: "日期选择器",
			Props: []ComponentPropDef{
				{Name: "format", Type: "string", DefaultValue: "YYYY-MM-DD", Description: "格式"},
				{Name: "showTime", Type: "boolean", DefaultValue: false, Description: "显示时间"},
			},
			Example: `<DatePicker format="YYYY-MM-DD" />`,
		},
		{
			Name: "Rate", Type: "rate", Category: "selection", Icon: "star",
			Description: "评分",
			Props: []ComponentPropDef{
				{Name: "count", Type: "number", DefaultValue: 5, Description: "星级数"},
				{Name: "allowHalf", Type: "boolean", DefaultValue: false, Description: "允许半选"},
			},
			Example: `<Rate count={5} allowHalf={true} />`,
		},
		{
			Name: "Upload", Type: "upload", Category: "advanced", Icon: "upload",
			Description: "上传组件",
			Props: []ComponentPropDef{
				{Name: "accept", Type: "string", DefaultValue: "", Description: "接受类型"},
				{Name: "maxCount", Type: "number", DefaultValue: nil, Description: "最大数量"},
			},
			Example: `<Upload accept=".jpg,.png" maxCount={5} />`,
		},
	}
	for _, comp := range components {
		f.componentRegistry.components[comp.Name] = comp
	}
}

func (r *ComponentRegistry) Register(comp ComponentDefinition) error {
	if _, exists := r.components[comp.Name]; exists {
		return fmt.Errorf("component %s already registered", comp.Name)
	}
	r.components[comp.Name] = comp
	return nil
}

func (r *ComponentRegistry) Get(name string) (*ComponentDefinition, bool) {
	comp, ok := r.components[name]
	return &comp, ok
}

func (r *ComponentRegistry) GetAll() []ComponentDefinition {
	result := make([]ComponentDefinition, 0, len(r.components))
	for _, comp := range r.components {
		result = append(result, comp)
	}
	return result
}

func (f *FormEngine) GenerateFormFromMetadata(modelName string) (*FormTemplate, error) {
	metadata, exists := f.metadataEngine.GetMetadata(modelName)
	if !exists {
		return nil, fmt.Errorf("model %s not found", modelName)
	}

	template := &FormTemplate{
		ID:          fmt.Sprintf("%s_form", strings.ToLower(modelName)),
		Name:        fmt.Sprintf("%s Form", modelName),
		Version:     1,
		Description: fmt.Sprintf("Auto-generated form for %s", modelName),
		Schema: FormSchema{
			Fields:     f.convertFieldsToFormFields(metadata.Fields),
			GridSystem: GridSystem{Columns: 24, Gutter: 16},
		},
		Layout: FormLayout{
			Type:        "vertical",
			LabelWidth:  120,
			LabelPosition: "top",
			Size:        "middle",
			Colon:       true,
		},
		Validation: FormValidation{
			Mode:         "inline",
			ShowMessage: true,
			AutoValidate: false,
		},
		I18n: map[string]string{
			"submit":          "提交",
			"cancel":          "取消",
			"reset":           "重置",
			"required":        "此字段为必填项",
			"validation_error": "验证失败",
			"save_success":    "保存成功",
			"save_failed":     "保存失败",
		},
	}
	return template, nil
}

func (f *FormEngine) convertFieldsToFormFields(fields []FieldMetadata) []FormFieldDef {
	formFields := make([]FormFieldDef, len(fields))
	for i, field := range fields {
		componentType := f.resolveComponentFromFieldType(field.Type, field.FormUI.Component)
		rules := make([]FormRuleDef, 0)
		if !field.IsNullable && !field.IsPrimaryKey {
			rules = append(rules, FormRuleDef{
				Type:    "required",
				Message: fmt.Sprintf("%s 是必填字段", field.FormUI.Label),
			})
		}
		formFields[i] = FormFieldDef{
			Key:          field.Name,
			Type:         componentType,
			Label:        field.FormUI.Label,
			Placeholder:  field.FormUI.Placeholder,
			DefaultValue: field.DefaultValue,
			Required:     !field.IsNullable,
			Readonly:     field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at",
			Hidden:       field.Name == "deleted_at",
			Span:         field.FormUI.Grid.Md,
			Props:        f.buildComponentProps(field, componentType),
			Rules:        rules,
			Options:      field.FormUI.Options,
			Description:  field.Comment,
		}
	}
	return formFields
}

func (f *FormEngine) resolveComponentFromFieldType(fieldType, preferredComponent string) string {
	if preferredComponent != "" && preferredComponent != "" {
		return preferredComponent
	}
	typeMap := map[string]string{
		"string": "Input", "varchar": "Input", "text": "Textarea",
		"int": "InputNumber", "integer": "InputNumber", "bigint": "InputNumber",
		"float": "InputNumber", "double": "InputNumber", "decimal": "InputNumber",
		"bool": "Switch", "boolean": "Switch",
		"date": "DatePicker", "datetime": "DatePicker", "timestamp": "DatePicker",
		"time": "TimePicker", "enum": "Select", "select": "Select",
		"email": "Input", "phone": "Input", "password": "Password",
		"image": "Upload", "file": "Upload", "color": "ColorPicker",
		"rate": "Rate", "slider": "Slider",
	}
	if comp, ok := typeMap[strings.ToLower(fieldType)]; ok {
		return comp
	}
	return "Input"
}

func (f *FormEngine) buildComponentProps(field FieldMetadata, componentType string) map[string]interface{} {
	props := make(map[string]interface{})
	if field.FormUI.Props != nil {
		props = field.FormUI.Props
	}
	switch componentType {
	case "Input":
		if props["placeholder"] == nil {
			props["placeholder"] = field.FormUI.Placeholder
		}
	case "InputNumber":
		if props["precision"] == nil && strings.Contains(strings.ToLower(field.Type), "decimal") {
			props["precision"] = 2
		}
	case "DatePicker":
		if props["format"] == nil {
			if strings.Contains(strings.ToLower(field.Type), "time") {
				props["showTime"] = true
				props["format"] = "YYYY-MM-DD HH:mm:ss"
			} else {
				props["format"] = "YYYY-MM-DD"
			}
		}
	case "Upload":
		if props["accept"] == nil {
			switch strings.ToLower(field.Type) {
			case "image":
				props["accept"] = ".jpg,.jpeg,.png,.gif,.webp"
				props["listType"] = "picture-card"
			default:
				props["accept"] = "*"
			}
		}
	case "Rate":
		if props["count"] == nil { props["count"] = 5 }
	case "Slider":
		if props["min"] == nil { props["min"] = 0 }
		if props["max"] == nil { props["max"] = 100 }
	}
	return props
}

func (f *FormEngine) GenerateReactCode(template *FormTemplate) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`// Auto-generated form: %s
import React from 'react';
import { Form, Input, Select, DatePicker, Switch, Upload, Rate, Slider } from 'antd';

interface %sFormData {
%s}

const %sForm: React.FC<%sFormProps> = ({ initialValues, onSubmit, onCancel }) => {
  const [form] = Form.useForm();
  return (
    <Form form={form} layout="%s" initialValues={initialValues} onFinish={onSubmit}>
%s
    </Form>
  );
};

export default %sForm;
`, template.Name, template.ID, generateInterface(template.Schema.Fields), template.ID, template.ID, template.Layout.Type, generateFormItems(template), template.ID))
	return sb.String(), nil
}

func generateInterface(fields []FormFieldDef) string {
	var sb strings.Builder
	for _, field := range fields {
		goType := inferTypeScriptType(field.Type)
		optional := ""
		if !field.Required { optional = "?" }
		sb.WriteString(fmt.Sprintf("  %s%s: %s;\n", field.Key, optional, goType))
	}
	return sb.String()
}

func inferTypeScriptType(componentType string) string {
	typeMap := map[string]string{
		"Input": "string", "Password": "string", "Textarea": "string",
		"InputNumber": "number", "Select": "string | number",
		"Switch": "boolean", "DatePicker": "string",
		"Rate": "number", "Slider": "number | [number, number]",
		"Upload": "any[]", "ColorPicker": "string",
	}
	if tsType, ok := typeMap[componentType]; ok { return tsType }
	return "any"
}

func generateFormItems(template *FormTemplate) string {
	var sb strings.Builder
	for _, field := range template.Schema.Fields {
		sb.WriteString(fmt.Sprintf(`      <Form.Item name="%s" label="%s" rules={{[{{ required: %v, message: '%s' }}]}}>\n`, field.Key, field.Label, field.Required, field.Label))
		sb.WriteString(fmt.Sprintf("        <%s />\n", field.Type))
		sb.WriteString("      </Form.Item>\n")
	}
	return sb.String()
}

func (f *FormEngine) ExportTemplateJSON(templateID string) ([]byte, error) {
	template, ok := f.templates[templateID]
	if !ok { return nil, fmt.Errorf("template %s not found", templateID) }
	return json.MarshalIndent(template, "", "  ")
}

func (f *FormEngine) ImportTemplateJSON(data []byte) (*FormTemplate, error) {
	var template FormTemplate
	if err := json.Unmarshal(data, &template); err != nil { return nil, err }
	f.templates[template.ID] = template
	log.Printf("[FormEngine] Imported template: %s", template.Name)
	return &template, nil
}

func (f *FormEngine) ValidateTemplate(template *FormTemplate) []string {
	var errors []string
	if template.Name == "" { errors = append(errors, "template name is required") }
	if len(template.Schema.Fields) == 0 { errors = append(errors, "at least one field is required") }
	keys := make(map[string]bool)
	for i, field := range template.Schema.Fields {
		if field.Key == "" { errors = append(errors, fmt.Sprintf("field[%d]: key is required", i)) }
		if field.Type == "" { errors = append(errors, fmt.Sprintf("field[%d]: type is required", i)) }
		if keys[field.Key] { errors = append(errors, fmt.Sprintf("duplicate field key: %s", field.Key)) }
		keys[field.Key] = true
	}
	return errors
}

func (f *FormEngine) GetAllComponents() []ComponentDefinition {
	return f.componentRegistry.GetAll()
}
