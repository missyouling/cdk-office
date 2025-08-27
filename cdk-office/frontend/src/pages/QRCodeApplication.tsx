import React, { useState } from 'react';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  CardDescription,
  CardFooter
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
  QrCode, 
  Plus, 
  Eye, 
  Edit, 
  Trash2, 
  Download,
  Share2
} from 'lucide-react';

interface QRCodeForm {
  id: string;
  name: string;
  type: 'survey' | 'registration' | 'feedback' | 'checkin';
  createdAt: string;
  createdBy: string;
}

// 模拟二维码表单数据
const mockForms: QRCodeForm[] = [
  {
    id: '1',
    name: '员工签到表单',
    type: 'checkin',
    createdAt: '2023-10-15',
    createdBy: '张三',
  },
  {
    id: '2',
    name: '会议室预订表单',
    type: 'registration',
    createdAt: '2023-10-10',
    createdBy: '李四',
  },
  {
    id: '3',
    name: '员工满意度调查',
    type: 'survey',
    createdAt: '2023-10-05',
    createdBy: '王五',
  },
  {
    id: '4',
    name: '访客登记表单',
    type: 'registration',
    createdAt: '2023-09-28',
    createdBy: '赵六',
  },
];

const QRCodeApplication: React.FC = () => {
  const [forms] = useState<QRCodeForm[]>(mockForms);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedForm, setSelectedForm] = useState<QRCodeForm | null>(null);

  const getFormTypeLabel = (type: QRCodeForm['type']) => {
    switch (type) {
      case 'survey': return '调查问卷';
      case 'registration': return '登记表';
      case 'feedback': return '反馈表';
      case 'checkin': return '签到表';
      default: return '未知类型';
    }
  };

  const getFormTypeColor = (type: QRCodeForm['type']) => {
    switch (type) {
      case 'survey': return 'default';
      case 'registration': return 'secondary';
      case 'feedback': return 'success';
      case 'checkin': return 'warning';
      default: return 'default';
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">二维码应用</h1>
          <p className="text-muted-foreground mt-2">
            创建和管理二维码表单
          </p>
        </div>
        <Button onClick={() => setIsCreateDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          创建表单
        </Button>
      </div>

      {/* 表单卡片网格 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {forms.map((form) => (
          <Card key={form.id} className="flex flex-col">
            <CardHeader>
              <div className="flex items-center space-x-3">
                <Avatar>
                  <AvatarFallback>
                    <QrCode className="h-5 w-5" />
                  </AvatarFallback>
                </Avatar>
                <div>
                  <CardTitle className="text-lg">{form.name}</CardTitle>
                  <CardDescription>
                    {form.createdAt} • {form.createdBy}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-grow">
              <div className="flex justify-center my-4">
                <QrCode className="h-24 w-24 text-muted-foreground" />
              </div>
              <div className="flex justify-center">
                <Badge variant={getFormTypeColor(form.type)}>
                  {getFormTypeLabel(form.type)}
                </Badge>
              </div>
            </CardContent>
            <CardFooter className="flex justify-between">
              <Button variant="outline" size="sm">
                <Eye className="h-4 w-4 mr-2" />
                预览
              </Button>
              <div className="flex space-x-1">
                <Button variant="ghost" size="icon">
                  <Edit className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon">
                  <Share2 className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon">
                  <Download className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </CardFooter>
          </Card>
        ))}
      </div>

      {/* 创建表单对话框 */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>创建二维码表单</DialogTitle>
            <DialogDescription>
              输入表单信息来创建新的二维码表单
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="formName">表单名称</Label>
              <Input id="formName" placeholder="输入表单名称" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="formType">表单类型</Label>
              <Select>
                <SelectTrigger id="formType">
                  <SelectValue placeholder="选择表单类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="checkin">签到表</SelectItem>
                  <SelectItem value="registration">登记表</SelectItem>
                  <SelectItem value="survey">调查问卷</SelectItem>
                  <SelectItem value="feedback">反馈表</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="formDescription">表单描述</Label>
              <Input id="formDescription" placeholder="输入表单描述" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={() => {
              setIsCreateDialogOpen(false);
              toast({
                title: "表单创建成功",
                description: "新的二维码表单已创建",
              });
            }}>
              创建
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default QRCodeApplication;