import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import api from '@/lib/api'

interface OptionRow {
  key: string
  value: string
}

export function SystemSettings() {
  const [options, setOptions] = useState<OptionRow[]>([])
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const res = await api.get('/option/')
      if (res.data?.success) setOptions(res.data.data || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const save = async (key: string, value: string) => {
    try {
      await api.put('/option/', { key, value })
      // Show success message
      console.log(`Saved ${key}: ${value}`)
    } catch (error) {
      console.error('Error saving option:', error)
    }
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>System Settings</CardTitle>
          <CardDescription>
            Configure system-wide settings and options.
          </CardDescription>
        </div>
        <Button variant="outline" onClick={load} disabled={loading}>
          Refresh
        </Button>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {options.map((opt, idx) => (
            <div key={idx} className="border rounded-lg p-4">
              <div className="text-sm font-medium text-muted-foreground mb-2">
                {opt.key}
              </div>
              <div className="flex gap-2">
                <Input
                  defaultValue={opt.value}
                  onBlur={(e) => save(opt.key, e.target.value)}
                  className="flex-1"
                />
                <Button
                  variant="outline"
                  onClick={(e) => {
                    const target = (e.currentTarget.previousSibling as HTMLInputElement)
                    save(opt.key, target.value)
                  }}
                >
                  Save
                </Button>
              </div>
            </div>
          ))}
          {!options.length && (
            <div className="col-span-full text-center text-sm text-muted-foreground py-8">
              No options available or insufficient permissions.
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export default SystemSettings
