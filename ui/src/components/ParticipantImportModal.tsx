import {
  Alert,
  Badge,
  Button,
  Code,
  FileInput,
  Group,
  List,
  Modal,
  PasswordInput,
  Stack,
  Tabs,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { AlertTriangle, CheckCircle2, Database, FileUp, Globe, Upload } from 'lucide-react';
import { useState } from 'react';

import {
  canRunHTTP,
  canRunPostgres,
  outcomeTone,
  problemsInFileOrder,
  summarise,
  type ImportSummary,
  type SourceKind,
} from '../domain/participantImport';

interface ParticipantImportModalProps {
  opened: boolean;
  onClose: () => void;
  running: boolean;
  result: ImportSummary | null;
  onImportFile: (file: File) => void;
  onImportHTTP: (url: string, method: string) => void;
  onImportPostgres: (dsn: string, query: string) => void;
}

/**
 * Bringing a directory in, from wherever it lives.
 *
 * The three sources are tabs rather than three buttons because they are one
 * decision — where does this list come from — made once. They produce the same
 * outcome, and the result panel below them is shared, which is the honest
 * rendering of a backend where only the fetching differs.
 */
export function ParticipantImportModal({
  opened,
  onClose,
  running,
  result,
  onImportFile,
  onImportHTTP,
  onImportPostgres,
}: ParticipantImportModalProps) {
  const [kind, setKind] = useState<SourceKind>('csv');
  const [file, setFile] = useState<File | null>(null);
  const [url, setUrl] = useState('');
  const [method, setMethod] = useState('GET');
  const [dsn, setDsn] = useState('');
  const [query, setQuery] = useState('');

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={<Group gap="xs"><Upload size={20} /><Text fw={800}>Import people</Text></Group>}
      size="xl"
      radius="lg"
    >
      <Stack gap="md">
        <Text size="sm" c="dimmed">
          Every source uses the same column names: <Code>username</Code>, and optionally{' '}
          <Code>display_name</Code>, <Code>email</Code>, <Code>groups</Code> and <Code>active</Code>.
          Only <Code>username</Code> is required.
        </Text>

        <Tabs value={kind} onChange={(value) => setKind((value ?? 'csv') as SourceKind)}>
          <Tabs.List>
            <Tabs.Tab value="csv" leftSection={<FileUp size={14} />}>Upload a file</Tabs.Tab>
            <Tabs.Tab value="http" leftSection={<Globe size={14} />}>API endpoint</Tabs.Tab>
            <Tabs.Tab value="postgres" leftSection={<Database size={14} />}>Database query</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="csv" pt="md">
            <Stack gap="sm">
              <FileInput
                label="CSV file"
                description="The first row names the columns, in any order"
                placeholder="Choose a file"
                accept=".csv,text/csv"
                value={file}
                onChange={setFile}
                leftSection={<FileUp size={16} />}
              />
              <Group justify="flex-end">
                <Button loading={running} disabled={!file} onClick={() => file && onImportFile(file)}>
                  Import
                </Button>
              </Group>
            </Stack>
          </Tabs.Panel>

          <Tabs.Panel value="http" pt="md">
            <Stack gap="sm">
              <TextInput
                label="Endpoint"
                description="Returns a JSON array, or an object with the list under data, items, results, users or participants"
                placeholder="https://hr.example.com/api/staff"
                value={url}
                onChange={(e) => setUrl(e.currentTarget.value)}
              />
              <TextInput
                label="Method"
                placeholder="GET"
                value={method}
                onChange={(e) => setMethod(e.currentTarget.value)}
              />
              <Text size="xs" c="dimmed">
                Addresses on private networks are refused unless this installation allows them.
              </Text>
              <Group justify="flex-end">
                <Button
                  loading={running}
                  disabled={!canRunHTTP({ url, method, headers: [] })}
                  onClick={() => onImportHTTP(url, method)}
                >
                  Import
                </Button>
              </Group>
            </Stack>
          </Tabs.Panel>

          <Tabs.Panel value="postgres" pt="md">
            <Stack gap="sm">
              <PasswordInput
                label="Database"
                description="Connection string. Stored encrypted and never shown again."
                placeholder="postgres://user:password@host:5432/hr"
                value={dsn}
                onChange={(e) => setDsn(e.currentTarget.value)}
              />
              <Textarea
                label="Query"
                description="Alias your columns into the names above"
                placeholder={'SELECT login AS username,\n       full_name AS display_name,\n       work_email AS email\nFROM staff\nWHERE is_current'}
                autosize
                minRows={4}
                value={query}
                onChange={(e) => setQuery(e.currentTarget.value)}
                styles={{ input: { fontFamily: 'monospace' } }}
              />
              <Group justify="flex-end">
                <Button
                  loading={running}
                  disabled={!canRunPostgres({ dsn, query })}
                  onClick={() => onImportPostgres(dsn, query)}
                >
                  Import
                </Button>
              </Group>
            </Stack>
          </Tabs.Panel>
        </Tabs>

        {result && <ImportOutcome result={result} />}
      </Stack>
    </Modal>
  );
}

/**
 * What the import did, including what it refused.
 *
 * A partial import is shown as a warning rather than a success even though the
 * good rows landed. Reporting only "imported" is what leaves somebody to
 * discover months later that forty people were never in the system.
 */
function ImportOutcome({ result }: { result: ImportSummary }) {
  const tone = outcomeTone(result);
  const problems = problemsInFileOrder(result);

  return (
    <Alert
      color={tone === 'partial' ? 'orange' : tone === 'success' ? 'green' : 'gray'}
      variant="light"
      radius="md"
      icon={tone === 'partial' ? <AlertTriangle size={16} /> : <CheckCircle2 size={16} />}
    >
      <Stack gap="xs">
        <Group gap="xs">
          <Text size="sm" fw={600}>{summarise(result)}</Text>
          {problems.length > 0 && (
            <Badge color="orange" variant="filled" size="sm">
              {problems.length} not imported
            </Badge>
          )}
        </Group>

        {problems.length > 0 && (
          <>
            <Text size="xs">
              The rest imported. These rows did not, and are listed where you can find them:
            </Text>
            <List size="xs" spacing={2} withPadding>
              {problems.slice(0, 20).map((problem) => (
                <List.Item key={`${problem.line}-${problem.reason}`}>
                  <Text size="xs" span ff="monospace">line {problem.line}</Text>
                  {problem.username ? <Text size="xs" span fw={600}> ({problem.username})</Text> : null}
                  <Text size="xs" span c="dimmed"> — {problem.reason}</Text>
                </List.Item>
              ))}
            </List>
            {problems.length > 20 && (
              <Text size="xs" c="dimmed">…and {problems.length - 20} more.</Text>
            )}
          </>
        )}
      </Stack>
    </Alert>
  );
}
