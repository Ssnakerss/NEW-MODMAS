package ddl

import (
	"fmt"
	"strings"
)

type ColumnDef struct {
	Name         string
	FieldType    string
	IsRequired   bool
	IsUnique     bool
	DefaultValue *string
	Options      map[string]interface{}
}

func BuildCreateTable(schema, tableName string, columns []ColumnDef) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", quote(schema), quote(tableName)))
	sb.WriteString("    _id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),\n")
	sb.WriteString("    _created_by UUID NOT NULL,\n")
	sb.WriteString("    _created_at TIMESTAMPTZ NOT NULL DEFAULT now(),\n")
	sb.WriteString("    _updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),\n")
	sb.WriteString("    _position   INT\n")

	for _, col := range columns {
		sb.WriteString(",\n    ")
		sb.WriteString(buildColumnDef(col))
	}

	sb.WriteString("\n)")

	return sb.String()
}

func BuildAddColumn(schema, tableName string, col ColumnDef) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s ADD COLUMN %s",
		quote(schema), quote(tableName), buildColumnDef(col),
	)
}

func BuildDropColumn(schema, tableName, columnName string) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s DROP COLUMN IF EXISTS %s",
		quote(schema), quote(tableName), quote(columnName),
	)
}

func BuildAlterColumnType(schema, tableName string, col ColumnDef) string {
	pgType := mapFieldType(col.FieldType, col.Options)
	using := fmt.Sprintf("%s::%s", quote(col.Name), pgType)
	return fmt.Sprintf(
		"ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s USING %s",
		quote(schema), quote(tableName), quote(col.Name), pgType, using,
	)
}

func BuildRenameColumn(schema, tableName, oldName, newName string) string {
	return fmt.Sprintf(
		"ALTER TABLE %s.%s RENAME COLUMN %s TO %s",
		quote(schema), quote(tableName), quote(oldName), quote(newName),
	)
}

func BuildCreateSchema(schema string) string {
	return fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", quote(schema))
}

func buildColumnDef(col ColumnDef) string {
	var sb strings.Builder

	pgType := mapFieldType(col.FieldType, col.Options)
	sb.WriteString(fmt.Sprintf("%s %s", quote(col.Name), pgType))

	if col.IsRequired {
		sb.WriteString(" NOT NULL")
	}
	if col.IsUnique {
		sb.WriteString(" UNIQUE")
	}
	if col.DefaultValue != nil {
		sb.WriteString(fmt.Sprintf(" DEFAULT %s", *col.DefaultValue))
	}

	if col.FieldType == "select" {
		if choices, ok := col.Options["choices"].([]interface{}); ok && len(choices) > 0 {
			vals := make([]string, 0, len(choices))
			for _, c := range choices {
				if m, ok := c.(map[string]interface{}); ok {
					if v, ok := m["value"].(string); ok {
						vals = append(vals, fmt.Sprintf("'%s'", escapeSQLString(v)))
					}
				}
			}
			if len(vals) > 0 {
				sb.WriteString(fmt.Sprintf(" CHECK (%s IN (%s))", quote(col.Name), strings.Join(vals, ",")))
			}
		}
	}

	if col.FieldType == "email" {
		sb.WriteString(fmt.Sprintf(" CHECK (%s ~* '^[^@\\s]+@[^@\\s]+\\.[^@\\s]')", quote(col.Name)))
	}

	if col.FieldType == "url" {
		sb.WriteString(fmt.Sprintf(" CHECK (%s ~* '^https?://')", quote(col.Name)))
	}

	return sb.String()
}

func mapFieldType(fieldType string, options map[string]interface{}) string {
	switch fieldType {
	case "text", "email", "url", "phone", "file", "select":
		return "TEXT"
	case "integer":
		return "BIGINT"
	case "decimal":
		return "NUMERIC(18,6)"
	case "boolean":
		return "BOOLEAN"
	case "date":
		return "DATE"
	case "datetime":
		return "TIMESTAMPTZ"
	case "multi_select":
		return "TEXT[]"
	case "relation":
		return "UUID"
	default:
		return "TEXT"
	}
}

func quote(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
