import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PersonalSettings } from './PersonalSettings'
import { SystemSettings } from './SystemSettings'
import { OperationSettings } from './OperationSettings'
import { OtherSettings } from './OtherSettings'
import { useAuthStore } from '@/lib/stores/auth'

export function SettingsPage() {
  const { user } = useAuthStore()
  const isRoot = user?.role >= 100

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <CardTitle>Settings</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="personal" className="w-full">
            <TabsList className="grid w-full grid-cols-2 lg:grid-cols-4">
              <TabsTrigger value="personal">Personal</TabsTrigger>
              {isRoot && <TabsTrigger value="operation">Operation</TabsTrigger>}
              {isRoot && <TabsTrigger value="system">System</TabsTrigger>}
              {isRoot && <TabsTrigger value="other">Other</TabsTrigger>}
            </TabsList>

            <TabsContent value="personal" className="mt-6">
              <PersonalSettings />
            </TabsContent>

            {isRoot && (
              <TabsContent value="operation" className="mt-6">
                <OperationSettings />
              </TabsContent>
            )}

            {isRoot && (
              <TabsContent value="system" className="mt-6">
                <SystemSettings />
              </TabsContent>
            )}

            {isRoot && (
              <TabsContent value="other" className="mt-6">
                <OtherSettings />
              </TabsContent>
            )}
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}

export default SettingsPage
