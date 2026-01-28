import React, { useState, useEffect } from 'react';
import { AppLayout } from '../components/AppLayout';
import { api } from '../services/api';
import type { AuditLog } from '../types';
import { ToastContainer } from '../components/Toast';
import { useToast } from '../hooks/useToast';

const AuditLogs: React.FC = () => {
  const toast = useToast();

  // State
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  // Filters state
  const [selectedEntityType, setSelectedEntityType] = useState('');
  const [selectedOperation, setSelectedOperation] = useState('');
  const [selectedFacility, setSelectedFacility] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');

  // Expanded row tracking
  const [expandedRows, setExpandedRows] = useState<Set<number>>(new Set());

  // Load logs on mount and filter/page change
  useEffect(() => {
    loadAuditLogs();
  }, [currentPage, pageSize, selectedEntityType, selectedOperation, selectedFacility, startDate, endDate]);

  const loadAuditLogs = async () => {
    setLoading(true);
    try {
      const response = await api.listAuditLogs({
        entityType: selectedEntityType || undefined,
        operation: selectedOperation || undefined,
        facility: selectedFacility || undefined,
        startTime: startDate ? new Date(startDate).toISOString() : undefined,
        endTime: endDate ? new Date(endDate).toISOString() : undefined,
        page: currentPage,
        pageSize,
      });
      setLogs(response.data);
      setTotalCount(response.pagination.totalCount);
      setTotalPages(response.pagination.totalPages);
    } catch (err: any) {
      console.error('Failed to load audit logs:', err);
      toast.error(err.response?.data?.message || 'Failed to load audit logs');
    } finally {
      setLoading(false);
    }
  };

  const toggleRowExpand = (logId: number) => {
    setExpandedRows(prev => {
      const next = new Set(prev);
      if (next.has(logId)) {
        next.delete(logId);
      } else {
        next.add(logId);
      }
      return next;
    });
  };

  const clearFilters = () => {
    setSelectedEntityType('');
    setSelectedOperation('');
    setSelectedFacility('');
    setStartDate('');
    setEndDate('');
    setCurrentPage(1);
  };

  const formatDateTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  const getUserDisplay = (log: AuditLog) => {
    if (log.userName) return log.userName;
    if (log.userId) return log.userId;
    if (log.ipAddress) return `IP: ${log.ipAddress}`;
    return '-';
  };

  return (
    <AppLayout>
      <ToastContainer toasts={toast.toasts} onClose={toast.removeToast} />
      <div className="px-4 py-6 sm:px-6 lg:px-12 lg:py-10">
        <div className="max-w-full mx-auto">
          {/* Header */}
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-slate-900">Audit Log</h1>
            <p className="mt-2 text-sm text-slate-600">
              View system activity and user actions
            </p>
          </div>

          {/* Filter Controls */}
          <div className="bg-white shadow rounded-lg p-6 mb-6">
            <h2 className="text-lg font-medium text-slate-900 mb-4">Filters</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {/* Entity Type */}
              <div>
                <label htmlFor="entityType" className="block text-sm font-medium text-slate-700 mb-1">
                  Entity Type
                </label>
                <select
                  id="entityType"
                  value={selectedEntityType}
                  onChange={(e) => {
                    setSelectedEntityType(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Entity Types</option>
                  <option value="issue">Issue</option>
                  <option value="jdcd_group">JDCD Group</option>
                  <option value="context_cache">Context Cache</option>
                </select>
              </div>

              {/* Operation */}
              <div>
                <label htmlFor="operation" className="block text-sm font-medium text-slate-700 mb-1">
                  Operation
                </label>
                <select
                  id="operation"
                  value={selectedOperation}
                  onChange={(e) => {
                    setSelectedOperation(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Operations</option>
                  <option value="ignore">Ignore</option>
                  <option value="unignore">Unignore</option>
                  <option value="delete_mop">Delete MOP</option>
                  <option value="delete_mo">Delete MO</option>
                  <option value="close_mo">Close MO</option>
                  <option value="align_earliest">Align Earliest</option>
                  <option value="refresh_all">Refresh All</option>
                </select>
              </div>

              {/* Facility */}
              <div>
                <label htmlFor="facility" className="block text-sm font-medium text-slate-700 mb-1">
                  Facility
                </label>
                <input
                  id="facility"
                  type="text"
                  value={selectedFacility}
                  onChange={(e) => {
                    setSelectedFacility(e.target.value);
                    setCurrentPage(1);
                  }}
                  placeholder="e.g., AZ1"
                  className="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              {/* Start Date */}
              <div>
                <label htmlFor="startDate" className="block text-sm font-medium text-slate-700 mb-1">
                  Start Date
                </label>
                <input
                  id="startDate"
                  type="date"
                  value={startDate}
                  onChange={(e) => {
                    setStartDate(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              {/* End Date */}
              <div>
                <label htmlFor="endDate" className="block text-sm font-medium text-slate-700 mb-1">
                  End Date
                </label>
                <input
                  id="endDate"
                  type="date"
                  value={endDate}
                  onChange={(e) => {
                    setEndDate(e.target.value);
                    setCurrentPage(1);
                  }}
                  className="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>

              {/* Clear Filters */}
              <div className="flex items-end">
                <button
                  onClick={clearFilters}
                  className="w-full px-4 py-2 border border-slate-300 rounded-md shadow-sm text-sm font-medium text-slate-700 bg-white hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Clear Filters
                </button>
              </div>
            </div>
          </div>

          {/* Audit Log Table */}
          <div className="bg-white shadow rounded-lg overflow-hidden">
            {loading ? (
              <div className="p-8 text-center">
                <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"></div>
                <p className="mt-2 text-sm text-slate-600">Loading audit logs...</p>
              </div>
            ) : logs.length === 0 ? (
              <div className="p-8 text-center text-slate-500">
                No audit logs found
              </div>
            ) : (
              <>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-slate-200">
                    <thead className="bg-slate-50">
                      <tr>
                        <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                          Timestamp
                        </th>
                        <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                          User
                        </th>
                        <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                          Entity
                        </th>
                        <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                          Operation
                        </th>
                        <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                          Facility
                        </th>
                        <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                          Details
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-slate-200">
                      {logs.map((log) => (
                        <React.Fragment key={log.id}>
                          <tr className="hover:bg-slate-50">
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-900">
                              {formatDateTime(log.timestamp)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-900">
                              {getUserDisplay(log)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-900">
                              <div>
                                <span className="font-medium">{log.entityType}</span>
                                {log.entityId && (
                                  <span className="text-slate-500 ml-1">#{log.entityId}</span>
                                )}
                              </div>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                {log.operation}
                              </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-900">
                              {log.facility || '-'}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm">
                              {log.metadata && (
                                <button
                                  onClick={() => toggleRowExpand(log.id)}
                                  className="text-blue-600 hover:text-blue-900 focus:outline-none"
                                >
                                  {expandedRows.has(log.id) ? 'Hide' : 'Show'}
                                </button>
                              )}
                            </td>
                          </tr>
                          {expandedRows.has(log.id) && log.metadata && (
                            <tr>
                              <td colSpan={6} className="px-6 py-4 bg-slate-50">
                                <div className="text-xs">
                                  <h4 className="font-medium text-slate-900 mb-2">Metadata:</h4>
                                  <pre className="bg-white p-4 rounded border border-slate-200 overflow-x-auto text-slate-700">
                                    {JSON.stringify(log.metadata, null, 2)}
                                  </pre>
                                </div>
                              </td>
                            </tr>
                          )}
                        </React.Fragment>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Pagination Controls */}
                <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-slate-200 sm:px-6">
                  <div className="flex-1 flex justify-between sm:hidden">
                    <button
                      onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                      disabled={currentPage === 1}
                      className="relative inline-flex items-center px-4 py-2 border border-slate-300 text-sm font-medium rounded-md text-slate-700 bg-white hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Previous
                    </button>
                    <button
                      onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                      disabled={currentPage === totalPages}
                      className="ml-3 relative inline-flex items-center px-4 py-2 border border-slate-300 text-sm font-medium rounded-md text-slate-700 bg-white hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Next
                    </button>
                  </div>
                  <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                    <div>
                      <p className="text-sm text-slate-700">
                        Showing <span className="font-medium">{(currentPage - 1) * pageSize + 1}</span> to{' '}
                        <span className="font-medium">{Math.min(currentPage * pageSize, totalCount)}</span> of{' '}
                        <span className="font-medium">{totalCount}</span> results
                      </p>
                    </div>
                    <div className="flex items-center space-x-4">
                      {/* Page Size Selector */}
                      <div className="flex items-center">
                        <label htmlFor="pageSize" className="text-sm text-slate-700 mr-2">
                          Per page:
                        </label>
                        <select
                          id="pageSize"
                          value={pageSize}
                          onChange={(e) => {
                            setPageSize(Number(e.target.value));
                            setCurrentPage(1);
                          }}
                          className="px-2 py-1 border border-slate-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                          <option value="25">25</option>
                          <option value="50">50</option>
                          <option value="100">100</option>
                          <option value="200">200</option>
                        </select>
                      </div>

                      {/* Pagination Buttons */}
                      <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
                        <button
                          onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                          disabled={currentPage === 1}
                          className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-slate-300 bg-white text-sm font-medium text-slate-500 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          <span className="sr-only">Previous</span>
                          ‹
                        </button>
                        <span className="relative inline-flex items-center px-4 py-2 border border-slate-300 bg-white text-sm font-medium text-slate-700">
                          Page {currentPage} of {totalPages}
                        </span>
                        <button
                          onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                          disabled={currentPage === totalPages}
                          className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-slate-300 bg-white text-sm font-medium text-slate-500 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          <span className="sr-only">Next</span>
                          ›
                        </button>
                      </nav>
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </AppLayout>
  );
};

export default AuditLogs;
