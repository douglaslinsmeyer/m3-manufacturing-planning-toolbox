import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { IssueSummary } from '../services/api';

interface Detector {
  name: string;
  label: string;
  description: string;
  enabled: boolean;
}

interface Props {
  summary: IssueSummary;
}

export const IssueBreakdownHierarchy: React.FC<Props> = ({ summary }) => {
  const [detectorLabels, setDetectorLabels] = useState<Record<string, string>>({});

  // Fetch detector metadata to get labels
  useEffect(() => {
    const fetchDetectors = async () => {
      try {
        // Use TRN as default - detector labels are environment-agnostic
        const response = await fetch('/api/detection/detectors?environment=TRN');
        const detectors: Detector[] = await response.json();
        const labels: Record<string, string> = {};
        detectors.forEach(d => labels[d.name] = d.label);
        setDetectorLabels(labels);
      } catch (err) {
        console.error('Failed to load detector labels:', err);
      }
    };
    fetchDetectors();
  }, []);

  // Extract facility/warehouse/detector hierarchy
  const facilityData = summary.by_facility_warehouse_detector;

  // Safety check for undefined/null data
  if (!facilityData || typeof facilityData !== 'object') {
    return (
      <div className="text-center py-8 text-slate-500">
        No breakdown data available.
      </div>
    );
  }

  const facilities = Object.keys(facilityData);

  // If no data, show empty state
  if (facilities.length === 0) {
    return (
      <div className="text-center py-8 text-slate-500">
        No issues detected in the current snapshot.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {facilities.map(facility => {
        const warehouses = facilityData[facility];
        const warehouseKeys = Object.keys(warehouses).filter(w => w !== ''); // Filter out empty warehouse keys

        // If no warehouses for this facility, skip
        if (warehouseKeys.length === 0) {
          return null;
        }

        return (
          <div key={facility} className="border border-slate-200 rounded-lg overflow-hidden">
            {/* Facility Header */}
            <div className="bg-slate-50 px-4 py-3 border-b border-slate-200">
              <h3 className="text-sm font-semibold text-slate-700">
                Facility: {facility}
                <span className="ml-2 text-slate-500 font-normal">
                  ({summary.by_facility[facility]} {summary.by_facility[facility] === 1 ? 'issue' : 'issues'})
                </span>
              </h3>
            </div>

            {/* Warehouse x Detector Table */}
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-white">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider">
                      Warehouse
                    </th>
                    {summary.by_detector && Object.keys(summary.by_detector).map(detector => (
                      <th key={detector} className="px-4 py-3 text-center text-xs font-medium text-slate-700 uppercase tracking-wider">
                        {detectorLabels[detector] || detector}
                      </th>
                    ))}
                    <th className="px-4 py-3 text-center text-xs font-medium text-slate-700 uppercase tracking-wider">
                      Total
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-slate-200">
                  {warehouseKeys.map(warehouse => {
                    const detectors = warehouses[warehouse];
                    const warehouseTotal = summary.by_warehouse[warehouse] || 0;

                    return (
                      <tr key={warehouse} className="hover:bg-slate-50 transition-colors">
                        <td className="px-4 py-3 text-sm font-medium text-slate-900">
                          {warehouse}
                        </td>
                        {summary.by_detector && Object.keys(summary.by_detector).map(detector => {
                          const count = detectors[detector] || 0;
                          return (
                            <td key={detector} className="px-4 py-3 text-center">
                              {count > 0 ? (
                                <Link
                                  to={`/issues?warehouse=${warehouse}&detector=${detector}`}
                                  className={`inline-block px-3 py-1 rounded-full text-sm font-medium transition-colors
                                             ${count > 10 ? 'bg-red-100 text-red-800 hover:bg-red-200' :
                                               count > 5 ? 'bg-orange-100 text-orange-800 hover:bg-orange-200' :
                                               'bg-yellow-100 text-yellow-800 hover:bg-yellow-200'}`}
                                >
                                  {count}
                                </Link>
                              ) : (
                                <span className="text-slate-300">—</span>
                              )}
                            </td>
                          );
                        })}
                        <td className="px-4 py-3 text-center text-sm font-semibold text-slate-900">
                          {warehouseTotal}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        );
      })}
    </div>
  );
};
