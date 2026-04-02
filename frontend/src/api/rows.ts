import { apiClient } from './client';
import type { Row, PaginatedRows, FilterCondition, SortCondition } from '../types';

interface FetchRowsParams {
  limit?: number;
  offset?: number;
  filters?: FilterCondition[];
  sorts?: SortCondition[];
}

interface UpsertRowPayload {
  [column_name: string]: unknown;
}

export const rowsApi = {
  list: (spreadsheetId: string, params: FetchRowsParams = {}) =>
    apiClient.post<PaginatedRows>(`/spreadsheets/\${spreadsheetId}/rows/query`, params).then(r => r.data),

  create: (spreadsheetId: string, data: UpsertRowPayload) =>
    apiClient.post<Row>(`/spreadsheets/\${spreadsheetId}/rows`, data).then(r => r.data),

  update: (spreadsheetId: string, rowId: string, data: UpsertRowPayload) =>
    apiClient.patch<Row>(`/spreadsheets/\${spreadsheetId}/rows/\${rowId}`, data).then(r => r.data),

  delete: (spreadsheetId: string, rowId: string) =>
    apiClient.delete(`/spreadsheets/\${spreadsheetId}/rows/\${rowId}`),

  bulkDelete: (spreadsheetId: string, rowIds: string[]) =>
    apiClient.post(`/spreadsheets/\${spreadsheetId}/rows/bulk-delete`, { row_ids: rowIds }),
};