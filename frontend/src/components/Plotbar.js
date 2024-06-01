import React from 'react';
import { Nav, Button } from 'react-bootstrap';
import '../styles/Plotbar.css';
import Plot from './Plot';

const Plotbar = ({ controller }) => {
    return (
        <div className='sidebar-inside'>
            <Plot />
        </div>
    );
};

export default Plotbar;
