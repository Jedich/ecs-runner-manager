import React, {useEffect} from 'react';
import { Navbar, Nav, Container, Button } from 'react-bootstrap';
import '../styles/Header.css';
import { useWS } from './Websocket';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faRotate } from '@fortawesome/free-solid-svg-icons'

const Header = ({ username }) => {
  const { sendMessage } = useWS();
  return (
    <Navbar variant="dark" bg="vvvvvv" className="bg-body-tertiary navbar">
      <Container>
        <Navbar.Brand href="#home">runner-manager</Navbar.Brand>
        <Button className='logout' variant="outline-info" onClick={
          useEffect(() => {
            sendMessage(JSON.stringify({ "event": "ctrls" }));
          }, [sendMessage])
        }><FontAwesomeIcon icon={faRotate} /></Button>
        <Navbar.Toggle />
        <Navbar.Collapse className="justify-content-end">
          <Navbar.Text>
            Signed in as: <a>Mark Otto</a>
          </Navbar.Text>
          <Button className='logout' variant="outline-info">Logout</Button>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
};

export default Header;
