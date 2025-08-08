import { useAuthStore } from '@/lib/stores/auth'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function DashboardPage() {
  const { user } = useAuthStore()

  if (!user) {
    return <div>Please log in to access the dashboard.</div>
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Dashboard</h1>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Account Info</CardTitle>
              <CardDescription>Your account details</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <p><strong>Username:</strong> {user.username}</p>
                <p><strong>Display Name:</strong> {user.display_name || 'Not set'}</p>
                <p><strong>Role:</strong> {user.role === 1 ? 'Admin' : 'User'}</p>
                <p><strong>Group:</strong> {user.group}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Quota Usage</CardTitle>
              <CardDescription>Your API usage statistics</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <p><strong>Total Quota:</strong> {user.quota?.toLocaleString() || 'N/A'}</p>
                <p><strong>Used:</strong> {user.used_quota?.toLocaleString() || 'N/A'}</p>
                <p><strong>Remaining:</strong> {user.quota && user.used_quota ? (user.quota - user.used_quota).toLocaleString() : 'N/A'}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Status</CardTitle>
              <CardDescription>Account status</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <p><strong>Status:</strong> {user.status === 1 ? 'Active' : 'Inactive'}</p>
                <p><strong>Email:</strong> {user.email || 'Not set'}</p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
