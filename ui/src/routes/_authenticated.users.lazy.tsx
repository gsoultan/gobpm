import { createLazyFileRoute } from '@tanstack/react-router'
import { UserList } from '../pages/UserList'
import { useAppStore } from '../store/useAppStore'
import { useEffect } from 'react'

export const Route = createLazyFileRoute('/_authenticated/users')({
  component: UserListRoute,
})

function UserListRoute() {
  const { setActiveTab } = useAppStore()
  useEffect(() => {
    setActiveTab('users')
  }, [setActiveTab])

  return <UserList />
}
