import React, { createContext, useEffect, useState, useContext } from 'react';
import useWebSocket from 'react-use-websocket';

const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const [ctrlsData, setCtrlsData] = useState({ controllers: [], empty: false });
  const [metricsData, setMetricsData] = useState({ runners: [] });
  const token = "";
  
  const { lastJsonMessage, sendMessage } = useWebSocket(
    `ws://localhost:9090/api/runners/ws?token=${token}`,
    {
      share: false,
      shouldReconnect: () => true,
    },
  );

  useEffect(() => {
    if (lastJsonMessage) {
      const { event, data } = lastJsonMessage;
      if (event === 'ctrls') {
        setCtrlsData({ controllers: data, empty: data.length === 0 });
      } else if (event === 'metrics') {
        setMetricsData(data);
      }
    }
  }, [lastJsonMessage]);

  return (
    <WebSocketContext.Provider value={{ ctrlsData, metricsData, sendMessage }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWS = () => {
  return useContext(WebSocketContext);
};