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