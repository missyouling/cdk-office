import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Avatar,
  Chip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import { AccountCircle, Edit, Email, Work, LocationOn, CalendarToday, AttachMoney } from '@mui/icons-material';
import { useRouter } from 'next/router';

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

const EmployeeDetails: React.FC = () => {
  const router = useRouter();
  const { id } = router.query;
  const [employee, setEmployee] = useState<Employee | null>(null);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editingEmployee, setEditingEmployee] = useState<Employee | null>(null);

  useEffect(() => {
    if (id) {
      const foundEmployee = mockEmployees.find(emp => emp.id === id);
      setEmployee(foundEmployee || null);
      if (foundEmployee) {
        setEditingEmployee({ ...foundEmployee });
      }
    }
  }, [id]);

  const handleEdit = () => {
    if (employee) {
      setEditingEmployee({ ...employee });
      setIsEditDialogOpen(true);
    }
  };

  const handleSave = () => {
    if (editingEmployee) {
      setEmployee(editingEmployee);
      setIsEditDialogOpen(false);
    }
  };

  if (!employee) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h5">员工未找到</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">员工详情</Typography>
        <Button
          variant="contained"
          startIcon={<Edit />}
          onClick={handleEdit}
        >
          编辑信息
        </Button>
      </Box>

      <Card>
        <CardContent>
          <Grid container spacing={3}>
            {/* 员工头像和基本信息 */}
            <Grid item xs={12} md={4}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Avatar sx={{ width: 120, height: 120, mb: 2 }}>
                  <AccountCircle sx={{ width: 100, height: 100 }} />
                </Avatar>
                <Typography variant="h5" sx={{ mb: 1 }}>
                  {employee.firstName} {employee.lastName}
                </Typography>
                <Chip
                  label={employee.status === 'active' ? '活跃' : employee.status === 'inactive' ? '非活跃' : '待定'}
                  color={
                    employee.status === 'active'
                      ? 'success'
                      : employee.status === 'inactive'
                      ? 'error'
                      : 'warning'
                  }
                />
              </Box>
            </Grid>

            {/* 详细信息 */}
            <Grid item xs={12} md={8}>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <Email sx={{ mr: 1, color: 'text.secondary' }} />
                    <Typography variant="body1">
                      <strong>邮箱:</strong> {employee.email}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <Work sx={{ mr: 1, color: 'text.secondary' }} />
                    <Typography variant="body1">
                      <strong>职位:</strong> {employee.jobTitle}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <LocationOn sx={{ mr: 1, color: 'text.secondary' }} />
                    <Typography variant="body1">
                      <strong>部门:</strong> {employee.department}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <AttachMoney sx={{ mr: 1, color: 'text.secondary' }} />
                    <Typography variant="body1">
                      <strong>薪资:</strong> ¥{employee.salary.toLocaleString('zh-CN')}
                    </Typography>
                  </Box>
                </Grid>
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <CalendarToday sx={{ mr: 1, color: 'text.secondary' }} />
                    <Typography variant="body1">
                      <strong>入职日期:</strong> {new Date(employee.startDate).toLocaleDateString('zh-CN')}
                    </Typography>
                  </Box>
                </Grid>
              </Grid>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

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
          <Button onClick={handleSave} variant="contained">保存</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default EmployeeDetails;