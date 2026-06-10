import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';

const Login: React.FC = () => {
  const { login } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    try {
      await login();
    } catch (err) {
      setError('Failed to initiate login. Please try again.');
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-900 flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      {/* Background pattern */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary-500/10 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-primary-600/10 rounded-full blur-3xl" />
      </div>

      <div className="relative mx-auto w-full max-w-md">
        {/* Logo */}
        <div className="flex justify-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary-500 shadow-lg">
            <span className="text-xl font-bold text-white">PW</span>
          </div>
        </div>
        <h2 className="mt-12 text-center text-3xl font-bold tracking-tight text-white">
          Production Planning Issues Workbench
        </h2>
        <p className="mt-4 text-center text-sm text-slate-400">
          Sign in with your PING account to access the workbench
        </p>
      </div>

      <div className="relative mt-12 mx-auto w-full max-w-md">
        <div className="bg-slate-800 border border-slate-700 py-12 px-8 shadow-xl rounded-xl sm:px-10">
          {/* Error Message */}
          {error && (
            <div className="mb-4 rounded-lg bg-error-500/10 border border-error-500/20 px-4 py-3">
              <p className="text-sm text-error-400">{error}</p>
            </div>
          )}

          {/* Sign In Button */}
          <button
            type="button"
            onClick={handleLogin}
            disabled={loading}
            className="w-full flex justify-center items-center gap-2 rounded-lg bg-primary-600 px-4 py-3 text-sm font-semibold text-white shadow-sm transition-all duration-200 hover:bg-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-slate-800 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? (
              <>
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Redirecting...
              </>
            ) : (
              <>
                {/* Microsoft logo */}
                <svg className="h-4 w-4" viewBox="0 0 21 21" aria-hidden="true">
                  <rect x="1" y="1" width="9" height="9" fill="#f25022" />
                  <rect x="11" y="1" width="9" height="9" fill="#7fba00" />
                  <rect x="1" y="11" width="9" height="9" fill="#00a4ef" />
                  <rect x="11" y="11" width="9" height="9" fill="#ffb900" />
                </svg>
                Sign in with Microsoft
              </>
            )}
          </button>

          {/* Help text */}
          <p className="mt-6 text-center text-xs text-slate-500">
            You will be redirected to Microsoft Entra ID for authentication
          </p>
        </div>

        {/* Footer */}
        <p className="mt-6 text-center text-xs text-slate-600">
          Need help? Contact your system administrator
        </p>
      </div>
    </div>
  );
};

export default Login;
