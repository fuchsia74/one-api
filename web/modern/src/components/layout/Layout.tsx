import { Outlet } from 'react-router-dom'
import { Header } from './Header'
import { Footer } from './Footer'
import { useResponsive } from '@/hooks/useResponsive'
import { cn } from '@/lib/utils'

export function Layout() {
  const { isMobile } = useResponsive()

  return (
    <div className="flex flex-col min-h-screen bg-background">
      <Header />

      <main className={cn(
        'flex-1 w-full',
        // Responsive padding and spacing
        isMobile ? 'px-2 py-4' : 'px-4 py-6',
        // Ensure proper spacing from header
        'mt-0'
      )}>
        <div className="w-full max-w-full">
          <Outlet />
        </div>
      </main>

      <Footer />
    </div>
  )
}
