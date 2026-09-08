import {
  Badge,
  Button,
  Card,
  Group,
  Stack,
  Table,
  Text,
  TextInput,
  ThemeIcon,
  Tooltip,
} from '@mantine/core';
import { KeyRound, Search, Upload, UserCircle, Users } from 'lucide-react';
import { useState, useTransition } from 'react';

import { PageHeader } from '../components/PageHeader';
import { ParticipantImportModal } from '../components/ParticipantImportModal';
import { ParticipantSources } from '../components/ParticipantSources';
import { EmptyState, ErrorState, TableLoadingState } from '../components/state';
import {
  STANDING_LABELS,
  standingOf,
  type ImportSummary,
  type Participant,
} from '../domain/participantImport';
import { matchesQuery } from '../domain/textSearch';
import { useAppStore } from '../store/useAppStore';
import {
  useImportParticipants,
  useParticipants,
} from '../hooks/useParticipants';

const COLUMNS = 4;

const STANDING_COLOURS: Record<string, string> = {
  ready: 'green',
  'no-credentials': 'yellow',
  inactive: 'gray',
};

/**
 * The people this project's processes can assign work to.
 *
 * Deliberately not the same page as platform accounts, and deliberately not
 * next to them in the navigation. They were one list, which meant one role
 * list answered two unrelated questions — who may approve an invoice, and who
 * may deploy a process model — so giving somebody an inbox gave them the
 * platform.
 *
 * These are per project: two projects naming an approver "ada" mean two
 * different people.
 */
export function ParticipantList() {
  const isAdmin = useAppStore((state) => state.user?.role === 'ADMIN');
  const { data, isLoading, error, refetch } = useParticipants();
  const importParticipants = useImportParticipants();
  const [query, setQuery] = useState('');
  const [importOpen, setImportOpen] = useState(false);
  const [result, setResult] = useState<ImportSummary | null>(null);
  const [, startTransition] = useTransition();

  const everyone: Participant[] = data?.participants ?? [];
  const people = query
    ? everyone.filter((p) => matchesQuery(query, p.username, p.display_name ?? '', p.email ?? ''))
    : everyone;

  const runImport = async (body: Record<string, unknown>) => {
    try {
      setResult(await importParticipants.mutateAsync(body));
    } catch {
      // The hook reports it; the dialog stays open so the source can be
      // corrected rather than retyped.
    }
  };

  if (error) {
    return <ErrorState error={error} action="load the people in this project" onRetry={() => refetch()} />;
  }

  return (
    <Stack gap="xl">
      <PageHeader
        title="People"
        description="Who this project's processes can assign work to. Separate from the accounts that administer Metis."
        actions={
          <Button
            variant="filled"
            color="indigo"
            leftSection={<Upload size={16} />}
            onClick={() => { setResult(null); setImportOpen(true); }}
          >
            Import
          </Button>
        }
      />

      <Card withBorder radius="lg" p={0}>
        <Group p="md" justify="space-between">
          <TextInput
            placeholder="Search by name or email"
            leftSection={<Search size={16} />}
            value={query}
            onChange={(e) => {
              const value = e.currentTarget.value;
              startTransition(() => setQuery(value));
            }}
            w={320}
          />
          {everyone.length > 0 && (
            <Text size="xs" c="dimmed">
              {query ? `${people.length} of ${everyone.length} match` : `${everyone.length} in this project`}
            </Text>
          )}
        </Group>

        {isLoading && <TableLoadingState columns={COLUMNS} />}

        {!isLoading && everyone.length === 0 && (
          <EmptyState
            icon={Users}
            title="Nobody here yet"
            description="Import a directory from a file, an API or a database query, so processes have somebody to assign work to."
          />
        )}

        {!isLoading && everyone.length > 0 && (
          <Table.ScrollContainer minWidth={720}>
            <Table verticalSpacing="sm" highlightOnHover>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>Person</Table.Th>
                  <Table.Th>Email</Table.Th>
                  <Table.Th>Teams</Table.Th>
                  <Table.Th>Standing</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {people.map((person) => {
                  const standing = standingOf(person);
                  const meta = STANDING_LABELS[standing];
                  return (
                    <Table.Tr key={person.id}>
                      <Table.Td>
                        <Group gap="sm">
                          <ThemeIcon variant="light" color="gray" radius="xl" size="md">
                            <UserCircle size={16} />
                          </ThemeIcon>
                          <Stack gap={0}>
                            <Text size="sm" fw={600}>{person.display_name || person.username}</Text>
                            <Text size="xs" c="dimmed" ff="monospace">{person.username}</Text>
                          </Stack>
                        </Group>
                      </Table.Td>
                      <Table.Td><Text size="sm" c="dimmed">{person.email || '—'}</Text></Table.Td>
                      <Table.Td>
                        <Group gap={4}>
                          {(person.groups ?? []).length === 0
                            ? <Text size="xs" c="dimmed">—</Text>
                            : (person.groups ?? []).map((team) => (
                                <Badge key={team} size="sm" variant="light" color="blue">{team}</Badge>
                              ))}
                        </Group>
                      </Table.Td>
                      <Table.Td>
                        <Tooltip label={meta.hint} multiline w={260}>
                          <Badge
                            size="sm"
                            variant="light"
                            color={STANDING_COLOURS[standing]}
                            leftSection={standing === 'no-credentials' ? <KeyRound size={10} /> : undefined}
                          >
                            {meta.label}
                          </Badge>
                        </Tooltip>
                      </Table.Td>
                    </Table.Tr>
                  );
                })}
              </Table.Tbody>
            </Table>
          </Table.ScrollContainer>
        )}
      </Card>

      {/*
        Directories are administrative: each holds a connection string or an
        endpoint token, so the list alone says where they live. The server
        refuses a non-admin either way; hiding it is so nobody is shown a
        control that will only ever say no.
      */}
      {isAdmin && <ParticipantSources />}

      <ParticipantImportModal
        opened={importOpen}
        onClose={() => setImportOpen(false)}
        running={importParticipants.isPending}
        result={result}
        onImportFile={(file) => runImport({ kind: 'csv', file })}
        onImportHTTP={(url, method) => runImport({ kind: 'http', url, method })}
        onImportPostgres={(dsn, query2) => runImport({ kind: 'postgres', dsn, query: query2 })}
      />
    </Stack>
  );
}
