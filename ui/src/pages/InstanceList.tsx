import {
  ActionIcon,
  Badge,
  Button,
  Card,
  Center,
  Drawer,
  Group,
  Pagination,
  Select,
  Skeleton,
  Stack,
  Table,
  Text,
  Tooltip,
} from '@mantine/core';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { AlertTriangle, Eye, RefreshCw } from 'lucide-react';
import { useState } from 'react';

import { IncidentInbox } from '../components/IncidentInbox';
import { PageHeader } from '../components/PageHeader';
import { StatusBadge } from '../components/StatusBadge';
import { ErrorState } from '../components/state';
import { STATUS } from '../components/statusVocabulary';
import {
  definitionName,
  humanizeNodeId,
  instanceReference,
  startedAtFromId,
  statusesOnPage,
  withStatus,
} from '../domain/instanceList';
import { useDefinitions } from '../hooks/useDefinitions';
import { useInstances } from '../hooks/useProcess';

dayjs.extend(relativeTime);

const PAGE_SIZES = ['25', '50', '100'];
const COLUMNS = 4;

export function InstanceList({ onViewInstance }: { onViewInstance: (instanceId: string, definitionId: string) => void }) {
  // Which instance's failures are on screen, if any.
  const [inspecting, setInspecting] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [status, setStatus] = useState<string | null>(null);
  const { data, isLoading, error, refetch } = useInstances(page, pageSize);
  // A listed instance carries only its definition's id, so every row read
  // "Process". The definitions are already a cached query; joining them here
  // costs one more request and gives each row the name of the process it is.
  const { data: definitionsData } = useDefinitions();
  const definitions = definitionsData?.definitions ?? [];
  const pageInfo = data?.pageInfo;

  // Changing the window size invalidates the current offset. Adjusted during
  // render rather than in an effect, which would render once with the old page
  // against the new size and then again to correct it.
  const [appliedPageSize, setAppliedPageSize] = useState(pageSize);
  if (pageSize !== appliedPageSize) {
    setAppliedPageSize(pageSize);
    setPage(1);
  }

  if (isLoading) {
    return (
      <Stack gap="xl">
        <Skeleton height={40} radius="md" />
        <Card withBorder radius="lg" p={0}>
          <Table verticalSpacing="md">
            <thead>
              <tr>
                <th>Process</th>
                <th>Status</th>
                <th>Where it is</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: 4 }).map((_, i) => (
                <tr key={i}>
                  <td><Skeleton height={16} width="50%" /></td>
                  <td><Skeleton height={16} width={80} /></td>
                  <td><Skeleton height={16} width="40%" /></td>
                  <td><Skeleton height={16} width={60} /></td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Card>
      </Stack>
    );
  }

  // A rejected request previously fell through to the empty state, so an
  // outage was reported to the user as "you have nothing".
  if (error) {
    return <ErrorState error={error} action="load your process instances" onRetry={() => refetch()} />;
  }

  const onPage = data?.instances ?? [];
  const instances = withStatus(onPage, status);
  // The server has no status filter, so this narrows the page on screen and
  // says so: a filter that looks server-wide and is not would hide the rest.
  const statusOptions = statusesOnPage(onPage).map((value) => ({ value, label: STATUS[value]?.label ?? value }));

  return (
    <Stack gap="xl">
      <PageHeader
        title="Process Instances"
        description="Every run of a process in this project, and where each one is."
        actions={
          <Button variant="light" leftSection={<RefreshCw size={16} />} onClick={() => refetch()}>Refresh</Button>
        }
      />

      {/* The failures drawer was mounted only while the page was loading, so
          the button that opens it did nothing once there were rows to click. */}
      <Drawer
        opened={inspecting !== null}
        onClose={() => setInspecting(null)}
        position="right"
        size="lg"
        title="What went wrong"
      >
        {inspecting && <IncidentInbox instanceId={inspecting} />}
      </Drawer>

      <Card withBorder radius="lg" p={0}>
        <Group px="md" py="sm" justify="flex-end">
          <Select
            aria-label="Show only one status, on this page"
            placeholder="Any status on this page"
            data={statusOptions}
            value={status}
            onChange={setStatus}
            clearable
            size="xs"
            w={220}
            comboboxProps={{ withinPortal: true }}
          />
        </Group>
        <Table verticalSpacing="md">
          <thead>
            <tr>
              <th>Process</th>
              <th>Status</th>
              <th>Where it is</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {instances.length === 0 ? (
              <tr>
                <td colSpan={COLUMNS}>
                  <Center py="xl">
                    <Text size="sm" c="dimmed">
                      {status ? 'Nothing on this page has that status.' : 'No instances found for this project'}
                    </Text>
                  </Center>
                </td>
              </tr>
            ) : (
              instances.map((inst) => {
                const startedAt = startedAtFromId(inst.id);
                return (
                  <tr key={inst.id}>
                    <td>
                      {/* The process is what a person identifies an instance
                          by; the started time and a short reference are what
                          tell two runs of the same process apart. */}
                      <Text size="sm" fw={500}>{definitionName(inst, definitions)}</Text>
                      <Text size="xs" c="dimmed">
                        {startedAt ? `Started ${dayjs(startedAt).fromNow()} · ` : ''}
                        {instanceReference(inst.id)}
                      </Text>
                    </td>
                    <td><StatusBadge status={inst.status} withIcon /></td>
                    <td>
                      <Group gap={4}>
                        {(inst.activeNodes ?? []).map((node) => (
                          <Badge key={node.id} size="sm" variant="light" color="blue">
                            {humanizeNodeId(node.id)}
                          </Badge>
                        ))}
                        {(inst.activeNodes ?? []).length === 0 && (
                          <Text size="xs" c="dimmed">
                            {inst.status === 'active' ? 'Starting…' : 'Nothing in progress'}
                          </Text>
                        )}
                      </Group>
                    </td>
                    <td>
                      <Group gap={6} wrap="nowrap">
                        <Tooltip label="Follow its path">
                          <ActionIcon
                            aria-label="View instance"
                            variant="light"
                            color="blue"
                            onClick={() => onViewInstance(inst.id, inst.definition?.id ?? '')}
                          >
                            <Eye size={16} />
                          </ActionIcon>
                        </Tooltip>
                        <Tooltip label="What went wrong, and try again">
                          <ActionIcon
                            aria-label="Show what failed"
                            variant="light"
                            color={inst.status === 'failed' ? 'red' : 'gray'}
                            onClick={() => setInspecting(inst.id)}
                          >
                            <AlertTriangle size={16} />
                          </ActionIcon>
                        </Tooltip>
                      </Group>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </Table>

        {/*
          Shown only when there is more than one page: controls that can never
          do anything are noise. The range is stated in words because "51–75 of
          1,240" answers both questions someone has about a long list, where a
          bare page number answers neither.
        */}
        {pageInfo && pageInfo.total > pageInfo.pageSize && (
          <Group justify="space-between" px="md" py="sm" wrap="wrap" gap="sm">
            <Text size="sm" c="dimmed">
              {`${(pageInfo.page - 1) * pageInfo.pageSize + 1}–` +
                `${Math.min(pageInfo.page * pageInfo.pageSize, pageInfo.total)}` +
                ` of ${pageInfo.total.toLocaleString()}`}
            </Text>
            <Group gap="sm" wrap="nowrap">
              <Select
                aria-label="Instances per page"
                data={PAGE_SIZES}
                value={String(pageSize)}
                onChange={(value) => value && setPageSize(Number(value))}
                size="xs"
                w={92}
                allowDeselect={false}
                comboboxProps={{ withinPortal: true }}
              />
              <Pagination
                // The server reports the page it served, after clamping; using
                // the requested value would let the highlight disagree with
                // what is on screen.
                value={pageInfo.page}
                onChange={setPage}
                total={Math.max(1, Math.ceil(pageInfo.total / pageInfo.pageSize))}
                size="sm"
                withEdges
                getControlProps={(control) => ({
                  'aria-label': {
                    first: 'First page',
                    last: 'Last page',
                    next: 'Next page',
                    previous: 'Previous page',
                  }[control] ?? undefined,
                })}
              />
            </Group>
          </Group>
        )}
      </Card>
    </Stack>
  );
}
