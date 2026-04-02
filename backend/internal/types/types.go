package types

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
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	TableName   string `json:"table_name"`
	DBSchema    string `json:"db_schema,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
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

// Добавить в конец файла:

// ─── Workspace ────────────────────────────────────────────────────────────────

type Workspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerID   string `json:"owner_id"`
	DBSchema  string `json:"db_schema"`
	CreatedAt string `json:"created_at"`
	Role      string `json:"role,omitempty"`
}
