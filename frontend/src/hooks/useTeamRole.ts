import { useQuery } from '@tanstack/react-query';
import { getTeam } from '../api/teams';
import { useAuthStore } from '../store/auth';

export interface UseTeamRoleResult {
  role: string | null;
  isLoading: boolean;
  canEdit: boolean;
  isCaptain: boolean;
  isCocaptain: boolean;
  isEditor: boolean;
}

export function useTeamRole(teamId: number): UseTeamRoleResult {
  const user = useAuthStore((s) => s.user);

  const shouldFetch = !isNaN(teamId) && teamId > 0 && !!user;

  const { data: teamData, isLoading } = useQuery({
    queryKey: ['team', teamId],
    queryFn: () => getTeam(teamId),
    enabled: shouldFetch,
  });

  if (!shouldFetch) {
    return { role: null, isLoading: false, canEdit: false, isCaptain: false, isCocaptain: false, isEditor: false };
  }

  const member = teamData?.members?.find((m) => m.user_id === user!.id);
  const role = member?.role ?? null;

  return {
    role,
    isLoading,
    canEdit: role === 'editor' || role === 'cocaptain' || role === 'captain',
    isCaptain: role === 'captain',
    isCocaptain: role === 'cocaptain',
    isEditor: role === 'editor',
  };
}
