CREATE TABLE meta.spreadsheets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES auth.workspaces(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    table_name   TEXT NOT NULL,
    description  TEXT,
    created_by   UUID NOT NULL REFERENCES auth.users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, table_name)
);

CREATE TYPE meta.field_type AS ENUM (
    'text','integer','decimal','boolean',
    'date','datetime','select','multi_select',
    'email','url','phone','file','relation'
);

CREATE TABLE meta.fields (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spreadsheet_id UUID NOT NULL REFERENCES meta.spreadsheets(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    column_name    TEXT NOT NULL,
    field_type     meta.field_type NOT NULL,
    position       INT NOT NULL DEFAULT 0,
    is_required    BOOL NOT NULL DEFAULT false,
    is_unique      BOOL NOT NULL DEFAULT false,
    default_value  TEXT,
    options        JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (spreadsheet_id, column_name)
);

CREATE INDEX idx_fields_spreadsheet ON meta.fields(spreadsheet_id);