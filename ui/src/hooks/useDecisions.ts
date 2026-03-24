import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';

export const useDecisions = () => {
  const { currentProjectId } = useAppStore();
  return useQuery({
    queryKey: ['decisions', currentProjectId],
    queryFn: ({ signal }) =>
      currentProjectId ? processService.listDecisions(currentProjectId, signal) : Promise.resolve({ decisions: [], err: "" }),
    enabled: !!currentProjectId,
  });
};

export const useDecision = (id: string | null) => {
  return useQuery({
    queryKey: ['decision', id],
    queryFn: ({ signal }) =>
      id ? processService.getDecision(id, signal) : Promise.resolve({ decision: null, err: "" }),
    enabled: !!id,
  });
};

export const useCreateDecision = () => {
  const { currentProjectId } = useAppStore();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params: any) => 
      processService.createDecision({ ...params, project_id: currentProjectId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['decisions'] });
    },
  });
};

export const useUpdateDecision = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...params }: any) => 
      processService.updateDecision(id, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['decisions'] });
      queryClient.invalidateQueries({ queryKey: ['decision'] });
    },
  });
};

export const useDeleteDecision = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => processService.deleteDecision(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['decisions'] });
    },
  });
};

export const useEvaluateDecision = () => {
  return useMutation({
    mutationFn: ({ key, variables, version }: { key: string, variables: any, version?: number }) => 
      processService.evaluateDecision(key, variables, version),
  });
};
