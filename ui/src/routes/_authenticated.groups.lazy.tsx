import { createLazyFileRoute } from '@tanstack/react-router'
import { GroupList } from '../pages/GroupList'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/groups')({
  component: GroupListRoute,
})

function GroupListRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('groups')
  }, [setActiveTab])

  return <GroupList />
}
