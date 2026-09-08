import {
  ActionIcon,
  Alert,
  Badge,
  Button,
  Group,
  Loader,
  Modal,
  Stack,
  Table,
  Text,
  TextInput,
  Tooltip,
} from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { AlertTriangle, ArrowRight, Plus, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';

import {
  carriedNodes,
  isApplicable,
  movedNodes,
  planSummary,
  toNodeMapping,
} from '../domain/instanceMigration';
import { useMigrateInstances, usePlanInstanceMigration } from '../hooks/useDefinitions';
import { errorMessage } from '../services/shared/errors';
import type { ApiMigrationPlan } from '../services/types';

/** A version, as this dialog needs to name it. */
export interface MigrationVersionRef {
  id: string;
  version: number;
}

interface MigrateInstancesModalProps {
  /** The version the work is on now, or null when the dialog is closed. */
  source: MigrationVersionRef | null;
  /** The version to move it to. */
  target: MigrationVersionRef | null;
  processKey: string;
  onClose: () => void;
}

interface MappingRow {
  /** Stable across edits, so a row keeps its focus when another is removed. */
  id: number;
  from: string;
  to: string;
}

/** Row ids are only unique within one open dialog, which is all they are for. */
let nextRowId = 0;

/**
 * Moves running instances from one version onto another.
 *
 * This is the escape hatch, not the normal path. Promoting a version and
 * letting the old one drain is what you do; this is for a version that must not
 * continue — a security fix, a calculation that was wrong. It rewrites work that
 * is already somebody's, so it previews by default: the plan is what you get
 * until you press the button that says apply.
 */
export function MigrateInstancesModal({ source, target, processKey, onClose }: MigrateInstancesModalProps) {
  const [rows, setRows] = useState<MappingRow[]>([]);
  const [plan, setPlan] = useState<ApiMigrationPlan | null>(null);
  const [refused, setRefused] = useState<string | null>(null);

  const preview = usePlanInstanceMigration();
  const apply = useMigrateInstances();

  const open = source !== null && target !== null;

  // The plan is fetched on open and re-fetched whenever the mapping changes,
  // because a mapping that strands a task must stop saying "ready to apply" the
  // moment it does.
  useEffect(() => {
    if (!source || !target) {
      setPlan(null);
      setRefused(null);
      return;
    }
    let cancelled = false;
    preview
      .mutateAsync({ source: source.id, target: target.id, mapping: toNodeMapping(rows) })
      .then((result) => {
        if (cancelled) return;
        setPlan(result.plan ?? null);
        setRefused(result.err ?? null);
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setPlan(null);
        setRefused(errorMessage(error, 'The plan could not be worked out.'));
      });
    return () => {
      cancelled = true;
    };
    // preview is a stable mutation object; including it would refetch on every
    // render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [source?.id, target?.id, JSON.stringify(rows)]);

  const handleApply = async () => {
    if (!source || !target) return;
    try {
      const result = await apply.mutateAsync({
        source: source.id,
        target: target.id,
        mapping: toNodeMapping(rows),
      });
      if (result.err) {
        setRefused(result.err);
        return;
      }
      notifications.show({
        title: 'Moved',
        message: `${plan?.instances ?? 0} instances now run on v${target.version}.`,
        color: 'green',
      });
      onClose();
    } catch (error: unknown) {
      setRefused(errorMessage(error, 'The instances could not be moved.'));
    }
  };

  const moved = plan ? movedNodes(plan) : [];
  const carried = plan ? carriedNodes(plan) : [];
  const ready = isApplicable(plan) && (plan?.instances ?? 0) > 0 && refused === null;

  return (
    <Modal
      opened={open}
      onClose={onClose}
      size="lg"
      radius="md"
      title={
        <Text fw={700}>
          Move running work to v{target?.version} — {processKey}
        </Text>
      }
    >
      <Stack gap="md">
        <Alert color="orange" icon={<AlertTriangle size={16} />} radius="md">
          <Text size="sm">
            This rewrites instances that have already started. The usual way to change version is to
            promote the new one and let v{source?.version} finish what it has — use this only when
            v{source?.version} must not continue.
          </Text>
        </Alert>

        {preview.isPending && !plan && (
          <Group gap="xs">
            <Loader size="xs" />
            <Text size="sm" c="dimmed">Working out what would move…</Text>
          </Group>
        )}

        {plan && <Text size="sm">{planSummary(plan)}</Text>}

        {refused && (
          <Alert color="red" icon={<AlertTriangle size={16} />} radius="md">
            <Text size="sm">{refused}</Text>
          </Alert>
        )}

        {plan?.refusals?.map((refusal) => (
          <Alert key={refusal} color="red" icon={<AlertTriangle size={16} />} radius="md">
            <Text size="sm">{refusal}</Text>
          </Alert>
        ))}

        {moved.length > 0 && (
          <Stack gap={4}>
            <Text size="sm" fw={600}>Work that moves</Text>
            <Table verticalSpacing="xs" horizontalSpacing="sm">
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>From</Table.Th>
                  <Table.Th>To</Table.Th>
                  <Table.Th ta="right">Tasks</Table.Th>
                  <Table.Th ta="right">Timers &amp; calls</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {moved.map((move) => (
                  <Table.Tr key={move.from}>
                    <Table.Td><Text size="xs" ff="monospace">{move.from}</Text></Table.Td>
                    <Table.Td>
                      <Group gap={4} wrap="nowrap">
                        <ArrowRight size={12} />
                        <Text size="xs" ff="monospace">{move.to}</Text>
                      </Group>
                    </Table.Td>
                    <Table.Td ta="right"><Text size="xs">{move.tasks}</Text></Table.Td>
                    <Table.Td ta="right"><Text size="xs">{move.jobs}</Text></Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </Stack>
        )}

        {carried.length > 0 && (
          <Group gap={4}>
            <Text size="xs" c="dimmed">Unchanged, because v{target?.version} still has them:</Text>
            {carried.map((move) => (
              <Badge key={move.from} size="xs" variant="light" color="gray">{move.from}</Badge>
            ))}
          </Group>
        )}

        <Stack gap={4}>
          <Group justify="space-between">
            <Text size="sm" fw={600}>Node mapping</Text>
            <Button
              size="compact-xs"
              variant="subtle"
              leftSection={<Plus size={12} />}
              onClick={() => setRows((current) => [...current, { id: nextRowId++, from: '', to: '' }])}
            >
              Add
            </Button>
          </Group>
          <Text size="xs" c="dimmed">
            Only needed where a node changed id. Anything you do not list is carried across
            unchanged.
          </Text>
          {rows.map((row, index) => (
            <Group key={row.id} gap="xs" wrap="nowrap">
              <TextInput
                size="xs"
                placeholder="node in v{source?.version}"
                aria-label={`Node in the old version, row ${index + 1}`}
                value={row.from}
                onChange={(event) => {
                  const value = event.currentTarget.value;
                  setRows((current) => current.map((r, i) => (i === index ? { ...r, from: value } : r)));
                }}
                style={{ flex: 1 }}
              />
              <ArrowRight size={14} />
              <TextInput
                size="xs"
                placeholder="node in the new version"
                aria-label={`Node in the new version, row ${index + 1}`}
                value={row.to}
                onChange={(event) => {
                  const value = event.currentTarget.value;
                  setRows((current) => current.map((r, i) => (i === index ? { ...r, to: value } : r)));
                }}
                style={{ flex: 1 }}
              />
              <Tooltip label="Remove this mapping">
                <ActionIcon
                  size="sm"
                  variant="subtle"
                  color="red"
                  aria-label={`Remove mapping row ${index + 1}`}
                  onClick={() => setRows((current) => current.filter((_, i) => i !== index))}
                >
                  <Trash2 size={14} />
                </ActionIcon>
              </Tooltip>
            </Group>
          ))}
        </Stack>

        <Group justify="flex-end">
          <Button variant="subtle" color="gray" onClick={onClose}>Cancel</Button>
          <Button
            color="orange"
            disabled={!ready}
            loading={apply.isPending}
            onClick={handleApply}
          >
            Move {plan?.instances ?? 0} {plan?.instances === 1 ? 'instance' : 'instances'}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
