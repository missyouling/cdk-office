import React, { useMemo, useState } from 'react';
import {
  MaterialReactTable,
  useMaterialReactTable,
  type MRT_ColumnDef,
  type MRT_Row,
} from 'material-react-table';
import {
  Box,
  Button,
  IconButton,
  MenuItem,
  Typography,
  Chip,
  Avatar,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Select,
  FormControl,
  InputLabel,
  Alert,
} from '@mui/material';
import { 
  AccountCircle, 
  Send, 
  Edit, 
  Delete,
  FileDownload,
  FileUpload,
  Print,
  Share,
} from '@mui/icons-material';

// 员工数据接口
interface Employee {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  jobTitle: string;
  department: string;
  salary: number;
  startDate: string;
  status: 'active' | 'inactive' | 'pending';
  avatar?: string;
}

// 模拟员工数据
const mockEmployees: Employee[] = [
  {
    id: '1',
    firstName: '张',
    lastName: '三',
    email: 'zhangsan@company.com',
    jobTitle: '软件工程师',
    department: '技术部',
    salary: 15000,
    startDate: '2022-03-15',
    status: 'active',
    avatar: '',
  },
  {
    id: '2',
    firstName: '李',
    lastName: '四',
    email: 'lisi@company.com',
    jobTitle: '产品经理',
    department: '产品部',
    salary: 18000,
    startDate: '2021-07-22',
    status: 'active',
    avatar: '',
  },
  {
    id: '3',
    firstName: '王',
    lastName: '五',
    email: 'wangwu@company.com',
    jobTitle: '设计师',
    department: '设计部',
    salary: 12000,
    startDate: '2023-01-10',
    status: 'pending',
    avatar: '',
  },
];

const EmployeeManagementTable: React.FC = () => {
  const [employees, setEmployees] = useState<Employee[]>(mockEmployees);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

  // 列定义
  const columns = useMemo<MRT_ColumnDef<Employee>[]>(
    () => [
      {
        id: 'employee',
        header: '员工信息',
        columns: [
          {
            accessorFn: (row) => `${row.firstName} ${row.lastName}`,
            id: 'name',
            header: '姓名',
            size: 150,
            Cell: ({ renderedCellValue, row }) => (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <Avatar>
                  <AccountCircle />
                </Avatar>
                <span>{renderedCellValue}</span>
              </Box>
            ),
          },
          {
            accessorKey: 'email',
            header: '邮箱',
            size: 200,
            enableClickToCopy: true,
          },
        ],
      },
      {
        id: 'job',
        header: '职位信息',
        columns: [
          {
            accessorKey: 'jobTitle',
            header: '职位',
            size: 150,
          },
          {
            accessorKey: 'department',
            header: '部门',
            size: 150,
          },
          {
            accessorKey: 'salary',
            header: '薪资',
            size: 120,
            Cell: ({ cell }) => (
              <Box
                component="span"
                sx={(theme) => ({
                  backgroundColor:
                    cell.getValue<number>() < 10000
                      ? theme.palette.error.dark
                      : cell.getValue<number>() >= 10000 &&
                        cell.getValue<number>() < 15000
                      ? theme.palette.warning.dark
                      : theme.palette.success.dark,
                  borderRadius: '0.25rem',
                  color: '#fff',
                  maxWidth: '9ch',
                  p: '0.25rem',
                })}
              >
                {cell.getValue<number>()?.toLocaleString?.('zh-CN', {
                  style: 'currency',
                  currency: 'CNY',
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 0,
                })}
              </Box>
            ),
          },
          {
            accessorFn: (row) => new Date(row.startDate),
            id: 'startDate',
            header: '入职日期',
            filterVariant: 'date',
            Cell: ({ cell }) => cell.getValue<Date>()?.toLocaleDateString('zh-CN'),
          },
          {
            accessorKey: 'status',
            header: '状态',
            filterVariant: 'select',
            Cell: ({ cell }) => (
              <Chip
                label={cell.getValue<string>()}
                color={
                  cell.getValue<string>() === 'active'
                    ? 'success'
                    : cell.getValue<string>() === 'inactive'
                    ? 'error'
                    : 'warning'
                }
              />
            ),
          },
        ],
      },
    ],
    [],
  );

  // 处理导出CSV
  const exportToCSV = (data: Employee[]) => {
    if (data.length === 0) {
      alert('没有选中的数据可以导出');
      return;
    }

    // 创建CSV内容
    const headers = ['ID', '姓名', '邮箱', '职位', '部门', '薪资', '入职日期', '状态'];
    const csvContent = [
      headers.join(','),
      ...data.map(emp => [
        emp.id,
        `${emp.firstName} ${emp.lastName}`,
        emp.email,
        emp.jobTitle,
        emp.department,
        emp.salary,
        emp.startDate,
        emp.status
      ].map(field => `"${field}"`).join(','))
    ].join('\n');

    // 创建下载链接
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `员工数据_${new Date().toISOString().slice(0, 10)}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // 处理导入CSV
  const importFromCSV = () => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.csv';
    
    input.onchange = (event: any) => {
      const file = event.target.files[0];
      if (!file) return;
      
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        const lines = content.split('\n');
        
        // 解析CSV数据
        const newEmployees: Employee[] = [];
        for (let i = 1; i < lines.length; i++) {
          const line = lines[i].trim();
          if (!line) continue;
          
          // 简化的CSV解析（实际应用中可能需要更复杂的解析）
          const fields = line.split(',').map(field => field.replace(/^"|"$/g, ''));
          if (fields.length >= 8) {
            newEmployees.push({
              id: fields[0] || `emp_${Date.now()}_${i}`,
              firstName: fields[1].split(' ')[0] || '',
              lastName: fields[1].split(' ')[1] || '',
              email: fields[2] || '',
              jobTitle: fields[3] || '',
              department: fields[4] || '',
              salary: parseInt(fields[5]) || 0,
              startDate: fields[6] || '',
              status: fields[7] as 'active' | 'inactive' | 'pending' || 'active',
            });
          }
        }
        
        // 更新员工数据
        setEmployees([...employees, ...newEmployees]);
        alert(`成功导入 ${newEmployees.length} 条员工记录`);
      };
      
      reader.readAsText(file, 'UTF-8');
    };
    
    input.click();
  };

  // 下载模板
  const downloadTemplate = () => {
    const templateContent = `ID,姓名,邮箱,职位,部门,薪资,入职日期,状态\n"","","","","","","","active/inactive/pending"`;
    const blob = new Blob([templateContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', '员工数据导入模板.csv');
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // 打印数据
  const printData = (data: Employee[]) => {
    if (data.length === 0) {
      alert('没有选中的数据可以打印');
      return;
    }

    // 创建打印窗口
    const printWindow = window.open('', '_blank');
    if (printWindow) {
      printWindow.document.write(`
        <html>
          <head>
            <title>员工数据</title>
            <style>
              table { border-collapse: collapse; width: 100%; }
              th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
              th { background-color: #f2f2f2; }
            </style>
          </head>
          <body>
            <h1>员工数据</h1>
            <table>
              <tr>
                <th>姓名</th>
                <th>邮箱</th>
                <th>职位</th>
                <th>部门</th>
                <th>薪资</th>
                <th>入职日期</th>
                <th>状态</th>
              </tr>
              ${data.map(emp => `
                <tr>
                  <td>${emp.firstName} ${emp.lastName}</td>
                  <td>${emp.email}</td>
                  <td>${emp.jobTitle}</td>
                  <td>${emp.department}</td>
                  <td>¥${emp.salary.toLocaleString('zh-CN')}</td>
                  <td>${new Date(emp.startDate).toLocaleDateString('zh-CN')}</td>
                  <td>${emp.status}</td>
                </tr>
              `).join('')}
            </table>
          </body>
        </html>
      `);
      printWindow.document.close();
      printWindow.print();
    }
  };

  // 编辑员工
  const handleEditEmployee = (employee: Employee) => {
    setEditingEmployee({ ...employee });
    setIsEditDialogOpen(true);
  };

  // 保存编辑的员工信息
  const handleSaveEmployee = () => {
    if (editingEmployee) {
      setEmployees(employees.map(emp => emp.id === editingEmployee.id ? editingEmployee : emp));
      setIsEditDialogOpen(false);
      setEditingEmployee(null);
    }
  };

  // 删除员工
  const handleDeleteEmployee = (employeeId: string) => {
    if (window.confirm('确定要删除这个员工吗？')) {
      setEmployees(employees.filter(emp => emp.id !== employeeId));
    }
  };

  // 表格配置
  const table = useMaterialReactTable({
    columns,
    data: employees,
    enableColumnFilterModes: true,
    enableColumnOrdering: true,
    enableGrouping: true,
    enableColumnPinning: true,
    enableFacetedValues: true,
    enableRowActions: true,
    enableRowSelection: true,
    enableEditing: true, // 启用编辑功能
    initialState: {
      showColumnFilters: true,
      showGlobalFilter: true,
      columnPinning: {
        left: ['mrt-row-select'],
        right: ['mrt-row-actions'],
      },
    },
    paginationDisplayMode: 'pages',
    positionToolbarAlertBanner: 'bottom',
    muiSearchTextFieldProps: {
      size: 'small',
      variant: 'outlined',
    },
    // 工具栏按钮
    renderTopToolbarCustomActions: ({ table }) => (
      <Box sx={{ display: 'flex', gap: '1rem', p: '0.5rem', flexWrap: 'wrap' }}>
        <Button
          color="primary"
          onClick={() => {
            // 导出选中行
            const selectedRows = table.getSelectedRowModel().flatRows;
            if (selectedRows.length > 0) {
              exportToCSV(selectedRows.map(row => row.original));
            } else {
              // 如果没有选中行，导出所有数据
              exportToCSV(employees);
            }
          }}
          variant="contained"
          startIcon={<FileDownload />}
        >
          导出数据
        </Button>
        <Button
          color="secondary"
          onClick={importFromCSV}
          variant="outlined"
          startIcon={<FileUpload />}
        >
          导入数据
        </Button>
        <Button
          color="success"
          onClick={downloadTemplate}
          variant="outlined"
          startIcon={<FileDownload />}
        >
          下载模板
        </Button>
        <Button
          color="info"
          onClick={() => {
            // 打印选中行
            const selectedRows = table.getSelectedRowModel().flatRows;
            if (selectedRows.length > 0) {
              printData(selectedRows.map(row => row.original));
            } else {
              // 如果没有选中行，打印所有数据
              printData(employees);
            }
          }}
          variant="outlined"
          startIcon={<Print />}
        >
          打印数据
        </Button>
        <Button
          color="warning"
          onClick={() => {
            // 分享选中行
            const selectedRows = table.getSelectedRowModel().flatRows;
            if (selectedRows.length > 0) {
              alert(`分享 ${selectedRows.length} 条员工记录的功能待实现`);
            } else {
              alert('请先选择要分享的员工记录');
            }
          }}
          variant="outlined"
          startIcon={<Share />}
        >
          分享选中
        </Button>
      </Box>
    ),
    // 行操作菜单
    renderRowActionMenuItems: ({ closeMenu, row }) => [
      <MenuItem
        key={0}
        onClick={() => {
          // 编辑员工
          handleEditEmployee(row.original);
          closeMenu();
        }}
        sx={{ m: 0 }}
      >
        <Edit />
        <span style={{ marginLeft: '0.5rem' }}>编辑</span>
      </MenuItem>,
      <MenuItem
        key={1}
        onClick={() => {
          // 删除员工
          handleDeleteEmployee(row.original.id);
          closeMenu();
        }}
        sx={{ m: 0 }}
      >
        <Delete />
        <span style={{ marginLeft: '0.5rem' }}>删除</span>
      </MenuItem>,
    ],
  });

  return (
    <>
      <MaterialReactTable table={table} />
      
      {/* 编辑员工对话框 */}
      <Dialog open={isEditDialogOpen} onClose={() => setIsEditDialogOpen(false)}>
        <DialogTitle>编辑员工信息</DialogTitle>
        <DialogContent>
          {editingEmployee && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: '1rem', minWidth: '400px', mt: '1rem' }}>
              <TextField
                label="名字"
                value={editingEmployee.firstName}
                onChange={(e) => setEditingEmployee({...editingEmployee, firstName: e.target.value})}
                fullWidth
              />
              <TextField
                label="姓氏"
                value={editingEmployee.lastName}
                onChange={(e) => setEditingEmployee({...editingEmployee, lastName: e.target.value})}
                fullWidth
              />
              <TextField
                label="邮箱"
                value={editingEmployee.email}
                onChange={(e) => setEditingEmployee({...editingEmployee, email: e.target.value})}
                fullWidth
              />
              <TextField
                label="职位"
                value={editingEmployee.jobTitle}
                onChange={(e) => setEditingEmployee({...editingEmployee, jobTitle: e.target.value})}
                fullWidth
              />
              <TextField
                label="部门"
                value={editingEmployee.department}
                onChange={(e) => setEditingEmployee({...editingEmployee, department: e.target.value})}
                fullWidth
              />
              <TextField
                label="薪资"
                type="number"
                value={editingEmployee.salary}
                onChange={(e) => setEditingEmployee({...editingEmployee, salary: Number(e.target.value)})}
                fullWidth
              />
              <TextField
                label="入职日期"
                type="date"
                value={editingEmployee.startDate}
                onChange={(e) => setEditingEmployee({...editingEmployee, startDate: e.target.value})}
                fullWidth
                InputLabelProps={{
                  shrink: true,
                }}
              />
              <FormControl fullWidth>
                <InputLabel>状态</InputLabel>
                <Select
                  value={editingEmployee.status}
                  onChange={(e) => setEditingEmployee({...editingEmployee, status: e.target.value as any})}
                >
                  <MenuItem value="active">活跃</MenuItem>
                  <MenuItem value="inactive">非活跃</MenuItem>
                  <MenuItem value="pending">待定</MenuItem>
                </Select>
              </FormControl>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsEditDialogOpen(false)}>取消</Button>
          <Button onClick={handleSaveEmployee} variant="contained">保存</Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default EmployeeManagementTable;