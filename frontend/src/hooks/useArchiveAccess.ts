import { useQuery } from '@tanstack/react-query';
import { getArchivePermissions } from '../api/admin';
import { useAuthStore } from '../store/auth';
import type { ArchivePermissions } from '../api/admin';

export function useArchiveAccess(): {
  permissions: ArchivePermissions | undefined;
  isLoading: boolean;
  canRead: boolean;
  canWrite: boolean;
} {
  const authReady = !useAuthStore((s) => s.isLoading);

  const { data: permissions, isLoading } = useQuery({
    queryKey: ['archive-permissions'],
    queryFn: getArchivePermissions,
    staleTime: 60_000,
    enabled: authReady,
  });

  return {
    permissions,
    isLoading: isLoading || !authReady,
    canRead: permissions?.can_read ?? false,
    canWrite: permissions?.can_write ?? false,
  };
}
