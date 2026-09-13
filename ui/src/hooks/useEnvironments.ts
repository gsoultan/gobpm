import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';
import type { ApiEnvironment } from '../services/types';
import { errorMessage } from '../services/shared/errors';

type EnvironmentsResult = Awaited<ReturnType<typeof processService.listEnvironments>>;

/**
 * A project's runtimes.
 *
 * Not cached hard: an administrator opening this page has usually just changed
 * something, or is about to, and a stale list of where the databases are is
 * worse than a second request.
 */
export const useEnvironments = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['environments', currentProjectId],
    queryFn: ({ signal }) =>
      currentProjectId && token
        ? processService.listEnvironments(currentProjectId, signal)
        : Promise.resolve({ environments: [], err: '' } as EnvironmentsResult),
    enabled: !!currentProjectId && !!token,
  });
};

export const useSaveEnvironment = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (environment: Partial<ApiEnvironment>) =>
      currentProjectId
        ? processService.saveEnvironment(currentProjectId, environment)
        : Promise.reject('No project selected'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environments', currentProjectId] });
    },
    onError: (error: unknown) => {
      notifications.show({
        title: 'Could not save this environment',
        message: errorMessage(error, 'It was not saved.'),
        color: 'red',
      });
    },
  });
};

export const useDeleteEnvironment = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (id: string) => processService.deleteEnvironment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['environments', currentProjectId] });
    },
    onError: (error: unknown) => {
      notifications.show({
        title: 'Could not remove this environment',
        message: errorMessage(error, 'It was not removed.'),
        color: 'red',
      });
    },
  });
};

/** Opens the described database and reports whether it answered. */
export const useTestEnvironmentConnection = () =>
  useMutation({
    mutationFn: (environment: { id?: string; driver?: string; connection?: Record<string, unknown> }) =>
      processService.testEnvironmentConnection(environment),
  });
