import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { AppLayout } from '../components/AppLayout';

function ArrowLeftIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
    </svg>
  );
}

function DocumentTextIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
    </svg>
  );
}

const ManufacturingOrderDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        {/* Back Link */}
        <Link
          to="/"
          className="inline-flex items-center gap-2 text-sm text-slate-500 hover:text-slate-700 no-underline mb-6"
        >
          <ArrowLeftIcon className="h-4 w-4" />
          Back to Dashboard
        </Link>

        {/* Page Header */}
        <div className="mb-6 lg:mb-10">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-primary-100 p-2">
              <DocumentTextIcon className="h-6 w-6 text-primary-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">Manufacturing Order #{id}</h1>
              <p className="mt-1 text-sm text-slate-500">
                Detailed view of manufacturing order
              </p>
            </div>
          </div>
        </div>

        {/* Content Placeholder */}
        <div className="rounded-xl bg-white p-8 shadow-sm ring-1 ring-slate-200">
          <div className="text-center">
            <DocumentTextIcon className="mx-auto h-12 w-12 text-slate-300" />
            <h3 className="mt-4 text-lg font-medium text-slate-900">Manufacturing Order Details</h3>
            <p className="mt-2 text-sm text-slate-500">
              Manufacturing order details and operations will be implemented here.
            </p>
          </div>
        </div>
      </div>
    </AppLayout>
  );
};

export default ManufacturingOrderDetail;
