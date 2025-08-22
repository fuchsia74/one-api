import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { api } from '@/lib/api'

export function AboutPage() {
  const [about, setAbout] = useState('')
  const [aboutLoaded, setAboutLoaded] = useState(false)

  const loadAbout = async () => {
    try {
      // Load cached content first
      setAbout(localStorage.getItem('about') || '')

      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/about')
      const { success, data } = res.data

      if (success && data) {
        let aboutContent = data

        // If it's not a URL, assume it's markdown and convert it
        if (!data.startsWith('https://')) {
          // For now, we'll just use the content as HTML
          // In a real implementation, you might want to use a markdown parser
          aboutContent = data
        }

        setAbout(aboutContent)
        localStorage.setItem('about', aboutContent)
      } else {
        console.error('Failed to load about content')
        if (!about) {
          setAbout('About content failed to load.')
        }
      }
    } catch (error) {
      console.error('Error loading about content:', error)
      if (!about) {
        setAbout('About content failed to load.')
      }
    } finally {
      setAboutLoaded(true)
    }
  }

  useEffect(() => {
    loadAbout()
  }, [])

  // If about is a URL, render as iframe
  if (about.startsWith('https://')) {
    return (
      <iframe
        src={about}
        className="w-full h-screen border-0"
        title="About"
      />
    )
  }

  // If no about content is configured, show default
  if (aboutLoaded && !about) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardHeader>
            <CardTitle>About</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-4">Welcome to One API</h2>
              <p className="text-muted-foreground mb-4">
                One API provides a unified interface for accessing multiple AI models and services.
                This platform allows you to integrate with various AI providers through a single,
                consistent API.
              </p>
            </div>

            <div className="flex gap-4">
              <Button asChild>
                <Link to="/models">View Supported Models</Link>
              </Button>
              <Button variant="outline" asChild>
                <a href="https://github.com/Laisky/one-api" target="_blank" rel="noopener noreferrer">
                  GitHub Repository
                </a>
              </Button>
            </div>

            <div className="border-t pt-6">
              <h3 className="font-semibold mb-2">Features</h3>
              <ul className="space-y-1 text-sm text-muted-foreground">
                <li>• Unified API for multiple AI providers</li>
                <li>• Token-based usage tracking and billing</li>
                <li>• User management and role-based access</li>
                <li>• Channel management for different providers</li>
                <li>• Comprehensive logging and monitoring</li>
                <li>• Quota management and top-up system</li>
              </ul>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Render custom about content
  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>About</CardTitle>
            <Button asChild>
              <Link to="/models">View Supported Models</Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div
            className="prose prose-lg prose-headings:font-semibold prose-headings:tracking-tight prose-a:text-primary hover:prose-a:underline prose-pre:bg-muted/60 prose-code:before:content-[''] prose-code:after:content-[''] max-w-none dark:prose-invert"
            dangerouslySetInnerHTML={{ __html: about }}
          />
        </CardContent>
      </Card>
    </div>
  )
}

export default AboutPage
