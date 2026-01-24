import React from 'react';
import { AppLayout } from '../components/AppLayout';

function ShoppingCartIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" />
    </svg>
  );
}

const CustomerOrders: React.FC = () => {
  return (
    <AppLayout>
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        {/* Page Header */}
        <div className="mb-6 lg:mb-10">
          <div className="flex items-center gap-3">
            <div className="rounded-lg bg-info-100 p-2">
              <ShoppingCartIcon className="h-6 w-6 text-info-600" />
            </div>
            <div>
              <h1 className="text-2xl font-semibold text-slate-900">Customer Orders</h1>
              <p className="mt-1 text-sm text-slate-500">
                Browse and search customer orders
              </p>
            </div>
          </div>
        </div>

        {/* Content Placeholder */}
        <div className="rounded-xl bg-white p-8 lg:p-12 shadow-sm ring-1 ring-slate-200">
          <div className="text-center">
            <ShoppingCartIcon className="mx-auto h-12 w-12 text-slate-300" />
            <h3 className="mt-4 text-lg font-medium text-slate-900">Customer Orders</h3>
            <p className="mt-2 text-sm text-slate-500">
              Customer orders list and details will be implemented here.
            </p>
          </div>
        </div>
      </div>
    </AppLayout>
  );
};

export default CustomerOrders;
