import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionsApi } from '../api/permissions';
import { usePermissionStore } from '../store/permissionStore';
import { useAuthStore } from '../store/authStore';
import type {
  SpreadsheetAccess,
  FieldAccess,
  RowAccessRule,
  PermissionRole,
} from '../types';

// ─── Spreadsheet-level permissions ───────────────────────────────────────────

export function useSpreadsheetPermissions(spreadsheetId: string) {
  const qc = useQueryClient();
  const { setPermissions } = usePermissionStore();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId],
    queryFn: async () => {
      const data = await permissionsApi.getSpreadsheetAccess(spreadsheetId);
      setPermissions(spreadsheetId, data);
      return data;
    },
    staleTime: 60_000,
  });

  const upsert = useMutation({
    mutationFn: (payload: Omit<SpreadsheetAccess, 'id' | 'principal_name'>) =>
      permissionsApi.upsertSpreadsheetAccess(spreadsheetId, payload),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  const remove = useMutation({
    mutationFn: (principalId: string) =>
      permissionsApi.removeSpreadsheetAccess(spreadsheetId, principalId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId] }),
  });

  return { ...query, upsert, remove };
}

// ─── Field-level permissions ──────────────────────────────────────────────────

export function useFieldPermissions(spreadsheetId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId, 'fields'],
    queryFn: () => permissionsApi.getFieldAccess(spreadsheetId),
    staleTime: 60_000,
  });

  const upsert = useMutation({
    mutationFn: ({ fieldId, data }: { fieldId: string; data: Omit<FieldAccess, 'field_id'> }) =>
      permissionsApi.upsertFieldAccess(fieldId, data),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'fields'] }),
  });

  return { ...query, upsert };
}

// ─── Row-level access rules ───────────────────────────────────────────────────

export function useRowRules(spreadsheetId: string) {
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ['permissions', spreadsheetId, 'rows'],
    queryFn: () => permissionsApi.getRowRules(spreadsheetId),
    staleTime: 60_000,
  });

  const upsert = useMutation({
    mutationFn: (rule: Omit<RowAccessRule, 'id'>) =>
      permissionsApi.upsertRowRule(spreadsheetId, rule),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  const remove = useMutation({
    mutationFn: (ruleId: string) =>
      permissionsApi.deleteRowRule(spreadsheetId, ruleId),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['permissions', spreadsheetId, 'rows'] }),
  });

  return { ...query, upsert, remove };
}

// ─── Current user ability check ───────────────────────────────────────────────

export function useAbility(spreadsheetId: string) {
  const { user } = useAuthStore();
  const { getPermissions } = usePermissionStore();

  const permissions: SpreadsheetAccess[] = getPermissions(spreadsheetId);

  const currentAccess = permissions.find(
    (p) => p.principal_id === user?.id,
  );

  const role: PermissionRole | null = currentAccess?.role ?? null;

  const can = (action: AbilityAction): boolean => {
    if (!role) return false;
    return ROLE_ABILITIES[role].includes(action);
  };

  const canAny = (actions: AbilityAction[]): boolean =>
    actions.some(can);

  const canAll = (actions: AbilityAction[]): boolean =>
    actions.every(can);

  // Проверяется отдельно загруженный список FieldAccess
  const canEditField = (fieldId: string, fieldAccess: FieldAccess[]): boolean => {
    const fa = fieldAccess.find(
      (f) => f.field_id === fieldId && f.principal_id === user?.id,
    );
    if (fa?.access === 'none') return false;
    if (fa?.access === 'view') return false;
    return can('row:edit');
  };

  const canViewField = (fieldId: string, fieldAccess: FieldAccess[]): boolean => {
    const fa = fieldAccess.find(
      (f) => f.field_id === fieldId && f.principal_id === user?.id,
    );
    if (fa?.access === 'none') return false;
    return true;
  };

  return {
    role,
    can,
    canAny,
    canAll,
    canEditField,
    canViewField,
    isOwner:  role === 'owner',
    isEditor: role === 'editor',
    isViewer: role === 'viewer',
  };
}

// ─── Ability matrix ───────────────────────────────────────────────────────────

export type AbilityAction =
  | 'spreadsheet:rename'
  | 'spreadsheet:delete'
  | 'field:create'
  | 'field:edit'
  | 'field:delete'
  | 'row:create'
  | 'row:edit'
  | 'row:delete'
  | 'member:invite'
  | 'member:remove'
  | 'member:change-role'
  | 'field-access:edit'
  | 'row-rules:edit';

const ROLE_ABILITIES: Record<PermissionRole, AbilityAction[]> = {
  owner: [
    'spreadsheet:rename',
    'spreadsheet:delete',
    'field:create',
    'field:edit',
    'field:delete',
    'row:create',
    'row:edit',
    'row:delete',
    'member:invite',
    'member:remove',
    'member:change-role',
    'field-access:edit',
    'row-rules:edit',
  ],
  editor: [
    'field:create',
    'field:edit',
    'row:create',
    'row:edit',
    'row:delete',
  ],
  viewer: [],
};