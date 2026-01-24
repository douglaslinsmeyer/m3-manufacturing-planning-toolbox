import React from 'react';
import { AppLayout } from '../components/AppLayout';

function TruckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 00-10.026 0 1.106 1.106 0 00-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" />
    </svg>
  );
}

const Deliveries: React.FC = () => {
  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        {/* Page Header */}
        <div className="mb-6 lg:mb-10">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-success-100 p-2">
              <TruckIcon className="h-6 w-6 text-success-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">Deliveries</h1>
              <p className="mt-1 text-sm text-slate-500">
                Track and manage delivery schedules
              </p>
            </div>
          </div>
        </div>

        {/* Content Placeholder */}
        <div className="rounded-xl bg-white p-8 lg:p-12 shadow-sm ring-1 ring-slate-200">
          <div className="text-center">
            <TruckIcon className="mx-auto h-12 w-12 text-slate-300" />
            <h3 className="mt-4 text-lg font-medium text-slate-900">Deliveries</h3>
            <p className="mt-2 text-sm text-slate-500">
              Delivery tracking and management will be implemented here.
            </p>
          </div>
        </div>
      </div>
    </AppLayout>
  );
};

export default Deliveries;
