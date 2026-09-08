/**
 * Defines a connector template by hand: its name, its key and the fields a
 * connection of it needs. Expert-mode only — it is a schema editor.
 */
import {
  ActionIcon,
  Button,
  Divider,
  Group,
  Modal,
  Paper,
  Select,
  SimpleGrid,
  Stack,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { Plus, Trash2 } from 'lucide-react';

import type { ApiConnectorProperty, CreateConnectorPayload } from '../../services/types';
import { CONNECTOR_ICONS } from './connectorIcons';

const MIN_NAME_LENGTH = 2;
const CATEGORIES = ['utility', 'social', 'messaging', 'crm', 'erp'];
const FIELD_TYPES = ['string', 'password', 'number', 'select', 'textarea', 'boolean'];

interface CustomConnectorModalProps {
  opened: boolean;
  onClose: () => void;
  onSubmit: (values: CreateConnectorPayload) => Promise<void>;
  pending: boolean;
}

export function CustomConnectorModal({ opened, onClose, onSubmit, pending }: CustomConnectorModalProps) {
  const form = useForm({
    initialValues: {
      name: '',
      key: '',
      description: '',
      icon: 'Zap',
      type: 'utility',
      schema: [] as ApiConnectorProperty[],
    },
    validate: {
      name: (value) => (value.length < MIN_NAME_LENGTH ? 'Name is too short' : null),
      key: (value) => (value.length < MIN_NAME_LENGTH ? 'Key is too short' : null),
    },
  });

  const submit = async (values: typeof form.values) => {
    await onSubmit(values);
    form.reset();
  };

  const addProperty = () =>
    form.insertListItem('schema', { key: '', label: '', type: 'string', required: false });

  return (
    <Modal opened={opened} onClose={onClose} title="Create a connector template" size="lg" radius="lg">
      <form onSubmit={form.onSubmit(submit)}>
        <Stack gap="md">
          <SimpleGrid cols={2}>
            <TextInput label="Connector name" placeholder="e.g. My Custom API" required {...form.getInputProps('name')} />
            <TextInput
              label="Key"
              description="What a service task names to use it"
              placeholder="e.g. my-custom-api"
              required
              {...form.getInputProps('key')}
            />
          </SimpleGrid>
          <Textarea label="Description" placeholder="What does this connector do?" {...form.getInputProps('description')} />
          <SimpleGrid cols={2}>
            <Select label="Icon" data={Object.keys(CONNECTOR_ICONS)} {...form.getInputProps('icon')} />
            <Select label="Category" data={CATEGORIES} {...form.getInputProps('type')} />
          </SimpleGrid>

          <Divider label="Settings a connection needs" labelPosition="center" />
          <Text size="xs" c="dimmed">Each field here is asked for when someone sets up a connection of this kind.</Text>

          <Stack gap="xs">
            {form.values.schema.map((_, index) => (
              <Paper key={index} withBorder p="xs" radius="md">
                <Group gap="xs" grow align="flex-end">
                  <TextInput label="Field key" placeholder="url" size="xs" {...form.getInputProps(`schema.${index}.key`)} />
                  <TextInput label="Label" placeholder="Target URL" size="xs" {...form.getInputProps(`schema.${index}.label`)} />
                  <Select label="Type" size="xs" data={FIELD_TYPES} {...form.getInputProps(`schema.${index}.type`)} />
                  <ActionIcon
                    aria-label={`Remove field ${form.values.schema[index].key || index + 1}`}
                    color="red"
                    variant="light"
                    onClick={() => form.removeListItem('schema', index)}
                  >
                    <Trash2 size={14} />
                  </ActionIcon>
                </Group>
              </Paper>
            ))}
            <Button variant="light" size="xs" leftSection={<Plus size={14} />} onClick={addProperty}>
              Add a field
            </Button>
          </Stack>

          <Group justify="flex-end" mt="xl">
            <Button variant="default" onClick={onClose}>Cancel</Button>
            <Button type="submit" color="indigo" loading={pending}>Create template</Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
