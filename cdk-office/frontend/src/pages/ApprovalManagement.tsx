'use client';

import React, { useState, useEffect } from 'react';
import { RefreshCw, Eye, Check, X, Plus, Calendar, AlertCircle } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
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
import { Textarea } from '@/components/ui/textarea';
import { Progress } from '@/components/ui/progress';
import { toast } from '@/components/ui/use-toast';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Alert, AlertDescription } from '@/components/ui/alert';

// 审批流程数据接口
interface ApprovalProcess {
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
}

const ApprovalManagement: React.FC = () => {
  const [approvals, setApprovals] = useState<ApprovalProcess[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedApproval, setSelectedApproval] = useState<ApprovalProcess | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [dialogAction, setDialogAction] = useState<'view' | 'approve' | 'reject'>('view');
  const [comment, setComment] = useState('');

  // 模拟获取审批流程数据
  useEffect(() => {
    const fetchApprovals = async () => {
      try {
        setLoading(true);
        
        // 使用真实API调用
        const response = await fetch('/api/v1/workflow/instances?status=all');
        if (response.ok) {
          const data = await response.json();
          
          // 转换后端数据格式为前端格式
          const transformedData: ApprovalProcess[] = (data.data || []).map((instance: any) => ({
            id: instance.id,
            name: instance.definition_name || '审批流程',
            documentName: instance.input?.documentName || instance.input?.title || '未知文档',
            requestorName: instance.started_by || '未知用户',
            approverName: instance.assigned_to || '待分配',
            status: mapInstanceStatusToApprovalStatus(instance.status),
            priority: instance.priority || 'normal',
            submittedAt: instance.started_at,
            deadline: instance.due_date,
            comments: instance.description || '',
          }));
          
          setApprovals(transformedData);
        } else {
          throw new Error('获取审批流程数据失败');
        }
      } catch (error) {
        console.error('获取审批流程数据失败:', error);
        
        // 降级使用模拟数据
        const mockData: ApprovalProcess[] = [
          {
            id: '1',
            name: '文档上传审批',
            documentName: '2023年Q4财务报告.pdf',
            requestorName: '张三',
            approverName: '李四',
            status: 'pending',
            priority: 'high',
            submittedAt: '2023-12-01T10:30:00Z',
            deadline: '2023-12-05T18:00:00Z',
            comments: '请审核Q4财务报告',
          },
          {
            id: '2',
            name: '文档更新审批',
            documentName: '员工手册_v2.0.docx',
            requestorName: '王五',
            approverName: '李四',
            status: 'approved',
            priority: 'normal',
            submittedAt: '2023-11-28T09:15:00Z',
            comments: '更新了员工福利政策',
          },
          {
            id: '3',
            name: '文档删除审批',
            documentName: '旧版产品规格书.pdf',
            requestorName: '赵六',
            approverName: '李四',
            status: 'pending',
            priority: 'low',
            submittedAt: '2023-12-02T14:20:00Z',
            deadline: '2023-12-07T18:00:00Z',
            comments: '请求删除过时文档',
          },
        ];
        
        setApprovals(mockData);
      } finally {
        setLoading(false);
      }
    };

    fetchApprovals();
  }, []);

  // 状态映射函数
  const mapInstanceStatusToApprovalStatus = (instanceStatus: string): 'pending' | 'approved' | 'rejected' | 'cancelled' => {
    switch (instanceStatus) {
      case 'running': return 'pending';
      case 'completed': return 'approved';
      case 'failed': return 'rejected';
      case 'cancelled': return 'cancelled';
      default: return 'pending';
    }
  };

  // 处理查看审批详情
  const handleViewApproval = (approval: ApprovalProcess) => {
    setSelectedApproval(approval);
    setDialogAction('view');
    setOpenDialog(true);
  };

  // 处理审批操作
  const handleApproveReject = (approval: ApprovalProcess, action: 'approve' | 'reject') => {
    setSelectedApproval(approval);
    setDialogAction(action);
    setComment('');
    setOpenDialog(true);
  };

  // 提交审批操作
  const handleSubmitAction = async () => {
    if (!selectedApproval) return;

    try {
      setLoading(true);
      
      // 使用真实API调用
      const endpoint = dialogAction === 'approve' ? 'approve' : 'reject';
      const response = await fetch(`/api/v1/workflow/instances/${selectedApproval.id}/${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          comment: comment,
          decision: dialogAction === 'approve' ? 'approved' : 'rejected'
        }),
      });
      
      if (response.ok) {
        // 更新本地状态
        setApprovals(prev => prev.map(approval => 
          approval.id === selectedApproval.id 
            ? { 
                ...approval, 
                status: dialogAction === 'approve' ? 'approved' : 'rejected',
                comments: comment || approval.comments
              } 
            : approval
        ));
        
        setOpenDialog(false);
        setSelectedApproval(null);
        setComment('');
        
        // 显示成功消息
        toast.success({
          title: `审批${dialogAction === 'approve' ? '通过' : '拒绝'}成功`,
        });
      } else {
        throw new Error(`审批${dialogAction === 'approve' ? '通过' : '拒绝'}失败`);
      }
    } catch (error) {
      console.error(`审批${dialogAction === 'approve' ? '通过' : '拒绝'}失败:`, error);
      toast.error({
        title: `审批${dialogAction === 'approve' ? '通过' : '拒绝'}失败`,
        description: '请重试或联系管理员',
      });
    } finally {
      setLoading(false);
    }
  };

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
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
      case 'approved': return 'success';
      case 'rejected': return 'destructive';
      case 'cancelled': return 'outline';
      default: return 'secondary';
    }
  };

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待审批';
      case 'approved': return '已通过';
      case 'rejected': return '已拒绝';
      case 'cancelled': return '已取消';
      default: return status;
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">审批管理</h1>
          <p className="text-muted-foreground mt-2">
            管理和处理文档审批流程
          </p>
        </div>
        <Button
          onClick={() => window.location.reload()}
          variant="outline"
          className="flex items-center gap-2"
        >
          <RefreshCw className="h-4 w-4" />
          刷新
        </Button>
      </div>

      {/* 审批列表 */}
      <Card>
        <CardHeader>
          <CardTitle>审批流程列表</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Progress className="w-32" />
              <span className="ml-2">加载中...</span>
            </div>
          ) : approvals.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>暂无审批流程</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>文档名称</TableHead>
                  <TableHead>流程名称</TableHead>
                  <TableHead>申请人</TableHead>
                  <TableHead>审批人</TableHead>
                  <TableHead>优先级</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>提交时间</TableHead>
                  <TableHead>截止时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {approvals.map((approval) => (
                  <TableRow key={approval.id}>
                    <TableCell className="font-medium">{approval.documentName}</TableCell>
                    <TableCell>{approval.name}</TableCell>
                    <TableCell>{approval.requestorName}</TableCell>
                    <TableCell>{approval.approverName}</TableCell>
                    <TableCell>
                      <Badge variant={getPriorityColor(approval.priority)}>
                        {approval.priority === 'urgent' && '紧急'}
                        {approval.priority === 'high' && '高'}
                        {approval.priority === 'normal' && '普通'}
                        {approval.priority === 'low' && '低'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={getStatusColor(approval.status)}>
                        {getStatusText(approval.status)}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatDate(approval.submittedAt)}</TableCell>
                    <TableCell>
                      {approval.deadline ? formatDate(approval.deadline) : '-'}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleViewApproval(approval)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        {approval.status === 'pending' && (
                          <>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleApproveReject(approval, 'approve')}
                              className="text-green-600 hover:text-green-700"
                            >
                              <Check className="h-4 w-4" />
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleApproveReject(approval, 'reject')}
                              className="text-red-600 hover:text-red-700"
                            >
                              <X className="h-4 w-4" />
                            </Button>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* 审批详情和操作对话框 */}
      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {dialogAction === 'view' ? '审批详情' : 
               dialogAction === 'approve' ? '通过审批' : '拒绝审批'}
            </DialogTitle>
          </DialogHeader>
          
          {selectedApproval && (
            <ScrollArea className="max-h-[60vh]">
              <div className="py-4 space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">审批名称</label>
                    <p className="mt-1">{selectedApproval.name}</p>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">文档名称</label>
                    <p className="mt-1">{selectedApproval.documentName}</p>
                  </div>
                </div>
                
                <Separator />
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">申请人</label>
                    <p className="mt-1">{selectedApproval.requestorName}</p>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">审批人</label>
                    <p className="mt-1">{selectedApproval.approverName}</p>
                  </div>
                </div>
                
                <Separator />
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">状态</label>
                    <div className="mt-1">
                      <Badge variant={getStatusColor(selectedApproval.status)}>
                        {getStatusText(selectedApproval.status)}
                      </Badge>
                    </div>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">优先级</label>
                    <div className="mt-1">
                      <Badge variant={getPriorityColor(selectedApproval.priority)}>
                        {selectedApproval.priority === 'urgent' && '紧急'}
                        {selectedApproval.priority === 'high' && '高'}
                        {selectedApproval.priority === 'normal' && '普通'}
                        {selectedApproval.priority === 'low' && '低'}
                      </Badge>
                    </div>
                  </div>
                </div>
                
                <Separator />
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">提交时间</label>
                    <p className="mt-1 flex items-center gap-2">
                      <Calendar className="h-4 w-4" />
                      {formatDate(selectedApproval.submittedAt)}
                    </p>
                  </div>
                  {selectedApproval.deadline && (
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">截止时间</label>
                      <p className="mt-1 flex items-center gap-2">
                        <Calendar className="h-4 w-4" />
                        {formatDate(selectedApproval.deadline)}
                      </p>
                    </div>
                  )}
                </div>
                
                {selectedApproval.comments && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">备注</label>
                    <p className="mt-1 text-sm">{selectedApproval.comments}</p>
                  </div>
                )}
                
                {(dialogAction === 'approve' || dialogAction === 'reject') && (
                  <div className="space-y-2">
                    <label className="text-sm font-medium text-muted-foreground">
                      审批意见
                    </label>
                    <Textarea
                      placeholder={`请输入${dialogAction === 'approve' ? '通过' : '拒绝'}意见`}
                      value={comment}
                      onChange={(e) => setComment(e.target.value)}
                      rows={4}
                    />
                  </div>
                )}
              </div>
            </ScrollArea>
          )}
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenDialog(false)}>
              取消
            </Button>
            {dialogAction !== 'view' && (
              <Button 
                variant={dialogAction === 'approve' ? 'default' : 'destructive'}
                onClick={handleSubmitAction}
                disabled={loading}
              >
                {loading ? '处理中...' : (dialogAction === 'approve' ? '通过' : '拒绝')}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ApprovalManagement;