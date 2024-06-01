import React, { useState } from 'react';
import Header from './components/Header';
import Sidebar from './components/Sidebar';
import RunnerController from './components/RunnerController';
import Plot from './components/Plot';
import Chart from 'chart.js/auto';
import './App.css';

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

  const handleOptionsClick = (controller) => {
    setHighlightedController(controller === highlightedController ? null : controller);
  };

  return (
    <div className="container-fluid">
      <Plot />
      <Header username={data.username} />
      
      <div className="row">
      
        <div className="col">
          {data.runner_controllers.map(controller => (
            <RunnerController
              key={controller.name}
              controller={controller}
              onOptionsClick={handleOptionsClick}
              isHighlighted={highlightedController === controller}
            />
          ))}
        </div>
        
        <div className={`col-3 sidebar-container ${highlightedController ? 'visible' : 'hidden'}`}>
          {highlightedController && <Sidebar controller={highlightedController} />}
        </div>
        
      </div>
    </div>
  );
}

export default App;
