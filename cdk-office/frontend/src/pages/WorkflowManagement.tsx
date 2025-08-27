'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
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
  CardActions,
  Grid,
  Tabs,
  Tab,
  FormControlLabel,
  Switch,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Visibility as ViewIcon,
  PlayArrow as StartIcon,
  Stop as StopIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';

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
    if (!window.confirm(`确认删除工作流"${workflow.name}"吗？`)) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/workflow/definitions/${workflow.id}`, {
        method: 'DELETE',
      });
      
      if (response.ok) {
        setWorkflows(prev => prev.filter(w => w.id !== workflow.id));
        alert('工作流删除成功');
      } else {
        throw new Error('删除失败');
      }
    } catch (error) {
      console.error('删除工作流失败:', error);
      alert('删除工作流失败，请重试');
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
        alert(`工作流已${newStatus === 'active' ? '启用' : '停用'}`);
      } else {
        throw new Error('状态更新失败');
      }
    } catch (error) {
      console.error('更新工作流状态失败:', error);
      alert('更新状态失败，请重试');
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
        alert(`工作流${dialogType === 'create' ? '创建' : '更新'}成功`);
      } else {
        throw new Error('保存失败');
      }
    } catch (error) {
      console.error('保存工作流失败:', error);
      alert('保存失败，请重试');
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
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ width: '100%', p: 3 }}>
      {/* 页面标题和操作按钮 */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          工作流管理
        </Typography>
        <Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={handleCreate}
            sx={{ mr: 2 }}
          >
            创建工作流
          </Button>
          <IconButton onClick={fetchWorkflows} color="primary">
            <RefreshIcon />
          </IconButton>
        </Box>
      </Box>

      {/* 工作流列表 */}
      {workflows.length === 0 ? (
        <Alert severity="info">暂无工作流定义</Alert>
      ) : (
        <Grid container spacing={3}>
          {workflows.map((workflow) => (
            <Grid item xs={12} md={6} lg={4} key={workflow.id}>
              <Card>
                <CardContent>
                  <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                    <Typography variant="h6" component="h2" gutterBottom>
                      {workflow.name}
                    </Typography>
                    <Chip 
                      label={getStatusLabel(workflow.status)}
                      color={getStatusColor(workflow.status)}
                      size="small"
                    />
                  </Box>
                  
                  <Typography color="textSecondary" gutterBottom>
                    {workflow.category}
                  </Typography>
                  
                  <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
                    {workflow.description}
                  </Typography>

                  <Box sx={{ mb: 2 }}>
                    <Typography variant="caption" color="textSecondary">
                      版本: {workflow.version} | 创建者: {workflow.createdBy}
                    </Typography>
                  </Box>

                  {/* 统计信息 */}
                  <Box sx={{ mb: 2, p: 1, bgcolor: 'grey.50', borderRadius: 1 }}>
                    <Grid container spacing={2}>
                      <Grid item xs={6}>
                        <Typography variant="caption" color="textSecondary">
                          总实例: {workflow.statistics.totalInstances}
                        </Typography>
                      </Grid>
                      <Grid item xs={6}>
                        <Typography variant="caption" color="textSecondary">
                          活跃: {workflow.statistics.activeInstances}
                        </Typography>
                      </Grid>
                      <Grid item xs={6}>
                        <Typography variant="caption" color="textSecondary">
                          完成: {workflow.statistics.completedInstances}
                        </Typography>
                      </Grid>
                      <Grid item xs={6}>
                        <Typography variant="caption" color="textSecondary">
                          成功率: {workflow.statistics.successRate}%
                        </Typography>
                      </Grid>
                    </Grid>
                  </Box>

                  <Typography variant="caption" color="textSecondary">
                    最后更新: {formatDate(workflow.updatedAt)}
                  </Typography>
                </CardContent>
                
                <CardActions>
                  <Tooltip title="查看详情">
                    <IconButton size="small" onClick={() => handleView(workflow)}>
                      <ViewIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="编辑">
                    <IconButton size="small" onClick={() => handleEdit(workflow)}>
                      <EditIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title={workflow.status === 'active' ? '停用' : '启用'}>
                    <IconButton 
                      size="small" 
                      onClick={() => handleToggleStatus(workflow)}
                      color={workflow.status === 'active' ? 'error' : 'success'}
                    >
                      {workflow.status === 'active' ? <StopIcon /> : <StartIcon />}
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="删除">
                    <IconButton 
                      size="small" 
                      color="error" 
                      onClick={() => handleDelete(workflow)}
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Tooltip>
                </CardActions>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* 工作流详情/编辑对话框 */}
      <Dialog 
        open={openDialog} 
        onClose={() => setOpenDialog(false)} 
        maxWidth="lg" 
        fullWidth
      >
        <DialogTitle>
          {dialogType === 'create' ? '创建工作流' : 
           dialogType === 'edit' ? '编辑工作流' : '工作流详情'}
        </DialogTitle>
        
        <DialogContent sx={{ p: 0 }}>
          <Tabs value={tabValue} onChange={(_, newValue) => setTabValue(newValue)}>
            <Tab label="基本信息" />
            <Tab label="工作流步骤" />
            <Tab label="配置设置" />
          </Tabs>

          {/* 基本信息标签页 */}
          <TabPanel value={tabValue} index={0}>
            <Box sx={{ p: 2 }}>
              <TextField
                fullWidth
                label="工作流名称"
                value={dialogType === 'view' ? selectedWorkflow?.name || '' : formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                disabled={dialogType === 'view'}
                sx={{ mb: 2 }}
              />
              
              <TextField
                fullWidth
                label="描述"
                multiline
                rows={3}
                value={dialogType === 'view' ? selectedWorkflow?.description || '' : formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                disabled={dialogType === 'view'}
                sx={{ mb: 2 }}
              />
              
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <TextField
                    fullWidth
                    label="分类"
                    value={dialogType === 'view' ? selectedWorkflow?.category || '' : formData.category}
                    onChange={(e) => setFormData(prev => ({ ...prev, category: e.target.value }))}
                    disabled={dialogType === 'view'}
                  />
                </Grid>
                <Grid item xs={6}>
                  <FormControl fullWidth disabled={dialogType === 'view'}>
                    <InputLabel>状态</InputLabel>
                    <Select
                      value={dialogType === 'view' ? selectedWorkflow?.status || '' : formData.status}
                      label="状态"
                      onChange={(e) => setFormData(prev => ({ ...prev, status: e.target.value as any }))}
                    >
                      <MenuItem value="draft">草稿</MenuItem>
                      <MenuItem value="active">活跃</MenuItem>
                      <MenuItem value="inactive">停用</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
              </Grid>
            </Box>
          </TabPanel>

          {/* 工作流步骤标签页 */}
          <TabPanel value={tabValue} index={1}>
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>工作流步骤</Typography>
              <Alert severity="info" sx={{ mb: 2 }}>
                工作流步骤配置功能正在开发中，请使用API进行详细配置
              </Alert>
            </Box>
          </TabPanel>

          {/* 配置设置标签页 */}
          <TabPanel value={tabValue} index={2}>
            <Box sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>配置设置</Typography>
              
              <Grid container spacing={2}>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="超时时间（秒）"
                    type="number"
                    value={dialogType === 'view' ? selectedWorkflow?.config.timeout || 0 : formData.config.timeout}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      config: { ...prev.config, timeout: parseInt(e.target.value) }
                    }))}
                    disabled={dialogType === 'view'}
                    sx={{ mb: 2 }}
                  />
                </Grid>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="重试次数"
                    type="number"
                    value={dialogType === 'view' ? selectedWorkflow?.config.retryAttempts || 0 : formData.config.retryAttempts}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      config: { ...prev.config, retryAttempts: parseInt(e.target.value) }
                    }))}
                    disabled={dialogType === 'view'}
                    sx={{ mb: 2 }}
                  />
                </Grid>
              </Grid>

              <FormControlLabel
                control={
                  <Switch
                    checked={dialogType === 'view' ? selectedWorkflow?.config.notificationEnabled || false : formData.config.notificationEnabled}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      config: { ...prev.config, notificationEnabled: e.target.checked }
                    }))}
                    disabled={dialogType === 'view'}
                  />
                }
                label="启用通知"
                sx={{ mb: 2 }}
              />

              <br />

              <FormControlLabel
                control={
                  <Switch
                    checked={dialogType === 'view' ? selectedWorkflow?.config.escalationEnabled || false : formData.config.escalationEnabled}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      config: { ...prev.config, escalationEnabled: e.target.checked }
                    }))}
                    disabled={dialogType === 'view'}
                  />
                }
                label="启用升级"
                sx={{ mb: 2 }}
              />
            </Box>
          </TabPanel>
        </DialogContent>
        
        <DialogActions>
          <Button onClick={() => setOpenDialog(false)}>取消</Button>
          {dialogType !== 'view' && (
            <Button variant="contained" onClick={handleSave}>
              {dialogType === 'create' ? '创建' : '保存'}
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default WorkflowManagement;