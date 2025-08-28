'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Chip,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Refresh as RefreshIcon,
  Visibility as VisibilityIcon,
  Check as ApproveIcon,
  Close as RejectIcon,
  Add as AddIcon,
} from '@mui/icons-material';

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
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));
        
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
        setLoading(false);
      } catch (error) {
        console.error('获取审批流程数据失败:', error);
        setLoading(false);
      }
    };

    fetchApprovals();
  }, []);

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
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));
      
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
      alert(`审批${dialogAction === 'approve' ? '通过' : '拒绝'}成功`);
    } catch (error) {
      console.error('审批操作失败:', error);
      alert('审批操作失败，请重试');
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

  // 刷新数据
  const handleRefresh = () => {
    setLoading(true);
    // 模拟刷新
    setTimeout(() => {
      setLoading(false);
    }, 500);
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ width: '100%', p: 3 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1" gutterBottom>
          审批流程管理
        </Typography>
        <Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            sx={{ mr: 2 }}
            onClick={() => alert('创建审批流程功能待实现')}
          >
            创建审批
          </Button>
          <IconButton onClick={handleRefresh} color="primary">
            <RefreshIcon />
          </IconButton>
        </Box>
      </Box>

      {approvals.length === 0 ? (
        <Alert severity="info">暂无审批流程数据</Alert>
      ) : (
        <TableContainer component={Paper}>
          <Table sx={{ minWidth: 650 }} aria-label="审批流程表格">
            <TableHead>
              <TableRow>
                <TableCell>审批名称</TableCell>
                <TableCell>文档名称</TableCell>
                <TableCell>申请人</TableCell>
                <TableCell>审批人</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>优先级</TableCell>
                <TableCell>提交时间</TableCell>
                <TableCell>截止时间</TableCell>
                <TableCell>操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {approvals.map((approval) => (
                <TableRow key={approval.id} hover>
                  <TableCell>{approval.name}</TableCell>
                  <TableCell>{approval.documentName}</TableCell>
                  <TableCell>{approval.requestorName}</TableCell>
                  <TableCell>{approval.approverName}</TableCell>
                  <TableCell>
                    <Chip 
                      label={approval.status === 'pending' ? '待审批' : 
                             approval.status === 'approved' ? '已通过' : 
                             approval.status === 'rejected' ? '已拒绝' : '已取消'}
                      color={getStatusColor(approval.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={getPriorityLabel(approval.priority)}
                      color={getPriorityColor(approval.priority)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{formatDate(approval.submittedAt)}</TableCell>
                  <TableCell>
                    {approval.deadline ? formatDate(approval.deadline) : '-'}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="查看详情">
                      <IconButton 
                        size="small" 
                        onClick={() => handleViewApproval(approval)}
                      >
                        <VisibilityIcon />
                      </IconButton>
                    </Tooltip>
                    {approval.status === 'pending' && (
                      <>
                        <Tooltip title="通过">
                          <IconButton 
                            size="small" 
                            color="success"
                            onClick={() => handleApproveReject(approval, 'approve')}
                          >
                            <ApproveIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="拒绝">
                          <IconButton 
                            size="small" 
                            color="error"
                            onClick={() => handleApproveReject(approval, 'reject')}
                          >
                            <RejectIcon />
                          </IconButton>
                        </Tooltip>
                      </>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* 审批详情和操作对话框 */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {dialogAction === 'view' ? '审批详情' : 
           dialogAction === 'approve' ? '通过审批' : '拒绝审批'}
        </DialogTitle>
        <DialogContent>
          {selectedApproval && (
            <Box sx={{ py: 2 }}>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">审批名称</Typography>
                <Typography variant="body1">{selectedApproval.name}</Typography>
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">文档名称</Typography>
                <Typography variant="body1">{selectedApproval.documentName}</Typography>
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">申请人</Typography>
                <Typography variant="body1">{selectedApproval.requestorName}</Typography>
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">审批人</Typography>
                <Typography variant="body1">{selectedApproval.approverName}</Typography>
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">状态</Typography>
                <Chip 
                  label={selectedApproval.status === 'pending' ? '待审批' : 
                         selectedApproval.status === 'approved' ? '已通过' : 
                         selectedApproval.status === 'rejected' ? '已拒绝' : '已取消'}
                  color={getStatusColor(selectedApproval.status)}
                  size="small"
                />
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">优先级</Typography>
                <Chip 
                  label={getPriorityLabel(selectedApproval.priority)}
                  color={getPriorityColor(selectedApproval.priority)}
                  size="small"
                />
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography variant="subtitle2" color="textSecondary">提交时间</Typography>
                <Typography variant="body1">{formatDate(selectedApproval.submittedAt)}</Typography>
              </Box>
              {selectedApproval.deadline && (
                <Box sx={{ mb: 2 }}>
                  <Typography variant="subtitle2" color="textSecondary">截止时间</Typography>
                  <Typography variant="body1">{formatDate(selectedApproval.deadline)}</Typography>
                </Box>
              )}
              {selectedApproval.comments && (
                <Box sx={{ mb: 2 }}>
                  <Typography variant="subtitle2" color="textSecondary">备注</Typography>
                  <Typography variant="body1">{selectedApproval.comments}</Typography>
                </Box>
              )}
              
              {(dialogAction === 'approve' || dialogAction === 'reject') && (
                <FormControl fullWidth sx={{ mt: 2 }}>
                  <TextField
                    label="审批意见"
                    multiline
                    rows={3}
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                    placeholder={`请输入${dialogAction === 'approve' ? '通过' : '拒绝'}意见`}
                  />
                </FormControl>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          {dialogAction !== 'view' && (
            <Button 
              variant="contained" 
              color={dialogAction === 'approve' ? 'success' : 'error'}
              onClick={handleSubmitAction}
            >
              {dialogAction === 'approve' ? '通过' : '拒绝'}
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default ApprovalManagement;