'use client';

import React from 'react';
import EmployeeManagementTable from '../components/employee/EmployeeManagementTable';

const EmployeeManagement: React.FC = () => {
  return (
    <div className="w-full">
      <div className="border-b bg-white px-6 py-4">
        <h1 className="text-2xl font-semibold text-gray-900">员工管理</h1>
      </div>
      <EmployeeManagementTable />
    </div>
  );
};

export default EmployeeManagement;