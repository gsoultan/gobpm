import { createLazyFileRoute } from '@tanstack/react-router'
import { TaskList } from '../pages/TaskList'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/tasks')({
  component: TaskListRoute,
})

function TaskListRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('tasks')
  }, [setActiveTab])

  return <TaskList />
}
