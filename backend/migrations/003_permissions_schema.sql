CREATE TABLE meta.spreadsheet_access (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spreadsheet_id UUID NOT NULL REFERENCES meta.spreadsheets(id) ON DELETE CASCADE,
    principal_id   UUID NOT NULL,
    principal_type TEXT NOT NULL CHECK (principal_type IN ('user','workspace_role')),
    can_view       BOOL NOT NULL DEFAULT false,
    can_insert     BOOL NOT NULL DEFAULT false,
    can_edit       BOOL NOT NULL DEFAULT false,
    can_delete     BOOL NOT NULL DEFAULT false,
    can_manage     BOOL NOT NULL DEFAULT false,
    UNIQUE (spreadsheet_id, principal_id, principal_type)
);

CREATE TABLE meta.field_access (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id       UUID NOT NULL REFERENCES meta.fields(id) ON DELETE CASCADE,
    principal_id   UUID NOT NULL,
    principal_type TEXT NOT NULL CHECK (principal_type IN ('user','workspace_role')),
    can_view       BOOL NOT NULL DEFAULT true,
    can_edit       BOOL NOT NULL DEFAULT true,
    UNIQUE (field_id, principal_id, principal_type)
);

CREATE TABLE meta.row_access_rules (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spreadsheet_id UUID NOT NULL REFERENCES meta.spreadsheets(id) ON DELETE CASCADE,
    principal_id   UUID NOT NULL,
    principal_type TEXT NOT NULL CHECK (principal_type IN ('user','workspace_role')),
    condition      JSONB NOT NULL,
    can_view       BOOL NOT NULL DEFAULT false,
    can_edit       BOOL NOT NULL DEFAULT false
);

CREATE INDEX idx_spreadsheet_access_principal ON meta.spreadsheet_access(principal_id);
CREATE INDEX idx_row_rules_spreadsheet ON meta.row_access_rules(spreadsheet_id);