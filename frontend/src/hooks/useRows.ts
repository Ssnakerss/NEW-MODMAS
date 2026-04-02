import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rowsApi } from '../api/rows';
import { useSpreadsheetStore } from '../store/spreadsheetStore';

const PAGE_SIZE = 50;

export function useRows(spreadsheetId: string) {
  const { filters, sorts } = useSpreadsheetStore();
  const qc = useQueryClient();

  const query = useInfiniteQuery({
    queryKey: ['rows', spreadsheetId, filters, sorts],
    queryFn: ({ pageParam = 0 }) =>
      rowsApi.list(spreadsheetId, {
        limit: PAGE_SIZE,
        offset: pageParam as number,
        filters,
        sorts,
      }),
    getNextPageParam: (last, pages) => {
      const loaded = pages.length * PAGE_SIZE;
      return loaded < last.total ? loaded : undefined;
    },
    initialPageParam: 0,
  });

  const allRows = query.data?.pages.flatMap((p) => p.data) ?? [];

  const createRow = useMutation({
    mutationFn: (data: Record<string, unknown>) => rowsApi.create(spreadsheetId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  const updateRow = useMutation({
    mutationFn: ({ rowId, data }: { rowId: string; data: Record<string, unknown> }) =>
      rowsApi.update(spreadsheetId, rowId, data),
    onMutate: async ({ rowId, data }) => {
      await qc.cancelQueries({ queryKey: ['rows', spreadsheetId] });
      const prev = qc.getQueryData(['rows', spreadsheetId]);
      qc.setQueryData(['rows', spreadsheetId, filters, sorts], (old: typeof query.data) => ({
        ...old,
        pages: old?.pages.map((page) => ({
          ...page,
          data: page.data.map((row) => (row._id === rowId ? { ...row, ...data } : row)),
        })),
      }));
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(['rows', spreadsheetId], ctx.prev);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  const deleteRow = useMutation({
    mutationFn: (rowId: string) => rowsApi.delete(spreadsheetId, rowId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rows', spreadsheetId] }),
  });

  return { ...query, allRows, createRow, updateRow, deleteRow };
}