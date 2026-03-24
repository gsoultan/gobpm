import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';

export const useConnectors = () => {
  return useQuery({
    queryKey: ['connectors'],
    queryFn: ({ signal }) => processService.listConnectors(signal),
  });
};

export const useCreateConnector = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (connector: any) => processService.createConnector(connector),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connectors'] });
    },
  });
};

export const useUpdateConnector = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (connector: any) => processService.updateConnector(connector),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connectors'] });
    },
  });
};

export const useDeleteConnector = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => processService.deleteConnector(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connectors'] });
    },
  });
};

export const useConnectorInstances = () => {
  const { currentProjectId } = useAppStore();
  return useQuery({
    queryKey: ['connector-instances', currentProjectId],
    queryFn: ({ signal }) =>
      currentProjectId ? processService.listConnectorInstances(currentProjectId, signal) : Promise.resolve({ instances: [], err: "" }),
    enabled: !!currentProjectId,
  });
};

export const useCreateConnectorInstance = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (instance: any) => processService.createConnectorInstance(instance),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connector-instances', currentProjectId] });
    },
  });
};

export const useUpdateConnectorInstance = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (instance: any) => processService.updateConnectorInstance(instance),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connector-instances', currentProjectId] });
    },
  });
};

export const useDeleteConnectorInstance = () => {
  const queryClient = useQueryClient();
  const { currentProjectId } = useAppStore();
  return useMutation({
    mutationFn: (id: string) => processService.deleteConnectorInstance(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connector-instances', currentProjectId] });
    },
  });
};

export const useExecuteConnector = () => {
  return useMutation({
    mutationFn: ({ connectorKey, config, payload }: { connectorKey: string; config: any; payload: any }) =>
      processService.executeConnector(connectorKey, config, payload),
  });
};

export const useExecuteScript = () => {
  return useMutation({
    mutationFn: ({ script, scriptFormat, variables }: { script: string; scriptFormat: string; variables: any }) =>
      processService.executeScript(script, scriptFormat, variables),
  });
};