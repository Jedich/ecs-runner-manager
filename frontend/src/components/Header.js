import React from 'react';
import { Navbar, Nav, Container, Button } from 'react-bootstrap';
import '../styles/Header.css';

const Header = ({ username }) => {
  return (
    <Navbar variant="dark" bg="vvvvvv" className="bg-body-tertiary navbar">
      <Container>
        <Navbar.Brand href="#home">runner-manager</Navbar.Brand>
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
