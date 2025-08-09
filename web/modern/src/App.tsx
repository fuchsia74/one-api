import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Layout } from '@/components/layout/Layout'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'
import { PasswordResetPage } from '@/pages/auth/PasswordResetPage'
import { PasswordResetConfirmPage } from '@/pages/auth/PasswordResetConfirmPage'
import { GitHubOAuthPage } from '@/pages/auth/GitHubOAuthPage'
import { LarkOAuthPage } from '@/pages/auth/LarkOAuthPage'
import { DashboardPage } from '@/pages/dashboard/DashboardPage'
import { TokensPage } from '@/pages/tokens/TokensPage'
import { EditTokenPage } from '@/pages/tokens/EditTokenPage'
import { LogsPage } from '@/pages/logs/LogsPage'
import { UsersPage } from '@/pages/users/UsersPage'
import { EditUserPage } from '@/pages/users/EditUserPage'
import { ChannelsPage } from '@/pages/channels/ChannelsPage'
import { EditChannelPage } from '@/pages/channels/EditChannelPage'
import { RedemptionsPage } from '@/pages/redemptions/RedemptionsPage'
import { EditRedemptionPage } from '@/pages/redemptions/EditRedemptionPage'
import { AboutPage } from '@/pages/about/AboutPage'
import { SettingsPage } from '@/pages/settings/SettingsPage'
import { ModelsPage } from '@/pages/models/ModelsPage'
import { TopUpPage } from '@/pages/topup/TopUpPage'
import { ChatPage } from '@/pages/chat/ChatPage'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="min-h-screen bg-background">
          <Routes>
            {/* Public auth routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/reset" element={<PasswordResetPage />} />
            <Route path="/user/reset" element={<PasswordResetConfirmPage />} />
            <Route path="/oauth/github" element={<GitHubOAuthPage />} />
            <Route path="/oauth/lark" element={<LarkOAuthPage />} />

            {/* Protected routes */}
            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                <Route index element={<HomePage />} />
                <Route path="dashboard" element={<DashboardPage />} />
                <Route path="tokens" element={<TokensPage />} />
                <Route path="tokens/add" element={<EditTokenPage />} />
                <Route path="tokens/edit/:id" element={<EditTokenPage />} />
                <Route path="logs" element={<LogsPage />} />
                <Route path="users" element={<UsersPage />} />
                <Route path="users/add" element={<EditUserPage />} />
                <Route path="users/edit/:id" element={<EditUserPage />} />
                <Route path="users/edit" element={<EditUserPage />} />
                <Route path="channels" element={<ChannelsPage />} />
                <Route path="channels/add" element={<EditChannelPage />} />
                <Route path="channels/edit/:id" element={<EditChannelPage />} />
                <Route path="redemptions" element={<RedemptionsPage />} />
                <Route path="redemptions/add" element={<EditRedemptionPage />} />
                <Route path="redemptions/edit/:id" element={<EditRedemptionPage />} />
                <Route path="about" element={<AboutPage />} />
                <Route path="settings" element={<SettingsPage />} />
                <Route path="models" element={<ModelsPage />} />
                <Route path="topup" element={<TopUpPage />} />
                <Route path="chat" element={<ChatPage />} />
              </Route>
            </Route>
          </Routes>
        </div>
      </Router>
    </QueryClientProvider>
  )
}

export default App
