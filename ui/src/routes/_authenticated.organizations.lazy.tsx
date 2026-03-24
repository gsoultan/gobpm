import { createLazyFileRoute } from '@tanstack/react-router'
import { OrganizationList } from '../pages/OrganizationList'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/organizations')({
  component: OrganizationListRoute,
})

function OrganizationListRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('organizations')
  }, [setActiveTab])

  return <OrganizationList />
}
