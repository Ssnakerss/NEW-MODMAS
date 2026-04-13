import { apiClient } from './client';
import type { Spreadsheet, Field } from '../types';

export interface SpreadsheetWithFields extends Omit<Spreadsheet, 'fields'> {
  fields: Field[];
}

interface CreateSpreadsheetPayload {
  name: string;
  description?: string;
  workspace_id: string;
}

export const spreadsheetsApi = {
  list: (workspaceId: string) =>
    apiClient.get<SpreadsheetWithFields[]>(`/workspaces/${workspaceId}/spreadsheets`).then(r => r.data),

  get: (id: string) =>
    apiClient.get<SpreadsheetWithFields>(`/spreadsheets/${id}`).then(r => r.data),

  create: (data: CreateSpreadsheetPayload) =>
    apiClient.post<SpreadsheetWithFields>('/spreadsheets', data).then(r => r.data),

  update: (id: string, data: Partial<Pick<SpreadsheetWithFields, 'name' | 'description'>>) =>
    apiClient.put<SpreadsheetWithFields>(`/spreadsheets/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    apiClient.delete(`/spreadsheets/${id}`),
};