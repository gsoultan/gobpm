import { RouterProvider, createRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'
import { useAppStore } from './store/useAppStore'
import './App.css'

// Create a new router instance
const router = createRouter({
  routeTree,
  context: {
    auth: undefined!, // will be updated in App component
  },
})

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

function App() {
  const auth = useAppStore()
  return <RouterProvider router={router} context={{ auth }} />
}

export default App
