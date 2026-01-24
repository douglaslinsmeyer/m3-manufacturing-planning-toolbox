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

function ClipboardDocumentCheckIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M11.35 3.836c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m8.9-4.414c.376.023.75.05 1.124.08 1.131.094 1.976 1.057 1.976 2.192V16.5A2.25 2.25 0 0118 18.75h-2.25m-7.5-10.5H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V18.75m-7.5-10.5h6.375c.621 0 1.125.504 1.125 1.125v9.375m-8.25-3l1.5 1.5 3-3.75" />
    </svg>
  );
}

const PlannedOrderDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        {/* Back Link */}
        <Link
          to="/production-orders"
          className="inline-flex items-center gap-2 text-sm text-slate-500 hover:text-slate-700 no-underline mb-6"
        >
          <ArrowLeftIcon className="h-4 w-4" />
          Back to Production Orders
        </Link>

        {/* Page Header */}
        <div className="mb-6 lg:mb-10">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-info-100 p-2">
              <ClipboardDocumentCheckIcon className="h-6 w-6 text-info-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">Planned Order #{id}</h1>
              <p className="mt-1 text-sm text-slate-500">
                Detailed view of planned manufacturing order
              </p>
            </div>
          </div>
        </div>

        {/* Content Placeholder */}
        <div className="rounded-xl bg-white p-8 shadow-sm ring-1 ring-slate-200">
          <div className="text-center">
            <ClipboardDocumentCheckIcon className="mx-auto h-12 w-12 text-slate-300" />
            <h3 className="mt-4 text-lg font-medium text-slate-900">Planned Order Details</h3>
            <p className="mt-2 text-sm text-slate-500">
              Planned order details and scheduling will be implemented here.
            </p>
          </div>
        </div>
      </div>
    </AppLayout>
  );
};

export default PlannedOrderDetail;
