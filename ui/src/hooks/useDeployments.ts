import { useQuery } from '@tanstack/react-query';
import { processService } from '../services/api';
import { useAppStore } from '../store/useAppStore';

export const useDeployments = () => {
  const { currentProjectId, token } = useAppStore();
  return useQuery({
    queryKey: ['deployments', currentProjectId],
    queryFn: ({ signal }) =>
      (currentProjectId && token) ? processService.listDeployments(currentProjectId, signal) : Promise.resolve({ deployments: [], err: "" }),
    enabled: !!currentProjectId && !!token,
  });
};
