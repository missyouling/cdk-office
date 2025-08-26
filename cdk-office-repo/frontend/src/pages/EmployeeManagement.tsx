'use client';

import React from 'react';
import { Box, Typography } from '@mui/material';
import EmployeeManagementTable from '../components/employee/EmployeeManagementTable';

const EmployeeManagement: React.FC = () => {
  return (
    <Box sx={{ width: '100%', p: 2 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        员工管理
      </Typography>
      <EmployeeManagementTable />
    </Box>
  );
};

export default EmployeeManagement;