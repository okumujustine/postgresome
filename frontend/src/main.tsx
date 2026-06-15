import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';
import { DatabaseInstanceProvider } from './lib/DatabaseInstanceProvider';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <DatabaseInstanceProvider>
        <App />
      </DatabaseInstanceProvider>
    </BrowserRouter>
  </React.StrictMode>,
);

