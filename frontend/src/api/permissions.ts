import { apiClient } from './client';
import type { SpreadsheetAccess, FieldAccess, RowAccessRule } from '../types';

export const permissionsApi = {
  getSpreadsheetAccess: (spreadsheetId: string) =>
    apiClient.get<SpreadsheetAccess[]>(`/spreadsheets/${spreadsheetId}/permissions`).then(r => r.data),

  upsertSpreadsheetAccess: (spreadsheetId: string, data: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
    apiClient.put(`/spreadsheets/${spreadsheetId}/permissions`, data),

  removeSpreadsheetAccess: (spreadsheetId: string, principalId: string) =>
    apiClient.delete(`/spreadsheets/${spreadsheetId}/permissions/${principalId}`),

  getFieldAccess: (spreadsheetId: string) =>
    apiClient.get<FieldAccess[]>(`/spreadsheets/${spreadsheetId}/field-permissions`).then(r => r.data),

  upsertFieldAccess: (fieldId: string, data: Omit<FieldAccess, 'field_id'>) =>
    apiClient.put(`/fields/${fieldId}/permissions`, data),

  getRowRules: (spreadsheetId: string) =>
    apiClient.get<RowAccessRule[]>(`/spreadsheets/${spreadsheetId}/row-rules`).then(r => r.data),

  upsertRowRule: (spreadsheetId: string, data: Omit<RowAccessRule, 'id'>) =>
    apiClient.put(`/spreadsheets/${spreadsheetId}/row-rules`, data),

  deleteRowRule: (spreadsheetId: string, ruleId: string) =>
    apiClient.delete(`/spreadsheets/${spreadsheetId}/row-rules/${ruleId}`),
};