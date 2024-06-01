import React, { useState } from 'react';
import Header from './Header';
import Sidebar from './Sidebar';
import RunnerController from './RunnerController';
import Plot from './Plot';
import Chart from 'chart.js/auto';
import '../styles/App.css';
import Plotbar from './Plotbar';

const data = {
  "username": "test",
  "runner_controllers": [
    {
      "name": "runner controller",
      "runners": [
        {
          "name": "runner1",
          "private_ipv4": "172.31.36.190",
          "status": "busy",
        },
        {
          "name": "runner2",
          "private_ipv4": "172.31.43.32",
          "status": "finished",
        }
      ]
    },
    {
      "name": "runner controller2",
      "runners": []
    }
  ]
};

function App() {
  const [highlightedController, setHighlightedController] = useState(null);
  const [plotController, setPlotController] = useState(null);

  const handleOptionsClick = (controller) => {
    if (plotController && controller !== highlightedController) {
      setPlotController(null);
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
  }

  return (
    <div className="container-fluid">
      <Header username={data.username} />
      <div className="row">
        <div className={`sidebar-container ${highlightedController ? 'visible' : 'hidden'}`}>
          {highlightedController && <Sidebar controller={highlightedController} />}
        </div>
        <div className="col">
          {data.runner_controllers.map(controller => (
            <RunnerController
              key={controller.name}
              controller={controller}
              onOptionsClick={handleOptionsClick}
              onPlotClick={handlePlotClick}
              isHighlighted={highlightedController === controller || plotController === controller}
            />
          ))}
        </div>
        <div className={`plotbar-container ${plotController ? 'visible' : 'hidden'}`}>
          {plotController && <Plotbar controller={plotController} />}
        </div>
      </div>
    </div>
  );
}

export default App;
