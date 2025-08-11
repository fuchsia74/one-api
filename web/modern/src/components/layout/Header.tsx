import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/lib/stores/auth'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/theme-toggle'
import { NavigationDrawer } from '@/components/ui/mobile-drawer'
import { useResponsive } from '@/hooks/useResponsive'
import {
  Menu,
  Home,
  Settings,
  Users,
  CreditCard,
  BarChart3,
  MessageSquare,
  Info,
  Zap,
  Gift,
  DollarSign,
  FileText
} from 'lucide-react'
import { useState } from 'react'

// Icon mapping for navigation items
const navigationIcons = {
  '/dashboard': Home,
  '/channels': Zap,
  '/tokens': CreditCard,
  '/logs': FileText,
  '/users': Users,
  '/redemptions': Gift,
  '/topup': DollarSign,
  '/models': BarChart3,
  '/chat': MessageSquare,
  '/about': Info,
  '/settings': Settings,
}

export function Header() {
  const { user, logout } = useAuthStore()
  const location = useLocation()
  const navigate = useNavigate()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const { isMobile, isTablet } = useResponsive()

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
  ].filter(item => item.show).map(item => ({
    ...item,
    href: item.to,
    icon: navigationIcons[item.to as keyof typeof navigationIcons],
    isActive: location.pathname === item.to
  }))

  const isActivePage = (path: string) => location.pathname === path

  const handleLogout = async () => {
    try {
      await api.get('/user/logout')
      logout()
      navigate('/login')
    } catch (error) {
      console.error('Logout failed:', error)
      // Force logout even if API call fails
      logout()
      navigate('/login')
    }
  }

  return (
    <header className="border-b bg-background/95 backdrop-blur-sm sticky top-0 z-50">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          {/* Logo and Brand */}
          <div className="flex items-center space-x-4">
            <Link
              to="/"
              className="text-xl font-bold hover:text-primary transition-colors"
            >
              {localStorage.getItem('system_name') || 'OneAPI'}
            </Link>

            {/* Desktop Navigation - Only show on large screens */}
            {user && !isMobile && !isTablet && (
              <nav className="hidden lg:flex items-center space-x-1">
                {navigationItems.map((item) => (
                  <Link
                    key={item.to}
                    to={item.to}
                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                      isActivePage(item.to)
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                    }`}
                  >
                    {item.name}
                  </Link>
                ))}
              </nav>
            )}
          </div>

          {/* Actions and User Menu */}
          <div className="flex items-center space-x-2">
            <ThemeToggle />

            {user ? (
              <>
                {/* User Welcome - Hide on mobile */}
                <span className="hidden md:inline text-sm text-muted-foreground truncate max-w-32">
                  Welcome, {user.display_name || user.username}
                </span>

                {/* Logout Button - Responsive sizing */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleLogout}
                  className="hidden sm:inline-flex"
                >
                  Logout
                </Button>

                {/* Mobile menu button - Show when navigation is hidden */}
                {(isMobile || isTablet) && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setMobileMenuOpen(true)}
                    className="lg:hidden touch-target"
                    aria-label="Open navigation menu"
                  >
                    <Menu className="h-5 w-5" />
                  </Button>
                )}
              </>
            ) : (
              <div className="flex items-center space-x-2">
                <Link
                  to="/register"
                  className={`font-medium text-muted-foreground hover:text-primary transition-colors ${
                    isMobile ? 'text-sm' : 'text-sm'
                  }`}
                >
                  Register
                </Link>
                <Button
                  asChild
                  size="sm"
                  className="touch-target"
                >
                  <Link to="/login">
                    Login
                  </Link>
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Mobile Navigation Drawer */}
      {user && (
        <NavigationDrawer
          isOpen={mobileMenuOpen}
          onClose={() => setMobileMenuOpen(false)}
          navigationItems={navigationItems}
          title="Navigation"
        />
      )}

      {/* Mobile Logout Button in Drawer */}
      {user && mobileMenuOpen && (
        <div className="fixed bottom-4 left-4 right-4 z-60 sm:hidden">
          <Button
            variant="outline"
            onClick={handleLogout}
            className="w-full touch-target"
          >
            Logout
          </Button>
        </div>
      )}
    </header>
  )
}
