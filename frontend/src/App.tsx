import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Provider } from 'react-redux'
import { store } from './app/store'
import { Layout } from './components/Layout'
import { QAManagementPage } from './features/qa/pages/QAManagementPage'
import { ChatPage } from './features/chat/pages/ChatPage'
import { Toaster } from './components/ui/sonner'

function App() {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Navigate to="/qa" replace />} />
            <Route path="qa" element={<QAManagementPage />} />
            <Route path="chat" element={<ChatPage />} />
          </Route>
        </Routes>
        <Toaster />
      </BrowserRouter>
    </Provider>
  )
}

export default App
