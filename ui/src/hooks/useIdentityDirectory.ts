import { useQuery } from '@tanstack/react-query';
import { processService } from '../services/api';

export const useUsers = (organizationId: string | null) => {
  return useQuery({
    queryKey: ['users', organizationId],
    queryFn: ({ signal }) =>
      organizationId ? processService.listUsers(organizationId, signal) : Promise.resolve({ users: [] }),
    enabled: !!organizationId,
  });
};

export const useUserGroups = (userId: string | null) => {
  return useQuery({
    queryKey: ['user-groups', userId],
    queryFn: ({ signal }) =>
      userId ? processService.listUserGroups(userId, signal) : Promise.resolve({ groups: [] }),
    enabled: !!userId,
  });
};

export const useGroups = (organizationId: string | null) => {
  return useQuery({
    queryKey: ['groups', organizationId],
    queryFn: ({ signal }) =>
      organizationId ? processService.listGroups(organizationId, signal) : Promise.resolve({ groups: [] }),
    enabled: !!organizationId,
  });
};