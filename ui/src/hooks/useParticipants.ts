import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';
import type { ImportSummary } from '../domain/participantImport';
import { errorMessage } from '../services/shared/errors';

type ParticipantsResult = Awaited<ReturnType<typeof processService.listParticipants>>;

/** The people this project's processes can assign work to. */
export const useParticipants = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['participants', currentProjectId],
    queryFn: ({ signal }) =>
      currentProjectId && token
        ? processService.listParticipants(currentProjectId, signal)
        : Promise.resolve({ participants: [], err: '' } as ParticipantsResult),
    enabled: !!currentProjectId && !!token,
  });
};

/**
 * Brings a directory in from a file, an API or a query.
 *
 * The reply is a summary rather than a bare success, because a partial import
 * is the normal case: the good rows land and the rest are reported. Treating it
 * as success-or-failure would throw away the half that says which rows did not.
 */
export const useImportParticipants = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation<ImportSummary, Error, Record<string, unknown>>({
    mutationFn: (source) =>
      currentProjectId
        ? processService.importParticipants(currentProjectId, source)
        : Promise.reject(new Error('No project selected')),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['participants', currentProjectId] });
    },
    onError: (error: unknown) => {
      notifications.show({
        title: 'The directory could not be read',
        message: errorMessage(error, 'Nothing was imported.'),
        color: 'red',
      });
    },
  });
};
