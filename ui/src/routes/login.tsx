import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAppStore } from '../store/useAppStore'
import { processService } from '../services/api'
import { z } from 'zod'

const loginSearchSchema = z.object({
  redirect: z.string().optional(),
})

export const Route = createFileRoute('/login')({
  validateSearch: loginSearchSchema,
  beforeLoad: async () => {
    // If system is not configured, redirect to setup
    try {
      const { status } = await processService.getSetupStatus()
      if (!status?.is_initialized) {
        throw redirect({ to: '/setup' })
      }
    } catch (e) {
      if (e instanceof Error && 'to' in e) throw e
      // On network error, assume not configured and redirect to setup
      throw redirect({ to: '/setup' })
    }

    // If already authenticated, redirect to home
    const { user, token } = useAppStore.getState()
    if (user && token) {
      throw redirect({ to: '/' })
    }
  },
})
