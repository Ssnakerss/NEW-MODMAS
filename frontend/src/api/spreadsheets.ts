import { apiClient } from './client';
import type { Spreadsheet } from '../types';

interface CreateSpreadsheetPayload {
  name: string;
  description?: string;
  workspace_id: string;
}

export const spreadsheetsApi = {
  list: (workspaceId: string) =>
    apiClient.get<Spreadsheet[]>(`/workspaces/${workspaceId}/spreadsheets`).then(r => r.data),

  get: (id: string) =>
    apiClient.get<Spreadsheet>(`/spreadsheets/${id}`).then(r => r.data),

  create: (data: CreateSpreadsheetPayload) =>
    apiClient.post<Spreadsheet>('/spreadsheets', data).then(r => r.data),

  update: (id: string, data: Partial<Pick<Spreadsheet, 'name' | 'description'>>) =>
    apiClient.put<Spreadsheet>(`/spreadsheets/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    apiClient.delete(`/spreadsheets/${id}`),
};