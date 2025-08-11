import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'

interface ModelData {
  input_price: number
  output_price: number
  max_tokens: number
}

interface ChannelInfo {
  models: Record<string, ModelData>
}

interface ModelsData {
  [channelName: string]: ChannelInfo
}

export function ModelsPage() {
  const [modelsData, setModelsData] = useState<ModelsData>({})
  const [filteredData, setFilteredData] = useState<ModelsData>({})
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedChannels, setSelectedChannels] = useState<string[]>([])

  const fetchModelsData = async () => {
    try {
      setLoading(true)
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/models/display')
      const { success, message, data } = res.data
      if (success) {
        setModelsData(data || {})
        setFilteredData(data || {})
      } else {
        console.error('Failed to fetch models:', message)
      }
    } catch (error) {
      console.error('Error fetching models:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchModelsData()
  }, [])

  useEffect(() => {
    let filtered = { ...modelsData }

    // Filter by selected channels
    if (selectedChannels.length > 0) {
      const channelFiltered: ModelsData = {}
      selectedChannels.forEach(channelName => {
        if (filtered[channelName]) {
          channelFiltered[channelName] = filtered[channelName]
        }
      })
      filtered = channelFiltered
    }

    // Filter by search term
    if (searchTerm) {
      const searchFiltered: ModelsData = {}
      Object.keys(filtered).forEach(channelName => {
        const channelData = filtered[channelName]
        const filteredModels: Record<string, ModelData> = {}

        Object.keys(channelData.models).forEach(modelName => {
          if (modelName.toLowerCase().includes(searchTerm.toLowerCase())) {
            filteredModels[modelName] = channelData.models[modelName]
          }
        })

        if (Object.keys(filteredModels).length > 0) {
          searchFiltered[channelName] = {
            ...channelData,
            models: filteredModels
          }
        }
      })
      filtered = searchFiltered
    }

    setFilteredData(filtered)
  }, [searchTerm, selectedChannels, modelsData])

  const formatPrice = (price: number): string => {
    if (price === 0) return 'Free'
    if (price < 0.001) return `$${price.toFixed(6)}`
    if (price < 1) return `$${price.toFixed(4)}`
    return `$${price.toFixed(2)}`
  }

  const formatMaxTokens = (maxTokens: number): string => {
    if (maxTokens === 0) return 'Unlimited'
    if (maxTokens >= 1000000) return `${(maxTokens / 1000000).toFixed(1)}M`
    if (maxTokens >= 1000) return `${(maxTokens / 1000).toFixed(0)}K`
    return maxTokens.toString()
  }

  const toggleChannelFilter = (channelName: string) => {
    if (selectedChannels.includes(channelName)) {
      setSelectedChannels(selectedChannels.filter(ch => ch !== channelName))
    } else {
      setSelectedChannels([...selectedChannels, channelName])
    }
  }

  const clearFilters = () => {
    setSearchTerm('')
    setSelectedChannels([])
  }

  const renderChannelModels = (channelName: string, channelInfo: ChannelInfo) => {
    const models = Object.keys(channelInfo.models).map(modelName => ({
      model: modelName,
      inputPrice: channelInfo.models[modelName].input_price,
      outputPrice: channelInfo.models[modelName].output_price,
      maxTokens: channelInfo.models[modelName].max_tokens
    }))

    return (
      <Card key={channelName} className="mb-6">
        <CardHeader>
          <CardTitle className="text-lg">
            {channelName} ({models.length} models)
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2 px-3 font-medium">Model</th>
                  <th className="text-left py-2 px-3 font-medium">Input Price (per 1M tokens)</th>
                  <th className="text-left py-2 px-3 font-medium">Output Price (per 1M tokens)</th>
                  <th className="text-left py-2 px-3 font-medium">Max Tokens</th>
                </tr>
              </thead>
              <tbody>
                {models.map(model => (
                  <tr key={model.model} className="border-b hover:bg-muted/50">
                    <td className="py-2 px-3 font-mono text-sm">{model.model}</td>
                    <td className="py-2 px-3">{formatPrice(model.inputPrice)}</td>
                    <td className="py-2 px-3">{formatPrice(model.outputPrice)}</td>
                    <td className="py-2 px-3">{formatMaxTokens(model.maxTokens)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            <span className="ml-3">Loading models...</span>
          </CardContent>
        </Card>
      </div>
    )
  }

  const totalModels = Object.values(filteredData).reduce((total, channelInfo) =>
    total + Object.keys(channelInfo.models).length, 0
  )

  const channelOptions = Object.keys(modelsData).sort()

  return (
    <div className="container mx-auto px-4 py-8">
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Supported Models</CardTitle>
          <CardDescription>
            Browse all models supported by the server, grouped by channel/adaptor with pricing information.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <div className="md:col-span-1">
              <Input
                placeholder="Search models..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
            <div className="md:col-span-1">
              <div className="flex flex-wrap gap-2">
                {channelOptions.map(channelName => (
                  <Badge
                    key={channelName}
                    variant={selectedChannels.includes(channelName) ? "default" : "outline"}
                    className="cursor-pointer"
                    onClick={() => toggleChannelFilter(channelName)}
                  >
                    {channelName} ({Object.keys(modelsData[channelName].models).length})
                  </Badge>
                ))}
              </div>
            </div>
            <div className="md:col-span-1">
              <Button variant="outline" onClick={clearFilters} className="w-full">
                Clear Filters
              </Button>
            </div>
          </div>

          {totalModels === 0 ? (
            <div className="text-center py-8">
              <h3 className="text-lg font-medium mb-2">No models found</h3>
              <p className="text-muted-foreground">Try adjusting your search terms or filters.</p>
            </div>
          ) : (
            <>
              <div className="mb-6">
                <h3 className="text-lg font-medium">
                  Found {totalModels} models in {Object.keys(filteredData).length} channels
                </h3>
              </div>
              {Object.keys(filteredData)
                .sort()
                .map(channelName => renderChannelModels(channelName, filteredData[channelName]))}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default ModelsPage
