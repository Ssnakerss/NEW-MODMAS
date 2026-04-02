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
        setTypeWarning(`Смена типа с «\${field.field_type}» на «\${newType}» может привести к потере данных!`);
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
      title={isEdit ? `Редактировать поле «\${field?.name}»` : 'Новое поле'}
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