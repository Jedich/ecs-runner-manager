import React from 'react';
import { Line } from 'react-chartjs-2';

const Plot = () => {
    
  // Sample data
  const data = {
    scales: {
        xAxes: [{
            title: "time",
            type: 'time',
            gridLines: {
                lineWidth: 2
            },
            time: {
                unit: "day",
                unitStepSize: 1000,
                displayFormats: {
                    millisecond: 'MMM DD',
                    second: 'MMM DD',
                    minute: 'MMM DD',
                    hour: 'MMM DD',
                    day: 'MMM DD',
                    week: 'MMM DD',
                    month: 'MMM DD',
                    quarter: 'MMM DD',
                    year: 'MMM DD',
                }
            }
        }]
    },
    labels: ["2024-05-31T20:23:19.000+00:00", "2024-06-01T00:00:00.000+00:00", "2024-06-02T00:00:00.000+00:00", "2024-06-03T00:00:00.000+00:00", "2024-06-04T00:00:00.000+00:00", "2024-06-05T00:00:00.000+00:00", "2024-06-06T00:00:00.000+00:00", "2024-06-07T00:00:00.000+00:00", "2024-06-08T00:00:00.000+00:00", "2024-06-09T00:00:00.000+00:00", "2024-06-10T00:00:00.000+00:00"],
    datasets: [
      {
        label: 'Example Dataset',
        data: [1.064960, 1.064960, 34.344960, 34.344960, 91.774976, 123.383808, 117.686272,
            117.686272, 117.559296, 117.559296, 117.559296
        ],
        fill: true,
        backgroundColor: 'rgba(75,192,192,0.2)',
        borderColor: 'rgba(75,192,192,1)',
        tension: 0.5
      },
      {
        label: 'Example Dataset 2',
        data: [536.870912, 536.870912, 536.870912, 536.870912, 536.870912, 536.870912, 536.870912,
            536.870912, 536.870912, 536.870912, 536.870912
        ],
        fill: false,
        borderColor: 'rgba(255,60,60,1)',
        tension: 0.5
      }
    ]
  };

  return (
    <div>
      <h2>Plot Example</h2>
      <Line data={data} />
    </div>
  );
};

export default Plot;
