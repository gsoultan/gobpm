import { requestJSON } from "../shared/rest";

type DecisionListResponse = {
  decisions?: any[];
  err?: string;
};

type DecisionResponse = {
  decision?: any;
  err?: string;
};

type CreateDecisionResponse = {
  id: string;
  err?: string;
};

type EvaluateDecisionResponse = {
  result?: any;
  err?: string;
};

export const decisionService = {
  async listDecisions(projectId: string, signal?: AbortSignal) {
    const data = await requestJSON<DecisionListResponse>(`/decisions?project_id=${projectId}`, { signal });
    return { decisions: data.decisions ?? [], err: data.err };
  },

  async getDecision(id: string, signal?: AbortSignal) {
    const data = await requestJSON<DecisionResponse>(`/decisions/${id}`, { signal });
    return { decision: data.decision, err: data.err };
  },

  async createDecision(params: any) {
    const data = await requestJSON<CreateDecisionResponse>("/decisions", {
      method: "POST",
      body: { decision: params },
    });
    return { id: data.id, err: data.err };
  },

  async updateDecision(id: string, params: any) {
    const data = await requestJSON<any>(`/decisions/${id}`, {
      method: "PUT",
      body: { decision: params },
    });
    return { err: data.err };
  },

  async deleteDecision(id: string) {
    const data = await requestJSON<any>(`/decisions/${id}`, {
      method: "DELETE",
    });
    return { err: data.err };
  },

  async evaluateDecision(key: string, variables: any = {}, version: number = 0, signal?: AbortSignal) {
    const data = await requestJSON<EvaluateDecisionResponse>("/decisions/evaluate", {
      method: "POST",
      body: { key, variables, version },
      signal,
    });
    return { result: data.result, err: data.err };
  },
};
