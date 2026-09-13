/**
 * The connections configured in this project.
 *
 * It says what is configured and when, and no more. It used to show a
 * hard-coded "Online" badge and a "99.8% (1.2k calls)" success rate on every
 * row; nothing measures either, and a health figure nobody measured is worse
 * than none.
 */
import {
  ActionIcon,
  Alert,
  Badge,
  Box,
  Button,
  Divider,
  Group,
  Paper,
  ScrollArea,
  Stack,
  Table,
  Text,
  ThemeIcon,
  Title,
  Tooltip,
} from '@mantine/core';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { AlertCircle, Play, Settings, Trash2 } from 'lucide-react';

import type { ApiConnector, ApiConnectorInstance } from '../../services/types';
import { ConnectorGlyph } from './ConnectorGlyph';

dayjs.extend(relativeTime);

interface ConnectorInstanceTableProps {
  instances: ApiConnectorInstance[];
  connectors: ApiConnector[];
  expertMode: boolean;
  onRefresh: () => void;
  refreshing: boolean;
  onTest: (instance: ApiConnectorInstance, connector: ApiConnector | undefined) => void;
  onEdit: (instance: ApiConnectorInstance) => void;
  onDelete: (instance: ApiConnectorInstance) => void;
}

export function ConnectorInstanceTable({
  instances,
  connectors,
  expertMode,
  onRefresh,
  refreshing,
  onTest,
  onEdit,
  onDelete,
}: ConnectorInstanceTableProps) {
  return (
    <Paper p="xl" radius="lg" withBorder shadow="sm">
      <Stack gap="md">
        <Group justify="space-between">
          <Box>
            <Title order={4}>Configured connections</Title>
            <Text size="xs" c="dimmed">The services this project's steps can call.</Text>
          </Box>
          <Button variant="subtle" size="xs" onClick={onRefresh} loading={refreshing}>
            Refresh
          </Button>
        </Group>
        <Divider />
        {instances.length === 0 ? (
          <Alert icon={<AlertCircle size={16} />} color="gray" variant="light">
            No connections yet. Pick a connector from the catalogue above to add one.
          </Alert>
        ) : (
          <ScrollArea>
            <Table verticalSpacing="md" horizontalSpacing="lg">
              <Table.Thead bg="gray.0">
                <Table.Tr>
                  <Table.Th>Connection</Table.Th>
                  <Table.Th>Service</Table.Th>
                  <Table.Th>Added</Table.Th>
                  <Table.Th ta="right">Actions</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {instances.map((instance) => {
                  const connector = connectors.find((c) => c.id === instance.connector?.id);
                  return (
                    <Table.Tr key={instance.id}>
                      <Table.Td>
                        <Group gap="sm">
                          <ThemeIcon size="md" variant="light" color="gray">
                            <ConnectorGlyph name={connector?.icon} size={18} />
                          </ThemeIcon>
                          <Stack gap={0}>
                            <Text fw={700} size="sm">{instance.name}</Text>
                            {expertMode && <Text size="xs" c="dimmed" ff="monospace">{instance.id}</Text>}
                          </Stack>
                        </Group>
                      </Table.Td>
                      <Table.Td>
                        <Badge variant="outline" color="gray" size="sm">{connector?.name || 'Unknown'}</Badge>
                      </Table.Td>
                      <Table.Td>
                        <Text size="sm" c="dimmed">
                          {instance.created_at ? dayjs(instance.created_at).fromNow() : '—'}
                        </Text>
                      </Table.Td>
                      <Table.Td>
                        <Group gap="xs" justify="flex-end">
                          <Tooltip label="Test connection">
                            <ActionIcon
                              aria-label={`Test the ${instance.name} connection`}
                              variant="light"
                              color="orange"
                              onClick={() => onTest(instance, connector)}
                            >
                              <Play size={16} />
                            </ActionIcon>
                          </Tooltip>
                          <Tooltip label="Configure">
                            <ActionIcon
                              aria-label={`Configure ${instance.name}`}
                              variant="light"
                              color="blue"
                              onClick={() => onEdit(instance)}
                            >
                              <Settings size={16} />
                            </ActionIcon>
                          </Tooltip>
                          <Tooltip label="Remove">
                            <ActionIcon
                              aria-label={`Remove the ${instance.name} connection`}
                              variant="light"
                              color="red"
                              onClick={() => onDelete(instance)}
                            >
                              <Trash2 size={16} />
                            </ActionIcon>
                          </Tooltip>
                        </Group>
                      </Table.Td>
                    </Table.Tr>
                  );
                })}
              </Table.Tbody>
            </Table>
          </ScrollArea>
        )}
      </Stack>
    </Paper>
  );
}
