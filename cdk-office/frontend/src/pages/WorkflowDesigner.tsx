'use client';

import React, { useState } from 'react';
import { 
  Play, 
  CheckCircle, 
  Bell, 
  Settings, 
  User, 
  Users, 
  Plus, 
  Edit, 
  Trash2, 
  Save, 
  Eye, 
  EyeOff,
  Workflow as WorkflowIcon
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter,
  DialogDescription
} from '@/components/ui/dialog';
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
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';
import { toast } from '@/components/ui/use-toast';

// 工作流步骤接口
interface WorkflowStep {
  id: string;
  name: string;
  type: 'start' | 'approval' | 'notification' | 'condition' | 'action' | 'end';
  assigneeType: 'user' | 'role' | 'group';
  assignee: string;
  config: {
    timeout?: number;
    retryAttempts?: number;
    condition?: string;
    template?: string;
    action?: string;
    [key: string]: any;
  };
  position: { x: number; y: number };
  order: number;
}

// 工作流定义接口
interface WorkflowDefinition {
  id?: string;
  name: string;
  description: string;
  category: string;
  version: string;
  steps: WorkflowStep[];
  config: {
    timeout: number;
    retryAttempts: number;
    notificationEnabled: boolean;
    escalationEnabled: boolean;
  };
}

const WorkflowDesigner: React.FC = () => {
  const [workflow, setWorkflow] = useState<WorkflowDefinition>({
    name: '',
    description: '',
    category: '',
    version: '1.0.0',
    steps: [],
    config: {
      timeout: 86400,
      retryAttempts: 3,
      notificationEnabled: true,
      escalationEnabled: false,
    },
  });

  const [selectedStep, setSelectedStep] = useState<WorkflowStep | null>(null);
  const [openStepDialog, setOpenStepDialog] = useState(false);
  const [openSaveDialog, setOpenSaveDialog] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [previewMode, setPreviewMode] = useState(false);

  // 添加新步骤
  const handleAddStep = (type: WorkflowStep['type']) => {
    const newStep: WorkflowStep = {
      id: `step_${Date.now()}`,
      name: getStepTypeName(type),
      type,
      assigneeType: 'user',
      assignee: '',
      config: getDefaultStepConfig(type),
      position: { x: 100, y: workflow.steps.length * 100 + 100 },
      order: workflow.steps.length + 1,
    };

    setSelectedStep(newStep);
    setIsEditing(false);
    setOpenStepDialog(true);
  };

  // 编辑步骤
  const handleEditStep = (step: WorkflowStep) => {
    setSelectedStep(step);
    setIsEditing(true);
    setOpenStepDialog(true);
  };

  // 删除步骤
  const handleDeleteStep = (stepId: string) => {
    if (window.confirm('确认删除这个步骤吗？')) {
      setWorkflow(prev => ({
        ...prev,
        steps: prev.steps.filter(step => step.id !== stepId),
      }));
    }
  };

  // 保存步骤
  const handleSaveStep = () => {
    if (!selectedStep) return;

    if (isEditing) {
      setWorkflow(prev => ({
        ...prev,
        steps: prev.steps.map(step => 
          step.id === selectedStep.id ? selectedStep : step
        ),
      }));
    } else {
      setWorkflow(prev => ({
        ...prev,
        steps: [...prev.steps, selectedStep],
      }));
    }

    setOpenStepDialog(false);
    setSelectedStep(null);
  };

  // 保存工作流
  const handleSaveWorkflow = async () => {
    try {
      // 检查必填字段
      if (!workflow.name) {
        toast({
          title: "保存失败",
          description: "请输入工作流名称",
          variant: "destructive",
        });
        return;
      }

      if (workflow.steps.length === 0) {
        toast({
          title: "保存失败",
          description: "请至少添加一个步骤",
          variant: "destructive",
        });
        return;
      }

      // 模拟API调用
      const response = await fetch('/api/v1/workflow/definitions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(workflow),
      });

      if (response.ok) {
        toast({
          title: "保存成功",
          description: "工作流已成功保存",
        });
        setOpenSaveDialog(false);
      } else {
        throw new Error('保存失败');
      }
    } catch (error) {
      console.error('保存工作流失败:', error);
      toast({
        title: "保存失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  // 预览工作流
  const handlePreviewWorkflow = () => {
    if (workflow.steps.length === 0) {
      toast({
        title: "无法预览",
        description: "请先添加步骤",
        variant: "destructive",
      });
      return;
    }
    setPreviewMode(!previewMode);
  };

  // 获取步骤类型名称
  const getStepTypeName = (type: WorkflowStep['type']) => {
    switch (type) {
      case 'start': return '开始';
      case 'approval': return '审批步骤';
      case 'notification': return '通知步骤';
      case 'condition': return '条件判断';
      case 'action': return '动作执行';
      case 'end': return '结束';
      default: return '未知步骤';
    }
  };

  // 获取步骤默认配置
  const getDefaultStepConfig = (type: WorkflowStep['type']) => {
    switch (type) {
      case 'approval':
        return { timeout: 86400, retryAttempts: 3 };
      case 'notification':
        return { template: 'default' };
      case 'condition':
        return { condition: '' };
      case 'action':
        return { action: '' };
      default:
        return {};
    }
  };

  // 获取步骤图标
  const getStepIcon = (type: WorkflowStep['type']) => {
    switch (type) {
      case 'start': return <Play className="h-4 w-4" />;
      case 'approval': return <CheckCircle className="h-4 w-4" />;
      case 'notification': return <Bell className="h-4 w-4" />;
      case 'condition': return <Settings className="h-4 w-4" />;
      case 'action': return <Settings className="h-4 w-4" />;
      case 'end': return <CheckCircle className="h-4 w-4" />;
      default: return <WorkflowIcon className="h-4 w-4" />;
    }
  };

  // 获取步骤类型颜色
  const getStepTypeColor = (type: WorkflowStep['type']) => {
    switch (type) {
      case 'start': return 'default';
      case 'approval': return 'default';
      case 'notification': return 'secondary';
      case 'condition': return 'outline';
      case 'action': return 'secondary';
      case 'end': return 'destructive';
      default: return 'default';
    }
  };

  // 获取处理人类型图标
  const getAssigneeTypeIcon = (assigneeType: string) => {
    switch (assigneeType) {
      case 'user': return <User className="h-4 w-4" />;
      case 'role': return <Settings className="h-4 w-4" />;
      case 'group': return <Users className="h-4 w-4" />;
      default: return <User className="h-4 w-4" />;
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和工具栏 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">工作流设计器</h1>
          <p className="text-muted-foreground mt-2">
            设计和管理工作流流程
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handlePreviewWorkflow}
            disabled={workflow.steps.length === 0}
          >
            {previewMode ? (
              <>
                <EyeOff className="h-4 w-4 mr-2" />
                编辑模式
              </>
            ) : (
              <>
                <Eye className="h-4 w-4 mr-2" />
                预览模式
              </>
            )}
          </Button>
          <Button
            onClick={() => setOpenSaveDialog(true)}
            disabled={workflow.steps.length === 0}
          >
            <Save className="h-4 w-4 mr-2" />
            保存工作流
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* 左侧面板 - 步骤库和工作流配置 */}
        <div className="lg:col-span-1 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>步骤库</CardTitle>
            </CardHeader>
            <CardContent>
              {!previewMode && (
                <div className="space-y-2">
                  <Button
                    variant="outline"
                    className="w-full justify-start"
                    onClick={() => handleAddStep('start')}
                  >
                    <Play className="h-4 w-4 mr-2 text-green-500" />
                    开始
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full justify-start"
                    onClick={() => handleAddStep('approval')}
                  >
                    <CheckCircle className="h-4 w-4 mr-2 text-blue-500" />
                    审批步骤
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full justify-start"
                    onClick={() => handleAddStep('notification')}
                  >
                    <Bell className="h-4 w-4 mr-2 text-purple-500" />
                    通知步骤
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full justify-start"
                    onClick={() => handleAddStep('condition')}
                  >
                    <Settings className="h-4 w-4 mr-2 text-yellow-500" />
                    条件判断
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full justify-start"
                    onClick={() => handleAddStep('action')}
                  >
                    <Settings className="h-4 w-4 mr-2 text-indigo-500" />
                    动作执行
                  </Button>
                  <Button
                    variant="outline"
                    className="w-full justify-start"
                    onClick={() => handleAddStep('end')}
                  >
                    <CheckCircle className="h-4 w-4 mr-2 text-red-500" />
                    结束
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>

          {/* 工作流信息 */}
          <Card>
            <CardHeader>
              <CardTitle>工作流信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">步骤数:</span>
                <span>{workflow.steps.length}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">版本:</span>
                <span>{workflow.version}</span>
              </div>
              {workflow.name && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">名称:</span>
                  <span>{workflow.name}</span>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* 中间画布 - 工作流设计区域 */}
        <div className="lg:col-span-3">
          <Card className="h-[600px]">
            <CardContent className="h-full p-0">
              {workflow.steps.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-center p-8">
                  <WorkflowIcon className="h-16 w-16 text-muted-foreground mb-4" />
                  <h3 className="text-xl font-semibold mb-2">工作流设计画布</h3>
                  <p className="text-muted-foreground">
                    从左侧步骤库中添加工作流步骤
                  </p>
                  {!previewMode && (
                    <Button 
                      className="mt-4" 
                      onClick={() => handleAddStep('approval')}
                    >
                      <Plus className="h-4 w-4 mr-2" />
                      添加第一个步骤
                    </Button>
                  )}
                </div>
              ) : (
                <ScrollArea className="h-full p-4">
                  {previewMode ? (
                    // 预览模式 - 使用步骤列表显示
                    <div className="space-y-4">
                      {workflow.steps.map((step, index) => (
                        <div key={step.id} className="border rounded-lg p-4">
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center">
                              <div className="mr-2">
                                {getStepIcon(step.type)}
                              </div>
                              <h4 className="font-semibold">{step.name}</h4>
                            </div>
                            <Badge variant={getStepTypeColor(step.type)}>
                              {getStepTypeName(step.type)}
                            </Badge>
                          </div>
                          
                          <div className="text-sm text-muted-foreground mb-2">
                            步骤 {index + 1}
                          </div>
                          
                          {step.assignee && (
                            <div className="flex items-center text-sm mb-1">
                              <div className="mr-2">
                                {getAssigneeTypeIcon(step.assigneeType)}
                              </div>
                              <span>
                                {step.assignee} ({step.assigneeType === 'user' ? '用户' : 
                                                step.assigneeType === 'role' ? '角色' : '用户组'})
                              </span>
                            </div>
                          )}
                          
                          {step.config.timeout && (
                            <div className="text-sm text-muted-foreground">
                              超时时间: {step.config.timeout}秒
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  ) : (
                    // 编辑模式 - 显示可编辑的步骤卡片
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {workflow.steps.map((step, index) => (
                        <Card key={step.id}>
                          <CardHeader className="pb-2">
                            <div className="flex items-center justify-between">
                              <div className="flex items-center">
                                <div className="mr-2">
                                  {getStepIcon(step.type)}
                                </div>
                                <CardTitle className="text-lg">{step.name}</CardTitle>
                              </div>
                              <Badge variant={getStepTypeColor(step.type)}>
                                {getStepTypeName(step.type)}
                              </Badge>
                            </div>
                            <div className="text-sm text-muted-foreground">
                              步骤 {index + 1}
                            </div>
                          </CardHeader>
                          <CardContent className="pb-2">
                            {step.assignee && (
                              <div className="flex items-center text-sm mb-2">
                                <div className="mr-2">
                                  {getAssigneeTypeIcon(step.assigneeType)}
                                </div>
                                <span>
                                  {step.assignee} ({step.assigneeType === 'user' ? '用户' : 
                                                  step.assigneeType === 'role' ? '角色' : '用户组'})
                                </span>
                              </div>
                            )}
                          </CardContent>
                          <CardFooter className="flex justify-end space-x-2 p-2">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleEditStep(step)}
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleDeleteStep(step.id)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </CardFooter>
                        </Card>
                      ))}
                    </div>
                  )}
                </ScrollArea>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* 步骤配置对话框 */}
      <Dialog open={openStepDialog} onOpenChange={setOpenStepDialog}>
        <DialogContent className="max-w-2xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>
              {isEditing ? '编辑步骤' : '添加步骤'} - {selectedStep && getStepTypeName(selectedStep.type)}
            </DialogTitle>
          </DialogHeader>
          
          {selectedStep && (
            <ScrollArea className="max-h-[60vh] pr-4">
              <div className="py-4 space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="step-name">步骤名称</Label>
                  <Input
                    id="step-name"
                    value={selectedStep.name}
                    onChange={(e) => setSelectedStep({...selectedStep, name: e.target.value})}
                    placeholder="输入步骤名称"
                  />
                </div>

                {(selectedStep.type === 'approval' || selectedStep.type === 'notification') && (
                  <>
                    <div className="space-y-2">
                      <Label>处理人类型</Label>
                      <Select
                        value={selectedStep.assigneeType}
                        onValueChange={(value) => setSelectedStep({
                          ...selectedStep, 
                          assigneeType: value as 'user' | 'role' | 'group'
                        })}
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="user">用户</SelectItem>
                          <SelectItem value="role">角色</SelectItem>
                          <SelectItem value="group">用户组</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="assignee">处理人</Label>
                      <Input
                        id="assignee"
                        value={selectedStep.assignee}
                        onChange={(e) => setSelectedStep({...selectedStep, assignee: e.target.value})}
                        placeholder={
                          selectedStep.assigneeType === 'user' ? '输入用户名或用户ID' :
                          selectedStep.assigneeType === 'role' ? '输入角色名称' :
                          '输入用户组名称'
                        }
                      />
                    </div>
                  </>
                )}

                {selectedStep.type === 'approval' && (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="timeout">超时时间（秒）</Label>
                      <Input
                        id="timeout"
                        type="number"
                        value={selectedStep.config.timeout || 86400}
                        onChange={(e) => setSelectedStep({
                          ...selectedStep,
                          config: {...selectedStep.config, timeout: parseInt(e.target.value) || 0}
                        })}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="retryAttempts">重试次数</Label>
                      <Input
                        id="retryAttempts"
                        type="number"
                        value={selectedStep.config.retryAttempts || 3}
                        onChange={(e) => setSelectedStep({
                          ...selectedStep,
                          config: {...selectedStep.config, retryAttempts: parseInt(e.target.value) || 0}
                        })}
                      />
                    </div>
                  </>
                )}

                {selectedStep.type === 'notification' && (
                  <div className="space-y-2">
                    <Label htmlFor="template">通知模板</Label>
                    <Input
                      id="template"
                      value={selectedStep.config.template || 'default'}
                      onChange={(e) => setSelectedStep({
                        ...selectedStep,
                        config: {...selectedStep.config, template: e.target.value}
                      })}
                      placeholder="输入模板名称"
                    />
                  </div>
                )}

                {selectedStep.type === 'condition' && (
                  <div className="space-y-2">
                    <Label htmlFor="condition">条件表达式</Label>
                    <Textarea
                      id="condition"
                      value={selectedStep.config.condition || ''}
                      onChange={(e) => setSelectedStep({
                        ...selectedStep,
                        config: {...selectedStep.config, condition: e.target.value}
                      })}
                      placeholder="输入条件判断逻辑"
                      rows={3}
                    />
                  </div>
                )}

                {selectedStep.type === 'action' && (
                  <div className="space-y-2">
                    <Label htmlFor="action">动作配置</Label>
                    <Textarea
                      id="action"
                      value={selectedStep.config.action || ''}
                      onChange={(e) => setSelectedStep({
                        ...selectedStep,
                        config: {...selectedStep.config, action: e.target.value}
                      })}
                      placeholder="输入动作执行配置"
                      rows={3}
                    />
                  </div>
                )}
              </div>
            </ScrollArea>
          )}
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenStepDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleSaveStep}
              disabled={!selectedStep?.name}
            >
              {isEditing ? '更新' : '添加'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 保存工作流对话框 */}
      <Dialog open={openSaveDialog} onOpenChange={setOpenSaveDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>保存工作流</DialogTitle>
            <DialogDescription>
              请输入工作流的基本信息
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="workflow-name">工作流名称 *</Label>
              <Input
                id="workflow-name"
                value={workflow.name}
                onChange={(e) => setWorkflow({...workflow, name: e.target.value})}
                placeholder="输入工作流名称"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">描述</Label>
              <Textarea
                id="description"
                value={workflow.description}
                onChange={(e) => setWorkflow({...workflow, description: e.target.value})}
                placeholder="输入工作流描述"
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="category">分类</Label>
              <Input
                id="category"
                value={workflow.category}
                onChange={(e) => setWorkflow({...workflow, category: e.target.value})}
                placeholder="输入分类"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="version">版本</Label>
              <Input
                id="version"
                value={workflow.version}
                onChange={(e) => setWorkflow({...workflow, version: e.target.value})}
                placeholder="输入版本号"
              />
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenSaveDialog(false)}>
              取消
            </Button>
            <Button 
              onClick={handleSaveWorkflow}
              disabled={!workflow.name || workflow.steps.length === 0}
            >
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default WorkflowDesigner;