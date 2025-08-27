'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Chip,
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  Alert,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Avatar,
} from '@mui/material';
import {
  CheckCircle as ApprovedIcon,
  Cancel as RejectedIcon,
  Pending as PendingIcon,
  History as HistoryIcon,
  Comment as CommentIcon,
  Person as PersonIcon,
  Event as EventIcon,
  Description as DocumentIcon,
  PriorityHigh as PriorityIcon,
} from '@mui/icons-material';

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
      alert(`审批${dialogAction === 'approve' ? '通过' : '拒绝'}成功`);
    } catch (err) {
      console.error('审批操作失败:', err);
      alert('审批操作失败，请重试');
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
      case 'submit': return <PersonIcon />;
      case 'approve': return <ApprovedIcon color="success" />;
      case 'reject': return <RejectedIcon color="error" />;
      case 'notify': return <CommentIcon />;
      default: return <EventIcon />;
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
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={3}>
        <Alert severity="error">{error}</Alert>
      </Box>
    );
  }

  if (!approval) {
    return (
      <Box p={3}>
        <Alert severity="info">未找到审批详情</Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ width: '100%', p: 3 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        审批详情
      </Typography>
      
      <Paper sx={{ p: 3, mb: 3 }}>
        <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
          <Box>
            <Typography variant="h5" gutterBottom>
              {approval.name}
            </Typography>
            <Box display="flex" alignItems="center" gap={1} mb={1}>
              <Chip 
                label={getStatusLabel(approval.status)}
                color={getStatusColor(approval.status)}
                size="small"
              />
              <Chip 
                label={getPriorityLabel(approval.priority)}
                color={getPriorityColor(approval.priority)}
                size="small"
                icon={<PriorityIcon />}
              />
            </Box>
          </Box>
          
          {approval.status === 'pending' && (
            <Box>
              <Button
                variant="contained"
                color="success"
                startIcon={<ApprovedIcon />}
                onClick={() => handleApproveReject('approve')}
                sx={{ mr: 1 }}
              >
                通过
              </Button>
              <Button
                variant="contained"
                color="error"
                startIcon={<RejectedIcon />}
                onClick={() => handleApproveReject('reject')}
              >
                拒绝
              </Button>
            </Box>
          )}
        </Box>
        
        <Divider sx={{ my: 2 }} />
        
        <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 3 }}>
          <Box>
            <Box display="flex" alignItems="center" mb={2}>
              <DocumentIcon sx={{ mr: 1, color: 'text.secondary' }} />
              <Typography variant="subtitle2" color="textSecondary">
                文档信息
              </Typography>
            </Box>
            <Typography variant="body1" paragraph>
              <strong>文档名称:</strong> {approval.documentName}
            </Typography>
          </Box>
          
          <Box>
            <Box display="flex" alignItems="center" mb={2}>
              <PersonIcon sx={{ mr: 1, color: 'text.secondary' }} />
              <Typography variant="subtitle2" color="textSecondary">
                人员信息
              </Typography>
            </Box>
            <Typography variant="body1" paragraph>
              <strong>申请人:</strong> {approval.requestorName}
            </Typography>
            <Typography variant="body1" paragraph>
              <strong>审批人:</strong> {approval.approverName}
            </Typography>
          </Box>
          
          <Box>
            <Box display="flex" alignItems="center" mb={2}>
              <EventIcon sx={{ mr: 1, color: 'text.secondary' }} />
              <Typography variant="subtitle2" color="textSecondary">
                时间信息
              </Typography>
            </Box>
            <Typography variant="body1" paragraph>
              <strong>提交时间:</strong> {formatDate(approval.submittedAt)}
            </Typography>
            {approval.deadline && (
              <Typography variant="body1" paragraph>
                <strong>截止时间:</strong> {formatDate(approval.deadline)}
              </Typography>
            )}
          </Box>
          
          <Box>
            <Box display="flex" alignItems="center" mb={2}>
              <CommentIcon sx={{ mr: 1, color: 'text.secondary' }} />
              <Typography variant="subtitle2" color="textSecondary">
                备注信息
              </Typography>
            </Box>
            <Typography variant="body1" paragraph>
              <strong>描述:</strong> {approval.description || '无'}
            </Typography>
            <Typography variant="body1" paragraph>
              <strong>备注:</strong> {approval.comments || '无'}
            </Typography>
          </Box>
        </Box>
      </Paper>
      
      <Paper sx={{ p: 3 }}>
        <Box display="flex" alignItems="center" mb={2}>
          <HistoryIcon sx={{ mr: 1, color: 'text.secondary' }} />
          <Typography variant="h6" component="h2">
            审批历史
          </Typography>
        </Box>
        
        {approval.history.length === 0 ? (
          <Alert severity="info">暂无审批历史记录</Alert>
        ) : (
          <List>
            {approval.history.map((record) => (
              <ListItem key={record.id} alignItems="flex-start">
                <ListItemIcon>
                  <Avatar sx={{ width: 24, height: 24 }}>
                    {getActionIcon(record.action)}
                  </Avatar>
                </ListItemIcon>
                <ListItemText
                  primary={
                    <Box display="flex" alignItems="center">
                      <Typography variant="subtitle2" sx={{ mr: 1 }}>
                        {record.actorName}
                      </Typography>
                      <Chip 
                        label={getActionLabel(record.action)} 
                        size="small" 
                        variant="outlined"
                      />
                    </Box>
                  }
                  secondary={
                    <>
                      <Typography variant="body2" color="textSecondary">
                        {record.comments}
                      </Typography>
                      <Typography variant="caption" display="block" color="textSecondary">
                        {formatDate(record.actionTime)}
                      </Typography>
                    </>
                  }
                />
              </ListItem>
            ))}
          </List>
        )}
      </Paper>
      
      {/* 审批操作对话框 */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {dialogAction === 'approve' ? '通过审批' : '拒绝审批'}
        </DialogTitle>
        <DialogContent>
          <TextField
            label="审批意见"
            multiline
            rows={4}
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            fullWidth
            margin="normal"
            placeholder={`请输入${dialogAction === 'approve' ? '通过' : '拒绝'}意见`}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          <Button 
            variant="contained" 
            color={dialogAction === 'approve' ? 'success' : 'error'}
            onClick={handleSubmitAction}
          >
            {dialogAction === 'approve' ? '通过' : '拒绝'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default ApprovalDetailPage;