import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { fieldsApi } from '../api/fields';
import { useSpreadsheetStore } from '../store/spreadsheetStore';
import type { Field, Spreadsheet } from '../types';
import type { CreateFieldPayload, UpdateFieldPayload } from '../api/fields';

// ─── Spreadsheet ──────────────────────────────────────────────────────────────

export function useSpreadsheet(spreadsheetId: string) {
  const qc = useQueryClient();
  const { setSpreadsheet } = useSpreadsheetStore();

  const query = useQuery({
    queryKey: ['spreadsheet', spreadsheetId],
    queryFn: async () => {
      const data = await spreadsheetsApi.get(spreadsheetId);
      setSpreadsheet(data);
      return data;
    },
    staleTime: 30_000,
  });

  const rename = useMutation({
    mutationFn: (name: string) =>
      spreadsheetsApi.update(spreadsheetId, { name }),
    onSuccess: (updated) => {
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], updated);
      setSpreadsheet(updated);
    },
  });

  const remove = useMutation({
    mutationFn: () => spreadsheetsApi.delete(spreadsheetId),
    onSuccess: () => {
      qc.removeQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      qc.invalidateQueries({ queryKey: ['spreadsheets'] });
    },
  });

  return { ...query, rename, remove };
}

// ─── Spreadsheets list ────────────────────────────────────────────────────────

export function useSpreadsheets(workspaceId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['spreadsheets', workspaceId],
    queryFn: () => spreadsheetsApi.list(workspaceId),
    staleTime: 30_000,
  });

  const create = useMutation({
    mutationFn: (payload: { name: string; description?: string }) =>
      spreadsheetsApi.create({ ...payload, workspace_id: workspaceId }),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['spreadsheets', workspaceId] }),
  });

  return { ...query, create };
}

// ─── Fields ───────────────────────────────────────────────────────────────────

export function useFields(spreadsheetId: string) {
  const qc = useQueryClient();
  const { addField, updateField, removeField } = useSpreadsheetStore();

  // Поля живут внутри spreadsheet — берём их оттуда
  const query = useQuery({
    queryKey: ['spreadsheet', spreadsheetId],
    select: (data: Spreadsheet) => data.fields ?? [],
    staleTime: 30_000,
  });

  const create = useMutation({
    mutationFn: (data: CreateFieldPayload) =>
      fieldsApi.create(spreadsheetId, data),
    onSuccess: (field) => {
      addField(field);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) =>
        old ? { ...old, fields: [...(old.fields ?? []), field] } : old,
      );
    },
  });

  const update = useMutation({
    mutationFn: ({ fieldId, data }: { fieldId: string; data: UpdateFieldPayload }) =>
      fieldsApi.update(fieldId, data),
    onMutate: async ({ fieldId, data }) => {
      await qc.cancelQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      const prev = qc.getQueryData<Spreadsheet>(['spreadsheet', spreadsheetId]);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) =>
        old
          ? {
              ...old,
              fields: old.fields?.map((f) =>
                f.id === fieldId ? { ...f, ...data } : f,
              ),
            }
          : old,
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(['spreadsheet', spreadsheetId], ctx.prev);
    },
    onSuccess: (field) => {
      updateField(field);
    },
  });

  const reorder = useMutation({
    mutationFn: (fieldIds: string[]) =>
      fieldsApi.reorder(spreadsheetId, fieldIds),
    onMutate: async (fieldIds) => {
      await qc.cancelQueries({ queryKey: ['spreadsheet', spreadsheetId] });
      const prev = qc.getQueryData<Spreadsheet>(['spreadsheet', spreadsheetId]);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) => {
        if (!old?.fields) return old;
        const map = Object.fromEntries(old.fields.map((f) => [f.id, f]));
        return {
          ...old,
          fields: fieldIds
            .filter((id) => map[id])
            .map((id, index) => ({ ...map[id], position: index })),
        };
      });
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev)
        qc.setQueryData(['spreadsheet', spreadsheetId], ctx.prev);
    },
    onSettled: () =>
      qc.invalidateQueries({ queryKey: ['spreadsheet', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (fieldId: string) => fieldsApi.delete(fieldId),
    onSuccess: (_data, fieldId) => {
      removeField(fieldId);
      qc.setQueryData<Spreadsheet>(['spreadsheet', spreadsheetId], (old) =>
        old
          ? { ...old, fields: old.fields?.filter((f) => f.id !== fieldId) }
          : old,
      );
    },
  });

  return { ...query, create, update, reorder, remove };
}