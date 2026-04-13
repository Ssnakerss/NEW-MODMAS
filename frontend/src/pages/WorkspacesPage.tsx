import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { workspacesApi } from '../api/workspaces';
import { useAuthStore } from '../store/authStore';
import { Button } from '../components/ui/Button';
import { Modal } from '../components/ui/Modal';
import type { Workspace, Spreadsheet } from '../types';
import type { SpreadsheetWithFields } from '../api/spreadsheets';
import type { WorkspaceWithSpreadsheetCount } from '../api/workspaces';


export const WorkspacesPage: React.FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const qc = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [newName, setNewName] = useState('');
  const [selectedWorkspace, setSelectedWorkspace] = useState<WorkspaceWithSpreadsheetCount | null>(null);

  // Получаем список workspace пользователя
  const { data: workspaces = [] } = useQuery({
    queryKey: ['workspaces'],
    queryFn: workspacesApi.list,
    enabled: !!user,
  });

  // Получаем spreadsheets для выбранного workspace
  const { data: spreadsheets = [], isLoading } = useQuery({
    queryKey: ['spreadsheets', selectedWorkspace?.id],
    queryFn: () => selectedWorkspace ? spreadsheetsApi.list(selectedWorkspace.id) : Promise.resolve([]),
    enabled: !!selectedWorkspace,
  });

  // Создаем workspace
  const createWorkspace = useMutation({
    mutationFn: () => workspacesApi.create({ name: newName }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspaces'] });
      setCreateOpen(false);
      setNewName('');
    },
  });

  // Создаем spreadsheet в выбранном workspace
  const createSpreadsheet = useMutation({
    mutationFn: () => spreadsheetsApi.create({ name: newName, workspace_id: selectedWorkspace!.id }),
    onSuccess: (s: SpreadsheetWithFields) => {
      qc.invalidateQueries({ queryKey: ['spreadsheets'] });
      setCreateOpen(false);
      setNewName('');
      navigate(`/spreadsheet/${s.id}`);
    },
  });

  // При изменении списка workspace выбираем первый (или уже выбранный)
  useEffect(() => {
    if (workspaces.length > 0 && !selectedWorkspace) {
      setSelectedWorkspace(workspaces[0]);
    } else if (workspaces.length === 0) {
      setSelectedWorkspace(null);
    }
  }, [workspaces, selectedWorkspace]);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <span className="text-xl">📊</span>
          <span className="font-semibold text-gray-900">DataGrid</span>
          
          {/* Селектор workspace */}
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">Workspace:</label>
            <select 
              value={selectedWorkspace?.id || ''}
              onChange={(e) => {
                const workspace = workspaces.find(w => w.id === e.target.value);
                setSelectedWorkspace(workspace || null);
              }}
              className="border border-gray-300 rounded-lg px-3 py-1 text-sm outline-none focus:ring-2 focus:ring-blue-500"
              disabled={workspaces.length === 0}
            >
              {workspaces.map((workspace) => (
                <option key={workspace.id} value={workspace.id}>
                  {workspace.name}
                </option>
              ))}
            </select>
          </div>
        </div>
        
        <div className="flex items-center gap-3">
          <span className="text-sm text-gray-600">{user?.name}</span>
          <Button variant="ghost" size="sm" onClick={logout}>Выйти</Button>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold text-gray-900">
            {selectedWorkspace ? `Таблицы: ${selectedWorkspace.name}` : 'Workspaces'}
          </h2>
          <Button onClick={() => setCreateOpen(true)}>
            {workspaces.length === 0 ? '+ Новый workspace' : '+ Новая таблица'}
          </Button>
        </div>

        {!selectedWorkspace ? (
          <div className="text-center py-16">
            <div className="text-5xl mb-4">🏗️</div>
            <p className="text-gray-500 mb-4">Нет доступных workspace</p>
            <Button onClick={() => setCreateOpen(true)}>+ Создать workspace</Button>
          </div>
        ) : isLoading ? (
          <div className="grid grid-cols-3 gap-4">
            {[1,2,3].map(i => (
              <div key={i} className="h-32 bg-gray-200 animate-pulse rounded-xl" />
            ))}
          </div>
        ) : spreadsheets.length === 0 ? (
          <div className="text-center py-16">
            <div className="text-5xl mb-4">📋</div>
            <p className="text-gray-500">Таблиц ещё нет. Создайте первую!</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {spreadsheets.map((s) => (
              <button
                key={s.id}
                onClick={() => navigate(`/spreadsheet/${s.id}`)}
                className="text-left p-5 bg-white rounded-xl border border-gray-200 hover:border-blue-400 hover:shadow-md transition-all group"
              >
                <div className="text-2xl mb-3">📊</div>
                <div className="font-medium text-gray-900 group-hover:text-blue-700">{s.name}</div>
                {s.description && (
                  <div className="text-sm text-gray-500 mt-1 truncate">{s.description}</div>
                )}
                <div className="text-xs text-gray-400 mt-3">
                  {s.fields.length} полей · {new Date(s.updated_at).toLocaleDateString('ru-RU')}
                </div>
              </button>
            ))}
          </div>
        )}
      </main>

      <Modal 
        open={createOpen} 
        onClose={() => setCreateOpen(false)} 
        title={workspaces.length === 0 ? "Новый workspace" : "Новая таблица"} 
        size="sm"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Название</label>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') createSpreadsheet.mutate(); }}
                placeholder={workspaces.length === 0 ? "Например: Мой персональный workspace" : "Например: Задачи, Клиенты..."}
              autoFocus
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setCreateOpen(false)}>Отмена</Button>
            {workspaces.length === 0 ? (
              <Button onClick={() => createWorkspace.mutate()} loading={createWorkspace.isPending} disabled={!newName.trim()}>
                Создать workspace
              </Button>
            ) : (
              <Button onClick={() => createSpreadsheet.mutate()} loading={createSpreadsheet.isPending} disabled={!newName.trim()}>
                Создать таблицу
              </Button>
            )}
          </div>
        </div>
      </Modal>
    </div>
  );
};