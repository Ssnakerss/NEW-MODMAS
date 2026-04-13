# опиши структуру прилож...

**User - --**

опиши структуру приложения. функции - пользователи могут создавать электронные таблицы,описывая поля и тип полей. можно натсраивать доступ определенных пользователей с таблицам, строкам и полям для просмотра и редактирования.интерфейс - веб. для хранения информации используется субд postgres

**Assistant - --**

# Структура приложения — «Умные таблицы» (No-code платформа)

---

## Общая архитектура

```
┌─────────────────────────────────────────────────────────┐
│                     Web Client (SPA)                    │
│              React / Vue + TypeScript                   │
└─────────────────────────┬───────────────────────────────┘
                          │ HTTP / WebSocket
┌─────────────────────────▼───────────────────────────────┐
│                    API Gateway / BFF                    │
│                   (REST + WebSocket)                    │
└──────┬──────────────────┬──────────────────┬────────────┘
       │                  │                  │
┌──────▼──────┐  ┌────────▼───────┐  ┌──────▼──────┐
│  Auth       │  │  Core API      │  │  Permission │
│  Service    │  │  Service       │  │  Service    │
└──────┬──────┘  └────────┬───────┘  └──────┬──────┘
       │                  │                  │
       └──────────────────▼──────────────────┘
                          │
              ┌───────────▼───────────┐
              │      PostgreSQL        │
              └───────────────────────┘
```

---

## Структура базы данных

### Схема `auth` — пользователи и организации

```sql
-- Пользователи
users
├── id            UUID PK
├── email         TEXT UNIQUE NOT NULL
├── password_hash TEXT NOT NULL
├── name          TEXT
├── avatar_url    TEXT
├── created_at    TIMESTAMPTZ
└── updated_at    TIMESTAMPTZ

-- Организации / рабочие пространства
workspaces
├── id         UUID PK
├── name       TEXT NOT NULL
├── owner_id   UUID FK → users.id
└── created_at TIMESTAMPTZ

-- Участники рабочего пространства
workspace_members
├── workspace_id UUID FK → workspaces.id
├── user_id      UUID FK → users.id
└── role         ENUM('owner', 'admin', 'member', 'viewer')
```

---

### Схема `tables` — структура таблиц

```sql
-- Пользовательские таблицы (метаданные)
spreadsheets
├── id           UUID PK
├── workspace_id UUID FK → workspaces.id
├── name         TEXT NOT NULL
├── description  TEXT
├── created_by   UUID FK → users.id
├── created_at   TIMESTAMPTZ
└── updated_at   TIMESTAMPTZ

-- Определение полей (колонок)
fields
├── id             UUID PK
├── spreadsheet_id UUID FK → spreadsheets.id
├── name           TEXT NOT NULL
├── field_type     ENUM(...)   -- см. ниже
├── position       INT         -- порядок колонок
├── is_required    BOOL
├── is_unique      BOOL
├── default_value  JSONB
└── options        JSONB       -- доп. настройки типа

-- Типы полей (field_type):
-- text, number, boolean, date, datetime,
-- select, multi_select, url, email,
-- phone, file, relation, formula, user
```

---

### Схема `data` — хранение данных

```sql
-- Строки таблицы
rows
├── id             UUID PK
├── spreadsheet_id UUID FK → spreadsheets.id
├── position       INT         -- порядок строк
├── created_by     UUID FK → users.id
├── created_at     TIMESTAMPTZ
└── updated_at     TIMESTAMPTZ

-- Значения ячеек (EAV-подход)
cell_values
├── id         UUID PK
├── row_id     UUID FK → rows.id
├── field_id   UUID FK → fields.id
├── value      JSONB       -- значение любого типа
└── updated_at TIMESTAMPTZ

-- Уникальность: (row_id, field_id)
```

> **Почему JSONB?** Позволяет хранить значения любых типов в одной таблице без schema migration при добавлении полей.

---

### Схема `permissions` — управление доступом

```sql
-- Уровни доступа к таблице в целом
spreadsheet_access
├── id             UUID PK
├── spreadsheet_id UUID FK → spreadsheets.id
├── principal_id   UUID         -- user_id или role_id
├── principal_type ENUM('user', 'workspace_role')
├── can_view       BOOL
├── can_edit       BOOL
├── can_delete     BOOL
└── can_share      BOOL

-- Доступ к конкретным полям (колонкам)
field_access
├── id           UUID PK
├── field_id     UUID FK → fields.id
├── principal_id UUID
├── principal_type ENUM('user', 'workspace_role')
├── can_view     BOOL
└── can_edit     BOOL

-- Доступ к конкретным строкам
row_access
├── id           UUID PK
├── row_id       UUID FK → rows.id
├── principal_id UUID
├── principal_type ENUM('user', 'workspace_role')
├── can_view     BOOL
└── can_edit     BOOL

-- Фильтры строк по условию (динамические права)
row_access_rules
├── id             UUID PK
├── spreadsheet_id UUID FK → spreadsheets.id
├── principal_id   UUID
├── principal_type ENUM('user', 'workspace_role')
├── condition      JSONB   -- {"field_id": "...", "op": "eq", "value": "..."}
└── permission     JSONB   -- {"can_view": true, "can_edit": false}
```

---

## Структура бэкенда

```
backend/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   │   ├── handler.go       # /register, /login, /refresh
│   │   ├── service.go       # JWT, bcrypt
│   │   └── repository.go
│   │
│   ├── workspace/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── spreadsheet/
│   │   ├── handler.go       # CRUD таблиц
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── field/
│   │   ├── handler.go       # CRUD полей
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── types/           # валидация по типу поля
│   │       ├── text.go
│   │       ├── number.go
│   │       ├── select.go
│   │       └── ...
│   │
│   ├── row/
│   │   ├── handler.go       # CRUD строк + ячеек
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── permission/
│   │   ├── handler.go       # настройка прав
│   │   ├── service.go       # проверка прав (middleware)
│   │   ├── repository.go
│   │   └── enforcer.go      # движок проверки доступа
│   │
│   └── middleware/
│       ├── auth.go          # проверка JWT
│       ├── permission.go    # проверка прав на ресурс
│       └── ratelimit.go
│
├── pkg/
│   ├── postgres/            # pool соединений
│   ├── jwt/
│   └── validator/
│
└── migrations/              # SQL миграции
    ├── 001_init_auth.sql
    ├── 002_init_tables.sql
    ├── 003_init_data.sql
    └── 004_init_permissions.sql
```

---

## Структура фронтенда

```
frontend/
├── src/
│   ├── pages/
│   │   ├── Login/
│   │   ├── Workspaces/
│   │   ├── Spreadsheet/        # главная страница таблицы
│   │   └── Settings/
│   │
│   ├── components/
│   │   ├── Grid/
│   │   │   ├── GridView.tsx     # рендер таблицы
│   │   │   ├── Cell.tsx         # ячейка
│   │   │   ├── HeaderCell.tsx   # заголовок колонки
│   │   │   └── RowActions.tsx
│   │   │
│   │   ├── FieldEditor/
│   │   │   ├── FieldTypeSelect.tsx
│   │   │   └── FieldOptionsForm.tsx
│   │   │
│   │   ├── Permissions/
│   │   │   ├── ShareModal.tsx
│   │   │   ├── UserAccessRow.tsx
│   │   │   ├── FieldAccessToggle.tsx
│   │   │   └── RowRuleBuilder.tsx
│   │   │
│   │   └── CellEditors/        # редакторы по типу поля
│   │       ├── TextEditor.tsx
│   │       ├── NumberEditor.tsx
│   │       ├── SelectEditor.tsx
│   │       ├── DateEditor.tsx
│   │       └── ...
│   │
│   ├── api/
│   │   ├── auth.ts
│   │   ├── spreadsheet.ts
│   │   ├── rows.ts
│   │   └── permissions.ts
│   │
│   ├── store/                  # Zustand / Redux
│   │   ├── spreadsheetStore.ts
│   │   ├── permissionStore.ts
│   │   └── authStore.ts
│   │
│   └── hooks/
│       ├── usePermission.ts    # проверка прав на клиенте
│       └── useGridData.ts
```

---

## API — основные эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/auth/register` | Регистрация |
| `POST` | `/auth/login` | Вход, получение JWT |
| `GET` | `/workspaces` | Список рабочих пространств |
| `GET` | `/spreadsheets/:id` | Метаданные таблицы |
| `POST` | `/spreadsheets` | Создать таблицу |
| `POST` | `/spreadsheets/:id/fields` | Добавить поле |
| `PUT` | `/spreadsheets/:id/fields/:fid` | Изменить поле |
| `GET` | `/spreadsheets/:id/rows` | Получить строки |
| `POST` | `/spreadsheets/:id/rows` | Создать строку |
| `PATCH` | `/rows/:id/cells` | Обновить ячейки строки |
| `GET` | `/spreadsheets/:id/permissions` | Получить права |
| `PUT` | `/spreadsheets/:id/permissions` | Настроить права |
| `PUT` | `/fields/:id/permissions` | Права на поле |
| `PUT` | `/rows/:id/permissions` | Права на строку |

---

## Логика проверки прав

```
Запрос на чтение строки
        │
        ▼
 Есть доступ к таблице? ──No──► 403
        │ Yes
        ▼
 Есть row_access для строки?
   ├── Yes → применить
   └── No  → проверить row_access_rules (условия по полям)
        │
        ▼
 Для каждого поля в строке:
   └── Есть field_access? → скрыть или показать значение
        │
        ▼
 Вернуть строку с видимыми полями
```

---

## Ключевые технические решения

| Задача | Решение |
|--------|---------|
| Хранение разнотипных данных | `cell_values.value JSONB` |
| Проверка типа при записи | Валидация на уровне сервиса по `field.field_type` |
| Производительность чтения | Индексы на `(spreadsheet_id, row_id)`, партиционирование `cell_values` |
| Права в реальном времени | Проверка в middleware перед каждым запросом |
| Совместное редактирование | WebSocket + оптимистичные обновления на клиенте |
| Миграции схемы | `golang-migrate` или `Flyway` |

**User - --**

обнови структуру программы с учетом того, что данные пользовательских таблиц должны храниться в аналогичных таблицах бд. т.е. набор полей, описанный пользователем должен соответствовать полям таблицы бд

**Assistant - --**

# Обновлённая структура — динамические таблицы в PostgreSQL

## Ключевое архитектурное изменение

Вместо EAV (`cell_values` с JSONB) каждая пользовательская таблица создаёт **реальную таблицу в PostgreSQL** в отдельной схеме.

```
PostgreSQL
├── schema: auth          — пользователи, workspace
├── schema: meta          — метаданные таблиц, полей, прав
└── schema: data_{wid}    — пользовательские данные workspace
    ├── table_{sid_1}     — реальная таблица пользователя №1
    ├── table_{sid_2}     — реальная таблица пользователя №2
    └── ...
```

---

## Структура базы данных

### Схема `auth`

```sql
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

---

### Схема `meta` — метаданные структуры

```sql
-- Описание пользовательской таблицы
spreadsheets
├── id           UUID PK
├── workspace_id UUID FK → workspaces.id
├── name         TEXT NOT NULL          -- отображаемое имя
├── table_name   TEXT NOT NULL          -- имя таблицы в PG: 'tbl_<uuid>'
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
├── column_name    TEXT NOT NULL          -- имя колонки в PG: 'col_<uuid>'
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
```

---

### Схема `meta` — права доступа

```sql
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
```

---

### Схема `data_{workspace_id}` — пользовательские данные

Каждая пользовательская таблица создаётся динамически через DDL:

```sql
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
```

---

## Маппинг типов полей → типы PostgreSQL

| `field_type` | Тип PostgreSQL | Доп. constraint |
|---|---|---|
| `text` | `TEXT` | — |
| `integer` | `BIGINT` | — |
| `decimal` | `NUMERIC(18,6)` | — |
| `boolean` | `BOOLEAN` | — |
| `date` | `DATE` | — |
| `datetime` | `TIMESTAMPTZ` | — |
| `select` | `TEXT` | `CHECK (col IN ('opt1','opt2'))` |
| `multi_select` | `TEXT[]` | — |
| `email` | `TEXT` | `CHECK (col ~* '^[^@]+@[^@]+$')` |
| `url` | `TEXT` | `CHECK (col ~* '^https?://')` |
| `phone` | `TEXT` | — |
| `file` | `TEXT` | — (URL в object storage) |
| `relation` | `UUID` | `REFERENCES data_X."tbl_Y"(_id)` |

---

## Структура бэкенда

```
backend/
├── cmd/server/main.go
│
├── internal/
│   │
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   │
│   ├── workspace/
│   │   ├── handler.go
│   │   ├── service.go        -- создаёт схему data_{id} в PG
│   │   └── repository.go
│   │
│   ├── spreadsheet/
│   │   ├── handler.go        -- CRUD метаданных
│   │   ├── service.go        -- оркестрация: meta + DDL
│   │   └── repository.go     -- работа с meta.spreadsheets
│   │
│   ├── field/
│   │   ├── handler.go
│   │   ├── service.go        -- оркестрация: meta + DDL
│   │   ├── repository.go     -- работа с meta.fields
│   │   └── types/
│   │       ├── registry.go   -- маппинг field_type → SQL тип
│   │       ├── text.go
│   │       ├── number.go
│   │       ├── select.go     -- генерация CHECK constraint
│   │       ├── relation.go   -- генерация FK
│   │       └── ...
│   │
│   ├── ddl/                  -- ключевой модуль
│   │   ├── executor.go       -- выполнение DDL-запросов
│   │   ├── builder.go        -- построение DDL-выражений
│   │   │   -- CreateTable(spreadsheet, fields) string
│   │   │   -- AddColumn(field) string
│   │   │   -- AlterColumn(oldField, newField) string
│   │   │   -- DropColumn(field) string
│   │   │   -- RenameColumn(field, newName) string
│   │   └── migration.go      -- транзакционное применение DDL
│   │                         -- (DDL + обновление meta в одной транзакции)
│   │
│   ├── row/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go     -- динамические INSERT/UPDATE/SELECT
│   │       -- buildSelectQuery(spreadsheet, fields, filters, user)
│   │       -- buildInsertQuery(spreadsheet, fields, values)
│   │       -- buildUpdateQuery(spreadsheet, fields, values)
│   │
│   ├── query/                -- построение безопасных SQL-запросов
│   │   ├── builder.go        -- SELECT с учётом видимых полей
│   │   ├── filter.go         -- WHERE из JSON-фильтров
│   │   └── sort.go
│   │
│   ├── permission/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── enforcer.go       -- проверка прав перед каждой операцией
│   │       -- CanViewSpreadsheet(userID, spreadsheetID)
│   │       -- VisibleFields(userID, spreadsheetID) []Field
│   │       -- EditableFields(userID, spreadsheetID) []Field
│   │       -- ApplyRowRules(userID, spreadsheetID, query) query
│   │
│   └── middleware/
│       ├── auth.go
│       └── permission.go
│
├── pkg/
│   ├── postgres/
│   │   ├── pool.go           -- pgxpool
│   │   └── naming.go        -- генерация безопасных имён таблиц/колонок
│   │       -- TableName(spreadsheetID) → "tbl_<uuid_hex>"
│   │       -- ColumnName(fieldID)      → "col_<uuid_hex>"
│   │       -- SchemaName(workspaceID)  → "data_<uuid_hex>"
│   └── jwt/
│
└── migrations/               -- миграции системных схем
    ├── 001_auth_schema.sql
    ├── 002_meta_schema.sql
    └── 003_permissions.sql
```

---

## Жизненный цикл операций

### Создание таблицы пользователем

```
POST /spreadsheets
{
  "name": "Заявки",
  "fields": [
    {"name": "Название", "type": "text",    "required": true},
    {"name": "Сумма",    "type": "decimal"},
    {"name": "Статус",   "type": "select",  "options": {"choices": ["новая","в работе","закрыта"]}}
  ]
}

         │
         ▼
1. Валидация полей (field/types/)
         │
         ▼
2. BEGIN TRANSACTION
         │
         ├─ INSERT INTO meta.spreadsheets (...)       -- сохранить метаданные таблицы
         ├─ INSERT INTO meta.fields (...) × N         -- сохранить метаданные полей
         │
         ├─ DDL: CREATE TABLE data_w1."tbl_uuid" (    -- создать реальную таблицу
         │         _id UUID PK,
         │         _created_by UUID,
         │         _created_at TIMESTAMPTZ,
         │         "col_uuid1" TEXT NOT NULL,
         │         "col_uuid2" NUMERIC(18,6),
         │         "col_uuid3" TEXT CHECK (...)
         │       )
         │
         └─ COMMIT
```

> **Важно:** DDL в PostgreSQL выполняется внутри транзакции — при ошибке таблица не создаётся, метаданные не сохраняются.

---

### Добавление/изменение поля

```
POST /spreadsheets/:id/fields
{"name": "Дедлайн", "type": "date"}

         │
         ▼
1. BEGIN TRANSACTION
2. INSERT INTO meta.fields (...)
3. DDL: ALTER TABLE data_w1."tbl_uuid"
           ADD COLUMN "col_newuuid" DATE
4. COMMIT


PUT /spreadsheets/:id/fields/:fid
{"type": "datetime"}  -- смена типа

         │
         ▼
1. Проверить совместимость типов (date → datetime ✓)
2. BEGIN TRANSACTION
3. UPDATE meta.fields SET field_type = 'datetime'
4. DDL: ALTER TABLE data_w1."tbl_uuid"
           ALTER COLUMN "col_uuid"
           TYPE TIMESTAMPTZ
           USING "col_uuid"::TIMESTAMPTZ
5. COMMIT
```

---

### Чтение строк с учётом прав

```
GET /spreadsheets/:id/rows?filter=...&sort=...

         │
         ▼
1. permission.VisibleFields(userID, spreadsheetID)
   → ["col_b1c2", "col_d3e4"]   -- col_f5g6 скрыт для этого пользователя
         │
         ▼
2. permission.ApplyRowRules(userID, spreadsheetID, baseQuery)
   → добавляет WHERE "col_owner" = '<userID>'  -- если есть row rule
         │
         ▼
3. query.Build():
   SELECT _id, _created_at, "col_b1c2", "col_d3e4"
   FROM data_w1."tbl_a3f9..."
   WHERE "col_owner" = $1
     AND <пользовательские фильтры>
   ORDER BY _position
   LIMIT $2 OFFSET $3
         │
         ▼
4. Вернуть строки, заменив column_name → field.name в ответе
```

---

## Структура фронтенда

```
frontend/src/
│
├── pages/
│   ├── Login/
│   ├── Workspaces/
│   ├── Spreadsheet/
│   │   ├── index.tsx            -- страница таблицы
│   │   ├── GridView.tsx         -- основная сетка
│   │   └── FieldBuilder.tsx     -- конструктор полей
│   └── Settings/
│
├── components/
│   ├── Grid/
│   │   ├── Grid.tsx
│   │   ├── HeaderCell.tsx       -- тип поля, меню
│   │   ├── Cell.tsx             -- диспетчер по типу
│   │   └── RowActions.tsx
│   │
│   ├── FieldEditor/
│   │   ├── FieldTypeSelect.tsx
│   │   ├── FieldOptionsForm.tsx
│   │   └── TypeChangeWarning.tsx  -- предупреждение при смене типа
│   │
│   ├── CellEditors/             -- по одному на тип
│   │   ├── TextEditor.tsx
│   │   ├── NumberEditor.tsx
│   │   ├── SelectEditor.tsx
│   │   ├── DateEditor.tsx
│   │   ├── BooleanEditor.tsx
│   │   ├── RelationEditor.tsx
│   │   └── FileEditor.tsx
│   │
│   └── Permissions/
│       ├── ShareModal.tsx
│       ├── FieldAccessMatrix.tsx  -- таблица: пользователь × поле
│       └── RowRuleBuilder.tsx     -- конструктор условий для строк
│
├── api/
│   ├── spreadsheet.ts
│   ├── fields.ts
│   ├── rows.ts
│   └── permissions.ts
│
└── store/
    ├── spreadsheetStore.ts   -- метаданные + структура полей
    ├── rowStore.ts           -- данные строк, пагинация
    └── permissionStore.ts
```

---

## Ограничения и их решения

| Проблема | Решение |
|---|---|
| DDL вне транзакции в некоторых СУБД | PostgreSQL поддерживает DDL в транзакциях — используем это |
| Конфликт имён таблиц/колонок | Имена генерируются из UUID, изолированы по схемам |
| Смена типа поля с потерей данных | Матрица допустимых конверсий + предупреждение пользователю |
| Удаление поля с данными | Сначала `ALTER TABLE DROP COLUMN`, потом DELETE из `meta.fields` |
| Большое число схем | Партиционирование по workspace, мониторинг числа объектов в PG |
| SQL-инъекции через имена | Все имена таблиц/колонок — UUID-производные, без пользовательского ввода в DDL |

**User - --**

сгенерируй код фронтенда

**Assistant - --**

# Фронтенд — полная реализация

## Стек
- **React 18** + **TypeScript**
- **Zustand** — стейт-менеджмент
- **TanStack Query** — серверный стейт
- **TanStack Virtual** — виртуализация строк
- **Tailwind CSS** — стили

---

## Структура файлов

```
frontend/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── types/
│   │   └── index.ts
│   ├── api/
│   │   ├── client.ts
│   │   ├── auth.ts
│   │   ├── spreadsheets.ts
│   │   ├── fields.ts
│   │   ├── rows.ts
│   │   └── permissions.ts
│   ├── store/
│   │   ├── authStore.ts
│   │   ├── spreadsheetStore.ts
│   │   └── permissionStore.ts
│   ├── hooks/
│   │   ├── useSpreadsheet.ts
│   │   ├── useRows.ts
│   │   └── usePermissions.ts
│   ├── components/
│   │   ├── ui/
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Select.tsx
│   │   │   └── Dropdown.tsx
│   │   ├── Grid/
│   │   │   ├── Grid.tsx
│   │   │   ├── HeaderCell.tsx
│   │   │   ├── Cell.tsx
│   │   │   └── AddColumnButton.tsx
│   │   ├── CellEditors/
│   │   │   ├── index.ts
│   │   │   ├── TextEditor.tsx
│   │   │   ├── NumberEditor.tsx
│   │   │   ├── BooleanEditor.tsx
│   │   │   ├── DateEditor.tsx
│   │   │   ├── SelectEditor.tsx
│   │   │   └── RelationEditor.tsx
│   │   ├── FieldEditor/
│   │   │   ├── FieldEditorModal.tsx
│   │   │   ├── FieldTypeSelect.tsx
│   │   │   └── SelectOptionsEditor.tsx
│   │   └── Permissions/
│   │       ├── ShareModal.tsx
│   │       ├── FieldAccessMatrix.tsx
│   │       └── RowRuleBuilder.tsx
│   └── pages/
│       ├── LoginPage.tsx
│       ├── WorkspacesPage.tsx
│       └── SpreadsheetPage.tsx
```

---

## `src/types/index.ts`

```typescript
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
}

export interface FieldAccess {
  field_id: string;
  principal_id: string;
  can_view: boolean;
  can_edit: boolean;
}

export interface RowAccessRule {
  id: string;
  spreadsheet_id: string;
  principal_id: string;
  condition: {
    column_name: string;
    op: 'eq' | 'neq' | 'eq_current_user' | 'contains';
    value?: string;
  };
  can_view: boolean;
  can_edit: boolean;
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
}
```

---

## `src/api/client.ts`

```typescript
import axios, { AxiosError } from 'axios';

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8080',
  headers: { 'Content-Type': 'application/json' },
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

apiClient.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

---

## `src/api/auth.ts`

```typescript
import { apiClient } from './client';
import type { User } from '../types';

interface LoginPayload { email: string; password: string }
interface RegisterPayload { email: string; password: string; name: string }
interface AuthResponse { access_token: string; user: User }

export const authApi = {
  login: (data: LoginPayload) =>
    apiClient.post<AuthResponse>('/auth/login', data).then(r => r.data),

  register: (data: RegisterPayload) =>
    apiClient.post<AuthResponse>('/auth/register', data).then(r => r.data),

  me: () =>
    apiClient.get<User>('/auth/me').then(r => r.data),
};
```

---

## `src/api/spreadsheets.ts`

```typescript
import { apiClient } from './client';
import type { Spreadsheet } from '../types';

interface CreateSpreadsheetPayload {
  name: string;
  description?: string;
  workspace_id: string;
}

export const spreadsheetsApi = {
  list: (workspaceId: string) =>
    apiClient.get<Spreadsheet[]>(`/workspaces/${workspaceId}/spreadsheets`).then(r => r.data),

  get: (id: string) =>
    apiClient.get<Spreadsheet>(`/spreadsheets/${id}`).then(r => r.data),

  create: (data: CreateSpreadsheetPayload) =>
    apiClient.post<Spreadsheet>('/spreadsheets', data).then(r => r.data),

  update: (id: string, data: Partial<Pick<Spreadsheet, 'name' | 'description'>>) =>
    apiClient.put<Spreadsheet>(`/spreadsheets/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    apiClient.delete(`/spreadsheets/${id}`),
};
```

---

## `src/api/fields.ts`

```typescript
import { apiClient } from './client';
import type { Field, FieldType, FieldOptions } from '../types';

export interface CreateFieldPayload {
  name: string;
  field_type: FieldType;
  is_required?: boolean;
  is_unique?: boolean;
  default_value?: string;
  options?: FieldOptions;
}

export interface UpdateFieldPayload extends Partial<CreateFieldPayload> {
  position?: number;
}

export const fieldsApi = {
  create: (spreadsheetId: string, data: CreateFieldPayload) =>
    apiClient.post<Field>(`/spreadsheets/${spreadsheetId}/fields`, data).then(r => r.data),

  update: (fieldId: string, data: UpdateFieldPayload) =>
    apiClient.put<Field>(`/fields/${fieldId}`, data).then(r => r.data),

  delete: (fieldId: string) =>
    apiClient.delete(`/fields/${fieldId}`),

  reorder: (spreadsheetId: string, fieldIds: string[]) =>
    apiClient.patch(`/spreadsheets/${spreadsheetId}/fields/reorder`, { field_ids: fieldIds }),
};
```

---

## `src/api/rows.ts`

```typescript
import { apiClient } from './client';
import type { Row, PaginatedRows, FilterCondition, SortCondition } from '../types';

interface FetchRowsParams {
  limit?: number;
  offset?: number;
  filters?: FilterCondition[];
  sorts?: SortCondition[];
}

interface UpsertRowPayload {
  [column_name: string]: unknown;
}

export const rowsApi = {
  list: (spreadsheetId: string, params: FetchRowsParams = {}) =>
    apiClient.post<PaginatedRows>(`/spreadsheets/${spreadsheetId}/rows/query`, params).then(r => r.data),

  create: (spreadsheetId: string, data: UpsertRowPayload) =>
    apiClient.post<Row>(`/spreadsheets/${spreadsheetId}/rows`, data).then(r => r.data),

  update: (spreadsheetId: string, rowId: string, data: UpsertRowPayload) =>
    apiClient.patch<Row>(`/spreadsheets/${spreadsheetId}/rows/${rowId}`, data).then(r => r.data),

  delete: (spreadsheetId: string, rowId: string) =>
    apiClient.delete(`/spreadsheets/${spreadsheetId}/rows/${rowId}`),

  bulkDelete: (spreadsheetId: string, rowIds: string[]) =>
    apiClient.post(`/spreadsheets/${spreadsheetId}/rows/bulk-delete`, { row_ids: rowIds }),
};
```

---

## `src/api/permissions.ts`

```typescript
import { apiClient } from './client';
import type { SpreadsheetAccess, FieldAccess, RowAccessRule } from '../types';

export const permissionsApi = {
  getSpreadsheetAccess: (spreadsheetId: string) =>
    apiClient.get<SpreadsheetAccess[]>(`/spreadsheets/${spreadsheetId}/permissions`).then(r => r.data),

  upsertSpreadsheetAccess: (spreadsheetId: string, data: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
    apiClient.put(`/spreadsheets/${spreadsheetId}/permissions`, data),

  removeSpreadsheetAccess: (spreadsheetId: string, principalId: string) =>
    apiClient.delete(`/spreadsheets/${spreadsheetId}/permissions/${principalId}`),

  getFieldAccess: (spreadsheetId: string) =>
    apiClient.get<FieldAccess[]>(`/spreadsheets/${spreadsheetId}/field-permissions`).then(r => r.data),

  upsertFieldAccess: (fieldId: string, data: Omit<FieldAccess, 'field_id'>) =>
    apiClient.put(`/fields/${fieldId}/permissions`, data),

  getRowRules: (spreadsheetId: string) =>
    apiClient.get<RowAccessRule[]>(`/spreadsheets/${spreadsheetId}/row-rules`).then(r => r.data),

  upsertRowRule: (spreadsheetId: string, data: Omit<RowAccessRule, 'id'>) =>
    apiClient.put(`/spreadsheets/${spreadsheetId}/row-rules`, data),

  deleteRowRule: (spreadsheetId: string, ruleId: string) =>
    apiClient.delete(`/spreadsheets/${spreadsheetId}/row-rules/${ruleId}`),
};
```

---

## `src/store/authStore.ts`

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types';

interface AuthState {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      setAuth: (user, token) => {
        localStorage.setItem('access_token', token);
        set({ user, token });
      },
      logout: () => {
        localStorage.removeItem('access_token');
        set({ user: null, token: null });
      },
      isAuthenticated: () => Boolean(get().token),
    }),
    { name: 'auth-storage', partialize: (s) => ({ token: s.token, user: s.user }) }
  )
);
```

---

## `src/store/spreadsheetStore.ts`

```typescript
import { create } from 'zustand';
import type { Spreadsheet, Field, FilterCondition, SortCondition } from '../types';

interface SpreadsheetState {
  spreadsheet: Spreadsheet | null;
  selectedRowIds: Set<string>;
  editingCell: { rowId: string; fieldId: string } | null;
  filters: FilterCondition[];
  sorts: SortCondition[];

  setSpreadsheet: (s: Spreadsheet) => void;
  updateField: (field: Field) => void;
  addField: (field: Field) => void;
  removeField: (fieldId: string) => void;

  setEditingCell: (rowId: string, fieldId: string) => void;
  clearEditingCell: () => void;

  toggleRowSelection: (rowId: string) => void;
  clearSelection: () => void;

  setFilters: (filters: FilterCondition[]) => void;
  setSorts: (sorts: SortCondition[]) => void;
}

export const useSpreadsheetStore = create<SpreadsheetState>((set) => ({
  spreadsheet: null,
  selectedRowIds: new Set(),
  editingCell: null,
  filters: [],
  sorts: [],

  setSpreadsheet: (spreadsheet) => set({ spreadsheet }),

  updateField: (field) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? {
            ...s.spreadsheet,
            fields: s.spreadsheet.fields.map((f) => (f.id === field.id ? field : f)),
          }
        : null,
    })),

  addField: (field) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? { ...s.spreadsheet, fields: [...s.spreadsheet.fields, field] }
        : null,
    })),

  removeField: (fieldId) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? { ...s.spreadsheet, fields: s.spreadsheet.fields.filter((f) => f.id !== fieldId) }
        : null,
    })),

  setEditingCell: (rowId, fieldId) => set({ editingCell: { rowId, fieldId } }),
  clearEditingCell: () => set({ editingCell: null }),

  toggleRowSelection: (rowId) =>
    set((s) => {
      const next = new Set(s.selectedRowIds);
      next.has(rowId) ? next.delete(rowId) : next.add(rowId);
      return { selectedRowIds: next };
    }),

  clearSelection: () => set({ selectedRowIds: new Set() }),

  setFilters: (filters) => set({ filters }),
  setSorts: (sorts) => set({ sorts }),
}));
```

---

## `src/hooks/useRows.ts`

```typescript
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rowsApi } from '../api/rows';
import { useSpreadsheetStore } from '../store/spreadsheetStore';

const PAGE_SIZE = 50;

export function useRows(spreadsheetId: string) {
  const { filters, sorts } = useSpreadsheetStore();
  const qc = useQueryClient();

  const query = useInfiniteQuery({
    queryKey: ['rows', spreadsheetId, filters, sorts],
    queryFn: ({ pageParam = 0 }) =>
      rowsApi.list(spreadsheetId, {
        limit: PAGE_SIZE,
        offset: pageParam as number,
        filters,
        sorts,
      }),
    getNextPageParam: (last, pages) => {
      const loaded = pages.length * PAGE_SIZE;
      return loaded < last.total ? loaded : undefined;
    },
    initialPageParam: 0,
  });

  const allRows = query.data?.pages.flatMap((p) => p.data) ?? [];

  const createRow = useMutation({
    mutationFn: (data: Record<string, unknown>) => rowsApi.create(spreadsheetId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  const updateRow = useMutation({
    mutationFn: ({ rowId, data }: { rowId: string; data: Record<string, unknown> }) =>
      rowsApi.update(spreadsheetId, rowId, data),
    onMutate: async ({ rowId, data }) => {
      await qc.cancelQueries({ queryKey: ['rows', spreadsheetId] });
      const prev = qc.getQueryData(['rows', spreadsheetId]);
      qc.setQueryData(['rows', spreadsheetId, filters, sorts], (old: typeof query.data) => ({
        ...old,
        pages: old?.pages.map((page) => ({
          ...page,
          data: page.data.map((row) => (row._id === rowId ? { ...row, ...data } : row)),
        })),
      }));
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(['rows', spreadsheetId], ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  const deleteRow = useMutation({
    mutationFn: (rowId: string) => rowsApi.delete(spreadsheetId, rowId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  return { ...query, allRows, createRow, updateRow, deleteRow };
}
```

---

## `src/components/ui/Modal.tsx`

```typescript
import React, { useEffect } from 'react';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

const sizeClass = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-2xl',
};

export const Modal: React.FC<ModalProps> = ({ open, onClose, title, children, size = 'md' }) => {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [onClose]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className={`relative z-10 w-full ${sizeClass[size]} mx-4 bg-white rounded-xl shadow-2xl`}>
        <div className="flex items-center justify-between px-6 py-4 border-b">
          <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
        </div>
        <div className="px-6 py-4">{children}</div>
      </div>
    </div>
  );
};
```

---

## `src/components/ui/Button.tsx`

```typescript
import React from 'react';
import { clsx } from 'clsx';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
}

const variantClass: Record<Variant, string> = {
  primary:   'bg-blue-600 text-white hover:bg-blue-700 disabled:bg-blue-300',
  secondary: 'bg-gray-100 text-gray-800 hover:bg-gray-200',
  ghost:     'bg-transparent text-gray-600 hover:bg-gray-100',
  danger:    'bg-red-600 text-white hover:bg-red-700',
};

const sizeClass: Record<Size, string> = {
  sm: 'px-2.5 py-1.5 text-sm',
  md: 'px-4 py-2 text-sm',
  lg: 'px-5 py-2.5 text-base',
};

export const Button: React.FC<ButtonProps> = ({
  variant = 'primary', size = 'md', loading, children, className, disabled, ...props
}) => (
  <button
    {...props}
    disabled={disabled || loading}
    className={clsx(
      'inline-flex items-center gap-2 font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500',
      variantClass[variant],
      sizeClass[size],
      className
    )}
  >
    {loading && (
      <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
      </svg>
    )}
    {children}
  </button>
);
```

---

## `src/components/CellEditors/TextEditor.tsx`

```typescript
import React, { useRef, useEffect } from 'react';

interface Props {
  value: string;
  onChange: (v: string) => void;
  onCommit: () => void;
  onCancel: () => void;
}

export const TextEditor: React.FC<Props> = ({ value, onChange, onCommit, onCancel }) => {
  const ref = useRef<HTMLInputElement>(null);

  useEffect(() => { ref.current?.focus(); ref.current?.select(); }, []);

  return (
    <input
      ref={ref}
      value={value ?? ''}
      onChange={(e) => onChange(e.target.value)}
      onBlur={onCommit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onCommit();
        if (e.key === 'Escape') onCancel();
      }}
      className="w-full h-full px-2 py-1 text-sm border-2 border-blue-500 rounded outline-none"
    />
  );
};
```

---

## `src/components/CellEditors/NumberEditor.tsx`

```typescript
import React, { useRef, useEffect } from 'react';

interface Props {
  value: number | null;
  onChange: (v: number | null) => void;
  onCommit: () => void;
  onCancel: () => void;
  isDecimal?: boolean;
}

export const NumberEditor: React.FC<Props> = ({ value, onChange, onCommit, onCancel, isDecimal }) => {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => { ref.current?.focus(); ref.current?.select(); }, []);

  return (
    <input
      ref={ref}
      type="number"
      step={isDecimal ? '0.000001' : '1'}
      value={value ?? ''}
      onChange={(e) => {
        const v = e.target.value === '' ? null : Number(e.target.value);
        onChange(v);
      }}
      onBlur={onCommit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onCommit();
        if (e.key === 'Escape') onCancel();
      }}
      className="w-full h-full px-2 py-1 text-sm text-right border-2 border-blue-500 rounded outline-none"
    />
  );
};
```

---

## `src/components/CellEditors/BooleanEditor.tsx`

```typescript
import React from 'react';

interface Props {
  value: boolean;
  onChange: (v: boolean) => void;
  onCommit: () => void;
}

export const BooleanEditor: React.FC<Props> = ({ value, onChange, onCommit }) => (
  <div className="flex items-center justify-center w-full h-full">
    <input
      type="checkbox"
      checked={Boolean(value)}
      onChange={(e) => { onChange(e.target.checked); onCommit(); }}
      className="w-4 h-4 text-blue-600 cursor-pointer"
      autoFocus
    />
  </div>
);
```

---

## `src/components/CellEditors/DateEditor.tsx`

```typescript
import React, { useRef, useEffect } from 'react';

interface Props {
  value: string;
  onChange: (v: string) => void;
  onCommit: () => void;
  onCancel: () => void;
  withTime?: boolean;
}

export const DateEditor: React.FC<Props> = ({ value, onChange, onCommit, onCancel, withTime }) => {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => { ref.current?.focus(); }, []);

  return (
    <input
      ref={ref}
      type={withTime ? 'datetime-local' : 'date'}
      value={value ?? ''}
      onChange={(e) => onChange(e.target.value)}
      onBlur={onCommit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onCommit();
        if (e.key === 'Escape') onCancel();
      }}
      className="w-full h-full px-2 py-1 text-sm border-2 border-blue-500 rounded outline-none"
    />
  );
};
```

---

## `src/components/CellEditors/SelectEditor.tsx`

```typescript
import React, { useRef, useEffect } from 'react';
import type { SelectOption } from '../../types';

interface Props {
  value: string | string[];
  options: SelectOption[];
  isMulti?: boolean;
  onChange: (v: string | string[]) => void;
  onCommit: () => void;
  onCancel: () => void;
}

export const SelectEditor: React.FC<Props> = ({ value, options, isMulti, onChange, onCommit, onCancel }) => {
  const ref = useRef<HTMLSelectElement>(null);
  useEffect(() => { ref.current?.focus(); }, []);

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    if (isMulti) {
      const selected = Array.from(e.target.selectedOptions).map(o => o.value);
      onChange(selected);
    } else {
      onChange(e.target.value);
      onCommit();
    }
  };

  return (
    <select
      ref={ref}
      multiple={isMulti}
      value={isMulti ? (value as string[]) : (value as string)}
      onChange={handleChange}
      onBlur={onCommit}
      onKeyDown={(e) => { if (e.key === 'Escape') onCancel(); }}
      className="w-full h-full px-2 py-1 text-sm border-2 border-blue-500 rounded outline-none bg-white"
    >
      {!isMulti && <option value="">— не выбрано —</option>}
      {options.map(opt => (
        <option key={opt.value} value={opt.value}>{opt.label}</option>
      ))}
    </select>
  );
};
```

---

## `src/components/CellEditors/index.ts`

```typescript
import type { FieldType, Field } from '../../types';
import { TextEditor } from './TextEditor';
import { NumberEditor } from './NumberEditor';
import { BooleanEditor } from './BooleanEditor';
import { DateEditor } from './DateEditor';
import { SelectEditor } from './SelectEditor';

export { TextEditor, NumberEditor, BooleanEditor, DateEditor, SelectEditor };

export function getEditor(fieldType: FieldType) {
  const map: Partial<Record<FieldType, unknown>> = {
    text: TextEditor,
    email: TextEditor,
    url: TextEditor,
    phone: TextEditor,
    file: TextEditor,
    integer: NumberEditor,
    decimal: NumberEditor,
    boolean: BooleanEditor,
    date: DateEditor,
    datetime: DateEditor,
    select: SelectEditor,
    multi_select: SelectEditor,
  };
  return map[fieldType] ?? TextEditor;
}

export function formatCellValue(value: unknown, field: Field): string {
  if (value === null || value === undefined) return '';

  switch (field.field_type) {
    case 'boolean':
      return value ? '✓' : '✗';
    case 'date':
      return value ? new Date(value as string).toLocaleDateString('ru-RU') : '';
    case 'datetime':
      return value ? new Date(value as string).toLocaleString('ru-RU') : '';
    case 'decimal':
      return typeof value === 'number' ? value.toLocaleString('ru-RU', { minimumFractionDigits: 2 }) : String(value);
    case 'select': {
      const opt = field.options?.choices?.find(c => c.value === value);
      return opt?.label ?? String(value);
    }
    case 'multi_select': {
      const arr = Array.isArray(value) ? value : [value];
      return arr.map(v => field.options?.choices?.find(c => c.value === v)?.label ?? v).join(', ');
    }
    default:
      return String(value);
  }
}
```

---

## `src/components/Grid/Cell.tsx`

```typescript
import React, { useState, useCallback } from 'react';
import type { Field, Row } from '../../types';
import { formatCellValue, TextEditor, NumberEditor, BooleanEditor, DateEditor, SelectEditor } from '../CellEditors';
import { useSpreadsheetStore } from '../../store/spreadsheetStore';

interface Props {
  row: Row;
  field: Field;
  canEdit: boolean;
  onUpdate: (rowId: string, colName: string, value: unknown) => void;
}

export const Cell: React.FC<Props> = React.memo(({ row, field, canEdit, onUpdate }) => {
  const { editingCell, setEditingCell, clearEditingCell } = useSpreadsheetStore();
  const [localValue, setLocalValue] = useState<unknown>(null);

  const isEditing =
    editingCell?.rowId === row._id && editingCell?.fieldId === field.id;

  const rawValue = row[field.column_name];

  const startEdit = useCallback(() => {
    if (!canEdit) return;
    setLocalValue(rawValue);
    setEditingCell(row._id, field.id);
  }, [canEdit, rawValue, row._id, field.id, setEditingCell]);

  const commit = useCallback(() => {
    if (localValue !== rawValue) {
      onUpdate(row._id, field.column_name, localValue);
    }
    clearEditingCell();
  }, [localValue, rawValue, row._id, field.column_name, onUpdate, clearEditingCell]);

  const cancel = useCallback(() => {
    clearEditingCell();
  }, [clearEditingCell]);

  const editorProps = {
    value: localValue as never,
    onChange: setLocalValue as never,
    onCommit: commit,
    onCancel: cancel,
  };

  const renderEditor = () => {
    switch (field.field_type) {
      case 'integer':
        return <NumberEditor {...editorProps} isDecimal={false} />;
      case 'decimal':
        return <NumberEditor {...editorProps} isDecimal />;
      case 'boolean':
        return <BooleanEditor {...editorProps} />;
      case 'date':
        return <DateEditor {...editorProps} />;
      case 'datetime':
        return <DateEditor {...editorProps} withTime />;
      case 'select':
        return <SelectEditor {...editorProps} options={field.options?.choices ?? []} />;
      case 'multi_select':
        return <SelectEditor {...editorProps} options={field.options?.choices ?? []} isMulti />;
      default:
        return <TextEditor {...editorProps} />;
    }
  };

  const getBadgeColor = (value: string) => {
    const opt = field.options?.choices?.find(c => c.value === value);
    return opt?.color ?? '#e5e7eb';
  };

  const renderValue = () => {
    if (rawValue === null || rawValue === undefined) {
      return <span className="text-gray-300 text-xs">—</span>;
    }
    if (field.field_type === 'boolean') {
      return (
        <span className={`text-lg ${rawValue ? 'text-green-500' : 'text-gray-300'}`}>
          {rawValue ? '✓' : '✗'}
        </span>
      );
    }
    if (field.field_type === 'select') {
      return (
        <span
          className="px-2 py-0.5 rounded-full text-xs font-medium"
          style={{ backgroundColor: getBadgeColor(rawValue as string) + '33', color: getBadgeColor(rawValue as string) }}
        >
          {formatCellValue(rawValue, field)}
        </span>
      );
    }
    if (field.field_type === 'multi_select') {
      const arr = Array.isArray(rawValue) ? rawValue : [rawValue];
      return (
        <div className="flex flex-wrap gap-1">
          {arr.map((v, i) => (
            <span key={i} className="px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded text-xs">
              {field.options?.choices?.find(c => c.value === v)?.label ?? String(v)}
            </span>
          ))}
        </div>
      );
    }
    return <span className="text-sm text-gray-800 truncate">{formatCellValue(rawValue, field)}</span>;
  };

  return (
    <div
      className={`
        relative w-full h-full min-h-[36px] flex items-center px-2
        border-r border-gray-200
        ${isEditing ? 'p-0' : ''}
        ${canEdit && !isEditing ? 'cursor-pointer hover:bg-blue-50' : ''}
      `}
      onDoubleClick={startEdit}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === 'F2') startEdit(); }}
      tabIndex={canEdit ? 0 : -1}
    >
      {isEditing ? renderEditor() : renderValue()}
    </div>
  );
});

Cell.displayName = 'Cell';
```

---

## `src/components/Grid/HeaderCell.tsx`

```typescript
import React, { useState } from 'react';
import type { Field } from '../../types';
import { Dropdown } from '../ui/Dropdown';
import { FieldEditorModal } from '../FieldEditor/FieldEditorModal';

const FIELD_TYPE_ICONS: Record<string, string> = {
  text: 'T',  integer: '#',  decimal: '.#',  boolean: '☑',
  date: '📅', datetime: '🕐', select: '▾',  multi_select: '▾▾',
  email: '@', url: '🔗',     phone: '☎',    file: '📎',
  relation: '↗',
};

interface Props {
  field: Field;
  canManage: boolean;
  onUpdate: (field: Field) => void;
  onDelete: (fieldId: string) => void;
  onSort: (fieldId: string, direction: 'asc' | 'desc') => void;
}

export const HeaderCell: React.FC<Props> = ({ field, canManage, onUpdate, onDelete, onSort }) => {
  const [editOpen, setEditOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  const menuItems = [
    { label: '↑ Сортировать А→Я', onClick: () => onSort(field.id, 'asc') },
    { label: '↓ Сортировать Я→А', onClick: () => onSort(field.id, 'desc') },
    ...(canManage ? [
      { label: '✏️ Редактировать поле', onClick: () => setEditOpen(true) },
      { label: '🗑️ Удалить поле', onClick: () => onDelete(field.id), danger: true },
    ] : []),
  ];

  return (
    <>
      <div
        className="flex items-center justify-between px-2 h-9 bg-gray-50 border-r border-gray-200 select-none group"
      >
        <div className="flex items-center gap-1.5 min-w-0">
          <span className="text-xs text-gray-400 font-mono shrink-0">
            {FIELD_TYPE_ICONS[field.field_type] ?? '?'}
          </span>
          <span className="text-xs font-medium text-gray-700 truncate">{field.name}</span>
          {field.is_required && <span className="text-red-500 text-xs shrink-0">*</span>}
        </div>
        <button
          onClick={(e) => { e.stopPropagation(); setMenuOpen(true); }}
          className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-gray-600 px-1 rounded"
        >
          ⋮
        </button>
      </div>

      {menuOpen && (
        <Dropdown items={menuItems} onClose={() => setMenuOpen(false)} />
      )}

      {editOpen && (
        <FieldEditorModal
          field={field}
          onClose={() => setEditOpen(false)}
          onSave={(updated) => { onUpdate(updated); setEditOpen(false); }}
        />
      )}
    </>
  );
};
```

---

## `src/components/Grid/Grid.tsx`

```typescript
import React, { useCallback, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import type { Field, Row } from '../../types';
import { Cell } from './Cell';
import { HeaderCell } from './HeaderCell';
import { AddColumnButton } from './AddColumnButton';

const ROW_HEIGHT = 36;
const INDEX_COL_WIDTH = 48;
const DEFAULT_COL_WIDTH = 200;

interface Props {
  fields: Field[];
  rows: Row[];
  canEdit: boolean;
  canManage: boolean;
  editableFields: Set<string>;
  selectedRowIds: Set<string>;
  hasNextPage?: boolean;
  isFetchingNextPage?: boolean;
  onFetchNextPage: () => void;
  onUpdateCell: (rowId: string, colName: string, value: unknown) => void;
  onAddRow: () => void;
  onDeleteRow: (rowId: string) => void;
  onToggleRow: (rowId: string) => void;
  onUpdateField: (field: Field) => void;
  onDeleteField: (fieldId: string) => void;
  onAddField: () => void;
  onSort: (fieldId: string, direction: 'asc' | 'desc') => void;
}

export const Grid: React.FC<Props> = ({
  fields, rows, canEdit, canManage, editableFields,
  selectedRowIds, hasNextPage, isFetchingNextPage,
  onFetchNextPage, onUpdateCell, onAddRow, onDeleteRow,
  onToggleRow, onUpdateField, onDeleteField, onAddField, onSort,
}) => {
  const parentRef = useRef<HTMLDivElement>(null);

  const totalWidth = INDEX_COL_WIDTH + fields.length * DEFAULT_COL_WIDTH + 80;

  const virtualizer = useVirtualizer({
    count: hasNextPage ? rows.length + 1 : rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  });

  const handleScroll = useCallback(() => {
    const el = parentRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 200;
    if (nearBottom && hasNextPage && !isFetchingNextPage) onFetchNextPage();
  }, [hasNextPage, isFetchingNextPage, onFetchNextPage]);

  return (
    <div className="flex flex-col h-full border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div
        className="flex border-b border-gray-200 bg-gray-50 sticky top-0 z-10"
        style={{ minWidth: totalWidth }}
      >
        {/* Index column */}
        <div
          className="flex items-center justify-center border-r border-gray-200 bg-gray-50 shrink-0"
          style={{ width: INDEX_COL_WIDTH }}
        >
          <input
            type="checkbox"
            onChange={(e) => rows.forEach(r => {
              if (e.target.checked !== selectedRowIds.has(r._id)) onToggleRow(r._id);
            })}
            checked={rows.length > 0 && rows.every(r => selectedRowIds.has(r._id))}
            className="w-3.5 h-3.5"
          />
        </div>

        {/* Field headers */}
        {fields.map((field) => (
          <div key={field.id} style={{ width: DEFAULT_COL_WIDTH, minWidth: DEFAULT_COL_WIDTH }}>
            <HeaderCell
              field={field}
              canManage={canManage}
              onUpdate={onUpdateField}
              onDelete={onDeleteField}
              onSort={onSort}
            />
          </div>
        ))}

        {canManage && <AddColumnButton onClick={onAddField} />}
      </div>

      {/* Body */}
      <div
        ref={parentRef}
        onScroll={handleScroll}
        className="flex-1 overflow-auto"
        style={{ minWidth: totalWidth }}
      >
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((vItem) => {
            const isLoader = vItem.index === rows.length;

            if (isLoader) {
              return (
                <div
                  key="loader"
                  style={{ position: 'absolute', top: vItem.start, height: ROW_HEIGHT, width: '100%' }}
                  className="flex items-center justify-center text-sm text-gray-400"
                >
                  Загрузка…
                </div>
              );
            }

            const row = rows[vItem.index];
            const isSelected = selectedRowIds.has(row._id);

            return (
              <div
                key={row._id}
                style={{ position: 'absolute', top: vItem.start, height: ROW_HEIGHT, width: '100%' }}
                className={`flex border-b border-gray-100 group ${isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'}`}
              >
                {/* Index + checkbox */}
                <div
                  className="flex items-center justify-center gap-1 border-r border-gray-200 shrink-0 text-xs text-gray-400"
                  style={{ width: INDEX_COL_WIDTH }}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => onToggleRow(row._id)}
                    className="w-3.5 h-3.5 opacity-0 group-hover:opacity-100"
                    onClick={(e) => e.stopPropagation()}
                  />
                  <span className="group-hover:hidden">{vItem.index + 1}</span>
                </div>

                {/* Cells */}
                {fields.map((field) => (
                  <div
                    key={field.id}
                    style={{ width: DEFAULT_COL_WIDTH, minWidth: DEFAULT_COL_WIDTH }}
                  >
                    <Cell
                      row={row}
                      field={field}
                      canEdit={canEdit && editableFields.has(field.id)}
                      onUpdate={onUpdateCell}
                    />
                  </div>
                ))}

                {/* Row actions */}
                {canEdit && (
                  <div className="flex items-center px-1 opacity-0 group-hover:opacity-100">
                    <button
                      onClick={() => onDeleteRow(row._id)}
                      className="text-red-400 hover:text-red-600 text-xs px-1"
                      title="Удалить строку"
                    >
                      ✕
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Footer — add row */}
      {canEdit && (
        <div className="border-t border-gray-200">
          <button
            onClick={onAddRow}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-50 w-full"
          >
            <span className="text-lg leading-none">+</span> Добавить строку
          </button>
        </div>
      )}
    </div>
  );
};
```

---

## `src/components/Grid/AddColumnButton.tsx`

```typescript
import React from 'react';

export const AddColumnButton: React.FC<{ onClick: () => void }> = ({ onClick }) => (
  <button
    onClick={onClick}
    className="flex items-center justify-center w-12 h-9 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors shrink-0"
    title="Добавить поле"
  >
    <span className="text-xl leading-none">+</span>
  </button>
);
```

---

## `src/components/FieldEditor/FieldEditorModal.tsx`

```typescript
import React, { useState } from 'react';
import type { Field, FieldType, FieldOptions } from '../../types';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { FieldTypeSelect } from './FieldTypeSelect';
import { SelectOptionsEditor } from './SelectOptionsEditor';
import { fieldsApi } from '../../api/fields';

interface Props {
  field?: Field;
  spreadsheetId?: string;
  onClose: () => void;
  onSave: (field: Field) => void;
}

const TYPE_CHANGE_COMPATIBILITY: Partial<Record<FieldType, FieldType[]>> = {
  text:     ['email', 'url', 'phone'],
  integer:  ['decimal', 'text'],
  decimal:  ['text'],
  date:     ['datetime', 'text'],
  datetime: ['text'],
  email:    ['text'],
  url:      ['text'],
  phone:    ['text'],
};

export const FieldEditorModal: React.FC<Props> = ({ field, spreadsheetId, onClose, onSave }) => {
  const isEdit = Boolean(field);

  const [name, setName] = useState(field?.name ?? '');
  const [fieldType, setFieldType] = useState<FieldType>(field?.field_type ?? 'text');
  const [isRequired, setIsRequired] = useState(field?.is_required ?? false);
  const [isUnique, setIsUnique] = useState(field?.is_unique ?? false);
  const [options, setOptions] = useState<FieldOptions>(field?.options ?? {});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [typeWarning, setTypeWarning] = useState('');

  const handleTypeChange = (newType: FieldType) => {
    if (isEdit && field) {
      const compatible = TYPE_CHANGE_COMPATIBILITY[field.field_type] ?? [];
      if (!compatible.includes(newType)) {
        setTypeWarning(`Смена типа с «${field.field_type}» на «${newType}» может привести к потере данных!`);
      } else {
        setTypeWarning('');
      }
    }
    setFieldType(newType);
  };

  const handleSave = async () => {
    if (!name.trim()) { setError('Введите название поля'); return; }
    setLoading(true);
    setError('');
    try {
      const payload = { name: name.trim(), field_type: fieldType, is_required: isRequired, is_unique: isUnique, options };
      let result: Field;
      if (isEdit && field) {
        result = await fieldsApi.update(field.id, payload);
      } else {
        result = await fieldsApi.create(spreadsheetId!, payload);
      }
      onSave(result);
    } catch (e: unknown) {
      setError((e as Error).message ?? 'Ошибка сохранения');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      open
      onClose={onClose}
      title={isEdit ? `Редактировать поле «${field?.name}»` : 'Новое поле'}
    >
      <div className="space-y-4">
        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Название</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Название поля"
            className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"
          />
        </div>

        {/* Type */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Тип поля</label>
          <FieldTypeSelect value={fieldType} onChange={handleTypeChange} />
        </div>

        {typeWarning && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 text-sm text-yellow-800">
            ⚠️ {typeWarning}
          </div>
        )}

        {/* Select options */}
        {(fieldType === 'select' || fieldType === 'multi_select') && (
          <SelectOptionsEditor
            choices={options.choices ?? []}
            onChange={(choices) => setOptions((o) => ({ ...o, choices }))}
          />
        )}

        {/* Flags */}
        <div className="flex gap-4">
          <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
            <input
              type="checkbox"
              checked={isRequired}
              onChange={(e) => setIsRequired(e.target.checked)}
              className="w-4 h-4"
            />
            Обязательное
          </label>
          <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
            <input
              type="checkbox"
              checked={isUnique}
              onChange={(e) => setIsUnique(e.target.checked)}
              className="w-4 h-4"
            />
            Уникальное
          </label>
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <div className="flex justify-end gap-2 pt-2">
          <Button variant="secondary" onClick={onClose}>Отмена</Button>
          <Button onClick={handleSave} loading={loading}>
            {isEdit ? 'Сохранить' : 'Создать поле'}
          </Button>
        </div>
      </div>
    </Modal>
  );
};
```

---

## `src/components/FieldEditor/FieldTypeSelect.tsx`

```typescript
import React from 'react';
import type { FieldType } from '../../types';

const FIELD_TYPES: { value: FieldType; label: string; icon: string; group: string }[] = [
  { value: 'text',         label: 'Текст',           icon: 'T',  group: 'Основные' },
  { value: 'integer',      label: 'Целое число',      icon: '#',  group: 'Основные' },
  { value: 'decimal',      label: 'Десятичное число', icon: '.#', group: 'Основные' },
  { value: 'boolean',      label: 'Да / Нет',         icon: '☑',  group: 'Основные' },
  { value: 'date',         label: 'Дата',             icon: '📅', group: 'Дата и время' },
  { value: 'datetime',     label: 'Дата и время',     icon: '🕐', group: 'Дата и время' },
  { value: 'select',       label: 'Выбор (один)',     icon: '▾',  group: 'Выбор' },
  { value: 'multi_select', label: 'Выбор (несколько)', icon: '▾▾', group: 'Выбор' },
  { value: 'email',        label: 'Email',            icon: '@',  group: 'Контакты' },
  { value: 'phone',        label: 'Телефон',          icon: '☎',  group: 'Контакты' },
  { value: 'url',          label: 'Ссылка',           icon: '🔗', group: 'Контакты' },
  { value: 'file',         label: 'Файл',             icon: '📎', group: 'Прочее' },
  { value: 'relation',     label: 'Связь',            icon: '↗',  group: 'Прочее' },
];

const GROUPS = ['Основные', 'Дата и время', 'Выбор', 'Контакты', 'Прочее'];

interface Props {
  value: FieldType;
  onChange: (v: FieldType) => void;
}

export const FieldTypeSelect: React.FC<Props> = ({ value, onChange }) => {
  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden max-h-48 overflow-y-auto">
      {GROUPS.map(group => {
        const items = FIELD_TYPES.filter(t => t.group === group);
        return (
          <div key={group}>
            <div className="px-3 py-1 bg-gray-50 text-xs font-medium text-gray-400 uppercase tracking-wide">
              {group}
            </div>
            {items.map(type => (
              <button
                key={type.value}
                type="button"
                onClick={() => onChange(type.value)}
                className={`w-full flex items-center gap-3 px-3 py-2 text-sm text-left transition-colors
                  ${value === type.value
                    ? 'bg-blue-50 text-blue-700 font-medium'
                    : 'text-gray-700 hover:bg-gray-50'
                  }`}
              >
                <span className="w-5 text-center text-base">{type.icon}</span>
                {type.label}
                {value === type.value && <span className="ml-auto text-blue-500">✓</span>}
              </button>
            ))}
          </div>
        );
      })}
    </div>
  );
};
```

---

## `src/components/FieldEditor/SelectOptionsEditor.tsx`

```typescript
import React, { useState } from 'react';
import type { SelectOption } from '../../types';

const COLORS = ['#ef4444','#f97316','#eab308','#22c55e','#3b82f6','#8b5cf6','#ec4899','#6b7280'];

interface Props {
  choices: SelectOption[];
  onChange: (choices: SelectOption[]) => void;
}

export const SelectOptionsEditor: React.FC<Props> = ({ choices, onChange }) => {
  const [newLabel, setNewLabel] = useState('');
  const [selectedColor, setSelectedColor] = useState(COLORS[0]);

  const addChoice = () => {
    if (!newLabel.trim()) return;
    const choice: SelectOption = {
      value: newLabel.trim().toLowerCase().replace(/\s+/g, '_'),
      label: newLabel.trim(),
      color: selectedColor,
    };
    onChange([...choices, choice]);
    setNewLabel('');
  };

  const removeChoice = (value: string) => {
    onChange(choices.filter(c => c.value !== value));
  };

  const updateColor = (value: string, color: string) => {
    onChange(choices.map(c => c.value === value ? { ...c, color } : c));
  };

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-gray-700">Варианты</label>

      <div className="space-y-1 max-h-40 overflow-y-auto">
        {choices.map(choice => (
          <div key={choice.value} className="flex items-center gap-2 group">
            <div className="flex gap-1">
              {COLORS.map(c => (
                <button
                  key={c}
                  type="button"
                  onClick={() => updateColor(choice.value, c)}
                  className={`w-4 h-4 rounded-full transition-transform ${choice.color === c ? 'scale-125 ring-2 ring-offset-1 ring-gray-400' : ''}`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
            <span
              className="flex-1 px-2 py-0.5 rounded-full text-xs font-medium"
              style={{ backgroundColor: (choice.color ?? '#6b7280') + '22', color: choice.color ?? '#6b7280' }}
            >
              {choice.label}
            </span>
            <button
              type="button"
              onClick={() => removeChoice(choice.value)}
              className="text-gray-300 hover:text-red-500 opacity-0 group-hover:opacity-100"
            >
              ✕
            </button>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-2">
        <div className="flex gap-1">
          {COLORS.map(c => (
            <button
              key={c}
              type="button"
              onClick={() => setSelectedColor(c)}
              className={`w-4 h-4 rounded-full ${selectedColor === c ? 'ring-2 ring-offset-1 ring-gray-400 scale-125' : ''}`}
              style={{ backgroundColor: c }}
            />
          ))}
        </div>
        <input
          value={newLabel}
          onChange={(e) => setNewLabel(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addChoice(); } }}
          placeholder="Новый вариант..."
          className="flex-1 border border-gray-300 rounded px-2 py-1 text-sm outline-none focus:ring-1 focus:ring-blue-500"
        />
        <button
          type="button"
          onClick={addChoice}
          className="px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm font-medium"
        >
          +
        </button>
      </div>
    </div>
  );
};
```

---

## `src/components/Permissions/ShareModal.tsx`

```typescript
import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../../api/permissions';
import type { SpreadsheetAccess } from '../../types';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { FieldAccessMatrix } from './FieldAccessMatrix';
import { RowRuleBuilder } from './RowRuleBuilder';
import type { Field } from '../../types';

type Tab = 'users' | 'fields' | 'rows';

interface Props {
  spreadsheetId: string;
  fields: Field[];
  onClose: () => void;
}

export const ShareModal: React.FC<Props> = ({ spreadsheetId, fields, onClose }) => {
  const [tab, setTab] = useState<Tab>('users');
  const [newEmail, setNewEmail] = useState('');
  const qc = useQueryClient();

  const { data: accesses = [], isLoading } = useQuery({
    queryKey: ['permissions', spreadsheetId],
    queryFn: () => permissionsApi.getSpreadsheetAccess(spreadsheetId),
  });

  const upsert = useMutation({
    mutationFn: (data: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
      permissionsApi.upsertSpreadsheetAccess(spreadsheetId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (principalId: string) =>
      permissionsApi.removeSpreadsheetAccess(spreadsheetId, principalId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const handleInvite = () => {
    if (!newEmail.trim()) return;
    // В реальности — поиск пользователя по email, затем upsert
    setNewEmail('');
  };

  const togglePerm = (access: SpreadsheetAccess, perm: keyof SpreadsheetAccess) => {
    upsert.mutate({
      principal_id: access.principal_id,
      principal_type: access.principal_type,
      can_view:   access.can_view,
      can_insert: access.can_insert,
      can_edit:   access.can_edit,
      can_delete: access.can_delete,
      can_manage: access.can_manage,
      [perm]: !access[perm],
    });
  };

  const TABS: { id: Tab; label: string }[] = [
    { id: 'users',  label: '👥 Пользователи' },
    { id: 'fields', label: '🔒 Поля' },
    { id: 'rows',   label: '📋 Строки' },
  ];

  const PERMS: { key: keyof SpreadsheetAccess; label: string }[] = [
    { key: 'can_view',   label: 'Просмотр' },
    { key: 'can_insert', label: 'Добавление' },
    { key: 'can_edit',   label: 'Редактирование' },
    { key: 'can_delete', label: 'Удаление' },
    { key: 'can_manage', label: 'Управление' },
  ];

  return (
    <Modal open onClose={onClose} title="Настройка доступа" size="xl">
      {/* Tabs */}
      <div className="flex border-b mb-4 -mx-6 px-6 gap-1">
        {TABS.map(t => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium rounded-t transition-colors
              ${tab === t.id ? 'border-b-2 border-blue-600 text-blue-700' : 'text-gray-600 hover:text-gray-900'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Users tab */}
      {tab === 'users' && (
        <div className="space-y-4">
          <div className="flex gap-2">
            <input
              value={newEmail}
              onChange={(e) => setNewEmail(e.target.value)}
              placeholder="Email пользователя..."
              className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
            />
            <Button onClick={handleInvite} size="sm">Пригласить</Button>
          </div>

          {isLoading ? (
            <p className="text-sm text-gray-400 text-center py-4">Загрузка...</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 pr-4 font-medium text-gray-600">Пользователь</th>
                    {PERMS.map(p => (
                      <th key={p.key} className="text-center py-2 px-2 font-medium text-gray-600 text-xs">
                        {p.label}
                      </th>
                    ))}
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {accesses.map((access) => (
                    <tr key={access.principal_id} className="border-b hover:bg-gray-50">
                      <td className="py-2 pr-4">
                        <span className="font-medium text-gray-800">{access.principal_name}</span>
                        <span className="ml-2 text-xs text-gray-400">{access.principal_type}</span>
                      </td>
                      {PERMS.map(p => (
                        <td key={p.key} className="text-center py-2 px-2">
                          <input
                            type="checkbox"
                            checked={Boolean(access[p.key])}
                            onChange={() => togglePerm(access, p.key)}
                            className="w-4 h-4 cursor-pointer"
                          />
                        </td>
                      ))}
                      <td className="py-2 pl-2">
                        <button
                          onClick={() => remove.mutate(access.principal_id)}
                          className="text-red-400 hover:text-red-600 text-xs"
                        >
                          Удалить
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {tab === 'fields' && (
        <FieldAccessMatrix spreadsheetId={spreadsheetId} fields={fields} accesses={accesses} />
      )}

      {tab === 'rows' && (
        <RowRuleBuilder spreadsheetId={spreadsheetId} fields={fields} />
      )}
    </Modal>
  );
};
```

---

## `src/components/Permissions/FieldAccessMatrix.tsx`

```typescript
import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../../api/permissions';
import type { Field, SpreadsheetAccess } from '../../types';

interface Props {
  spreadsheetId: string;
  fields: Field[];
  accesses: SpreadsheetAccess[];
}

export const FieldAccessMatrix: React.FC<Props> = ({ spreadsheetId, fields, accesses }) => {
  const qc = useQueryClient();

  const { data: fieldAccesses = [] } = useQuery({
    queryKey: ['field-permissions', spreadsheetId],
    queryFn: () => permissionsApi.getFieldAccess(spreadsheetId),
  });

  const upsert = useMutation({
    mutationFn: ({ fieldId, data }: { fieldId: string; data: { principal_id: string; can_view: boolean; can_edit: boolean } }) =>
      permissionsApi.upsertFieldAccess(fieldId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['field-permissions', spreadsheetId] }),
  });

  const getAccess = (fieldId: string, principalId: string) =>
    fieldAccesses.find(fa => fa.field_id === fieldId && fa.principal_id === principalId);

  const toggle = (fieldId: string, principalId: string, perm: 'can_view' | 'can_edit') => {
    const current = getAccess(fieldId, principalId);
    upsert.mutate({
      fieldId,
      data: {
        principal_id: principalId,
        can_view:  perm === 'can_view' ? !(current?.can_view ?? true) : (current?.can_view ?? true),
        can_edit:  perm === 'can_edit' ? !(current?.can_edit ?? true) : (current?.can_edit ?? true),
      },
    });
  };

  return (
    <div className="overflow-auto max-h-96">
      <table className="w-full text-xs border-collapse">
        <thead className="sticky top-0 bg-white z-10">
          <tr>
            <th className="text-left py-2 pr-4 font-medium text-gray-600 min-w-32">Поле / Пользователь</th>
            {accesses.map(a => (
              <th key={a.principal_id} className="text-center px-2 py-2 font-medium text-gray-600">
                <div>{a.principal_name}</div>
                <div className="flex gap-1 justify-center mt-1">
                  <span className="text-gray-400">👁</span>
                  <span className="text-gray-400">✏️</span>
                </div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {fields.map(field => (
            <tr key={field.id} className="border-t hover:bg-gray-50">
              <td className="py-2 pr-4 font-medium text-gray-700">{field.name}</td>
              {accesses.map(access => {
                const fa = getAccess(field.id, access.principal_id);
                const canView = fa?.can_view ?? true;
                const canEdit = fa?.can_edit ?? true;
                return (
                  <td key={access.principal_id} className="text-center px-2 py-2">
                    <div className="flex gap-1 justify-center">
                      <input
                        type="checkbox"
                        checked={canView}
                        onChange={() => toggle(field.id, access.principal_id, 'can_view')}
                        className="w-3.5 h-3.5 cursor-pointer"
                        title="Просмотр"
                      />
                      <input
                        type="checkbox"
                        checked={canEdit}
                        disabled={!canView}
                        onChange={() => toggle(field.id, access.principal_id, 'can_edit')}
                        className="w-3.5 h-3.5 cursor-pointer disabled:opacity-40"
                        title="Редактирование"
                      />
                    </div>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
```

---

## `src/components/Permissions/RowRuleBuilder.tsx`

```typescript
import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../../api/permissions';
import type { Field, RowAccessRule } from '../../types';
import { Button } from '../ui/Button';

const OPS = [
  { value: 'eq',             label: 'равно' },
  { value: 'neq',            label: 'не равно' },
  { value: 'contains',       label: 'содержит' },
  { value: 'eq_current_user', label: '= текущий пользователь' },
];

interface Props {
  spreadsheetId: string;
  fields: Field[];
}

export const RowRuleBuilder: React.FC<Props> = ({ spreadsheetId, fields }) => {
  const qc = useQueryClient();

  const { data: rules = [] } = useQuery({
    queryKey: ['row-rules', spreadsheetId],
    queryFn: () => permissionsApi.getRowRules(spreadsheetId),
  });

  const deleteRule = useMutation({
    mutationFn: (ruleId: string) => permissionsApi.deleteRowRule(spreadsheetId, ruleId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['row-rules', spreadsheetId] }),
  });

  const addRule = useMutation({
    mutationFn: (data: Omit<RowAccessRule, 'id'>) => permissionsApi.upsertRowRule(spreadsheetId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['row-rules', spreadsheetId] }),
  });

  const getFieldName = (colName: string) =>
    fields.find(f => f.column_name === colName)?.name ?? colName;

  return (
    <div className="space-y-4">
      <p className="text-sm text-gray-500">
        Правила ограничивают, какие строки видит или может редактировать пользователь.
      </p>

      <div className="space-y-2">
        {rules.map(rule => (
          <div key={rule.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg text-sm">
            <span className="font-medium text-gray-700">{rule.principal_id}</span>
            <span className="text-gray-400">→</span>
            <span className="text-gray-600">
              если <strong>{getFieldName(rule.condition.column_name)}</strong>{' '}
              {OPS.find(o => o.value === rule.condition.op)?.label}{' '}
              {rule.condition.value && <strong>{rule.condition.value}</strong>}
            </span>
            <div className="ml-auto flex items-center gap-2">
              {rule.can_view && <span className="px-2 py-0.5 bg-green-100 text-green-700 rounded text-xs">просмотр</span>}
              {rule.can_edit && <span className="px-2 py-0.5 bg-blue-100 text-blue-700 rounded text-xs">редактирование</span>}
              <button
                onClick={() => deleteRule.mutate(rule.id)}
                className="text-red-400 hover:text-red-600 ml-2"
              >
                ✕
              </button>
            </div>
          </div>
        ))}

        {rules.length === 0 && (
          <p className="text-center text-sm text-gray-400 py-6">Правила не настроены</p>
        )}
      </div>

      <Button
        variant="secondary"
        size="sm"
        onClick={() => {
          // Упрощённый пример добавления правила
          addRule.mutate({
            spreadsheet_id: spreadsheetId,
            principal_id: '',
            principal_type: 'user',
            condition: { column_name: fields[0]?.column_name ?? '', op: 'eq_current_user' },
            can_view: true,
            can_edit: false,
          });
        }}
      >
        + Добавить правило
      </Button>
    </div>
  );
};
```

---

## `src/pages/SpreadsheetPage.tsx`

```typescript
import React, { useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { fieldsApi } from '../api/fields';
import { useRows } from '../hooks/useRows';
import { useSpreadsheetStore } from '../store/spreadsheetStore';
import { Grid } from '../components/Grid/Grid';
import { FieldEditorModal } from '../components/FieldEditor/FieldEditorModal';
import { ShareModal } from '../components/Permissions/ShareModal';
import { Button } from '../components/ui/Button';
import type { Field } from '../types';

export const SpreadsheetPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const {
    spreadsheet, setSpreadsheet,
    addField, updateField, removeField,
    selectedRowIds, toggleRowSelection, clearSelection,
    filters, sorts, setFilters, setSorts,
  } = useSpreadsheetStore();

  const [addFieldOpen, setAddFieldOpen] = useState(false);
  const [shareOpen, setShareOpen] = useState(false);

  const { isLoading } = useQuery({
    queryKey: ['spreadsheet', id],
    queryFn: () => spreadsheetsApi.get(id!),
    onSuccess: setSpreadsheet,
    enabled: Boolean(id),
  });

  const { allRows, fetchNextPage, hasNextPage, isFetchingNextPage, createRow, updateRow, deleteRow } = useRows(id!);

  const handleUpdateCell = useCallback(async (rowId: string, colName: string, value: unknown) => {
    await updateRow.mutateAsync({ rowId, data: { [colName]: value } });
  }, [updateRow]);

  const handleAddRow = useCallback(async () => {
    await createRow.mutateAsync({});
  }, [createRow]);

  const handleDeleteRow = useCallback(async (rowId: string) => {
    await deleteRow.mutateAsync(rowId);
  }, [deleteRow]);

  const handleDeleteField = useCallback(async (fieldId: string) => {
    if (!window.confirm('Удалить поле и все данные в нём?')) return;
    await fieldsApi.delete(fieldId);
    removeField(fieldId);
  }, [removeField]);

  const handleSort = useCallback((fieldId: string, direction: 'asc' | 'desc') => {
    setSorts([{ field_id: fieldId, direction }]);
  }, [setSorts]);

  if (isLoading || !spreadsheet) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full" />
      </div>
    );
  }

  const sortedFields = [...spreadsheet.fields].sort((a, b) => a.position - b.position);

  return (
    <div className="flex flex-col h-screen bg-white">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white shrink-0">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-gray-900">{spreadsheet.name}</h1>
          <span className="text-sm text-gray-400">{allRows.length} строк</span>
        </div>

        <div className="flex items-center gap-2">
          {selectedRowIds.size > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-blue-700 bg-blue-50 px-2 py-1 rounded">
                Выбрано: {selectedRowIds.size}
              </span>
              <Button
                variant="danger"
                size="sm"
                onClick={async () => {
                  for (const rowId of selectedRowIds) await deleteRow.mutateAsync(rowId);
                  clearSelection();
                }}
              >
                Удалить выбранные
              </Button>
              <Button variant="ghost" size="sm" onClick={clearSelection}>Отмена</Button>
            </div>
          )}
          <Button variant="secondary" size="sm" onClick={() => setShareOpen(true)}>
            🔒 Доступ
          </Button>
          <Button size="sm" onClick={handleAddRow} loading={createRow.isPending}>
            + Строка
          </Button>
        </div>
      </div>

      {/* Grid */}
      <div className="flex-1 overflow-hidden p-4">
        <Grid
          fields={sortedFields}
          rows={allRows}
          canEdit={true}
          canManage={true}
          editableFields={new Set(sortedFields.map(f => f.id))}
          selectedRowIds={selectedRowIds}
          hasNextPage={hasNextPage}
          isFetchingNextPage={isFetchingNextPage}
          onFetchNextPage={fetchNextPage}
          onUpdateCell={handleUpdateCell}
          onAddRow={handleAddRow}
          onDeleteRow={handleDeleteRow}
          onToggleRow={toggleRowSelection}
          onUpdateField={(field: Field) => { updateField(field); }}
          onDeleteField={handleDeleteField}
          onAddField={() => setAddFieldOpen(true)}
          onSort={handleSort}
        />
      </div>

      {/* Modals */}
      {addFieldOpen && (
        <FieldEditorModal
          spreadsheetId={spreadsheet.id}
          onClose={() => setAddFieldOpen(false)}
          onSave={(field) => { addField(field); setAddFieldOpen(false); }}
        />
      )}
      {shareOpen && (
        <ShareModal
          spreadsheetId={spreadsheet.id}
          fields={sortedFields}
          onClose={() => setShareOpen(false)}
        />
      )}
    </div>
  );
};
```

---

## `src/pages/LoginPage.tsx`

```typescript
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/auth';
import { useAuthStore } from '../store/authStore';
import { Button } from '../components/ui/Button';

export const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const { setAuth } = useAuthStore();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [form, setForm] = useState({ email: '', password: '', name: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const result = mode === 'login'
        ? await authApi.login({ email: form.email, password: form.password })
        : await authApi.register(form);
      setAuth(result.user, result.access_token);
      navigate('/workspaces');
    } catch {
      setError('Неверный email или пароль');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="w-full max-w-sm bg-white rounded-2xl shadow-lg p-8">
        <div className="text-center mb-8">
          <div className="text-4xl mb-2">📊</div>
          <h1 className="text-2xl font-bold text-gray-900">DataGrid</h1>
          <p className="text-gray-500 text-sm mt-1">Умные таблицы для вашей команды</p>
        </div>

        <div className="flex rounded-lg bg-gray-100 p-1 mb-6">
          {(['login', 'register'] as const).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={`flex-1 py-2 text-sm font-medium rounded-md transition-colors
                ${mode === m ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600'}`}
            >
              {m === 'login' ? 'Войти' : 'Регистрация'}
            </button>
          ))}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {mode === 'register' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Имя</label>
              <input
                value={form.name}
                onChange={(e) => setForm(f => ({ ...f, name: e.target.value }))}
                placeholder="Иван Иванов"
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              type="email"
              value={form.email}
              onChange={(e) => setForm(f => ({ ...f, email: e.target.value }))}
              placeholder="you@example.com"
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm(f => ({ ...f, password: e.target.value }))}
              placeholder="••••••••"
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {error && <p className="text-sm text-red-600 text-center">{error}</p>}

          <Button type="submit" className="w-full justify-center" loading={loading}>
            {mode === 'login' ? 'Войти' : 'Создать аккаунт'}
          </Button>
        </form>
      </div>
    </div>
  );
};
```

---

## `src/pages/WorkspacesPage.tsx`

```typescript
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { useAuthStore } from '../store/authStore';
import { Button } from '../components/ui/Button';
import { Modal } from '../components/ui/Modal';
import type { Workspace } from '../types';

// Временная заглушка: первый workspace из store
const MOCK_WORKSPACE_ID = 'ws-1';

export const WorkspacesPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [newName, setNewName] = useState('');

  const { data: spreadsheets = [], isLoading } = useQuery({
    queryKey: ['spreadsheets', MOCK_WORKSPACE_ID],
    queryFn: () => spreadsheetsApi.list(MOCK_WORKSPACE_ID),
  });

  const create = useMutation({
    mutationFn: () => spreadsheetsApi.create({ name: newName, workspace_id: MOCK_WORKSPACE_ID }),
    onSuccess: (s) => {
      qc.invalidateQueries({ queryKey: ['spreadsheets'] });
      setCreateOpen(false);
      setNewName('');
      navigate(`/spreadsheet/${s.id}`);
    },
  });

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xl">📊</span>
          <span className="font-semibold text-gray-900">DataGrid</span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-600">{user?.name}</span>
          <Button variant="ghost" size="sm" onClick={logout}>Выйти</Button>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-gray-900">Мои таблицы</h2>
          <Button onClick={() => setCreateOpen(true)}>+ Новая таблица</Button>
        </div>

        {isLoading ? (
          <div className="grid grid-cols-3 gap-4">
            {[1,2,3].map(i => (
              <div key={i} className="h-32 bg-gray-200 animate-pulse rounded-xl" />
            ))}
          </div>
        ) : spreadsheets.length === 0 ? (
          <div className="text-center py-16">
            <div className="text-5xl mb-4">📋</div>
            <p className="text-gray-500">Таблиц ещё нет. Создайте первую!</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {spreadsheets.map((s) => (
              <button
                key={s.id}
                onClick={() => navigate(`/spreadsheet/${s.id}`)}
                className="text-left p-5 bg-white rounded-xl border border-gray-200 hover:border-blue-400 hover:shadow-md transition-all group"
              >
                <div className="text-2xl mb-3">📊</div>
                <div className="font-medium text-gray-900 group-hover:text-blue-700">{s.name}</div>
                {s.description && (
                  <div className="text-sm text-gray-500 mt-1 truncate">{s.description}</div>
                )}
                <div className="text-xs text-gray-400 mt-3">
                  {s.fields.length} полей · {new Date(s.updated_at).toLocaleDateString('ru-RU')}
                </div>
              </button>
            ))}
          </div>
        )}
      </main>

      <Modal open={createOpen} onClose={() => setCreateOpen(false)} title="Новая таблица" size="sm">
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Название</label>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') create.mutate(); }}
              placeholder="Например: Задачи, Клиенты..."
              autoFocus
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setCreateOpen(false)}>Отмена</Button>
            <Button onClick={() => create.mutate()} loading={create.isPending} disabled={!newName.trim()}>
              Создать
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};
```

---

## `src/App.tsx`

```typescript
import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuthStore } from './store/authStore';
import { LoginPage } from './pages/LoginPage';
import { WorkspacesPage } from './pages/WorkspacesPage';
import { SpreadsheetPage } from './pages/SpreadsheetPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  return isAuthenticated() ? <>{children}</> : <Navigate to="/login" replace />;
};

export const App: React.FC = () => (
  <QueryClientProvider client={queryClient}>
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/workspaces" element={<PrivateRoute><WorkspacesPage /></PrivateRoute>} />
        <Route path="/spreadsheet/:id" element={<PrivateRoute><SpreadsheetPage /></PrivateRoute>} />
        <Route path="*" element={<Navigate to="/workspaces" replace />} />
      </Routes>
    </BrowserRouter>
  </QueryClientProvider>
);
```

---

## `src/main.tsx`

```typescript
import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './App';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

---

## `package.json`

```json
{
  "name": "datagrid-frontend",
  "version": "0.1.0",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.0.0",
    "@tanstack/react-virtual": "^3.0.0",
    "axios": "^1.6.0",
    "clsx": "^2.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "zustand": "^4.4.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

---

## `tailwind.config.js`

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: { extend: {} },
  plugins: [],
};
```

**User - --**

не хватает файла index.css, добавь

**Assistant - --**

## `src/index.css`

```css
/* ─── Tailwind directives ─────────────────────────────────── */
@tailwind base;
@tailwind components;
@tailwind utilities;

/* ─── CSS Custom Properties ───────────────────────────────── */
:root {
  --color-primary:        #2563eb;
  --color-primary-hover:  #1d4ed8;
  --color-danger:         #dc2626;
  --color-border:         #e5e7eb;
  --color-bg:             #f9fafb;
  --color-surface:        #ffffff;
  --color-text:           #111827;
  --color-text-muted:     #6b7280;

  --radius-sm:  4px;
  --radius-md:  8px;
  --radius-lg:  12px;
  --radius-xl:  16px;

  --shadow-sm:  0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md:  0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  --shadow-lg:  0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --shadow-xl:  0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);

  --font-sans:  'Inter', ui-sans-serif, system-ui, -apple-system, sans-serif;
  --font-mono:  'JetBrains Mono', ui-monospace, 'Cascadia Code', monospace;
}

/* ─── Base Reset ──────────────────────────────────────────── */
@layer base {
  *,
  *::before,
  *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  html {
    font-size: 16px;
    -webkit-text-size-adjust: 100%;
    tab-size: 4;
  }

  body {
    font-family: var(--font-sans);
    font-size: 0.875rem;    /* 14px */
    line-height: 1.5;
    color: var(--color-text);
    background-color: var(--color-bg);
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  /* Убираем стандартные стили кнопок */
  button {
    font-family: inherit;
    cursor: pointer;
    border: none;
    background: none;
  }

  button:disabled {
    cursor: not-allowed;
  }

  /* Убираем стандартные стили инпутов */
  input,
  select,
  textarea {
    font-family: inherit;
    font-size: inherit;
  }

  /* Скроллбар — WebKit */
  ::-webkit-scrollbar {
    width: 8px;
    height: 8px;
  }

  ::-webkit-scrollbar-track {
    background: transparent;
  }

  ::-webkit-scrollbar-thumb {
    background-color: #d1d5db;
    border-radius: 4px;
    border: 2px solid transparent;
    background-clip: content-box;
  }

  ::-webkit-scrollbar-thumb:hover {
    background-color: #9ca3af;
  }

  ::-webkit-scrollbar-corner {
    background: transparent;
  }

  /* Скроллбар — Firefox */
  * {
    scrollbar-width: thin;
    scrollbar-color: #d1d5db transparent;
  }

  /* Выделение текста */
  ::selection {
    background-color: #bfdbfe;
    color: #1e3a5f;
  }

  /* Фокус — доступность */
  :focus-visible {
    outline: 2px solid var(--color-primary);
    outline-offset: 2px;
    border-radius: var(--radius-sm);
  }

  /* Убираем outline при клике мышью */
  :focus:not(:focus-visible) {
    outline: none;
  }

  /* Ссылки */
  a {
    color: var(--color-primary);
    text-decoration: none;
  }

  a:hover {
    text-decoration: underline;
  }

  /* Изображения */
  img,
  svg {
    display: block;
    max-width: 100%;
  }

  /* Таблицы */
  table {
    border-collapse: collapse;
    border-spacing: 0;
  }

  /* Моноширинный текст */
  code,
  kbd,
  pre,
  samp {
    font-family: var(--font-mono);
    font-size: 0.875em;
  }
}

/* ─── Компонентные утилиты ────────────────────────────────── */
@layer components {

  /* Сетка таблицы — фиксируем высоту строк */
  .grid-row {
    @apply flex border-b border-gray-100;
    height: 36px;
    min-height: 36px;
  }

  .grid-cell {
    @apply flex items-center px-2 border-r border-gray-200 overflow-hidden;
    min-height: 36px;
  }

  .grid-header-cell {
    @apply flex items-center px-2 border-r border-gray-200 bg-gray-50 select-none;
    height: 36px;
  }

  /* Значок типа поля */
  .field-type-icon {
    @apply text-xs font-mono text-gray-400 shrink-0 w-5 text-center;
  }

  /* Бейдж для select-значений */
  .select-badge {
    @apply inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium whitespace-nowrap;
  }

  /* Карточка таблицы на странице Workspaces */
  .spreadsheet-card {
    @apply bg-white rounded-xl border border-gray-200 p-5 text-left
           transition-all duration-150 cursor-pointer
           hover:border-blue-400 hover:shadow-md;
  }

  /* Строка таблицы прав */
  .permission-row {
    @apply flex items-center gap-3 px-3 py-2 rounded-lg text-sm;
  }

  /* Разделитель группы в FieldTypeSelect */
  .field-type-group-label {
    @apply px-3 py-1 bg-gray-50 text-xs font-medium text-gray-400 uppercase tracking-wider;
  }

  /* Анимация скелетона */
  .skeleton {
    @apply animate-pulse bg-gray-200 rounded;
  }

  /* Overlay для модальных окон */
  .modal-overlay {
    @apply fixed inset-0 z-50 flex items-center justify-center;
  }

  .modal-backdrop {
    @apply absolute inset-0 bg-black/50 backdrop-blur-sm;
  }

  .modal-content {
    @apply relative z-10 w-full mx-4 bg-white rounded-xl shadow-2xl;
  }
}

/* ─── Утилиты ─────────────────────────────────────────────── */
@layer utilities {

  /* Плавное появление */
  .animate-fade-in {
    animation: fade-in 0.15s ease-out;
  }

  .animate-slide-down {
    animation: slide-down 0.2s ease-out;
  }

  .animate-slide-up {
    animation: slide-up 0.2s ease-out;
  }

  /* Обрезка текста в ячейке */
  .cell-text {
    @apply truncate text-sm text-gray-800;
  }

  /* Прокрутка только по X (для Grid) */
  .scroll-x-only {
    overflow-x: auto;
    overflow-y: hidden;
  }

  /* Sticky колонка-индекс */
  .col-sticky {
    position: sticky;
    left: 0;
    z-index: 1;
    background: inherit;
  }

  /* Подсветка активной ячейки */
  .cell-active {
    @apply ring-2 ring-inset ring-blue-500;
  }

  /* Подсветка выбранной строки */
  .row-selected {
    @apply bg-blue-50;
  }

  /* Тонкая граница */
  .border-light {
    border-color: var(--color-border);
  }
}

/* ─── Keyframes ───────────────────────────────────────────── */
@keyframes fade-in {
  from { opacity: 0; }
  to   { opacity: 1; }
}

@keyframes slide-down {
  from { opacity: 0; transform: translateY(-8px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes slide-up {
  from { opacity: 0; transform: translateY(8px); }
  to   { opacity: 1; transform: translateY(0); }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ─── Специфика Grid ──────────────────────────────────────── */

/* Контейнер виртуализированной сетки */
.grid-container {
  position: relative;
  will-change: transform;
}

/* Запрещаем перенос в заголовке */
.grid-header {
  white-space: nowrap;
}

/* Отключаем pointer-events при drag-resize колонки */
.grid-resizing * {
  pointer-events: none;
  user-select: none;
}

/* Ручка resize колонки */
.col-resize-handle {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  cursor: col-resize;
  background: transparent;
  transition: background 0.15s;
}

.col-resize-handle:hover,
.col-resize-handle:active {
  background: var(--color-primary);
}

/* ─── Специфика форм ──────────────────────────────────────── */

.form-input {
  @apply w-full border border-gray-300 rounded-lg px-3 py-2 text-sm
         outline-none transition-shadow
         focus:ring-2 focus:ring-blue-500 focus:border-blue-500;
}

.form-label {
  @apply block text-sm font-medium text-gray-700 mb-1;
}

.form-error {
  @apply text-sm text-red-600 mt-1;
}

.form-hint {
  @apply text-xs text-gray-400 mt-1;
}

/* ─── Специфика Dropdown ──────────────────────────────────── */

.dropdown-menu {
  @apply absolute z-50 min-w-40 bg-white border border-gray-200
         rounded-xl shadow-lg py-1 animate-slide-down;
}

.dropdown-item {
  @apply flex items-center gap-2 px-3 py-2 text-sm text-gray-700
         cursor-pointer transition-colors hover:bg-gray-50;
}

.dropdown-item.danger {
  @apply text-red-600 hover:bg-red-50;
}

.dropdown-separator {
  @apply my-1 border-t border-gray-100;
}

/* ─── Специфика Toast / уведомлений ──────────────────────── */

.toast {
  @apply flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg
         text-sm font-medium animate-slide-up;
}

.toast-success {
  @apply bg-green-50 text-green-800 border border-green-200;
}

.toast-error {
  @apply bg-red-50 text-red-800 border border-red-200;
}

.toast-info {
  @apply bg-blue-50 text-blue-800 border border-blue-200;
}

/* ─── Адаптивность ────────────────────────────────────────── */

/* На мобильных — скрываем лишние колонки Grid */
@media (max-width: 640px) {
  .grid-header-cell,
  .grid-cell {
    min-width: 120px;
  }

  .modal-content {
    @apply mx-2 rounded-lg;
  }
}

/* ─── Печать ──────────────────────────────────────────────── */

@media print {
  .modal-overlay,
  .dropdown-menu,
  button {
    display: none !important;
  }

  .grid-container {
    overflow: visible !important;
  }
}
```

**User - --**

при запуске приложения выдает такую оштбку:
@tailwind base;
^

SyntaxError: Invalid or unexpected token
    at wrapSafe (node:internal/modules/cjs/loader:1743:18)
    at Module._compile (node:internal/modules/cjs/loader:1786:20)
    at node:internal/modules/cjs/loader:1943:10
    at Object.transformer (D:\Git_personal\NEW MODMAS\frontend\src\node_modules\tsx\dist\register-D46fvsV_.cjs:3:844)
    at Module.load (node:internal/modules/cjs/loader:1533:32)
    at Module._load (node:internal/modules/cjs/loader:1335:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.require (node:internal/modules/cjs/loader:1556:12)
    at require (node:internal/modules/helpers:152:16)
    at <anonymous> (D:\Git_personal\NEW MODMAS\frontend\src\main.tsx:4:8)

Node.js v24.14.1

**Assistant - --**

Ошибка возникает потому что `tsx` пытается выполнить CSS-файл как JavaScript. Это значит приложение запускается не через Vite, а напрямую через `tsx`. 

Проверьте несколько вещей:

---

## 1. Убедитесь что есть `vite.config.ts`

```typescript
// vite.config.ts  (в корне frontend/)
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
});
```

---

## 2. Проверьте `package.json` — скрипт `dev` должен запускать Vite

```json
{
  "name": "datagrid-frontend",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev":     "vite",
    "build":   "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.0.0",
    "@tanstack/react-virtual": "^3.0.0",
    "axios": "^1.6.0",
    "clsx": "^2.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "zustand": "^4.4.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

---

## 3. Убедитесь что есть `index.html` в корне `frontend/`

```html
<!DOCTYPE html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>DataGrid</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

---

## 4. Добавьте конфиги PostCSS и Tailwind

```javascript
// postcss.config.js  (в корне frontend/)
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
```

```javascript
// tailwind.config.js  (в корне frontend/)
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
```

---

## 5. Итоговая структура файлов в корне `frontend/`

```
frontend/
├── index.html          ← обязательно
├── vite.config.ts      ← обязательно
├── postcss.config.js   ← обязательно для Tailwind
├── tailwind.config.js  ← обязательно для Tailwind
├── tsconfig.json
├── package.json
└── src/
    ├── main.tsx
    ├── index.css
    └── ...
```

---

## 6. `tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"]
}
```

---

## 7. Установка и запуск

```bash
# Перейти в папку frontend
cd frontend

# Установить зависимости
npm install

# Запустить через Vite (НЕ через tsx)
npm run dev
```

---

Корень проблемы — запуск `tsx src/main.tsx` напрямую вместо `vite`. Node.js не понимает CSS с директивами `@tailwind`. Vite обрабатывает CSS через PostCSS перед передачей в браузер.

**User - --**

где должны распологаться файлы vite.config.ts и package.json

**Assistant - --**

Все конфигурационные файлы должны лежать в **корне папки `frontend/`**, рядом с `index.html`.

```
frontend/                   ← корень проекта фронтенда
├── index.html
├── package.json            ← здесь
├── vite.config.ts          ← здесь
├── postcss.config.js       ← здесь
├── tailwind.config.js      ← здесь
├── tsconfig.json           ← здесь
└── src/
    ├── main.tsx
    ├── index.css
    ├── App.tsx
    ├── types/
    ├── api/
    ├── store/
    ├── hooks/
    ├── components/
    └── pages/
```

Команды запускаются тоже из папки `frontend/`:

```bash
cd frontend
npm install
npm run dev
```

**User - --**

сгенерируй код для бэкенда

**Assistant - --**

# Бэкенд — полная реализация

## Стек
- **Go 1.22**
- **pgx v5** — драйвер PostgreSQL
- **Chi** — HTTP роутер
- **golang-jwt** — JWT токены
- **golang-migrate** — миграции БД

---

## Структура файлов

```
backend/
├── cmd/server/main.go
├── config/config.go
├── migrations/
│   ├── 001_auth_schema.sql
│   ├── 002_meta_schema.sql
│   └── 003_permissions_schema.sql
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── workspace/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── spreadsheet/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── field/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── types.go
│   ├── row/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── repository.go
│   ├── permission/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── enforcer.go
│   ├── ddl/
│   │   ├── builder.go
│   │   └── executor.go
│   └── middleware/
│       ├── auth.go
│       └── cors.go
└── pkg/
    ├── postgres/pool.go
    ├── jwt/jwt.go
    ├── hasher/hasher.go
    └── response/response.go
```

---

## `go.mod`

```go
module github.com/yourorg/datagrid

go 1.22

require (
    github.com/go-chi/chi/v5       v5.0.12
    github.com/go-chi/cors         v1.2.1
    github.com/golang-jwt/jwt/v5   v5.2.1
    github.com/golang-migrate/migrate/v4 v4.17.1
    github.com/google/uuid         v1.6.0
    github.com/jackc/pgx/v5        v5.5.5
    golang.org/x/crypto            v0.22.0
)
```

---

## `config/config.go`

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    DB       DBConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Host string
    Port string
}

type DBConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type JWTConfig struct {
    Secret          string
    AccessTokenTTL  time.Duration
    RefreshTokenTTL time.Duration
}

func (d DBConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
    )
}

func Load() *Config {
    accessTTL, _ := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MINUTES", "60"))
    refreshTTL, _ := strconv.Atoi(getEnv("JWT_REFRESH_TTL_DAYS", "30"))

    return &Config{
        Server: ServerConfig{
            Host: getEnv("SERVER_HOST", "0.0.0.0"),
            Port: getEnv("SERVER_PORT", "8080"),
        },
        DB: DBConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", "postgres"),
            DBName:   getEnv("DB_NAME", "datagrid"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
        JWT: JWTConfig{
            Secret:          getEnv("JWT_SECRET", "change-me-in-production"),
            AccessTokenTTL:  time.Duration(accessTTL) * time.Minute,
            RefreshTokenTTL: time.Duration(refreshTTL) * 24 * time.Hour,
        },
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

## `pkg/postgres/pool.go`

```go
package postgres

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("parse pool config: %w", err)
    }

    cfg.MaxConns = 20
    cfg.MinConns = 2

    pool, err := pgxpool.NewWithConfig(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }

    return pool, nil
}
```

---

## `pkg/jwt/jwt.go`

```go
package jwt

import (
    "errors"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type Manager struct {
    secret    string
    accessTTL time.Duration
}

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func NewManager(secret string, accessTTL time.Duration) *Manager {
    return &Manager{secret: secret, accessTTL: accessTTL}
}

func (m *Manager) Generate(userID, email string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ID:        uuid.New().String(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString([]byte(m.secret))
    if err != nil {
        return "", fmt.Errorf("sign token: %w", err)
    }
    return signed, nil
}

func (m *Manager) Validate(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return []byte(m.secret), nil
    })
    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    return claims, nil
}
```

---

## `pkg/hasher/hasher.go`

```go
package hasher

import (
    "fmt"

    "golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("hash password: %w", err)
    }
    return string(hash), nil
}

func Check(password, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

---

## `pkg/response/response.go`

```go
package response

import (
    "encoding/json"
    "net/http"
)

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, data any) {
    JSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data any) {
    JSON(w, http.StatusCreated, data)
}

func NoContent(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}

func BadRequest(w http.ResponseWriter, msg string) {
    JSON(w, http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: msg})
}

func Unauthorized(w http.ResponseWriter, msg string) {
    JSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Message: msg})
}

func Forbidden(w http.ResponseWriter, msg string) {
    JSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Message: msg})
}

func NotFound(w http.ResponseWriter, msg string) {
    JSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Message: msg})
}

func InternalError(w http.ResponseWriter, msg string) {
    JSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: msg})
}

func Conflict(w http.ResponseWriter, msg string) {
    JSON(w, http.StatusConflict, ErrorResponse{Error: "conflict", Message: msg})
}
```

---

## Миграции

### `migrations/001_auth_schema.sql`

```sql
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS meta;

CREATE TABLE auth.users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    avatar_url    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth.workspaces (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    owner_id   UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    db_schema  TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth.workspace_members (
    workspace_id UUID NOT NULL REFERENCES auth.workspaces(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role         TEXT NOT NULL DEFAULT 'member'
                     CHECK (role IN ('owner','admin','member','viewer')),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_user ON auth.workspace_members(user_id);
```

### `migrations/002_meta_schema.sql`

```sql
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
```

### `migrations/003_permissions_schema.sql`

```sql
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
```

---

## `internal/middleware/auth.go`

```go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/yourorg/datagrid/pkg/jwt"
    "github.com/yourorg/datagrid/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "user_email"

func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            if header == "" {
                response.Unauthorized(w, "missing authorization header")
                return
            }

            parts := strings.SplitN(header, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" {
                response.Unauthorized(w, "invalid authorization format")
                return
            }

            claims, err := jwtManager.Validate(parts[1])
            if err != nil {
                response.Unauthorized(w, "invalid or expired token")
                return
            }

            ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
            ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func GetUserID(ctx context.Context) string {
    if v, ok := ctx.Value(UserIDKey).(string); ok {
        return v
    }
    return ""
}
```

---

## `internal/middleware/cors.go`

```go
package middleware

import (
    "github.com/go-chi/cors"
    "net/http"
)

func CORS() func(http.Handler) http.Handler {
    return cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           300,
    })
}
```

---

## `internal/auth/repository.go`

```go
package auth

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID           string  `json:"id"`
    Email        string  `json:"email"`
    PasswordHash string  `json:"-"`
    Name         string  `json:"name"`
    AvatarURL    *string `json:"avatar_url,omitempty"`
}

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, email, passwordHash, name string) (*User, error) {
    user := &User{}
    err := r.pool.QueryRow(ctx, `
        INSERT INTO auth.users (email, password_hash, name)
        VALUES ($1, $2, $3)
        RETURNING id, email, password_hash, name, avatar_url
    `, email, passwordHash, name).Scan(
        &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL,
    )
    if err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }
    return user, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
    user := &User{}
    err := r.pool.QueryRow(ctx, `
        SELECT id, email, password_hash, name, avatar_url
        FROM auth.users
        WHERE email = $1
    `, email).Scan(
        &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL,
    )
    if err != nil {
        return nil, fmt.Errorf("get user by email: %w", err)
    }
    return user, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
    user := &User{}
    err := r.pool.QueryRow(ctx, `
        SELECT id, email, password_hash, name, avatar_url
        FROM auth.users
        WHERE id = $1
    `, id).Scan(
        &user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL,
    )
    if err != nil {
        return nil, fmt.Errorf("get user by id: %w", err)
    }
    return user, nil
}
```

---

## `internal/auth/service.go`

```go
package auth

import (
    "context"
    "errors"
    "fmt"

    "github.com/yourorg/datagrid/pkg/hasher"
    "github.com/yourorg/datagrid/pkg/jwt"
)

type Service struct {
    repo       *Repository
    jwtManager *jwt.Manager
}

func NewService(repo *Repository, jwtManager *jwt.Manager) *Service {
    return &Service{repo: repo, jwtManager: jwtManager}
}

type RegisterInput struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
}

type LoginInput struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type AuthResponse struct {
    AccessToken string `json:"access_token"`
    User        *User  `json:"user"`
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
    if input.Email == "" || input.Password == "" || input.Name == "" {
        return nil, errors.New("email, password and name are required")
    }
    if len(input.Password) < 6 {
        return nil, errors.New("password must be at least 6 characters")
    }

    hash, err := hasher.Hash(input.Password)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }

    user, err := s.repo.Create(ctx, input.Email, hash, input.Name)
    if err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    token, err := s.jwtManager.Generate(user.ID, user.Email)
    if err != nil {
        return nil, fmt.Errorf("generate token: %w", err)
    }

    return &AuthResponse{AccessToken: token, User: user}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
    user, err := s.repo.GetByEmail(ctx, input.Email)
    if err != nil {
        return nil, errors.New("invalid credentials")
    }

    if !hasher.Check(input.Password, user.PasswordHash) {
        return nil, errors.New("invalid credentials")
    }

    token, err := s.jwtManager.Generate(user.ID, user.Email)
    if err != nil {
        return nil, fmt.Errorf("generate token: %w", err)
    }

    return &AuthResponse{AccessToken: token, User: user}, nil
}

func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
    return s.repo.GetByID(ctx, userID)
}
```

---

## `internal/auth/handler.go`

```go
package auth

import (
    "encoding/json"
    "net/http"

    "github.com/yourorg/datagrid/internal/middleware"
    "github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var input RegisterInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    res, err := h.service.Register(r.Context(), input)
    if err != nil {
        response.BadRequest(w, err.Error())
        return
    }

    response.Created(w, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var input LoginInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    res, err := h.service.Login(r.Context(), input)
    if err != nil {
        response.Unauthorized(w, err.Error())
        return
    }

    response.OK(w, res)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())
    user, err := h.service.Me(r.Context(), userID)
    if err != nil {
        response.NotFound(w, "user not found")
        return
    }
    response.OK(w, user)
}
```

---

## `internal/ddl/builder.go`

```go
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
        sb.WriteString(fmt.Sprintf(" CHECK (%s ~* '^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$')", quote(col.Name)))
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
```

---

## `internal/ddl/executor.go`

```go
package ddl

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Executor struct {
    pool *pgxpool.Pool
}

func NewExecutor(pool *pgxpool.Pool) *Executor {
    return &Executor{pool: pool}
}

// ExecInTx выполняет DDL-запрос и функцию с метаданными в одной транзакции
func (e *Executor) ExecInTx(ctx context.Context, ddlQuery string, metaFn func(pgx.Tx) error) error {
    tx, err := e.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer func() { _ = tx.Rollback(ctx) }()

    if _, err := tx.Exec(ctx, ddlQuery); err != nil {
        return fmt.Errorf("exec ddl: %w", err)
    }

    if err := metaFn(tx); err != nil {
        return fmt.Errorf("meta fn: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }

    return nil
}

// ExecRaw выполняет DDL без транзакции (для CREATE SCHEMA)
func (e *Executor) ExecRaw(ctx context.Context, ddlQuery string) error {
    if _, err := e.pool.Exec(ctx, ddlQuery); err != nil {
        return fmt.Errorf("exec raw ddl: %w", err)
    }
    return nil
}
```

---

## `internal/workspace/repository.go`

```go
package workspace

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Workspace struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    OwnerID   string `json:"owner_id"`
    DBSchema  string `json:"db_schema"`
    CreatedAt string `json:"created_at"`
    Role      string `json:"role,omitempty"`
}

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, name, ownerID, schema string) (*Workspace, error) {
    ws := &Workspace{}
    err := r.pool.QueryRow(ctx, `
        INSERT INTO auth.workspaces (name, owner_id, db_schema)
        VALUES ($1, $2, $3)
        RETURNING id, name, owner_id, db_schema, created_at
    `, name, ownerID, schema).Scan(
        &ws.ID, &ws.Name, &ws.OwnerID, &ws.DBSchema, &ws.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("create workspace: %w", err)
    }
    return ws, nil
}

func (r *Repository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO auth.workspace_members (workspace_id, user_id, role)
        VALUES ($1, $2, $3)
        ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = $3
    `, workspaceID, userID, role)
    return err
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*Workspace, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT w.id, w.name, w.owner_id, w.db_schema, w.created_at, wm.role
        FROM auth.workspaces w
        JOIN auth.workspace_members wm ON wm.workspace_id = w.id
        WHERE wm.user_id = $1
        ORDER BY w.created_at DESC
    `, userID)
    if err != nil {
        return nil, fmt.Errorf("list workspaces: %w", err)
    }
    defer rows.Close()

    var result []*Workspace
    for rows.Next() {
        ws := &Workspace{}
        if err := rows.Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.DBSchema, &ws.CreatedAt, &ws.Role); err != nil {
            return nil, err
        }
        result = append(result, ws)
    }
    return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Workspace, error) {
    ws := &Workspace{}
    err := r.pool.QueryRow(ctx, `
        SELECT id, name, owner_id, db_schema, created_at
        FROM auth.workspaces WHERE id = $1
    `, id).Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.DBSchema, &ws.CreatedAt)
    if err != nil {
        return nil, fmt.Errorf("get workspace: %w", err)
    }
    return ws, nil
}
```

---

## `internal/workspace/service.go`

```go
package workspace

import (
    "context"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/yourorg/datagrid/internal/ddl"
)

type Service struct {
    repo    *Repository
    ddlExec *ddl.Executor
}

func NewService(repo *Repository, ddlExec *ddl.Executor) *Service {
    return &Service{repo: repo, ddlExec: ddlExec}
}

func (s *Service) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
    schemaName := "data_" + strings.ReplaceAll(uuid.New().String(), "-", "")

    if err := s.ddlExec.ExecRaw(ctx, ddl.BuildCreateSchema(schemaName)); err != nil {
        return nil, fmt.Errorf("create schema: %w", err)
    }

    ws, err := s.repo.Create(ctx, name, ownerID, schemaName)
    if err != nil {
        return nil, err
    }

    if err := s.repo.AddMember(ctx, ws.ID, ownerID, "owner"); err != nil {
        return nil, fmt.Errorf("add owner member: %w", err)
    }

    return ws, nil
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]*Workspace, error) {
    return s.repo.ListByUser(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Workspace, error) {
    return s.repo.GetByID(ctx, id)
}
```

---

## `internal/workspace/handler.go`

```go
package workspace

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yourorg/datagrid/internal/middleware"
    "github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserID(r.Context())
    workspaces, err := h.service.ListByUser(r.Context(), userID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.OK(w, workspaces)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }
    if input.Name == "" {
        response.BadRequest(w, "name is required")
        return
    }

    userID := middleware.GetUserID(r.Context())
    ws, err := h.service.Create(r.Context(), input.Name, userID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.Created(w, ws)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    ws, err := h.service.GetByID(r.Context(), id)
    if err != nil {
        response.NotFound(w, "workspace not found")
        return
    }
    response.OK(w, ws)
}
```

---

## `internal/field/types.go`

```go
package field

import "fmt"

var typeConversionMatrix = map[string][]string{
    "text":     {"email", "url", "phone"},
    "integer":  {"decimal", "text"},
    "decimal":  {"text"},
    "date":     {"datetime", "text"},
    "datetime": {"text"},
    "email":    {"text"},
    "url":      {"text"},
    "phone":    {"text"},
}

func IsConversionAllowed(from, to string) bool {
    if from == to {
        return true
    }
    allowed, ok := typeConversionMatrix[from]
    if !ok {
        return false
    }
    for _, t := range allowed {
        if t == to {
            return true
        }
    }
    return false
}

func ValidFieldType(t string) bool {
    valid := map[string]bool{
        "text": true, "integer": true, "decimal": true, "boolean": true,
        "date": true, "datetime": true, "select": true, "multi_select": true,
        "email": true, "url": true, "phone": true, "file": true, "relation": true,
    }
    return valid[t]
}

func ValidateOptions(fieldType string, options map[string]interface{}) error {
    if (fieldType == "select" || fieldType == "multi_select") {
        choices, ok := options["choices"]
        if !ok || choices == nil {
            return fmt.Errorf("select field requires 'choices' option")
        }
    }
    return nil
}
```

---

## `internal/field/repository.go`

```go
package field

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Field struct {
    ID             string                 `json:"id"`
    SpreadsheetID  string                 `json:"spreadsheet_id"`
    Name           string                 `json:"name"`
    ColumnName     string                 `json:"column_name"`
    FieldType      string                 `json:"field_type"`
    Position       int                    `json:"position"`
    IsRequired     bool                   `json:"is_required"`
    IsUnique       bool                   `json:"is_unique"`
    DefaultValue   *string                `json:"default_value,omitempty"`
    Options        map[string]interface{} `json:"options,omitempty"`
}

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, f *Field) (*Field, error) {
    optionsJSON, _ := json.Marshal(f.Options)

    result := &Field{}
    var optRaw []byte

    err := tx.QueryRow(ctx, `
        INSERT INTO meta.fields
            (spreadsheet_id, name, column_name, field_type, position, is_required, is_unique, default_value, options)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
        RETURNING id, spreadsheet_id, name, column_name, field_type,
                  position, is_required, is_unique, default_value, options
    `, f.SpreadsheetID, f.Name, f.ColumnName, f.FieldType,
        f.Position, f.IsRequired, f.IsUnique, f.DefaultValue, optionsJSON,
    ).Scan(
        &result.ID, &result.SpreadsheetID, &result.Name, &result.ColumnName,
        &result.FieldType, &result.Position, &result.IsRequired, &result.IsUnique,
        &result.DefaultValue, &optRaw,
    )
    if err != nil {
        return nil, fmt.Errorf("create field: %w", err)
    }

    if optRaw != nil {
        _ = json.Unmarshal(optRaw, &result.Options)
    }
    return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Field, error) {
    f := &Field{}
    var optRaw []byte
    err := r.pool.QueryRow(ctx, `
        SELECT id, spreadsheet_id, name, column_name, field_type,
               position, is_required, is_unique, default_value, options
        FROM meta.fields WHERE id = $1
    `, id).Scan(
        &f.ID, &f.SpreadsheetID, &f.Name, &f.ColumnName, &f.FieldType,
        &f.Position, &f.IsRequired, &f.IsUnique, &f.DefaultValue, &optRaw,
    )
    if err != nil {
        return nil, fmt.Errorf("get field: %w", err)
    }
    if optRaw != nil {
        _ = json.Unmarshal(optRaw, &f.Options)
    }
    return f, nil
}

func (r *Repository) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*Field, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT id, spreadsheet_id, name, column_name, field_type,
               position, is_required, is_unique, default_value, options
        FROM meta.fields
        WHERE spreadsheet_id = $1
        ORDER BY position
    `, spreadsheetID)
    if err != nil {
        return nil, fmt.Errorf("list fields: %w", err)
    }
    defer rows.Close()

    var fields []*Field
    for rows.Next() {
        f := &Field{}
        var optRaw []byte
        if err := rows.Scan(
            &f.ID, &f.SpreadsheetID, &f.Name, &f.ColumnName, &f.FieldType,
            &f.Position, &f.IsRequired, &f.IsUnique, &f.DefaultValue, &optRaw,
        ); err != nil {
            return nil, err
        }
        if optRaw != nil {
            _ = json.Unmarshal(optRaw, &f.Options)
        }
        fields = append(fields, f)
    }
    return fields, nil
}

func (r *Repository) Update(ctx context.Context, tx pgx.Tx, f *Field) (*Field, error) {
    optionsJSON, _ := json.Marshal(f.Options)
    result := &Field{}
    var optRaw []byte

    err := tx.QueryRow(ctx, `
        UPDATE meta.fields SET
            name = $1, field_type = $2, position = $3,
            is_required = $4, is_unique = $5,
            default_value = $6, options = $7
        WHERE id = $8
        RETURNING id, spreadsheet_id, name, column_name, field_type,
                  position, is_required, is_unique, default_value, options
    `, f.Name, f.FieldType, f.Position, f.IsRequired, f.IsUnique,
        f.DefaultValue, optionsJSON, f.ID,
    ).Scan(
        &result.ID, &result.SpreadsheetID, &result.Name, &result.ColumnName,
        &result.FieldType, &result.Position, &result.IsRequired, &result.IsUnique,
        &result.DefaultValue, &optRaw,
    )
    if err != nil {
        return nil, fmt.Errorf("update field: %w", err)
    }
    if optRaw != nil {
        _ = json.Unmarshal(optRaw, &result.Options)
    }
    return result, nil
}

func (r *Repository) Delete(ctx context.Context, tx pgx.Tx, id string) error {
    _, err := tx.Exec(ctx, `DELETE FROM meta.fields WHERE id = $1`, id)
    return err
}

func (r *Repository) GetMaxPosition(ctx context.Context, spreadsheetID string) (int, error) {
    var pos int
    err := r.pool.QueryRow(ctx, `
        SELECT COALESCE(MAX(position), -1) FROM meta.fields WHERE spreadsheet_id = $1
    `, spreadsheetID).Scan(&pos)
    return pos, err
}
```

---

## `internal/field/service.go`

```go
package field

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/yourorg/datagrid/internal/ddl"
    spreadsheetRepo "github.com/yourorg/datagrid/internal/spreadsheet"
)

type Service struct {
    repo        *Repository
    spreadRepo  *spreadsheetRepo.Repository
    ddlExec     *ddl.Executor
}

func NewService(repo *Repository, spreadRepo *spreadsheetRepo.Repository, ddlExec *ddl.Executor) *Service {
    return &Service{repo: repo, spreadRepo: spreadRepo, ddlExec: ddlExec}
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

func (s *Service) Create(ctx context.Context, spreadsheetID string, input CreateFieldInput) (*Field, error) {
    if input.Name == "" {
        return nil, fmt.Errorf("field name is required")
    }
    if !ValidFieldType(input.FieldType) {
        return nil, fmt.Errorf("invalid field type: %s", input.FieldType)
    }

    spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
    if err != nil {
        return nil, fmt.Errorf("spreadsheet not found: %w", err)
    }

    maxPos, _ := s.repo.GetMaxPosition(ctx, spreadsheetID)
    columnName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")

    optionsRaw, _ := json.Marshal(input.Options)
    var optionsMap map[string]interface{}
    _ = json.Unmarshal(optionsRaw, &optionsMap)

    colDef := ddl.ColumnDef{
        Name:         columnName,
        FieldType:    input.FieldType,
        IsRequired:   input.IsRequired,
        IsUnique:     input.IsUnique,
        DefaultValue: input.DefaultValue,
        Options:      optionsMap,
    }

    ddlQuery := ddl.BuildAddColumn(spread.DBSchema, spread.TableName, colDef)

    var createdField *Field

    err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx interface{ QueryRow(context.Context, string, ...any) interface{ Scan(...any) error } }) error {
        // приведение типа
        return nil
    })

    // Используем альтернативный подход через pool напрямую
    err = func() error {
        f := &Field{
            SpreadsheetID: spreadsheetID,
            Name:          input.Name,
            ColumnName:    columnName,
            FieldType:     input.FieldType,
            Position:      maxPos + 1,
            IsRequired:    input.IsRequired,
            IsUnique:      input.IsUnique,
            DefaultValue:  input.DefaultValue,
            Options:       optionsMap,
        }

        return s.ddlExec.ExecInTxFull(ctx, ddlQuery, func(tx interface {
            QueryRow(ctx context.Context, sql string, args ...any) pgxRow
        }) error {
            var err error
            createdField, err = s.repo.Create(ctx, txWrapper{tx}, f)
            return err
        })
    }()

    return createdField, err
}

func (s *Service) Delete(ctx context.Context, fieldID string) error {
    f, err := s.repo.GetByID(ctx, fieldID)
    if err != nil {
        return fmt.Errorf("field not found: %w", err)
    }

    spread, err := s.spreadRepo.GetByID(ctx, f.SpreadsheetID)
    if err != nil {
        return fmt.Errorf("spreadsheet not found: %w", err)
    }

    ddlQuery := ddl.BuildDropColumn(spread.DBSchema, spread.TableName, f.ColumnName)

    return s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgxTx) error {
        return s.repo.Delete(ctx, tx, fieldID)
    })
}

func (s *Service) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*Field, error) {
    return s.repo.ListBySpreadsheet(ctx, spreadsheetID)
}
```

---

## `internal/spreadsheet/repository.go`

```go
package spreadsheet

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

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

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, s *Spreadsheet) (*Spreadsheet, error) {
    result := &Spreadsheet{}
    err := tx.QueryRow(ctx, `
        INSERT INTO meta.spreadsheets (workspace_id, name, table_name, description, created_by)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, workspace_id, name, table_name, description, created_by, created_at, updated_at
    `, s.WorkspaceID, s.Name, s.TableName, s.Description, s.CreatedBy,
    ).Scan(
        &result.ID, &result.WorkspaceID, &result.Name, &result.TableName,
        &result.Description, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("create spreadsheet: %w", err)
    }
    return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Spreadsheet, error) {
    s := &Spreadsheet{}
    err := r.pool.QueryRow(ctx, `
        SELECT s.id, s.workspace_id, s.name, s.table_name,
               COALESCE(s.description,''), s.created_by, s.created_at, s.updated_at,
               w.db_schema
        FROM meta.spreadsheets s
        JOIN auth.workspaces w ON w.id = s.workspace_id
        WHERE s.id = $1
    `, id).Scan(
        &s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
        &s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
        &s.DBSchema,
    )
    if err != nil {
        return nil, fmt.Errorf("get spreadsheet: %w", err)
    }
    return s, nil
}

func (r *Repository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*Spreadsheet, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT id, workspace_id, name, table_name,
               COALESCE(description,''), created_by, created_at, updated_at
        FROM meta.spreadsheets
        WHERE workspace_id = $1
        ORDER BY created_at DESC
    `, workspaceID)
    if err != nil {
        return nil, fmt.Errorf("list spreadsheets: %w", err)
    }
    defer rows.Close()

    var result []*Spreadsheet
    for rows.Next() {
        s := &Spreadsheet{}
        if err := rows.Scan(
            &s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
            &s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        result = append(result, s)
    }
    return result, nil
}

func (r *Repository) Update(ctx context.Context, id, name, description string) (*Spreadsheet, error) {
    s := &Spreadsheet{}
    err := r.pool.QueryRow(ctx, `
        UPDATE meta.spreadsheets
        SET name = $1, description = $2, updated_at = now()
        WHERE id = $3
        RETURNING id, workspace_id, name, table_name,
                  COALESCE(description,''), created_by, created_at, updated_at
    `, name, description, id).Scan(
        &s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
        &s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
    )
    return s, err
}

func (r *Repository) Delete(ctx context.Context, tx pgx.Tx, id string) error {
    _, err := tx.Exec(ctx, `DELETE FROM meta.spreadsheets WHERE id = $1`, id)
    return err
}
```

---

## `internal/spreadsheet/service.go`

```go
package spreadsheet

import (
    "context"
    "fmt"
    "strings"

    "github.com/google/uuid"
    "github.com/yourorg/datagrid/internal/ddl"
    "github.com/yourorg/datagrid/internal/field"
    workspaceRepo "github.com/yourorg/datagrid/internal/workspace"
)

type Service struct {
    repo       *Repository
    fieldRepo  *field.Repository
    wsRepo     *workspaceRepo.Repository
    ddlExec    *ddl.Executor
}

func NewService(
    repo *Repository,
    fieldRepo *field.Repository,
    wsRepo *workspaceRepo.Repository,
    ddlExec *ddl.Executor,
) *Service {
    return &Service{repo: repo, fieldRepo: fieldRepo, wsRepo: wsRepo, ddlExec: ddlExec}
}

type CreateInput struct {
    WorkspaceID string              `json:"workspace_id"`
    Name        string              `json:"name"`
    Description string              `json:"description"`
    Fields      []field.CreateFieldInput `json:"fields"`
}

type SpreadsheetWithFields struct {
    *Spreadsheet
    Fields []*field.Field `json:"fields"`
}

func (s *Service) Create(ctx context.Context, input CreateInput, createdBy string) (*SpreadsheetWithFields, error) {
    if input.Name == "" {
        return nil, fmt.Errorf("name is required")
    }

    ws, err := s.wsRepo.GetByID(ctx, input.WorkspaceID)
    if err != nil {
        return nil, fmt.Errorf("workspace not found: %w", err)
    }

    tableName := "tbl_" + strings.ReplaceAll(uuid.New().String(), "-", "")

    // Подготовить колонки для DDL
    colDefs := make([]ddl.ColumnDef, 0, len(input.Fields))
    fieldRecords := make([]*field.Field, 0, len(input.Fields))

    for i, fi := range input.Fields {
        colName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")
        colDefs = append(colDefs, ddl.ColumnDef{
            Name:         colName,
            FieldType:    fi.FieldType,
            IsRequired:   fi.IsRequired,
            IsUnique:     fi.IsUnique,
            DefaultValue: fi.DefaultValue,
            Options:      fi.Options,
        })
        fieldRecords = append(fieldRecords, &field.Field{
            Name:         fi.Name,
            ColumnName:   colName,
            FieldType:    fi.FieldType,
            Position:     i,
            IsRequired:   fi.IsRequired,
            IsUnique:     fi.IsUnique,
            DefaultValue: fi.DefaultValue,
            Options:      fi.Options,
        })
    }

    ddlQuery := ddl.BuildCreateTable(ws.DBSchema, tableName, colDefs)

    spread := &Spreadsheet{
        WorkspaceID: input.WorkspaceID,
        Name:        input.Name,
        TableName:   tableName,
        Description: input.Description,
        CreatedBy:   createdBy,
    }

    var result *SpreadsheetWithFields

    err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgxTx) error {
        created, err := s.repo.Create(ctx, tx, spread)
        if err != nil {
            return err
        }

        createdFields := make([]*field.Field, 0, len(fieldRecords))
        for _, fr := range fieldRecords {
            fr.SpreadsheetID = created.ID
            cf, err := s.fieldRepo.Create(ctx, tx, fr)
            if err != nil {
                return fmt.Errorf("create field %s: %w", fr.Name, err)
            }
            createdFields = append(createdFields, cf)
        }

        result = &SpreadsheetWithFields{
            Spreadsheet: created,
            Fields:      createdFields,
        }
        return nil
    })

    return result, err
}

func (s *Service) GetWithFields(ctx context.Context, id string) (*SpreadsheetWithFields, error) {
    spread, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    fields, err := s.fieldRepo.ListBySpreadsheet(ctx, id)
    if err != nil {
        return nil, err
    }

    return &SpreadsheetWithFields{Spreadsheet: spread, Fields: fields}, nil
}

func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string) ([]*SpreadsheetWithFields, error) {
    spreads, err := s.repo.ListByWorkspace(ctx, workspaceID)
    if err != nil {
        return nil, err
    }

    result := make([]*SpreadsheetWithFields, 0, len(spreads))
    for _, sp := range spreads {
        fields, _ := s.fieldRepo.ListBySpreadsheet(ctx, sp.ID)
        result = append(result, &SpreadsheetWithFields{Spreadsheet: sp, Fields: fields})
    }
    return result, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
    spread, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    dropQuery := fmt.Sprintf(`DROP TABLE IF EXISTS "%s"."%s"`, spread.DBSchema, spread.TableName)

    return s.ddlExec.ExecInTx(ctx, dropQuery, func(tx pgxTx) error {
        return s.repo.Delete(ctx, tx, id)
    })
}
```

---

## `internal/spreadsheet/handler.go`

```go
package spreadsheet

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yourorg/datagrid/internal/middleware"
    "github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    userID := middleware.GetUserID(r.Context())
    result, err := h.service.Create(r.Context(), input, userID)
    if err != nil {
        response.BadRequest(w, err.Error())
        return
    }

    response.Created(w, result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    result, err := h.service.GetWithFields(r.Context(), id)
    if err != nil {
        response.NotFound(w, "spreadsheet not found")
        return
    }
    response.OK(w, result)
}

func (h *Handler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
    workspaceID := chi.URLParam(r, "workspaceId")
    result, err := h.service.ListByWorkspace(r.Context(), workspaceID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.OK(w, result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    var input struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    result, err := h.service.repo.Update(r.Context(), id, input.Name, input.Description)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.OK(w, result)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if err := h.service.Delete(r.Context(), id); err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.NoContent(w)
}
```

---

## `internal/row/repository.go`

```go
package row

import (
    "context"
    "fmt"
    "strings"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yourorg/datagrid/internal/field"
)

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

type QueryParams struct {
    Limit   int
    Offset  int
    Filters []FilterCondition
    Sorts   []SortCondition
}

type FilterCondition struct {
    FieldID string `json:"field_id"`
    Op      string `json:"op"`
    Value   string `json:"value"`
}

type SortCondition struct {
    FieldID   string `json:"field_id"`
    Direction string `json:"direction"`
}

type RowData map[string]interface{}

func (r *Repository) List(
    ctx context.Context,
    schema, table string,
    fields []*field.Field,
    params QueryParams,
) ([]RowData, int, error) {
    if params.Limit <= 0 { params.Limit = 50 }
    if params.Limit > 500 { params.Limit = 500 }

    // Карта field_id → column_name
    fieldMap := make(map[string]string, len(fields))
    for _, f := range fields {
        fieldMap[f.ID] = f.ColumnName
    }

    args := []interface{}{}
    argIdx := 1

    // WHERE
    whereClauses := []string{}
    for _, filter := range params.Filters {
        colName, ok := fieldMap[filter.FieldID]
        if !ok { continue }
        col := fmt.Sprintf(`"%s"`, colName)

        switch filter.Op {
        case "eq":
            whereClauses = append(whereClauses, fmt.Sprintf(`%s = $%d`, col, argIdx))
            args = append(args, filter.Value); argIdx++
        case "neq":
            whereClauses = append(whereClauses, fmt.Sprintf(`%s != $%d`, col, argIdx))
            args = append(args, filter.Value); argIdx++
        case "contains":
            whereClauses = append(whereClauses, fmt.Sprintf(`%s ILIKE $%d`, col, argIdx))
            args = append(args, "%"+filter.Value+"%"); argIdx++
        case "is_empty":
            whereClauses = append(whereClauses, fmt.Sprintf(`(%s IS NULL OR %s = '')`, col, col))
        case "is_not_empty":
            whereClauses = append(whereClauses, fmt.Sprintf(`(%s IS NOT NULL AND %s != '')`, col, col))
        case "gt":
            whereClauses = append(whereClauses, fmt.Sprintf(`%s > $%d`, col, argIdx))
            args = append(args, filter.Value); argIdx++
        case "lt":
            whereClauses = append(whereClauses, fmt.Sprintf(`%s < $%d`, col, argIdx))
            args = append(args, filter.Value); argIdx++
        }
    }

    whereSQL := ""
    if len(whereClauses) > 0 {
        whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
    }

    // ORDER BY
    orderClauses := []string{}
    for _, sort := range params.Sorts {
        colName, ok := fieldMap[sort.FieldID]
        if !ok { continue }
        dir := "ASC"
        if strings.ToLower(sort.Direction) == "desc" { dir = "DESC" }
        orderClauses = append(orderClauses, fmt.Sprintf(`"%s" %s`, colName, dir))
    }
    orderClauses = append(orderClauses, "_position ASC NULLS LAST", "_created_at ASC")
    orderSQL := "ORDER BY " + strings.Join(orderClauses, ", ")

    tableRef := fmt.Sprintf(`"%s"."%s"`, schema, table)

    // COUNT
    countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableRef, whereSQL)
    var total int
    if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
        return nil, 0, fmt.Errorf("count rows: %w", err)
    }

    // SELECT
    limitArgs := append(args, params.Limit, params.Offset)
    selectSQL := fmt.Sprintf(`
        SELECT * FROM %s
        %s
        %s
        LIMIT $%d OFFSET $%d
    `, tableRef, whereSQL, orderSQL, argIdx, argIdx+1)

    rows, err := r.pool.Query(ctx, selectSQL, limitArgs...)
    if err != nil {
        return nil, 0, fmt.Errorf("query rows: %w", err)
    }
    defer rows.Close()

    return scanRows(rows)
}

func (r *Repository) Create(ctx context.Context, schema, table string, data RowData, createdBy string) (RowData, error) {
    cols := []string{"_created_by"}
    placeholders := []string{"$1"}
    args := []interface{}{createdBy}
    idx := 2

    for col, val := range data {
        cols = append(cols, fmt.Sprintf(`"%s"`, col))
        placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
        args = append(args, val)
        idx++
    }

    sql := fmt.Sprintf(
        `INSERT INTO "%s"."%s" (%s) VALUES (%s) RETURNING *`,
        schema, table,
        strings.Join(cols, ", "),
        strings.Join(placeholders, ", "),
    )

    rows, err := r.pool.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("insert row: %w", err)
    }
    defer rows.Close()

    result, _, err := scanRows(rows)
    if err != nil || len(result) == 0 {
        return nil, fmt.Errorf("scan inserted row: %w", err)
    }
    return result[0], nil
}

func (r *Repository) Update(ctx context.Context, schema, table, rowID string, data RowData) (RowData, error) {
    setClauses := []string{"_updated_at = now()"}
    args := []interface{}{}
    idx := 1

    for col, val := range data {
        setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, idx))
        args = append(args, val)
        idx++
    }

    args = append(args, rowID)
    sql := fmt.Sprintf(
        `UPDATE "%s"."%s" SET %s WHERE _id = $%d RETURNING *`,
        schema, table,
        strings.Join(setClauses, ", "),
        idx,
    )

    rows, err := r.pool.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("update row: %w", err)
    }
    defer rows.Close()

    result, _, err := scanRows(rows)
    if err != nil || len(result) == 0 {
        return nil, fmt.Errorf("scan updated row: %w", err)
    }
    return result[0], nil
}

func (r *Repository) Delete(ctx context.Context, schema, table, rowID string) error {
    sql := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE _id = $1`, schema, table)
    _, err := r.pool.Exec(ctx, sql, rowID)
    return err
}

func scanRows(rows pgx.Rows) ([]RowData, int, error) {
    descriptions := rows.FieldDescriptions()
    var result []RowData

    for rows.Next() {
        values, err := rows.Values()
        if err != nil {
            return nil, 0, err
        }
        row := make(RowData, len(descriptions))
        for i, desc := range descriptions {
            row[string(desc.Name)] = values[i]
        }
        result = append(result, row)
    }

    return result, len(result), rows.Err()
}
```

---

## `internal/row/handler.go`

```go
package row

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yourorg/datagrid/internal/field"
    "github.com/yourorg/datagrid/internal/middleware"
    spreadsheetRepo "github.com/yourorg/datagrid/internal/spreadsheet"
    "github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
    repo       *Repository
    spreadRepo *spreadsheetRepo.Repository
    fieldRepo  *field.Repository
}

func NewHandler(repo *Repository, spreadRepo *spreadsheetRepo.Repository, fieldRepo *field.Repository) *Handler {
    return &Handler{repo: repo, spreadRepo: spreadRepo, fieldRepo: fieldRepo}
}

type queryRequest struct {
    Limit   int               `json:"limit"`
    Offset  int               `json:"offset"`
    Filters []FilterCondition `json:"filters"`
    Sorts   []SortCondition   `json:"sorts"`
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")

    var req queryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        req = queryRequest{Limit: 50}
    }

    spread, err := h.spreadRepo.GetByID(r.Context(), spreadsheetID)
    if err != nil {
        response.NotFound(w, "spreadsheet not found")
        return
    }

    fields, err := h.fieldRepo.ListBySpreadsheet(r.Context(), spreadsheetID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }

    rows, total, err := h.repo.List(r.Context(), spread.DBSchema, spread.TableName, fields, QueryParams{
        Limit:   req.Limit,
        Offset:  req.Offset,
        Filters: req.Filters,
        Sorts:   req.Sorts,
    })
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }

    response.OK(w, map[string]interface{}{
        "data":   rows,
        "total":  total,
        "limit":  req.Limit,
        "offset": req.Offset,
    })
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")

    var data RowData
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        data = RowData{}
    }

    spread, err := h.spreadRepo.GetByID(r.Context(), spreadsheetID)
    if err != nil {
        response.NotFound(w, "spreadsheet not found")
        return
    }

    userID := middleware.GetUserID(r.Context())
    row, err := h.repo.Create(r.Context(), spread.DBSchema, spread.TableName, data, userID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }

    response.Created(w, row)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")
    rowID := chi.URLParam(r, "rowId")

    var data RowData
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    spread, err := h.spreadRepo.GetByID(r.Context(), spreadsheetID)
    if err != nil {
        response.NotFound(w, "spreadsheet not found")
        return
    }

    // Убираем системные поля из обновления
    delete(data, "_id")
    delete(data, "_created_by")
    delete(data, "_created_at")
    delete(data, "_updated_at")

    row, err := h.repo.Update(r.Context(), spread.DBSchema, spread.TableName, rowID, data)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }

    response.OK(w, row)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")
    rowID := chi.URLParam(r, "rowId")

    spread, err := h.spreadRepo.GetByID(r.Context(), spreadsheetID)
    if err != nil {
        response.NotFound(w, "spreadsheet not found")
        return
    }

    if err := h.repo.Delete(r.Context(), spread.DBSchema, spread.TableName, rowID); err != nil {
        response.InternalError(w, err.Error())
        return
    }

    response.NoContent(w)
}
```

---

## `internal/permission/repository.go`

```go
package permission

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type SpreadsheetAccess struct {
    ID            string `json:"id"`
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
    ID            string `json:"id"`
    FieldID       string `json:"field_id"`
    PrincipalID   string `json:"principal_id"`
    PrincipalType string `json:"principal_type"`
    CanView       bool   `json:"can_view"`
    CanEdit       bool   `json:"can_edit"`
}

type Repository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{pool: pool}
}

func (r *Repository) GetSpreadsheetAccess(ctx context.Context, spreadsheetID string) ([]*SpreadsheetAccess, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT sa.id, sa.spreadsheet_id, sa.principal_id, sa.principal_type,
               COALESCE(u.name, sa.principal_type) as principal_name,
               sa.can_view, sa.can_insert, sa.can_edit, sa.can_delete, sa.can_manage
        FROM meta.spreadsheet_access sa
        LEFT JOIN auth.users u ON u.id::text = sa.principal_id::text
        WHERE sa.spreadsheet_id = $1
    `, spreadsheetID)
    if err != nil {
        return nil, fmt.Errorf("get spreadsheet access: %w", err)
    }
    defer rows.Close()

    var result []*SpreadsheetAccess
    for rows.Next() {
        a := &SpreadsheetAccess{}
        if err := rows.Scan(
            &a.ID, &a.SpreadsheetID, &a.PrincipalID, &a.PrincipalType,
            &a.PrincipalName, &a.CanView, &a.CanInsert, &a.CanEdit, &a.CanDelete, &a.CanManage,
        ); err != nil {
            return nil, err
        }
        result = append(result, a)
    }
    return result, nil
}

func (r *Repository) UpsertSpreadsheetAccess(ctx context.Context, a *SpreadsheetAccess) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO meta.spreadsheet_access
            (spreadsheet_id, principal_id, principal_type, can_view, can_insert, can_edit, can_delete, can_manage)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (spreadsheet_id, principal_id, principal_type)
        DO UPDATE SET
            can_view = $4, can_insert = $5, can_edit = $6,
            can_delete = $7, can_manage = $8
    `, a.SpreadsheetID, a.PrincipalID, a.PrincipalType,
        a.CanView, a.CanInsert, a.CanEdit, a.CanDelete, a.CanManage,
    )
    return err
}

func (r *Repository) RemoveSpreadsheetAccess(ctx context.Context, spreadsheetID, principalID string) error {
    _, err := r.pool.Exec(ctx, `
        DELETE FROM meta.spreadsheet_access
        WHERE spreadsheet_id = $1 AND principal_id = $2
    `, spreadsheetID, principalID)
    return err
}

func (r *Repository) CheckAccess(ctx context.Context, spreadsheetID, userID string) (*SpreadsheetAccess, error) {
    a := &SpreadsheetAccess{}
    err := r.pool.QueryRow(ctx, `
        SELECT can_view, can_insert, can_edit, can_delete, can_manage
        FROM meta.spreadsheet_access
        WHERE spreadsheet_id = $1 AND principal_id = $2
    `, spreadsheetID, userID).Scan(
        &a.CanView, &a.CanInsert, &a.CanEdit, &a.CanDelete, &a.CanManage,
    )
    if err != nil {
        return nil, err
    }
    return a, nil
}

func (r *Repository) GetFieldAccess(ctx context.Context, spreadsheetID string) ([]*FieldAccess, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT fa.id, fa.field_id, fa.principal_id, fa.principal_type, fa.can_view, fa.can_edit
        FROM meta.field_access fa
        JOIN meta.fields f ON f.id = fa.field_id
        WHERE f.spreadsheet_id = $1
    `, spreadsheetID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []*FieldAccess
    for rows.Next() {
        fa := &FieldAccess{}
        if err := rows.Scan(&fa.ID, &fa.FieldID, &fa.PrincipalID, &fa.PrincipalType, &fa.CanView, &fa.CanEdit); err != nil {
            return nil, err
        }
        result = append(result, fa)
    }
    return result, nil
}

func (r *Repository) UpsertFieldAccess(ctx context.Context, fa *FieldAccess) error {
    _, err := r.pool.Exec(ctx, `
        INSERT INTO meta.field_access (field_id, principal_id, principal_type, can_view, can_edit)
        VALUES ($1,$2,$3,$4,$5)
        ON CONFLICT (field_id, principal_id, principal_type)
        DO UPDATE SET can_view = $4, can_edit = $5
    `, fa.FieldID, fa.PrincipalID, fa.PrincipalType, fa.CanView, fa.CanEdit)
    return err
}
```

---

## `internal/permission/handler.go`

```go
package permission

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
    repo *Repository
}

func NewHandler(repo *Repository) *Handler {
    return &Handler{repo: repo}
}

func (h *Handler) GetSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")
    accesses, err := h.repo.GetSpreadsheetAccess(r.Context(), spreadsheetID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.OK(w, accesses)
}

func (h *Handler) UpsertSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")
    var input SpreadsheetAccess
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }
    input.SpreadsheetID = spreadsheetID

    if err := h.repo.UpsertSpreadsheetAccess(r.Context(), &input); err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.NoContent(w)
}

func (h *Handler) RemoveSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")
    principalID := chi.URLParam(r, "principalId")

    if err := h.repo.RemoveSpreadsheetAccess(r.Context(), spreadsheetID, principalID); err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.NoContent(w)
}

func (h *Handler) GetFieldAccess(w http.ResponseWriter, r *http.Request) {
    spreadsheetID := chi.URLParam(r, "id")
    accesses, err := h.repo.GetFieldAccess(r.Context(), spreadsheetID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.OK(w, accesses)
}

func (h *Handler) UpsertFieldAccess(w http.ResponseWriter, r *http.Request) {
    fieldID := chi.URLParam(r, "fieldId")
    var input FieldAccess
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }
    input.FieldID = fieldID

    if err := h.repo.UpsertFieldAccess(r.Context(), &input); err != nil {
        response.InternalError(w, err.Error())
        return
    }
    response.NoContent(w)
}
```

---

## `cmd/server/main.go`

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    chiMiddleware "github.com/go-chi/chi/v5/middleware"

    "github.com/yourorg/datagrid/config"
    authPkg "github.com/yourorg/datagrid/internal/auth"
    "github.com/yourorg/datagrid/internal/ddl"
    fieldPkg "github.com/yourorg/datagrid/internal/field"
    "github.com/yourorg/datagrid/internal/middleware"
    permissionPkg "github.com/yourorg/datagrid/internal/permission"
    rowPkg "github.com/yourorg/datagrid/internal/row"
    spreadsheetPkg "github.com/yourorg/datagrid/internal/spreadsheet"
    workspacePkg "github.com/yourorg/datagrid/internal/workspace"
    "github.com/yourorg/datagrid/pkg/jwt"
    "github.com/yourorg/datagrid/pkg/postgres"
)

func main() {
    cfg := config.Load()

    // База данных
    ctx := context.Background()
    pool, err := postgres.NewPool(ctx, cfg.DB.DSN())
    if err != nil {
        log.Fatalf("connect to db: %v", err)
    }
    defer pool.Close()

    log.Println("connected to database")

    // JWT
    jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)

    // DDL executor
    ddlExec := ddl.NewExecutor(pool)

    // Repositories
    authRepo        := authPkg.NewRepository(pool)
    wsRepo          := workspacePkg.NewRepository(pool)
    spreadsheetRepo := spreadsheetPkg.NewRepository(pool)
    fieldRepo       := fieldPkg.NewRepository(pool)
    rowRepo         := rowPkg.NewRepository(pool)
    permRepo        := permissionPkg.NewRepository(pool)

    // Services
    authService      := authPkg.NewService(authRepo, jwtManager)
    wsService        := workspacePkg.NewService(wsRepo, ddlExec)
    spreadsheetSvc   := spreadsheetPkg.NewService(spreadsheetRepo, fieldRepo, wsRepo, ddlExec)
    fieldService     := fieldPkg.NewService(fieldRepo, spreadsheetRepo, ddlExec)

    // Handlers
    authHandler        := authPkg.NewHandler(authService)
    wsHandler          := workspacePkg.NewHandler(wsService)
    spreadsheetHandler := spreadsheetPkg.NewHandler(spreadsheetSvc)
    fieldHandler       := fieldPkg.NewHandler(fieldService)
    rowHandler         := rowPkg.NewHandler(rowRepo, spreadsheetRepo, fieldRepo)
    permHandler        := permissionPkg.NewHandler(permRepo)

    // Router
    r := chi.NewRouter()

    r.Use(chiMiddleware.Logger)
    r.Use(chiMiddleware.Recoverer)
    r.Use(chiMiddleware.RequestID)
    r.Use(middleware.CORS())

    // Public routes
    r.Post("/auth/register", authHandler.Register)
    r.Post("/auth/login",    authHandler.Login)

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(middleware.Auth(jwtManager))

        r.Get("/auth/me", authHandler.Me)

        // Workspaces
        r.Get("/workspaces",      wsHandler.List)
        r.Post("/workspaces",     wsHandler.Create)
        r.Get("/workspaces/{id}", wsHandler.Get)

        // Spreadsheets
        r.Post("/spreadsheets",                              spreadsheetHandler.Create)
        r.Get("/workspaces/{workspaceId}/spreadsheets",      spreadsheetHandler.ListByWorkspace)
        r.Get("/spreadsheets/{id}",                          spreadsheetHandler.Get)
        r.Put("/spreadsheets/{id}",                          spreadsheetHandler.Update)
        r.Delete("/spreadsheets/{id}",                       spreadsheetHandler.Delete)

        // Fields
        r.Post("/spreadsheets/{id}/fields",   fieldHandler.Create)
        r.Put("/fields/{fieldId}",            fieldHandler.Update)
        r.Delete("/fields/{fieldId}",         fieldHandler.Delete)

        // Rows
        r.Post("/spreadsheets/{id}/rows/query", rowHandler.Query)
        r.Post("/spreadsheets/{id}/rows",        rowHandler.Create)
        r.Patch("/spreadsheets/{id}/rows/{rowId}", rowHandler.Update)
        r.Delete("/spreadsheets/{id}/rows/{rowId}", rowHandler.Delete)

        // Permissions
        r.Get("/spreadsheets/{id}/permissions",               permHandler.GetSpreadsheetAccess)
        r.Put("/spreadsheets/{id}/permissions",               permHandler.UpsertSpreadsheetAccess)
        r.Delete("/spreadsheets/{id}/permissions/{principalId}", permHandler.RemoveSpreadsheetAccess)
        r.Get("/spreadsheets/{id}/field-permissions",         permHandler.GetFieldAccess)
        r.Put("/fields/{fieldId}/permissions",                permHandler.UpsertFieldAccess)
    })

    // HTTP Server
    addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
    srv := &http.Server{
        Addr:         addr,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        log.Printf("server started on %s", addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()

    <-quit
    log.Println("shutting down...")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("server shutdown: %v", err)
    }

    log.Println("server stopped")
}
```

---

## `.env`

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=datagrid
DB_SSLMODE=disable

JWT_SECRET=your-super-secret-key-change-in-production
JWT_ACCESS_TTL_MINUTES=60
JWT_REFRESH_TTL_DAYS=30
```

---

## Команды для запуска

```bash
# Создать БД
createdb datagrid

# Применить миграции вручную
psql -d datagrid -f migrations/001_auth_schema.sql
psql -d datagrid -f migrations/002_meta_schema.sql
psql -d datagrid -f migrations/003_permissions_schema.sql

# Установить зависимости
go mod tidy

# Запустить
go run ./cmd/server/main.go
```

**User - --**

не хватает файла internal-field-handler.go
создай

**Assistant - --**

## `internal/field/handler.go`

```go
package field

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	var input CreateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "field name is required")
		return
	}

	if !ValidFieldType(input.FieldType) {
		response.BadRequest(w, "invalid field type: "+input.FieldType)
		return
	}

	if input.Options != nil {
		if err := ValidateOptions(input.FieldType, input.Options); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Create(r.Context(), spreadsheetID, input)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, field)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	var input UpdateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.FieldType != nil && !ValidFieldType(*input.FieldType) {
		response.BadRequest(w, "invalid field type: "+*input.FieldType)
		return
	}

	field, err := h.service.Update(r.Context(), fieldID, input)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, field)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	if err := h.service.Delete(r.Context(), fieldID); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

func (h *Handler) ListBySpreadsheet(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	fields, err := h.service.ListBySpreadsheet(r.Context(), spreadsheetID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, fields)
}
```

---

Также нужно обновить `internal/field/service.go` — в предыдущей версии метод `Update` отсутствовал, а логика `Create` содержала незавершённый код. Вот исправленная версия:

## `internal/field/service.go` (исправленная)

```go
package field

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourorg/datagrid/internal/ddl"
	spreadsheetRepo "github.com/yourorg/datagrid/internal/spreadsheet"
)

type Service struct {
	repo       *Repository
	spreadRepo *spreadsheetRepo.Repository
	ddlExec    *ddl.Executor
}

func NewService(repo *Repository, spreadRepo *spreadsheetRepo.Repository, ddlExec *ddl.Executor) *Service {
	return &Service{repo: repo, spreadRepo: spreadRepo, ddlExec: ddlExec}
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

func (s *Service) Create(ctx context.Context, spreadsheetID string, input CreateFieldInput) (*Field, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("field name is required")
	}
	if !ValidFieldType(input.FieldType) {
		return nil, fmt.Errorf("invalid field type: %s", input.FieldType)
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	maxPos, _ := s.repo.GetMaxPosition(ctx, spreadsheetID)
	columnName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	colDef := ddl.ColumnDef{
		Name:         columnName,
		FieldType:    input.FieldType,
		IsRequired:   input.IsRequired,
		IsUnique:     input.IsUnique,
		DefaultValue: input.DefaultValue,
		Options:      input.Options,
	}

	ddlQuery := ddl.BuildAddColumn(spread.DBSchema, spread.TableName, colDef)

	newField := &Field{
		SpreadsheetID: spreadsheetID,
		Name:          input.Name,
		ColumnName:    columnName,
		FieldType:     input.FieldType,
		Position:      maxPos + 1,
		IsRequired:    input.IsRequired,
		IsUnique:      input.IsUnique,
		DefaultValue:  input.DefaultValue,
		Options:       input.Options,
	}

	var created *Field

	err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, tx, newField)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("create field: %w", err)
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, fieldID string, input UpdateFieldInput) (*Field, error) {
	existing, err := s.repo.GetByID(ctx, fieldID)
	if err != nil {
		return nil, fmt.Errorf("field not found: %w", err)
	}

	// Применяем изменения к существующему полю
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.IsRequired != nil {
		existing.IsRequired = *input.IsRequired
	}
	if input.IsUnique != nil {
		existing.IsUnique = *input.IsUnique
	}
	if input.DefaultValue != nil {
		existing.DefaultValue = input.DefaultValue
	}
	if input.Options != nil {
		existing.Options = input.Options
	}

	// Если меняется тип — нужен ALTER COLUMN
	if input.FieldType != nil && *input.FieldType != existing.FieldType {
		if !IsConversionAllowed(existing.FieldType, *input.FieldType) {
			return nil, fmt.Errorf(
				"type conversion from '%s' to '%s' is not allowed",
				existing.FieldType, *input.FieldType,
			)
		}

		spread, err := s.spreadRepo.GetByID(ctx, existing.SpreadsheetID)
		if err != nil {
			return nil, fmt.Errorf("spreadsheet not found: %w", err)
		}

		oldType := existing.FieldType
		existing.FieldType = *input.FieldType

		colDef := ddl.ColumnDef{
			Name:      existing.ColumnName,
			FieldType: existing.FieldType,
			Options:   existing.Options,
		}

		ddlQuery := ddl.BuildAlterColumnType(spread.DBSchema, spread.TableName, colDef)

		var updated *Field
		err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
			var txErr error
			updated, txErr = s.repo.Update(ctx, tx, existing)
			return txErr
		})
		if err != nil {
			// Откатываем тип в памяти при ошибке
			existing.FieldType = oldType
			return nil, fmt.Errorf("alter column type: %w", err)
		}

		return updated, nil
	}

	// Тип не меняется — только метаданные, DDL не нужен
	var updated *Field

	err = s.ddlExec.ExecInTx(ctx, "" , func(tx pgx.Tx) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, tx, existing)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("update field metadata: %w", err)
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, fieldID string) error {
	f, err := s.repo.GetByID(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("field not found: %w", err)
	}

	spread, err := s.spreadRepo.GetByID(ctx, f.SpreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	ddlQuery := ddl.BuildDropColumn(spread.DBSchema, spread.TableName, f.ColumnName)

	return s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		return s.repo.Delete(ctx, tx, fieldID)
	})
}

func (s *Service) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*Field, error) {
	return s.repo.ListBySpreadsheet(ctx, spreadsheetID)
}
```

---

Также нужно обновить `internal/ddl/executor.go` — метод `ExecInTx` должен поддерживать пустой DDL-запрос (для случая обновления только метаданных):

## `internal/ddl/executor.go` (исправленный)

```go
package ddl

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor struct {
	pool *pgxpool.Pool
}

func NewExecutor(pool *pgxpool.Pool) *Executor {
	return &Executor{pool: pool}
}

// ExecInTx выполняет DDL-запрос и метафункцию в одной транзакции.
// Если ddlQuery пустой — выполняется только метафункция.
func (e *Executor) ExecInTx(ctx context.Context, ddlQuery string, metaFn func(pgx.Tx) error) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if ddlQuery != "" {
		if _, err := tx.Exec(ctx, ddlQuery); err != nil {
			return fmt.Errorf("exec ddl %q: %w", ddlQuery, err)
		}
	}

	if err := metaFn(tx); err != nil {
		return fmt.Errorf("meta fn: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// ExecRaw выполняет DDL без транзакции (например CREATE SCHEMA).
func (e *Executor) ExecRaw(ctx context.Context, ddlQuery string) error {
	if _, err := e.pool.Exec(ctx, ddlQuery); err != nil {
		return fmt.Errorf("exec raw ddl: %w", err)
	}
	return nil
}
```

**User - --**

не хватает файла internal - permission - enforcer.go

**Assistant - --**

## `internal/permission/enforcer.go`

```go
package permission

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	fieldPkg "github.com/yourorg/datagrid/internal/field"
)

// Enforcer проверяет права доступа пользователя к объектам системы
type Enforcer struct {
	pool    *pgxpool.Pool
	permRepo *Repository
	fieldRepo *fieldPkg.Repository
}

func NewEnforcer(pool *pgxpool.Pool, permRepo *Repository, fieldRepo *fieldPkg.Repository) *Enforcer {
	return &Enforcer{
		pool:      pool,
		permRepo:  permRepo,
		fieldRepo: fieldRepo,
	}
}

// ─── Права на таблицу ─────────────────────────────────────────────────────────

// SpreadsheetPerms содержит все права пользователя на конкретную таблицу
type SpreadsheetPerms struct {
	CanView   bool
	CanInsert bool
	CanEdit   bool
	CanDelete bool
	CanManage bool
}

// GetSpreadsheetPerms возвращает права пользователя на таблицу.
// Владелец workspace получает полные права автоматически.
func (e *Enforcer) GetSpreadsheetPerms(ctx context.Context, userID, spreadsheetID string) (*SpreadsheetPerms, error) {
	// Проверяем, является ли пользователь владельцем workspace
	isOwner, err := e.isWorkspaceOwner(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("check workspace owner: %w", err)
	}
	if isOwner {
		return &SpreadsheetPerms{
			CanView:   true,
			CanInsert: true,
			CanEdit:   true,
			CanDelete: true,
			CanManage: true,
		}, nil
	}

	// Проверяем явные права
	access, err := e.permRepo.CheckAccess(ctx, spreadsheetID, userID)
	if err != nil {
		// Нет записи — нет доступа
		return &SpreadsheetPerms{}, nil
	}

	return &SpreadsheetPerms{
		CanView:   access.CanView,
		CanInsert: access.CanInsert,
		CanEdit:   access.CanEdit,
		CanDelete: access.CanDelete,
		CanManage: access.CanManage,
	}, nil
}

// RequireView возвращает ошибку если пользователь не может просматривать таблицу
func (e *Enforcer) RequireView(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanView {
		return ErrForbidden("view", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireInsert возвращает ошибку если пользователь не может добавлять строки
func (e *Enforcer) RequireInsert(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanInsert {
		return ErrForbidden("insert", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireEdit возвращает ошибку если пользователь не может редактировать строки
func (e *Enforcer) RequireEdit(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanEdit {
		return ErrForbidden("edit", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireDelete возвращает ошибку если пользователь не может удалять строки
func (e *Enforcer) RequireDelete(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanDelete {
		return ErrForbidden("delete", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireManage возвращает ошибку если пользователь не может управлять структурой таблицы
func (e *Enforcer) RequireManage(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanManage {
		return ErrForbidden("manage", "spreadsheet", spreadsheetID)
	}
	return nil
}

// ─── Права на поля ────────────────────────────────────────────────────────────

// FieldPerms содержит права пользователя на конкретное поле
type FieldPerms struct {
	CanView bool
	CanEdit bool
}

// GetFieldPerms возвращает права пользователя на конкретное поле.
// Если явных ограничений нет — по умолчанию разрешено всё (наследует от таблицы).
func (e *Enforcer) GetFieldPerms(ctx context.Context, userID, fieldID string) (*FieldPerms, error) {
	var canView, canEdit bool

	err := e.pool.QueryRow(ctx, `
		SELECT
			COALESCE(fa.can_view, true),
			COALESCE(fa.can_edit, true)
		FROM meta.fields f
		LEFT JOIN meta.field_access fa
			ON fa.field_id = f.id
			AND fa.principal_id = $2
		WHERE f.id = $1
	`, fieldID, userID).Scan(&canView, &canEdit)
	if err != nil {
		return nil, fmt.Errorf("get field perms: %w", err)
	}

	return &FieldPerms{CanView: canView, CanEdit: canEdit}, nil
}

// VisibleFields возвращает только те поля, которые пользователь может видеть
func (e *Enforcer) VisibleFields(ctx context.Context, userID, spreadsheetID string) ([]*fieldPkg.Field, error) {
	fields, err := e.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	// Загружаем все field_access записи для этого пользователя одним запросом
	hiddenFields, err := e.getHiddenFieldIDs(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if len(hiddenFields) == 0 {
		return fields, nil
	}

	visible := make([]*fieldPkg.Field, 0, len(fields))
	for _, f := range fields {
		if !hiddenFields[f.ID] {
			visible = append(visible, f)
		}
	}
	return visible, nil
}

// EditableFields возвращает только те поля, которые пользователь может редактировать
func (e *Enforcer) EditableFields(ctx context.Context, userID, spreadsheetID string) ([]*fieldPkg.Field, error) {
	fields, err := e.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	readonlyFields, err := e.getReadonlyFieldIDs(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if len(readonlyFields) == 0 {
		return fields, nil
	}

	editable := make([]*fieldPkg.Field, 0, len(fields))
	for _, f := range fields {
		if !readonlyFields[f.ID] {
			editable = append(editable, f)
		}
	}
	return editable, nil
}

// ─── Row-level фильтрация ─────────────────────────────────────────────────────

// RowFilter описывает SQL-условие для фильтрации строк
type RowFilter struct {
	SQL  string
	Args []interface{}
}

// GetRowFilters возвращает SQL-условия для фильтрации строк
// на основе row_access_rules для данного пользователя
func (e *Enforcer) GetRowFilters(ctx context.Context, userID, spreadsheetID string) ([]RowFilter, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT condition, can_view
		FROM meta.row_access_rules
		WHERE spreadsheet_id = $1
		  AND principal_id = $2
		  AND can_view = true
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}
	defer rows.Close()

	type ruleRow struct {
		Condition map[string]interface{}
		CanView   bool
	}

	var rules []ruleRow
	for rows.Next() {
		var condJSON []byte
		var canView bool
		if err := rows.Scan(&condJSON, &canView); err != nil {
			return nil, err
		}

		var cond map[string]interface{}
		if err := parseJSON(condJSON, &cond); err != nil {
			continue
		}
		rules = append(rules, ruleRow{Condition: cond, CanView: canView})
	}

	if len(rules) == 0 {
		return nil, nil
	}

	var filters []RowFilter
	for _, rule := range rules {
		filter, err := e.buildRowFilter(rule.Condition, userID)
		if err != nil {
			continue
		}
		filters = append(filters, filter)
	}

	return filters, nil
}

// buildRowFilter преобразует JSON-условие правила в SQL-фрагмент
func (e *Enforcer) buildRowFilter(condition map[string]interface{}, userID string) (RowFilter, error) {
	colName, ok := condition["column_name"].(string)
	if !ok || colName == "" {
		return RowFilter{}, fmt.Errorf("missing column_name in condition")
	}

	op, ok := condition["op"].(string)
	if !ok || op == "" {
		return RowFilter{}, fmt.Errorf("missing op in condition")
	}

	quotedCol := fmt.Sprintf(`"%s"`, colName)

	switch op {
	case "eq_current_user":
		return RowFilter{
			SQL:  fmt.Sprintf(`%s = $1`, quotedCol),
			Args: []interface{}{userID},
		}, nil

	case "eq":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s = $1`, quotedCol),
			Args: []interface{}{val},
		}, nil

	case "neq":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s != $1`, quotedCol),
			Args: []interface{}{val},
		}, nil

	case "contains":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s ILIKE $1`, quotedCol),
			Args: []interface{}{"%" + val + "%"},
		}, nil

	default:
		return RowFilter{}, fmt.Errorf("unknown op: %s", op)
	}
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// isWorkspaceOwner проверяет, является ли пользователь владельцем workspace
// к которому принадлежит таблица
func (e *Enforcer) isWorkspaceOwner(ctx context.Context, userID, spreadsheetID string) (bool, error) {
	var isOwner bool
	err := e.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM auth.workspaces w
			JOIN meta.spreadsheets s ON s.workspace_id = w.id
			WHERE s.id = $1
			  AND w.owner_id = $2
		)
	`, spreadsheetID, userID).Scan(&isOwner)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}

// getHiddenFieldIDs возвращает set field_id, которые скрыты для пользователя
func (e *Enforcer) getHiddenFieldIDs(ctx context.Context, userID, spreadsheetID string) (map[string]bool, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT fa.field_id
		FROM meta.field_access fa
		JOIN meta.fields f ON f.id = fa.field_id
		WHERE f.spreadsheet_id = $1
		  AND fa.principal_id = $2
		  AND fa.can_view = false
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get hidden fields: %w", err)
	}
	defer rows.Close()

	hidden := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		hidden[id] = true
	}
	return hidden, nil
}

// getReadonlyFieldIDs возвращает set field_id, которые нельзя редактировать
func (e *Enforcer) getReadonlyFieldIDs(ctx context.Context, userID, spreadsheetID string) (map[string]bool, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT fa.field_id
		FROM meta.field_access fa
		JOIN meta.fields f ON f.id = fa.field_id
		WHERE f.spreadsheet_id = $1
		  AND fa.principal_id = $2
		  AND fa.can_edit = false
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get readonly fields: %w", err)
	}
	defer rows.Close()

	readonly := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		readonly[id] = true
	}
	return readonly, nil
}

// ─── Ошибки ───────────────────────────────────────────────────────────────────

// AccessError описывает ошибку доступа
type AccessError struct {
	Action     string
	Resource   string
	ResourceID string
}

func (e *AccessError) Error() string {
	return fmt.Sprintf("forbidden: cannot %s %s %s", e.Action, e.Resource, e.ResourceID)
}

func ErrForbidden(action, resource, resourceID string) error {
	return &AccessError{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
	}
}

// IsForbidden проверяет, является ли ошибка ошибкой доступа
func IsForbidden(err error) bool {
	_, ok := err.(*AccessError)
	return ok
}

// ─── Утилиты ──────────────────────────────────────────────────────────────────

func parseJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty json")
	}
	import_json := func() error {
		import "encoding/json"
		return json.Unmarshal(data, v)
	}
	return import_json()
}
```

Функция `parseJSON` написана некорректно — в Go нельзя делать локальный импорт. Исправим её:

## `internal/permission/enforcer.go` (финальная версия)

```go
package permission

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	fieldPkg "github.com/yourorg/datagrid/internal/field"
)

// Enforcer проверяет права доступа пользователя к объектам системы
type Enforcer struct {
	pool      *pgxpool.Pool
	permRepo  *Repository
	fieldRepo *fieldPkg.Repository
}

func NewEnforcer(pool *pgxpool.Pool, permRepo *Repository, fieldRepo *fieldPkg.Repository) *Enforcer {
	return &Enforcer{
		pool:      pool,
		permRepo:  permRepo,
		fieldRepo: fieldRepo,
	}
}

// ─── Права на таблицу ─────────────────────────────────────────────────────────

// SpreadsheetPerms содержит все права пользователя на конкретную таблицу
type SpreadsheetPerms struct {
	CanView   bool
	CanInsert bool
	CanEdit   bool
	CanDelete bool
	CanManage bool
}

// GetSpreadsheetPerms возвращает права пользователя на таблицу.
// Владелец workspace получает полные права автоматически.
func (e *Enforcer) GetSpreadsheetPerms(ctx context.Context, userID, spreadsheetID string) (*SpreadsheetPerms, error) {
	isOwner, err := e.isWorkspaceOwner(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("check workspace owner: %w", err)
	}
	if isOwner {
		return &SpreadsheetPerms{
			CanView:   true,
			CanInsert: true,
			CanEdit:   true,
			CanDelete: true,
			CanManage: true,
		}, nil
	}

	access, err := e.permRepo.CheckAccess(ctx, spreadsheetID, userID)
	if err != nil {
		// Нет записи — нет доступа
		return &SpreadsheetPerms{}, nil
	}

	return &SpreadsheetPerms{
		CanView:   access.CanView,
		CanInsert: access.CanInsert,
		CanEdit:   access.CanEdit,
		CanDelete: access.CanDelete,
		CanManage: access.CanManage,
	}, nil
}

// RequireView возвращает ошибку если пользователь не может просматривать таблицу
func (e *Enforcer) RequireView(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanView {
		return ErrForbidden("view", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireInsert возвращает ошибку если пользователь не может добавлять строки
func (e *Enforcer) RequireInsert(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanInsert {
		return ErrForbidden("insert", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireEdit возвращает ошибку если пользователь не может редактировать строки
func (e *Enforcer) RequireEdit(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanEdit {
		return ErrForbidden("edit", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireDelete возвращает ошибку если пользователь не может удалять строки
func (e *Enforcer) RequireDelete(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanDelete {
		return ErrForbidden("delete", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireManage возвращает ошибку если пользователь не может управлять структурой таблицы
func (e *Enforcer) RequireManage(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanManage {
		return ErrForbidden("manage", "spreadsheet", spreadsheetID)
	}
	return nil
}

// ─── Права на поля ────────────────────────────────────────────────────────────

// FieldPerms содержит права пользователя на конкретное поле
type FieldPerms struct {
	CanView bool
	CanEdit bool
}

// GetFieldPerms возвращает права пользователя на конкретное поле.
// Если явных ограничений нет — по умолчанию разрешено всё.
func (e *Enforcer) GetFieldPerms(ctx context.Context, userID, fieldID string) (*FieldPerms, error) {
	var canView, canEdit bool

	err := e.pool.QueryRow(ctx, `
		SELECT
			COALESCE(fa.can_view, true),
			COALESCE(fa.can_edit, true)
		FROM meta.fields f
		LEFT JOIN meta.field_access fa
			ON fa.field_id = f.id
			AND fa.principal_id = $2
		WHERE f.id = $1
	`, fieldID, userID).Scan(&canView, &canEdit)
	if err != nil {
		return nil, fmt.Errorf("get field perms: %w", err)
	}

	return &FieldPerms{CanView: canView, CanEdit: canEdit}, nil
}

// VisibleFields возвращает только те поля, которые пользователь может видеть
func (e *Enforcer) VisibleFields(ctx context.Context, userID, spreadsheetID string) ([]*fieldPkg.Field, error) {
	fields, err := e.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	hiddenFields, err := e.getHiddenFieldIDs(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if len(hiddenFields) == 0 {
		return fields, nil
	}

	visible := make([]*fieldPkg.Field, 0, len(fields))
	for _, f := range fields {
		if !hiddenFields[f.ID] {
			visible = append(visible, f)
		}
	}
	return visible, nil
}

// EditableFields возвращает только те поля, которые пользователь может редактировать
func (e *Enforcer) EditableFields(ctx context.Context, userID, spreadsheetID string) ([]*fieldPkg.Field, error) {
	fields, err := e.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	readonlyFields, err := e.getReadonlyFieldIDs(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if len(readonlyFields) == 0 {
		return fields, nil
	}

	editable := make([]*fieldPkg.Field, 0, len(fields))
	for _, f := range fields {
		if !readonlyFields[f.ID] {
			editable = append(editable, f)
		}
	}
	return editable, nil
}

// ─── Row-level фильтрация ─────────────────────────────────────────────────────

// RowFilter описывает SQL-условие для фильтрации строк
type RowFilter struct {
	SQL  string
	Args []interface{}
}

// GetRowFilters возвращает SQL-условия для фильтрации строк
// на основе row_access_rules для данного пользователя
func (e *Enforcer) GetRowFilters(ctx context.Context, userID, spreadsheetID string) ([]RowFilter, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT condition
		FROM meta.row_access_rules
		WHERE spreadsheet_id = $1
		  AND principal_id = $2
		  AND can_view = true
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}
	defer rows.Close()

	var filters []RowFilter
	for rows.Next() {
		var condJSON []byte
		if err := rows.Scan(&condJSON); err != nil {
			return nil, err
		}

		var cond map[string]interface{}
		if err := json.Unmarshal(condJSON, &cond); err != nil {
			continue
		}

		filter, err := e.buildRowFilter(cond, userID)
		if err != nil {
			continue
		}
		filters = append(filters, filter)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return filters, nil
}

// buildRowFilter преобразует JSON-условие правила в SQL-фрагмент
func (e *Enforcer) buildRowFilter(condition map[string]interface{}, userID string) (RowFilter, error) {
	colName, ok := condition["column_name"].(string)
	if !ok || colName == "" {
		return RowFilter{}, fmt.Errorf("missing column_name in condition")
	}

	op, ok := condition["op"].(string)
	if !ok || op == "" {
		return RowFilter{}, fmt.Errorf("missing op in condition")
	}

	quotedCol := fmt.Sprintf(`"%s"`, colName)

	switch op {
	case "eq_current_user":
		return RowFilter{
			SQL:  fmt.Sprintf(`%s = $1`, quotedCol),
			Args: []interface{}{userID},
		}, nil

	case "eq":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s = $1`, quotedCol),
			Args: []interface{}{val},
		}, nil

	case "neq":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s != $1`, quotedCol),
			Args: []interface{}{val},
		}, nil

	case "contains":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s ILIKE $1`, quotedCol),
			Args: []interface{}{"%" + val + "%"},
		}, nil

	default:
		return RowFilter{}, fmt.Errorf("unknown op: %s", op)
	}
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// isWorkspaceOwner проверяет, является ли пользователь владельцем workspace
func (e *Enforcer) isWorkspaceOwner(ctx context.Context, userID, spreadsheetID string) (bool, error) {
	var isOwner bool
	err := e.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM auth.workspaces w
			JOIN meta.spreadsheets s ON s.workspace_id = w.id
			WHERE s.id = $1
			  AND w.owner_id = $2
		)
	`, spreadsheetID, userID).Scan(&isOwner)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}

// getHiddenFieldIDs возвращает set field_id скрытых для пользователя полей
func (e *Enforcer) getHiddenFieldIDs(ctx context.Context, userID, spreadsheetID string) (map[string]bool, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT fa.field_id
		FROM meta.field_access fa
		JOIN meta.fields f ON f.id = fa.field_id
		WHERE f.spreadsheet_id = $1
		  AND fa.principal_id = $2
		  AND fa.can_view = false
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get hidden fields: %w", err)
	}
	defer rows.Close()

	hidden := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		hidden[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hidden, nil
}

// getReadonlyFieldIDs возвращает set field_id полей недоступных для редактирования
func (e *Enforcer) getReadonlyFieldIDs(ctx context.Context, userID, spreadsheetID string) (map[string]bool, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT fa.field_id
		FROM meta.field_access fa
		JOIN meta.fields f ON f.id = fa.field_id
		WHERE f.spreadsheet_id = $1
		  AND fa.principal_id = $2
		  AND fa.can_edit = false
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get readonly fields: %w", err)
	}
	defer rows.Close()

	readonly := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		readonly[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return readonly, nil
}

// ─── Ошибки ───────────────────────────────────────────────────────────────────

// AccessError описывает ошибку доступа
type AccessError struct {
	Action     string
	Resource   string
	ResourceID string
}

func (e *AccessError) Error() string {
	return fmt.Sprintf("forbidden: cannot %s %s %s", e.Action, e.Resource, e.ResourceID)
}

// ErrForbidden создаёт ошибку запрета доступа
func ErrForbidden(action, resource, resourceID string) error {
	return &AccessError{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
	}
}

// IsForbidden проверяет, является ли ошибка ошибкой доступа
func IsForbidden(err error) bool {
	_, ok := err.(*AccessError)
	return ok
}
```

**User - --**

не хватает файла internal-row-service.go

**Assistant - --**

## `internal/row/service.go`

```go
package row

import (
	"context"
	"fmt"
	"strings"

	fieldPkg "github.com/yourorg/datagrid/internal/field"
	permissionPkg "github.com/yourorg/datagrid/internal/permission"
	spreadsheetPkg "github.com/yourorg/datagrid/internal/spreadsheet"
)

type Service struct {
	repo        *Repository
	spreadRepo  *spreadsheetPkg.Repository
	fieldRepo   *fieldPkg.Repository
	enforcer    *permissionPkg.Enforcer
}

func NewService(
	repo *Repository,
	spreadRepo *spreadsheetPkg.Repository,
	fieldRepo *fieldPkg.Repository,
	enforcer *permissionPkg.Enforcer,
) *Service {
	return &Service{
		repo:       repo,
		spreadRepo: spreadRepo,
		fieldRepo:  fieldRepo,
		enforcer:   enforcer,
	}
}

// ─── Query ────────────────────────────────────────────────────────────────────

type QueryInput struct {
	Limit   int
	Offset  int
	Filters []FilterCondition
	Sorts   []SortCondition
}

type QueryResult struct {
	Data   []RowData `json:"data"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

func (s *Service) Query(ctx context.Context, userID, spreadsheetID string, input QueryInput) (*QueryResult, error) {
	// Проверяем право на просмотр
	if err := s.enforcer.RequireView(ctx, userID, spreadsheetID); err != nil {
		return nil, err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	// Получаем только видимые поля
	visibleFields, err := s.enforcer.VisibleFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get visible fields: %w", err)
	}

	// Фильтруем входящие фильтры — оставляем только по видимым полям
	visibleFieldIDs := make(map[string]bool, len(visibleFields))
	for _, f := range visibleFields {
		visibleFieldIDs[f.ID] = true
	}

	allowedFilters := make([]FilterCondition, 0, len(input.Filters))
	for _, filter := range input.Filters {
		if visibleFieldIDs[filter.FieldID] {
			allowedFilters = append(allowedFilters, filter)
		}
	}

	allowedSorts := make([]SortCondition, 0, len(input.Sorts))
	for _, sort := range input.Sorts {
		if visibleFieldIDs[sort.FieldID] {
			allowedSorts = append(allowedSorts, sort)
		}
	}

	// Добавляем row-level фильтры из enforcer
	rowFilters, err := s.enforcer.GetRowFilters(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get row filters: %w", err)
	}

	if input.Limit <= 0 {
		input.Limit = 50
	}

	params := QueryParams{
		Limit:      input.Limit,
		Offset:     input.Offset,
		Filters:    allowedFilters,
		Sorts:      allowedSorts,
		RowFilters: rowFilters,
	}

	rows, total, err := s.repo.List(ctx, spread.DBSchema, spread.TableName, visibleFields, params)
	if err != nil {
		return nil, fmt.Errorf("list rows: %w", err)
	}

	// Скрываем скрытые поля из результата
	sanitized := s.sanitizeRows(rows, visibleFields)

	return &QueryResult{
		Data:   sanitized,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, userID, spreadsheetID string, data RowData) (RowData, error) {
	if err := s.enforcer.RequireInsert(ctx, userID, spreadsheetID); err != nil {
		return nil, err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	// Получаем редактируемые поля
	editableFields, err := s.enforcer.EditableFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get editable fields: %w", err)
	}

	// Валидируем и фильтруем входящие данные
	cleanData, err := s.validateAndFilterData(ctx, spreadsheetID, data, editableFields)
	if err != nil {
		return nil, fmt.Errorf("validate data: %w", err)
	}

	row, err := s.repo.Create(ctx, spread.DBSchema, spread.TableName, cleanData, userID)
	if err != nil {
		return nil, fmt.Errorf("create row: %w", err)
	}

	return row, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *Service) Update(ctx context.Context, userID, spreadsheetID, rowID string, data RowData) (RowData, error) {
	if err := s.enforcer.RequireEdit(ctx, userID, spreadsheetID); err != nil {
		return nil, err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	// Получаем редактируемые поля
	editableFields, err := s.enforcer.EditableFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get editable fields: %w", err)
	}

	// Удаляем системные поля
	s.stripSystemFields(data)

	// Валидируем и фильтруем
	cleanData, err := s.validateAndFilterData(ctx, spreadsheetID, data, editableFields)
	if err != nil {
		return nil, fmt.Errorf("validate data: %w", err)
	}

	if len(cleanData) == 0 {
		return nil, fmt.Errorf("no editable fields in request")
	}

	row, err := s.repo.Update(ctx, spread.DBSchema, spread.TableName, rowID, cleanData)
	if err != nil {
		return nil, fmt.Errorf("update row: %w", err)
	}

	return row, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) Delete(ctx context.Context, userID, spreadsheetID, rowID string) error {
	if err := s.enforcer.RequireDelete(ctx, userID, spreadsheetID); err != nil {
		return err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	if err := s.repo.Delete(ctx, spread.DBSchema, spread.TableName, rowID); err != nil {
		return fmt.Errorf("delete row: %w", err)
	}

	return nil
}

// ─── BulkDelete ───────────────────────────────────────────────────────────────

type BulkDeleteInput struct {
	RowIDs []string `json:"row_ids"`
}

func (s *Service) BulkDelete(ctx context.Context, userID, spreadsheetID string, input BulkDeleteInput) error {
	if err := s.enforcer.RequireDelete(ctx, userID, spreadsheetID); err != nil {
		return err
	}

	if len(input.RowIDs) == 0 {
		return fmt.Errorf("row_ids is required")
	}

	if len(input.RowIDs) > 500 {
		return fmt.Errorf("cannot delete more than 500 rows at once")
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	if err := s.repo.BulkDelete(ctx, spread.DBSchema, spread.TableName, input.RowIDs); err != nil {
		return fmt.Errorf("bulk delete rows: %w", err)
	}

	return nil
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// validateAndFilterData проверяет данные на соответствие полям и правам доступа
func (s *Service) validateAndFilterData(
	ctx context.Context,
	spreadsheetID string,
	data RowData,
	editableFields []*fieldPkg.Field,
) (RowData, error) {
	// Строим карту column_name → field для быстрого поиска
	editableByColumn := make(map[string]*fieldPkg.Field, len(editableFields))
	for _, f := range editableFields {
		editableByColumn[f.ColumnName] = f
	}

	// Получаем все поля для валидации обязательных
	allFields, err := s.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	clean := make(RowData, len(data))

	// Фильтруем: оставляем только редактируемые поля
	for colName, value := range data {
		f, ok := editableByColumn[colName]
		if !ok {
			// Колонка не найдена среди редактируемых — пропускаем
			continue
		}

		// Валидируем значение
		if err := validateFieldValue(f, value); err != nil {
			return nil, fmt.Errorf("field %q: %w", f.Name, err)
		}

		clean[colName] = value
	}

	// Проверяем обязательные поля при создании (если данные не пустые)
	for _, f := range allFields {
		if !f.IsRequired {
			continue
		}
		val, provided := data[f.ColumnName]
		if !provided {
			continue
		}
		if isEmptyValue(val) {
			return nil, fmt.Errorf("field %q is required", f.Name)
		}
	}

	return clean, nil
}

// validateFieldValue проверяет значение поля на соответствие его типу
func validateFieldValue(f *fieldPkg.Field, value interface{}) error {
	if value == nil {
		if f.IsRequired {
			return fmt.Errorf("required field cannot be null")
		}
		return nil
	}

	switch f.FieldType {
	case "integer":
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer, got decimal")
			}
		case int, int32, int64:
			// ok
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}

	case "decimal":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// ok
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case "select":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for select, got %T", value)
		}
		if f.Options != nil {
			if err := validateSelectChoice(f, strVal); err != nil {
				return err
			}
		}

	case "multi_select":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for multi_select, got %T", value)
		}
		if f.Options != nil {
			for _, item := range arr {
				strItem, ok := item.(string)
				if !ok {
					return fmt.Errorf("multi_select values must be strings")
				}
				if err := validateSelectChoice(f, strItem); err != nil {
					return err
				}
			}
		}

	case "email":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for email")
		}
		if strVal != "" && !strings.Contains(strVal, "@") {
			return fmt.Errorf("invalid email format")
		}

	case "url":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for url")
		}
		if strVal != "" && !strings.HasPrefix(strVal, "http://") && !strings.HasPrefix(strVal, "https://") {
			return fmt.Errorf("url must start with http:// or https://")
		}
	}

	return nil
}

// validateSelectChoice проверяет, что значение входит в список допустимых вариантов
func validateSelectChoice(f *fieldPkg.Field, value string) error {
	choices, ok := f.Options["choices"].([]interface{})
	if !ok {
		return nil
	}

	for _, c := range choices {
		choice, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := choice["value"].(string); ok && v == value {
			return nil
		}
	}

	return fmt.Errorf("value %q is not in allowed choices", value)
}

// stripSystemFields удаляет системные поля из данных запроса
func (s *Service) stripSystemFields(data RowData) {
	delete(data, "_id")
	delete(data, "_created_by")
	delete(data, "_created_at")
	delete(data, "_updated_at")
	delete(data, "_position")
}

// sanitizeRows оставляет в строках только видимые поля плюс системные
func (s *Service) sanitizeRows(rows []RowData, visibleFields []*fieldPkg.Field) []RowData {
	if len(visibleFields) == 0 {
		return rows
	}

	// Строим set видимых column_name
	visibleCols := make(map[string]bool, len(visibleFields)+5)
	for _, f := range visibleFields {
		visibleCols[f.ColumnName] = true
	}

	// Системные поля всегда видны
	systemCols := []string{"_id", "_created_by", "_created_at", "_updated_at", "_position"}
	for _, col := range systemCols {
		visibleCols[col] = true
	}

	sanitized := make([]RowData, 0, len(rows))
	for _, row := range rows {
		clean := make(RowData, len(visibleCols))
		for col, val := range row {
			if visibleCols[col] {
				clean[col] = val
			}
		}
		sanitized = append(sanitized, clean)
	}

	return sanitized
}

// isEmptyValue проверяет, является ли значение пустым
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) == ""
	}
	return false
}
```

---

Также нужно обновить `internal/row/repository.go` — добавить метод `BulkDelete` и поле `RowFilters` в `QueryParams`:

## Дополнение к `internal/row/repository.go`

```go
// Добавить в структуру QueryParams поле RowFilters:

type QueryParams struct {
	Limit      int
	Offset     int
	Filters    []FilterCondition
	Sorts      []SortCondition
	RowFilters []permissionPkg.RowFilter  // фильтры row-level security
}

// Добавить метод BulkDelete в Repository:

func (r *Repository) BulkDelete(ctx context.Context, schema, table string, rowIDs []string) error {
	if len(rowIDs) == 0 {
		return nil
	}

	// Строим $1, $2, ... $N
	placeholders := make([]string, len(rowIDs))
	args := make([]interface{}, len(rowIDs))
	for i, id := range rowIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	sql := fmt.Sprintf(
		`DELETE FROM "%s"."%s" WHERE _id IN (%s)`,
		schema, table,
		strings.Join(placeholders, ", "),
	)

	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("bulk delete rows: %w", err)
	}

	return nil
}
```

---

И обновить `internal/row/handler.go` — заменить прямое использование репозитория на сервис:

## `internal/row/handler.go` (обновлённый)

```go
package row

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/datagrid/internal/middleware"
	permissionPkg "github.com/yourorg/datagrid/internal/permission"
	"github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type queryRequest struct {
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	Filters []FilterCondition `json:"filters"`
	Sorts   []SortCondition   `json:"sorts"`
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = queryRequest{Limit: 50}
	}

	result, err := h.service.Query(r.Context(), userID, spreadsheetID, QueryInput{
		Limit:   req.Limit,
		Offset:  req.Offset,
		Filters: req.Filters,
		Sorts:   req.Sorts,
	})
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var data RowData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		data = RowData{}
	}

	row, err := h.service.Create(r.Context(), userID, spreadsheetID, data)
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, row)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	rowID := chi.URLParam(r, "rowId")
	userID := middleware.GetUserID(r.Context())

	var data RowData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	row, err := h.service.Update(r.Context(), userID, spreadsheetID, rowID, data)
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, row)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	rowID := chi.URLParam(r, "rowId")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.Delete(r.Context(), userID, spreadsheetID, rowID); err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

func (h *Handler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var input BulkDeleteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.BulkDelete(r.Context(), userID, spreadsheetID, input); err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}
```

**User - --**

не хватает файла internal-permission-service.go

**Assistant - --**

## `internal/permission/service.go`

```go
package permission

import (
	"context"
	"fmt"

	spreadsheetPkg "github.com/yourorg/datagrid/internal/spreadsheet"
	workspacePkg "github.com/yourorg/datagrid/internal/workspace"
)

type Service struct {
	repo        *Repository
	spreadRepo  *spreadsheetPkg.Repository
	wsRepo      *workspacePkg.Repository
	enforcer    *Enforcer
}

func NewService(
	repo *Repository,
	spreadRepo *spreadsheetPkg.Repository,
	wsRepo *workspacePkg.Repository,
	enforcer *Enforcer,
) *Service {
	return &Service{
		repo:       repo,
		spreadRepo: spreadRepo,
		wsRepo:     wsRepo,
		enforcer:   enforcer,
	}
}

// ─── Spreadsheet Access ───────────────────────────────────────────────────────

type UpsertSpreadsheetAccessInput struct {
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	CanView       bool   `json:"can_view"`
	CanInsert     bool   `json:"can_insert"`
	CanEdit       bool   `json:"can_edit"`
	CanDelete     bool   `json:"can_delete"`
	CanManage     bool   `json:"can_manage"`
}

type RemoveSpreadsheetAccessInput struct {
	PrincipalID string `json:"principal_id"`
}

// GetSpreadsheetAccess возвращает все права доступа к таблице.
// Только пользователи с правом can_manage могут просматривать список прав.
func (s *Service) GetSpreadsheetAccess(ctx context.Context, requesterID, spreadsheetID string) ([]*SpreadsheetAccess, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	accesses, err := s.repo.GetSpreadsheetAccess(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet access: %w", err)
	}

	return accesses, nil
}

// UpsertSpreadsheetAccess создаёт или обновляет права доступа к таблице.
// Только пользователи с правом can_manage могут изменять права.
func (s *Service) UpsertSpreadsheetAccess(
	ctx context.Context,
	requesterID, spreadsheetID string,
	input UpsertSpreadsheetAccessInput,
) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	if input.PrincipalID == "" {
		return fmt.Errorf("principal_id is required")
	}

	if input.PrincipalType == "" {
		return fmt.Errorf("principal_type is required")
	}

	if input.PrincipalType != "user" && input.PrincipalType != "workspace_role" {
		return fmt.Errorf("principal_type must be 'user' or 'workspace_role'")
	}

	// Нельзя изменить права владельца workspace
	isOwner, err := s.isWorkspaceOwner(ctx, input.PrincipalID, spreadsheetID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if isOwner {
		return fmt.Errorf("cannot modify owner permissions")
	}

	access := &SpreadsheetAccess{
		SpreadsheetID: spreadsheetID,
		PrincipalID:   input.PrincipalID,
		PrincipalType: input.PrincipalType,
		CanView:       input.CanView,
		CanInsert:     input.CanInsert,
		CanEdit:       input.CanEdit,
		CanDelete:     input.CanDelete,
		CanManage:     input.CanManage,
	}

	if err := s.repo.UpsertSpreadsheetAccess(ctx, access); err != nil {
		return fmt.Errorf("upsert spreadsheet access: %w", err)
	}

	return nil
}

// RemoveSpreadsheetAccess удаляет права доступа пользователя к таблице.
// Только пользователи с правом can_manage могут удалять права.
func (s *Service) RemoveSpreadsheetAccess(
	ctx context.Context,
	requesterID, spreadsheetID, principalID string,
) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	// Нельзя удалить права владельца workspace
	isOwner, err := s.isWorkspaceOwner(ctx, principalID, spreadsheetID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if isOwner {
		return fmt.Errorf("cannot remove owner permissions")
	}

	if err := s.repo.RemoveSpreadsheetAccess(ctx, spreadsheetID, principalID); err != nil {
		return fmt.Errorf("remove spreadsheet access: %w", err)
	}

	return nil
}

// ─── Field Access ─────────────────────────────────────────────────────────────

type UpsertFieldAccessInput struct {
	FieldID       string `json:"field_id"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	CanView       bool   `json:"can_view"`
	CanEdit       bool   `json:"can_edit"`
}

// GetFieldAccess возвращает права доступа ко всем полям таблицы.
func (s *Service) GetFieldAccess(ctx context.Context, requesterID, spreadsheetID string) ([]*FieldAccess, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	accesses, err := s.repo.GetFieldAccess(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get field access: %w", err)
	}

	return accesses, nil
}

// UpsertFieldAccess создаёт или обновляет права доступа к полю.
func (s *Service) UpsertFieldAccess(
	ctx context.Context,
	requesterID, spreadsheetID string,
	input UpsertFieldAccessInput,
) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	if input.FieldID == "" {
		return fmt.Errorf("field_id is required")
	}

	if input.PrincipalID == "" {
		return fmt.Errorf("principal_id is required")
	}

	if input.PrincipalType == "" {
		return fmt.Errorf("principal_type is required")
	}

	if input.PrincipalType != "user" && input.PrincipalType != "workspace_role" {
		return fmt.Errorf("principal_type must be 'user' or 'workspace_role'")
	}

	fa := &FieldAccess{
		FieldID:       input.FieldID,
		PrincipalID:   input.PrincipalID,
		PrincipalType: input.PrincipalType,
		CanView:       input.CanView,
		CanEdit:       input.CanEdit,
	}

	if err := s.repo.UpsertFieldAccess(ctx, fa); err != nil {
		return fmt.Errorf("upsert field access: %w", err)
	}

	return nil
}

// ─── Row Access Rules ─────────────────────────────────────────────────────────

type UpsertRowRuleInput struct {
	PrincipalID   string                 `json:"principal_id"`
	PrincipalType string                 `json:"principal_type"`
	Condition     map[string]interface{} `json:"condition"`
	CanView       bool                   `json:"can_view"`
	CanEdit       bool                   `json:"can_edit"`
}

type RowAccessRule struct {
	ID            string                 `json:"id"`
	SpreadsheetID string                 `json:"spreadsheet_id"`
	PrincipalID   string                 `json:"principal_id"`
	PrincipalType string                 `json:"principal_type"`
	Condition     map[string]interface{} `json:"condition"`
	CanView       bool                   `json:"can_view"`
	CanEdit       bool                   `json:"can_edit"`
}

// GetRowRules возвращает все row-level правила для таблицы.
func (s *Service) GetRowRules(ctx context.Context, requesterID, spreadsheetID string) ([]*RowAccessRule, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	rules, err := s.repo.GetRowRules(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}

	return rules, nil
}

// UpsertRowRule создаёт или обновляет row-level правило.
func (s *Service) UpsertRowRule(
	ctx context.Context,
	requesterID, spreadsheetID string,
	input UpsertRowRuleInput,
) (*RowAccessRule, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	if input.PrincipalID == "" {
		return nil, fmt.Errorf("principal_id is required")
	}

	if input.PrincipalType != "user" && input.PrincipalType != "workspace_role" {
		return nil, fmt.Errorf("principal_type must be 'user' or 'workspace_role'")
	}

	if len(input.Condition) == 0 {
		return nil, fmt.Errorf("condition is required")
	}

	// Валидируем condition
	if err := validateRowCondition(input.Condition); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	rule := &RowAccessRule{
		SpreadsheetID: spreadsheetID,
		PrincipalID:   input.PrincipalID,
		PrincipalType: input.PrincipalType,
		Condition:     input.Condition,
		CanView:       input.CanView,
		CanEdit:       input.CanEdit,
	}

	created, err := s.repo.UpsertRowRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("upsert row rule: %w", err)
	}

	return created, nil
}

// DeleteRowRule удаляет row-level правило по ID.
func (s *Service) DeleteRowRule(ctx context.Context, requesterID, spreadsheetID, ruleID string) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	if err := s.repo.DeleteRowRule(ctx, ruleID); err != nil {
		return fmt.Errorf("delete row rule: %w", err)
	}

	return nil
}

// ─── My Permissions ───────────────────────────────────────────────────────────

type MyPermissions struct {
	Spreadsheet *SpreadsheetPerms        `json:"spreadsheet"`
	Fields      map[string]*FieldPerms   `json:"fields"`
}

// GetMyPermissions возвращает все права текущего пользователя на таблицу.
// Используется фронтендом для отображения доступных действий.
func (s *Service) GetMyPermissions(ctx context.Context, userID, spreadsheetID string) (*MyPermissions, error) {
	spreadPerms, err := s.enforcer.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet perms: %w", err)
	}

	// Если нет права на просмотр — возвращаем только это
	if !spreadPerms.CanView {
		return &MyPermissions{
			Spreadsheet: spreadPerms,
			Fields:      map[string]*FieldPerms{},
		}, nil
	}

	// Загружаем права по полям
	visibleFields, err := s.enforcer.VisibleFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get visible fields: %w", err)
	}

	fieldPerms := make(map[string]*FieldPerms, len(visibleFields))
	for _, f := range visibleFields {
		fp, err := s.enforcer.GetFieldPerms(ctx, userID, f.ID)
		if err != nil {
			continue
		}
		fieldPerms[f.ID] = fp
	}

	return &MyPermissions{
		Spreadsheet: spreadPerms,
		Fields:      fieldPerms,
	}, nil
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// isWorkspaceOwner проверяет, является ли пользователь владельцем workspace таблицы
func (s *Service) isWorkspaceOwner(ctx context.Context, userID, spreadsheetID string) (bool, error) {
	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return false, fmt.Errorf("get spreadsheet: %w", err)
	}

	ws, err := s.wsRepo.GetByID(ctx, spread.WorkspaceID)
	if err != nil {
		return false, fmt.Errorf("get workspace: %w", err)
	}

	return ws.OwnerID == userID, nil
}

// validateRowCondition проверяет корректность JSON-условия row-level правила
func validateRowCondition(condition map[string]interface{}) error {
	op, ok := condition["op"].(string)
	if !ok || op == "" {
		return fmt.Errorf("condition must have 'op' field")
	}

	validOps := map[string]bool{
		"eq":               true,
		"neq":              true,
		"contains":         true,
		"eq_current_user":  true,
	}

	if !validOps[op] {
		return fmt.Errorf("unknown op %q, valid ops: eq, neq, contains, eq_current_user", op)
	}

	if _, ok := condition["column_name"].(string); !ok {
		return fmt.Errorf("condition must have 'column_name' field")
	}

	// Для операций с явным значением требуем поле value
	if op == "eq" || op == "neq" || op == "contains" {
		if _, ok := condition["value"]; !ok {
			return fmt.Errorf("condition with op %q must have 'value' field", op)
		}
	}

	return nil
}
```

---

Также нужно добавить в `internal/permission/repository.go` методы для работы с row rules и обновить `internal/permission/handler.go`:

## Дополнение к `internal/permission/repository.go`

```go
// Добавить следующие методы в Repository:

func (r *Repository) GetRowRules(ctx context.Context, spreadsheetID string) ([]*RowAccessRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, spreadsheet_id, principal_id, principal_type,
		       condition, can_view, can_edit
		FROM meta.row_access_rules
		WHERE spreadsheet_id = $1
		ORDER BY id
	`, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}
	defer rows.Close()

	var result []*RowAccessRule
	for rows.Next() {
		rule := &RowAccessRule{}
		var condJSON []byte
		if err := rows.Scan(
			&rule.ID, &rule.SpreadsheetID, &rule.PrincipalID,
			&rule.PrincipalType, &condJSON, &rule.CanView, &rule.CanEdit,
		); err != nil {
			return nil, err
		}
		if condJSON != nil {
			_ = json.Unmarshal(condJSON, &rule.Condition)
		}
		result = append(result, rule)
	}
	return result, rows.Err()
}

func (r *Repository) UpsertRowRule(ctx context.Context, rule *RowAccessRule) (*RowAccessRule, error) {
	condJSON, err := json.Marshal(rule.Condition)
	if err != nil {
		return nil, fmt.Errorf("marshal condition: %w", err)
	}

	result := &RowAccessRule{}
	var condRaw []byte

	err = r.pool.QueryRow(ctx, `
		INSERT INTO meta.row_access_rules
		    (spreadsheet_id, principal_id, principal_type, condition, can_view, can_edit)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (spreadsheet_id, principal_id, principal_type)
		DO UPDATE SET
		    condition = $4,
		    can_view  = $5,
		    can_edit  = $6
		RETURNING id, spreadsheet_id, principal_id, principal_type,
		          condition, can_view, can_edit
	`, rule.SpreadsheetID, rule.PrincipalID, rule.PrincipalType,
		condJSON, rule.CanView, rule.CanEdit,
	).Scan(
		&result.ID, &result.SpreadsheetID, &result.PrincipalID,
		&result.PrincipalType, &condRaw, &result.CanView, &result.CanEdit,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert row rule: %w", err)
	}

	if condRaw != nil {
		_ = json.Unmarshal(condRaw, &result.Condition)
	}

	return result, nil
}

func (r *Repository) DeleteRowRule(ctx context.Context, ruleID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM meta.row_access_rules WHERE id = $1
	`, ruleID)
	return err
}
```

---

## `internal/permission/handler.go` (обновлённый)

```go
package permission

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/datagrid/internal/middleware"
	"github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Spreadsheet Access ───────────────────────────────────────────────────────

func (h *Handler) GetSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	accesses, err := h.service.GetSpreadsheetAccess(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, accesses)
}

func (h *Handler) UpsertSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var input UpsertSpreadsheetAccessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.UpsertSpreadsheetAccess(r.Context(), userID, spreadsheetID, input); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}

func (h *Handler) RemoveSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	principalID := chi.URLParam(r, "principalId")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.RemoveSpreadsheetAccess(r.Context(), userID, spreadsheetID, principalID); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Field Access ─────────────────────────────────────────────────────────────

func (h *Handler) GetFieldAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	accesses, err := h.service.GetFieldAccess(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, accesses)
}

func (h *Handler) UpsertFieldAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	fieldID := chi.URLParam(r, "fieldId")
	userID := middleware.GetUserID(r.Context())

	var input UpsertFieldAccessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	input.FieldID = fieldID

	if err := h.service.UpsertFieldAccess(r.Context(), userID, spreadsheetID, input); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Row Rules ────────────────────────────────────────────────────────────────

func (h *Handler) GetRowRules(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	rules, err := h.service.GetRowRules(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, rules)
}

func (h *Handler) UpsertRowRule(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var input UpsertRowRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	rule, err := h.service.UpsertRowRule(r.Context(), userID, spreadsheetID, input)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, rule)
}

func (h *Handler) DeleteRowRule(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	ruleID := chi.URLParam(r, "ruleId")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.DeleteRowRule(r.Context(), userID, spreadsheetID, ruleID); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── My Permissions ───────────────────────────────────────────────────────────

func (h *Handler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	perms, err := h.service.GetMyPermissions(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, perms)
}
```

---

И обновить роуты в `cmd/server/main.go` — добавить новые эндпоинты:

## Дополнение к роутам в `cmd/server/main.go`

```go
// Заменить блок Permissions на расширенный:

// Permissions
r.Get("/spreadsheets/{id}/permissions",
    permHandler.GetSpreadsheetAccess)
r.Put("/spreadsheets/{id}/permissions",
    permHandler.UpsertSpreadsheetAccess)
r.Delete("/spreadsheets/{id}/permissions/{principalId}",
    permHandler.RemoveSpreadsheetAccess)

// Field permissions
r.Get("/spreadsheets/{id}/field-permissions",
    permHandler.GetFieldAccess)
r.Put("/spreadsheets/{id}/field-permissions/{fieldId}",
    permHandler.UpsertFieldAccess)

// Row rules
r.Get("/spreadsheets/{id}/row-rules",
    permHandler.GetRowRules)
r.Put("/spreadsheets/{id}/row-rules",
    permHandler.UpsertRowRule)
r.Delete("/spreadsheets/{id}/row-rules/{ruleId}",
    permHandler.DeleteRowRule)

// My permissions
r.Get("/spreadsheets/{id}/my-permissions",
    permHandler.GetMyPermissions)

// Rows — добавить bulk delete
r.Delete("/spreadsheets/{id}/rows",
    rowHandler.BulkDelete)
```

**User - --**

файлы internal-spreadsheet-service.go и internal- field-service.go ссылаются друг на друга

**Assistant - --**

Это классическая проблема циклических зависимостей в Go. Решение — вынести общие типы в отдельный пакет `internal/types`, а зависимости передавать через интерфейсы.

---

## Новая структура

```
internal/
├── types/
│   └── types.go          ← общие типы (Field, Spreadsheet и др.)
├── field/
│   ├── repository.go     ← работает с types.Field
│   ├── service.go        ← не импортирует spreadsheet
│   └── handler.go
├── spreadsheet/
│   ├── repository.go     ← работает с types.Spreadsheet
│   ├── service.go        ← не импортирует field напрямую
│   └── handler.go
```

---

## `internal/types/types.go`

```go
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
```

---

## `internal/field/repository.go`

```go
package field

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/datagrid/internal/types"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, f *types.Field) (*types.Field, error) {
	optionsJSON, _ := json.Marshal(f.Options)

	result := &types.Field{}
	var optRaw []byte

	err := tx.QueryRow(ctx, `
		INSERT INTO meta.fields
			(spreadsheet_id, name, column_name, field_type, position,
			 is_required, is_unique, default_value, options)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, spreadsheet_id, name, column_name, field_type,
		          position, is_required, is_unique, default_value, options
	`, f.SpreadsheetID, f.Name, f.ColumnName, f.FieldType,
		f.Position, f.IsRequired, f.IsUnique, f.DefaultValue, optionsJSON,
	).Scan(
		&result.ID, &result.SpreadsheetID, &result.Name, &result.ColumnName,
		&result.FieldType, &result.Position, &result.IsRequired, &result.IsUnique,
		&result.DefaultValue, &optRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("create field: %w", err)
	}

	if optRaw != nil {
		_ = json.Unmarshal(optRaw, &result.Options)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*types.Field, error) {
	f := &types.Field{}
	var optRaw []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, spreadsheet_id, name, column_name, field_type,
		       position, is_required, is_unique, default_value, options
		FROM meta.fields WHERE id = $1
	`, id).Scan(
		&f.ID, &f.SpreadsheetID, &f.Name, &f.ColumnName, &f.FieldType,
		&f.Position, &f.IsRequired, &f.IsUnique, &f.DefaultValue, &optRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("get field: %w", err)
	}

	if optRaw != nil {
		_ = json.Unmarshal(optRaw, &f.Options)
	}
	return f, nil
}

func (r *Repository) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*types.Field, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, spreadsheet_id, name, column_name, field_type,
		       position, is_required, is_unique, default_value, options
		FROM meta.fields
		WHERE spreadsheet_id = $1
		ORDER BY position
	`, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}
	defer rows.Close()

	var fields []*types.Field
	for rows.Next() {
		f := &types.Field{}
		var optRaw []byte
		if err := rows.Scan(
			&f.ID, &f.SpreadsheetID, &f.Name, &f.ColumnName, &f.FieldType,
			&f.Position, &f.IsRequired, &f.IsUnique, &f.DefaultValue, &optRaw,
		); err != nil {
			return nil, err
		}
		if optRaw != nil {
			_ = json.Unmarshal(optRaw, &f.Options)
		}
		fields = append(fields, f)
	}

	return fields, rows.Err()
}

func (r *Repository) Update(ctx context.Context, tx pgx.Tx, f *types.Field) (*types.Field, error) {
	optionsJSON, _ := json.Marshal(f.Options)
	result := &types.Field{}
	var optRaw []byte

	err := tx.QueryRow(ctx, `
		UPDATE meta.fields SET
			name          = $1,
			field_type    = $2,
			position      = $3,
			is_required   = $4,
			is_unique     = $5,
			default_value = $6,
			options       = $7
		WHERE id = $8
		RETURNING id, spreadsheet_id, name, column_name, field_type,
		          position, is_required, is_unique, default_value, options
	`, f.Name, f.FieldType, f.Position, f.IsRequired, f.IsUnique,
		f.DefaultValue, optionsJSON, f.ID,
	).Scan(
		&result.ID, &result.SpreadsheetID, &result.Name, &result.ColumnName,
		&result.FieldType, &result.Position, &result.IsRequired, &result.IsUnique,
		&result.DefaultValue, &optRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("update field: %w", err)
	}

	if optRaw != nil {
		_ = json.Unmarshal(optRaw, &result.Options)
	}
	return result, nil
}

func (r *Repository) Delete(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `DELETE FROM meta.fields WHERE id = $1`, id)
	return err
}

func (r *Repository) GetMaxPosition(ctx context.Context, spreadsheetID string) (int, error) {
	var pos int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1)
		FROM meta.fields
		WHERE spreadsheet_id = $1
	`, spreadsheetID).Scan(&pos)
	return pos, err
}
```

---

## `internal/field/service.go`

```go
package field

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourorg/datagrid/internal/ddl"
	"github.com/yourorg/datagrid/internal/types"
)

// SpreadsheetGetter — интерфейс для получения таблицы.
// Позволяет избежать циклической зависимости со spreadsheet-пакетом.
type SpreadsheetGetter interface {
	GetByID(ctx context.Context, id string) (*types.Spreadsheet, error)
}

type Service struct {
	repo        *Repository
	spreadGetter SpreadsheetGetter
	ddlExec     *ddl.Executor
}

func NewService(repo *Repository, spreadGetter SpreadsheetGetter, ddlExec *ddl.Executor) *Service {
	return &Service{
		repo:         repo,
		spreadGetter: spreadGetter,
		ddlExec:      ddlExec,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, spreadsheetID string, input types.CreateFieldInput) (*types.Field, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("field name is required")
	}
	if !ValidFieldType(input.FieldType) {
		return nil, fmt.Errorf("invalid field type: %s", input.FieldType)
	}
	if input.Options != nil {
		if err := ValidateOptions(input.FieldType, input.Options); err != nil {
			return nil, err
		}
	}

	spread, err := s.spreadGetter.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	maxPos, _ := s.repo.GetMaxPosition(ctx, spreadsheetID)
	columnName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	colDef := ddl.ColumnDef{
		Name:         columnName,
		FieldType:    input.FieldType,
		IsRequired:   input.IsRequired,
		IsUnique:     input.IsUnique,
		DefaultValue: input.DefaultValue,
		Options:      input.Options,
	}

	ddlQuery := ddl.BuildAddColumn(spread.DBSchema, spread.TableName, colDef)

	newField := &types.Field{
		SpreadsheetID: spreadsheetID,
		Name:          input.Name,
		ColumnName:    columnName,
		FieldType:     input.FieldType,
		Position:      maxPos + 1,
		IsRequired:    input.IsRequired,
		IsUnique:      input.IsUnique,
		DefaultValue:  input.DefaultValue,
		Options:       input.Options,
	}

	var created *types.Field

	err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, tx, newField)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("create field: %w", err)
	}

	return created, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *Service) Update(ctx context.Context, fieldID string, input types.UpdateFieldInput) (*types.Field, error) {
	existing, err := s.repo.GetByID(ctx, fieldID)
	if err != nil {
		return nil, fmt.Errorf("field not found: %w", err)
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.IsRequired != nil {
		existing.IsRequired = *input.IsRequired
	}
	if input.IsUnique != nil {
		existing.IsUnique = *input.IsUnique
	}
	if input.DefaultValue != nil {
		existing.DefaultValue = input.DefaultValue
	}
	if input.Options != nil {
		existing.Options = input.Options
	}

	// Если меняется тип — нужен ALTER COLUMN
	if input.FieldType != nil && *input.FieldType != existing.FieldType {
		if !IsConversionAllowed(existing.FieldType, *input.FieldType) {
			return nil, fmt.Errorf(
				"type conversion from '%s' to '%s' is not allowed",
				existing.FieldType, *input.FieldType,
			)
		}

		spread, err := s.spreadGetter.GetByID(ctx, existing.SpreadsheetID)
		if err != nil {
			return nil, fmt.Errorf("spreadsheet not found: %w", err)
		}

		oldType := existing.FieldType
		existing.FieldType = *input.FieldType

		colDef := ddl.ColumnDef{
			Name:      existing.ColumnName,
			FieldType: existing.FieldType,
			Options:   existing.Options,
		}

		ddlQuery := ddl.BuildAlterColumnType(spread.DBSchema, spread.TableName, colDef)

		var updated *types.Field
		err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
			var txErr error
			updated, txErr = s.repo.Update(ctx, tx, existing)
			return txErr
		})
		if err != nil {
			existing.FieldType = oldType
			return nil, fmt.Errorf("alter column type: %w", err)
		}

		return updated, nil
	}

	// Тип не меняется — только метаданные
	var updated *types.Field

	err = s.ddlExec.ExecInTx(ctx, "", func(tx pgx.Tx) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, tx, existing)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("update field metadata: %w", err)
	}

	return updated, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) Delete(ctx context.Context, fieldID string) error {
	f, err := s.repo.GetByID(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("field not found: %w", err)
	}

	spread, err := s.spreadGetter.GetByID(ctx, f.SpreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	ddlQuery := ddl.BuildDropColumn(spread.DBSchema, spread.TableName, f.ColumnName)

	return s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		return s.repo.Delete(ctx, tx, fieldID)
	})
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *Service) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*types.Field, error) {
	return s.repo.ListBySpreadsheet(ctx, spreadsheetID)
}
```

---

## `internal/spreadsheet/repository.go`

```go
package spreadsheet

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/datagrid/internal/types"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, s *types.Spreadsheet) (*types.Spreadsheet, error) {
	result := &types.Spreadsheet{}
	err := tx.QueryRow(ctx, `
		INSERT INTO meta.spreadsheets (workspace_id, name, table_name, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, name, table_name,
		          COALESCE(description,''), created_by, created_at, updated_at
	`, s.WorkspaceID, s.Name, s.TableName, s.Description, s.CreatedBy,
	).Scan(
		&result.ID, &result.WorkspaceID, &result.Name, &result.TableName,
		&result.Description, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create spreadsheet: %w", err)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*types.Spreadsheet, error) {
	s := &types.Spreadsheet{}
	err := r.pool.QueryRow(ctx, `
		SELECT s.id, s.workspace_id, s.name, s.table_name,
		       COALESCE(s.description,''), s.created_by,
		       s.created_at, s.updated_at,
		       w.db_schema
		FROM meta.spreadsheets s
		JOIN auth.workspaces w ON w.id = s.workspace_id
		WHERE s.id = $1
	`, id).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
		&s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		&s.DBSchema,
	)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}
	return s, nil
}

func (r *Repository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*types.Spreadsheet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, name, table_name,
		       COALESCE(description,''), created_by, created_at, updated_at
		FROM meta.spreadsheets
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list spreadsheets: %w", err)
	}
	defer rows.Close()

	var result []*types.Spreadsheet
	for rows.Next() {
		s := &types.Spreadsheet{}
		if err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
			&s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id, name, description string) (*types.Spreadsheet, error) {
	s := &types.Spreadsheet{}
	err := r.pool.QueryRow(ctx, `
		UPDATE meta.spreadsheets
		SET name = $1, description = $2, updated_at = now()
		WHERE id = $3
		RETURNING id, workspace_id, name, table_name,
		          COALESCE(description,''), created_by, created_at, updated_at
	`, name, description, id).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
		&s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *Repository) Delete(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `DELETE FROM meta.spreadsheets WHERE id = $1`, id)
	return err
}
```

---

## `internal/spreadsheet/service.go`

```go
package spreadsheet

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourorg/datagrid/internal/ddl"
	"github.com/yourorg/datagrid/internal/types"
)

// FieldRepository — интерфейс для работы с полями.
// Позволяет избежать циклической зависимости с field-пакетом.
type FieldRepository interface {
	Create(ctx context.Context, tx pgx.Tx, f *types.Field) (*types.Field, error)
	ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*types.Field, error)
}

// WorkspaceGetter — интерфейс для получения workspace.
type WorkspaceGetter interface {
	GetByID(ctx context.Context, id string) (*types.Workspace, error)
}

type Service struct {
	repo      *Repository
	fieldRepo FieldRepository
	wsGetter  WorkspaceGetter
	ddlExec   *ddl.Executor
}

func NewService(
	repo *Repository,
	fieldRepo FieldRepository,
	wsGetter WorkspaceGetter,
	ddlExec *ddl.Executor,
) *Service {
	return &Service{
		repo:      repo,
		fieldRepo: fieldRepo,
		wsGetter:  wsGetter,
		ddlExec:   ddlExec,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, input types.CreateSpreadsheetInput, createdBy string) (*types.SpreadsheetWithFields, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.WorkspaceID == "" {
		return nil, fmt.Errorf("workspace_id is required")
	}

	ws, err := s.wsGetter.GetByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}

	tableName := "tbl_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	// Готовим DDL-колонки и метаданные полей
	colDefs := make([]ddl.ColumnDef, 0, len(input.Fields))
	fieldRecords := make([]*types.Field, 0, len(input.Fields))

	for i, fi := range input.Fields {
		colName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")
		colDefs = append(colDefs, ddl.ColumnDef{
			Name:         colName,
			FieldType:    fi.FieldType,
			IsRequired:   fi.IsRequired,
			IsUnique:     fi.IsUnique,
			DefaultValue: fi.DefaultValue,
			Options:      fi.Options,
		})
		fieldRecords = append(fieldRecords, &types.Field{
			Name:         fi.Name,
			ColumnName:   colName,
			FieldType:    fi.FieldType,
			Position:     i,
			IsRequired:   fi.IsRequired,
			IsUnique:     fi.IsUnique,
			DefaultValue: fi.DefaultValue,
			Options:      fi.Options,
		})
	}

	ddlQuery := ddl.BuildCreateTable(ws.DBSchema, tableName, colDefs)

	spread := &types.Spreadsheet{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		TableName:   tableName,
		Description: input.Description,
		CreatedBy:   createdBy,
	}

	var result *types.SpreadsheetWithFields

	err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		created, err := s.repo.Create(ctx, tx, spread)
		if err != nil {
			return fmt.Errorf("create spreadsheet: %w", err)
		}

		createdFields := make([]*types.Field, 0, len(fieldRecords))
		for _, fr := range fieldRecords {
			fr.SpreadsheetID = created.ID
			cf, err := s.fieldRepo.Create(ctx, tx, fr)
			if err != nil {
				return fmt.Errorf("create field '%s': %w", fr.Name, err)
			}
			createdFields = append(createdFields, cf)
		}

		result = &types.SpreadsheetWithFields{
			Spreadsheet: created,
			Fields:      createdFields,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (s *Service) GetWithFields(ctx context.Context, id string) (*types.SpreadsheetWithFields, error) {
	spread, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	fields, err := s.fieldRepo.ListBySpreadsheet(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	return &types.SpreadsheetWithFields{
		Spreadsheet: spread,
		Fields:      fields,
	}, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string) ([]*types.SpreadsheetWithFields, error) {
	spreads, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list spreadsheets: %w", err)
	}

	result := make([]*types.SpreadsheetWithFields, 0, len(spreads))
	for _, sp := range spreads {
		fields, _ := s.fieldRepo.ListBySpreadsheet(ctx, sp.ID)
		result = append(result, &types.SpreadsheetWithFields{
			Spreadsheet: sp,
			Fields:      fields,
		})
	}
	return result, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *Service) Update(ctx context.Context, id, name, description string) (*types.Spreadsheet, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.repo.Update(ctx, id, name, description)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) Delete(ctx context.Context, id string) error {
	spread, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	dropQuery := fmt.Sprintf(
		`DROP TABLE IF EXISTS "%s"."%s"`,
		spread.DBSchema, spread.TableName,
	)

	return s.ddlExec.ExecInTx(ctx, dropQuery, func(tx pgx.Tx) error {
		return s.repo.Delete(ctx, tx, id)
	})
}
```

---

Также нужно добавить тип `Workspace` в `internal/types/types.go` и обновить `workspace/repository.go`:

## Дополнение к `internal/types/types.go`

```go
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
```

---

## Обновление `cmd/server/main.go` — wire зависимостей

```go
// Repositories
authRepo        := authPkg.NewRepository(pool)
wsRepo          := workspacePkg.NewRepository(pool)
spreadsheetRepo := spreadsheetPkg.NewRepository(pool)
fieldRepo       := fieldPkg.NewRepository(pool)
rowRepo         := rowPkg.NewRepository(pool)
permRepo        := permissionPkg.NewRepository(pool)

// field.Service получает интерфейс SpreadsheetGetter → spreadsheetRepo напрямую
fieldService := fieldPkg.NewService(fieldRepo, spreadsheetRepo, ddlExec)

// spreadsheet.Service получает интерфейс FieldRepository → fieldRepo напрямую
spreadsheetSvc := spreadsheetPkg.NewService(spreadsheetRepo, fieldRepo, wsRepo, ddlExec)
```

Оба сервиса теперь зависят только от **интерфейсов** и **репозиториев** — циклических импортов нет.

**User - --**

приведи полный код обновленного файла cmd/server/main.go

**Assistant - --**

## `cmd/server/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/yourorg/datagrid/config"
	authPkg "github.com/yourorg/datagrid/internal/auth"
	"github.com/yourorg/datagrid/internal/ddl"
	fieldPkg "github.com/yourorg/datagrid/internal/field"
	"github.com/yourorg/datagrid/internal/middleware"
	permissionPkg "github.com/yourorg/datagrid/internal/permission"
	rowPkg "github.com/yourorg/datagrid/internal/row"
	spreadsheetPkg "github.com/yourorg/datagrid/internal/spreadsheet"
	workspacePkg "github.com/yourorg/datagrid/internal/workspace"
	"github.com/yourorg/datagrid/pkg/jwt"
	"github.com/yourorg/datagrid/pkg/postgres"
)

func main() {
	cfg := config.Load()

	// ─── База данных ──────────────────────────────────────────────────────────
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	log.Println("connected to database")

	// ─── Инфраструктура ───────────────────────────────────────────────────────
	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)
	ddlExec    := ddl.NewExecutor(pool)

	// ─── Repositories ─────────────────────────────────────────────────────────
	authRepo        := authPkg.NewRepository(pool)
	wsRepo          := workspacePkg.NewRepository(pool)
	spreadsheetRepo := spreadsheetPkg.NewRepository(pool)
	fieldRepo       := fieldPkg.NewRepository(pool)
	rowRepo         := rowPkg.NewRepository(pool)
	permRepo        := permissionPkg.NewRepository(pool)

	// ─── Services ─────────────────────────────────────────────────────────────

	// field.Service использует интерфейс SpreadsheetGetter — передаём spreadsheetRepo
	fieldService := fieldPkg.NewService(
		fieldRepo,
		spreadsheetRepo, // реализует field.SpreadsheetGetter
		ddlExec,
	)

	// spreadsheet.Service использует интерфейсы FieldRepository и WorkspaceGetter
	spreadsheetSvc := spreadsheetPkg.NewService(
		spreadsheetRepo,
		fieldRepo,      // реализует spreadsheet.FieldRepository
		wsRepo,         // реализует spreadsheet.WorkspaceGetter
		ddlExec,
	)

	// workspace.Service
	wsService := workspacePkg.NewService(wsRepo, ddlExec)

	// auth.Service
	authService := authPkg.NewService(authRepo, jwtManager)

	// permission.Enforcer и permission.Service
	enforcer := permissionPkg.NewEnforcer(pool, permRepo, fieldRepo)

	permService := permissionPkg.NewService(
		permRepo,
		spreadsheetRepo,
		wsRepo,
		enforcer,
	)

	// row.Service
	rowService := rowPkg.NewService(
		rowRepo,
		spreadsheetRepo,
		fieldRepo,
		enforcer,
	)

	// ─── Handlers ─────────────────────────────────────────────────────────────
	authHandler        := authPkg.NewHandler(authService)
	wsHandler          := workspacePkg.NewHandler(wsService)
	spreadsheetHandler := spreadsheetPkg.NewHandler(spreadsheetSvc)
	fieldHandler       := fieldPkg.NewHandler(fieldService)
	rowHandler         := rowPkg.NewHandler(rowService)
	permHandler        := permissionPkg.NewHandler(permService)

	// ─── Router ───────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Timeout(30 * time.Second))
	r.Use(middleware.CORS())

	// ─── Public routes ────────────────────────────────────────────────────────
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login",    authHandler.Login)

	// ─── Protected routes ─────────────────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))

		// Auth
		r.Get("/auth/me", authHandler.Me)

		// ── Workspaces ────────────────────────────────────────────────────────
		r.Get("/workspaces",       wsHandler.List)
		r.Post("/workspaces",      wsHandler.Create)
		r.Get("/workspaces/{id}",  wsHandler.Get)

		// ── Spreadsheets ──────────────────────────────────────────────────────
		r.Post("/spreadsheets",                         spreadsheetHandler.Create)
		r.Get("/workspaces/{workspaceId}/spreadsheets", spreadsheetHandler.ListByWorkspace)
		r.Get("/spreadsheets/{id}",                     spreadsheetHandler.Get)
		r.Put("/spreadsheets/{id}",                     spreadsheetHandler.Update)
		r.Delete("/spreadsheets/{id}",                  spreadsheetHandler.Delete)

		// ── Fields ────────────────────────────────────────────────────────────
		r.Post("/spreadsheets/{id}/fields", fieldHandler.Create)
		r.Get("/spreadsheets/{id}/fields",  fieldHandler.ListBySpreadsheet)
		r.Put("/fields/{fieldId}",          fieldHandler.Update)
		r.Delete("/fields/{fieldId}",       fieldHandler.Delete)

		// ── Rows ──────────────────────────────────────────────────────────────
		r.Post("/spreadsheets/{id}/rows/query",          rowHandler.Query)
		r.Post("/spreadsheets/{id}/rows",                rowHandler.Create)
		r.Patch("/spreadsheets/{id}/rows/{rowId}",       rowHandler.Update)
		r.Delete("/spreadsheets/{id}/rows/{rowId}",      rowHandler.Delete)
		r.Delete("/spreadsheets/{id}/rows",              rowHandler.BulkDelete)

		// ── Permissions — spreadsheet level ───────────────────────────────────
		r.Get("/spreadsheets/{id}/permissions",                    permHandler.GetSpreadsheetAccess)
		r.Put("/spreadsheets/{id}/permissions",                    permHandler.UpsertSpreadsheetAccess)
		r.Delete("/spreadsheets/{id}/permissions/{principalId}",   permHandler.RemoveSpreadsheetAccess)

		// ── Permissions — field level ─────────────────────────────────────────
		r.Get("/spreadsheets/{id}/field-permissions",              permHandler.GetFieldAccess)
		r.Put("/spreadsheets/{id}/field-permissions/{fieldId}",    permHandler.UpsertFieldAccess)

		// ── Permissions — row rules ────────────────────────────────────────────
		r.Get("/spreadsheets/{id}/row-rules",                      permHandler.GetRowRules)
		r.Put("/spreadsheets/{id}/row-rules",                      permHandler.UpsertRowRule)
		r.Delete("/spreadsheets/{id}/row-rules/{ruleId}",          permHandler.DeleteRowRule)

		// ── Permissions — my permissions ──────────────────────────────────────
		r.Get("/spreadsheets/{id}/my-permissions",                 permHandler.GetMyPermissions)
	})

	// ─── HTTP Server ──────────────────────────────────────────────────────────
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ─── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server started on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server stopped")
}
```

**User - --**

приведи полный код обновленного internal/field/handler.go

**Assistant - --**

## `internal/field/handler.go`

```go
package field

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/datagrid/internal/types"
	"github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	var input types.CreateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "field name is required")
		return
	}

	if !ValidFieldType(input.FieldType) {
		response.BadRequest(w, "invalid field type: "+input.FieldType)
		return
	}

	if input.Options != nil {
		if err := ValidateOptions(input.FieldType, input.Options); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Create(r.Context(), spreadsheetID, input)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, field)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	var input types.UpdateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.FieldType != nil && !ValidFieldType(*input.FieldType) {
		response.BadRequest(w, "invalid field type: "+*input.FieldType)
		return
	}

	if input.FieldType != nil && input.Options != nil {
		if err := ValidateOptions(*input.FieldType, input.Options); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Update(r.Context(), fieldID, input)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, field)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	if err := h.service.Delete(r.Context(), fieldID); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (h *Handler) ListBySpreadsheet(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	fields, err := h.service.ListBySpreadsheet(r.Context(), spreadsheetID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, fields)
}
```

**User - --**

приведи полный код обновленного файла internal/spreadsheet/handler.go

**Assistant - --**

## `internal/spreadsheet/handler.go`

```go
package spreadsheet

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/datagrid/internal/middleware"
	"github.com/yourorg/datagrid/internal/types"
	"github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var input types.CreateSpreadsheetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	if input.WorkspaceID == "" {
		response.BadRequest(w, "workspace_id is required")
		return
	}

	// Валидируем поля если они переданы
	for i, f := range input.Fields {
		if f.Name == "" {
			response.BadRequest(w, fmt.Sprintf("field[%d]: name is required", i))
			return
		}
		if !isValidFieldType(f.FieldType) {
			response.BadRequest(w, fmt.Sprintf("field[%d]: invalid field type '%s'", i, f.FieldType))
			return
		}
		if f.Options != nil {
			if err := validateFieldOptions(f.FieldType, f.Options); err != nil {
				response.BadRequest(w, fmt.Sprintf("field[%d]: %s", i, err.Error()))
				return
			}
		}
	}

	result, err := h.service.Create(r.Context(), input, userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, result)
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.service.GetWithFields(r.Context(), id)
	if err != nil {
		response.NotFound(w, "spreadsheet not found")
		return
	}

	response.OK(w, result)
}

// ─── ListByWorkspace ──────────────────────────────────────────────────────────

func (h *Handler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceId")

	result, err := h.service.ListByWorkspace(r.Context(), workspaceID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	result, err := h.service.Update(r.Context(), id, input.Name, input.Description)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

// isValidFieldType проверяет допустимость типа поля.
// Дублирует логику field.ValidFieldType без импорта field-пакета.
func isValidFieldType(fieldType string) bool {
	switch fieldType {
	case "text", "integer", "decimal", "boolean",
		"date", "datetime", "select", "multi_select",
		"email", "url", "phone", "attachment":
		return true
	}
	return false
}

// validateFieldOptions проверяет опции поля при создании таблицы.
// Дублирует логику field.ValidateOptions без импорта field-пакета.
func validateFieldOptions(fieldType string, options map[string]interface{}) error {
	switch fieldType {
	case "select", "multi_select":
		choices, ok := options["choices"]
		if !ok {
			return fmt.Errorf("select field must have 'choices' option")
		}
		arr, ok := choices.([]interface{})
		if !ok || len(arr) == 0 {
			return fmt.Errorf("'choices' must be a non-empty array")
		}
		for i, c := range arr {
			choice, ok := c.(map[string]interface{})
			if !ok {
				return fmt.Errorf("choice[%d] must be an object", i)
			}
			if _, ok := choice["value"].(string); !ok {
				return fmt.Errorf("choice[%d] must have a 'value' string field", i)
			}
			if _, ok := choice["label"].(string); !ok {
				return fmt.Errorf("choice[%d] must have a 'label' string field", i)
			}
		}
	}
	return nil
}
```

Обратите внимание — в `Create` используется `fmt.Sprintf`, поэтому нужно добавить `"fmt"` в импорты. Финальная версия с правильными импортами:

```go
package spreadsheet

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/datagrid/internal/middleware"
	"github.com/yourorg/datagrid/internal/types"
	"github.com/yourorg/datagrid/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var input types.CreateSpreadsheetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	if input.WorkspaceID == "" {
		response.BadRequest(w, "workspace_id is required")
		return
	}

	for i, f := range input.Fields {
		if f.Name == "" {
			response.BadRequest(w, fmt.Sprintf("field[%d]: name is required", i))
			return
		}
		if !isValidFieldType(f.FieldType) {
			response.BadRequest(w, fmt.Sprintf("field[%d]: invalid field type '%s'", i, f.FieldType))
			return
		}
		if f.Options != nil {
			if err := validateFieldOptions(f.FieldType, f.Options); err != nil {
				response.BadRequest(w, fmt.Sprintf("field[%d]: %s", i, err.Error()))
				return
			}
		}
	}

	result, err := h.service.Create(r.Context(), input, userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, result)
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.service.GetWithFields(r.Context(), id)
	if err != nil {
		response.NotFound(w, "spreadsheet not found")
		return
	}

	response.OK(w, result)
}

// ─── ListByWorkspace ──────────────────────────────────────────────────────────

func (h *Handler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceId")

	result, err := h.service.ListByWorkspace(r.Context(), workspaceID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	result, err := h.service.Update(r.Context(), id, input.Name, input.Description)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

func isValidFieldType(fieldType string) bool {
	switch fieldType {
	case "text", "integer", "decimal", "boolean",
		"date", "datetime", "select", "multi_select",
		"email", "url", "phone", "attachment":
		return true
	}
	return false
}

func validateFieldOptions(fieldType string, options map[string]interface{}) error {
	switch fieldType {
	case "select", "multi_select":
		choices, ok := options["choices"]
		if !ok {
			return fmt.Errorf("select field must have 'choices' option")
		}
		arr, ok := choices.([]interface{})
		if !ok || len(arr) == 0 {
			return fmt.Errorf("'choices' must be a non-empty array")
		}
		for i, c := range arr {
			choice, ok := c.(map[string]interface{})
			if !ok {
				return fmt.Errorf("choice[%d] must be an object", i)
			}
			if _, ok := choice["value"].(string); !ok {
				return fmt.Errorf("choice[%d] must have a 'value' string field", i)
			}
			if _, ok := choice["label"].(string); !ok {
				return fmt.Errorf("choice[%d] must have a 'label' string field", i)
			}
		}
	}
	return nil
}
```

**User - --**

напиши команды запуска приложения

**Assistant - --**

## Команды запуска приложения

### 1. Переменные окружения

Создай файл `.env` в корне проекта:

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=datagrid
DB_PASSWORD=secret
DB_NAME=datagrid
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-super-secret-key-min-32-chars
JWT_ACCESS_TOKEN_TTL=24h
```

---

### 2. Makefile

```makefile
.PHONY: run build test migrate-up migrate-down docker-up docker-down

# ─── Запуск ───────────────────────────────────────────────────────────────────

run:
	go run ./cmd/server/main.go

run-env:
	env $(cat .env | xargs) go run ./cmd/server/main.go

build:
	go build -o bin/datagrid ./cmd/server/main.go

build-run: build
	./bin/datagrid

# ─── Зависимости ─────────────────────────────────────────────────────────────

deps:
	go mod tidy
	go mod download

# ─── Миграции ─────────────────────────────────────────────────────────────────

migrate-up:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

migrate-drop:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" drop -f

# ─── Docker ───────────────────────────────────────────────────────────────────

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-rebuild:
	docker compose up -d --build

docker-logs:
	docker compose logs -f app

# ─── Тесты ───────────────────────────────────────────────────────────────────

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ─── Линтер ──────────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./...
```

---

### 3. `docker-compose.yml`

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    env_file:
      - .env
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: datagrid
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: datagrid
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U datagrid"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  postgres_data:
```

---

### 4. `Dockerfile`

```dockerfile
# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/datagrid ./cmd/server/main.go

# ── Run stage ─────────────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/datagrid .

EXPOSE 8080

CMD ["./datagrid"]
```

---

### 5. `config/config.go`

```go
package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	JWT    JWTConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret         string
	AccessTokenTTL time.Duration
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "datagrid"),
			Password: getEnv("DB_PASSWORD", "secret"),
			Name:     getEnv("DB_NAME", "datagrid"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "changeme-min-32-chars-secret-key"),
			AccessTokenTTL: getDuration("JWT_ACCESS_TOKEN_TTL", 24*time.Hour),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}
```

---

### 6. Последовательность первого запуска

```bash
# 1. Установить зависимости
go mod tidy

# 2. Поднять PostgreSQL
make docker-up

# 3. Применить миграции
make migrate-up

# 4. Запустить приложение локально
make run

# ── или всё через Docker ──────────────────────────────────────────────────────

# Поднять всё сразу (postgres + app)
make docker-rebuild

# Смотреть логи
make docker-logs
```

---

### 7. Проверка работоспособности

```bash
# Регистрация
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123","name":"Admin"}'

# Логин
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'

# Создание workspace (подставить токен из логина)
curl -X POST http://localhost:8080/workspaces \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"My Workspace"}'

# Создание таблицы
curl -X POST http://localhost:8080/spreadsheets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "workspace_id": "<workspace_id>",
    "name": "Contacts",
    "fields": [
      {"name":"Name",  "field_type":"text",    "is_required":true},
      {"name":"Email", "field_type":"email",   "is_required":true},
      {"name":"Age",   "field_type":"integer", "is_required":false}
    ]
  }'
```

**User - --**

при запуске веб части открывается пустая страница

**Assistant - --**

Нужно посмотреть что именно происходит. Дай ответы на несколько вопросов:

**1. Какой фреймворк используется на фронтенде?**
- React / Next.js / Vue / Svelte / другой?

**2. Какие ошибки в консоли браузера?**
```
F12 → Console → скопируй ошибки
```

**3. Какие ошибки в Network?**
```
F12 → Network → посмотри на красные запросы
```

**4. Как запускается фронтенд?**
```bash
# npm?
npm run dev
npm run start

# или через Docker?
docker compose up
```

**5. Что в терминале при запуске?**
```bash
# скопируй вывод команды запуска
```

---

Пока жду ответы — вот **самые частые причины** пустой страницы:

| Причина | Симптом в консоли |
|---|---|
| Ошибка сборки JS | `SyntaxError` / `ReferenceError` |
| Не найден роут | Пустой `<div id="root">` в Elements |
| CORS ошибка | `Access-Control-Allow-Origin` |
| API недоступен | `Failed to fetch` / `net::ERR_CONNECTION_REFUSED` |
| Env переменные не заданы | `Cannot read properties of undefined` |
| Неверный `base path` | 404 на JS/CSS файлы в Network |

Скопируй ошибки из консоли — сразу скажу в чём проблема.

**User - --**

framework: react
команда запуска:  npm run dev
package.json:
{
  "name": "datagrid-frontend",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev":     "vite",
    "build":   "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.0.0",
    "@tanstack/react-virtual": "^3.0.0",
    "axios": "^1.6.0",
    "clsx": "^2.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "zustand": "^4.4.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}

**Assistant - --**

Покажи содержимое следующих файлов:

```
src/main.tsx
src/App.tsx
```

И структуру папки `src`:
```bash
find src -type f | sort
```

А также — есть ли файлы конфигурации в корне:
```bash
ls -la
# ожидаем увидеть:
# vite.config.ts
# tsconfig.json
# tailwind.config.js
# postcss.config.js
# index.html
```

И содержимое `index.html`:
```bash
cat index.html
```

---

Пока жду — **самые частые причины** пустой страницы в Vite + React:

| Файл | Типичная проблема |
|---|---|
| `index.html` | Нет `<div id="root">` или неверный путь к `main.tsx` |
| `main.tsx` | `document.getElementById('root')` возвращает `null` |
| `App.tsx` | Компонент ничего не рендерит / падает с ошибкой |
| `vite.config.ts` | Неверный `base` путь |
| `tailwind.config.js` | Не настроен `content` — стили не применяются |

Скопируй содержимое этих файлов и сразу найдём проблему.

**User - --**

main.tsx:
import React from 'react';
import ReactDOM from 'react-dom/client';
import { App } from './App';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

App.tsx:
import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAuthStore } from './store/authStore';
import { LoginPage } from './pages/LoginPage';
import { WorkspacesPage } from './pages/WorkspacesPage';
import { SpreadsheetPage } from './pages/SpreadsheetPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthStore();
  return isAuthenticated() ? <>{children}</> : <Navigate to="/login" replace />;
};

export const App: React.FC = () => (
  <QueryClientProvider client={queryClient}>
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/workspaces" element={<PrivateRoute><WorkspacesPage /></PrivateRoute>} />
        <Route path="/spreadsheet/:id" element={<PrivateRoute><SpreadsheetPage /></PrivateRoute>} />
        <Route path="*" element={<Navigate to="/workspaces" replace />} />
      </Routes>
    </BrowserRouter>
  </QueryClientProvider>
);

index.html:
<!DOCTYPE html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>DataGrid</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>

**Assistant - --**

`index.html` и `main.tsx` выглядят корректно. Проблема скорее всего в одном из импортируемых файлов. Покажи:

```bash
# 1. Структура src/
find src -type f | sort

# 2. Содержимое этих файлов:
cat src/index.css
cat src/store/authStore.ts
cat src/pages/LoginPage.tsx
cat src/pages/WorkspacesPage.tsx
cat src/pages/SpreadsheetPage.tsx
```

И конфиги:

```bash
cat vite.config.ts
cat tailwind.config.js
cat postcss.config.js
cat tsconfig.json
```

---

А пока — **проверь консоль браузера прямо сейчас**:

```
F12 → Console
```

Скорее всего там одна из этих ошибок:

| Ошибка в консоли | Причина |
|---|---|
| `Failed to resolve import` | Не найден один из импортируемых файлов |
| `useAuthStore is not a function` | Ошибка в `authStore.ts` |
| `Cannot read properties of null` | `getElementById('root')` вернул null |
| `Unexpected token` | Синтаксическая ошибка в одном из файлов |
| `tailwindcss` / `postcss` ошибка | Неверный конфиг postcss |

Скопируй **текст ошибки из консоли** — это быстрее всего укажет на причину.

**User - --**

Uncaught SyntaxError: The requested module '/src/components/ui/Dropdown.tsx' does not provide an export named 'Dropdown' (at HeaderCell.tsx:3:10)

**Assistant - --**

Покажи содержимое файла `src/components/ui/Dropdown.tsx`:

```bash
cat src/components/ui/Dropdown.tsx
```

И файла где он импортируется:

```bash
cat src/components/ui/HeaderCell.tsx
```

---

Скорее всего одна из двух причин:

**1. Default export вместо named export**
```tsx
// Dropdown.tsx — сейчас так (неверно для named import)
export default function Dropdown() { ... }

// HeaderCell.tsx — импортирует так
import { Dropdown } from './Dropdown'; // ❌ не найдёт
```

**Решение:**
```tsx
// Вариант А — поменять экспорт в Dropdown.tsx
export const Dropdown = () => { ... } // ✅ named export

// Вариант Б — поменять импорт в HeaderCell.tsx
import Dropdown from './Dropdown'; // ✅ default import
```

**2. Файл `Dropdown.tsx` пустой или не существует**
```bash
ls src/components/ui/
```

Покажи содержимое файлов — исправим сразу.

**User - --**

файл Dropdown.tsx пустой.   создай содержимое этого файла по аналогии с уже существующим Button.trx:
import React from 'react';
import { clsx } from 'clsx';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
}

const variantClass: Record<Variant, string> = {
  primary:   'bg-blue-600 text-white hover:bg-blue-700 disabled:bg-blue-300',
  secondary: 'bg-gray-100 text-gray-800 hover:bg-gray-200',
  ghost:     'bg-transparent text-gray-600 hover:bg-gray-100',
  danger:    'bg-red-600 text-white hover:bg-red-700',
};

const sizeClass: Record<Size, string> = {
  sm: 'px-2.5 py-1.5 text-sm',
  md: 'px-4 py-2 text-sm',
  lg: 'px-5 py-2.5 text-base',
};

export const Button: React.FC<ButtonProps> = ({
  variant = 'primary', size = 'md', loading, children, className, disabled, ...props
}) => (
  <button
    {...props}
    disabled={disabled || loading}
    className={clsx(
      'inline-flex items-center gap-2 font-medium rounded-lg transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500',
      variantClass[variant],
      sizeClass[size],
      className
    )}
  >
    {loading && (
      <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
      </svg>
    )}
    {children}
  </button>
);

**Assistant - --**

## `src/components/ui/Dropdown.tsx`

```tsx
import React, { useEffect, useRef, useState } from 'react';
import { clsx } from 'clsx';

type DropdownSize = 'sm' | 'md' | 'lg';
type DropdownAlign = 'left' | 'right';

export interface DropdownItem {
  label: string;
  value: string;
  icon?: React.ReactNode;
  danger?: boolean;
  disabled?: boolean;
  divider?: boolean;
}

interface DropdownProps {
  trigger: React.ReactNode;
  items: DropdownItem[];
  onSelect: (item: DropdownItem) => void;
  size?: DropdownSize;
  align?: DropdownAlign;
  disabled?: boolean;
  className?: string;
}

const sizeClass: Record<DropdownSize, string> = {
  sm: 'text-xs py-1',
  md: 'text-sm py-1.5',
  lg: 'text-base py-2',
};

const itemSizeClass: Record<DropdownSize, string> = {
  sm: 'px-3 py-1 text-xs',
  md: 'px-4 py-2 text-sm',
  lg: 'px-5 py-2.5 text-base',
};

const alignClass: Record<DropdownAlign, string> = {
  left:  'left-0',
  right: 'right-0',
};

export const Dropdown: React.FC<DropdownProps> = ({
  trigger,
  items,
  onSelect,
  size = 'md',
  align = 'left',
  disabled = false,
  className,
}) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // Закрытие по клику вне компонента
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Закрытие по Escape
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleSelect = (item: DropdownItem) => {
    if (item.disabled) return;
    onSelect(item);
    setOpen(false);
  };

  return (
    <div ref={ref} className={clsx('relative inline-block', className)}>
      {/* Триггер */}
      <div
        onClick={() => !disabled && setOpen(prev => !prev)}
        className={clsx(disabled && 'opacity-50 cursor-not-allowed')}
      >
        {trigger}
      </div>

      {/* Меню */}
      {open && (
        <div
          className={clsx(
            'absolute z-50 mt-1 min-w-[10rem] bg-white border border-gray-200',
            'rounded-lg shadow-lg overflow-hidden',
            sizeClass[size],
            alignClass[align],
          )}
        >
          {items.map((item, index) => (
            <React.Fragment key={item.value ?? index}>
              {/* Разделитель */}
              {item.divider && (
                <div className="my-1 border-t border-gray-100" />
              )}

              {/* Пункт меню */}
              <button
                type="button"
                disabled={item.disabled}
                onClick={() => handleSelect(item)}
                className={clsx(
                  'w-full flex items-center gap-2 text-left transition-colors',
                  'focus:outline-none focus:bg-gray-50',
                  itemSizeClass[size],
                  item.danger
                    ? 'text-red-600 hover:bg-red-50'
                    : 'text-gray-700 hover:bg-gray-100',
                  item.disabled && 'opacity-40 cursor-not-allowed pointer-events-none',
                )}
              >
                {item.icon && (
                  <span className="shrink-0 w-4 h-4">{item.icon}</span>
                )}
                {item.label}
              </button>
            </React.Fragment>
          ))}
        </div>
      )}
    </div>
  );
};
```

---

### Пример использования в `HeaderCell.tsx`

```tsx
import { Dropdown, DropdownItem } from '../ui/Dropdown';

const menuItems: DropdownItem[] = [
  { label: 'Переименовать', value: 'rename' },
  { label: 'Сортировать А → Я', value: 'sort-asc' },
  { label: 'Сортировать Я → А', value: 'sort-desc' },
  { divider: true, label: '', value: 'divider-1' },
  { label: 'Удалить', value: 'delete', danger: true },
];

<Dropdown
  trigger={<button>⋯</button>}
  items={menuItems}
  align="right"
  onSelect={(item) => console.log(item.value)}
/>
```

**User - --**

создай аналогично файл Input.tsx

**Assistant - --**

## `src/components/ui/Input.tsx`

```tsx
import React, { forwardRef } from 'react';
import { clsx } from 'clsx';

type InputSize = 'sm' | 'md' | 'lg';
type InputVariant = 'default' | 'error' | 'success';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;
  size?: InputSize;
  variant?: InputVariant;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  loading?: boolean;
}

const sizeClass: Record<InputSize, string> = {
  sm: 'px-2.5 py-1.5 text-sm',
  md: 'px-3 py-2 text-sm',
  lg: 'px-4 py-2.5 text-base',
};

const variantClass: Record<InputVariant, string> = {
  default: 'border-gray-300 focus:border-blue-500 focus:ring-blue-500',
  error:   'border-red-400 focus:border-red-500 focus:ring-red-500',
  success: 'border-green-400 focus:border-green-500 focus:ring-green-500',
};

export const Input = forwardRef<HTMLInputElement, InputProps>(({
  label,
  hint,
  error,
  size = 'md',
  variant = 'default',
  leftIcon,
  rightIcon,
  loading = false,
  className,
  disabled,
  id,
  ...props
}, ref) => {
  const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-');
  const resolvedVariant: InputVariant = error ? 'error' : variant;

  return (
    <div className="flex flex-col gap-1 w-full">

      {/* Label */}
      {label && (
        <label
          htmlFor={inputId}
          className="text-sm font-medium text-gray-700"
        >
          {label}
          {props.required && (
            <span className="ml-1 text-red-500">*</span>
          )}
        </label>
      )}

      {/* Input wrapper */}
      <div className="relative flex items-center">

        {/* Left icon */}
        {leftIcon && (
          <span className="absolute left-3 flex items-center text-gray-400 pointer-events-none">
            {leftIcon}
          </span>
        )}

        {/* Input */}
        <input
          {...props}
          ref={ref}
          id={inputId}
          disabled={disabled || loading}
          className={clsx(
            'w-full rounded-lg border bg-white outline-none transition-colors',
            'focus:ring-2 focus:ring-offset-0',
            'placeholder:text-gray-400',
            'disabled:bg-gray-50 disabled:text-gray-400 disabled:cursor-not-allowed',
            sizeClass[size],
            variantClass[resolvedVariant],
            leftIcon  && 'pl-9',
            (rightIcon || loading) && 'pr-9',
            className,
          )}
        />

        {/* Right icon / spinner */}
        {(rightIcon || loading) && (
          <span className="absolute right-3 flex items-center text-gray-400 pointer-events-none">
            {loading ? (
              <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle
                  className="opacity-25"
                  cx="12" cy="12" r="10"
                  stroke="currentColor" strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8v8H4z"
                />
              </svg>
            ) : rightIcon}
          </span>
        )}
      </div>

      {/* Error message */}
      {error && (
        <p className="text-xs text-red-500 flex items-center gap-1">
          <svg className="w-3.5 h-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
          </svg>
          {error}
        </p>
      )}

      {/* Hint */}
      {hint && !error && (
        <p className="text-xs text-gray-400">{hint}</p>
      )}

    </div>
  );
});

Input.displayName = 'Input';
```

---

### Примеры использования

```tsx
import { Input } from '../ui/Input';

// Обычный
<Input placeholder="Введите текст" />

// С label и hint
<Input
  label="Email"
  hint="Мы не передаём данные третьим лицам"
  type="email"
  placeholder="example@mail.com"
  required
/>

// С ошибкой
<Input
  label="Пароль"
  type="password"
  error="Минимум 8 символов"
/>

// С иконкой слева
<Input
  label="Поиск"
  leftIcon={<SearchIcon className="w-4 h-4" />}
  placeholder="Найти..."
/>

// С состоянием загрузки
<Input
  label="Логин"
  loading={isChecking}
  value={username}
  onChange={e => setUsername(e.target.value)}
/>

// Через ref (для форм)
const ref = useRef<HTMLInputElement>(null);
<Input ref={ref} label="Имя" />
```

**User - --**

ошибка в строке interface Input Props extends React.InputHTMLAttributes
 - Interface 'InputProps' incorrectly extends interface InputHTMLAttributes. Types of property 'size' are incompatible

**Assistant - --**

Проблема в том, что `React.InputHTMLAttributes` уже содержит свойство `size: number`, а мы переопределяем его как `'sm' | 'md' | 'lg'`.

**Решение — исключить `size` из базового интерфейса через `Omit`:**

```tsx
interface InputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?: string;
  hint?: string;
  error?: string;
  size?: InputSize;          // теперь конфликта нет
  variant?: InputVariant;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  loading?: boolean;
}
```

Аналогичная ситуация встречается с другими пропсами HTML-элементов — на будущее:

| Конфликтующий проп | Элемент | Тип в HTML | Решение |
|---|---|---|---|
| `size` | `input`, `select` | `number` | `Omit<..., 'size'>` |
| `color` | `input` | `string` | `Omit<..., 'color'>` |
| `label` | `option` | `string` | `Omit<..., 'label'>` |
| `translate` | любой | `'yes' \| 'no'` | `Omit<..., 'translate'>` |

**User - --**

создай по аналогии файл Select.tsx

**Assistant - --**

## `src/components/ui/Select.tsx`

```tsx
import React, { forwardRef } from 'react';
import { clsx } from 'clsx';

type SelectSize = 'sm' | 'md' | 'lg';
type SelectVariant = 'default' | 'error' | 'success';

export interface SelectOption {
  label: string;
  value: string;
  disabled?: boolean;
}

export interface SelectGroup {
  label: string;
  options: SelectOption[];
}

interface SelectProps extends Omit<React.SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label?: string;
  hint?: string;
  error?: string;
  size?: SelectSize;
  variant?: SelectVariant;
  options?: SelectOption[];
  groups?: SelectGroup[];
  placeholder?: string;
  loading?: boolean;
}

const sizeClass: Record<SelectSize, string> = {
  sm: 'px-2.5 py-1.5 text-sm',
  md: 'px-3 py-2 text-sm',
  lg: 'px-4 py-2.5 text-base',
};

const variantClass: Record<SelectVariant, string> = {
  default: 'border-gray-300 focus:border-blue-500 focus:ring-blue-500',
  error:   'border-red-400 focus:border-red-500 focus:ring-red-500',
  success: 'border-green-400 focus:border-green-500 focus:ring-green-500',
};

export const Select = forwardRef<HTMLSelectElement, SelectProps>(({
  label,
  hint,
  error,
  size = 'md',
  variant = 'default',
  options = [],
  groups = [],
  placeholder,
  loading = false,
  className,
  disabled,
  id,
  ...props
}, ref) => {
  const selectId = id ?? label?.toLowerCase().replace(/\s+/g, '-');
  const resolvedVariant: SelectVariant = error ? 'error' : variant;

  return (
    <div className="flex flex-col gap-1 w-full">

      {/* Label */}
      {label && (
        <label
          htmlFor={selectId}
          className="text-sm font-medium text-gray-700"
        >
          {label}
          {props.required && (
            <span className="ml-1 text-red-500">*</span>
          )}
        </label>
      )}

      {/* Select wrapper */}
      <div className="relative flex items-center">

        {/* Select */}
        <select
          {...props}
          ref={ref}
          id={selectId}
          disabled={disabled || loading}
          className={clsx(
            'w-full rounded-lg border bg-white outline-none transition-colors appearance-none',
            'focus:ring-2 focus:ring-offset-0',
            'disabled:bg-gray-50 disabled:text-gray-400 disabled:cursor-not-allowed',
            sizeClass[size],
            variantClass[resolvedVariant],
            'pr-9',
            className,
          )}
        >
          {/* Placeholder */}
          {placeholder && (
            <option value="" disabled hidden>
              {placeholder}
            </option>
          )}

          {/* Flat options */}
          {options.map(option => (
            <option
              key={option.value}
              value={option.value}
              disabled={option.disabled}
            >
              {option.label}
            </option>
          ))}

          {/* Grouped options */}
          {groups.map(group => (
            <optgroup key={group.label} label={group.label}>
              {group.options.map(option => (
                <option
                  key={option.value}
                  value={option.value}
                  disabled={option.disabled}
                >
                  {option.label}
                </option>
              ))}
            </optgroup>
          ))}
        </select>

        {/* Chevron / spinner */}
        <span className="absolute right-3 flex items-center text-gray-400 pointer-events-none">
          {loading ? (
            <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle
                className="opacity-25"
                cx="12" cy="12" r="10"
                stroke="currentColor" strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8v8H4z"
              />
            </svg>
          ) : (
            <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M5.22 8.22a.75.75 0 011.06 0L10 11.94l3.72-3.72a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.22 9.28a.75.75 0 010-1.06z" clipRule="evenodd" />
            </svg>
          )}
        </span>
      </div>

      {/* Error message */}
      {error && (
        <p className="text-xs text-red-500 flex items-center gap-1">
          <svg className="w-3.5 h-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
          </svg>
          {error}
        </p>
      )}

      {/* Hint */}
      {hint && !error && (
        <p className="text-xs text-gray-400">{hint}</p>
      )}

    </div>
  );
});

Select.displayName = 'Select';
```

---

### Примеры использования

```tsx
import { Select } from '../ui/Select';

// Простой список
<Select
  options={[
    { label: 'Текст',  value: 'text' },
    { label: 'Число', value: 'number' },
    { label: 'Дата',  value: 'date' },
  ]}
  placeholder="Выберите тип"
/>

// С label и hint
<Select
  label="Тип колонки"
  hint="Тип нельзя изменить после создания"
  options={typeOptions}
  required
/>

// С ошибкой
<Select
  label="Workspace"
  error="Обязательное поле"
  options={workspaceOptions}
/>

// Сгруппированные опции
<Select
  label="Формат"
  groups={[
    {
      label: 'Числа',
      options: [
        { label: 'Целое',    value: 'int' },
        { label: 'Дробное',  value: 'float' },
      ],
    },
    {
      label: 'Текст',
      options: [
        { label: 'Строка',   value: 'string' },
        { label: 'Markdown', value: 'markdown' },
      ],
    },
  ]}
/>

// С состоянием загрузки
<Select
  label="Таблица"
  loading={isLoading}
  options={tableOptions}
/>

// Через ref
const ref = useRef<HTMLSelectElement>(null);
<Select ref={ref} options={options} />
```

**User - --**

файл frontend/src/hooks/userPermission.ts пустой, содай его

**Assistant - --**

Покажи несколько файлов для контекста:

```bash
cat src/store/authStore.ts
cat src/types/index.ts   # или где хранятся типы
```

И где используется `usePermission`:

```bash
grep -r "usePermission" src/ --include="*.ts" --include="*.tsx" -l
# затем покажи один из этих файлов
```

---

Пока жду — создам **разумную базовую версию** на основе типичной структуры таких хуков:

## `src/hooks/usePermission.ts`

```ts
import { useAuthStore } from '../store/authStore';

export type Role = 'owner' | 'editor' | 'viewer';

export type Action =
  | 'workspace:create'
  | 'workspace:delete'
  | 'workspace:rename'
  | 'spreadsheet:create'
  | 'spreadsheet:delete'
  | 'spreadsheet:rename'
  | 'row:create'
  | 'row:delete'
  | 'row:edit'
  | 'column:create'
  | 'column:delete'
  | 'column:edit'
  | 'member:invite'
  | 'member:remove'
  | 'member:change-role';

// Что может делать каждая роль
const ROLE_PERMISSIONS: Record<Role, Action[]> = {
  owner: [
    'workspace:create',
    'workspace:delete',
    'workspace:rename',
    'spreadsheet:create',
    'spreadsheet:delete',
    'spreadsheet:rename',
    'row:create',
    'row:delete',
    'row:edit',
    'column:create',
    'column:delete',
    'column:edit',
    'member:invite',
    'member:remove',
    'member:change-role',
  ],
  editor: [
    'spreadsheet:create',
    'spreadsheet:rename',
    'row:create',
    'row:delete',
    'row:edit',
    'column:create',
    'column:edit',
  ],
  viewer: [],
};

interface UsePermissionReturn {
  role: Role | null;
  can: (action: Action) => boolean;
  canAny: (actions: Action[]) => boolean;
  canAll: (actions: Action[]) => boolean;
  isOwner: boolean;
  isEditor: boolean;
  isViewer: boolean;
}

export const usePermission = (workspaceRole?: Role): UsePermissionReturn => {
  const { user } = useAuthStore();

  // Роль берём из аргумента (контекст workspace) или из глобального стора
  const role: Role | null = workspaceRole ?? (user?.role as Role) ?? null;

  const allowedActions: Action[] = role ? ROLE_PERMISSIONS[role] : [];

  const can = (action: Action): boolean =>
    allowedActions.includes(action);

  const canAny = (actions: Action[]): boolean =>
    actions.some(action => allowedActions.includes(action));

  const canAll = (actions: Action[]): boolean =>
    actions.every(action => allowedActions.includes(action));

  return {
    role,
    can,
    canAny,
    canAll,
    isOwner:  role === 'owner',
    isEditor: role === 'editor',
    isViewer: role === 'viewer',
  };
};
```

---

### Примеры использования

```tsx
import { usePermission } from '../hooks/usePermission';

// Базовое использование
const { can, isOwner } = usePermission();

// С ролью из контекста workspace
const { can, role } = usePermission(workspace.currentUserRole);

// Скрыть кнопку удаления для не-owner
{can('workspace:delete') && (
  <Button variant="danger" onClick={handleDelete}>
    Удалить
  </Button>
)}

// Заблокировать редактирование для viewer
<Input
  value={value}
  disabled={!can('row:edit')}
/>

// Проверить хотя бы одно из прав
{canAny(['row:create', 'row:edit']) && <Edittoolbar />}
```

---

Как только покажешь `authStore.ts` и типы — подправлю хук под реальную структуру `user` объекта.

**User - --**

контекст - мы создаем приложение, которое позволяет пользователям создавать электронные таблицы,описывая поля и тип полей. можно натсраивать доступ определенных пользователей с таблицам, строкам и полям для просмотра и редактирования.интерфейс - веб. для хранения информации используется субд postgres
Стек
React 18 + TypeScript
Zustand — стейт-менеджмент
TanStack Query — серверный стейт
TanStack Virtual — виртуализация строк
Tailwind CSS — стили
Структура файлов
frontend/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── types/
│   │   └── index.ts
│   ├── api/
│   │   ├── client.ts
│   │   ├── auth.ts
│   │   ├── spreadsheets.ts
│   │   ├── fields.ts
│   │   ├── rows.ts
│   │   └── permissions.ts
│   ├── store/
│   │   ├── authStore.ts
│   │   ├── spreadsheetStore.ts
│   │   └── permissionStore.ts
│   ├── hooks/
│   │   ├── useSpreadsheet.ts
│   │   ├── useRows.ts
│   │   └── usePermissions.ts
│   ├── components/
│   │   ├── ui/
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Select.tsx
│   │   │   └── Dropdown.tsx
│   │   ├── Grid/
│   │   │   ├── Grid.tsx
│   │   │   ├── HeaderCell.tsx
│   │   │   ├── Cell.tsx
│   │   │   └── AddColumnButton.tsx
│   │   ├── CellEditors/
│   │   │   ├── index.ts
│   │   │   ├── TextEditor.tsx
│   │   │   ├── NumberEditor.tsx
│   │   │   ├── BooleanEditor.tsx
│   │   │   ├── DateEditor.tsx
│   │   │   ├── SelectEditor.tsx
│   │   │   └── RelationEditor.tsx
│   │   ├── FieldEditor/
│   │   │   ├── FieldEditorModal.tsx
│   │   │   ├── FieldTypeSelect.tsx
│   │   │   └── SelectOptionsEditor.tsx
│   │   └── Permissions/
│   │       ├── ShareModal.tsx
│   │       ├── FieldAccessMatrix.tsx
│   │       └── RowRuleBuilder.tsx
│   └── pages/
│       ├── LoginPage.tsx
│       ├── WorkspacesPage.tsx
│       └── SpreadsheetPage.tsx

файл authStore.ts:
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '../types';

interface AuthState {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      setAuth: (user, token) => {
        localStorage.setItem('access_token', token);
        set({ user, token });
      },
      logout: () => {
        localStorage.removeItem('access_token');
        set({ user: null, token: null });
      },
      isAuthenticated: () => Boolean(get().token),
    }),
    { name: 'auth-storage', partialize: (s) => ({ token: s.token, user: s.user }) }
  )
);

аналогичный файл userRows.ts:
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rowsApi } from '../api/rows';
import { useSpreadsheetStore } from '../store/spreadsheetStore';

const PAGE_SIZE = 50;

export function useRows(spreadsheetId: string) {
  const { filters, sorts } = useSpreadsheetStore();
  const qc = useQueryClient();

  const query = useInfiniteQuery({
    queryKey: ['rows', spreadsheetId, filters, sorts],
    queryFn: ({ pageParam = 0 }) =>
      rowsApi.list(spreadsheetId, {
        limit: PAGE_SIZE,
        offset: pageParam as number,
        filters,
        sorts,
      }),
    getNextPageParam: (last, pages) => {
      const loaded = pages.length * PAGE_SIZE;
      return loaded < last.total ? loaded : undefined;
    },
    initialPageParam: 0,
  });

  const allRows = query.data?.pages.flatMap((p) => p.data) ?? [];

  const createRow = useMutation({
    mutationFn: (data: Record<string, unknown>) => rowsApi.create(spreadsheetId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  const updateRow = useMutation({
    mutationFn: ({ rowId, data }: { rowId: string; data: Record<string, unknown> }) =>
      rowsApi.update(spreadsheetId, rowId, data),
    onMutate: async ({ rowId, data }) => {
      await qc.cancelQueries({ queryKey: ['rows', spreadsheetId] });
      const prev = qc.getQueryData(['rows', spreadsheetId]);
      qc.setQueryData(['rows', spreadsheetId, filters, sorts], (old: typeof query.data) => ({
        ...old,
        pages: old?.pages.map((page) => ({
          ...page,
          data: page.data.map((row) => (row._id === rowId ? { ...row, ...data } : row)),
        })),
      }));
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(['rows', spreadsheetId], ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  const deleteRow = useMutation({
    mutationFn: (rowId: string) => rowsApi.delete(spreadsheetId, rowId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  return { ...query, allRows, createRow, updateRow, deleteRow };
}

**Assistant - --**

## `src/hooks/usePermissions.ts`

```ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../api/permissions';
import { usePermissionStore } from '../store/permissionStore';
import { useAuthStore } from '../store/authStore';
import type {
  SpreadsheetPermission,
  FieldPermission,
  RowAccessRule,
  PermissionRole,
} from '../types';

// ─── Spreadsheet-level permissions ───────────────────────────────────────────

export function useSpreadsheetPermissions(spreadsheetId: string) {
  const qc = useQueryClient();
  const { setPermissions } = usePermissionStore();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId],
    queryFn: async () => {
      const data = await permissionsApi.listSpreadsheet(spreadsheetId);
      setPermissions(spreadsheetId, data);
      return data;
    },
    staleTime: 60_000,
  });

  const invite = useMutation({
    mutationFn: (payload: { email: string; role: PermissionRole }) =>
      permissionsApi.invite(spreadsheetId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const changeRole = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: PermissionRole }) =>
      permissionsApi.changeRole(spreadsheetId, userId, role),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const revoke = useMutation({
    mutationFn: (userId: string) =>
      permissionsApi.revoke(spreadsheetId, userId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  return { ...query, invite, changeRole, revoke };
}

// ─── Field-level permissions ──────────────────────────────────────────────────

export function useFieldPermissions(spreadsheetId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId, 'fields'],
    queryFn: () => permissionsApi.listFields(spreadsheetId),
    staleTime: 60_000,
  });

  const update = useMutation({
    mutationFn: (payload: FieldPermission) =>
      permissionsApi.updateField(spreadsheetId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'fields'] }),
  });

  const remove = useMutation({
    mutationFn: ({ fieldId, userId }: { fieldId: string; userId: string }) =>
      permissionsApi.removeField(spreadsheetId, fieldId, userId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'fields'] }),
  });

  return { ...query, update, remove };
}

// ─── Row-level access rules ───────────────────────────────────────────────────

export function useRowRules(spreadsheetId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId, 'rows'],
    queryFn: () => permissionsApi.listRowRules(spreadsheetId),
    staleTime: 60_000,
  });

  const create = useMutation({
    mutationFn: (rule: Omit<RowAccessRule, 'id'>) =>
      permissionsApi.createRowRule(spreadsheetId, rule),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  const update = useMutation({
    mutationFn: (rule: RowAccessRule) =>
      permissionsApi.updateRowRule(spreadsheetId, rule),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  const remove = useMutation({
    mutationFn: (ruleId: string) =>
      permissionsApi.deleteRowRule(spreadsheetId, ruleId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  return { ...query, create, update, remove };
}

// ─── Current user ability check ───────────────────────────────────────────────

export function useAbility(spreadsheetId: string) {
  const { user } = useAuthStore();
  const { getPermissions } = usePermissionStore();

  const permissions: SpreadsheetPermission[] = getPermissions(spreadsheetId);

  const currentUserPermission = permissions.find(
    (p) => p.userId === user?.id,
  );

  const role: PermissionRole | null = currentUserPermission?.role ?? null;

  const can = (action: AbilityAction): boolean => {
    if (!role) return false;
    return ROLE_ABILITIES[role].includes(action);
  };

  const canAny = (actions: AbilityAction[]): boolean =>
    actions.some(can);

  const canAll = (actions: AbilityAction[]): boolean =>
    actions.every(can);

  const canEditField = (fieldId: string): boolean => {
    const fp = currentUserPermission?.fieldPermissions?.find(
      (f) => f.fieldId === fieldId,
    );
    // Явный запрет поля перекрывает роль
    if (fp?.access === 'none') return false;
    if (fp?.access === 'view') return false;
    return can('row:edit');
  };

  const canViewField = (fieldId: string): boolean => {
    const fp = currentUserPermission?.fieldPermissions?.find(
      (f) => f.fieldId === fieldId,
    );
    if (fp?.access === 'none') return false;
    return true;
  };

  return {
    role,
    can,
    canAny,
    canAll,
    canEditField,
    canViewField,
    isOwner:  role === 'owner',
    isEditor: role === 'editor',
    isViewer: role === 'viewer',
  };
}

// ─── Ability matrix ───────────────────────────────────────────────────────────

export type AbilityAction =
  | 'spreadsheet:rename'
  | 'spreadsheet:delete'
  | 'field:create'
  | 'field:edit'
  | 'field:delete'
  | 'row:create'
  | 'row:edit'
  | 'row:delete'
  | 'member:invite'
  | 'member:remove'
  | 'member:change-role'
  | 'field-access:edit'
  | 'row-rules:edit';

const ROLE_ABILITIES: Record<PermissionRole, AbilityAction[]> = {
  owner: [
    'spreadsheet:rename',
    'spreadsheet:delete',
    'field:create',
    'field:edit',
    'field:delete',
    'row:create',
    'row:edit',
    'row:delete',
    'member:invite',
    'member:remove',
    'member:change-role',
    'field-access:edit',
    'row-rules:edit',
  ],
  editor: [
    'field:create',
    'field:edit',
    'row:create',
    'row:edit',
    'row:delete',
  ],
  viewer: [],
};
```

---

Теперь нужны типы в `src/types/index.ts`. Добавь если их ещё нет:

```ts
export type PermissionRole = 'owner' | 'editor' | 'viewer';

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

export interface RowAccessRule {
  id: string;
  spreadsheetId: string;
  userId: string;
  // Условие: поле + оператор + значение
  fieldId: string;
  operator: 'eq' | 'neq' | 'gt' | 'lt' | 'contains' | 'is_empty';
  value: string | number | boolean | null;
  access: 'view' | 'none';
}
```

---

### Примеры использования

```tsx
import { useAbility } from '../hooks/usePermissions';

const { can, canEditField, canViewField, isOwner } = useAbility(spreadsheetId);

// Скрыть кнопку удаления
{can('spreadsheet:delete') && <Button variant="danger">Удалить</Button>}

// Заблокировать ячейку
<Cell
  editable={canEditField(field.id)}
  visible={canViewField(field.id)}
/>

// Показать вкладку настройки доступа только owner
{isOwner && <ShareModal spreadsheetId={spreadsheetId} />}
```

**User - --**

поправь код этого файла с учётом кода файла src/api/permissions.ts
import { apiClient } from './client';
import type { SpreadsheetAccess, FieldAccess, RowAccessRule } from '../types';

export const permissionsApi = {
  getSpreadsheetAccess: (spreadsheetId: string) =>
    apiClient.get<SpreadsheetAccess[]>(`/spreadsheets/\${spreadsheetId}/permissions`).then(r => r.data),

  upsertSpreadsheetAccess: (spreadsheetId: string, data: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
    apiClient.put(`/spreadsheets/\${spreadsheetId}/permissions`, data),

  removeSpreadsheetAccess: (spreadsheetId: string, principalId: string) =>
    apiClient.delete(`/spreadsheets/\${spreadsheetId}/permissions/\${principalId}`),

  getFieldAccess: (spreadsheetId: string) =>
    apiClient.get<FieldAccess[]>(`/spreadsheets/\${spreadsheetId}/field-permissions`).then(r => r.data),

  upsertFieldAccess: (fieldId: string, data: Omit<FieldAccess, 'field_id'>) =>
    apiClient.put(`/fields/\${fieldId}/permissions`, data),

  getRowRules: (spreadsheetId: string) =>
    apiClient.get<RowAccessRule[]>(`/spreadsheets/\${spreadsheetId}/row-rules`).then(r => r.data),

  upsertRowRule: (spreadsheetId: string, data: Omit<RowAccessRule, 'id'>) =>
    apiClient.put(`/spreadsheets/\${spreadsheetId}/row-rules`, data),

  deleteRowRule: (spreadsheetId: string, ruleId: string) =>
    apiClient.delete(`/spreadsheets/\${spreadsheetId}/row-rules/\${ruleId}`),
};

**Assistant - --**

## `src/hooks/usePermissions.ts`

```ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../api/permissions';
import { usePermissionStore } from '../store/permissionStore';
import { useAuthStore } from '../store/authStore';
import type {
  SpreadsheetAccess,
  FieldAccess,
  RowAccessRule,
  PermissionRole,
} from '../types';

// ─── Spreadsheet-level permissions ───────────────────────────────────────────

export function useSpreadsheetPermissions(spreadsheetId: string) {
  const qc = useQueryClient();
  const { setPermissions } = usePermissionStore();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId],
    queryFn: async () => {
      const data = await permissionsApi.getSpreadsheetAccess(spreadsheetId);
      setPermissions(spreadsheetId, data);
      return data;
    },
    staleTime: 60_000,
  });

  const upsert = useMutation({
    mutationFn: (payload: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
      permissionsApi.upsertSpreadsheetAccess(spreadsheetId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (principalId: string) =>
      permissionsApi.removeSpreadsheetAccess(spreadsheetId, principalId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  return { ...query, upsert, remove };
}

// ─── Field-level permissions ──────────────────────────────────────────────────

export function useFieldPermissions(spreadsheetId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId, 'fields'],
    queryFn: () => permissionsApi.getFieldAccess(spreadsheetId),
    staleTime: 60_000,
  });

  const upsert = useMutation({
    mutationFn: ({ fieldId, data }: { fieldId: string; data: Omit<FieldAccess, 'field_id'> }) =>
      permissionsApi.upsertFieldAccess(fieldId, data),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'fields'] }),
  });

  return { ...query, upsert };
}

// ─── Row-level access rules ───────────────────────────────────────────────────

export function useRowRules(spreadsheetId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId, 'rows'],
    queryFn: () => permissionsApi.getRowRules(spreadsheetId),
    staleTime: 60_000,
  });

  const upsert = useMutation({
    mutationFn: (rule: Omit<RowAccessRule, 'id'>) =>
      permissionsApi.upsertRowRule(spreadsheetId, rule),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  const remove = useMutation({
    mutationFn: (ruleId: string) =>
      permissionsApi.deleteRowRule(spreadsheetId, ruleId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  return { ...query, upsert, remove };
}

// ─── Current user ability check ───────────────────────────────────────────────

export function useAbility(spreadsheetId: string) {
  const { user } = useAuthStore();
  const { getPermissions } = usePermissionStore();

  const permissions: SpreadsheetAccess[] = getPermissions(spreadsheetId);

  const currentAccess = permissions.find(
    (p) => p.principal_id === user?.id,
  );

  const role: PermissionRole | null = currentAccess?.role ?? null;

  const can = (action: AbilityAction): boolean => {
    if (!role) return false;
    return ROLE_ABILITIES[role].includes(action);
  };

  const canAny = (actions: AbilityAction[]): boolean =>
    actions.some(can);

  const canAll = (actions: AbilityAction[]): boolean =>
    actions.every(can);

  // Проверяется отдельно загруженный список FieldAccess
  const canEditField = (fieldId: string, fieldAccess: FieldAccess[]): boolean => {
    const fa = fieldAccess.find(
      (f) => f.field_id === fieldId && f.principal_id === user?.id,
    );
    if (fa?.access === 'none') return false;
    if (fa?.access === 'view') return false;
    return can('row:edit');
  };

  const canViewField = (fieldId: string, fieldAccess: FieldAccess[]): boolean => {
    const fa = fieldAccess.find(
      (f) => f.field_id === fieldId && f.principal_id === user?.id,
    );
    if (fa?.access === 'none') return false;
    return true;
  };

  return {
    role,
    can,
    canAny,
    canAll,
    canEditField,
    canViewField,
    isOwner:  role === 'owner',
    isEditor: role === 'editor',
    isViewer: role === 'viewer',
  };
}

// ─── Ability matrix ───────────────────────────────────────────────────────────

export type AbilityAction =
  | 'spreadsheet:rename'
  | 'spreadsheet:delete'
  | 'field:create'
  | 'field:edit'
  | 'field:delete'
  | 'row:create'
  | 'row:edit'
  | 'row:delete'
  | 'member:invite'
  | 'member:remove'
  | 'member:change-role'
  | 'field-access:edit'
  | 'row-rules:edit';

const ROLE_ABILITIES: Record<PermissionRole, AbilityAction[]> = {
  owner: [
    'spreadsheet:rename',
    'spreadsheet:delete',
    'field:create',
    'field:edit',
    'field:delete',
    'row:create',
    'row:edit',
    'row:delete',
    'member:invite',
    'member:remove',
    'member:change-role',
    'field-access:edit',
    'row-rules:edit',
  ],
  editor: [
    'field:create',
    'field:edit',
    'row:create',
    'row:edit',
    'row:delete',
  ],
  viewer: [],
};
```

---

Добавь в `src/types/index.ts` если ещё нет:

```ts
export type PermissionRole = 'owner' | 'editor' | 'viewer';

export interface SpreadsheetAccess {
  id: string;
  principal_id: string;
  principal_name: string;
  role: PermissionRole;
}

export interface FieldAccess {
  field_id: string;
  principal_id: string;
  access: 'view' | 'edit' | 'none';
}

export interface RowAccessRule {
  id: string;
  spreadsheetId: string;
  fieldId: string;
  principal_id: string;
  operator: 'eq' | 'neq' | 'gt' | 'lt' | 'contains' | 'is_empty';
  value: string | number | boolean | null;
  access: 'view' | 'none';
}
```

---

### Примеры использования

```tsx
import { useAbility, useFieldPermissions } from '../hooks/usePermissions';

const { can, canEditField, canViewField, isOwner } = useAbility(spreadsheetId);
const { data: fieldAccess = [] } = useFieldPermissions(spreadsheetId);

// Скрыть кнопку удаления
{can('spreadsheet:delete') && (
  <Button variant="danger">Удалить</Button>
)}

// Заблокировать ячейку с учётом field-level доступа
<Cell
  editable={canEditField(field.id, fieldAccess)}
  visible={canViewField(field.id, fieldAccess)}
/>

// Показать управление доступом только owner
{isOwner && <ShareModal spreadsheetId={spreadsheetId} />}
```

**User - --**

создай файл src/store/permissioStore.ts
вот код аналогичного файла src/store/spreadsheetStore.ts для примера:
import { create } from 'zustand';
import type { Spreadsheet, Field, FilterCondition, SortCondition } from '../types';

interface SpreadsheetState {
  spreadsheet: Spreadsheet | null;
  selectedRowIds: Set<string>;
  editingCell: { rowId: string; fieldId: string } | null;
  filters: FilterCondition[];
  sorts: SortCondition[];

  setSpreadsheet: (s: Spreadsheet) => void;
  updateField: (field: Field) => void;
  addField: (field: Field) => void;
  removeField: (fieldId: string) => void;

  setEditingCell: (rowId: string, fieldId: string) => void;
  clearEditingCell: () => void;

  toggleRowSelection: (rowId: string) => void;
  clearSelection: () => void;

  setFilters: (filters: FilterCondition[]) => void;
  setSorts: (sorts: SortCondition[]) => void;
}

export const useSpreadsheetStore = create<SpreadsheetState>((set) => ({
  spreadsheet: null,
  selectedRowIds: new Set(),
  editingCell: null,
  filters: [],
  sorts: [],

  setSpreadsheet: (spreadsheet) => set({ spreadsheet }),

  updateField: (field) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? {
            ...s.spreadsheet,
            fields: s.spreadsheet.fields.map((f) => (f.id === field.id ? field : f)),
          }
        : null,
    })),

  addField: (field) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? { ...s.spreadsheet, fields: [...s.spreadsheet.fields, field] }
        : null,
    })),

  removeField: (fieldId) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? { ...s.spreadsheet, fields: s.spreadsheet.fields.filter((f) => f.id !== fieldId) }
        : null,
    })),

  setEditingCell: (rowId, fieldId) => set({ editingCell: { rowId, fieldId } }),
  clearEditingCell: () => set({ editingCell: null }),

  toggleRowSelection: (rowId) =>
    set((s) => {
      const next = new Set(s.selectedRowIds);
      next.has(rowId) ? next.delete(rowId) : next.add(rowId);
      return { selectedRowIds: next };
    }),

  clearSelection: () => set({ selectedRowIds: new Set() }),

  setFilters: (filters) => set({ filters }),
  setSorts: (sorts) => set({ sorts }),
}));

**Assistant - --**

## `src/store/permissionStore.ts`

```ts
import { create } from 'zustand';
import type { SpreadsheetAccess, FieldAccess, RowAccessRule } from '../types';

interface PermissionState {
  // Доступы к таблицам: spreadsheetId → список доступов
  spreadsheetAccess: Record<string, SpreadsheetAccess[]>;
  // Доступы к полям: spreadsheetId → список доступов
  fieldAccess: Record<string, FieldAccess[]>;
  // Правила доступа к строкам: spreadsheetId → список правил
  rowRules: Record<string, RowAccessRule[]>;

  // Spreadsheet access
  setPermissions: (spreadsheetId: string, access: SpreadsheetAccess[]) => void;
  getPermissions: (spreadsheetId: string) => SpreadsheetAccess[];
  upsertPermission: (spreadsheetId: string, access: SpreadsheetAccess) => void;
  removePermission: (spreadsheetId: string, principalId: string) => void;

  // Field access
  setFieldAccess: (spreadsheetId: string, access: FieldAccess[]) => void;
  getFieldAccess: (spreadsheetId: string) => FieldAccess[];
  upsertFieldAccess: (spreadsheetId: string, access: FieldAccess) => void;

  // Row rules
  setRowRules: (spreadsheetId: string, rules: RowAccessRule[]) => void;
  getRowRules: (spreadsheetId: string) => RowAccessRule[];
  upsertRowRule: (spreadsheetId: string, rule: RowAccessRule) => void;
  removeRowRule: (spreadsheetId: string, ruleId: string) => void;

  // Сброс при смене таблицы
  clearSpreadsheet: (spreadsheetId: string) => void;
  clearAll: () => void;
}

export const usePermissionStore = create<PermissionState>((set, get) => ({
  spreadsheetAccess: {},
  fieldAccess: {},
  rowRules: {},

  // ─── Spreadsheet access ───────────────────────────────────────────────────

  setPermissions: (spreadsheetId, access) =>
    set((s) => ({
      spreadsheetAccess: { ...s.spreadsheetAccess, [spreadsheetId]: access },
    })),

  getPermissions: (spreadsheetId) =>
    get().spreadsheetAccess[spreadsheetId] ?? [],

  upsertPermission: (spreadsheetId, access) =>
    set((s) => {
      const current = s.spreadsheetAccess[spreadsheetId] ?? [];
      const exists = current.some((a) => a.principal_id === access.principal_id);
      return {
        spreadsheetAccess: {
          ...s.spreadsheetAccess,
          [spreadsheetId]: exists
            ? current.map((a) => (a.principal_id === access.principal_id ? access : a))
            : [...current, access],
        },
      };
    }),

  removePermission: (spreadsheetId, principalId) =>
    set((s) => ({
      spreadsheetAccess: {
        ...s.spreadsheetAccess,
        [spreadsheetId]: (s.spreadsheetAccess[spreadsheetId] ?? []).filter(
          (a) => a.principal_id !== principalId,
        ),
      },
    })),

  // ─── Field access ─────────────────────────────────────────────────────────

  setFieldAccess: (spreadsheetId, access) =>
    set((s) => ({
      fieldAccess: { ...s.fieldAccess, [spreadsheetId]: access },
    })),

  getFieldAccess: (spreadsheetId) =>
    get().fieldAccess[spreadsheetId] ?? [],

  upsertFieldAccess: (spreadsheetId, access) =>
    set((s) => {
      const current = s.fieldAccess[spreadsheetId] ?? [];
      const exists = current.some(
        (a) => a.field_id === access.field_id && a.principal_id === access.principal_id,
      );
      return {
        fieldAccess: {
          ...s.fieldAccess,
          [spreadsheetId]: exists
            ? current.map((a) =>
                a.field_id === access.field_id && a.principal_id === access.principal_id
                  ? access
                  : a,
              )
            : [...current, access],
        },
      };
    }),

  // ─── Row rules ────────────────────────────────────────────────────────────

  setRowRules: (spreadsheetId, rules) =>
    set((s) => ({
      rowRules: { ...s.rowRules, [spreadsheetId]: rules },
    })),

  getRowRules: (spreadsheetId) =>
    get().rowRules[spreadsheetId] ?? [],

  upsertRowRule: (spreadsheetId, rule) =>
    set((s) => {
      const current = s.rowRules[spreadsheetId] ?? [];
      const exists = current.some((r) => r.id === rule.id);
      return {
        rowRules: {
          ...s.rowRules,
          [spreadsheetId]: exists
            ? current.map((r) => (r.id === rule.id ? rule : r))
            : [...current, rule],
        },
      };
    }),

  removeRowRule: (spreadsheetId, ruleId) =>
    set((s) => ({
      rowRules: {
        ...s.rowRules,
        [spreadsheetId]: (s.rowRules[spreadsheetId] ?? []).filter(
          (r) => r.id !== ruleId,
        ),
      },
    })),

  // ─── Reset ────────────────────────────────────────────────────────────────

  clearSpreadsheet: (spreadsheetId) =>
    set((s) => {
      const spreadsheetAccess = { ...s.spreadsheetAccess };
      const fieldAccess = { ...s.fieldAccess };
      const rowRules = { ...s.rowRules };
      delete spreadsheetAccess[spreadsheetId];
      delete fieldAccess[spreadsheetId];
      delete rowRules[spreadsheetId];
      return { spreadsheetAccess, fieldAccess, rowRules };
    }),

  clearAll: () =>
    set({ spreadsheetAccess: {}, fieldAccess: {}, rowRules: {} }),
}));
```

---

### Примеры использования

```tsx
import { usePermissionStore } from '../store/permissionStore';

// Чтение доступов из стора (без запроса к серверу)
const access = usePermissionStore((s) => s.getPermissions(spreadsheetId));

// Обновить роль локально после мутации
const { upsertPermission } = usePermissionStore();
upsertPermission(spreadsheetId, { ...updatedAccess });

// Удалить участника локально
const { removePermission } = usePermissionStore();
removePermission(spreadsheetId, principalId);

// Очистить при уходе со страницы таблицы
useEffect(() => {
  return () => clearSpreadsheet(spreadsheetId);
}, [spreadsheetId]);
```

**User - --**

создай файл src/hooks/useSpreadsheet.ts

**Assistant - --**

## `src/hooks/useSpreadsheet.ts`

```ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { fieldsApi } from '../api/fields';
import { useSpreadsheetStore } from '../store/spreadsheetStore';
import type { Field, Spreadsheet } from '../types';

// ─── Spreadsheet ──────────────────────────────────────────────────────────────

export function useSpreadsheet(spreadsheetId: string) {
  const qc = useQueryClient();
  const { setSpreadsheet } = useSpreadsheetStore();

  const query = useQuery({
    queryKey: ['spreadsheet', spreadsheetId],
    queryFn: async () => {
      const data = await spreadsheetsApi.getById(spreadsheetId);
      setSpreadsheet(data);
      return data;
    },
    staleTime: 30_000,
  });

  const rename = useMutation({
    mutationFn: (name: string) =>
      spreadsheetsApi.update(spreadsheetId, { name }),
    onSuccess: (updated) => {
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], updated);
      setSpreadsheet(updated);
    },
  });

  const remove = useMutation({
    mutationFn: () => spreadsheetsApi.delete(spreadsheetId),
    onSuccess: () => {
      qc.removeQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      qc.invalidateQueries({ queryKey: ['spreadsheets'] });
    },
  });

  return { ...query, rename, remove };
}

// ─── Spreadsheets list ────────────────────────────────────────────────────────

export function useSpreadsheets(workspaceId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['spreadsheets', workspaceId],
    queryFn: () => spreadsheetsApi.list(workspaceId),
    staleTime: 30_000,
  });

  const create = useMutation({
    mutationFn: (name: string) =>
      spreadsheetsApi.create(workspaceId, { name }),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['spreadsheets', workspaceId] }),
  });

  return { ...query, create };
}

// ─── Fields ───────────────────────────────────────────────────────────────────

export function useFields(spreadsheetId: string) {
  const qc = useQueryClient();
  const { addField, updateField, removeField } = useSpreadsheetStore();

  const query = useQuery({
    queryKey: ['fields', spreadsheetId],
    queryFn: () => fieldsApi.list(spreadsheetId),
    staleTime: 30_000,
  });

  const create = useMutation({
    mutationFn: (data: Omit<Field, 'id' | 'spreadsheet_id' | 'order'>) =>
      fieldsApi.create(spreadsheetId, data),
    onSuccess: (field) => {
      addField(field);
      qc.invalidateQueries({ queryKey: ['fields', spreadsheetId] });
    },
  });

  const update = useMutation({
    mutationFn: ({ fieldId, data }: { fieldId: string; data: Partial<Field> }) =>
      fieldsApi.update(spreadsheetId, fieldId, data),
    onMutate: async ({ fieldId, data }) => {
      await qc.cancelQueries({ queryKey: ['fields', spreadsheetId] });
      const prev = qc.getQueryData<Field[]>(['fields', spreadsheetId]);
      qc.setQueryData<Field[]>(['fields', spreadsheetId], (old = []) =>
        old.map((f) => (f.id === fieldId ? { ...f, ...data } : f)),
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(['fields', spreadsheetId], ctx.prev);
    },
    onSuccess: (field) => {
      updateField(field);
      qc.invalidateQueries({ queryKey: ['fields', spreadsheetId] });
    },
  });

  const reorder = useMutation({
    mutationFn: (orderedIds: string[]) =>
      fieldsApi.reorder(spreadsheetId, orderedIds),
    onMutate: async (orderedIds) => {
      await qc.cancelQueries({ queryKey: ['fields', spreadsheetId] });
      const prev = qc.getQueryData<Field[]>(['fields', spreadsheetId]);
      qc.setQueryData<Field[]>(['fields', spreadsheetId], (old = []) => {
        const map = Object.fromEntries(old.map((f) => [f.id, f]));
        return orderedIds.map((id, index) => ({ ...map[id], order: index }));
      });
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(['fields', spreadsheetId], ctx.prev);
    },
    onSettled: () =>
      qc.invalidateQueries({ queryKey: ['fields', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (fieldId: string) =>
      fieldsApi.delete(spreadsheetId, fieldId),
    onSuccess: (_data, fieldId) => {
      removeField(fieldId);
      qc.invalidateQueries({ queryKey: ['fields', spreadsheetId] });
    },
  });

  return { ...query, create, update, reorder, remove };
}
```

---

Нужны методы в `src/api/spreadsheets.ts` и `src/api/fields.ts`. Покажи их код — подправлю под реальные сигнатуры. Пока ожидаемый интерфейс:

```ts
// src/api/spreadsheets.ts
export const spreadsheetsApi = {
  list:    (workspaceId: string) => ...,
  getById: (spreadsheetId: string) => ...,
  create:  (workspaceId: string, data: { name: string }) => ...,
  update:  (spreadsheetId: string, data: Partial<Spreadsheet>) => ...,
  delete:  (spreadsheetId: string) => ...,
};

// src/api/fields.ts
export const fieldsApi = {
  list:    (spreadsheetId: string) => ...,
  create:  (spreadsheetId: string, data: Omit<Field, 'id' | 'spreadsheet_id' | 'order'>) => ...,
  update:  (spreadsheetId: string, fieldId: string, data: Partial<Field>) => ...,
  reorder: (spreadsheetId: string, orderedIds: string[]) => ...,
  delete:  (spreadsheetId: string, fieldId: string) => ...,
};
```

---

### Примеры использования

```tsx
import { useSpreadsheet, useFields } from '../hooks/useSpreadsheet';

// Загрузка таблицы
const { data: spreadsheet, isLoading } = useSpreadsheet(spreadsheetId);

// Переименование
const { rename } = useSpreadsheet(spreadsheetId);
rename.mutate('Новое название');

// Поля
const { data: fields = [], create, update, reorder, remove } = useFields(spreadsheetId);

// Создать поле
create.mutate({ name: 'Телефон', type: 'text', required: false });

// Обновить поле
update.mutate({ fieldId: field.id, data: { name: 'Новое имя' } });

// Изменить порядок
reorder.mutate(['field-3', 'field-1', 'field-2']);

// Удалить поле
remove.mutate(field.id);
```

**User - --**

api/spreadsheets.ts:
import { apiClient } from './client';
import type { Spreadsheet } from '../types';

interface CreateSpreadsheetPayload {
  name: string;
  description?: string;
  workspace_id: string;
}

export const spreadsheetsApi = {
  list: (workspaceId: string) =>
    apiClient.get<Spreadsheet[]>(`/workspaces/\${workspaceId}/spreadsheets`).then(r => r.data),

  get: (id: string) =>
    apiClient.get<Spreadsheet>(`/spreadsheets/\${id}`).then(r => r.data),

  create: (data: CreateSpreadsheetPayload) =>
    apiClient.post<Spreadsheet>('/spreadsheets', data).then(r => r.data),

  update: (id: string, data: Partial<Pick<Spreadsheet, 'name' | 'description'>>) =>
    apiClient.put<Spreadsheet>(`/spreadsheets/\${id}`, data).then(r => r.data),

  delete: (id: string) =>
    apiClient.delete(`/spreadsheets/\${id}`),
};

api/fields.ts:
import { apiClient } from './client';
import type { Field, FieldType, FieldOptions } from '../types';

export interface CreateFieldPayload {
  name: string;
  field_type: FieldType;
  is_required?: boolean;
  is_unique?: boolean;
  default_value?: string;
  options?: FieldOptions;
}

export interface UpdateFieldPayload extends Partial<CreateFieldPayload> {
  position?: number;
}

export const fieldsApi = {
  create: (spreadsheetId: string, data: CreateFieldPayload) =>
    apiClient.post<Field>(`/spreadsheets/\${spreadsheetId}/fields`, data).then(r => r.data),

  update: (fieldId: string, data: UpdateFieldPayload) =>
    apiClient.put<Field>(`/fields/\${fieldId}`, data).then(r => r.data),

  delete: (fieldId: string) =>
    apiClient.delete(`/fields/\${fieldId}`),

  reorder: (spreadsheetId: string, fieldIds: string[]) =>
    apiClient.patch(`/spreadsheets/\${spreadsheetId}/fields/reorder`, { field_ids: fieldIds }),
};

**Assistant - --**

## `src/hooks/useSpreadsheet.ts`

```ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { fieldsApi } from '../api/fields';
import { useSpreadsheetStore } from '../store/spreadsheetStore';
import type { Field, Spreadsheet } from '../types';
import type { CreateFieldPayload, UpdateFieldPayload } from '../api/fields';

// ─── Spreadsheet ──────────────────────────────────────────────────────────────

export function useSpreadsheet(spreadsheetId: string) {
  const qc = useQueryClient();
  const { setSpreadsheet } = useSpreadsheetStore();

  const query = useQuery({
    queryKey: ['spreadsheet', spreadsheetId],
    queryFn: async () => {
      const data = await spreadsheetsApi.get(spreadsheetId);
      setSpreadsheet(data);
      return data;
    },
    staleTime: 30_000,
  });

  const rename = useMutation({
    mutationFn: (name: string) =>
      spreadsheetsApi.update(spreadsheetId, { name }),
    onSuccess: (updated) => {
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], updated);
      setSpreadsheet(updated);
    },
  });

  const remove = useMutation({
    mutationFn: () => spreadsheetsApi.delete(spreadsheetId),
    onSuccess: () => {
      qc.removeQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      qc.invalidateQueries({ queryKey: ['spreadsheets'] });
    },
  });

  return { ...query, rename, remove };
}

// ─── Spreadsheets list ────────────────────────────────────────────────────────

export function useSpreadsheets(workspaceId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['spreadsheets', workspaceId],
    queryFn: () => spreadsheetsApi.list(workspaceId),
    staleTime: 30_000,
  });

  const create = useMutation({
    mutationFn: (payload: { name: string; description?: string }) =>
      spreadsheetsApi.create({ ...payload, workspace_id: workspaceId }),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['spreadsheets', workspaceId] }),
  });

  return { ...query, create };
}

// ─── Fields ───────────────────────────────────────────────────────────────────

export function useFields(spreadsheetId: string) {
  const qc = useQueryClient();
  const { addField, updateField, removeField } = useSpreadsheetStore();

  // Поля живут внутри spreadsheet — берём их оттуда
  const query = useQuery({
    queryKey: ['spreadsheet', spreadsheetId],
    select: (data: Spreadsheet) => data.fields ?? [],
    staleTime: 30_000,
  });

  const create = useMutation({
    mutationFn: (data: CreateFieldPayload) =>
      fieldsApi.create(spreadsheetId, data),
    onSuccess: (field) => {
      addField(field);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) =>
        old ? { ...old, fields: [...(old.fields ?? []), field] } : old,
      );
    },
  });

  const update = useMutation({
    mutationFn: ({ fieldId, data }: { fieldId: string; data: UpdateFieldPayload }) =>
      fieldsApi.update(fieldId, data),
    onMutate: async ({ fieldId, data }) => {
      await qc.cancelQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      const prev = qc.getQueryData<Spreadsheet>(['spreadsheet', spreadsheetId]);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) =>
        old
          ? {
              ...old,
              fields: old.fields?.map((f) =>
                f.id === fieldId ? { ...f, ...data } : f,
              ),
            }
          : old,
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(['spreadsheet', spreadsheetId], ctx.prev);
    },
    onSuccess: (field) => {
      updateField(field);
    },
  });

  const reorder = useMutation({
    mutationFn: (fieldIds: string[]) =>
      fieldsApi.reorder(spreadsheetId, fieldIds),
    onMutate: async (fieldIds) => {
      await qc.cancelQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      const prev = qc.getQueryData<Spreadsheet>(['spreadsheet', spreadsheetId]);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) => {
        if (!old?.fields) return old;
        const map = Object.fromEntries(old.fields.map((f) => [f.id, f]));
        return {
          ...old,
          fields: fieldIds
            .filter((id) => map[id])
            .map((id, index) => ({ ...map[id], position: index })),
        };
      });
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(['spreadsheet', spreadsheetId], ctx.prev);
    },
    onSettled: () =>
      qc.invalidateQueries({ queryKey: ['spreadsheet', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (fieldId: string) => fieldsApi.delete(fieldId),
    onSuccess: (_data, fieldId) => {
      removeField(fieldId);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) =>
        old
          ? { ...old, fields: old.fields?.filter((f) => f.id !== fieldId) }
          : old,
      );
    },
  });

  return { ...query, create, update, reorder, remove };
}
```

---

### Что изменилось по сравнению с предыдущей версией

| Было | Стало |
|---|---|
| `spreadsheetsApi.getById` | `spreadsheetsApi.get` |
| `spreadsheetsApi.create(workspaceId, { name })` | `spreadsheetsApi.create({ name, workspace_id })` |
| `fieldsApi.list` (отдельный запрос) | `select` из кэша `['spreadsheet', id]` |
| `fieldsApi.update(spreadsheetId, fieldId, data)` | `fieldsApi.update(fieldId, data)` |
| `fieldsApi.delete(spreadsheetId, fieldId)` | `fieldsApi.delete(fieldId)` |
| `fieldsApi.reorder(spreadsheetId, orderedIds)` | `fieldsApi.reorder(spreadsheetId, fieldIds)` с `field_ids` |

---

### Примеры использования

```tsx
import { useSpreadsheet, useSpreadsheets, useFields } from '../hooks/useSpreadsheet';

// Загрузка таблицы
const { data: spreadsheet, isLoading } = useSpreadsheet(spreadsheetId);

// Список таблиц воркспейса + создание
const { data: spreadsheets, create } = useSpreadsheets(workspaceId);
create.mutate({ name: 'Новая таблица', description: 'Описание' });

// Поля — берутся из кэша spreadsheet без лишнего запроса
const { data: fields = [], create, update, reorder, remove } = useFields(spreadsheetId);

// Создать поле
create.mutate({ name: 'Телефон', field_type: 'text', is_required: false });

// Обновить поле
update.mutate({ fieldId: field.id, data: { name: 'Новое имя' } });

// Изменить порядок (drag & drop)
reorder.mutate(['field-3', 'field-1', 'field-2']);

// Удалить поле
remove.mutate(field.id);
```

