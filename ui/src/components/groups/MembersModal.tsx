/**
 * Who is in a group, and adding or removing someone.
 *
 * The picker holds the whole directory but renders only a handful of matches
 * at a time: a Select listing six hundred people is a scroll, not a choice.
 */
import { ActionIcon, Button, Group, Modal, Select, Stack, Table, Text, Tooltip } from '@mantine/core';
import { notifications } from '@mantine/notifications';
import { UserMinus, UserPlus } from 'lucide-react';
import { useState } from 'react';

import { useAddMembership, useGroupMembers, useRemoveMembership, useUsers } from '../../hooks/useUser';
import { errorMessage } from '../../services/shared/errors';
import type { ApiGroup } from '../../services/types';

/** How many matches the picker shows at once. */
const PICKER_LIMIT = 20;

interface MembersModalProps {
  group: ApiGroup | null;
  opened: boolean;
  onClose: () => void;
}

export function MembersModal({ group, opened, onClose }: MembersModalProps) {
  const { data: usersData } = useUsers();
  const { data: membersData, isLoading: membersLoading } = useGroupMembers(group?.id ?? '');
  const addMembership = useAddMembership();
  const removeMembership = useRemoveMembership();
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const allUsers = usersData?.users ?? [];
  const members = membersData?.users ?? [];
  const memberIds = new Set(members.map((m) => m.id));
  const candidates = allUsers
    .filter((u) => !memberIds.has(u.id))
    .map((u) => ({ value: u.id, label: `${u.full_name || u.username} (@${u.username})` }));

  const add = async () => {
    if (!selectedUserId || !group) return;
    try {
      await addMembership.mutateAsync({ groupId: group.id, userId: selectedUserId });
      setSelectedUserId(null);
      notifications.show({ title: 'Added', message: `They are now in ${group.name}.`, color: 'green' });
    } catch (error: unknown) {
      notifications.show({ title: 'Could not add them', message: errorMessage(error, 'Failed to add member'), color: 'red' });
    }
  };

  const remove = async (userId: string, who: string) => {
    if (!group) return;
    try {
      await removeMembership.mutateAsync({ groupId: group.id, userId });
      notifications.show({ title: 'Removed', message: `${who} is no longer in ${group.name}.`, color: 'green' });
    } catch (error: unknown) {
      notifications.show({ title: 'Could not remove them', message: errorMessage(error, 'Failed to remove member'), color: 'red' });
    }
  };

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={<Text fw={700}>Members of {group?.name}</Text>}
      radius="lg"
      size="lg"
    >
      <Stack gap="md">
        <Group align="flex-end">
          <Select
            label="Add someone"
            description={`Type a name to search ${candidates.length} people`}
            placeholder="Start typing a name…"
            data={candidates}
            value={selectedUserId}
            onChange={setSelectedUserId}
            searchable
            limit={PICKER_LIMIT}
            nothingFoundMessage="Nobody matches"
            style={{ flex: 1 }}
          />
          <Button leftSection={<UserPlus size={16} />} onClick={add} disabled={!selectedUserId} loading={addMembership.isPending}>
            Add
          </Button>
        </Group>

        {membersLoading ? (
          <Text c="dimmed">Loading members…</Text>
        ) : members.length === 0 ? (
          <Text c="dimmed" ta="center" py="md">Nobody is in this group yet.</Text>
        ) : (
          <Table verticalSpacing="sm" highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>User</Table.Th>
                <Table.Th>Email</Table.Th>
                <Table.Th ta="right">Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {members.map((m) => {
                const who = m.full_name || m.username;
                return (
                  <Table.Tr key={m.id}>
                    <Table.Td>
                      <Stack gap={0}>
                        <Text fw={600} size="sm">{who}</Text>
                        <Text size="xs" c="dimmed">@{m.username}</Text>
                      </Stack>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm">{m.email || '—'}</Text>
                    </Table.Td>
                    <Table.Td>
                      <Group justify="flex-end">
                        <Tooltip label="Remove from group">
                          <ActionIcon aria-label={`Remove ${who} from the group`} variant="light" color="red" onClick={() => remove(m.id, who)}>
                            <UserMinus size={16} />
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
    </Modal>
  );
}
