import { apiClient } from './client';
import type { Field, FieldType, FieldOptions } from '../types';

export interface CreateFieldPayload {
  name: string;
  field_type: FieldType;
  is_required?: boolean;
  is_unique?: boolean;
  default_value?: string;
  options?: FieldOptions;
}

export interface UpdateFieldPayload extends Partial<CreateFieldPayload> {
  position?: number;
}

export const fieldsApi = {
  create: (spreadsheetId: string, data: CreateFieldPayload) =>
    apiClient.post<Field>(`/spreadsheets/${spreadsheetId}/fields`, data).then(r => r.data),

  update: (fieldId: string, data: UpdateFieldPayload) =>
    apiClient.put<Field>(`/fields/${fieldId}`, data).then(r => r.data),

  delete: (fieldId: string) =>
    apiClient.delete(`/fields/${fieldId}`),

  reorder: (spreadsheetId: string, fieldIds: string[]) =>
    apiClient.patch(`/spreadsheets/${spreadsheetId}/fields/reorder`, { field_ids: fieldIds }),
};