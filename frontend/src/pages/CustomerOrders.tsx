import React from 'react';
import { Link } from 'react-router-dom';

const CustomerOrders: React.FC = () => {
  return (
    <div className="page-container">
      <div className="page-header">
        <Link to="/" className="back-link">← Back to Dashboard</Link>
        <h1>Customer Orders</h1>
      </div>
      <div className="page-content">
        <p>Customer orders view will be implemented here</p>
      </div>
    </div>
  );
};

export default CustomerOrders;
