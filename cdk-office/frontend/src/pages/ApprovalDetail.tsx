'use client';

import React, { useState, useEffect } from 'react';
import {
  CheckCircle,
  X,
  Clock,
  History,
  MessageSquare,
  User,
  Calendar,
  FileText,
  AlertTriangle,
  ArrowLeft,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { toast } from '@/components/ui/use-toast';

// 审批历史记录接口
interface ApprovalHistory {
  id: string;
  actorName: string;
  action: string;
  comments: string;
  actionTime: string;
}

// 审批流程详情接口
interface ApprovalDetail {
  id: string;
  name: string;
  documentName: string;
  requestorName: string;
  approverName: string;
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  priority: 'low' | 'normal' | 'high' | 'urgent';
  submittedAt: string;
  deadline?: string;
  comments?: string;
  description?: string;
  history: ApprovalHistory[];
}

const ApprovalDetailPage: React.FC<{ approvalId: string }> = ({ approvalId }) => {
  const [approval, setApproval] = useState<ApprovalDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [openDialog, setOpenDialog] = useState(false);
  const [dialogAction, setDialogAction] = useState<'approve' | 'reject'>('approve');
  const [comment, setComment] = useState('');
  const [error, setError] = useState<string | null>(null);

  // 模拟获取审批详情数据
  useEffect(() => {
    const fetchApprovalDetail = async () => {
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 800));
        
        // 模拟数据
        const mockData: ApprovalDetail = {
          id: approvalId,
          name: '文档上传审批',
          documentName: '2023年Q4财务报告.pdf',
          requestorName: '张三',
          approverName: '李四',
          status: 'pending',
          priority: 'high',
          submittedAt: '2023-12-01T10:30:00Z',
          deadline: '2023-12-05T18:00:00Z',
          comments: '请审核Q4财务报告',
          description: '这是2023年第四季度的财务报告，请审批后上传至知识库。',
          history: [
            {
              id: '1',
              actorName: '张三',
              action: 'submit',
              comments: '提交审批',
              actionTime: '2023-12-01T10:30:00Z',
            },
            {
              id: '2',
              actorName: '系统',
              action: 'notify',
              comments: '发送审批通知给李四',
              actionTime: '2023-12-01T10:31:00Z',
            }
          ]
        };
        
        setApproval(mockData);
        setLoading(false);
      } catch (err) {
        setError('获取审批详情失败');
        setLoading(false);
      }
    };

    if (approvalId) {
      fetchApprovalDetail();
    }
  }, [approvalId]);

  // 处理审批操作
  const handleApproveReject = (action: 'approve' | 'reject') => {
    setDialogAction(action);
    setComment('');
    setOpenDialog(true);
  };

  // 提交审批操作
  const handleSubmitAction = async () => {
    if (!approval) return;

    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 更新本地状态
      setApproval({
        ...approval,
        status: dialogAction === 'approve' ? 'approved' : 'rejected',
        comments: comment || approval.comments
      });
      
      setOpenDialog(false);
      setComment('');
      
      // 显示成功消息
      toast({
        title: "成功",
        description: `审批${dialogAction === 'approve' ? '通过' : '拒绝'}成功`,
      });
    } catch (err) {
      console.error('审批操作失败:', err);
      toast({
        title: "错误",
        description: "审批操作失败，请重试",
        variant: "destructive",
      });
    }
  };

  // 获取状态标签
  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'pending': return '待审批';
      case 'approved': return '已通过';
      case 'rejected': return '已拒绝';
      case 'cancelled': return '已取消';
      default: return status;
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'warning';
      case 'approved': return 'success';
      case 'rejected': return 'error';
      case 'cancelled': return 'default';
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

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  // 获取操作图标
  const getActionIcon = (action: string) => {
    switch (action) {
      case 'submit': return <User className="h-4 w-4" />;
      case 'approve': return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'reject': return <X className="h-4 w-4 text-red-500" />;
      case 'notify': return <MessageSquare className="h-4 w-4" />;
      default: return <Calendar className="h-4 w-4" />;
    }
  };

  // 获取操作标签
  const getActionLabel = (action: string) => {
    switch (action) {
      case 'submit': return '提交';
      case 'approve': return '通过';
      case 'reject': return '拒绝';
      case 'notify': return '通知';
      case 'comment': return '评论';
      default: return action;
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="space-y-6">
          <Skeleton className="h-8 w-48" />
          <Card>
            <CardHeader>
              <Skeleton className="h-6 w-3/4" />
              <div className="flex gap-2">
                <Skeleton className="h-6 w-16" />
                <Skeleton className="h-6 w-16" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-4">
                <Skeleton className="h-20" />
                <Skeleton className="h-20" />
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-red-600" />
            <p className="text-red-800 font-medium">错误</p>
          </div>
          <p className="text-red-700 mt-2">{error}</p>
        </div>
      </div>
    );
  }

  if (!approval) {
    return (
      <div className="container mx-auto p-6">
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-blue-600" />
            <p className="text-blue-800 font-medium">信息</p>
          </div>
          <p className="text-blue-700 mt-2">未找到审批详情</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex items-center gap-4 mb-6">
        <Button variant="outline" size="icon">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h1 className="text-3xl font-bold">审批详情</h1>
      </div>
      
      <Card className="mb-6">
        <CardHeader>
          <div className="flex justify-between items-start">
            <div className="space-y-2">
              <CardTitle className="text-2xl">{approval.name}</CardTitle>
              <div className="flex items-center gap-2">
                <Badge 
                  variant={approval.status === 'pending' ? 'secondary' : 
                          approval.status === 'approved' ? 'default' : 
                          approval.status === 'rejected' ? 'destructive' : 'outline'}
                >
                  {getStatusLabel(approval.status)}
                </Badge>
                <Badge 
                  variant={approval.priority === 'urgent' ? 'destructive' : 
                          approval.priority === 'high' ? 'secondary' : 'outline'}
                >
                  <AlertTriangle className="h-3 w-3 mr-1" />
                  {getPriorityLabel(approval.priority)}
                </Badge>
              </div>
            </div>
            
            {approval.status === 'pending' && (
              <div className="flex gap-2">
                <Button
                  className="bg-green-600 hover:bg-green-700"
                  onClick={() => handleApproveReject('approve')}
                >
                  <CheckCircle className="h-4 w-4 mr-2" />
                  通过
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => handleApproveReject('reject')}
                >
                  <X className="h-4 w-4 mr-2" />
                  拒绝
                </Button>
              </div>
            )}
          </div>
        </CardHeader>
        
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-4">
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <FileText className="h-4 w-4 text-muted-foreground" />
                  <Label className="text-sm font-medium text-muted-foreground">文档信息</Label>
                </div>
                <p className="text-sm">
                  <span className="font-medium">文档名称:</span> {approval.documentName}
                </p>
              </div>
              
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <Label className="text-sm font-medium text-muted-foreground">人员信息</Label>
                </div>
                <div className="space-y-1">
                  <p className="text-sm">
                    <span className="font-medium">申请人:</span> {approval.requestorName}
                  </p>
                  <p className="text-sm">
                    <span className="font-medium">审批人:</span> {approval.approverName}
                  </p>
                </div>
              </div>
            </div>
            
            <div className="space-y-4">
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <Label className="text-sm font-medium text-muted-foreground">时间信息</Label>
                </div>
                <div className="space-y-1">
                  <p className="text-sm">
                    <span className="font-medium">提交时间:</span> {formatDate(approval.submittedAt)}
                  </p>
                  {approval.deadline && (
                    <p className="text-sm">
                      <span className="font-medium">截止时间:</span> {formatDate(approval.deadline)}
                    </p>
                  )}
                </div>
              </div>
              
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <MessageSquare className="h-4 w-4 text-muted-foreground" />
                  <Label className="text-sm font-medium text-muted-foreground">备注信息</Label>
                </div>
                <div className="space-y-1">
                  <p className="text-sm">
                    <span className="font-medium">描述:</span> {approval.description || '无'}
                  </p>
                  <p className="text-sm">
                    <span className="font-medium">备注:</span> {approval.comments || '无'}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <History className="h-5 w-5 text-muted-foreground" />
            <CardTitle>审批历史</CardTitle>
          </div>
        </CardHeader>
        
        <CardContent>
          {approval.history.length === 0 ? (
            <div className="text-center py-8">
              <History className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold">暂无审批历史记录</h3>
              <p className="text-muted-foreground">该审批尚无操作记录</p>
            </div>
          ) : (
            <ScrollArea className="h-64">
              <div className="space-y-4">
                {approval.history.map((record, index) => (
                  <div key={record.id} className="flex gap-4">
                    <div className="flex flex-col items-center">
                      <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground">
                        {getActionIcon(record.action)}
                      </div>
                      {index < approval.history.length - 1 && (
                        <div className="h-8 w-px bg-border" />
                      )}
                    </div>
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center gap-2">
                        <h4 className="text-sm font-medium">{record.actorName}</h4>
                        <Badge variant="outline" className="text-xs">
                          {getActionLabel(record.action)}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground">
                        {record.comments}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {formatDate(record.actionTime)}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}
        </CardContent>
      </Card>
      
      {/* 审批操作对话框 */}
      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              {dialogAction === 'approve' ? '通过审批' : '拒绝审批'}
            </DialogTitle>
            <DialogDescription>
              请输入{dialogAction === 'approve' ? '通过' : '拒绝'}意见，将作为审批记录保存。
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="comment">审批意见</Label>
              <Textarea
                id="comment"
                rows={4}
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                placeholder={`请输入${dialogAction === 'approve' ? '通过' : '拒绝'}意见`}
                className="mt-1"
              />
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenDialog(false)}>
              取消
            </Button>
            <Button 
              className={dialogAction === 'approve' ? 'bg-green-600 hover:bg-green-700' : ''}
              variant={dialogAction === 'approve' ? 'default' : 'destructive'}
              onClick={handleSubmitAction}
            >
              {dialogAction === 'approve' ? '通过' : '拒绝'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ApprovalDetailPage;