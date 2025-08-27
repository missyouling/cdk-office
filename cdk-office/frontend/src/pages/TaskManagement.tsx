'use client';

import React, { useState, useEffect } from 'react';
import {
  CheckCircle,
  XCircle,
  MessageSquare,
  Eye,
  RotateCcw,
  Clock,
  User,
  FileText,
  AlertTriangle,
  Send,
  Paperclip,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { Progress } from '@/components/ui/progress';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { toast } from '@/components/ui/use-toast';
import { Label } from '@/components/ui/label';

// 任务接口
interface Task {
  id: string;
  instanceId: string;
  workflowName: string;
  stepName: string;
  title: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'rejected';
  priority: 'low' | 'normal' | 'high' | 'urgent';
  assigneeId: string;
  assigneeName: string;
  requestorId: string;
  requestorName: string;
  createdAt: string;
  dueDate?: string;
  completedAt?: string;
  data: any;
  comments: TaskComment[];
  attachments: TaskAttachment[];
}

interface TaskComment {
  id: string;
  taskId: string;
  userId: string;
  userName: string;
  content: string;
  createdAt: string;
}

interface TaskAttachment {
  id: string;
  taskId: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  uploadedBy: string;
  uploadedAt: string;
}

const TaskManagement: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [dialogType, setDialogType] = useState<'view' | 'process'>('view');
  const [tabValue, setTabValue] = useState('pending');
  const [comment, setComment] = useState('');
  const [processing, setProcessing] = useState(false);

  // 获取任务列表
  useEffect(() => {
    fetchTasks();
  }, [tabValue]);

  const fetchTasks = async () => {
    try {
      setLoading(true);
      const url = tabValue === 'all' 
        ? '/api/v1/workflow/tasks/assigned'
        : `/api/v1/workflow/tasks/assigned?status=${tabValue}`;
      
      const response = await fetch(url);
      if (response.ok) {
        const data = await response.json();
        setTasks(data.data || []);
      } else {
        throw new Error('获取任务列表失败');
      }
    } catch (error) {
      console.error('获取任务列表失败:', error);
      // 使用模拟数据
      const mockData: Task[] = [
        {
          id: 'task1',
          instanceId: 'inst1',
          workflowName: '文档审批流程',
          stepName: '部门负责人审批',
          title: 'Q4财务报告审批',
          description: '请审核2023年第四季度财务报告，确认数据准确性后批准上传至知识库。',
          status: 'pending',
          priority: 'high',
          assigneeId: 'user1',
          assigneeName: '李四',
          requestorId: 'user2',
          requestorName: '张三',
          createdAt: '2023-12-01T10:30:00Z',
          dueDate: '2023-12-05T18:00:00Z',
          data: {
            documentId: 'doc123',
            documentName: 'Q4财务报告.pdf',
            documentSize: '2.5MB',
            requestType: 'upload'
          },
          comments: [
            {
              id: 'comment1',
              taskId: 'task1',
              userId: 'user2',
              userName: '张三',
              content: '财务报告已准备完毕，请审核',
              createdAt: '2023-12-01T10:30:00Z'
            }
          ],
          attachments: [
            {
              id: 'att1',
              taskId: 'task1',
              fileName: 'Q4财务报告.pdf',
              fileSize: 2621440,
              fileType: 'application/pdf',
              uploadedBy: '张三',
              uploadedAt: '2023-12-01T10:25:00Z'
            }
          ]
        },
        {
          id: 'task2',
          instanceId: 'inst2',
          workflowName: '员工入职流程',
          stepName: 'HR初审',
          title: '新员工王五入职审批',
          description: '请审核新员工王五的入职申请，确认其资格和材料完整性。',
          status: 'pending',
          priority: 'normal',
          assigneeId: 'user1',
          assigneeName: '李四',
          requestorId: 'user3',
          requestorName: 'HR部门',
          createdAt: '2023-12-02T09:15:00Z',
          dueDate: '2023-12-06T17:00:00Z',
          data: {
            employeeName: '王五',
            department: '技术部',
            position: '前端工程师',
            startDate: '2023-12-15'
          },
          comments: [],
          attachments: []
        }
      ];
      
      // 根据状态过滤
      const filteredData = tabValue === 'all' 
        ? mockData 
        : mockData.filter(task => task.status === tabValue);
      
      setTasks(filteredData);
    } finally {
      setLoading(false);
    }
  };

  // 查看任务详情
  const handleViewTask = (task: Task) => {
    setSelectedTask(task);
    setDialogType('view');
    setComment('');
    setOpenDialog(true);
  };

  // 处理任务
  const handleProcessTask = (task: Task) => {
    setSelectedTask(task);
    setDialogType('process');
    setComment('');
    setOpenDialog(true);
  };

  // 批准任务
  const handleApproveTask = async () => {
    if (!selectedTask) return;

    try {
      setProcessing(true);
      const response = await fetch(`/api/v1/workflow/tasks/${selectedTask.id}/approve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          comment: comment,
          decision: 'approved'
        }),
      });

      if (response.ok) {
        setTasks(prev => prev.map(task => 
          task.id === selectedTask.id 
            ? { ...task, status: 'completed', completedAt: new Date().toISOString() }
            : task
        ));
        
        setOpenDialog(false);
        toast.success({ title: '任务已批准' });
      } else {
        throw new Error('批准任务失败');
      }
    } catch (error) {
      console.error('批准任务失败:', error);
      toast.error({ title: '批准任务失败，请重试' });
    } finally {
      setProcessing(false);
    }
  };

  // 拒绝任务
  const handleRejectTask = async () => {
    if (!selectedTask) return;

    if (!comment.trim()) {
      toast.error({ title: '拒绝任务时必须填写原因' });
      return;
    }

    try {
      setProcessing(true);
      const response = await fetch(`/api/v1/workflow/tasks/${selectedTask.id}/reject`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          comment: comment,
          decision: 'rejected'
        }),
      });

      if (response.ok) {
        setTasks(prev => prev.map(task => 
          task.id === selectedTask.id 
            ? { ...task, status: 'rejected', completedAt: new Date().toISOString() }
            : task
        ));
        
        setOpenDialog(false);
        toast.success({ title: '任务已拒绝' });
      } else {
        throw new Error('拒绝任务失败');
      }
    } catch (error) {
      console.error('拒绝任务失败:', error);
      toast.error({ title: '拒绝任务失败，请重试' });
    } finally {
      setProcessing(false);
    }
  };

  // 添加评论
  const handleAddComment = async () => {
    if (!selectedTask || !comment.trim()) return;

    try {
      const response = await fetch(`/api/v1/workflow/tasks/${selectedTask.id}/comments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: comment }),
      });

      if (response.ok) {
        const newComment = {
          id: Date.now().toString(),
          taskId: selectedTask.id,
          userId: 'current-user',
          userName: '当前用户',
          content: comment,
          createdAt: new Date().toISOString()
        };

        setSelectedTask(prev => prev ? {
          ...prev,
          comments: [...prev.comments, newComment]
        } : null);
        
        setComment('');
        toast.success({ title: '评论已添加' });
      }
    } catch (error) {
      console.error('添加评论失败:', error);
      toast.error({ title: '添加评论失败' });
    }
  };

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  // 格式化文件大小
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // 获取优先级颜色
  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent': return 'destructive';
      case 'high': return 'warning';
      case 'normal': return 'secondary';
      case 'low': return 'outline';
      default: return 'secondary';
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'warning';
      case 'in_progress': return 'info';
      case 'completed': return 'success';
      case 'rejected': return 'destructive';
      default: return 'secondary';
    }
  };

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待处理';
      case 'in_progress': return '处理中';
      case 'completed': return '已完成';
      case 'rejected': return '已拒绝';
      default: return status;
    }
  };

  const filteredTasks = tasks.filter(task => {
    if (tabValue === 'all') return true;
    return task.status === tabValue;
  });

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">我的任务</h1>
          <p className="text-muted-foreground mt-2">
            管理和处理分配给您的工作流任务
          </p>
        </div>
        <Button
          onClick={fetchTasks}
          variant="outline"
          className="flex items-center gap-2"
        >
          <RotateCcw className="h-4 w-4" />
          刷新
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待处理</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {tasks.filter(task => task.status === 'pending').length}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">处理中</CardTitle>
            <Progress className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {tasks.filter(task => task.status === 'in_progress').length}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已完成</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {tasks.filter(task => task.status === 'completed').length}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总任务</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{tasks.length}</div>
          </CardContent>
        </Card>
      </div>

      {/* 任务列表 */}
      <Card>
        <CardHeader>
          <CardTitle>任务列表</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs value={tabValue} onValueChange={setTabValue}>
            <TabsList className="grid w-full grid-cols-5">
              <TabsTrigger value="all">全部</TabsTrigger>
              <TabsTrigger value="pending">待处理</TabsTrigger>
              <TabsTrigger value="in_progress">处理中</TabsTrigger>
              <TabsTrigger value="completed">已完成</TabsTrigger>
              <TabsTrigger value="rejected">已拒绝</TabsTrigger>
            </TabsList>
            
            <TabsContent value={tabValue} className="mt-6">
              {loading ? (
                <div className="flex items-center justify-center py-8">
                  <Progress className="w-32" />
                  <span className="ml-2">加载中...</span>
                </div>
              ) : filteredTasks.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>暂无任务</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>任务信息</TableHead>
                      <TableHead>工作流</TableHead>
                      <TableHead>优先级</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>申请人</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>截止时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredTasks.map((task) => (
                      <TableRow key={task.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{task.title}</div>
                            <div className="text-sm text-muted-foreground">
                              {task.stepName}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>{task.workflowName}</TableCell>
                        <TableCell>
                          <Badge variant={getPriorityColor(task.priority)}>
                            {task.priority === 'urgent' && '紧急'}
                            {task.priority === 'high' && '高'}
                            {task.priority === 'normal' && '普通'}
                            {task.priority === 'low' && '低'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getStatusColor(task.status)}>
                            {getStatusText(task.status)}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Avatar className="h-6 w-6">
                              <AvatarFallback className="text-xs">
                                {task.requestorName.charAt(0)}
                              </AvatarFallback>
                            </Avatar>
                            {task.requestorName}
                          </div>
                        </TableCell>
                        <TableCell>{formatDate(task.createdAt)}</TableCell>
                        <TableCell>
                          {task.dueDate ? formatDate(task.dueDate) : '-'}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleViewTask(task)}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                            {task.status === 'pending' && (
                              <Button
                                size="sm"
                                onClick={() => handleProcessTask(task)}
                              >
                                处理
                              </Button>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
      {/* 任务详情对话框 */}
      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {dialogType === 'view' ? '任务详情' : '处理任务'}
            </DialogTitle>
          </DialogHeader>
          
          {selectedTask && (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>任务信息</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label className="text-sm font-medium">任务标题</Label>
                      <p className="mt-1">{selectedTask.title}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium">工作流</Label>
                      <p className="mt-1">{selectedTask.workflowName}</p>
                    </div>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">任务描述</Label>
                    <p className="mt-1 text-sm">{selectedTask.description}</p>
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <Label className="text-sm font-medium">状态</Label>
                      <div className="mt-1">
                        <Badge variant={getStatusColor(selectedTask.status)}>
                          {getStatusText(selectedTask.status)}
                        </Badge>
                      </div>
                    </div>
                    <div>
                      <Label className="text-sm font-medium">优先级</Label>
                      <div className="mt-1">
                        <Badge variant={getPriorityColor(selectedTask.priority)}>
                          {selectedTask.priority === 'urgent' && '紧急'}
                          {selectedTask.priority === 'high' && '高'}
                          {selectedTask.priority === 'normal' && '普通'}
                          {selectedTask.priority === 'low' && '低'}
                        </Badge>
                      </div>
                    </div>
                    <div>
                      <Label className="text-sm font-medium">申请人</Label>
                      <p className="mt-1">{selectedTask.requestorName}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
              
              {/* 处理区域 */}
              {dialogType === 'process' && selectedTask.status === 'pending' && (
                <Card>
                  <CardHeader>
                    <CardTitle>处理意见</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <Textarea
                      placeholder="请输入处理意见..."
                      value={comment}
                      onChange={(e) => setComment(e.target.value)}
                      rows={4}
                    />
                  </CardContent>
                </Card>
              )}
              
              {/* 评论列表 */}
              {selectedTask.comments.length > 0 && (
                <Card>
                  <CardHeader>
                    <CardTitle>评论历史</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-3">
                      {selectedTask.comments.map((comment) => (
                        <div key={comment.id} className="flex gap-3 p-3 bg-muted rounded">
                          <Avatar className="h-8 w-8">
                            <AvatarFallback>{comment.userName.charAt(0)}</AvatarFallback>
                          </Avatar>
                          <div className="flex-1">
                            <div className="flex items-center gap-2 mb-1">
                              <span className="font-medium">{comment.userName}</span>
                              <span className="text-xs text-muted-foreground">
                                {formatDate(comment.createdAt)}
                              </span>
                            </div>
                            <p className="text-sm">{comment.content}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
          
          <DialogFooter>
            {dialogType === 'process' && selectedTask?.status === 'pending' && (
              <div className="flex gap-2">
                <Button
                  variant="destructive"
                  onClick={handleRejectTask}
                  disabled={processing}
                >
                  拒绝
                </Button>
                <Button
                  onClick={handleApproveTask}
                  disabled={processing}
                >
                  {processing ? '处理中...' : '批准'}
                </Button>
              </div>
            )}
            <Button variant="outline" onClick={() => setOpenDialog(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default TaskManagement;