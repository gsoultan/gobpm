import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';
import { errorMessage } from '../services/shared/errors';

type SourcesResult = Awaited<ReturnType<typeof processService.listSources>>;

/** The directories this project syncs from. */
export const useSources = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['participantSources', currentProjectId],
    queryFn: ({ signal }) =>
      currentProjectId && token
        ? processService.listSources(currentProjectId, signal)
        : Promise.resolve({ sources: [], err: '' } as SourcesResult),
    enabled: !!currentProjectId && !!token,
  });
};

export const useSaveSource = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (source: Record<string, unknown>) =>
      currentProjectId
        ? processService.saveSource(currentProjectId, source)
        : Promise.reject(new Error('No project selected')),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['participantSources', currentProjectId] });
    },
  });
};

export const useDeleteSource = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (id: string) => processService.deleteSource(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['participantSources', currentProjectId] });
    },
    onError: (error: unknown) => {
      notifications.show({
        title: 'Could not remove that directory',
        message: errorMessage(error, 'It is still there.'),
        color: 'red',
      });
    },
  });
};

/**
 * Reads one directory now.
 *
 * Invalidates the participant list as well as the sources: a sync is the whole
 * point, and leaving the list stale would show the old directory next to a
 * fresh "synced" message.
 */
export const useSyncSource = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (id: string) => processService.syncSource(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['participantSources', currentProjectId] });
      queryClient.invalidateQueries({ queryKey: ['participants', currentProjectId] });
    },
  });
};
