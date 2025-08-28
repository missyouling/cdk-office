'use client';

import React, { useState, useEffect } from 'react';
import {
  Plus,
  Edit,
  Trash2,
  Eye,
  Play,
  Square,
  RefreshCw,
  Settings,
  Users,
  Clock,
  AlertCircle,
  CheckCircle2,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
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
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Switch } from '@/components/ui/switch';
import { Skeleton } from '@/components/ui/skeleton';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { toast } from '@/components/ui/use-toast';

// 工作流定义接口
interface WorkflowDefinition {
  id: string;
  name: string;
  description: string;
  version: string;
  category: string;
  status: 'active' | 'inactive' | 'draft';
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  steps: WorkflowStep[];
  config: WorkflowConfig;
  statistics: {
    totalInstances: number;
    activeInstances: number;
    completedInstances: number;
    successRate: number;
  };
}

interface WorkflowStep {
  id: string;
  name: string;
  type: 'approval' | 'notification' | 'condition' | 'action';
  assignee: string;
  assigneeType: 'user' | 'role' | 'group';
  config: any;
  order: number;
}

interface WorkflowConfig {
  timeout: number;
  retryAttempts: number;
  notificationEnabled: boolean;
  escalationEnabled: boolean;
  escalationTimeout: number;
}

const WorkflowManagement: React.FC = () => {
  const [workflows, setWorkflows] = useState<WorkflowDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedWorkflow, setSelectedWorkflow] = useState<WorkflowDefinition | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [dialogType, setDialogType] = useState<'create' | 'edit' | 'view'>('view');
  const [tabValue, setTabValue] = useState(0);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: '',
    status: 'draft' as const,
    config: {
      timeout: 86400,
      retryAttempts: 3,
      notificationEnabled: true,
      escalationEnabled: false,
      escalationTimeout: 7200,
    },
    steps: [] as WorkflowStep[],
  });

  // 获取工作流列表
  useEffect(() => {
    fetchWorkflows();
  }, []);

  const fetchWorkflows = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/workflow/definitions');
      if (response.ok) {
        const data = await response.json();
        setWorkflows(data.data || []);
      } else {
        throw new Error('获取工作流列表失败');
      }
    } catch (error) {
      console.error('获取工作流列表失败:', error);
      // 使用模拟数据
      const mockData: WorkflowDefinition[] = [
        {
          id: '1',
          name: '文档审批流程',
          description: '用于文档上传、修改、删除的审批流程',
          version: '1.0.0',
          category: '文档管理',
          status: 'active',
          createdBy: '管理员',
          createdAt: '2023-12-01T10:00:00Z',
          updatedAt: '2023-12-01T10:00:00Z',
          steps: [
            {
              id: 'step1',
              name: '部门负责人审批',
              type: 'approval',
              assignee: 'dept_manager',
              assigneeType: 'role',
              config: { timeout: 86400 },
              order: 1,
            },
            {
              id: 'step2',
              name: '总经理审批',
              type: 'approval',
              assignee: 'general_manager',
              assigneeType: 'role',
              config: { timeout: 172800 },
              order: 2,
            },
          ],
          config: {
            timeout: 259200,
            retryAttempts: 3,
            notificationEnabled: true,
            escalationEnabled: true,
            escalationTimeout: 7200,
          },
          statistics: {
            totalInstances: 45,
            activeInstances: 8,
            completedInstances: 37,
            successRate: 92.3,
          },
        },
        {
          id: '2',
          name: '员工入职流程',
          description: '新员工入职的完整审批和准备流程',
          version: '2.1.0',
          category: '人事管理',
          status: 'active',
          createdBy: 'HR',
          createdAt: '2023-11-15T09:00:00Z',
          updatedAt: '2023-11-28T14:30:00Z',
          steps: [
            {
              id: 'step1',
              name: 'HR初审',
              type: 'approval',
              assignee: 'hr_specialist',
              assigneeType: 'role',
              config: { timeout: 43200 },
              order: 1,
            },
            {
              id: 'step2',
              name: '部门经理确认',
              type: 'approval',
              assignee: 'dept_manager',
              assigneeType: 'role',
              config: { timeout: 86400 },
              order: 2,
            },
          ],
          config: {
            timeout: 172800,
            retryAttempts: 2,
            notificationEnabled: true,
            escalationEnabled: false,
            escalationTimeout: 3600,
          },
          statistics: {
            totalInstances: 23,
            activeInstances: 3,
            completedInstances: 20,
            successRate: 95.7,
          },
        },
      ];
      setWorkflows(mockData);
    } finally {
      setLoading(false);
    }
  };

  // 打开创建对话框
  const handleCreate = () => {
    setFormData({
      name: '',
      description: '',
      category: '',
      status: 'draft',
      config: {
        timeout: 86400,
        retryAttempts: 3,
        notificationEnabled: true,
        escalationEnabled: false,
        escalationTimeout: 7200,
      },
      steps: [],
    });
    setDialogType('create');
    setTabValue(0);
    setOpenDialog(true);
  };

  // 打开编辑对话框
  const handleEdit = (workflow: WorkflowDefinition) => {
    setSelectedWorkflow(workflow);
    setFormData({
      name: workflow.name,
      description: workflow.description,
      category: workflow.category,
      status: workflow.status,
      config: workflow.config,
      steps: workflow.steps,
    });
    setDialogType('edit');
    setTabValue(0);
    setOpenDialog(true);
  };

  // 打开查看对话框
  const handleView = (workflow: WorkflowDefinition) => {
    setSelectedWorkflow(workflow);
    setDialogType('view');
    setTabValue(0);
    setOpenDialog(true);
  };

  // 删除工作流
  const handleDelete = async (workflow: WorkflowDefinition) => {
    try {
      const response = await fetch(`/api/v1/workflow/definitions/${workflow.id}`, {
        method: 'DELETE',
      });
      
      if (response.ok) {
        setWorkflows(prev => prev.filter(w => w.id !== workflow.id));
        toast({
          title: "成功",
          description: "工作流删除成功",
        });
      } else {
        throw new Error('删除失败');
      }
    } catch (error) {
      console.error('删除工作流失败:', error);
      toast({
        title: "错误",
        description: "删除工作流失败，请重试",
        variant: "destructive",
      });
    }
  };

  // 启动/停止工作流
  const handleToggleStatus = async (workflow: WorkflowDefinition) => {
    try {
      const newStatus = workflow.status === 'active' ? 'inactive' : 'active';
      const response = await fetch(`/api/v1/workflow/definitions/${workflow.id}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });

      if (response.ok) {
        setWorkflows(prev => prev.map(w => 
          w.id === workflow.id ? { ...w, status: newStatus } : w
        ));
        toast({
          title: "成功",
          description: `工作流已${newStatus === 'active' ? '启用' : '停用'}`,
        });
      } else {
        throw new Error('状态更新失败');
      }
    } catch (error) {
      console.error('更新工作流状态失败:', error);
      toast({
        title: "错误",
        description: "更新状态失败，请重试",
        variant: "destructive",
      });
    }
  };

  // 保存工作流
  const handleSave = async () => {
    try {
      const url = dialogType === 'create' 
        ? '/api/v1/workflow/definitions'
        : `/api/v1/workflow/definitions/${selectedWorkflow?.id}`;
      
      const method = dialogType === 'create' ? 'POST' : 'PUT';
      
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        const result = await response.json();
        
        if (dialogType === 'create') {
          setWorkflows(prev => [...prev, result.data]);
        } else {
          setWorkflows(prev => prev.map(w => 
            w.id === selectedWorkflow?.id ? result.data : w
          ));
        }
        
        setOpenDialog(false);
        toast({
          title: "成功",
          description: `工作流${dialogType === 'create' ? '创建' : '更新'}成功`,
        });
      } else {
        throw new Error('保存失败');
      }
    } catch (error) {
      console.error('保存工作流失败:', error);
      toast({
        title: "错误",
        description: "保存失败，请重试",
        variant: "destructive",
      });
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success';
      case 'inactive': return 'default';
      case 'draft': return 'warning';
      default: return 'default';
    }
  };

  // 获取状态标签
  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'active': return '活跃';
      case 'inactive': return '停用';
      case 'draft': return '草稿';
      default: return status;
    }
  };

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="space-y-6">
          <Skeleton className="h-8 w-48" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <Skeleton className="h-6 w-3/4" />
                  <Skeleton className="h-4 w-1/2" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-16 w-full" />
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold">工作流管理</h1>
          <p className="text-muted-foreground mt-1">管理和配置工作流定义</p>
        </div>
        <div className="flex gap-2">
          <Button onClick={handleCreate} className="gap-2">
            <Plus className="h-4 w-4" />
            创建工作流
          </Button>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="icon" onClick={fetchWorkflows}>
                  <RefreshCw className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>刷新</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      {/* 工作流列表 */}
      {workflows.length === 0 ? (
        <Card className="flex flex-col items-center justify-center py-12">
          <AlertCircle className="h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-semibold">暂无工作流定义</h3>
          <p className="text-muted-foreground text-center max-w-sm mt-2">
            开始创建你的第一个工作流来自动化业务流程
          </p>
          <Button onClick={handleCreate} className="mt-4 gap-2">
            <Plus className="h-4 w-4" />
            创建工作流
          </Button>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {workflows.map((workflow) => (
            <Card key={workflow.id} className="group hover:shadow-lg transition-shadow">
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="space-y-1">
                    <CardTitle className="text-lg">{workflow.name}</CardTitle>
                    <Badge 
                      variant={workflow.status === 'active' ? 'default' : 
                              workflow.status === 'draft' ? 'secondary' : 'outline'}
                    >
                      {getStatusLabel(workflow.status)}
                    </Badge>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    v{workflow.version}
                  </div>
                </div>
              </CardHeader>
              
              <CardContent className="space-y-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">{workflow.category}</p>
                  <p className="text-sm mt-1">{workflow.description}</p>
                </div>

                {/* 统计信息 */}
                <div className="bg-muted/50 rounded-lg p-3 space-y-2">
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3" />
                      总实例: {workflow.statistics.totalInstances}
                    </div>
                    <div className="flex items-center gap-1">
                      <Play className="h-3 w-3" />
                      活跃: {workflow.statistics.activeInstances}
                    </div>
                    <div className="flex items-center gap-1">
                      <CheckCircle2 className="h-3 w-3" />
                      完成: {workflow.statistics.completedInstances}
                    </div>
                    <div className="flex items-center gap-1">
                      <Settings className="h-3 w-3" />
                      成功率: {workflow.statistics.successRate}%
                    </div>
                  </div>
                </div>

                <div className="text-xs text-muted-foreground">
                  创建者: {workflow.createdBy} • 最后更新: {formatDate(workflow.updatedAt)}
                </div>
              </CardContent>
              
              <CardFooter className="flex justify-between">
                <div className="flex gap-1">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button variant="ghost" size="sm" onClick={() => handleView(workflow)}>
                          <Eye className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>查看详情</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button variant="ghost" size="sm" onClick={() => handleEdit(workflow)}>
                          <Edit className="h-4 w-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>编辑</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>

                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button 
                          variant="ghost" 
                          size="sm" 
                          onClick={() => handleToggleStatus(workflow)}
                          className={workflow.status === 'active' ? 'text-orange-600' : 'text-green-600'}
                        >
                          {workflow.status === 'active' ? 
                            <Square className="h-4 w-4" /> : 
                            <Play className="h-4 w-4" />
                          }
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {workflow.status === 'active' ? '停用' : '启用'}
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>

                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>确认删除</AlertDialogTitle>
                      <AlertDialogDescription>
                        你确定要删除工作流“{workflow.name}”吗？这个操作不可撤销。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction 
                        onClick={() => handleDelete(workflow)}
                        className="bg-red-600 hover:bg-red-700"
                      >
                        删除
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </CardFooter>
            </Card>
          ))}
        </div>
      )}

      {/* 工作流详情/编辑对话框 */}
      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {dialogType === 'create' ? '创建工作流' : 
               dialogType === 'edit' ? '编辑工作流' : '工作流详情'}
            </DialogTitle>
            <DialogDescription>
              {dialogType === 'create' ? '创建新的工作流定义' :
               dialogType === 'edit' ? '修改工作流配置' : '查看工作流详细信息'}
            </DialogDescription>
          </DialogHeader>
          
          <div className="py-4">
            <Tabs value={tabValue.toString()} onValueChange={(value) => setTabValue(parseInt(value))}>
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="0">基本信息</TabsTrigger>
                <TabsTrigger value="1">工作流步骤</TabsTrigger>
                <TabsTrigger value="2">配置设置</TabsTrigger>
              </TabsList>

              {/* 基本信息标签页 */}
              <TabsContent value="0" className="space-y-4 mt-4">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="workflow-name">工作流名称</Label>
                    <Input
                      id="workflow-name"
                      value={dialogType === 'view' ? selectedWorkflow?.name || '' : formData.name}
                      onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                      disabled={dialogType === 'view'}
                      placeholder="输入工作流名称"
                    />
                  </div>
                  
                  <div>
                    <Label htmlFor="workflow-description">描述</Label>
                    <Textarea
                      id="workflow-description"
                      rows={3}
                      value={dialogType === 'view' ? selectedWorkflow?.description || '' : formData.description}
                      onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                      disabled={dialogType === 'view'}
                      placeholder="输入工作流描述"
                    />
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="workflow-category">分类</Label>
                      <Input
                        id="workflow-category"
                        value={dialogType === 'view' ? selectedWorkflow?.category || '' : formData.category}
                        onChange={(e) => setFormData(prev => ({ ...prev, category: e.target.value }))}
                        disabled={dialogType === 'view'}
                        placeholder="输入分类"
                      />
                    </div>
                    
                    <div>
                      <Label htmlFor="workflow-status">状态</Label>
                      <Select 
                        value={dialogType === 'view' ? selectedWorkflow?.status || '' : formData.status}
                        onValueChange={(value) => setFormData(prev => ({ ...prev, status: value as any }))}
                        disabled={dialogType === 'view'}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="选择状态" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="draft">草稿</SelectItem>
                          <SelectItem value="active">活跃</SelectItem>
                          <SelectItem value="inactive">停用</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                </div>
              </TabsContent>

              {/* 工作流步骤标签页 */}
              <TabsContent value="1" className="space-y-4 mt-4">
                <div className="space-y-4">
                  <h3 className="text-lg font-semibold">工作流步骤</h3>
                  <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                    <div className="flex items-center gap-2">
                      <AlertCircle className="h-5 w-5 text-blue-600" />
                      <p className="text-blue-800 font-medium">功能开发中</p>
                    </div>
                    <p className="text-blue-700 mt-2">
                      工作流步骤配置功能正在开发中，请使用API进行详细配置
                    </p>
                  </div>
                </div>
              </TabsContent>

              {/* 配置设置标签页 */}
              <TabsContent value="2" className="space-y-4 mt-4">
                <div className="space-y-6">
                  <h3 className="text-lg font-semibold">配置设置</h3>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label htmlFor="timeout">超时时间（秒）</Label>
                      <Input
                        id="timeout"
                        type="number"
                        value={dialogType === 'view' ? selectedWorkflow?.config.timeout || 0 : formData.config.timeout}
                        onChange={(e) => setFormData(prev => ({
                          ...prev,
                          config: { ...prev.config, timeout: parseInt(e.target.value) }
                        }))}
                        disabled={dialogType === 'view'}
                      />
                    </div>
                    
                    <div>
                      <Label htmlFor="retryAttempts">重试次数</Label>
                      <Input
                        id="retryAttempts"
                        type="number"
                        value={dialogType === 'view' ? selectedWorkflow?.config.retryAttempts || 0 : formData.config.retryAttempts}
                        onChange={(e) => setFormData(prev => ({
                          ...prev,
                          config: { ...prev.config, retryAttempts: parseInt(e.target.value) }
                        }))}
                        disabled={dialogType === 'view'}
                      />
                    </div>
                  </div>

                  <Separator />

                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label>启用通知</Label>
                        <p className="text-sm text-muted-foreground">
                          在工作流执行过程中发送通知
                        </p>
                      </div>
                      <Switch
                        checked={dialogType === 'view' ? selectedWorkflow?.config.notificationEnabled || false : formData.config.notificationEnabled}
                        onCheckedChange={(checked) => setFormData(prev => ({
                          ...prev,
                          config: { ...prev.config, notificationEnabled: checked }
                        }))}
                        disabled={dialogType === 'view'}
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      <div className="space-y-0.5">
                        <Label>启用升级</Label>
                        <p className="text-sm text-muted-foreground">
                          在超时时自动升级处理
                        </p>
                      </div>
                      <Switch
                        checked={dialogType === 'view' ? selectedWorkflow?.config.escalationEnabled || false : formData.config.escalationEnabled}
                        onCheckedChange={(checked) => setFormData(prev => ({
                          ...prev,
                          config: { ...prev.config, escalationEnabled: checked }
                        }))}
                        disabled={dialogType === 'view'}
                      />
                    </div>
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenDialog(false)}>
              取消
            </Button>
            {dialogType !== 'view' && (
              <Button onClick={handleSave}>
                {dialogType === 'create' ? '创建' : '保存'}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default WorkflowManagement;