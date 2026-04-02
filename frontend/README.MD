# Фронтенд — полная реализация

## Стек

* **React 18** + **TypeScript**
* **Zustand** — стейт-менеджмент
* **TanStack Query** — серверный стейт
* **TanStack Virtual** — виртуализация строк
* **Tailwind CSS** — стили

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
