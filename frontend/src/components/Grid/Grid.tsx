import React, { useCallback, useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import type { Field, Row } from '../../types';
import { Cell } from './Cell';
import { HeaderCell } from './HeaderCell';
import { AddColumnButton } from './AddColumnButton';

const ROW_HEIGHT = 36;
const INDEX_COL_WIDTH = 48;
const DEFAULT_COL_WIDTH = 200;

interface Props {
  fields: Field[];
  rows: Row[];
  canEdit: boolean;
  canManage: boolean;
  editableFields: Set<string>;
  selectedRowIds: Set<string>;
  hasNextPage?: boolean;
  isFetchingNextPage?: boolean;
  onFetchNextPage: () => void;
  onUpdateCell: (rowId: string, colName: string, value: unknown) => void;
  onAddRow: () => void;
  onDeleteRow: (rowId: string) => void;
  onToggleRow: (rowId: string) => void;
  onUpdateField: (field: Field) => void;
  onDeleteField: (fieldId: string) => void;
  onAddField: () => void;
  onSort: (fieldId: string, direction: 'asc' | 'desc') => void;
}

export const Grid: React.FC<Props> = ({
  fields, rows, canEdit, canManage, editableFields,
  selectedRowIds, hasNextPage, isFetchingNextPage,
  onFetchNextPage, onUpdateCell, onAddRow, onDeleteRow,
  onToggleRow, onUpdateField, onDeleteField, onAddField, onSort,
}) => {
  const parentRef = useRef<HTMLDivElement>(null);

  const totalWidth = INDEX_COL_WIDTH + fields.length * DEFAULT_COL_WIDTH + 80;

  const virtualizer = useVirtualizer({
    count: hasNextPage ? rows.length + 1 : rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  });

  const handleScroll = useCallback(() => {
    const el = parentRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 200;
    if (nearBottom && hasNextPage && !isFetchingNextPage) onFetchNextPage();
  }, [hasNextPage, isFetchingNextPage, onFetchNextPage]);

  return (
    <div className="flex flex-col h-full border border-gray-200 rounded-lg overflow-hidden">
      {/* Header */}
      <div
        className="flex border-b border-gray-200 bg-gray-50 sticky top-0 z-10"
        style={{ minWidth: totalWidth }}
      >
        {/* Index column */}
        <div
          className="flex items-center justify-center border-r border-gray-200 bg-gray-50 shrink-0"
          style={{ width: INDEX_COL_WIDTH }}
        >
          <input
            type="checkbox"
            onChange={(e) => rows.forEach(r => {
              if (e.target.checked !== selectedRowIds.has(r._id)) onToggleRow(r._id);
            })}
            checked={rows.length > 0 && rows.every(r => selectedRowIds.has(r._id))}
            className="w-3.5 h-3.5"
          />
        </div>

        {/* Field headers */}
        {fields.map((field) => (
          <div key={field.id} style={{ width: DEFAULT_COL_WIDTH, minWidth: DEFAULT_COL_WIDTH }}>
            <HeaderCell
              field={field}
              canManage={canManage}
              onUpdate={onUpdateField}
              onDelete={onDeleteField}
              onSort={onSort}
            />
          </div>
        ))}

        {canManage && <AddColumnButton onClick={onAddField} />}
      </div>

      {/* Body */}
      <div
        ref={parentRef}
        onScroll={handleScroll}
        className="flex-1 overflow-auto"
        style={{ minWidth: totalWidth }}
      >
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((vItem) => {
            const isLoader = vItem.index === rows.length;

            if (isLoader) {
              return (
                <div
                  key="loader"
                  style={{ position: 'absolute', top: vItem.start, height: ROW_HEIGHT, width: '100%' }}
                  className="flex items-center justify-center text-sm text-gray-400"
                >
                  Загрузка…
                </div>
              );
            }

            const row = rows[vItem.index];
            const isSelected = selectedRowIds.has(row._id);

            return (
              <div
                key={row._id}
                style={{ position: 'absolute', top: vItem.start, height: ROW_HEIGHT, width: '100%' }}
                className={`flex border-b border-gray-100 group ${isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'}`}
              >
                {/* Index + checkbox */}
                <div
                  className="flex items-center justify-center gap-1 border-r border-gray-200 shrink-0 text-xs text-gray-400"
                  style={{ width: INDEX_COL_WIDTH }}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    onChange={() => onToggleRow(row._id)}
                    className="w-3.5 h-3.5 opacity-0 group-hover:opacity-100"
                    onClick={(e) => e.stopPropagation()}
                  />
                  <span className="group-hover:hidden">{vItem.index + 1}</span>
                </div>

                {/* Cells */}
                {fields.map((field) => (
                  <div
                    key={field.id}
                    style={{ width: DEFAULT_COL_WIDTH, minWidth: DEFAULT_COL_WIDTH }}
                  >
                    <Cell
                      row={row}
                      field={field}
                      canEdit={canEdit && editableFields.has(field.id)}
                      onUpdate={onUpdateCell}
                    />
                  </div>
                ))}

                {/* Row actions */}
                {canEdit && (
                  <div className="flex items-center px-1 opacity-0 group-hover:opacity-100">
                    <button
                      onClick={() => onDeleteRow(row._id)}
                      className="text-red-400 hover:text-red-600 text-xs px-1"
                      title="Удалить строку"
                    >
                      ✕
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Footer — add row */}
      {canEdit && (
        <div className="border-t border-gray-200">
          <button
            onClick={onAddRow}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-50 w-full"
          >
            <span className="text-lg leading-none">+</span> Добавить строку
          </button>
        </div>
      )}
    </div>
  );
};