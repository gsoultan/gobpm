import {
  authService,
  collaborationService,
  connectorService,
  decisionService,
  definitionService,
  identityService,
  notificationService,
  organizationService,
  processRuntimeService,
  projectService,
  setupService,
  signalService,
  taskService,
} from "./domains";

export type { Task, Project } from "./types";

export const processService: any = {
  ...authService,
  ...organizationService,
  ...projectService,
  ...processRuntimeService,
  ...taskService,
  ...definitionService,
  ...decisionService,
  ...signalService,
  ...connectorService,
  ...collaborationService,
  ...identityService,
  ...notificationService,
  ...setupService,
};
