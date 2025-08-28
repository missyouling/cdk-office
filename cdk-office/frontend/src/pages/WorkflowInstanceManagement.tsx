'use client';

import React, { useState, useEffect } from 'react';
import {
  Eye,
  Square,
  Play,
  RefreshCw,
  CheckCircle2,
  X,
  Clock,
  AlertTriangle,
  MessageSquare,
  User,
  CalendarClock,
  Activity,
  Filter,
  Search,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Progress } from '@/components/ui/progress';
import { Skeleton } from '@/components/ui/skeleton';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { toast } from '@/components/ui/use-toast';

// 工作流实例接口
interface WorkflowInstance {
  id: string;
  definitionId: string;
  definitionName: string;
  status: 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
  currentStep: string;
  progress: number;
  startedAt: string;
  completedAt?: string;
  startedBy: string;
  assignedTo?: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  input: any;
  output?: any;
  tasks: WorkflowTask[];
  history: WorkflowHistory[];
}

interface WorkflowTask {
  id: string;
  instanceId: string;
  stepId: string;
  stepName: string;
  assigneeId: string;
  assigneeName: string;
  status: 'pending' | 'in_progress' | 'completed' | 'rejected';
  dueDate?: string;
  createdAt: string;
  completedAt?: string;
  comments?: string;
}

interface WorkflowHistory {
  id: string;
  instanceId: string;
  action: string;
  actor: string;
  timestamp: string;
  details: string;
  stepName?: string;
}

const WorkflowInstanceManagement: React.FC = () => {
  const [instances, setInstances] = useState<WorkflowInstance[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedInstance, setSelectedInstance] = useState<WorkflowInstance | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [tabValue, setTabValue] = useState(0);
  const [statusFilter, setStatusFilter] = useState('all');

  // 获取工作流实例列表
  useEffect(() => {
    fetchInstances();
  }, [statusFilter]);

  const fetchInstances = async () => {
    try {
      setLoading(true);
      const url = statusFilter === 'all' 
        ? '/api/v1/workflow/instances'
        : `/api/v1/workflow/instances?status=${statusFilter}`;
      
      const response = await fetch(url);
      if (response.ok) {
        const data = await response.json();
        setInstances(data.data || []);
      } else {
        throw new Error('获取工作流实例失败');
      }
    } catch (error) {
      console.error('获取工作流实例失败:', error);
      // 使用模拟数据
      const mockData: WorkflowInstance[] = [
        {
          id: 'inst1',
          definitionId: 'def1',
          definitionName: '文档审批流程',
          status: 'running',
          currentStep: '部门负责人审批',
          progress: 50,
          startedAt: '2023-12-01T10:30:00Z',
          startedBy: '张三',
          assignedTo: '李四',
          priority: 'normal',
          input: {
            documentId: 'doc123',
            documentName: 'Q4财务报告.pdf',
            requestType: 'upload'
          },
          tasks: [
            {
              id: 'task1',
              instanceId: 'inst1',
              stepId: 'step1',
              stepName: '部门负责人审批',
              assigneeId: 'user2',
              assigneeName: '李四',
              status: 'pending',
              dueDate: '2023-12-05T18:00:00Z',
              createdAt: '2023-12-01T10:30:00Z',
            }
          ],
          history: [
            {
              id: 'hist1',
              instanceId: 'inst1',
              action: 'started',
              actor: '张三',
              timestamp: '2023-12-01T10:30:00Z',
              details: '启动文档审批流程',
              stepName: '提交申请',
            },
            {
              id: 'hist2',
              instanceId: 'inst1',
              action: 'assigned',
              actor: '系统',
              timestamp: '2023-12-01T10:31:00Z',
              details: '分配给李四进行审批',
              stepName: '部门负责人审批',
            }
          ]
        },
        {
          id: 'inst2',
          definitionId: 'def2',
          definitionName: '员工入职流程',
          status: 'completed',
          currentStep: '已完成',
          progress: 100,
          startedAt: '2023-11-28T09:00:00Z',
          completedAt: '2023-11-30T16:30:00Z',
          startedBy: 'HR',
          priority: 'high',
          input: {
            employeeName: '王五',
            department: '技术部',
            position: '前端工程师'
          },
          output: {
            accountCreated: true,
            equipmentAssigned: true,
            orientationCompleted: true
          },
          tasks: [],
          history: [
            {
              id: 'hist3',
              instanceId: 'inst2',
              action: 'started',
              actor: 'HR',
              timestamp: '2023-11-28T09:00:00Z',
              details: '启动员工入职流程',
            },
            {
              id: 'hist4',
              instanceId: 'inst2',
              action: 'completed',
              actor: '系统',
              timestamp: '2023-11-30T16:30:00Z',
              details: '入职流程已完成',
            }
          ]
        }
      ];
      setInstances(mockData);
    } finally {
      setLoading(false);
    }
  };

  // 查看实例详情
  const handleViewInstance = (instance: WorkflowInstance) => {
    setSelectedInstance(instance);
    setTabValue(0);
    setOpenDialog(true);
  };

  // 停止实例
  const handleStopInstance = async (instance: WorkflowInstance) => {
    try {
      const response = await fetch(`/api/v1/workflow/instances/${instance.id}/stop`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ reason: '手动停止' }),
      });

      if (response.ok) {
        setInstances(prev => prev.map(inst => 
          inst.id === instance.id ? { ...inst, status: 'cancelled' } : inst
        ));
        toast({
          title: "成功",
          description: "工作流实例已停止",
        });
      } else {
        throw new Error('停止失败');
      }
    } catch (error) {
      console.error('停止工作流实例失败:', error);
      toast({
        title: "错误",
        description: "停止失败，请重试",
        variant: "destructive",
      });
    }
  };

  // 重启实例
  const handleRestartInstance = async (instance: WorkflowInstance) => {
    try {
      const response = await fetch(`/api/v1/workflow/instances/${instance.id}/restart`, {
        method: 'POST',
      });

      if (response.ok) {
        setInstances(prev => prev.map(inst => 
          inst.id === instance.id ? { ...inst, status: 'running' } : inst
        ));
        toast({
          title: "成功",
          description: "工作流实例已重启",
        });
      } else {
        throw new Error('重启失败');
      }
    } catch (error) {
      console.error('重启工作流实例失败:', error);
      toast({
        title: "错误",
        description: "重启失败，请重试",
        variant: "destructive",
      });
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'primary';
      case 'completed': return 'success';
      case 'failed': return 'error';
      case 'cancelled': return 'default';
      case 'paused': return 'warning';
      default: return 'default';
    }
  };

  // 获取状态标签
  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'running': return '运行中';
      case 'completed': return '已完成';
      case 'failed': return '失败';
      case 'cancelled': return '已取消';
      case 'paused': return '暂停';
      default: return status;
    }
  };

  // 获取优先级颜色
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'low': return 'success';
      case 'normal': return 'info';
      case 'high': return 'warning';
      case 'urgent': return 'error';
      default: return 'default';
    }
  };

  // 获取优先级标签
  const getPriorityLabel = (priority: string) => {
    switch (priority) {
      case 'low': return '低';
      case 'normal': return '中';
      case 'high': return '高';
      case 'urgent': return '紧急';
      default: return priority;
    }
  };

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  // 计算持续时间
  const getDuration = (startDate: string, endDate?: string) => {
    const start = new Date(startDate);
    const end = endDate ? new Date(endDate) : new Date();
    const diff = end.getTime() - start.getTime();
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    return `${hours}小时${minutes}分钟`;
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="space-y-6">
          <Skeleton className="h-8 w-48" />
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <div className="flex justify-between">
                    <Skeleton className="h-6 w-1/3" />
                    <Skeleton className="h-6 w-16" />
                  </div>
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-2 w-full" />
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
      {/* 页面标题和过滤器 */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold">工作流实例管理</h1>
          <p className="text-muted-foreground mt-1">监控和管理正在运行的工作流实例</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4" />
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="状态筛选" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部</SelectItem>
                <SelectItem value="running">运行中</SelectItem>
                <SelectItem value="completed">已完成</SelectItem>
                <SelectItem value="failed">失败</SelectItem>
                <SelectItem value="cancelled">已取消</SelectItem>
                <SelectItem value="paused">暂停</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="outline" size="icon" onClick={fetchInstances}>
                  <RefreshCw className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>刷新</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      {/* 实例列表 */}
      {instances.length === 0 ? (
        <Card className="flex flex-col items-center justify-center py-16">
          <Activity className="h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-semibold">暂无工作流实例</h3>
          <p className="text-muted-foreground text-center max-w-sm mt-2">
            当前没有正在运行或历史的工作流实例
          </p>
        </Card>
      ) : (
        <div className="space-y-4">
          {instances.map((instance) => (
            <Card key={instance.id} className="hover:shadow-md transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex justify-between items-start">
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <CardTitle className="text-lg">{instance.definitionName}</CardTitle>
                      <Badge 
                        variant={instance.status === 'running' ? 'default' : 
                                instance.status === 'completed' ? 'secondary' : 
                                instance.status === 'failed' ? 'destructive' : 'outline'}
                      >
                        {getStatusLabel(instance.status)}
                      </Badge>
                      <Badge 
                        variant={instance.priority === 'urgent' ? 'destructive' : 
                                instance.priority === 'high' ? 'secondary' : 'outline'}
                      >
                        {getPriorityLabel(instance.priority)}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      ID: {instance.id} • 当前步骤: {instance.currentStep}
                    </p>
                  </div>
                  <div className="flex gap-1">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button 
                            variant="ghost" 
                            size="sm" 
                            onClick={() => handleViewInstance(instance)}
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>查看详情</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                    
                    {instance.status === 'running' && (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="text-red-600"
                              onClick={() => handleStopInstance(instance)}
                            >
                              <Square className="h-4 w-4" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>停止</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    )}
                    
                    {(instance.status === 'failed' || instance.status === 'cancelled') && (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              className="text-green-600"
                              onClick={() => handleRestartInstance(instance)}
                            >
                              <Play className="h-4 w-4" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>重启</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    )}
                  </div>
                </div>
              </CardHeader>
              
              <CardContent className="pt-0">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
                  <div className="flex items-center gap-2 text-sm">
                    <User className="h-4 w-4 text-muted-foreground" />
                    <span>发起人: {instance.startedBy}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <CalendarClock className="h-4 w-4 text-muted-foreground" />
                    <span>开始: {formatDate(instance.startedAt)}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span>持续: {getDuration(instance.startedAt, instance.completedAt)}</span>
                  </div>
                  {instance.assignedTo && (
                    <div className="flex items-center gap-2 text-sm">
                      <User className="h-4 w-4 text-muted-foreground" />
                      <span>当前处理: {instance.assignedTo}</span>
                    </div>
                  )}
                </div>
                
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">进度:</span>
                  <Progress value={instance.progress} className="flex-1" />
                  <span className="text-sm font-medium">{instance.progress}%</span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 实例详情对话框 */}
      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="max-w-6xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              工作流实例详情 - {selectedInstance?.definitionName}
            </DialogTitle>
            <DialogDescription>
              查看工作流实例的详细信息、任务和执行历史
            </DialogDescription>
          </DialogHeader>
          
          <div className="py-4">
            <Tabs value={tabValue.toString()} onValueChange={(value) => setTabValue(parseInt(value))}>
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="0">基本信息</TabsTrigger>
                <TabsTrigger value="1">任务列表</TabsTrigger>
                <TabsTrigger value="2">执行历史</TabsTrigger>
                <TabsTrigger value="3">输入输出</TabsTrigger>
              </TabsList>

              {/* 基本信息 */}
              <TabsContent value="0" className="space-y-4 mt-4">
                {selectedInstance && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg">实例信息</CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">实例ID</Label>
                          <p className="text-sm mt-1">{selectedInstance.id}</p>
                        </div>
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">工作流名称</Label>
                          <p className="text-sm mt-1">{selectedInstance.definitionName}</p>
                        </div>
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">状态</Label>
                          <div className="mt-1">
                            <Badge 
                              variant={selectedInstance.status === 'running' ? 'default' : 
                                      selectedInstance.status === 'completed' ? 'secondary' : 
                                      selectedInstance.status === 'failed' ? 'destructive' : 'outline'}
                            >
                              {getStatusLabel(selectedInstance.status)}
                            </Badge>
                          </div>
                        </div>
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">当前步骤</Label>
                          <p className="text-sm mt-1">{selectedInstance.currentStep}</p>
                        </div>
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">进度</Label>
                          <div className="mt-2 flex items-center gap-2">
                            <Progress value={selectedInstance.progress} className="flex-1" />
                            <span className="text-sm font-medium">{selectedInstance.progress}%</span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                    
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-lg">时间信息</CardTitle>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">发起人</Label>
                          <p className="text-sm mt-1">{selectedInstance.startedBy}</p>
                        </div>
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">开始时间</Label>
                          <p className="text-sm mt-1">{formatDate(selectedInstance.startedAt)}</p>
                        </div>
                        {selectedInstance.completedAt && (
                          <div>
                            <Label className="text-sm font-medium text-muted-foreground">完成时间</Label>
                            <p className="text-sm mt-1">{formatDate(selectedInstance.completedAt)}</p>
                          </div>
                        )}
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">持续时间</Label>
                          <p className="text-sm mt-1">
                            {getDuration(selectedInstance.startedAt, selectedInstance.completedAt)}
                          </p>
                        </div>
                        <div>
                          <Label className="text-sm font-medium text-muted-foreground">优先级</Label>
                          <div className="mt-1">
                            <Badge 
                              variant={selectedInstance.priority === 'urgent' ? 'destructive' : 
                                      selectedInstance.priority === 'high' ? 'secondary' : 'outline'}
                            >
                              {getPriorityLabel(selectedInstance.priority)}
                            </Badge>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                )}
              </TabsContent>

              {/* 任务列表 */}
              <TabsContent value="1" className="mt-4">
                {selectedInstance?.tasks.length === 0 ? (
                  <div className="text-center py-8">
                    <Clock className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-semibold">暂无任务</h3>
                    <p className="text-muted-foreground">该实例当前没有活跃任务</p>
                  </div>
                ) : (
                  <div className="border rounded-lg">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>任务ID</TableHead>
                          <TableHead>步骤名称</TableHead>
                          <TableHead>处理人</TableHead>
                          <TableHead>状态</TableHead>
                          <TableHead>截止时间</TableHead>
                          <TableHead>创建时间</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {selectedInstance?.tasks.map((task) => (
                          <TableRow key={task.id}>
                            <TableCell className="font-mono text-xs">{task.id}</TableCell>
                            <TableCell>{task.stepName}</TableCell>
                            <TableCell>{task.assigneeName}</TableCell>
                            <TableCell>
                              <div className="flex items-center gap-2">
                                {task.status === 'pending' && <Clock className="h-4 w-4 text-orange-500" />}
                                {task.status === 'in_progress' && <Activity className="h-4 w-4 text-blue-500" />}
                                {task.status === 'completed' && <CheckCircle2 className="h-4 w-4 text-green-500" />}
                                {task.status === 'rejected' && <X className="h-4 w-4 text-red-500" />}
                                <span className="text-sm">
                                  {task.status === 'pending' ? '待处理' :
                                   task.status === 'in_progress' ? '处理中' :
                                   task.status === 'completed' ? '已完成' : '已拒绝'}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell>
                              {task.dueDate ? formatDate(task.dueDate) : '-'}
                            </TableCell>
                            <TableCell>{formatDate(task.createdAt)}</TableCell>
                          </TableRow>
                        )) || []}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </TabsContent>

              {/* 执行历史 */}
              <TabsContent value="2" className="mt-4">
                {selectedInstance?.history.length === 0 ? (
                  <div className="text-center py-8">
                    <MessageSquare className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                    <h3 className="text-lg font-semibold">暂无执行历史</h3>
                    <p className="text-muted-foreground">该实例尚无执行记录</p>
                  </div>
                ) : (
                  <ScrollArea className="h-96">
                    <div className="space-y-4">
                      {selectedInstance?.history.map((hist, index) => (
                        <div key={hist.id} className="flex gap-4">
                          <div className="flex flex-col items-center">
                            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground">
                              {hist.action === 'started' && <Play className="h-4 w-4" />}
                              {hist.action === 'completed' && <CheckCircle2 className="h-4 w-4" />}
                              {hist.action === 'failed' && <AlertTriangle className="h-4 w-4" />}
                              {hist.action === 'assigned' && <User className="h-4 w-4" />}
                              {!['started', 'completed', 'failed', 'assigned'].includes(hist.action) && <Clock className="h-4 w-4" />}
                            </div>
                            {index < (selectedInstance?.history.length || 0) - 1 && (
                              <div className="h-8 w-px bg-border" />
                            )}
                          </div>
                          <div className="flex-1 space-y-1">
                            <div className="flex items-center justify-between">
                              <h4 className="text-sm font-medium">
                                {hist.stepName || hist.action}
                              </h4>
                              <span className="text-xs text-muted-foreground">
                                {formatDate(hist.timestamp)}
                              </span>
                            </div>
                            <p className="text-sm text-muted-foreground">
                              {hist.details} - {hist.actor}
                            </p>
                          </div>
                        </div>
                      )) || []}
                    </div>
                  </ScrollArea>
                )}
              </TabsContent>

              {/* 输入输出 */}
              <TabsContent value="3" className="mt-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">输入数据</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <ScrollArea className="h-64">
                        <pre className="text-xs bg-muted p-3 rounded border overflow-auto">
                          {JSON.stringify(selectedInstance?.input, null, 2)}
                        </pre>
                      </ScrollArea>
                    </CardContent>
                  </Card>
                  
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-lg">输出数据</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <ScrollArea className="h-64">
                        <pre className="text-xs bg-muted p-3 rounded border overflow-auto">
                          {selectedInstance?.output 
                            ? JSON.stringify(selectedInstance.output, null, 2)
                            : '暂无输出数据'}
                        </pre>
                      </ScrollArea>
                    </CardContent>
                  </Card>
                </div>
              </TabsContent>
            </Tabs>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default WorkflowInstanceManagement;