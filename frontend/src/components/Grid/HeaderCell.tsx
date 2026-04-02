import React, { useState } from 'react';
import type { Field } from '../../types';
import { Dropdown, type DropdownItem } from '../ui/Dropdown';
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

  const handleSelect = (item: DropdownItem) => {
    switch (item.value) {
      case 'sort-asc':
        onSort(field.id, 'asc');
        break;
      case 'sort-desc':
        onSort(field.id, 'desc');
        break;
      case 'edit':
        setEditOpen(true);
        break;
      case 'delete':
        onDelete(field.id);
        break;
      default:
        break;
    }
    setMenuOpen(false);
  };

  const menuItems = [
    { label: '↑ Сортировать А→Я', value: 'sort-asc' },
    { label: '↓ Сортировать Я→А', value: 'sort-desc' },
    ...(canManage ? [
      { label: '✏️ Редактировать поле', value: 'edit' },
      { label: '🗑️ Удалить поле', value: 'delete', danger: true },
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
        <Dropdown
          trigger={
            <button
              className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-gray-600 px-1 rounded"
              onClick={(e) => {
                e.stopPropagation();
                setMenuOpen(prev => !prev);
              }}
            >
              ⋮
            </button>
          }
          items={menuItems}
          onSelect={handleSelect}
          align="right"
          size="sm"
          open={menuOpen}
          onClose={() => setMenuOpen(false)}
        />
      </div>

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