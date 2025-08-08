import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function HomePage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4">Welcome to One API</h1>
          <p className="text-xl text-muted-foreground">
            Modern AI API management platform
          </p>
        </div>

        <div className="grid md:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Modern UI</CardTitle>
              <CardDescription>
                Built with shadcn/ui and Tailwind CSS
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Experience a clean, modern interface with excellent mobile support.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Better Performance</CardTitle>
              <CardDescription>
                Powered by Vite and React 18
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Fast loading times and smooth user experience.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Enhanced Accessibility</CardTitle>
              <CardDescription>
                WCAG 2.1 AA compliant
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">
                Accessible to all users with keyboard navigation and screen reader support.
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="text-center mt-8">
          <Button size="lg">
            Get Started
          </Button>
        </div>
      </div>
    </div>
  )
}
