import { 
  Title, 
  Text, 
  Paper, 
  Stack, 
  Group, 
  Switch, 
  Divider, 
  Box,
  SimpleGrid,
  ActionIcon
} from '@mantine/core';
import { 
  Moon, 
  Sun, 
  ShieldCheck,
  ShieldOff
} from 'lucide-react';
import { useAppStore } from '../store/useAppStore';
import { EnvironmentSettings } from '../components/EnvironmentSettings';
import { PageHeader } from '../components/PageHeader';
import { ComingSoonButton } from '../components/state/ComingSoon';

export function Settings() {
  // Environments are administrative: the list alone says where every runtime's
  // database lives. The server refuses a non-admin either way; hiding it here
  // is so nobody is shown a control that will only ever say no.
  const isAdmin = useAppStore((state) => state.user?.role === 'ADMIN');
  const theme = useAppStore((state) => state.theme);
  const toggleTheme = useAppStore((state) => state.toggleTheme);
  const expertMode = useAppStore((state) => state.expertMode);
  const setExpertMode = useAppStore((state) => state.setExpertMode);

  return (
    <Stack gap="xl">
      <PageHeader 
        title="Application Settings" 
        description="Configure your workspace and preferences."
      />

      {/*
        Full width, above the preference grid: an environment names a database
        and a port, which is infrastructure rather than a preference, and it is
        the thing an administrator opens this page to change.
      */}
      {isAdmin && <EnvironmentSettings />}

      <SimpleGrid cols={{ base: 1, md: 2 }} spacing="xl">
        <Stack gap="lg">
          <Paper p="xl" radius="lg" withBorder shadow="sm">
            <Title order={2} size="h5" mb="lg">Appearance</Title>
            <Stack gap="md">
              <Group justify="space-between">
                <Box>
                  <Text fw={600} size="sm">Interface Theme</Text>
                  <Text size="xs" c="dimmed">Choose between light and dark mode</Text>
                </Box>
                <Group gap={0}>
                  <ActionIcon aria-label="Use light theme" 
                    variant={theme === 'light' ? 'filled' : 'light'} 
                    onClick={() => theme === 'dark' && toggleTheme()}
                    size="lg"
                    radius="md"
                  >
                    <Sun size={18} />
                  </ActionIcon>
                  <ActionIcon aria-label="Use dark theme" 
                    variant={theme === 'dark' ? 'filled' : 'light'} 
                    onClick={() => theme === 'light' && toggleTheme()}
                    size="lg"
                    radius="md"
                    ml="xs"
                  >
                    <Moon size={18} />
                  </ActionIcon>
                </Group>
              </Group>
              
              <Divider />

              <Group justify="space-between">
                <Box>
                  <Text fw={600} size="sm">Expert Mode</Text>
                  <Text size="xs" c="dimmed">Show advanced technical settings and schemas</Text>
                </Box>
                <Group gap="xs">
                  {expertMode ? <ShieldCheck size={16} color="green" /> : <ShieldOff size={16} color="gray" />}
                  <Switch 
                    aria-label="Expert mode"
                    checked={expertMode} 
                    onChange={(event) => setExpertMode(event.currentTarget.checked)} 
                    size="md" 
                  />
                </Group>
              </Group>
            </Stack>
          </Paper>
          {/*
            A "Notifications" panel with email and push switches stood here.
            Neither was wired to anything — one even rendered as already on —
            so a user could turn on email notifications, get none, and not know
            whether the feature or their mail was broken. Notifications arrive
            in the bell in the header; there is nothing to configure yet.
          */}
        </Stack>

        <Stack gap="lg">
          <Paper p="xl" radius="lg" withBorder shadow="sm">
            <Title order={2} size="h5" mb="lg">Security &amp; API</Title>
            <Stack gap="md">
              <Group justify="space-between">
                <Box>
                  <Text fw={600} size="sm">Two-Factor Authentication</Text>
                  <Text size="xs" c="dimmed">Add an extra layer of security to your account</Text>
                </Box>
                <ComingSoonButton variant="light" color="blue" size="xs" label="Two-factor authentication is not implemented yet">Enable</ComingSoonButton>
              </Group>
              <Divider />
              <Group justify="space-between">
                <Box>
                  <Text fw={600} size="sm">API Keys</Text>
                  <Text size="xs" c="dimmed">Manage tokens for external API access</Text>
                </Box>
                <ComingSoonButton variant="outline" color="gray" size="xs" label="API keys are not available yet">Manage</ComingSoonButton>
              </Group>
            </Stack>
          </Paper>

          <Paper p="xl" radius="lg" withBorder shadow="sm" style={{ borderColor: 'var(--mantine-color-red-2)' }}>
            <Title order={2} size="h5" mb="lg" c="red">Danger Zone</Title>
            <Stack gap="md">
              <Group justify="space-between">
                <Box>
                  <Text fw={600} size="sm">Clear Cache</Text>
                  <Text size="xs" c="dimmed">Reset local storage and application data</Text>
                </Box>
                <ComingSoonButton variant="light" color="red" size="xs" label="Clearing local application data is not implemented yet">Clear</ComingSoonButton>
              </Group>
              {/*
                A red "Delete Account" button stood here with no handler. There
                is no self-service deletion; an administrator removes accounts.
              */}
            </Stack>
          </Paper>
        </Stack>
      </SimpleGrid>
      
      {/*
        Neither of these buttons was ever wired up, so the entire page was
        decorative: a user could change a setting, press Save, and get no
        feedback of any kind. Settings that DO persist (theme, expert mode)
        already save themselves through the store on change, which is why the
        page needs no save button once the unwired ones are removed.
      */}
      <Group justify="flex-end" mt="xl">
        <ComingSoonButton variant="default" label="Restoring defaults is not implemented yet">
          Reset to Defaults
        </ComingSoonButton>
      </Group>
    </Stack>
  );
}
