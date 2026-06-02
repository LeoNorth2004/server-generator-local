package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

type WorkflowEngine struct {
	db           *gorm.DB
	definitions  map[string]*WorkflowDefinition
	instances    map[string]*WorkflowInstance
	executors    map[string]WorkflowExecutor
	mu           sync.RWMutex
}

type WorkflowDefinition struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"not null"`
	Version     int                    `json:"version" gorm:"default:1"`
	Description string                 `json:"description"`
	Entity      string                 `json:"entity" gorm:"index"` // which model this workflow applies to
	Status      string                 `json:"status" gorm:"default:draft"` // draft, active, archived
	Nodes       []WorkflowNodeDef      `json:"nodes" gorm:"type:jsonb"`
	Edges       []WorkflowEdgeDef      `json:"edges" gorm:"type:jsonb"`
	Variables   []WorkflowVariable     `json:"variables" gorm:"type:jsonb"`
	Config      WorkflowDefConfig      `json:"config" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   uint                   `json:"created_by"`
}

type WorkflowNodeDef struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // start, end, approval, condition, action, parallel, timer, notification
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config"`
	Position    Position               `json:"position"`
	Assignee    *AssigneeConfig        `json:"assignee,omitempty"`
	Conditions  []ConditionRule        `json:"conditions,omitempty"`
	TimeLimit   *TimeLimitConfig       `json:"time_limit,omitempty"`
}

type AssigneeConfig struct {
	Type     string   `json:"type"` // user, role, field, expression, initiator, supervisor
	Value    []string `json:"value"`
	Field    string   `json:"field,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type ConditionRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, neq, gt, gte, lt, lte, in, not_in, contains, is_empty, is_not_empty
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic"` // and, or
}

type TimeLimitConfig struct {
	Duration int    `json:"duration"` // in hours
	Unit     string `json:"unit"`     // hour, day, week
	Action   string `json:"action"`   // auto_approve, auto_reject, notify, escalate
}

type WorkflowEdgeDef struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Label     string `json:"label,omitempty"`
	Condition string `json:"condition,omitempty"` // expression for conditional routing
	Priority  int    `json:"priority"`          // for multiple outgoing edges
}

type WorkflowVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // string, number, boolean, date, user, entity
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Description  string      `json:"description"`
	Scope        string      `json:"scope"` // global, instance, node
}

type WorkflowDefConfig struct {
	AutoStart       bool                   `json:"auto_start"`
	AllowRecall     bool                   `json:"allow_recall"`
	AllowDelegate   bool                   `json:"allow_delegate"`
	AllowUrgent     bool                   `json:"allow_urgent"`
	NotifyOnComplete bool                  `json:"notify_on_complete"`
	NotifyOnReject  bool                   `json:"notify_on_reject"`
	CustomActions   []CustomActionConfig   `json:"custom_actions,omitempty"`
	Hooks           []WorkflowHookConfig   `json:"hooks,omitempty"`
}

type CustomActionConfig struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Type        string                 `json:"type"` // approve, reject, return, delegate, add_signer, transfer, comment, urgent
	Config      map[string]interface{} `json:"config,omitempty"`
	Permissions []string               `json:"permissions"`
}

type WorkflowHookConfig struct {
	Event   string `json:"event"` // before_start, after_complete, on_approval, on_rejection, on_timeout
	Action  string `json:"action"` // call_api, send_notification, update_field, create_entity
	Config  string `json:"config"`
	Async   bool   `json:"async"`
}

type WorkflowInstance struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	DefinitionID  string                 `json:"definition_id" gorm:"index;not null"`
	EntityID      string                 `json:"entity_id" gorm:"index;not null"`
	EntityType    string                 `json:"entity_type" gorm:"index"`
	Status        string                 `json:"status" gorm:"index"` // running, completed, rejected, cancelled, suspended
	CurrentNodes  []string               `json:"current_nodes" gorm:"type:jsonb"`
	History       []WorkflowHistoryEntry `json:"history" gorm:"type:jsonb"`
	Variables     map[string]interface{} `json:"variables" gorm:"type:jsonb"`
	Initiator     uint                   `json:"initiator"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type WorkflowHistoryEntry struct {
	NodeID     string                 `json:"node_id"`
	NodeName   string                 `json:"node_name"`
	Action     string                 `json:"action"` // enter, leave, approve, reject, return, delegate, timeout
	Actor      uint                   `json:"actor"`
	ActorName  string                 `json:"actor_name"`
	Comment    string                 `json:"comment"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Duration   int64                  `json:"duration_ms"`
}

type WorkflowTask struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	InstanceID  string                 `json:"instance_id" gorm:"index;not null"`
	NodeID      string                 `json:"node_id" gorm:"index"`
	NodeName    string                 `json:"node_name"`
	Assignee    uint                   `json:"assignee" gorm:"index"`
	Status      string                 `json:"status" gorm:"index"` // pending, approved, rejected, delegated, returned, cancelled
	Priority    int                    `json:"priority" gorm:"default:0"`
	DueDate     *time.Time             `json:"due_date"`
	CompletedAt *time.Time             `json:"completed_at"`
	Comment     string                 `json:"comment"`
	Data        map[string]interface{} `json:"data" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type WorkflowExecutor interface {
 CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error)
 Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error)
 OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error
 OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error
 Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error
}

type ExecutionResult struct {
 Success     bool                   `json:"success"`
 Action      string                 `json:"action"` // approve, reject, return, delegate, next, complete
 NextNodes   []string               `json:"next_nodes,omitempty"`
 Data        map[string]interface{} `json:"data,omitempty"`
	Message     string                 `json:"message,omitempty"`
	Error       error                  `json:"-"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
}

func NewWorkflowEngine(db *gorm.DB) *WorkflowEngine {
 engine := &WorkflowEngine{
  db:        db,
  definitions: make(map[string]*WorkflowDefinition),
  instances:  make(map[string]*WorkflowInstance),
  executors:  make(map[string]WorkflowExecutor),
 }

 engine.registerDefaultExecutors()

 return engine
}

func (w *WorkflowEngine) registerDefaultExecutors() {
 w.executors["start"] = &StartNodeExecutor{}
 w.executors["end"] = &EndNodeExecutor{}
 w.executors["approval"] = &ApprovalNodeExecutor{db: w.db}
 w.executors["condition"] = &ConditionNodeExecutor{}
 w.executors["action"] = &ActionNodeExecutor{}
 w.executors["parallel"] = &ParallelNodeExecutor{}
 w.executors["timer"] = &TimerNodeExecutor{}
 w.executors["notification"] = &NotificationNodeExecutor{}
}

func (w *WorkflowEngine) RegisterDefinition(ctx context.Context, def *WorkflowDefinition) error {
 w.mu.Lock()
 defer w.mu.Unlock()

 if err := w.validateDefinition(def); err != nil {
  return fmt.Errorf("invalid workflow definition: %w", err)
 }

 if def.ID == "" {
  def.ID = generateUUID()
 }

 if def.Status == "" {
  def.Status = "draft"
 }

 if err := w.db.Create(def).Error; err != nil {
  return fmt.Errorf("failed to save workflow definition: %w", err)
 }

 w.definitions[def.ID] = def

 log.Printf("[WorkflowEngine] Registered definition: %s (v%d)", def.Name, def.Version)

 return nil
}

func (w *WorkflowEngine) GetDefinition(id string) (*WorkflowDefinition, bool) {
 w.mu.RLock()
 defer w.mu.RUnlock()

 def, ok := w.definitions[id]
 if !ok {
  var dbDef WorkflowDefinition
  if err := w.db.First(&dbDef, "id = ?", id).Error; err != nil {
   return nil, false
  }
  w.definitions[id] = &dbDef
  return &dbDef, true
 }
 return def, true
}

func (w *WorkflowEngine) StartInstance(ctx context.Context, definitionID, entityID, entityType string, initiator uint, variables map[string]interface{}) (*WorkflowInstance, error) {
 def, exists := w.GetDefinition(definitionID)
 if !exists {
  return nil, fmt.Errorf("workflow definition %s not found", definitionID)
 }

 if def.Status != "active" {
  return nil, fmt.Errorf("workflow definition %s is not active (status: %s)", definitionID, def.Status)
 }

 startNode := findStartNode(def.Nodes)
 if startNode == nil {
  return nil, fmt.Errorf("workflow definition %s has no start node", definitionID)
 }

 instance := &WorkflowInstance{
  ID:           generateUUID(),
  DefinitionID: definitionID,
  EntityID:     entityID,
  EntityType:   entityType,
  Status:       "running",
  CurrentNodes: []string{startNode.ID},
  Variables:    make(map[string]interface{}),
  Initiator:    initiator,
  StartedAt:    time.Now(),
  History: []WorkflowHistoryEntry{
   {
    NodeID:    startNode.ID,
    NodeName:  startNode.Name,
    Action:    "enter",
    Actor:     initiator,
    Timestamp: time.Now(),
   },
  },
 }

 if variables != nil {
  for k, v := range variables {
   instance.Variables[k] = v
  }
 }

 if err := w.db.Create(instance).Error; err != nil {
  return nil, fmt.Errorf("failed to create workflow instance: %w", err)
 }

 w.instances[instance.ID] = instance

 executor, ok := w.executors[startNode.Type]
 if !ok {
  return nil, fmt.Errorf("no executor for node type: %s", startNode.Type)
 }

 if err := executor.OnEnter(ctx, startNode, instance); err != nil {
  log.Printf("[WorkflowEngine] Warning: OnEnter failed for start node: %v", err)
 }

 result, err := executor.Execute(ctx, startNode, instance, variables)
 if err != nil {
  return nil, fmt.Errorf("failed to execute start node: %w", err)
 }

 if result.Action == "next" && len(result.NextNodes) > 0 {
  if err := w.transitionToNext(ctx, instance, startNode, result); err != nil {
   return nil, fmt.Errorf("failed to transition from start node: %w", err)
  }
 }

 log.Printf("[WorkflowEngine] Started instance: %s for entity: %s/%s", instance.ID, entityType, entityID)

 return instance, nil
}

func (w *WorkflowEngine) CompleteTask(ctx context.Context, taskID string, actor uint, action string, comment string, data map[string]interface{}) (*WorkflowInstance, error) {
 var task WorkflowTask
 if err := w.db.First(&task, "id = ?", taskID).Error; err != nil {
  return nil, fmt.Errorf("task %s not found: %w", taskID, err)
 }

 if task.Status != "pending" {
  return nil, fmt.Errorf("task %s is not pending (status: %s)", taskID, task.Status)
 }

 instance, exists := w.GetInstance(task.InstanceID)
 if !exists {
  return nil, fmt.Errorf("instance %s not found", task.InstanceID)
 }

 def, _ := w.GetDefinition(instance.DefinitionID)
 currentNode := findNodeByID(def.Nodes, task.NodeID)
 if currentNode == nil {
  return nil, fmt.Errorf("node %s not found in definition", task.NodeID)
 }

 executor, ok := w.executors[currentNode.Type]
 if !ok {
  return nil, fmt.Errorf("no executor for node type: %s", currentNode.Type)
 }

 input := map[string]interface{}{
  "action":  action,
  "comment": comment,
  "actor":   actor,
  "data":    data,
 }

 result, err := executor.Execute(ctx, currentNode, instance, input)
 if err != nil {
  return nil, fmt.Errorf("failed to execute task: %w", err)
 }

 task.Status = action
 task.Comment = comment
 now := time.Now()
 task.CompletedAt = &now
 if data != nil {
  task.Data = data
 }
 w.db.Save(&task)

 entry := WorkflowHistoryEntry{
  NodeID:    currentNode.ID,
  NodeName:  currentNode.Name,
  Action:    action,
  Actor:     actor,
  Comment:   comment,
  Data:      data,
  Timestamp: now,
 }
 instance.History = append(instance.History, entry)

 switch action {
 case "approve":
  if err := w.transitionToNext(ctx, instance, currentNode, result); err != nil {
   return nil, err
  }
 case "reject":
  instance.Status = "rejected"
  now := time.Now()
  instance.CompletedAt = &now
 case "return":
  if err := w.returnToPrevious(ctx, instance, currentNode, data); err != nil {
   return nil, err
  }
 default:
  if len(result.NextNodes) > 0 {
   if err := w.transitionToNext(ctx, instance, currentNode, result); err != nil {
    return nil, err
   }
  }
 }

 instance.UpdatedAt = time.Now()
 w.db.Save(instance)

 log.Printf("[WorkflowEngine] Completed task: %s with action: %s", taskID, action)

 return instance, nil
}

func (w *WorkflowEngine) transitionToNext(ctx context.Context, instance *WorkflowInstance, currentNode *WorkflowNodeDef, result *ExecutionResult) error {
 def, _ := w.GetDefinition(instance.DefinitionID)

 newCurrentNodes := make([]string, 0)

 for _, nextNodeID := range result.NextNodes {
  nextNode := findNodeByID(def.Nodes, nextNodeID)
  if nextNode == nil {
   continue
  }

  executor, ok := w.executors[nextNode.Type]
  if !ok {
   continue
  }

  if canExecute, err := executor.CanExecute(ctx, nextNode, instance); err == nil && canExecute {
   newCurrentNodes = append(newCurrentNodes, nextNodeID)

   if err := executor.OnEnter(ctx, nextNode, instance); err != nil {
    log.Printf("[WorkflowEngine] OnEnter failed for node %s: %v", nextNodeID, err)
   }

   if nextNode.Type == "approval" || nextNode.Type == "action" {
    w.createTasksForNode(ctx, instance, nextNode)
   } else if nextNode.Type == "end" {
    instance.Status = "completed"
    now := time.Now()
    instance.CompletedAt = &now
   }
  }
 }

 if len(newCurrentNodes) == 0 {
  instance.Status = "completed"
  now := time.Now()
  instance.CompletedAt = &now
 } else {
  instance.CurrentNodes = newCurrentNodes
 }

 executor, _ := w.executors[currentNode.Type]
 if executor != nil {
  if err := executor.OnLeave(ctx, currentNode, instance); err != nil {
   log.Printf("[WorkflowEngine] OnLeave failed for node %s: %v", currentNode.ID, err)
  }
 }

 return nil
}

func (w *WorkflowEngine) returnToPrevious(ctx context.Context, instance *WorkflowInstance, currentNode *WorkflowNodeDef, data map[string]interface{}) error {
 def, _ := w.GetDefinition(instance.DefinitionID)

 var previousNodeID string
 for i := len(instance.History) - 1; i >= 0; i-- {
  if instance.History[i].Action == "enter" && instance.History[i].NodeID != currentNode.ID {
   previousNodeID = instance.History[i].NodeID
   break
  }
 }

 if previousNodeID == "" {
  return fmt.Errorf("no previous node found to return to")
 }

 previousNode := findNodeByID(def.Nodes, previousNodeID)
 if previousNode == nil {
  return fmt.Errorf("previous node %s not found", previousNodeID)
 }

 instance.CurrentNodes = []string{previousNodeID}
 w.createTasksForNode(ctx, instance, previousNode)

 return nil
}

func (w *WorkflowEngine) createTasksForNode(ctx context.Context, instance *WorkflowInstance, node *WorkflowNodeDef) error {
 if node.Assignee == nil {
  return nil
 }

 assignees, err := w.resolveAssignees(ctx, node.Assignee, instance)
 if err != nil {
  return fmt.Errorf("failed to resolve assignees: %w", err)
 }

 for _, assignee := range assignees {
  task := &WorkflowTask{
   ID:         generateUUID(),
   InstanceID: instance.ID,
   NodeID:     node.ID,
   NodeName:   node.Name,
   Assignee:   assignee,
   Status:     "pending",
   CreatedAt:  time.Now(),
  }

  if node.TimeLimit != nil {
   dueTime := time.Now().Add(time.Duration(node.TimeLimit.Duration) * time.Hour)
   task.DueDate = &dueTime
  }

  if err := w.db.Create(task).Error; err != nil {
   log.Printf("[WorkflowEngine] Failed to create task: %v", err)
  }
 }

 return nil
}

func (w *WorkflowEngine) resolveAssignees(ctx context.Context, config *AssigneeConfig, instance *WorkflowInstance) ([]uint, error) {
 switch config.Type {
 case "user":
  var userIDs []uint
  for _, v := range config.Value {
   var id uint
   if _, err := fmt.Sscanf(v, "%d", &id); err == nil {
    userIDs = append(userIDs, id)
   }
  }
  return userIDs, nil

 case "role":
  return w.getUsersByRole(config.Value...)

 case "initiator":
  return []uint{instance.Initiator}, nil

 case "field":
  if val, ok := instance.Variables[config.Field]; ok {
   if id, ok := val.(float64); ok {
    return []uint{uint(id)}, nil
   }
  }
  return nil, nil

 default:
  return nil, fmt.Errorf("unsupported assignee type: %s", config.Type)
 }
}

func (w *WorkflowEngine) getUsersByRole(roles ...string) ([]uint, error) {
 return []uint{}, nil
}

func (w *WorkflowEngine) GetInstance(id string) (*WorkflowInstance, bool) {
 w.mu.RLock()
 defer w.mu.RUnlock()

 instance, ok := w.instances[id]
 if !ok {
  var dbInstance WorkflowInstance
  if err := w.db.First(&dbInstance, "id = ?", id).Error; err != nil {
   return nil, false
  }
  w.instances[id] = &dbInstance
  return &dbInstance, true
 }
 return instance, true
}

func (w *WorkflowEngine) GetTasksForUser(userID uint, status string) ([]WorkflowTask, error) {
 var tasks []WorkflowTask
 query := w.db.Where("assignee = ?", userID)

 if status != "" {
  query = query.Where("status = ?", status)
 }

 if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
  return nil, fmt.Errorf("failed to get tasks: %w", err)
 }

 return tasks, nil
}

func (w *WorkflowEngine) validateDefinition(def *WorkflowDefinition) error {
 if def.Name == "" {
  return fmt.Errorf("name is required")
 }

 hasStart := false
 hasEnd := false
 nodeIDs := make(map[string]bool)

 for _, node := range def.Nodes {
  if node.ID == "" {
   return fmt.Errorf("node ID is required")
  }
  if nodeIDs[node.ID] {
   return fmt.Errorf("duplicate node ID: %s", node.ID)
  }
  nodeIDs[node.ID] = true

  if node.Type == "start" {
   hasStart = true
  }
  if node.Type == "end" {
   hasEnd = true
  }
 }

 if !hasStart {
  return fmt.Errorf("workflow must have a start node")
 }
 if !hasEnd {
  return fmt.Errorf("workflow must have an end node")
 }

 for _, edge := range def.Edges {
  if !nodeIDs[edge.Source] {
   return fmt.Errorf("edge source node %s not found", edge.Source)
  }
  if !nodeIDs[edge.Target] {
   return fmt.Errorf("edge target node %s not found", edge.Target)
  }
 }

 return nil
}

func findStartNode(nodes []WorkflowNodeDef) *WorkflowNodeDef {
 for i := range nodes {
  if nodes[i].Type == "start" {
   return &nodes[i]
  }
 }
 return nil
}

func findNodeByID(nodes []WorkflowNodeDef, id string) *WorkflowNodeDef {
 for i := range nodes {
  if nodes[i].ID == id {
   return &nodes[i]
  }
 }
 return nil
}

type StartNodeExecutor struct{}

func (e *StartNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) {
 return true, nil
}

func (e *StartNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 edges := findOutgoingEdges(getDefinitionFromInstance(instance), node.ID)
 nextNodes := make([]string, len(edges))
 for i, edge := range edges {
  nextNodes[i] = edge.Target
 }

 return &ExecutionResult{
  Success:   true,
  Action:    "next",
  NextNodes: nextNodes,
 }, nil
}

func (e *StartNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error {
 return nil
}

func (e *StartNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error {
 return nil
}

func (e *StartNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error {
 return nil
}

type EndNodeExecutor struct{}

func (e *EndNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) {
 return true, nil
}

func (e *EndNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 return &ExecutionResult{
  Success: true,
  Action:  "complete",
 }, nil
}

func (e *EndNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *EndNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *EndNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error { return nil }

type ApprovalNodeExecutor struct {
 db *gorm.DB
}

func (e *ApprovalNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) {
 return true, nil
}

func (e *ApprovalNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 action, _ := input["action"].(string)

 switch strings.ToLower(action) {
 case "approve":
  edges := findOutgoingEdgesByLabel(getDefinitionFromInstance(instance), node.ID, "approve")
  nextNodes := make([]string, len(edges))
  for i, edge := range edges {
   nextNodes[i] = edge.Target
  }
  return &ExecutionResult{Success: true, Action: "approve", NextNodes: nextNodes}, nil

 case "reject":
  return &ExecutionResult{Success: true, Action: "reject"}, nil

 case "return":
  return &ExecutionResult{Success: true, Action: "return"}, nil

 case "delegate":
  return &ExecutionResult{Success: true, Action: "delegate"}, nil

 default:
  return &ExecutionResult{Success: false, Error: fmt.Errorf("unknown action: %s", action)}, nil
 }
}

func (e *ApprovalNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ApprovalNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ApprovalNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error {
 action, _ := input["action"].(string)
 validActions := map[string]bool{"approve": true, "reject": true, "return": true, "delegate": true}
 if !validActions[strings.ToLower(action)] {
  return fmt.Errorf("invalid action: %s", action)
 }
 return nil
}

type ConditionNodeExecutor struct{}

func (e *ConditionNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) { return true, nil }
func (e *ConditionNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 for _, cond := range node.Conditions {
  value, exists := instance.Variables[cond.Field]
  if !exists {
   continue
  }

  if evaluateCondition(cond.Operator, value, cond.Value) {
   edges := findOutgoingEdgesByCondition(getDefinitionFromInstance(instance), node.ID, cond.Field)
   nextNodes := make([]string, len(edges))
   for i, edge := range edges {
    nextNodes[i] = edge.Target
   }
   return &ExecutionResult{Success: true, Action: "next", NextNodes: nextNodes}, nil
  }
 }

 edges := findOutgoingEdgesByLabel(getDefinitionFromInstance(instance), node.ID, "else")
 nextNodes := make([]string, len(edges))
 for i, edge := range edges {
  nextNodes[i] = edge.Target
 }
 return &ExecutionResult{Success: true, Action: "next", NextNodes: nextNodes}, nil
}
func (e *ConditionNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ConditionNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ConditionNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error { return nil }

type ActionNodeExecutor struct{}

func (e *ActionNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) { return true, nil }
func (e *ActionNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 actionType, _ := node.Config["type"].(string)

 switch actionType {
 case "update_field":
  if field, ok := node.Config["field"].(string); ok {
   if value, ok := node.Config["value"]; ok {
    instance.Variables[field] = value
   }
  }
 case "set_variable":
  if variable, ok := node.Config["variable"].(string); ok {
   if value, ok := node.Config["value"]; ok {
    instance.Variables[variable] = value
   }
  }
 case "call_api":
  log.Printf("[ActionNode] Would call API: %v", node.Config)
 }

 edges := findOutgoingEdges(getDefinitionFromInstance(instance), node.ID)
 nextNodes := make([]string, len(edges))
 for i, edge := range edges {
  nextNodes[i] = edge.Target
 }
 return &ExecutionResult{Success: true, Action: "next", NextNodes: nextNodes}, nil
}
func (e *ActionNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ActionNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ActionNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error { return nil }

type ParallelNodeExecutor struct{}

func (e *ParallelNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) { return true, nil }
func (e *ParallelNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 edges := findOutgoingEdges(getDefinitionFromInstance(instance), node.ID)
 nextNodes := make([]string, len(edges))
 for i, edge := range edges {
  nextNodes[i] = edge.Target
 }
 return &ExecutionResult{Success: true, Action: "next", NextNodes: nextNodes}, nil
}
func (e *ParallelNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ParallelNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *ParallelNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error { return nil }

type TimerNodeExecutor struct{}

func (e *TimerNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) { return true, nil }
func (e *TimerNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 duration, _ := node.Config["duration"].(int)
 unit, _ := node.Config["unit"].(string)

 var waitTime time.Duration
 switch unit {
 case "second":
  waitTime = time.Duration(duration) * time.Second
 case "minute":
  waitTime = time.Duration(duration) * time.Minute
 case "hour":
  waitTime = time.Duration(duration) * time.Hour
 case "day":
  waitTime = time.Duration(duration) * 24 * time.Hour
 default:
  waitTime = time.Duration(duration) * time.Hour
 }

 log.Printf("[TimerNode] Waiting for %v", waitTime)
 time.Sleep(waitTime)

 edges := findOutgoingEdges(getDefinitionFromInstance(instance), node.ID)
 nextNodes := make([]string, len(edges))
 for i, edge := range edges {
  nextNodes[i] = edge.Target
 }
 return &ExecutionResult{Success: true, Action: "next", NextNodes: nextNodes}, nil
}
func (e *TimerNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *TimerNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *TimerNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error { return nil }

type NotificationNodeExecutor struct{}

func (e *NotificationNodeExecutor) CanExecute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) (bool, error) { return true, nil }
func (e *NotificationNodeExecutor) Execute(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance, input map[string]interface{}) (*ExecutionResult, error) {
 notificationType, _ := node.Config["type"].(string)
 message, _ := node.Config["message"].(string)

 log.Printf("[NotificationNode] Sending %s: %s", notificationType, message)

 edges := findOutgoingEdges(getDefinitionFromInstance(instance), node.ID)
 nextNodes := make([]string, len(edges))
 for i, edge := range edges {
  nextNodes[i] = edge.Target
 }
 return &ExecutionResult{Success: true, Action: "next", NextNodes: nextNodes}, nil
}
func (e *NotificationNodeExecutor) OnEnter(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *NotificationNodeExecutor) OnLeave(ctx context.Context, node *WorkflowNodeDef, instance *WorkflowInstance) error { return nil }
func (e *NotificationNodeExecutor) Validate(ctx context.Context, node *WorkflowNodeDef, input map[string]interface{}) error { return nil }

func getDefinitionFromInstance(instance *WorkflowInstance) *WorkflowDefinition {
 return nil
}

func findOutgoingEdges(def *WorkflowDefinition, nodeID string) []WorkflowEdgeDef {
 if def == nil {
  return []WorkflowEdgeDef{}
 }
 var edges []WorkflowEdgeDef
 for _, edge := range def.Edges {
  if edge.Source == nodeID {
   edges = append(edges, edge)
  }
 }
 return edges
}

func findOutgoingEdgesByLabel(def *WorkflowDefinition, nodeID, label string) []WorkflowEdgeDef {
 if def == nil {
  return []WorkflowEdgeDef{}
 }
 var edges []WorkflowEdgeDef
 for _, edge := range def.Edges {
  if edge.Source == nodeID && edge.Label == label {
   edges = append(edges, edge)
  }
 }
 return edges
}

func findOutgoingEdgesByCondition(def *WorkflowDefinition, nodeID, condition string) []WorkflowEdgeDef {
 if def == nil {
  return []WorkflowEdgeDef{}
 }
 var edges []WorkflowEdgeDef
 for _, edge := range def.Edges {
  if edge.Source == nodeID && edge.Condition == condition {
   edges = append(edges, edge)
  }
 }
 return edges
}

func evaluateCondition(operator string, actual, expected interface{}) bool {
 switch operator {
 case "eq":
  return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
 case "neq":
  return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected)
 case "is_empty":
  return actual == nil || actual == ""
 case "is_not_empty":
  return actual != nil && actual != ""
 default:
  return false
 }
}

func generateUUID() string {
 return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (w *WorkflowEngine) ExportDefinitionJSON(definitionID string) ([]byte, error) {
 def, exists := w.GetDefinition(definitionID)
 if !exists {
  return nil, fmt.Errorf("definition %s not found", definitionID)
 }

 return json.MarshalIndent(def, "", "  ")
}

func (w *WorkflowEngine) ImportDefinitionJSON(data []byte) (*WorkflowDefinition, error) {
 var def WorkflowDefinition
 if err := json.Unmarshal(data, &def); err != nil {
  return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
 }

 if err := w.RegisterDefinition(context.Background(), &def); err != nil {
  return nil, fmt.Errorf("failed to register imported definition: %w", err)
 }

 return &def, nil
}

func (w *WorkflowEngine) GetDefinitionsByEntity(entity string) ([]*WorkflowDefinition, error) {
 var defs []*WorkflowDefinition
 if err := w.db.Where("entity = ? AND status = ?", entity, "active").Find(&defs).Error; err != nil {
  return nil, fmt.Errorf("failed to get definitions: %w", err)
 }
 return defs, nil
}
