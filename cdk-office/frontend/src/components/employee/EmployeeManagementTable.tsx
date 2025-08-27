import React, { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { 
  User, 
  Edit, 
  Trash2,
  Download,
  Upload,
  Printer,
  Share,
  MoreHorizontal,
  Search,
  Filter,
} from 'lucide-react';

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
  const [selectedEmployees, setSelectedEmployees] = useState<string[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [departmentFilter, setDepartmentFilter] = useState<string>('all');

  // 过滤员工数据
  const filteredEmployees = employees.filter(employee => {
    const matchesSearch = 
      `${employee.firstName} ${employee.lastName}`.toLowerCase().includes(searchTerm.toLowerCase()) ||
      employee.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
      employee.jobTitle.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesStatus = statusFilter === 'all' || employee.status === statusFilter;
    const matchesDepartment = departmentFilter === 'all' || employee.department === departmentFilter;
    
    return matchesSearch && matchesStatus && matchesDepartment;
  });

  // 获取唯一的部门列表
  const departments = Array.from(new Set(employees.map(emp => emp.department)));

  // 处理行选择
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedEmployees(filteredEmployees.map(emp => emp.id));
    } else {
      setSelectedEmployees([]);
    }
  };

  const handleSelectRow = (employeeId: string, checked: boolean) => {
    if (checked) {
      setSelectedEmployees([...selectedEmployees, employeeId]);
    } else {
      setSelectedEmployees(selectedEmployees.filter(id => id !== employeeId));
    }
  };

  // 状态徽章样式
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'active':
        return <Badge variant="success">在职</Badge>;
      case 'inactive':
        return <Badge variant="error">离职</Badge>;
      case 'pending':
        return <Badge variant="warning">待入职</Badge>;
      default:
        return <Badge variant="default">{status}</Badge>;
    }
  };

  // 薪资格式化
  const formatSalary = (salary: number) => {
    return salary.toLocaleString('zh-CN', {
      style: 'currency',
      currency: 'CNY',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
  };

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
      setSelectedEmployees(selectedEmployees.filter(id => id !== employeeId));
    }
  };

  return (
    <div className="w-full p-6 space-y-4">
      {/* 顶部工具栏 */}
      <div className="flex flex-col sm:flex-row gap-4 justify-between items-start sm:items-center">
        <div className="flex flex-col sm:flex-row gap-4 flex-1">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="搜索员工..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 w-64"
            />
          </div>
          
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="筛选状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">所有状态</SelectItem>
              <SelectItem value="active">在职</SelectItem>
              <SelectItem value="inactive">离职</SelectItem>
              <SelectItem value="pending">待入职</SelectItem>
            </SelectContent>
          </Select>

          <Select value={departmentFilter} onValueChange={setDepartmentFilter}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="筛选部门" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">所有部门</SelectItem>
              {departments.map(dept => (
                <SelectItem key={dept} value={dept}>{dept}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => exportToCSV(selectedEmployees.length > 0 
              ? employees.filter(emp => selectedEmployees.includes(emp.id))
              : filteredEmployees
            )}
          >
            <Download className="h-4 w-4 mr-2" />
            导出数据
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={importFromCSV}
          >
            <Upload className="h-4 w-4 mr-2" />
            导入数据
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={downloadTemplate}
          >
            <Download className="h-4 w-4 mr-2" />
            下载模板
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => printData(selectedEmployees.length > 0 
              ? employees.filter(emp => selectedEmployees.includes(emp.id))
              : filteredEmployees
            )}
          >
            <Printer className="h-4 w-4 mr-2" />
            打印
          </Button>
        </div>
      </div>

      {/* 选中信息 */}
      {selectedEmployees.length > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-blue-700">
              已选中 {selectedEmployees.length} 个员工
            </span>
            <div className="flex gap-2">
              <Button size="sm" variant="destructive">
                批量删除
              </Button>
              <Button size="sm" variant="outline">
                批量导出
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* 数据表格 */}
      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={selectedEmployees.length === filteredEmployees.length && filteredEmployees.length > 0}
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              <TableHead>员工信息</TableHead>
              <TableHead>职位</TableHead>
              <TableHead>部门</TableHead>
              <TableHead>薪资</TableHead>
              <TableHead>入职日期</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="w-20">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredEmployees.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-gray-500">
                  没有找到匹配的员工数据
                </TableCell>
              </TableRow>
            ) : (
              filteredEmployees.map((employee) => (
                <TableRow key={employee.id}>
                  <TableCell>
                    <Checkbox
                      checked={selectedEmployees.includes(employee.id)}
                      onCheckedChange={(checked) => handleSelectRow(employee.id, checked as boolean)}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-3">
                      <Avatar>
                        <AvatarFallback>
                          <User className="h-4 w-4" />
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium">{employee.firstName} {employee.lastName}</div>
                        <div className="text-sm text-gray-500">{employee.email}</div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{employee.jobTitle}</TableCell>
                  <TableCell>{employee.department}</TableCell>
                  <TableCell>
                    <span className={`px-2 py-1 rounded text-sm font-medium ${
                      employee.salary < 10000 ? 'bg-red-100 text-red-800' :
                      employee.salary < 15000 ? 'bg-yellow-100 text-yellow-800' :
                      'bg-green-100 text-green-800'
                    }`}>
                      {formatSalary(employee.salary)}
                    </span>
                  </TableCell>
                  <TableCell>{new Date(employee.startDate).toLocaleDateString('zh-CN')}</TableCell>
                  <TableCell>{getStatusBadge(employee.status)}</TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <DropdownMenuItem onClick={() => handleEditEmployee(employee)}>
                          <Edit className="h-4 w-4 mr-2" />
                          编辑
                        </DropdownMenuItem>
                        <DropdownMenuItem 
                          onClick={() => handleDeleteEmployee(employee.id)}
                          className="text-red-600"
                        >
                          <Trash2 className="h-4 w-4 mr-2" />
                          删除
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* 编辑员工对话框 */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>编辑员工信息</DialogTitle>
          </DialogHeader>
          {editingEmployee && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="firstName">名字</Label>
                <Input
                  id="firstName"
                  value={editingEmployee.firstName}
                  onChange={(e) => setEditingEmployee({...editingEmployee, firstName: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="lastName">姓氏</Label>
                <Input
                  id="lastName"
                  value={editingEmployee.lastName}
                  onChange={(e) => setEditingEmployee({...editingEmployee, lastName: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">邮箱</Label>
                <Input
                  id="email"
                  type="email"
                  value={editingEmployee.email}
                  onChange={(e) => setEditingEmployee({...editingEmployee, email: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="jobTitle">职位</Label>
                <Input
                  id="jobTitle"
                  value={editingEmployee.jobTitle}
                  onChange={(e) => setEditingEmployee({...editingEmployee, jobTitle: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="department">部门</Label>
                <Input
                  id="department"
                  value={editingEmployee.department}
                  onChange={(e) => setEditingEmployee({...editingEmployee, department: e.target.value})}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="salary">薪资</Label>
                <Input
                  id="salary"
                  type="number"
                  value={editingEmployee.salary}
                  onChange={(e) => setEditingEmployee({...editingEmployee, salary: parseInt(e.target.value) || 0})}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="status">状态</Label>
                <Select value={editingEmployee.status} onValueChange={(value) => setEditingEmployee({...editingEmployee, status: value as any})}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">在职</SelectItem>
                    <SelectItem value="inactive">离职</SelectItem>
                    <SelectItem value="pending">待入职</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex justify-end space-x-2 pt-4">
                <Button variant="outline" onClick={() => setIsEditDialogOpen(false)}>
                  取消
                </Button>
                <Button onClick={handleSaveEmployee}>
                  保存
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default EmployeeManagementTable;