# Фронтенд — полная реализация

## Стек

* **React 18** + **TypeScript**
* **Zustand** — стейт-менеджмент
* **TanStack Query** — серверный стейт
* **TanStack Virtual** — виртуализация строк
* **Tailwind CSS** — стили

---


## Структура базы данных

### Схема auth

```
users
├── id            UUID PK DEFAULT gen_random_uuid()
├── email         TEXT UNIQUE NOT NULL
├── password_hash TEXT NOT NULL
├── name          TEXT
├── avatar_url    TEXT
├── created_at    TIMESTAMPTZ DEFAULT now()
└── updated_at    TIMESTAMPTZ DEFAULT now()

workspaces
├── id         UUID PK
├── name       TEXT NOT NULL
├── owner_id   UUID FK → users.id
├── db_schema  TEXT UNIQUE NOT NULL  -- 'data_<uuid>' — имя схемы в PG
└── created_at TIMESTAMPTZ

workspace_members
├── workspace_id UUID FK → workspaces.id
├── user_id      UUID FK → users.id
├── role         ENUM('owner', 'admin', 'member', 'viewer')
└── PRIMARY KEY (workspace_id, user_id)
```


### Схема meta — метаданные структуры


-- Описание пользовательской таблицы
spreadsheets
├── id           UUID PK
├── workspace_id UUID FK → workspaces.id
├── name         TEXT NOT NULL          -- отображаемое имя
├── table_name   TEXT NOT NULL          -- имя таблицы в PG: 'tbl_`<uuid>`'
├── description  TEXT
├── created_by   UUID FK → users.id
├── created_at   TIMESTAMPTZ
└── updated_at   TIMESTAMPTZ
-- UNIQUE (workspace_id, table_name)

-- Описание полей (колонок) пользовательской таблицы
fields
├── id             UUID PK
├── spreadsheet_id UUID FK → spreadsheets.id
├── name           TEXT NOT NULL          -- отображаемое имя
├── column_name    TEXT NOT NULL          -- имя колонки в PG: 'col_`<uuid>`'
├── field_type     field_type_enum NOT NULL
├── position       INT NOT NULL
├── is_required    BOOL DEFAULT false
├── is_unique      BOOL DEFAULT false
├── default_value  TEXT                   -- SQL-совместимое значение
├── options        JSONB                  -- доп. параметры (см. ниже)
└── created_at     TIMESTAMPTZ
-- UNIQUE (spreadsheet_id, column_name)

-- Перечисление типов полей
CREATE TYPE field_type_enum AS ENUM (
    'text',          -- TEXT
    'integer',       -- BIGINT
    'decimal',       -- NUMERIC(18,6)
    'boolean',       -- BOOLEAN
    'date',          -- DATE
    'datetime',      -- TIMESTAMPTZ
    'select',        -- TEXT + CHECK constraint
    'multi_select',  -- TEXT[]
    'email',         -- TEXT + CHECK (email format)
    'url',           -- TEXT + CHECK (url format)
    'phone',         -- TEXT
    'file',          -- TEXT (URL to storage)
    'relation'       -- UUID FK → другая таблица
);



### Схема meta — права доступа


-- Доступ к таблице целиком
spreadsheet_access
├── id               UUID PK
├── spreadsheet_id   UUID FK → spreadsheets.id
├── principal_id     UUID        -- users.id или ссылка на роль
├── principal_type   ENUM('user', 'workspace_role')
├── can_view         BOOL DEFAULT false
├── can_insert       BOOL DEFAULT false
├── can_edit         BOOL DEFAULT false
├── can_delete       BOOL DEFAULT false
└── can_manage       BOOL DEFAULT false  -- изменение структуры и прав

-- Доступ к конкретным полям (колонкам)
field_access
├── id             UUID PK
├── field_id       UUID FK → fields.id
├── principal_id   UUID
├── principal_type ENUM('user', 'workspace_role')
├── can_view       BOOL DEFAULT false
└── can_edit       BOOL DEFAULT false

-- Доступ к строкам по условию (Row-Level Security правила)
row_access_rules
├── id             UUID PK
├── spreadsheet_id UUID FK → spreadsheets.id
├── principal_id   UUID
├── principal_type ENUM('user', 'workspace_role')
├── condition      JSONB    -- {"column_name": "owner_col_uuid", "op": "eq_current_user"}
├── can_view       BOOL
└── can_edit       BOOL


### Схема data_ — пользовательские данные


-- Пример: пользователь создал таблицу "Заявки" с полями:
-- Название (text), Сумма (decimal), Статус (select), Дата (date), Автор (relation)

CREATE TABLE data_w1."tbl_a3f9..." (
    -- Системные колонки (добавляются всегда)
    _id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    _created_by  UUID NOT NULL REFERENCES auth.users(id),
    _created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    _updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    _position    INT,   -- порядок строк

    -- Пользовательские колонки
    "col_b1c2..."  TEXT,                          -- Название     (text)
    "col_d3e4..."  NUMERIC(18,6),                 -- Сумма        (decimal)
    "col_f5g6..."  TEXT CHECK (... IN (...)),      -- Статус       (select)
    "col_h7i8..."  DATE,                          -- Дата         (date)
    "col_j9k0..."  UUID REFERENCES data_w1."tbl_..." (_id)  -- Автор (relation)
);

-- Индекс для сортировки
CREATE INDEX ON data_w1."tbl_a3f9..." (_position);
CREATE INDEX ON data_w1."tbl_a3f9..." (_created_at);
