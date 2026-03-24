import { signalClient } from "../shared/connect";

export const signalService = {
  async broadcastSignal(projectId: string, signalName: string, variables: Record<string, unknown> = {}, signal?: AbortSignal) {
    const response = await signalClient.broadcastSignal({ projectId, signalName, variables }, { signal });
    return { err: response.error };
  },

  async sendMessage(projectId: string, messageName: string, correlationKey: string, variables: Record<string, unknown> = {}, signal?: AbortSignal) {
    const response = await signalClient.sendMessage({ projectId, messageName, correlationKey, variables }, { signal });
    return { err: response.error };
  },
};
