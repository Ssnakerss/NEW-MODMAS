import { apiClient } from './client';
import type { Workspace } from '../types';

export interface CreateWorkspacePayload {
  name: string;
}

export interface WorkspaceWithSpreadsheetCount extends Workspace {
  spreadsheet_count?: number;
}

export const workspacesApi = {
  list: () =>
    apiClient.get<WorkspaceWithSpreadsheetCount[]>('/workspaces').then(r => r.data),

  get: (id: string) =>
    apiClient.get<Workspace>(`/workspaces/${id}`).then(r => r.data),

  create: (data: CreateWorkspacePayload) =>
    apiClient.post<Workspace>('/workspaces', data).then(r => r.data),
};