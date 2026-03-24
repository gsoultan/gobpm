import { API_BASE_URL } from "./config";
import { getAuthHeaders } from "./auth";

type RequestOptions = {
  method?: string;
  body?: unknown;
  signal?: AbortSignal;
  headers?: Record<string, string>;
  auth?: boolean;
};

const toErrorMessage = (statusText: string, fallback: string) => {
  if (statusText) {
    return statusText;
  }

  return fallback;
};

export const requestJSON = async <T>(path: string, options: RequestOptions = {}): Promise<T> => {
  const { method = "GET", body, signal, headers = {}, auth = true } = options;

  const finalHeaders: Record<string, string> = {
    ...headers,
    ...(auth ? getAuthHeaders() : {}),
  };

  if (body !== undefined && !finalHeaders["Content-Type"]) {
    finalHeaders["Content-Type"] = "application/json";
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers: finalHeaders,
    body: body === undefined ? undefined : JSON.stringify(body),
    signal,
  });

  if (response.status === 204) {
    return {} as T;
  }

  const data = (await response.json().catch(() => ({}))) as Record<string, unknown>;
  if (!response.ok) {
    const errorMessage = typeof data.error === "string"
      ? data.error
      : toErrorMessage(response.statusText, "Request failed");
    throw new Error(errorMessage);
  }

  return data as T;
};
