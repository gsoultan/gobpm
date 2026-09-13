import type { ApiConnector } from '../../services/types';

/** What the setup wizard is editing: a name and the connector's config values. */
export interface InstanceFormState {
  name: string;
  config: Record<string, unknown>;
}

/** The form for a connection that does not exist yet: every field at its default. */
export function emptyFormFor(connector: ApiConnector): InstanceFormState {
  const config: Record<string, unknown> = {};
  for (const prop of connector.schema ?? []) {
    config[prop.key] = prop.default_value || '';
  }
  return { name: `${connector.name} Instance`, config };
}
