import { Link } from 'react-router-dom'
import { useAuthStore } from '@/lib/stores/auth'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'

export function Header() {
  const { user, logout } = useAuthStore()

  return (
    <header className="border-b bg-background">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Link to="/" className="text-xl font-bold">
              One API
            </Link>
            {user && (
              <nav className="flex space-x-4">
                <Link
                  to="/dashboard"
                  className="text-sm font-medium hover:text-primary"
                >
                  Dashboard
                </Link>
                <Link
                  to="/channels"
                  className="text-sm font-medium hover:text-primary"
                >
                  Channels
                </Link>
                <Link
                  to="/users"
                  className="text-sm font-medium hover:text-primary"
                >
                  Users
                </Link>
                <Link
                  to="/tokens"
                  className="text-sm font-medium hover:text-primary"
                >
                  Tokens
                </Link>
                <Link
                  to="/logs"
                  className="text-sm font-medium hover:text-primary"
                >
                  Logs
                </Link>
                <Link
                  to="/redemptions"
                  className="text-sm font-medium hover:text-primary"
                >
                  Redemptions
                </Link>
                <Link to="/models" className="text-sm font-medium hover:text-primary">Models</Link>
                <Link to="/topup" className="text-sm font-medium hover:text-primary">Top Up</Link>
                <Link to="/about" className="text-sm font-medium hover:text-primary">About</Link>
                <Link to="/settings" className="text-sm font-medium hover:text-primary">Settings</Link>
              </nav>
            )}
          </div>
          <div className="flex items-center space-x-4">
      {user ? (
              <>
                <span className="text-sm text-muted-foreground">
                  Welcome, {user.display_name || user.username}
                </span>
        <Button variant="outline" onClick={async () => { await api.get('/user/logout'); logout(); }}>
                  Logout
                </Button>
              </>
            ) : (
              <Link to="/login" className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 text-sm text-white hover:opacity-90">Login</Link>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
