import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './Login.css';

const Login: React.FC = () => {
  const { login } = useAuth();
  const [selectedEnv, setSelectedEnv] = useState<'TRN' | 'PRD'>('TRN');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    try {
      await login(selectedEnv);
    } catch (err) {
      setError('Failed to initiate login. Please try again.');
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1 className="login-title">M3 Manufacturing Planning Tools</h1>
        <p className="login-subtitle">Select your environment to continue</p>

        <div className="environment-selector">
          <button
            className={`env-button ${selectedEnv === 'TRN' ? 'selected' : ''}`}
            onClick={() => setSelectedEnv('TRN')}
          >
            <div className="env-label">Training</div>
            <div className="env-code">TRN</div>
          </button>
          <button
            className={`env-button ${selectedEnv === 'PRD' ? 'selected' : ''}`}
            onClick={() => setSelectedEnv('PRD')}
          >
            <div className="env-label">Production</div>
            <div className="env-code">PRD</div>
          </button>
        </div>

        {error && <div className="login-error">{error}</div>}

        <button
          className="login-button"
          onClick={handleLogin}
          disabled={loading}
        >
          {loading ? 'Redirecting...' : 'Sign In with M3'}
        </button>

        <p className="login-info">
          You will be redirected to Infor M3 to sign in with your credentials
        </p>
      </div>
    </div>
  );
};

export default Login;
