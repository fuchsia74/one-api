import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Home } from 'lucide-react'

export function NotFoundPage() {
  const navigate = useNavigate()
  const [seconds, setSeconds] = useState(5)

  useEffect(() => {
    const tick = setInterval(() => setSeconds((s) => Math.max(0, s - 1)), 1000)
    const timer = setTimeout(() => navigate('/', { replace: true }), 5000)
    return () => {
      clearInterval(tick)
      clearTimeout(timer)
    }
  }, [navigate])

  return (
    <div className="flex flex-col items-center justify-center text-center py-16 gap-6">
      <div>
        <h1 className="text-4xl font-bold">404</h1>
        <p className="text-muted-foreground mt-2">Page not found</p>
      </div>

      <p className="text-sm text-muted-foreground">
        Redirecting to home in {seconds}s...
      </p>

      <div className="flex items-center gap-3">
        <Button asChild>
          <Link to="/">
            <Home className="mr-2 h-4 w-4" /> Go Home Now
          </Link>
        </Button>
        <Button variant="outline" onClick={() => navigate(-1)}>Go Back</Button>
      </div>
    </div>
  )
}

export default NotFoundPage
