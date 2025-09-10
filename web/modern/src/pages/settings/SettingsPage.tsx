import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ResponsivePageContainer } from '@/components/ui/responsive-container'
import { useResponsive } from '@/hooks/useResponsive'
import { PersonalSettings } from './PersonalSettings'
import { SystemSettings } from './SystemSettings'
import { OperationSettings } from './OperationSettings'
import { OtherSettings } from './OtherSettings'
import { useAuthStore } from '@/lib/stores/auth'
import { cn } from '@/lib/utils'

export function SettingsPage() {
  const { user } = useAuthStore()
  const { isMobile } = useResponsive()
  const isRoot = user?.role >= 100

  const tabCount = 1 + (isRoot ? 3 : 0) // Personal + 3 admin tabs

  return (
    <ResponsivePageContainer
      title="Settings"
      description="Configure your account and system settings"
    >
      <Card>
        <CardContent className={cn(
          isMobile ? "p-4" : "p-6"
        )}>
          <Tabs defaultValue="personal" className="w-full">
            <TabsList className={cn(
              "grid w-full",
              isMobile ? "grid-cols-1 h-auto flex-col" :
                tabCount === 1 ? "grid-cols-1" :
                  tabCount === 2 ? "grid-cols-2" :
                    tabCount === 3 ? "grid-cols-3" :
                      "grid-cols-2 lg:grid-cols-4"
            )}>
              <TabsTrigger
                value="personal"
                className={cn(isMobile ? "w-full justify-start" : "")}
              >
                Personal
              </TabsTrigger>
              {isRoot && (
                <TabsTrigger
                  value="operation"
                  className={cn(isMobile ? "w-full justify-start" : "")}
                >
                  Operation
                </TabsTrigger>
              )}
              {isRoot && (
                <TabsTrigger
                  value="system"
                  className={cn(isMobile ? "w-full justify-start" : "")}
                >
                  System
                </TabsTrigger>
              )}
              {isRoot && (
                <TabsTrigger
                  value="other"
                  className={cn(isMobile ? "w-full justify-start" : "")}
                >
                  Other
                </TabsTrigger>
              )}
            </TabsList>

            <TabsContent value="personal" className={cn(isMobile ? "mt-4" : "mt-6")}>
              <PersonalSettings />
            </TabsContent>

            {isRoot && (
              <TabsContent value="operation" className={cn(isMobile ? "mt-4" : "mt-6")}>
                <OperationSettings />
              </TabsContent>
            )}

            {isRoot && (
              <TabsContent value="system" className={cn(isMobile ? "mt-4" : "mt-6")}>
                <SystemSettings />
              </TabsContent>
            )}

            {isRoot && (
              <TabsContent value="other" className={cn(isMobile ? "mt-4" : "mt-6")}>
                <OtherSettings />
              </TabsContent>
            )}
          </Tabs>
        </CardContent>
      </Card>
    </ResponsivePageContainer>
  )
}

export default SettingsPage
