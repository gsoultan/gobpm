import { createLazyFileRoute } from '@tanstack/react-router'
import { Connectors } from '../pages/Connectors'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/connectors')({
  component: ConnectorsRoute,
})

function ConnectorsRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('connectors')
  }, [setActiveTab])

  return <Connectors />
}
