export type FieldType =
  | 'text'
  | 'integer'
  | 'decimal'
  | 'boolean'
  | 'date'
  | 'datetime'
  | 'select'
  | 'multi_select'
  | 'email'
  | 'url'
  | 'phone'
  | 'file'
  | 'relation';

export interface SelectOption {
  value: string;
  label: string;
  color?: string;
}

export interface FieldOptions {
  choices?: SelectOption[];          // select, multi_select
  relation_spreadsheet_id?: string;  // relation
  relation_display_field_id?: string;
}

export interface Field {
  id: string;
  spreadsheet_id: string;
  name: string;
  column_name: string;
  field_type: FieldType;
  position: number;
  is_required: boolean;
  is_unique: boolean;
  default_value?: string;
  options?: FieldOptions;
}

export interface Spreadsheet {
  id: string;
  workspace_id: string;
  name: string;
  description?: string;
  fields: Field[];
  created_at: string;
  updated_at: string;
}

export interface Row {
  _id: string;
  _created_by: string;
  _created_at: string;
  _updated_at: string;
  [column_name: string]: unknown;
}

export interface Workspace {
  id: string;
  name: string;
  role: 'owner' | 'admin' | 'member' | 'viewer';
}

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
}

// Permissions
export type Permission = 'can_view' | 'can_insert' | 'can_edit' | 'can_delete' | 'can_manage';

export interface SpreadsheetAccess {
  id: string;
  principal_id: string;
  principal_type: 'user' | 'workspace_role';
  principal_name: string;
  can_view: boolean;
  can_insert: boolean;
  can_edit: boolean;
  can_delete: boolean;
  can_manage: boolean;
  role?: PermissionRole;
}

export interface FieldAccess {
  field_id: string;
  principal_id: string;
  can_view: boolean;
  can_edit: boolean;
  access?: 'view' | 'edit' | 'none';
}



export interface PaginatedRows {
  data: Row[];
  total: number;
  limit: number;
  offset: number;
}

export interface FilterCondition {
  field_id: string;
  op: 'eq' | 'neq' | 'contains' | 'gt' | 'lt' | 'is_empty' | 'is_not_empty';
  value?: string;
}

export interface SortCondition {
  field_id: string;
  direction: 'asc' | 'desc';
}export type PermissionRole = 'owner' | 'editor' | 'viewer';

export interface FieldPermission {
  fieldId: string;
  userId: string;
  access: 'view' | 'edit' | 'none';
}

export interface SpreadsheetPermission {
  userId: string;
  email: string;
  name: string;
  role: PermissionRole;
  fieldPermissions?: FieldPermission[];
}





// export interface RowAccessRule {
//   id: string;
//   spreadsheetId: string;
//   fieldId: string;
//   principal_id: string;
//   operator: 'eq' | 'neq' | 'gt' | 'lt' | 'contains' | 'is_empty';
//   value: string | number | boolean | null;
//   access: 'view' | 'none';
// }

export interface RowAccessRule {
  id: string;
  spreadsheet_id: string;
  principal_id: string;
  principal_type: 'user' | 'workspace_role';
  condition: {
    column_name: string;
    op: 'eq' | 'neq' | 'eq_current_user' | 'contains';
    value?: string;
  };
  can_view: boolean;
  can_edit: boolean;
}

// export interface RowAccessRule {
//   id: string;
//   spreadsheetId: string;
//   userId: string;
//   // Условие: поле + оператор + значение
//   fieldId: string;
//   operator: 'eq' | 'neq' | 'gt' | 'lt' | 'contains' | 'is_empty';
//   value: string | number | boolean | null;
//   access: 'view' | 'none';
// }