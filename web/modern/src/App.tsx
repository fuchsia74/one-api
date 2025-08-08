import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Layout } from '@/components/layout/Layout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/auth/LoginPage'
import { DashboardPage } from '@/pages/dashboard/DashboardPage'
import { TokensPage } from '@/pages/tokens/TokensPage'
import { LogsPage } from '@/pages/logs/LogsPage'
import { UsersPage } from '@/pages/users/UsersPage'
import { ChannelsPage } from '@/pages/channels/ChannelsPage'
import { RedemptionsPage } from '@/pages/redemptions/RedemptionsPage'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="min-h-screen bg-background">
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route element={<ProtectedRoute />}>
            <Route path="/" element={<Layout />}>
              <Route index element={<HomePage />} />
              <Route path="dashboard" element={<DashboardPage />} />
              <Route path="tokens" element={<TokensPage />} />
              <Route path="logs" element={<LogsPage />} />
              <Route path="users" element={<UsersPage />} />
              <Route path="channels" element={<ChannelsPage />} />
              <Route path="redemptions" element={<RedemptionsPage />} />
            </Route>
            </Route>
          </Routes>
        </div>
      </Router>
    </QueryClientProvider>
  )
}

export default App
