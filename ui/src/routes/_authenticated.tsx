import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { MainLayout } from '../containers/MainLayout'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { useAppStore } from '../store/useAppStore'
import { processService } from '../services/api'

export const Route = createFileRoute('/_authenticated')({
  component: AuthenticatedLayout,
  beforeLoad: async ({ location }) => {
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

    // If not authenticated, redirect to login
    const { user, token } = useAppStore.getState()
    if (!user || !token) {
      throw redirect({
        to: '/login',
        search: {
          redirect: location.href,
        },
      })
    }
  },
})

function AuthenticatedLayout() {
  const { activeTab, setActiveTab } = useAppStore()

  return (
    <MainLayout activeTab={activeTab} onTabChange={setActiveTab}>
      <ErrorBoundary>
        <Outlet />
      </ErrorBoundary>
    </MainLayout>
  )
}
