import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/lib/stores/auth'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Menu, X } from 'lucide-react'
import { useState } from 'react'

export function Header() {
  const { user, logout } = useAuthStore()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const isAdmin = user?.role >= 10
  const chatLink = localStorage.getItem('chat_link')

  const navigationItems = [
    { name: 'Dashboard', to: '/dashboard', show: true },
    { name: 'Channels', to: '/channels', show: isAdmin },
    { name: 'Tokens', to: '/tokens', show: true },
    { name: 'Logs', to: '/logs', show: true },
    { name: 'Users', to: '/users', show: isAdmin },
    { name: 'Redemptions', to: '/redemptions', show: isAdmin },
    { name: 'Top Up', to: '/topup', show: !isAdmin },
    { name: 'Models', to: '/models', show: true },
    { name: 'Chat', to: '/chat', show: !!chatLink },
    { name: 'About', to: '/about', show: true },
    { name: 'Settings', to: '/settings', show: isAdmin },
  ].filter(item => item.show)

  const isActivePage = (path: string) => location.pathname === path

  return (
    <header className="border-b bg-background sticky top-0 z-50">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Link to="/" className="text-xl font-bold">
              {localStorage.getItem('system_name') || 'OneAPI'}
            </Link>

            {/* Desktop Navigation */}
            {user && (
              <nav className="hidden lg:flex space-x-4">
                {navigationItems.map((item) => (
                  <Link
                    key={item.to}
                    to={item.to}
                    className={`text-sm font-medium transition-colors ${
                      isActivePage(item.to)
                        ? 'text-primary'
                        : 'text-muted-foreground hover:text-primary'
                    }`}
                  >
                    {item.name}
                  </Link>
                ))}
              </nav>
            )}
          </div>

          <div className="flex items-center space-x-4">
            {user ? (
              <>
                <span className="hidden md:inline text-sm text-muted-foreground">
                  Welcome, {user.display_name || user.username}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={async () => {
                    await api.get('/user/logout');
                    logout();
                  }}
                >
                  Logout
                </Button>

                {/* Mobile menu button */}
                <Button
                  variant="ghost"
                  size="sm"
                  className="lg:hidden"
                  onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                >
                  {mobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
                </Button>
              </>
            ) : (
              <div className="flex items-center space-x-2">
                <Link
                  to="/register"
                  className="text-sm font-medium text-muted-foreground hover:text-primary"
                >
                  Register
                </Link>
                <Link
                  to="/login"
                  className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 text-sm text-primary-foreground hover:bg-primary/90"
                >
                  Login
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Mobile Navigation */}
        {user && mobileMenuOpen && (
          <nav className="lg:hidden mt-4 pb-4 border-t pt-4">
            <div className="flex flex-col space-y-2">
              {navigationItems.map((item) => (
                <Link
                  key={item.to}
                  to={item.to}
                  className={`text-sm font-medium py-2 px-3 rounded-md transition-colors ${
                    isActivePage(item.to)
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:text-primary hover:bg-muted'
                  }`}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {item.name}
                </Link>
              ))}
            </div>
          </nav>
        )}
      </div>
    </header>
  )
}
