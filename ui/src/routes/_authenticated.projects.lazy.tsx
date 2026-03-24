import { createLazyFileRoute } from '@tanstack/react-router'
import { ProjectList } from '../pages/ProjectList'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/projects')({
  component: ProjectListRoute,
})

function ProjectListRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('projects')
  }, [setActiveTab])

  return <ProjectList />
}
