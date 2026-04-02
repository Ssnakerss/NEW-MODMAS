import React, { useState, useCallback, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { spreadsheetsApi } from '../api/spreadsheets';
import { fieldsApi } from '../api/fields';
import { useRows } from '../hooks/useRows';
import { useSpreadsheetStore } from '../store/spreadsheetStore';
import { Grid } from '../components/Grid/Grid';
import { FieldEditorModal } from '../components/FieldEditor/FieldEditorModal';
import { ShareModal } from '../components/Permissions/ShareModal';
import { Button } from '../components/ui/Button';
import type { Field } from '../types';

export const SpreadsheetPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const {
    spreadsheet, setSpreadsheet,
    addField, updateField, removeField,
    selectedRowIds, toggleRowSelection, clearSelection,
    filters, sorts, setFilters, setSorts,
  } = useSpreadsheetStore();

  const [addFieldOpen, setAddFieldOpen] = useState(false);
  const [shareOpen, setShareOpen] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['spreadsheet', id],
    queryFn: () => spreadsheetsApi.get(id!),
    enabled: Boolean(id),
  });
  useEffect(() => {
    if (data) setSpreadsheet(data);
  }, [data]);

  const { allRows, fetchNextPage, hasNextPage, isFetchingNextPage, createRow, updateRow, deleteRow } = useRows(id!);

  const handleUpdateCell = useCallback(async (rowId: string, colName: string, value: unknown) => {
    await updateRow.mutateAsync({ rowId, data: { [colName]: value } });
  }, [updateRow]);

  const handleAddRow = useCallback(async () => {
    await createRow.mutateAsync({});
  }, [createRow]);

  const handleDeleteRow = useCallback(async (rowId: string) => {
    await deleteRow.mutateAsync(rowId);
  }, [deleteRow]);

  const handleDeleteField = useCallback(async (fieldId: string) => {
    if (!window.confirm('Удалить поле и все данные в нём?')) return;
    await fieldsApi.delete(fieldId);
    removeField(fieldId);
  }, [removeField]);

  const handleSort = useCallback((fieldId: string, direction: 'asc' | 'desc') => {
    setSorts([{ field_id: fieldId, direction }]);
  }, [setSorts]);

  if (isLoading || !spreadsheet) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full" />
      </div>
    );
  }

  const sortedFields = [...spreadsheet.fields].sort((a, b) => a.position - b.position);

  return (
    <div className="flex flex-col h-screen bg-white">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white shrink-0">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-gray-900">{spreadsheet.name}</h1>
          <span className="text-sm text-gray-400">{allRows.length} строк</span>
        </div>

        <div className="flex items-center gap-2">
          {selectedRowIds.size > 0 && (
            <div className="flex items-center gap-2">
              <span className="text-sm text-blue-700 bg-blue-50 px-2 py-1 rounded">
                Выбрано: {selectedRowIds.size}
              </span>
              <Button
                variant="danger"
                size="sm"
                onClick={async () => {
                  const results = await Promise.allSettled(
                    Array.from(selectedRowIds).map(rowId => deleteRow.mutateAsync(rowId))
                  );
                  const failed = results.filter(r => r.status === 'rejected');
                  if (failed.length === 0) {
                    clearSelection();
                  } else {
                    console.error(`Failed to delete ${failed.length} rows`, failed);
                    // Можно показать toast, но пока просто оставляем выбор
                  }
                }}
              >
                Удалить выбранные
              </Button>
              <Button variant="ghost" size="sm" onClick={clearSelection}>Отмена</Button>
            </div>
          )}
          <Button variant="secondary" size="sm" onClick={() => setShareOpen(true)}>
            🔒 Доступ
          </Button>
          <Button size="sm" onClick={handleAddRow} loading={createRow.isPending}>
            + Строка
          </Button>
        </div>
      </div>

      {/* Grid */}
      <div className="flex-1 overflow-hidden p-4">
        <Grid
          fields={sortedFields}
          rows={allRows}
          canEdit={true}
          canManage={true}
          editableFields={new Set(sortedFields.map(f => f.id))}
          selectedRowIds={selectedRowIds}
          hasNextPage={hasNextPage}
          isFetchingNextPage={isFetchingNextPage}
          onFetchNextPage={fetchNextPage}
          onUpdateCell={handleUpdateCell}
          onAddRow={handleAddRow}
          onDeleteRow={handleDeleteRow}
          onToggleRow={toggleRowSelection}
          onUpdateField={(field: Field) => { updateField(field); }}
          onDeleteField={handleDeleteField}
          onAddField={() => setAddFieldOpen(true)}
          onSort={handleSort}
        />
      </div>

      {/* Modals */}
      {addFieldOpen && (
        <FieldEditorModal
          spreadsheetId={spreadsheet.id}
          onClose={() => setAddFieldOpen(false)}
          onSave={(field) => { addField(field); setAddFieldOpen(false); }}
        />
      )}
      {shareOpen && (
        <ShareModal
          spreadsheetId={spreadsheet.id}
          fields={sortedFields}
          onClose={() => setShareOpen(false)}
        />
      )}
    </div>
  );
};