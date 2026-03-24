import { createFileRoute, redirect } from '@tanstack/react-router'
import { processService } from '../services/api'

export const Route = createFileRoute('/setup')({
  beforeLoad: async () => {
    try {
      const { status } = await processService.getSetupStatus()
      if (status?.is_initialized) {
        throw redirect({ to: '/login' })
      }
    } catch (e) {
      // If it's a redirect, re-throw it
      if (e instanceof Error && 'to' in e) throw e
      // On network error, allow access (server might not be ready yet)
    }
  },
})
