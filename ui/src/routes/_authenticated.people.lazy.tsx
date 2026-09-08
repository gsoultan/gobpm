import { createLazyFileRoute } from '@tanstack/react-router'
import { useEffect } from 'react'

import { ParticipantList } from '../pages/ParticipantList'
import { useAppStore } from '../store/useAppStore'

export const Route = createLazyFileRoute('/_authenticated/people')({
  component: ParticipantListRoute,
})

function ParticipantListRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('people')
  }, [setActiveTab])

  return <ParticipantList />
}
