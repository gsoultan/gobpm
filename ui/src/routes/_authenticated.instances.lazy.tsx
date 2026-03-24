import { createLazyFileRoute, useNavigate } from '@tanstack/react-router'
import { InstanceList } from '../pages/InstanceList'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/instances')({
  component: InstanceListRoute,
})

function InstanceListRoute() {
  const { setActiveTab } = useAppStore()
  const navigate = useNavigate()

  useEffect(() => {
    setActiveTab('instances')
  }, [setActiveTab])

  const handleView = (instanceId: string, definitionId: string) => {
    navigate({
      to: '/designer',
      search: { instanceId, definitionId }
    })
  }

  return <InstanceList onViewInstance={handleView} />
}
