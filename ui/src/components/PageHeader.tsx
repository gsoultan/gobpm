import { Group, Title, Text, Box, Stack } from '@mantine/core';
import React from 'react';

interface PageHeaderProps {
  title: string;
  description?: string;
  actions?: React.ReactNode;
}

export function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <Box 
      mb="xl" 
      pb="lg" 
      style={{ 
        borderBottom: '1px solid var(--mantine-color-gray-2)',
        backgroundColor: 'var(--mantine-color-white)',
        margin: '0 -24px 32px -24px',
        padding: '24px 24px 32px 24px'
      }}
    >
      <Group justify="space-between" align="center">
        <Stack gap={4}>
          <Title order={1} fw={800} lts={-0.5}>
            <Text span variant="gradient" gradient={{ from: 'blue.8', to: 'cyan.7' }} inherit>
              {title}
            </Text>
          </Title>
          {description && (
            <Text c="dimmed" size="md" fw={500}>
              {description}
            </Text>
          )}
        </Stack>
        {actions && (
          <Group gap="sm">
            {actions}
          </Group>
        )}
      </Group>
    </Box>
  );
}
