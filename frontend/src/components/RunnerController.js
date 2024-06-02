import React, { useState } from 'react';
import { Card, Button } from 'react-bootstrap';
import RunnerCard from './RunnerCard';
import '../styles/RunnerController.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faChartLine, faList } from '@fortawesome/free-solid-svg-icons'
import { faSquareCaretDown, faSquareCaretRight } from '@fortawesome/free-regular-svg-icons'


const RunnerController = ({ controller, onOptionsClick, onPlotClick, isHighlighted }) => {
  let [expanded, setExpanded] = useState(null);

  const handleExpand = () => {
    setExpanded(!expanded);
  };


  return (
    <Card className={`runner-controller ${expanded ? 'expanded' : ''} ${isHighlighted ? 'highlighted' : ''}`}>
      <Card.Header className="d-flex justify-content-between align-items-center">
        
        <div className='a'>{controller['name']}</div>
        <div>
          <Button variant="link" className={`bttn ${controller.runners.length === 0 ? 'disabled' : ''}`} onClick={handleExpand}>
            {expanded ? <FontAwesomeIcon icon={faSquareCaretDown} size="xl" /> : <FontAwesomeIcon icon={faSquareCaretRight} size="xl" />}
          </Button>
          <Button variant="link" className="bttn" onClick={() => onPlotClick(controller)}>
          <FontAwesomeIcon icon={faChartLine} size="xl" />
          </Button>
          <Button variant="link" className="bttn" onClick={() => onOptionsClick(controller)}>
          <FontAwesomeIcon icon={faList} size="xl" />
          </Button>
        </div>
      </Card.Header>
      <div className={`runner-cards-container visible`}>
        {controller.runners.map(runner => (
          <RunnerCard key={runner.name} runner={runner} />
        ))}

      </div>
    </Card>
  );
};

export default RunnerController;
