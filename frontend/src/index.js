import React from 'react';
import ReactDOM from 'react-dom/client';
import 'bootstrap/dist/css/bootstrap.min.css';
import "@fontsource/ubuntu/400.css";
import './index.css';
import App from './components/App';
import { WebSocketProvider, useWebSocket } from './components/Websocket';


const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <WebSocketProvider>
    <App />
    </WebSocketProvider>
  </React.StrictMode>
);
