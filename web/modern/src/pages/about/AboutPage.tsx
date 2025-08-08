import { useEffect, useState } from 'react'
import api from '@/lib/api'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

export function AboutPage() {
  const [html, setHtml] = useState('')
  useEffect(() => { (async () => { const res = await api.get('/about'); if (res.data?.success) setHtml(res.data.data || '') })() }, [])
  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader><CardTitle>About</CardTitle></CardHeader>
        <CardContent>
          <div className="prose prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: html }} />
        </CardContent>
      </Card>
    </div>
  )
}

export default AboutPage
