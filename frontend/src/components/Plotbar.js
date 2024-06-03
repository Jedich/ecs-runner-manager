import React, { useState } from 'react';
import { Nav, Button, Dropdown } from 'react-bootstrap';
import '../styles/Plotbar.css';
import Plot from './Plot';
import { useWS } from './Websocket';

const Plotbar = ({ controller }) => {
    const [selectedItem, setSelectedItem] = useState('Select an item');
    const currentTime = new Date();
    let before = 0;
    const timeMap = [
        {
            'value': new Date(currentTime.getTime() - 5 * 60 * 1000),
            'desc': '5 minutes ago',
        },
        {
            'value': new Date(currentTime.getTime() - 15 * 60 * 1000),
            'desc': '15 minutes ago',
        },
        {
            'value': new Date(currentTime.getTime() - 30 * 60 * 1000),
            'desc': '30 minutes ago',
        },
        {
            'value': new Date(currentTime.getTime() - 45 * 60 * 1000),
            'desc': '45 minutes ago',
        },
        {
            'value': new Date(currentTime.getTime() - 1 * 60 * 60 * 1000),
            'desc': '1 hour ago',
        },
        {
            'value': new Date(currentTime.getTime() - 3 * 60 * 60 * 1000),
            'desc': '3 hours ago',
        },
        {
            'value': new Date(currentTime.getTime() - 6 * 60 * 60 * 1000),
            'desc': '6 hours ago',
        }
    ];
    const handleSelect = (eventKey) => {
        console.log(`selected ${eventKey}`);
        setSelectedItem(timeMap[eventKey].value);
    };
    const niceColors = [
        'rgba(255, 99, 132, 1)',
        'rgba(54, 162, 235, 1)',
        'rgba(255, 206, 86, 1)',
        'rgba(75, 192, 192, 1)',
        'rgba(153, 102, 255, 1)',
        'rgba(255, 159, 64, 1)',
        'rgba(255, 99, 132, 1)',
        'rgba(54, 162, 235, 1)',
        'rgba(255, 206, 86, 1)',
        'rgba(75, 192, 192, 1)'];
    const { metricsData } = useWS();
    const data = metricsData ? metricsData : { runners: [{ metrics: [] }] };
    const newOpacity = '0.2';

    const hoursBefore = selectedItem ? selectedItem : new Date(currentTime.getTime() - 1 * 60 * 60 * 1000);
    const addAlpha = (color, opacity) => {
        // coerce values so it is between 0 and 1.
        var _opacity = Math.round(Math.min(Math.max(opacity ?? 1, 0), 1) * 255);
        return color + _opacity.toString(16).toUpperCase();
    };

    let mem = [];
    let cpu = [];
    data.runners.sort((a, b) => a.name.localeCompare(b.name))
    for (let i = 0; i < data.runners.length; i++) {
        let runner = data.runners[i]
        if (controller.runners[i].name === null) {
            runner.name = "runner"
        } else {
            runner.name = controller.runners[i].name;
        }
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
        let color = niceColors[i % niceColors.length];
        let bgColor = color.replace(/[^,]+(?=\))/, newOpacity);
        let memdata = {
            label: runner.name,
            data: [],
            fill: true,
            backgroundColor: bgColor,
            borderColor: color,
            tension: 0.5
        }
        let cpudata = {
            label: runner.name,
            data: [],
            fill: true,
            backgroundColor: bgColor,
            borderColor: color,
            tension: 0.5
        }
        for (let j = 0; j < data.runners[i].metrics.length; j++) {
            let metric = data.runners[i].metrics[j];
            let time = new Date(metric.timestamp);
            if (metric.metadata.ecs_memory_bytes) {
                if (hoursBefore != null && time < hoursBefore) {
                    continue;
                }
                memdata.data.push({ x: time, y: metric.metadata.ecs_memory_bytes / 1000000 });
            }
            if (metric.metadata.ecs_cpu_seconds_total) {
                if (hoursBefore != null && time < hoursBefore) {
                    continue;
                }
                cpudata.data.push({ x: time, y: metric.metadata.ecs_cpu_seconds_total / 100 });
            }
            // if (metric['ecs_memory_bytes']) {
            //     memdata.data.push(metric['ecs_memory_bytes']);
            // }
        }
        mem.push(memdata);
        cpu.push(cpudata);
    }

    return (
        <div className='sidebar-inside'>
            {mem.length > 0 ?
                <div>

                    <Dropdown onSelect={handleSelect}>
                        <Dropdown.Toggle variant="success" id="dropdown-basic">
                            Dropdown Button
                        </Dropdown.Toggle>

                        <Dropdown.Menu>
                            {timeMap.map((time, index) => (
                                <Dropdown.Item eventKey={index}>{time.desc}</Dropdown.Item>
                            ))}
                        </Dropdown.Menu>
                    </Dropdown>
                    <Plot name={"Memory, MB."} datasets={mem} />
                    <Plot name={"CPU, %"} datasets={cpu} />
                </div>
                :
                <div>Loading...</div>
            }

        </div>
    );
};

export default Plotbar;
