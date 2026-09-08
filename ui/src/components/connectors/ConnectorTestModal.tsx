/**
 * Sends a sample payload through a connector with the form's current settings.
 *
 * A stored secret is not available to a test: the execute endpoint takes the
 * config it is given and has no connection to fill a sentinel from. So when
 * any setting still reads "unchanged", the test is refused here with the
 * reason, rather than sending the literal word to the third party.
 */
import { Alert, Box, Button, Modal, Stack, Text, Textarea } from '@mantine/core';
import { AlertCircle, CheckCircle2 } from 'lucide-react';

import { unchangedKeys } from '../../domain/connectorSecrets';

export type TestResult = { success: true; data: unknown } | { success: false; error: string };

interface ConnectorTestModalProps {
  opened: boolean;
  onClose: () => void;
  connectorName?: string;
  config: Record<string, unknown>;
  payload: string;
  onPayloadChange: (payload: string) => void;
  onExecute: () => void;
  pending: boolean;
  result: TestResult | null;
}

export function ConnectorTestModal({
  opened,
  onClose,
  connectorName,
  config,
  payload,
  onPayloadChange,
  onExecute,
  pending,
  result,
}: ConnectorTestModalProps) {
  const withheld = unchangedKeys(config);

  return (
    <Modal opened={opened} onClose={onClose} title={`Test ${connectorName ?? 'connector'}`} size="md">
      <Stack gap="md">
        {withheld.length > 0 && (
          <Alert icon={<AlertCircle size={16} />} color="yellow" title="Stored credentials are not sent to a test">
            <Text size="xs">
              {withheld.join(', ')} {withheld.length === 1 ? 'is' : 'are'} kept on the server and cannot be read back.
              Enter {withheld.length === 1 ? 'it' : 'them'} in the form to test with{' '}
              {withheld.length === 1 ? 'it' : 'them'}.
            </Text>
          </Alert>
        )}
        <Textarea
          label="Test payload (JSON)"
          description="Sample data to send through the connector"
          value={payload}
          onChange={(e) => onPayloadChange(e.currentTarget.value)}
          styles={{ input: { fontFamily: 'monospace' } }}
          minRows={5}
          autosize
        />
        <Button fullWidth onClick={onExecute} loading={pending} disabled={withheld.length > 0}>
          Run test
        </Button>

        {result && (
          <Alert
            icon={result.success ? <CheckCircle2 size={16} /> : <AlertCircle size={16} />}
            color={result.success ? 'green' : 'red'}
            title={result.success ? 'It worked' : 'It failed'}
          >
            <Box mt="xs">
              <Text size="xs" component="pre" style={{ whiteSpace: 'pre-wrap' }}>
                {result.success ? JSON.stringify(result.data, null, 2) : result.error}
              </Text>
            </Box>
          </Alert>
        )}
      </Stack>
    </Modal>
  );
}
