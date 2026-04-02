# Создание структуры каталогов и файлов для backend/

$base = "backend"

# ─── Директории ───────────────────────────────────────────────────────────────

$dirs = @(
    "$base/cmd/server",
    "$base/config",
    "$base/migrations",
    "$base/internal/auth",
    "$base/internal/workspace",
    "$base/internal/spreadsheet",
    "$base/internal/field",
    "$base/internal/row",
    "$base/internal/permission",
    "$base/internal/ddl",
    "$base/internal/middleware",
    "$base/pkg/postgres",
    "$base/pkg/jwt",
    "$base/pkg/hasher",
    "$base/pkg/response"
)

foreach ($dir in $dirs) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    Write-Host "  [DIR]  $dir" -ForegroundColor Cyan
}

# ─── Файлы ────────────────────────────────────────────────────────────────────

$files = @(
    # cmd
    "$base/cmd/server/main.go",

    # config
    "$base/config/config.go",

    # migrations
    "$base/migrations/001_auth_schema.sql",
    "$base/migrations/002_meta_schema.sql",
    "$base/migrations/003_permissions_schema.sql",

    # internal/auth
    "$base/internal/auth/handler.go",
    "$base/internal/auth/service.go",
    "$base/internal/auth/repository.go",

    # internal/workspace
    "$base/internal/workspace/handler.go",
    "$base/internal/workspace/service.go",
    "$base/internal/workspace/repository.go",

    # internal/spreadsheet
    "$base/internal/spreadsheet/handler.go",
    "$base/internal/spreadsheet/service.go",
    "$base/internal/spreadsheet/repository.go",

    # internal/field
    "$base/internal/field/handler.go",
    "$base/internal/field/service.go",
    "$base/internal/field/repository.go",
    "$base/internal/field/types.go",

    # internal/row
    "$base/internal/row/handler.go",
    "$base/internal/row/service.go",
    "$base/internal/row/repository.go",

    # internal/permission
    "$base/internal/permission/handler.go",
    "$base/internal/permission/service.go",
    "$base/internal/permission/repository.go",
    "$base/internal/permission/enforcer.go",

    # internal/ddl
    "$base/internal/ddl/builder.go",
    "$base/internal/ddl/executor.go",

    # internal/middleware
    "$base/internal/middleware/auth.go",
    "$base/internal/middleware/cors.go",

    # pkg
    "$base/pkg/postgres/pool.go",
    "$base/pkg/jwt/jwt.go",
    "$base/pkg/hasher/hasher.go",
    "$base/pkg/response/response.go"
)

foreach ($file in $files) {
    New-Item -ItemType File -Path $file -Force | Out-Null
    Write-Host "  [FILE] $file" -ForegroundColor Green
}

Write-Host "`nГотово! Структура создана в '$base'" -ForegroundColor Yellow