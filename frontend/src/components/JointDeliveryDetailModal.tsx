import React from 'react';

interface JointDeliveryOrder {
  number: string;
  type: 'MO' | 'MOP';
  date: string;  // YYYYMMDD
  co_line: string;
  product_number?: string;
  mo_type?: string;
  quantity?: number;
  confirmed_delivery_date?: string;  // YYYYMMDD
  requested_delivery_date?: string;  // YYYYMMDD
}

function ArrowsRightLeftIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
    </svg>
  );
}

interface JointDeliveryDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAlignEarliest?: () => void;
  issueData: {
    jdcd: string;
    dates: string[];  // YYYYMMDD format
    min_date: number;
    max_date: number;
    num_co_lines: number;
    num_production_orders: number;
    orders: JointDeliveryOrder[];
    tolerance_days: number;
    item_number?: string;
    warehouse: string;
    company: string;
  };
  coNumber?: string;
  currentOrderNumber?: string;  // Highlight this order in the table
}

// Format M3 date (YYYYMMDD) to readable format
function formatDate(dateInt: number | string): string {
  if (!dateInt) return '';
  const str = dateInt.toString();
  if (str.length !== 8) return str;
  const year = str.substring(0, 4);
  const month = str.substring(4, 6);
  const day = str.substring(6, 8);
  const date = new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

// Calculate days between two M3 dates
function dateDiffDays(date1: number | string, date2: number | string): number {
  const str1 = date1.toString();
  const str2 = date2.toString();
  const d1 = new Date(
    parseInt(str1.substring(0, 4)),
    parseInt(str1.substring(4, 6)) - 1,
    parseInt(str1.substring(6, 8))
  );
  const d2 = new Date(
    parseInt(str2.substring(0, 4)),
    parseInt(str2.substring(4, 6)) - 1,
    parseInt(str2.substring(6, 8))
  );
  return Math.abs(Math.floor((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24)));
}

export const JointDeliveryDetailModal: React.FC<JointDeliveryDetailModalProps> = ({
  isOpen,
  onClose,
  onAlignEarliest,
  issueData,
  coNumber,
  currentOrderNumber,
}) => {
  if (!isOpen) return null;

  const dateVarianceDays = dateDiffDays(issueData.min_date, issueData.max_date);

  // Sort orders by date (earliest first)
  const sortedOrders = [...issueData.orders].sort((a, b) => {
    const dateA = parseInt(a.date);
    const dateB = parseInt(b.date);
    return dateA - dateB;
  });

  // Extract unique delivery dates
  const confirmedDeliveryDates = Array.from(
    new Set(issueData.orders.map(o => o.confirmed_delivery_date).filter(d => d && d !== '0'))
  );
  const requestedDeliveryDates = Array.from(
    new Set(issueData.orders.map(o => o.requested_delivery_date).filter(d => d && d !== '0'))
  );

  // Determine variance badge color
  const getVarianceBadgeColor = (days: number) => {
    if (days > 7) return 'bg-red-100 text-red-800 ring-red-300';
    if (days >= 3) return 'bg-orange-100 text-orange-800 ring-orange-300';
    return 'bg-yellow-100 text-yellow-800 ring-yellow-300';
  };

  // Handle escape key
  React.useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, onClose]);

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        {/* Backdrop */}
        <div
          className="fixed inset-0 bg-slate-500 bg-opacity-75 transition-opacity"
          onClick={onClose}
        />

        {/* Modal */}
        <div className="relative transform overflow-hidden rounded-lg bg-white shadow-xl transition-all w-full max-w-4xl">
          {/* Header */}
          <div className="bg-gradient-to-r from-blue-600 to-blue-700 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-white">
                Joint Delivery Group Details
              </h3>
              <p className="text-sm text-blue-100 mt-1">
                JDCD: <span className="font-mono font-medium">{issueData.jdcd}</span>
                {coNumber && <span className="ml-3">• CO: {coNumber}</span>}
              </p>
            </div>
            <button
              onClick={onClose}
              className="text-blue-100 hover:text-white transition-colors"
            >
              <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Content */}
          <div className="px-6 py-5">
            {/* Summary Card */}
            <div className="bg-slate-50 rounded-lg p-4 mb-5 border border-slate-200">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <div className="text-xs font-medium text-slate-600 uppercase tracking-wider mb-1">
                    Date Range
                  </div>
                  <div className="text-sm font-semibold text-slate-900">
                    {formatDate(issueData.min_date)}
                    <span className="mx-2 text-slate-400">→</span>
                    {formatDate(issueData.max_date)}
                  </div>
                </div>
                <div>
                  <div className="text-xs font-medium text-slate-600 uppercase tracking-wider mb-1">
                    Variance
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-sm font-semibold ring-1 ${getVarianceBadgeColor(dateVarianceDays)}`}>
                    {dateVarianceDays} {dateVarianceDays === 1 ? 'day' : 'days'}
                  </span>
                </div>
                <div>
                  <div className="text-xs font-medium text-slate-600 uppercase tracking-wider mb-1">
                    Production Orders
                  </div>
                  <div className="text-sm font-semibold text-slate-900">
                    {issueData.num_production_orders} {issueData.num_production_orders === 1 ? 'order' : 'orders'}
                  </div>
                </div>
              </div>

              {/* Delivery Dates Section */}
              {(confirmedDeliveryDates.length > 0 || requestedDeliveryDates.length > 0) && (
                <div className="mt-4 pt-4 border-t border-slate-200">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {confirmedDeliveryDates.length > 0 && (
                      <div>
                        <div className="text-xs font-medium text-slate-600 uppercase tracking-wider mb-1">
                          Confirmed Delivery Date{confirmedDeliveryDates.length > 1 ? 's' : ''}
                        </div>
                        <div className="text-sm font-semibold text-slate-900">
                          {confirmedDeliveryDates.map(d => formatDate(d)).join(', ')}
                        </div>
                      </div>
                    )}
                    {requestedDeliveryDates.length > 0 && (
                      <div>
                        <div className="text-xs font-medium text-slate-600 uppercase tracking-wider mb-1">
                          Requested Delivery Date{requestedDeliveryDates.length > 1 ? 's' : ''}
                        </div>
                        <div className="text-sm font-semibold text-slate-900">
                          {requestedDeliveryDates.map(d => formatDate(d)).join(', ')}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {issueData.tolerance_days !== undefined && (
                <div className="mt-3 pt-3 border-t border-slate-200">
                  <span className="text-xs text-slate-600">
                    Tolerance: {issueData.tolerance_days} {issueData.tolerance_days === 1 ? 'day' : 'days'}
                    {issueData.tolerance_days === 0 && ' (exact match required)'}
                  </span>
                </div>
              )}
            </div>

            {/* Production Orders Table */}
            <div className="overflow-hidden border border-slate-200 rounded-lg">
              <div className="overflow-x-auto max-h-96">
                <table className="min-w-full divide-y divide-slate-200">
                  <thead className="bg-slate-50 sticky top-0">
                    <tr>
                      <th className="px-3 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider">
                        Order #
                      </th>
                      <th className="px-3 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider">
                        Type
                      </th>
                      <th className="px-3 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider">
                        CO Line
                      </th>
                      <th className="px-3 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider">
                        Item/Product
                      </th>
                      <th className="px-3 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider">
                        Planned Start
                      </th>
                      <th className="px-3 py-3 text-center text-xs font-medium text-slate-700 uppercase tracking-wider">
                        Variance
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-slate-200">
                    {sortedOrders.map((order, index) => {
                      const orderDateInt = parseInt(order.date);
                      const varianceFromMin = dateDiffDays(issueData.min_date, orderDateInt);
                      const isMinDate = orderDateInt === issueData.min_date;
                      const isMaxDate = orderDateInt === issueData.max_date;
                      const isCurrentOrder = currentOrderNumber === order.number;

                      return (
                        <tr
                          key={`${order.type}-${order.number}`}
                          className={`
                            ${isCurrentOrder ? 'ring-2 ring-blue-500 bg-blue-50' : ''}
                            ${isMinDate && !isCurrentOrder ? 'bg-green-50' : ''}
                            ${isMaxDate && !isCurrentOrder ? 'bg-orange-50' : ''}
                            ${!isMinDate && !isMaxDate && !isCurrentOrder ? 'hover:bg-slate-50' : ''}
                            transition-colors
                          `}
                        >
                          <td className="px-3 py-3 text-sm">
                            <div className="flex items-center gap-2">
                              <span className="font-mono font-medium text-slate-900">
                                {order.number}
                              </span>
                              {isCurrentOrder && (
                                <span className="text-xs text-blue-600 font-medium">(Current)</span>
                              )}
                            </div>
                          </td>
                          <td className="px-3 py-3 text-sm">
                            <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                              order.type === 'MO'
                                ? 'bg-purple-100 text-purple-800'
                                : 'bg-indigo-100 text-indigo-800'
                            }`}>
                              {order.type}
                            </span>
                          </td>
                          <td className="px-3 py-3 text-sm font-mono text-slate-700">
                            {order.co_line || '—'}
                          </td>
                          <td className="px-3 py-3 text-sm text-slate-700">
                            <div className="max-w-xs truncate" title={order.product_number || issueData.item_number}>
                              {order.product_number || issueData.item_number || '—'}
                            </div>
                          </td>
                          <td className="px-3 py-3 text-sm">
                            <div className="flex items-center gap-2">
                              <span className="font-medium text-slate-900">
                                {formatDate(order.date)}
                              </span>
                              {isMinDate && (
                                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                                  Earliest
                                </span>
                              )}
                              {isMaxDate && (
                                <span className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-orange-100 text-orange-800">
                                  Latest
                                </span>
                              )}
                            </div>
                          </td>
                          <td className="px-3 py-3 text-center text-sm">
                            {varianceFromMin === 0 ? (
                              <span className="text-slate-400">—</span>
                            ) : (
                              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                                varianceFromMin > 7 ? 'bg-red-100 text-red-800' :
                                varianceFromMin >= 3 ? 'bg-orange-100 text-orange-800' :
                                'bg-yellow-100 text-yellow-800'
                              }`}>
                                +{varianceFromMin}d
                              </span>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Info Message */}
            {issueData.num_production_orders > 1 && (
              <div className="mt-4 rounded-lg bg-blue-50 p-3 text-sm text-blue-800">
                <div className="flex items-start gap-2">
                  <svg className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p>
                    All production orders in this joint delivery group (JDCD {issueData.jdcd}) should ideally have the same planned start date
                    to ensure synchronized delivery. Current variance: <strong>{dateVarianceDays} days</strong>.
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="bg-slate-50 px-6 py-4 flex justify-between items-center">
            {/* Left: Align button */}
            {onAlignEarliest && (
              <button
                type="button"
                onClick={onAlignEarliest}
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-500 transition-colors shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              >
                <ArrowsRightLeftIcon className="h-4 w-4" />
                Align to Earliest Date
              </button>
            )}

            {/* Right: Close button */}
            <button
              type="button"
              onClick={onClose}
              className="inline-flex justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-slate-900 shadow-sm ring-1 ring-inset ring-slate-300 hover:bg-slate-50 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
