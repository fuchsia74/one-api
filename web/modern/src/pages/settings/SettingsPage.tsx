import { useEffect, useState } from 'react'
import api from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

interface OptionRow { key: string; value: string }

export function SettingsPage() {
  const [options, setOptions] = useState<OptionRow[]>([])
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try { const res = await api.get('/option/'); if (res.data?.success) setOptions(res.data.data || []) } finally { setLoading(false) }
  }
  useEffect(() => { load() }, [])

  const save = async (key: string, value: string) => {
    await api.put('/option/', { key, value })
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader className="flex items-center justify-between">
          <CardTitle>System Settings</CardTitle>
          <Button variant="outline" onClick={load} disabled={loading}>Refresh</Button>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {options.map((opt, idx) => (
              <div key={idx} className="border rounded p-3">
                <div className="text-xs text-muted-foreground mb-1">{opt.key}</div>
                <div className="flex gap-2">
                  <Input defaultValue={opt.value} onBlur={(e)=>save(opt.key, e.target.value)} />
                  <Button variant="outline" onClick={(e)=>{
                    const target = (e.currentTarget.previousSibling as HTMLInputElement); save(opt.key, target.value)
                  }}>Save</Button>
                </div>
              </div>
            ))}
            {!options.length && <div className="text-sm text-muted-foreground">No options or insufficient permission.</div>}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default SettingsPage
