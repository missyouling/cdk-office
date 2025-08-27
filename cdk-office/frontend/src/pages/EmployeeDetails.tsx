import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  CardDescription
} from '@/components/ui/card';
import { 
  Button 
} from '@/components/ui/button';
import { 
  Input 
} from '@/components/ui/input';
import { 
  Label 
} from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { 
  Badge 
} from '@/components/ui/badge';
import { 
  Avatar, 
  AvatarFallback 
} from '@/components/ui/avatar';
import { 
  toast 
} from '@/components/ui/use-toast';

import { 
  User, 
  Edit, 
  Mail, 
  Briefcase, 
  MapPin, 
  Calendar,
  DollarSign
} from 'lucide-react';

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
  // 在实际项目中应该从router.query获取id
  const id = '1'; // 简化处理
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
      toast({
        title: "保存成功",
        description: "员工信息已更新",
      });
    }
  };

  if (!employee) {
    return (
      <div className="container mx-auto p-6">
        <h2 className="text-2xl font-bold">员工未找到</h2>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">员工详情</h1>
          <p className="text-muted-foreground mt-2">
            查看和管理员工详细信息
          </p>
        </div>
        <Button onClick={handleEdit}>
          <Edit className="h-4 w-4 mr-2" />
          编辑信息
        </Button>
      </div>

      {/* 员工信息卡片 */}
      <Card>
        <CardContent className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* 员工头像和基本信息 */}
            <div className="md:col-span-1 flex flex-col items-center">
              <Avatar className="h-32 w-32 mb-4">
                <AvatarFallback>
                  <User className="h-16 w-16" />
                </AvatarFallback>
              </Avatar>
              <h2 className="text-2xl font-bold">
                {employee.firstName} {employee.lastName}
              </h2>
              <Badge 
                variant={
                  employee.status === 'active' 
                    ? 'success' 
                    : employee.status === 'inactive' 
                      ? 'destructive' 
                      : 'warning'
                }
                className="mt-2"
              >
                {employee.status === 'active' ? '活跃' : employee.status === 'inactive' ? '非活跃' : '待定'}
              </Badge>
            </div>

            {/* 详细信息 */}
            <div className="md:col-span-2 space-y-4">
              <div className="flex items-center space-x-3">
                <Mail className="h-5 w-5 text-muted-foreground" />
                <span className="font-medium">邮箱:</span>
                <span>{employee.email}</span>
              </div>
              <div className="flex items-center space-x-3">
                <Briefcase className="h-5 w-5 text-muted-foreground" />
                <span className="font-medium">职位:</span>
                <span>{employee.jobTitle}</span>
              </div>
              <div className="flex items-center space-x-3">
                <MapPin className="h-5 w-5 text-muted-foreground" />
                <span className="font-medium">部门:</span>
                <span>{employee.department}</span>
              </div>
              <div className="flex items-center space-x-3">
                <DollarSign className="h-5 w-5 text-muted-foreground" />
                <span className="font-medium">薪资:</span>
                <span>¥{employee.salary.toLocaleString('zh-CN')}</span>
              </div>
              <div className="flex items-center space-x-3">
                <Calendar className="h-5 w-5 text-muted-foreground" />
                <span className="font-medium">入职日期:</span>
                <span>{new Date(employee.startDate).toLocaleDateString('zh-CN')}</span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 编辑员工信息对话框 */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>编辑员工信息</DialogTitle>
            <DialogDescription>
              更新员工的详细信息
            </DialogDescription>
          </DialogHeader>
          {editingEmployee && (
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="firstName">名字</Label>
                <Input
                  id="firstName"
                  value={editingEmployee.firstName}
                  onChange={(e) => setEditingEmployee({ ...editingEmployee, firstName: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lastName">姓氏</Label>
                <Input
                  id="lastName"
                  value={editingEmployee.lastName}
                  onChange={(e) => setEditingEmployee({ ...editingEmployee, lastName: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">邮箱</Label>
                <Input
                  id="email"
                  type="email"
                  value={editingEmployee.email}
                  onChange={(e) => setEditingEmployee({ ...editingEmployee, email: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="jobTitle">职位</Label>
                <Input
                  id="jobTitle"
                  value={editingEmployee.jobTitle}
                  onChange={(e) => setEditingEmployee({ ...editingEmployee, jobTitle: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="department">部门</Label>
                <Input
                  id="department"
                  value={editingEmployee.department}
                  onChange={(e) => setEditingEmployee({ ...editingEmployee, department: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="salary">薪资</Label>
                <Input
                  id="salary"
                  type="number"
                  value={editingEmployee.salary}
                  onChange={(e) => setEditingEmployee({ ...editingEmployee, salary: Number(e.target.value) })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="status">状态</Label>
                <Select
                  value={editingEmployee.status}
                  onValueChange={(value) => setEditingEmployee({ 
                    ...editingEmployee, 
                    status: value as 'active' | 'inactive' | 'pending' 
                  })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">活跃</SelectItem>
                    <SelectItem value="inactive">非活跃</SelectItem>
                    <SelectItem value="pending">待定</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsEditDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSave}>
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default EmployeeDetails;