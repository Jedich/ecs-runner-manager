import React from 'react';
import { Nav, Button } from 'react-bootstrap';
import '../styles/Sidebar.css';

const Sidebar = ({ controller }) => {
  return (
    <Nav className={`col-md-12 d-none d-md-block sidebar`}
      activeKey="/home"
      onSelect={selectedKey => alert(`selected ${selectedKey}`)}
    >
      <div className="sidebar-inside">
      <h3>{controller.name}</h3>
      <div>aaaaaaa</div>

      <div className="sidebar-sticky"></div>
      <Nav.Item>
        <Nav.Link href="/home">Active</Nav.Link>
      </Nav.Item>
      </div>
    </Nav>
  );
};

export default Sidebar;