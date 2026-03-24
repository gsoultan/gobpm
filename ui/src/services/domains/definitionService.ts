import { definitionClient, statsClient } from "../shared/connect";
import { requestJSON } from "../shared/rest";
import type {
  CreateDefinitionPayload,
  ExportDefinitionResponse,
  ImportDefinitionResponse,
} from "../types";

export const definitionService = {
  async listDefinitions(projectId: string, signal?: AbortSignal) {
    const response = await definitionClient.listDefinitions({ projectId }, { signal });
    return { definitions: response.definitions ?? [], err: response.error };
  },

  async createDefinition(projectId: string, definition: CreateDefinitionPayload, signal?: AbortSignal) {
    const response = await definitionClient.createDefinition({
      projectId,
      key: definition?.key ?? "",
      name: definition?.name ?? "",
      nodes: definition?.nodes ?? [],
      flows: definition?.flows ?? [],
    }, { signal });

    return { id: response.id, err: response.error };
  },

  async getDefinition(_projectId: string, id: string, signal?: AbortSignal) {
    const response = await definitionClient.getDefinition({ id }, { signal });
    return { definition: response.definition, err: response.error };
  },

  async deleteDefinition(id: string, signal?: AbortSignal) {
    const response = await definitionClient.deleteDefinition({ id }, { signal });
    return { err: response.error };
  },

  async exportDefinition(id: string, signal?: AbortSignal) {
    return requestJSON<ExportDefinitionResponse>(`/definitions/${id}/export`, {
      method: "GET",
      signal,
    });
  },

  async importDefinition(xml: string, signal?: AbortSignal) {
    return requestJSON<ImportDefinitionResponse>("/definitions/import", {
      method: "POST",
      body: { xml: btoa(xml) },
      signal,
    });
  },

  async getProcessStatistics(projectId: string, signal?: AbortSignal) {
    const response = await statsClient.getProcessStatistics({ projectId }, { signal });
    return { stats: response, err: response.error };
  },
};
