import React from 'react';
import { Link, useParams } from 'react-router-dom';

const PlannedOrderDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  return (
    <div className="page-container">
      <div className="page-header">
        <Link to="/production-orders" className="back-link">← Back to Production Orders</Link>
        <h1>Planned Order Details</h1>
      </div>
      <div className="page-content">
        <p>MOP details for ID: {id}</p>
      </div>
    </div>
  );
};

export default PlannedOrderDetail;
