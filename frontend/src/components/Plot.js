import React from 'react';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  TimeScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from "chart.js";
import 'chartjs-adapter-date-fns';
import 'chartjs-plugin-zoom';

ChartJS.register(
  TimeScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const Plot = ({ name, datasets }) => {

  // Sample data
  const data = {
    datasets: datasets
  };

  const options = {
    maintainAspectRatio: false,
    plugins: {
      zoom: {
        zoom: {
          wheel: {
            enabled: true,
          },
          pinch: {
            enabled: true
          },
          mode: 'xy',
        }
      }
    },
    scales: {
      x: {
        type: 'time',
        time: {
          unit: 'second'
        }
      },
      y: {
        beginAtZero: true
      }
    },
  };

  return (
    <div>
      <h3>{name}</h3>
      <div style={{ height: "300px", position: "relative", marginBottom: "1%", padding: "1%" }}>
        <Line data={data} options={options} />
      </div>
    </div>
  );
};

export default Plot;
