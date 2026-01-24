import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';

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
            <span className="text-xl font-bold text-white">M3</span>
          </div>
        </div>
        <h2 className="mt-12 text-center text-3xl font-bold tracking-tight text-white">
          Manufacturing Planning Tools
        </h2>
        <p className="mt-4 text-center text-sm text-slate-400">
          Sign in to access your planning dashboard
        </p>
      </div>

      <div className="relative mt-12 mx-auto w-full max-w-md">
        <div className="bg-slate-800 border border-slate-700 py-12 px-8 shadow-xl rounded-xl sm:px-10">
          {/* Environment Selection */}
          <div className="mb-8">
            <label className="block text-sm font-medium text-slate-300 mb-4">
              Select Environment
            </label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setSelectedEnv('TRN')}
                className={`relative flex flex-col items-center justify-center rounded-lg border-2 py-4 px-4 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 focus:ring-offset-slate-800 ${
                  selectedEnv === 'TRN'
                    ? 'border-primary-500 bg-primary-500/10'
                    : 'border-slate-600 bg-slate-700/50 hover:border-slate-500 hover:bg-slate-700'
                }`}
              >
                <span className={`text-2xl font-bold ${selectedEnv === 'TRN' ? 'text-primary-400' : 'text-slate-300'}`}>
                  TRN
                </span>
                <span className={`mt-1 text-xs ${selectedEnv === 'TRN' ? 'text-primary-400' : 'text-slate-500'}`}>
                  Training
                </span>
                {selectedEnv === 'TRN' && (
                  <div className="absolute -top-1 -right-1 h-3 w-3 rounded-full bg-primary-500 ring-2 ring-slate-800" />
                )}
              </button>

              <button
                type="button"
                onClick={() => setSelectedEnv('PRD')}
                className={`relative flex flex-col items-center justify-center rounded-lg border-2 py-4 px-4 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-error-500 focus:ring-offset-2 focus:ring-offset-slate-800 ${
                  selectedEnv === 'PRD'
                    ? 'border-error-500 bg-error-500/10'
                    : 'border-slate-600 bg-slate-700/50 hover:border-slate-500 hover:bg-slate-700'
                }`}
              >
                <span className={`text-2xl font-bold ${selectedEnv === 'PRD' ? 'text-error-400' : 'text-slate-300'}`}>
                  PRD
                </span>
                <span className={`mt-1 text-xs ${selectedEnv === 'PRD' ? 'text-error-400' : 'text-slate-500'}`}>
                  Production
                </span>
                {selectedEnv === 'PRD' && (
                  <div className="absolute -top-1 -right-1 h-3 w-3 rounded-full bg-error-500 ring-2 ring-slate-800" />
                )}
              </button>
            </div>
          </div>

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
              'Sign in with Infor M3'
            )}
          </button>

          {/* Help text */}
          <p className="mt-6 text-center text-xs text-slate-500">
            You will be redirected to Infor M3 for authentication
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
