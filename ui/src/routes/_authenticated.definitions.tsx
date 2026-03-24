import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/definitions')({
  beforeLoad: () => {
    throw redirect({ to: '/models', search: { tab: 'processes' } })
  }
})
