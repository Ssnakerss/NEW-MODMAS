package types

import "time"

// ─── Field ────────────────────────────────────────────────────────────────────

type Field struct {
	ID            string                 `json:"id"`
	SpreadsheetID string                 `json:"spreadsheet_id"`
	Name          string                 `json:"name"`
	ColumnName    string                 `json:"column_name"`
	FieldType     string                 `json:"field_type"`
	Position      int                    `json:"position"`
	IsRequired    bool                   `json:"is_required"`
	IsUnique      bool                   `json:"is_unique"`
	DefaultValue  *string                `json:"default_value,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}

type CreateFieldInput struct {
	Name         string                 `json:"name"`
	FieldType    string                 `json:"field_type"`
	IsRequired   bool                   `json:"is_required"`
	IsUnique     bool                   `json:"is_unique"`
	DefaultValue *string                `json:"default_value"`
	Options      map[string]interface{} `json:"options"`
}

type UpdateFieldInput struct {
	Name         *string                `json:"name"`
	FieldType    *string                `json:"field_type"`
	IsRequired   *bool                  `json:"is_required"`
	IsUnique     *bool                  `json:"is_unique"`
	DefaultValue *string                `json:"default_value"`
	Options      map[string]interface{} `json:"options"`
}

// ─── Spreadsheet ──────────────────────────────────────────────────────────────

type Spreadsheet struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	TableName   string    `json:"table_name"`
	DBSchema    string    `json:"db_schema,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SpreadsheetWithFields struct {
	*Spreadsheet
	Fields []*Field `json:"fields"`
}

type CreateSpreadsheetInput struct {
	WorkspaceID string             `json:"workspace_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Fields      []CreateFieldInput `json:"fields"`
}

// ─── Workspace ────────────────────────────────────────────────────────────────

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	DBSchema  string    `json:"db_schema"`
	CreatedAt time.Time `json:"created_at"`
	Role      string    `json:"role,omitempty"`
}

// ─── Spreadsheet Access ──────────────────────────────────────────────────────

type SpreadsheetAccess struct {
	ID            string `json:"id,omitempty"`
	SpreadsheetID string `json:"spreadsheet_id"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	PrincipalName string `json:"principal_name,omitempty"`
	CanView       bool   `json:"can_view"`
	CanInsert     bool   `json:"can_insert"`
	CanEdit       bool   `json:"can_edit"`
	CanDelete     bool   `json:"can_delete"`
	CanManage     bool   `json:"can_manage"`
}

type FieldAccess struct {
	ID            string `json:"id,omitempty"`
	FieldID       string `json:"field_id"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	CanView       bool   `json:"can_view"`
	CanEdit       bool   `json:"can_edit"`
}
