import React from 'react';
import { Link } from 'react-router-dom';

const Deliveries: React.FC = () => {
  return (
    <div className="page-container">
      <div className="page-header">
        <Link to="/" className="back-link">← Back to Dashboard</Link>
        <h1>Deliveries</h1>
      </div>
      <div className="page-content">
        <p>Deliveries view will be implemented here</p>
      </div>
    </div>
  );
};

export default Deliveries;
