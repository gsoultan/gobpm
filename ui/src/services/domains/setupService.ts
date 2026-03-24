import { requestJSON } from "../shared/rest";
import type { SetupRequest } from "../types";

export type SetupStatus = 'not_configured' | 'configured' | 'ready';

export const setupService = {
  async getSetupStatus(signal?: AbortSignal) {
    const data = await requestJSON<{ status?: SetupStatus; err?: string }>("/setup/status", {
      signal,
      auth: false,
    });

    return { status: data.status, err: data.err };
  },

  async setup(req: SetupRequest, signal?: AbortSignal) {
    const data = await requestJSON<{ err?: string }>("/setup", {
      method: "POST",
      body: req,
      signal,
      auth: false,
    });

    return { err: data.err };
  },

  async testConnection(req: Pick<SetupRequest, 'database_driver' | 'db_host' | 'db_port' | 'db_username' | 'db_password' | 'db_name' | 'db_ssl_enabled'>, signal?: AbortSignal) {
    const data = await requestJSON<{ success: boolean; message: string }>("/setup/test-connection", {
      method: "POST",
      body: req,
      signal,
      auth: false,
    });

    return { success: data.success, message: data.message };
  },
};
