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