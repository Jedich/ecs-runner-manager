import React, { useState, useEffect } from 'react';
import Header from './Header';
import Sidebar from './Sidebar';
import RunnerController from './RunnerController';
import Plot from './Plot';
import Chart from 'chart.js/auto';
import '../styles/App.css';
import Plotbar from './Plotbar';
import { useWS } from './Websocket';
import AuthModal from './Auth';

const user = {
  "username": "test",
}

function App() {
  const { ctrlsData, sendMessage } = useWS();
  const data = ctrlsData ? ctrlsData : { controllers: [] };
  const [highlightedController, setHighlightedController] = useState(null);
  const [plotController, setPlotController] = useState(null);

  useEffect(() => {
    sendMessage(JSON.stringify({ "event": "ctrls" }));
  }, [sendMessage]);

  const handleOptionsClick = (controller) => {
    if (plotController && controller !== highlightedController) {
      setPlotController(null);
      sendMessage(JSON.stringify({ "ctrl_id": "" }));
    }
    setHighlightedController(controller === highlightedController ? null : controller);
  };

  const handlePlotClick = (controller) => {
    if (!highlightedController) {
      setHighlightedController(controller);
    }
    if (controller !== highlightedController) {
      setHighlightedController(controller);
    }
    setPlotController(controller === plotController ? null : controller);
    sendMessage(JSON.stringify({ "ctrl_id": controller === plotController ? "" : controller.id, "event": "metrics" }));
  }
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      setIsLoggedIn(true);
    } else {
      setShowModal(true);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    setIsLoggedIn(false);
    setShowModal(true);
  };
  return (
    <div className="container-fluid">
      {isLoggedIn ?
        <AuthModal
          show={showModal}
          setIsLoggedIn={setIsLoggedIn}
          setShowModal={setShowModal}
        />
        : null}
      <Header username={user.username} />
      <div className="row">
        <div className={`sidebar-container ${highlightedController ? 'visible' : 'hidden'}`}>
          {highlightedController && <Sidebar controller={highlightedController} />}
        </div>
        <div className="col">
          {data.controllers.length > 0 ? (
            data.controllers.map(controller => (
              <RunnerController
                key={controller.name}
                controller={controller}
                onOptionsClick={handleOptionsClick}
                onPlotClick={handlePlotClick}
                isHighlighted={highlightedController === controller || plotController === controller}
              />
            ))
          ) : (
            <div>Loading...</div>
          )}
        </div>
        <div className={`plotbar-container ${plotController ? 'visible' : 'hidden'}`}>
          {plotController && <Plotbar controller={plotController} />}
        </div>
      </div>
    </div>
  );
}

export default App;
