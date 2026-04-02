# Создание структуры каталогов и файлов для frontend/src/

$base = "frontend"

# ─── Директории ───────────────────────────────────────────────────────────────

$dirs = @(
    "$base/src/types",
    "$base/src/api",
    "$base/src/store",
    "$base/src/hooks",
    "$base/src/components/ui",
    "$base/src/components/Grid",
    "$base/src/components/CellEditors",
    "$base/src/components/FieldEditor",
    "$base/src/components/Permissions",
    "$base/src/pages"
)

foreach ($dir in $dirs) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    Write-Host "  [DIR]  $dir" -ForegroundColor Cyan
}

# ─── Файлы ────────────────────────────────────────────────────────────────────

$files = @(
    # src root
    "$base/src/main.tsx",
    "$base/src/App.tsx",

    # types
    "$base/src/types/index.ts",

    # api
    "$base/src/api/client.ts",
    "$base/src/api/auth.ts",
    "$base/src/api/spreadsheets.ts",
    "$base/src/api/fields.ts",
    "$base/src/api/rows.ts",
    "$base/src/api/permissions.ts",

    # store
    "$base/src/store/authStore.ts",
    "$base/src/store/spreadsheetStore.ts",
    "$base/src/store/permissionStore.ts",

    # hooks
    "$base/src/hooks/useSpreadsheet.ts",
    "$base/src/hooks/useRows.ts",
    "$base/src/hooks/usePermissions.ts",

    # components/ui
    "$base/src/components/ui/Button.tsx",
    "$base/src/components/ui/Input.tsx",
    "$base/src/components/ui/Modal.tsx",
    "$base/src/components/ui/Select.tsx",
    "$base/src/components/ui/Dropdown.tsx",

    # components/Grid
    "$base/src/components/Grid/Grid.tsx",
    "$base/src/components/Grid/HeaderCell.tsx",
    "$base/src/components/Grid/Cell.tsx",
    "$base/src/components/Grid/AddColumnButton.tsx",

    # components/CellEditors
    "$base/src/components/CellEditors/index.ts",
    "$base/src/components/CellEditors/TextEditor.tsx",
    "$base/src/components/CellEditors/NumberEditor.tsx",
    "$base/src/components/CellEditors/BooleanEditor.tsx",
    "$base/src/components/CellEditors/DateEditor.tsx",
    "$base/src/components/CellEditors/SelectEditor.tsx",
    "$base/src/components/CellEditors/RelationEditor.tsx",

    # components/FieldEditor
    "$base/src/components/FieldEditor/FieldEditorModal.tsx",
    "$base/src/components/FieldEditor/FieldTypeSelect.tsx",
    "$base/src/components/FieldEditor/SelectOptionsEditor.tsx",

    # components/Permissions
    "$base/src/components/Permissions/ShareModal.tsx",
    "$base/src/components/Permissions/FieldAccessMatrix.tsx",
    "$base/src/components/Permissions/RowRuleBuilder.tsx",

    # pages
    "$base/src/pages/LoginPage.tsx",
    "$base/src/pages/WorkspacesPage.tsx",
    "$base/src/pages/SpreadsheetPage.tsx"
)

foreach ($file in $files) {
    New-Item -ItemType File -Path $file -Force | Out-Null
    Write-Host "  [FILE] $file" -ForegroundColor Green
}

Write-Host "`nГотово! Структура создана в '$base'" -ForegroundColor Yellow