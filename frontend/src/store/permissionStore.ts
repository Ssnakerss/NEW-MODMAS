import { create } from 'zustand';
import type { SpreadsheetAccess, FieldAccess, RowAccessRule } from '../types';

interface PermissionState {
  // Доступы к таблицам: spreadsheetId → список доступов
  spreadsheetAccess: Record<string, SpreadsheetAccess[]>;
  // Доступы к полям: spreadsheetId → список доступов
  fieldAccess: Record<string, FieldAccess[]>;
  // Правила доступа к строкам: spreadsheetId → список правил
  rowRules: Record<string, RowAccessRule[]>;

  // Spreadsheet access
  setPermissions: (spreadsheetId: string, access: SpreadsheetAccess[]) => void;
  getPermissions: (spreadsheetId: string) => SpreadsheetAccess[];
  upsertPermission: (spreadsheetId: string, access: SpreadsheetAccess) => void;
  removePermission: (spreadsheetId: string, principalId: string) => void;

  // Field access
  setFieldAccess: (spreadsheetId: string, access: FieldAccess[]) => void;
  getFieldAccess: (spreadsheetId: string) => FieldAccess[];
  upsertFieldAccess: (spreadsheetId: string, access: FieldAccess) => void;

  // Row rules
  setRowRules: (spreadsheetId: string, rules: RowAccessRule[]) => void;
  getRowRules: (spreadsheetId: string) => RowAccessRule[];
  upsertRowRule: (spreadsheetId: string, rule: RowAccessRule) => void;
  removeRowRule: (spreadsheetId: string, ruleId: string) => void;

  // Сброс при смене таблицы
  clearSpreadsheet: (spreadsheetId: string) => void;
  clearAll: () => void;
}

export const usePermissionStore = create<PermissionState>((set, get) => ({
  spreadsheetAccess: {},
  fieldAccess: {},
  rowRules: {},

  // ─── Spreadsheet access ───────────────────────────────────────────────────

  setPermissions: (spreadsheetId, access) =>
    set((s) => ({
      spreadsheetAccess: { ...s.spreadsheetAccess, [spreadsheetId]: access },
    })),

  getPermissions: (spreadsheetId) =>
    get().spreadsheetAccess[spreadsheetId] ?? [],

  upsertPermission: (spreadsheetId, access) =>
    set((s) => {
      const current = s.spreadsheetAccess[spreadsheetId] ?? [];
      const exists = current.some((a) => a.principal_id === access.principal_id);
      return {
        spreadsheetAccess: {
          ...s.spreadsheetAccess,
          [spreadsheetId]: exists
            ? current.map((a) => (a.principal_id === access.principal_id ? access : a))
            : [...current, access],
        },
      };
    }),

  removePermission: (spreadsheetId, principalId) =>
    set((s) => ({
      spreadsheetAccess: {
        ...s.spreadsheetAccess,
        [spreadsheetId]: (s.spreadsheetAccess[spreadsheetId] ?? []).filter(
          (a) => a.principal_id !== principalId,
        ),
      },
    })),

  // ─── Field access ─────────────────────────────────────────────────────────

  setFieldAccess: (spreadsheetId, access) =>
    set((s) => ({
      fieldAccess: { ...s.fieldAccess, [spreadsheetId]: access },
    })),

  getFieldAccess: (spreadsheetId) =>
    get().fieldAccess[spreadsheetId] ?? [],

  upsertFieldAccess: (spreadsheetId, access) =>
    set((s) => {
      const current = s.fieldAccess[spreadsheetId] ?? [];
      const exists = current.some(
        (a) => a.field_id === access.field_id && a.principal_id === access.principal_id,
      );
      return {
        fieldAccess: {
          ...s.fieldAccess,
          [spreadsheetId]: exists
            ? current.map((a) =>
                a.field_id === access.field_id && a.principal_id === access.principal_id
                  ? access
                  : a,
              )
            : [...current, access],
        },
      };
    }),

  // ─── Row rules ────────────────────────────────────────────────────────────

  setRowRules: (spreadsheetId, rules) =>
    set((s) => ({
      rowRules: { ...s.rowRules, [spreadsheetId]: rules },
    })),

  getRowRules: (spreadsheetId) =>
    get().rowRules[spreadsheetId] ?? [],

  upsertRowRule: (spreadsheetId, rule) =>
    set((s) => {
      const current = s.rowRules[spreadsheetId] ?? [];
      const exists = current.some((r) => r.id === rule.id);
      return {
        rowRules: {
          ...s.rowRules,
          [spreadsheetId]: exists
            ? current.map((r) => (r.id === rule.id ? rule : r))
            : [...current, rule],
        },
      };
    }),

  removeRowRule: (spreadsheetId, ruleId) =>
    set((s) => ({
      rowRules: {
        ...s.rowRules,
        [spreadsheetId]: (s.rowRules[spreadsheetId] ?? []).filter(
          (r) => r.id !== ruleId,
        ),
      },
    })),

  // ─── Reset ────────────────────────────────────────────────────────────────

  clearSpreadsheet: (spreadsheetId) =>
    set((s) => {
      const spreadsheetAccess = { ...s.spreadsheetAccess };
      const fieldAccess = { ...s.fieldAccess };
      const rowRules = { ...s.rowRules };
      delete spreadsheetAccess[spreadsheetId];
      delete fieldAccess[spreadsheetId];
      delete rowRules[spreadsheetId];
      return { spreadsheetAccess, fieldAccess, rowRules };
    }),

  clearAll: () =>
    set({ spreadsheetAccess: {}, fieldAccess: {}, rowRules: {} }),
}));