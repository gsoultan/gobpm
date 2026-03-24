import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { DefinitionService } from "../../gen/services/definition_connect";
import { OrganizationService } from "../../gen/services/organization_connect";
import { ProcessService } from "../../gen/services/process_connect";
import { ProjectService } from "../../gen/services/project_connect";
import { SignalService } from "../../gen/services/signal_connect";
import { StatsService } from "../../gen/services/stats_connect";
import { TaskService } from "../../gen/services/task_connect";
import { UserService } from "../../gen/services/user_connect";
import { API_BASE_URL } from "./config";
import { getAuthToken } from "./auth";

export const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  interceptors: [
    (next) => async (req) => {
      const token = getAuthToken();
      if (token) {
        req.header.set("Authorization", `Bearer ${token}`);
      }
      return next(req);
    },
  ],
});

export const organizationClient = createPromiseClient(OrganizationService, transport);
export const projectClient = createPromiseClient(ProjectService, transport);
export const processClient = createPromiseClient(ProcessService, transport);
export const taskClient = createPromiseClient(TaskService, transport);
export const definitionClient = createPromiseClient(DefinitionService, transport);
export const signalClient = createPromiseClient(SignalService, transport);
export const statsClient = createPromiseClient(StatsService, transport);
export const userClient = createPromiseClient(UserService, transport);
