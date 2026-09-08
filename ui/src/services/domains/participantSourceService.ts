import { requestJSON } from "../shared/rest";
import { raiseIfRefused } from "../raise";
import type { ImportSummary, ParticipantSource } from "../../domain/participantImport";

interface ListSourcesResponse {
  sources?: ParticipantSource[];
  err?: string;
}

interface SaveSourceResponse {
  id?: string;
  err?: string;
}

interface SyncSourceResponse extends ImportSummary {
  deactivated?: number;
  err?: string;
}

/**
 * The directories a project keeps its participants in step with.
 *
 * Every call is administrative: a source holds a connection string or an
 * endpoint token, so even the list says where each directory lives.
 */
export const participantSourceService = {
  async listSources(projectId: string, signal?: AbortSignal) {
    const response = await requestJSON<ListSourcesResponse>(
      `/participant-sources?project_id=${encodeURIComponent(projectId)}`,
      { method: "GET", signal },
    );
    return { sources: response.sources ?? [], err: response.err };
  },

  async saveSource(projectId: string, source: Record<string, unknown>, signal?: AbortSignal) {
    const response = await requestJSON<SaveSourceResponse>("/participant-sources", {
      method: "POST",
      body: { project_id: projectId, ...source },
      signal,
    });
    return { id: raiseIfRefused(response).id };
  },

  async deleteSource(id: string, signal?: AbortSignal) {
    const response = await requestJSON<{ err?: string }>(
      `/participant-sources/${encodeURIComponent(id)}`,
      { method: "DELETE", signal },
    );
    return { err: raiseIfRefused(response).err };
  },

  /**
   * Reads one directory now.
   *
   * The same path a scheduled run takes, lock included — so pressing this while
   * a scheduled run is in flight is refused rather than doubling the work.
   */
  async syncSource(id: string, signal?: AbortSignal) {
    const response = await requestJSON<SyncSourceResponse>(
      `/participant-sources/${encodeURIComponent(id)}/sync`,
      { method: "POST", signal },
    );
    return raiseIfRefused(response) as ImportSummary & { deactivated?: number };
  },
};
