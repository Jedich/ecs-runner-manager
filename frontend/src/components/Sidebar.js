import React from 'react';
import { Card } from 'react-bootstrap';
import './Sidebar.css';

const Sidebar = ({ controller }) => {
  return (
    <Card className="sidebar">
      <Card.Body>
        <Card.Title>Controller: {controller.name}</Card.Title>
        <Card.Text>
          Number of Runners: {controller.runners.length}
        </Card.Text>
      </Card.Body>
    </Card>
  );
};

export default Sidebar;
