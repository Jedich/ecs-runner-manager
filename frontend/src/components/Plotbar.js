import React from 'react';
import { Nav, Button } from 'react-bootstrap';
import '../styles/Plotbar.css';
import Plot from './Plot';
import { useWS } from './Websocket';

const Plotbar = ({ controller }) => {
    const { metricsData } = useWS();
    const data = metricsData ? metricsData : { runners: [{ metrics: [] }] };
    console.log(JSON.stringify(data));
    const currentTime = new Date();
    const threeHoursBefore = new Date(currentTime.getTime() - 1 * 60 * 60 * 1000);
    const addAlpha = (color, opacity) => {
        // coerce values so it is between 0 and 1.
        var _opacity = Math.round(Math.min(Math.max(opacity ?? 1, 0), 1) * 255);
        return color + _opacity.toString(16).toUpperCase();
    };

    let mem = [];
    metricsData.runners.forEach(runner => {
        if (runner.name === null) {
            return;
        }
        // let memdata = {
        //     label: runner['name'],
        //     data: [],
        //     fill: true,
        //     backgroundColor: addAlpha(runner['color'], 0.2),
        //     borderColor: addAlpha(runner['color'], 1),
        //     tension: 0.5
        // }
        let memlimitdata = {
            label: runner.name + '_limit',
            data: [],
            fill: false,
            borderColor: 'rgba(255,60,60,1)',
            tension: 0.5
        }
        for (let i = 0; i < runner.metrics.length; i++) {
            let metric = runner.metrics[i];
            if (metric.metadata.ecs_memory_bytes) {
                let time = new Date(metric.timestamp);
                // if (time < threeHoursBefore) {
                //     continue;
                // }
                memlimitdata.data.push({x: time, y: metric.metadata.ecs_memory_bytes/1000000});
            }
            // if (metric['ecs_memory_bytes']) {
            //     memdata.data.push(metric['ecs_memory_bytes']);
            // }
          }
        mem.push(memlimitdata);
    });

    console.log(mem);

    return (
        <div className='sidebar-inside'>
            {mem.length > 0 ?
                <Plot name={"Memory"} datasets={mem} />
                :
                <div>Loading...</div>
            }

        </div>
    );
};

export default Plotbar;
