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
                  \${value === type.value
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