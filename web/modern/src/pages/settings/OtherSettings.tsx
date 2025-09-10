import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Separator } from '@/components/ui/separator'
import { api } from '@/lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Info } from 'lucide-react'

const otherSchema = z.object({
  Footer: z.string().default(''),
  Notice: z.string().default(''),
  About: z.string().default(''),
  SystemName: z.string().default(''),
  Logo: z.string().default(''),
  HomePageContent: z.string().default(''),
  Theme: z.string().default(''),
})

type OtherForm = z.infer<typeof otherSchema>

export function OtherSettings() {
  const [loading, setLoading] = useState(true)
  const [updateData, setUpdateData] = useState<{ tag_name: string; content: string } | null>(null)

  // Descriptions for each setting on this page
  const descriptions = useMemo<Record<string, string>>(
    () => ({
      SystemName: 'System display name shown in the UI and emails.',
      Logo: 'Logo image URL displayed in the header/login screens.',
      Theme: 'UI theme name. Supported: default, berry, air, modern.',
      Notice: 'Site‑wide announcement content shown to all users. Markdown supported.',
      About: 'Content for the About page. Markdown supported.',
      HomePageContent: 'Content displayed on the home page. Markdown supported.',
      Footer: 'Footer content rendered site‑wide. HTML allowed; use responsibly.',
    }),
    []
  )

  const form = useForm<OtherForm>({
    resolver: zodResolver(otherSchema),
    defaultValues: {
      Footer: '',
      Notice: '',
      About: '',
      SystemName: '',
      Logo: '',
      HomePageContent: '',
      Theme: '',
    },
  })

  const loadOptions = async () => {
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/option/')
      const { success, data } = res.data
      if (success && data) {
        const formData: any = {}
        data.forEach((item: { key: string; value: string }) => {
          const key = item.key
          if (key in form.getValues()) {
            formData[key] = item.value
          }
        })
        form.reset(formData)
      }
    } catch (error) {
      console.error('Error loading options:', error)
    } finally {
      setLoading(false)
    }
  }

  const updateOption = async (key: string, value: string) => {
    try {
      setLoading(true)
      // Unified API call - complete URL with /api prefix
      await api.put('/api/option/', { key, value })
      console.log(`Updated ${key}`)
    } catch (error) {
      console.error(`Error updating ${key}:`, error)
    } finally {
      setLoading(false)
    }
  }

  const submitField = async (key: keyof OtherForm) => {
    const value = form.getValues(key)
    await updateOption(key, value)
  }

  const checkUpdate = async () => {
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/option/update')
      const { success, data } = res.data
      if (success && data) {
        setUpdateData(data)
      }
    } catch (error) {
      console.error('Error checking for updates:', error)
    }
  }

  const openGitHubRelease = () => {
    window.open('https://github.com/Laisky/one-api/releases/latest', '_blank')
  }

  useEffect(() => {
    loadOptions()
  }, [])

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <span className="ml-3">Loading other settings...</span>
        </CardContent>
      </Card>
    )
  }

  return (
    <TooltipProvider>
      <div className="space-y-6">
        {/* System Branding */}
        <Card>
          <CardHeader>
            <CardTitle>System Branding</CardTitle>
            <CardDescription>Configure system name, logo, and visual appearance</CardDescription>
          </CardHeader>
          <CardContent>
            <Form {...form}>
              <div className="space-y-4">
                <FormField
                  control={form.control}
                  name="SystemName"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        System Name
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About System Name">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.SystemName}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="flex gap-2">
                        <FormControl>
                          <Input placeholder="One API" {...field} />
                        </FormControl>
                        <Button onClick={() => submitField('SystemName')}>Save</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="Logo"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        Logo URL
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Logo URL">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.Logo}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="flex gap-2">
                        <FormControl>
                          <Input placeholder="https://..." {...field} />
                        </FormControl>
                        <Button onClick={() => submitField('Logo')}>Save</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="Theme"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        Theme
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Theme">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.Theme}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="flex gap-2">
                        <FormControl>
                          <Input placeholder="default" {...field} />
                        </FormControl>
                        <Button onClick={() => submitField('Theme')}>Save</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </Form>
          </CardContent>
        </Card>

        {/* Content Management */}
        <Card>
          <CardHeader>
            <CardTitle>Content Management</CardTitle>
            <CardDescription>Configure notices, about page, and home page content</CardDescription>
          </CardHeader>
          <CardContent>
            <Form {...form}>
              <div className="space-y-4">
                <FormField
                  control={form.control}
                  name="Notice"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        Notice (supports Markdown)
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Notice">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.Notice}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="space-y-2">
                        <FormControl>
                          <Textarea
                            placeholder="Enter notice content..."
                            className="min-h-[100px]"
                            {...field}
                          />
                        </FormControl>
                        <Button onClick={() => submitField('Notice')}>Save Notice</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="About"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        About Page Content (supports Markdown)
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About About Page Content">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.About}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="space-y-2">
                        <FormControl>
                          <Textarea
                            placeholder="Enter about page content..."
                            className="min-h-[100px]"
                            {...field}
                          />
                        </FormControl>
                        <Button onClick={() => submitField('About')}>Save About</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="HomePageContent"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        Home Page Content (supports Markdown)
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Home Page Content">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.HomePageContent}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="space-y-2">
                        <FormControl>
                          <Textarea
                            placeholder="Enter home page content..."
                            className="min-h-[100px]"
                            {...field}
                          />
                        </FormControl>
                        <Button onClick={() => submitField('HomePageContent')}>Save Home Content</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="Footer"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="flex items-center gap-2">
                        Footer Content (supports HTML)
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Footer Content">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.Footer}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
                      <div className="space-y-2">
                        <FormControl>
                          <Textarea
                            placeholder="Enter footer content..."
                            className="min-h-[80px]"
                            {...field}
                          />
                        </FormControl>
                        <Button onClick={() => submitField('Footer')}>Save Footer</Button>
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </Form>
          </CardContent>
        </Card>

        {/* System Updates */}
        <Card>
          <CardHeader>
            <CardTitle>System Updates</CardTitle>
            <CardDescription>Check for updates and manage system version</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex gap-2">
                <Button onClick={checkUpdate}>Check for Updates</Button>
                <Button variant="outline" onClick={openGitHubRelease}>
                  View GitHub Releases
                </Button>
              </div>

              {updateData && (
                <div className="p-4 bg-muted rounded-lg">
                  <h4 className="font-medium mb-2">Update Available: {updateData.tag_name}</h4>
                  <div className="text-sm text-muted-foreground">
                    {updateData.content}
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </TooltipProvider>
  )
}

export default OtherSettings
