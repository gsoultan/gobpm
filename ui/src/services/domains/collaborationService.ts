import { requestJSON } from "../shared/rest";

export const collaborationService = {
  async broadcastCollaboration(event: any, signal?: AbortSignal) {
    const data = await requestJSON<{ err?: string }>("/collaboration/broadcast", {
      method: "POST",
      body: { event },
      signal,
    });

    return { err: data.err };
  },
};
