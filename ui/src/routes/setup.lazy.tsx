import { createLazyFileRoute, useNavigate } from '@tanstack/react-router'
import { Setup } from '../pages/Setup'

export const Route = createLazyFileRoute('/setup')({
  component: SetupRoute,
})

function SetupRoute() {
  const navigate = useNavigate()

  return <Setup onComplete={() => navigate({ to: '/login' })} />
}
