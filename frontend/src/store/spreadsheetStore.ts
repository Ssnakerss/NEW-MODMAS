import { create } from 'zustand';
import type { Spreadsheet, Field, FilterCondition, SortCondition } from '../types';

interface SpreadsheetState {
  spreadsheet: Spreadsheet | null;
  selectedRowIds: Set<string>;
  editingCell: { rowId: string; fieldId: string } | null;
  filters: FilterCondition[];
  sorts: SortCondition[];

  setSpreadsheet: (s: Spreadsheet) => void;
  updateField: (field: Field) => void;
  addField: (field: Field) => void;
  removeField: (fieldId: string) => void;

  setEditingCell: (rowId: string, fieldId: string) => void;
  clearEditingCell: () => void;

  toggleRowSelection: (rowId: string) => void;
  clearSelection: () => void;

  setFilters: (filters: FilterCondition[]) => void;
  setSorts: (sorts: SortCondition[]) => void;
}

export const useSpreadsheetStore = create<SpreadsheetState>((set) => ({
  spreadsheet: null,
  selectedRowIds: new Set(),
  editingCell: null,
  filters: [],
  sorts: [],

  setSpreadsheet: (spreadsheet) => set({ spreadsheet }),

  updateField: (field) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? {
            ...s.spreadsheet,
            fields: s.spreadsheet.fields.map((f) => (f.id === field.id ? field : f)),
          }
        : null,
    })),

  addField: (field) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? { ...s.spreadsheet, fields: [...s.spreadsheet.fields, field] }
        : null,
    })),

  removeField: (fieldId) =>
    set((s) => ({
      spreadsheet: s.spreadsheet
        ? { ...s.spreadsheet, fields: s.spreadsheet.fields.filter((f) => f.id !== fieldId) }
        : null,
    })),

  setEditingCell: (rowId, fieldId) => set({ editingCell: { rowId, fieldId } }),
  clearEditingCell: () => set({ editingCell: null }),

  toggleRowSelection: (rowId) =>
    set((s) => {
      const next = new Set(s.selectedRowIds);
      next.has(rowId) ? next.delete(rowId) : next.add(rowId);
      return { selectedRowIds: next };
    }),

  clearSelection: () => set({ selectedRowIds: new Set() }),

  setFilters: (filters) => set({ filters }),
  setSorts: (sorts) => set({ sorts }),
}));