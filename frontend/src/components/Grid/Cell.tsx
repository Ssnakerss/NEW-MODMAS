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
        <span className={`text-lg \${rawValue ? 'text-green-500' : 'text-gray-300'}`}>
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
        \${isEditing ? 'p-0' : ''}
        \${canEdit && !isEditing ? 'cursor-pointer hover:bg-blue-50' : ''}
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