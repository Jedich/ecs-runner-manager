import React, { useState } from 'react';
import { Card, Button } from 'react-bootstrap';
import RunnerCard from './RunnerCard';
import './RunnerController.css';
import { CgMoreVerticalO, CgChevronDownR, CgChevronRightR, CgPoll } from "react-icons/cg";

const RunnerController = ({ controller, onOptionsClick, isHighlighted }) => {
  const [expanded, setExpanded] = useState(false);

  const handleExpand = () => {
    setExpanded(!expanded);
  };

  return (
    <Card className={`runner-controller ${expanded ? 'expanded' : ''} ${isHighlighted ? 'highlighted' : ''}`}>
      <Card.Header className="d-flex justify-content-between align-items-center">
      <Button variant="link" className="options-button" onClick={() => onOptionsClick(controller)}>
      <CgMoreVerticalO />
          </Button>
        <span>{controller.name}</span>
        <div>
          <Button variant="link" className="expand-button" onClick={handleExpand}>
            {expanded ? <CgChevronDownR /> : <CgChevronRightR />}<CgPoll />
          </Button>
          
        </div>
      </Card.Header>
      <div className={`runner-cards-container ${expanded ? 'visible' : 'hidden'}`}>
        {controller.runners.map(runner => (
          <RunnerCard key={runner.name} runner={runner} />
          
        ))}
        
      </div>
    </Card>
  );
};

export default RunnerController;
