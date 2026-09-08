import { Button, Group, Notification, Stack, Text } from '@mantine/core';
import { CloudOff, RefreshCw, Upload } from 'lucide-react';

import { useOutbox } from './useOutbox';
import { useOnlineStatus, useServiceWorker } from './useServiceWorker';

/**
 * The two things the service worker needs to say out loud.
 *
 * A new version is waiting, and the connection has gone. Both are fixed to the
 * bottom of the screen rather than shown as a toast, because a toast that
 * disappears after four seconds is no use for a state that persists.
 */
export function ServiceWorkerPrompt() {
  const { updateReady, update, dismiss } = useServiceWorker();
  const online = useOnlineStatus();
  const { summary, flush } = useOutbox();

  if (!updateReady && online && summary === '') {
    return null;
  }

  return (
    <Stack
      gap="xs"
      style={{ position: 'fixed', bottom: 16, right: 16, zIndex: 400, maxWidth: 360 }}
    >
      {!online && (
        <Notification
          icon={<CloudOff size={18} />}
          color="orange"
          title="You are offline"
          withCloseButton={false}
        >
          <Text size="sm">
            You can still read what has already loaded, and complete the tasks in
            your inbox. They will be sent when you are back online.
          </Text>
        </Notification>
      )}

      {/*
        Somebody's unsent work. Shown whether online or not: if it is still here
        while the connection is back, something is wrong and saying nothing
        would leave them believing an approval had gone through.
      */}
      {summary !== '' && (
        <Notification
          icon={<Upload size={18} />}
          color={online ? 'blue' : 'gray'}
          title={summary}
          withCloseButton={false}
        >
          <Stack gap="xs">
            <Text size="sm">
              {online
                ? 'Sending them now. They are kept here until the server confirms each one.'
                : 'Kept on this device until you are back online.'}
            </Text>
            {online && (
              <Group gap="xs">
                <Button size="xs" variant="light" onClick={() => void flush()}>
                  Try again now
                </Button>
              </Group>
            )}
          </Stack>
        </Notification>
      )}

      {updateReady && (
        <Notification
          icon={<RefreshCw size={18} />}
          color="blue"
          title="A new version is ready"
          onClose={dismiss}
        >
          <Stack gap="xs">
            <Text size="sm">
              Reload when you are at a good stopping point. Anything you are part
              way through typing is not saved yet.
            </Text>
            <Group gap="xs">
              <Button size="xs" onClick={update}>
                Reload now
              </Button>
              <Button size="xs" variant="subtle" color="gray" onClick={dismiss}>
                Later
              </Button>
            </Group>
          </Stack>
        </Notification>
      )}
    </Stack>
  );
}
