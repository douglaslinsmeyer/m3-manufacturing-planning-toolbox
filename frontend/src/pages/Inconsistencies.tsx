import React from 'react';
import { Link } from 'react-router-dom';

const Inconsistencies: React.FC = () => {
  return (
    <div className="page-container">
      <div className="page-header">
        <Link to="/" className="back-link">← Back to Dashboard</Link>
        <h1>Planning Inconsistencies</h1>
      </div>
      <div className="page-content">
        <p>Inconsistencies analysis will be implemented here</p>
      </div>
    </div>
  );
};

export default Inconsistencies;
