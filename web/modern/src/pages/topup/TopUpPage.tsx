import { useState } from 'react'
import api from '@/lib/api'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

export function TopUpPage() {
  const [key, setKey] = useState('')
  const [result, setResult] = useState<string|number>('')

  const redeem = async () => {
    const res = await api.post('/user/topup', { key })
    if (res.data?.success) setResult(res.data.data)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader><CardTitle>Redeem Code</CardTitle></CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <Input placeholder="Enter redemption code" value={key} onChange={(e)=>setKey(e.target.value)} />
            <Button onClick={redeem}>Redeem</Button>
          </div>
          {result !== '' && <div className="mt-3 text-sm">Quota added: <span className="font-medium">{String(result)}</span></div>}
        </CardContent>
      </Card>
    </div>
  )
}

export default TopUpPage
