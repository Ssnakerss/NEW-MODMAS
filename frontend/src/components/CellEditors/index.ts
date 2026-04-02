import type { FieldType, Field } from '../../types';
import { TextEditor } from './TextEditor';
import { NumberEditor } from './NumberEditor';
import { BooleanEditor } from './BooleanEditor';
import { DateEditor } from './DateEditor';
import { SelectEditor } from './SelectEditor';
import { RelationEditor } from './RelationEditor';

export { TextEditor, NumberEditor, BooleanEditor, DateEditor, SelectEditor, RelationEditor };

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
    relation: RelationEditor,
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