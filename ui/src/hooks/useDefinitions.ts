import { notifications } from '@mantine/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';

export const useDefinitions = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['definitions', currentProjectId],
    queryFn: ({ signal }) =>
      (currentProjectId && token) ? processService.listDefinitions(currentProjectId, signal) : Promise.resolve({ definitions: [], err: "" }),
    enabled: !!currentProjectId && !!token,
  });
};

export const useDefinition = (id: string | null) => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['definition', currentProjectId, id],
    queryFn: ({ signal }) =>
      (currentProjectId && id && token) ? processService.getDefinition(currentProjectId, id, signal) : Promise.resolve({ definition: null, err: "" }),
    enabled: !!currentProjectId && !!id && !!token,
  });
};

export const useCreateDefinition = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (definition: any) =>
      currentProjectId ? processService.createDefinition(currentProjectId, definition) : Promise.reject('No project selected'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['definitions', currentProjectId] });
    },
  });
};

export const useDeleteDefinition = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (id: string) => processService.deleteDefinition(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['definitions', currentProjectId] });
    },
  });
};

export const useExportDefinition = () => {
  return useMutation({
    mutationFn: (id: string) => processService.exportDefinition(id),
  });
};

export const useImportDefinition = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (xml: string) => processService.importDefinition(xml),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['definitions', currentProjectId] });
      notifications.show({
        title: 'Success',
        message: 'BPMN model imported successfully.',
        color: 'teal',
      });
    },
    onError: (error: any) => {
      notifications.show({
        title: 'Import Error',
        message: error.message || 'Failed to import BPMN model.',
        color: 'red',
      });
    }
  });
};