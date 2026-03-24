import { useQuery } from '@tanstack/react-query';
import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';

export const useProcessStatistics = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['stats', currentProjectId],
    queryFn: ({ signal }) =>
      (currentProjectId && token) ? processService.getProcessStatistics(currentProjectId, signal) : Promise.resolve({ stats: null, err: "" }),
    enabled: !!currentProjectId && !!token,
  });
};

export const useInstance = (id: string | null) => {
  return useQuery({
    queryKey: ['instance', id],
    queryFn: ({ signal }) =>
      id ? processService.getInstance(id, signal) : Promise.resolve({ instance: null, err: "" }),
    enabled: !!id,
  });
};

export const useExecutionPath = (id: string | null) => {
  return useQuery({
    queryKey: ['execution-path', id],
    queryFn: ({ signal }) =>
      id ? processService.getExecutionPath(id, signal) : Promise.resolve({ nodes: [], node_frequencies: {}, err: "" }),
    enabled: !!id,
  });
};

export const useAuditLogs = (id: string | null) => {
  return useQuery({
    queryKey: ['audit-logs', id],
    queryFn: ({ signal }) =>
      id ? processService.getAuditLogs(id, signal) : Promise.resolve({ entries: [], err: "" }),
    enabled: !!id,
  });
};

export const useSubProcesses = (parentInstanceId: string | null) => {
  return useQuery({
    queryKey: ['subProcesses', parentInstanceId],
    queryFn: ({ signal }) =>
      parentInstanceId ? processService.listSubProcesses(parentInstanceId, signal) : Promise.resolve({ instances: [], err: "" }),
    enabled: !!parentInstanceId,
  });
};

export const useInstances = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['instances', currentProjectId],
    queryFn: ({ signal }) =>
      (currentProjectId && token) ? processService.listInstances(currentProjectId, signal) : Promise.resolve({ instances: [], err: "" }),
    enabled: !!currentProjectId && !!token,
  });
};