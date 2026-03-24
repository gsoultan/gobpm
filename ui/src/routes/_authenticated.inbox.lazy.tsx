import { createLazyFileRoute } from '@tanstack/react-router'
import { TaskInbox } from '../pages/TaskInbox'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/inbox')({
  component: TaskInboxRoute,
})

function TaskInboxRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('inbox')
  }, [setActiveTab])

  return <TaskInbox />
}
