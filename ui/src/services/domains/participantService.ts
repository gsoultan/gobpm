import { API_BASE_URL } from "../shared/config";
import { getAuthHeaders } from "../shared/auth";
import { requestJSON } from "../shared/rest";
import { raiseIfRefused } from "../raise";
import type { ImportSummary, Participant } from "../../domain/participantImport";

interface ListParticipantsResponse {
  participants?: Participant[];
  err?: string;
}

interface ImportParticipantsResponse extends ImportSummary {
  err?: string;
}

export const participantService = {
  async listParticipants(projectId: string, signal?: AbortSignal) {
    const response = await requestJSON<ListParticipantsResponse>(
      `/participants?project_id=${encodeURIComponent(projectId)}`,
      { method: "GET", signal },
    );
    return { participants: response.participants ?? [], err: response.err };
  },

  /**
   * Imports a directory.
   *
   * A file goes as multipart because it is a file; an endpoint or a query goes
   * as JSON. One method rather than three, because the caller is answering one
   * question — where does this list come from — and the reply is the same
   * summary either way.
   */
  async importParticipants(projectId: string, source: Record<string, unknown>, signal?: AbortSignal) {
    if (source.kind === "csv" && source.file instanceof File) {
      const form = new FormData();
      form.append("project_id", projectId);
      form.append("file", source.file);
      // Not requestJSON: it sets a JSON content type, and multipart needs the
      // browser to set its own boundary.
      const response = await fetch(`${API_BASE_URL}/participants/import`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: form,
        signal,
      });
      const body = (await response.json().catch(() => ({}))) as ImportParticipantsResponse;
      if (!response.ok) {
        throw new Error(body.err || `Import failed (${response.status})`);
      }
      return raiseIfRefused(body) as ImportSummary;
    }

    const body = await requestJSON<ImportParticipantsResponse>("/participants/import", {
      method: "POST",
      body: { project_id: projectId, ...source },
      signal,
    });
    return raiseIfRefused(body) as ImportSummary;
  },
};
