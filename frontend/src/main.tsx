import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import App from './App.tsx'
import { DatabaseInstanceProvider } from './lib/DatabaseInstanceProvider.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <DatabaseInstanceProvider>
        <App />
      </DatabaseInstanceProvider>
    </BrowserRouter>
  </StrictMode>,
)
