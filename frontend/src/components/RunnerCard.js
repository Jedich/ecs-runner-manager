import React from 'react';
import { Card } from 'react-bootstrap';
import './RunnerCard.css';

const RunnerCard = ({ runner }) => {
  const getStatusColor = (status) => {
    return status === 'busy' ? 'orange' : 'green';
  };

  return (
    <Card className="m-2" style={{ width: '18rem' }}>
      <Card.Body>
        <Card.Title>{runner.name}</Card.Title>
        <Card.Text>Private IPv4: {runner.private_ipv4}</Card.Text>
        <Card.Text>
          Status: <span className="status-circle" style={{ backgroundColor: getStatusColor(runner.status) }}></span>
        </Card.Text>
      </Card.Body>
    </Card>
  );
};

export default RunnerCard;
