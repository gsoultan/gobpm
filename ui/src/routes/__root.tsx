import { createRootRoute, Outlet, ScrollRestoration } from '@tanstack/react-router'
import { MantineProvider, createTheme, rem, Card, Button, Table, Paper, ActionIcon, Badge, TextInput } from '@mantine/core'
import { Notifications } from '@mantine/notifications'
import { useAppStore } from '../store/useAppStore'
import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'

const mantineTheme = createTheme({
  primaryColor: 'blue',
  primaryShade: 6,
  defaultRadius: 'md',
  fontFamily: 'Inter, system-ui, sans-serif',
  headings: {
    fontFamily: 'Inter, system-ui, sans-serif',
    fontWeight: '800',
  },
  shadows: {
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
  },
  components: {
    Card: Card.extend({
      defaultProps: {
        withBorder: true,
        padding: 'xl',
        radius: 'lg',
        shadow: 'md',
      },
      styles: {
        root: {
          transition: 'transform 200ms ease, shadow 200ms ease',
          '&:hover': {
            transform: 'translateY(-2px)',
            boxShadow: 'var(--mantine-shadow-lg)',
          }
        }
      }
    }),
    Button: Button.extend({
      defaultProps: {
        radius: 'md',
        fw: 600,
      }
    }),
    Table: Table.extend({
      defaultProps: {
        verticalSpacing: 'md',
        horizontalSpacing: 'xl',
      },
      styles: {
        thead: {
          backgroundColor: 'var(--mantine-color-gray-0)',
        },
        th: {
          textTransform: 'uppercase',
          fontSize: rem(11),
          letterSpacing: rem(1),
          fontWeight: 700,
          color: 'var(--mantine-color-dimmed)',
        }
      }
    }),
    Paper: Paper.extend({
      defaultProps: {
        radius: 'md',
        withBorder: true,
      }
    }),
    ActionIcon: ActionIcon.extend({
      defaultProps: {
        radius: 'md',
        variant: 'light',
      }
    }),
    Badge: Badge.extend({
      defaultProps: {
        radius: 'sm',
        fw: 700,
      }
    }),
    TextInput: TextInput.extend({
      defaultProps: {
        radius: 'md',
      }
    })
  }
})

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  const { theme } = useAppStore()

  return (
    <MantineProvider theme={mantineTheme} forceColorScheme={theme}>
      <Notifications />
      <Outlet />
      <ScrollRestoration />
    </MantineProvider>
  )
}
