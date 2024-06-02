// WebSocketProvider.js
import React, { createContext, useEffect, useState, useContext } from 'react';
import useWebSocket, { ReadyState } from 'react-use-websocket';

const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const [data, setData] = useState({controllers: []});
  const token = ""
  const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket(
      `ws://localhost:9090/api/runners/ws?token=${token}`,
      {
        share: false,
        shouldReconnect: () => true,
      },
    )
    useEffect(() => {
        console.log(`Got a new message: ${JSON.stringify(lastJsonMessage)}`)
        setData(lastJsonMessage);
      }, [lastJsonMessage])    

  return (
    <WebSocketContext.Provider value={{ data }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWS = () => {
  return useContext(WebSocketContext);
};