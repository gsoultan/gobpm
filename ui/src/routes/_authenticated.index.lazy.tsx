import { createLazyFileRoute } from '@tanstack/react-router'
import { Dashboard } from '../pages/Dashboard'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/')({
  component: DashboardRoute,
})

function DashboardRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('dashboard')
  }, [setActiveTab])

  return <Dashboard />
}
