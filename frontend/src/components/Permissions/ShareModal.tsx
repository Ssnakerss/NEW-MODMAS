import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../../api/permissions';
import type { SpreadsheetAccess } from '../../types';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { FieldAccessMatrix } from './FieldAccessMatrix';
import { RowRuleBuilder } from './RowRuleBuilder';
import type { Field } from '../../types';

type Tab = 'users' | 'fields' | 'rows';

interface Props {
  spreadsheetId: string;
  fields: Field[];
  onClose: () => void;
}

export const ShareModal: React.FC<Props> = ({ spreadsheetId, fields, onClose }) => {
  const [tab, setTab] = useState<Tab>('users');
  const [newEmail, setNewEmail] = useState('');
  const qc = useQueryClient();

  const { data: accesses = [], isLoading } = useQuery({
    queryKey: ['permissions', spreadsheetId],
    queryFn: () => permissionsApi.getSpreadsheetAccess(spreadsheetId),
  });

  const upsert = useMutation({
    mutationFn: (data: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
      permissionsApi.upsertSpreadsheetAccess(spreadsheetId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (principalId: string) =>
      permissionsApi.removeSpreadsheetAccess(spreadsheetId, principalId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const handleInvite = () => {
    if (!newEmail.trim()) return;
    // В реальности — поиск пользователя по email, затем upsert
    setNewEmail('');
  };

  const togglePerm = (access: SpreadsheetAccess, perm: keyof SpreadsheetAccess) => {
    upsert.mutate({
      principal_id: access.principal_id,
      principal_type: access.principal_type,
      can_view:   access.can_view,
      can_insert: access.can_insert,
      can_edit:   access.can_edit,
      can_delete: access.can_delete,
      can_manage: access.can_manage,
      [perm]: !access[perm],
    });
  };

  const TABS: { id: Tab; label: string }[] = [
    { id: 'users',  label: '👥 Пользователи' },
    { id: 'fields', label: '🔒 Поля' },
    { id: 'rows',   label: '📋 Строки' },
  ];

  const PERMS: { key: keyof SpreadsheetAccess; label: string }[] = [
    { key: 'can_view',   label: 'Просмотр' },
    { key: 'can_insert', label: 'Добавление' },
    { key: 'can_edit',   label: 'Редактирование' },
    { key: 'can_delete', label: 'Удаление' },
    { key: 'can_manage', label: 'Управление' },
  ];

  return (
    <Modal open onClose={onClose} title="Настройка доступа" size="xl">
      {/* Tabs */}
      <div className="flex border-b mb-4 -mx-6 px-6 gap-1">
        {TABS.map(t => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 text-sm font-medium rounded-t transition-colors
              \${tab === t.id ? 'border-b-2 border-blue-600 text-blue-700' : 'text-gray-600 hover:text-gray-900'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Users tab */}
      {tab === 'users' && (
        <div className="space-y-4">
          <div className="flex gap-2">
            <input
              value={newEmail}
              onChange={(e) => setNewEmail(e.target.value)}
              placeholder="Email пользователя..."
              className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
            />
            <Button onClick={handleInvite} size="sm">Пригласить</Button>
          </div>

          {isLoading ? (
            <p className="text-sm text-gray-400 text-center py-4">Загрузка...</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 pr-4 font-medium text-gray-600">Пользователь</th>
                    {PERMS.map(p => (
                      <th key={p.key} className="text-center py-2 px-2 font-medium text-gray-600 text-xs">
                        {p.label}
                      </th>
                    ))}
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {accesses.map((access) => (
                    <tr key={access.principal_id} className="border-b hover:bg-gray-50">
                      <td className="py-2 pr-4">
                        <span className="font-medium text-gray-800">{access.principal_name}</span>
                        <span className="ml-2 text-xs text-gray-400">{access.principal_type}</span>
                      </td>
                      {PERMS.map(p => (
                        <td key={p.key} className="text-center py-2 px-2">
                          <input
                            type="checkbox"
                            checked={Boolean(access[p.key])}
                            onChange={() => togglePerm(access, p.key)}
                            className="w-4 h-4 cursor-pointer"
                          />
                        </td>
                      ))}
                      <td className="py-2 pl-2">
                        <button
                          onClick={() => remove.mutate(access.principal_id)}
                          className="text-red-400 hover:text-red-600 text-xs"
                        >
                          Удалить
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {tab === 'fields' && (
        <FieldAccessMatrix spreadsheetId={spreadsheetId} fields={fields} accesses={accesses} />
      )}

      {tab === 'rows' && (
        <RowRuleBuilder spreadsheetId={spreadsheetId} fields={fields} />
      )}
    </Modal>
  );
};