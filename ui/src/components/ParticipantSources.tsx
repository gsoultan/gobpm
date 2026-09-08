import {
  ActionIcon,
  Alert,
  Badge,
  Button,
  Card,
  Group,
  Modal,
  PasswordInput,
  Select,
  Stack,
  Switch,
  Table,
  Text,
  Textarea,
  TextInput,
  ThemeIcon,
  Tooltip,
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { AlertTriangle, Database, Globe, Plus, RefreshCw, Repeat, Trash2 } from 'lucide-react';
import { useState } from 'react';

import {
  ON_MISSING,
  ON_MISSING_LABELS,
  SCHEDULE_OPTIONS,
  SOURCE_HEALTH_LABELS,
  scheduleLabel,
  sourceHealth,
  summarise,
  type ParticipantSource,
} from '../domain/participantImport';
import {
  useDeleteSource,
  useSaveSource,
  useSources,
  useSyncSource,
} from '../hooks/useParticipantSources';
import { errorMessage } from '../services/shared/errors';

dayjs.extend(relativeTime);

type Draft = {
  id?: string;
  name: string;
  kind: 'http' | 'postgres';
  url: string;
  dsn: string;
  query: string;
  schedule: string;
  on_missing: string;
  enabled: boolean;
};

const emptyDraft = (): Draft => ({
  name: '',
  kind: 'http',
  url: '',
  dsn: '',
  query: '',
  schedule: '',
  on_missing: 'leave',
  enabled: true,
});

/**
 * The directories this project keeps its people in step with.
 *
 * Separate from the one-off import above it, because they are different acts: an
 * upload is something somebody did once, a source is a standing statement that
 * another system is where the answer lives. Only the second is worth a schedule.
 */
export function ParticipantSources() {
  const { data, isLoading, error } = useSources();
  const save = useSaveSource();
  const remove = useDeleteSource();
  const sync = useSyncSource();
  const [draft, setDraft] = useState<Draft | null>(null);

  const sources = data?.sources ?? [];

  const edit = (source: ParticipantSource) => {
    setDraft({
      id: source.id,
      name: source.name,
      kind: source.kind,
      url: String(source.config?.url ?? ''),
      dsn: String(source.config?.dsn ?? ''),
      query: String(source.config?.query ?? ''),
      schedule: source.schedule ?? '',
      on_missing: source.on_missing,
      enabled: source.enabled,
    });
  };

  const submit = async () => {
    if (!draft) return;
    const config = draft.kind === 'http' ? { url: draft.url } : { dsn: draft.dsn, query: draft.query };
    try {
      await save.mutateAsync({
        id: draft.id,
        name: draft.name,
        kind: draft.kind,
        config,
        schedule: draft.schedule,
        on_missing: draft.on_missing,
        enabled: draft.enabled,
      });
      setDraft(null);
    } catch (err: unknown) {
      notifications.show({
        title: 'Could not save this directory',
        message: errorMessage(err, 'It was not saved.'),
        color: 'red',
      });
    }
  };

  const runNow = async (source: ParticipantSource) => {
    try {
      const summary = await sync.mutateAsync(source.id);
      notifications.show({
        title: `${source.name} synced`,
        message: summarise(summary) + (summary.deactivated ? `, ${summary.deactivated} stood down` : ''),
        color: (summary.problems?.length ?? 0) > 0 ? 'orange' : 'green',
      });
    } catch (err: unknown) {
      notifications.show({
        title: `${source.name} could not be synced`,
        message: errorMessage(err, 'The directory was not read.'),
        color: 'red',
      });
    }
  };

  return (
    <Card withBorder radius="lg" p="lg">
      <Stack gap="md">
        <Group justify="space-between" align="flex-start">
          <Group gap="sm">
            <ThemeIcon variant="light" color="indigo" size="lg" radius="md">
              <Repeat size={18} />
            </ThemeIcon>
            <Stack gap={0}>
              <Text fw={700}>Directories</Text>
              <Text size="xs" c="dimmed">
                Systems this project keeps its people in step with, on a schedule or on demand.
              </Text>
            </Stack>
          </Group>
          <Button size="compact-sm" variant="light" leftSection={<Plus size={14} />} onClick={() => setDraft(emptyDraft())}>
            Add directory
          </Button>
        </Group>

        {error && (
          <Alert color="red" icon={<AlertTriangle size={16} />} radius="md">
            <Text size="sm">{errorMessage(error, 'The directories could not be loaded.')}</Text>
          </Alert>
        )}

        {!isLoading && sources.length === 0 && !error && (
          <Alert color="gray" variant="light" radius="md">
            <Text size="sm">
              No directories yet. Add one to keep this project's people in step with an API or
              another database automatically.
            </Text>
          </Alert>
        )}

        {sources.length > 0 && (
          <Table verticalSpacing="sm">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Directory</Table.Th>
                <Table.Th>Runs</Table.Th>
                <Table.Th>Missing people</Table.Th>
                <Table.Th>Last run</Table.Th>
                <Table.Th ta="right">Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {sources.map((source) => {
                const health = sourceHealth(source);
                const meta = SOURCE_HEALTH_LABELS[health];
                return (
                  <Table.Tr key={source.id}>
                    <Table.Td>
                      <Group gap="xs">
                        {source.kind === 'http' ? <Globe size={14} /> : <Database size={14} />}
                        <Stack gap={0}>
                          <Text size="sm" fw={600}>{source.name}</Text>
                          <Text size="xs" c="dimmed">{source.kind === 'http' ? 'API endpoint' : 'Database query'}</Text>
                        </Stack>
                      </Group>
                    </Table.Td>
                    <Table.Td><Text size="sm">{scheduleLabel(source.schedule)}</Text></Table.Td>
                    <Table.Td>
                      <Tooltip label={ON_MISSING_LABELS[source.on_missing]?.hint ?? ''} multiline w={280}>
                        <Badge size="sm" variant="light" color={source.on_missing === 'deactivate' ? 'orange' : 'gray'}>
                          {ON_MISSING_LABELS[source.on_missing]?.label ?? source.on_missing}
                        </Badge>
                      </Tooltip>
                    </Table.Td>
                    <Table.Td>
                      <Group gap={6}>
                        <Badge size="sm" variant="light" color={meta.colour}>{meta.label}</Badge>
                        {source.last_run && (
                          <Tooltip label={source.last_run.detail || `${source.last_run.created} added, ${source.last_run.updated} updated`}>
                            <Text size="xs" c="dimmed">{dayjs(source.last_run.at).fromNow()}</Text>
                          </Tooltip>
                        )}
                      </Group>
                    </Table.Td>
                    <Table.Td>
                      <Group justify="flex-end" gap="xs">
                        <Tooltip label="Read this directory now">
                          <Button
                            size="compact-xs"
                            variant="light"
                            leftSection={<RefreshCw size={12} />}
                            loading={sync.isPending}
                            onClick={() => runNow(source)}
                          >
                            Sync now
                          </Button>
                        </Tooltip>
                        <Button size="compact-xs" variant="subtle" onClick={() => edit(source)}>Edit</Button>
                        <Tooltip label="Removes the directory. The people it brought in stay.">
                          <ActionIcon
                            aria-label={`Remove ${source.name}`}
                            variant="subtle"
                            color="red"
                            loading={remove.isPending}
                            onClick={() => remove.mutate(source.id)}
                          >
                            <Trash2 size={14} />
                          </ActionIcon>
                        </Tooltip>
                      </Group>
                    </Table.Td>
                  </Table.Tr>
                );
              })}
            </Table.Tbody>
          </Table>
        )}
      </Stack>

      <Modal
        opened={draft !== null}
        onClose={() => setDraft(null)}
        title={<Text fw={800}>{draft?.id ? `Edit ${draft.name}` : 'New directory'}</Text>}
        size="lg"
        radius="lg"
      >
        {draft && (
          <Stack gap="sm">
            <TextInput
              label="Name"
              placeholder="HR system"
              value={draft.name}
              onChange={(e) => setDraft({ ...draft, name: e.currentTarget.value })}
            />
            <Select
              label="Read from"
              data={[
                { value: 'http', label: 'API endpoint' },
                { value: 'postgres', label: 'Database query' },
              ]}
              value={draft.kind}
              allowDeselect={false}
              onChange={(value) => setDraft({ ...draft, kind: (value ?? 'http') as Draft['kind'] })}
            />

            {draft.kind === 'http' ? (
              <TextInput
                label="Endpoint"
                description="Returns a JSON array, or an object with the list under data, items, results, users or participants"
                placeholder="https://hr.example.com/api/staff"
                value={draft.url}
                onChange={(e) => setDraft({ ...draft, url: e.currentTarget.value })}
              />
            ) : (
              <>
                <PasswordInput
                  label="Database"
                  description="Connection string. Stored encrypted; leave as it is to keep the stored one."
                  placeholder="postgres://user:password@host:5432/hr"
                  value={draft.dsn}
                  onChange={(e) => setDraft({ ...draft, dsn: e.currentTarget.value })}
                />
                <Textarea
                  label="Query"
                  description="Alias your columns into username, display_name, email, groups, active"
                  autosize
                  minRows={3}
                  value={draft.query}
                  onChange={(e) => setDraft({ ...draft, query: e.currentTarget.value })}
                  styles={{ input: { fontFamily: 'monospace' } }}
                />
              </>
            )}

            <Select
              label="Run"
              data={SCHEDULE_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
              value={draft.schedule}
              allowDeselect={false}
              onChange={(value) => setDraft({ ...draft, schedule: value ?? '' })}
            />
            <Select
              label="When somebody is no longer in the directory"
              description={ON_MISSING_LABELS[draft.on_missing]?.hint}
              data={ON_MISSING.map((value) => ({ value, label: ON_MISSING_LABELS[value].label }))}
              value={draft.on_missing}
              allowDeselect={false}
              onChange={(value) => setDraft({ ...draft, on_missing: value ?? 'leave' })}
            />
            <Switch
              label="Run this directory"
              checked={draft.enabled}
              onChange={(e) => setDraft({ ...draft, enabled: e.currentTarget.checked })}
            />

            <Group justify="flex-end">
              <Button variant="default" onClick={() => setDraft(null)} disabled={save.isPending}>Cancel</Button>
              <Button color="indigo" loading={save.isPending} disabled={!draft.name.trim()} onClick={submit}>
                {draft.id ? 'Save' : 'Add directory'}
              </Button>
            </Group>
          </Stack>
        )}
      </Modal>
    </Card>
  );
}
