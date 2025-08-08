import { useEffect, useState } from 'react'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface ModelInfo { id?: string; object?: string; data?: any }

export function ModelsPage() {
  const [models, setModels] = useState<any[]>([])
  useEffect(()=>{ (async()=>{ const res = await api.get('/models/display'); if (res.data?.success) setModels(res.data.data || []) })() }, [])
  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader><CardTitle>Available Models</CardTitle></CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {models.map((m, idx) => (
              <div key={idx} className="border rounded p-3">
                <div className="font-medium">{m.name || m.id}</div>
                {m.display_name && <div className="text-sm text-muted-foreground">{m.display_name}</div>}
                {m.provider && <div className="text-xs mt-1">Provider: {m.provider}</div>}
              </div>
            ))}
            {!models.length && <div className="text-sm text-muted-foreground">No models</div>}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default ModelsPage
