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
  Card,
  CardContent,
  Grid,
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot,
  TimelineOppositeContent,
  LinearProgress,
  Tabs,
  Tab,
} from '@mui/material';
import {
  Visibility as ViewIcon,
  Stop as StopIcon,
  PlayArrow as StartIcon,
  Refresh as RefreshIcon,
  CheckCircle as CompleteIcon,
  Cancel as CancelIcon,
  Pending as PendingIcon,
  Error as ErrorIcon,
  Comment as CommentIcon,
  Person as PersonIcon,
  Schedule as ScheduleIcon,
} from '@mui/icons-material';

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

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

const TabPanel: React.FC<TabPanelProps> = ({ children, value, index }) => {
  return (
    <div hidden={value !== index}>
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
};

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
    if (!window.confirm(`确认停止工作流实例"${instance.definitionName}"吗？`)) {
      return;
    }

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
        alert('工作流实例已停止');
      } else {
        throw new Error('停止失败');
      }
    } catch (error) {
      console.error('停止工作流实例失败:', error);
      alert('停止失败，请重试');
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
        alert('工作流实例已重启');
      } else {
        throw new Error('重启失败');
      }
    } catch (error) {
      console.error('重启工作流实例失败:', error);
      alert('重启失败，请重试');
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

  // 获取任务状态图标
  const getTaskStatusIcon = (status: string) => {
    switch (status) {
      case 'pending': return <PendingIcon color="warning" />;
      case 'in_progress': return <CircularProgress size={20} />;
      case 'completed': return <CompleteIcon color="success" />;
      case 'rejected': return <CancelIcon color="error" />;
      default: return <PendingIcon />;
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
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ width: '100%', p: 3 }}>
      {/* 页面标题和过滤器 */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          工作流实例管理
        </Typography>
        <Box display="flex" alignItems="center" gap={2}>
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>状态筛选</InputLabel>
            <Select
              value={statusFilter}
              label="状态筛选"
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <MenuItem value="all">全部</MenuItem>
              <MenuItem value="running">运行中</MenuItem>
              <MenuItem value="completed">已完成</MenuItem>
              <MenuItem value="failed">失败</MenuItem>
              <MenuItem value="cancelled">已取消</MenuItem>
              <MenuItem value="paused">暂停</MenuItem>
            </Select>
          </FormControl>
          <IconButton onClick={fetchInstances} color="primary">
            <RefreshIcon />
          </IconButton>
        </Box>
      </Box>

      {/* 实例列表 */}
      {instances.length === 0 ? (
        <Alert severity="info">暂无工作流实例</Alert>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>实例ID</TableCell>
                <TableCell>工作流名称</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>当前步骤</TableCell>
                <TableCell>进度</TableCell>
                <TableCell>优先级</TableCell>
                <TableCell>发起人</TableCell>
                <TableCell>开始时间</TableCell>
                <TableCell>持续时间</TableCell>
                <TableCell>操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {instances.map((instance) => (
                <TableRow key={instance.id} hover>
                  <TableCell>{instance.id}</TableCell>
                  <TableCell>{instance.definitionName}</TableCell>
                  <TableCell>
                    <Chip 
                      label={getStatusLabel(instance.status)}
                      color={getStatusColor(instance.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{instance.currentStep}</TableCell>
                  <TableCell>
                    <Box display="flex" alignItems="center" gap={1}>
                      <LinearProgress 
                        variant="determinate" 
                        value={instance.progress} 
                        sx={{ width: 80 }}
                      />
                      <Typography variant="caption">
                        {instance.progress}%
                      </Typography>
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={getPriorityLabel(instance.priority)}
                      color={getPriorityColor(instance.priority)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{instance.startedBy}</TableCell>
                  <TableCell>{formatDate(instance.startedAt)}</TableCell>
                  <TableCell>
                    {getDuration(instance.startedAt, instance.completedAt)}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="查看详情">
                      <IconButton 
                        size="small" 
                        onClick={() => handleViewInstance(instance)}
                      >
                        <ViewIcon />
                      </IconButton>
                    </Tooltip>
                    {instance.status === 'running' && (
                      <Tooltip title="停止">
                        <IconButton 
                          size="small" 
                          color="error"
                          onClick={() => handleStopInstance(instance)}
                        >
                          <StopIcon />
                        </IconButton>
                      </Tooltip>
                    )}
                    {(instance.status === 'failed' || instance.status === 'cancelled') && (
                      <Tooltip title="重启">
                        <IconButton 
                          size="small" 
                          color="primary"
                          onClick={() => handleRestartInstance(instance)}
                        >
                          <StartIcon />
                        </IconButton>
                      </Tooltip>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* 实例详情对话框 */}
      <Dialog 
        open={openDialog} 
        onClose={() => setOpenDialog(false)} 
        maxWidth="lg" 
        fullWidth
      >
        <DialogTitle>
          工作流实例详情 - {selectedInstance?.definitionName}
        </DialogTitle>
        
        <DialogContent sx={{ p: 0 }}>
          <Tabs value={tabValue} onChange={(_, newValue) => setTabValue(newValue)}>
            <Tab label="基本信息" />
            <Tab label="任务列表" />
            <Tab label="执行历史" />
            <Tab label="输入输出" />
          </Tabs>

          {/* 基本信息 */}
          <TabPanel value={tabValue} index={0}>
            {selectedInstance && (
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6" gutterBottom>实例信息</Typography>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">实例ID</Typography>
                        <Typography variant="body1">{selectedInstance.id}</Typography>
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">工作流名称</Typography>
                        <Typography variant="body1">{selectedInstance.definitionName}</Typography>
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">状态</Typography>
                        <Chip 
                          label={getStatusLabel(selectedInstance.status)}
                          color={getStatusColor(selectedInstance.status)}
                          size="small"
                        />
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">当前步骤</Typography>
                        <Typography variant="body1">{selectedInstance.currentStep}</Typography>
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">进度</Typography>
                        <Box display="flex" alignItems="center" gap={1}>
                          <LinearProgress 
                            variant="determinate" 
                            value={selectedInstance.progress} 
                            sx={{ flexGrow: 1 }}
                          />
                          <Typography variant="body2">
                            {selectedInstance.progress}%
                          </Typography>
                        </Box>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6" gutterBottom>时间信息</Typography>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">发起人</Typography>
                        <Typography variant="body1">{selectedInstance.startedBy}</Typography>
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">开始时间</Typography>
                        <Typography variant="body1">{formatDate(selectedInstance.startedAt)}</Typography>
                      </Box>
                      {selectedInstance.completedAt && (
                        <Box sx={{ mb: 2 }}>
                          <Typography variant="body2" color="textSecondary">完成时间</Typography>
                          <Typography variant="body1">{formatDate(selectedInstance.completedAt)}</Typography>
                        </Box>
                      )}
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">持续时间</Typography>
                        <Typography variant="body1">
                          {getDuration(selectedInstance.startedAt, selectedInstance.completedAt)}
                        </Typography>
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="body2" color="textSecondary">优先级</Typography>
                        <Chip 
                          label={getPriorityLabel(selectedInstance.priority)}
                          color={getPriorityColor(selectedInstance.priority)}
                          size="small"
                        />
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
            )}
          </TabPanel>

          {/* 任务列表 */}
          <TabPanel value={tabValue} index={1}>
            {selectedInstance?.tasks.length === 0 ? (
              <Alert severity="info">暂无任务</Alert>
            ) : (
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>任务ID</TableCell>
                      <TableCell>步骤名称</TableCell>
                      <TableCell>处理人</TableCell>
                      <TableCell>状态</TableCell>
                      <TableCell>截止时间</TableCell>
                      <TableCell>创建时间</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {selectedInstance?.tasks.map((task) => (
                      <TableRow key={task.id}>
                        <TableCell>{task.id}</TableCell>
                        <TableCell>{task.stepName}</TableCell>
                        <TableCell>{task.assigneeName}</TableCell>
                        <TableCell>
                          <Box display="flex" alignItems="center" gap={1}>
                            {getTaskStatusIcon(task.status)}
                            <Typography variant="body2">
                              {task.status === 'pending' ? '待处理' :
                               task.status === 'in_progress' ? '处理中' :
                               task.status === 'completed' ? '已完成' : '已拒绝'}
                            </Typography>
                          </Box>
                        </TableCell>
                        <TableCell>
                          {task.dueDate ? formatDate(task.dueDate) : '-'}
                        </TableCell>
                        <TableCell>{formatDate(task.createdAt)}</TableCell>
                      </TableRow>
                    )) || []}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </TabPanel>

          {/* 执行历史 */}
          <TabPanel value={tabValue} index={2}>
            {selectedInstance?.history.length === 0 ? (
              <Alert severity="info">暂无执行历史</Alert>
            ) : (
              <Timeline>
                {selectedInstance?.history.map((hist) => (
                  <TimelineItem key={hist.id}>
                    <TimelineOppositeContent sx={{ m: 'auto 0' }} variant="body2" color="textSecondary">
                      {formatDate(hist.timestamp)}
                    </TimelineOppositeContent>
                    <TimelineSeparator>
                      <TimelineDot color="primary">
                        {hist.action === 'started' ? <StartIcon /> :
                         hist.action === 'completed' ? <CompleteIcon /> :
                         hist.action === 'failed' ? <ErrorIcon /> :
                         hist.action === 'assigned' ? <PersonIcon /> : 
                         <ScheduleIcon />}
                      </TimelineDot>
                      <TimelineConnector />
                    </TimelineSeparator>
                    <TimelineContent sx={{ py: '12px', px: 2 }}>
                      <Typography variant="h6" component="span">
                        {hist.stepName || hist.action}
                      </Typography>
                      <Typography color="textSecondary">
                        {hist.details} - {hist.actor}
                      </Typography>
                    </TimelineContent>
                  </TimelineItem>
                )) || []}
              </Timeline>
            )}
          </TabPanel>

          {/* 输入输出 */}
          <TabPanel value={tabValue} index={3}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Typography variant="h6" gutterBottom>输入数据</Typography>
                <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
                  <pre>
                    {JSON.stringify(selectedInstance?.input, null, 2)}
                  </pre>
                </Paper>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="h6" gutterBottom>输出数据</Typography>
                <Paper sx={{ p: 2, bgcolor: 'grey.50' }}>
                  <pre>
                    {selectedInstance?.output 
                      ? JSON.stringify(selectedInstance.output, null, 2)
                      : '暂无输出数据'}
                  </pre>
                </Paper>
              </Grid>
            </Grid>
          </TabPanel>
        </DialogContent>
        
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>关闭</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default WorkflowInstanceManagement;