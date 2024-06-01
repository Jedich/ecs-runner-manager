import React from 'react';
import { Card } from 'react-bootstrap';
import '../styles/RunnerCard.css';


const RunnerCard = ({ runner }) => {
  const getStatusColor = (status) => {
    let statuses = {
      'idle': 'gray',
      'busy': 'orange',
      'finished': 'green',
      'error': 'red'
    }
    return statuses[status];
  };

  return (
    <Card className="m-2" style={{ width: '200px' }}>
      <Card.Body>
        <Card.Title>{runner.name}</Card.Title>
        <Card.Text>Private IPv4: {runner.private_ipv4}</Card.Text>
        <Card.Text>
          <span className="status-circle" style={{ backgroundColor: getStatusColor(runner.status) }}></span>
        </Card.Text>
      </Card.Body>
    </Card>
  );
};

export default RunnerCard;
