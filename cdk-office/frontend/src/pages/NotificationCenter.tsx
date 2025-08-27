'use client';

import React, { useState, useEffect } from 'react';
import { 
  RefreshCw, 
  Eye, 
  Trash2, 
  Archive, 
  Mail, 
  Bell, 
  MoreVertical, 
  Settings, 
  Smartphone,
  Monitor,
  Volume2
} from 'lucide-react';

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
import { Checkbox } from '@/components/ui/checkbox';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Progress } from '@/components/ui/progress';
import { toast } from '@/components/ui/use-toast';
import { Separator } from '@/components/ui/separator';
import { ScrollArea } from '@/components/ui/scroll-area';

// 通知数据接口
interface Notification {
  id: string;
  title: string;
  content: string;
  type: 'system' | 'approval' | 'document' | 'task' | 'mention';
  category: 'general' | 'urgent' | 'important';
  priority: 'low' | 'normal' | 'high' | 'urgent';
  isRead: boolean;
  isArchived: boolean;
  createdAt: string;
  relatedId?: string;
  relatedType?: string;
  actionRequired: boolean;
}

// 通知偏好设置接口
interface NotificationPreference {
  id: string;
  userId: string;
  emailEnabled: boolean;
  emailFrequency: 'immediately' | 'daily' | 'weekly';
  pushEnabled: boolean;
  inAppEnabled: boolean;
  smsEnabled: boolean;
  desktopEnabled: boolean;
  soundEnabled: boolean;
}

const NotificationCenter: React.FC = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedNotification, setSelectedNotification] = useState<Notification | null>(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [dialogAction, setDialogAction] = useState<'view' | 'settings'>('view');
  const [preference, setPreference] = useState<NotificationPreference | null>(null);
  const [tabValue, setTabValue] = useState('all');
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);

  // 模拟获取通知数据
  useEffect(() => {
    const fetchNotifications = async () => {
      try {
        // 模拟API调用
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        const mockData: Notification[] = [
          {
            id: '1',
            title: '审批请求',
            content: '您有一个新的文档审批请求需要处理',
            type: 'approval',
            category: 'important',
            priority: 'high',
            isRead: false,
            isArchived: false,
            createdAt: '2023-12-01T10:30:00Z',
            relatedId: 'approval-1',
            relatedType: 'approval',
            actionRequired: true,
          },
          {
            id: '2',
            title: '文档更新',
            content: '您关注的文档已更新',
            type: 'document',
            category: 'general',
            priority: 'normal',
            isRead: true,
            isArchived: false,
            createdAt: '2023-11-28T09:15:00Z',
            relatedId: 'doc-1',
            relatedType: 'document',
            actionRequired: false,
          },
          {
            id: '3',
            title: '系统维护通知',
            content: '系统将于今晚00:00-02:00进行维护',
            type: 'system',
            category: 'urgent',
            priority: 'urgent',
            isRead: false,
            isArchived: false,
            createdAt: '2023-12-02T14:20:00Z',
            actionRequired: false,
          },
          {
            id: '4',
            title: '任务提醒',
            content: '您分配的任务即将到期',
            type: 'task',
            category: 'important',
            priority: 'high',
            isRead: true,
            isArchived: false,
            createdAt: '2023-11-30T16:45:00Z',
            relatedId: 'task-1',
            relatedType: 'task',
            actionRequired: true,
          },
        ];
        
        setNotifications(mockData);
        setUnreadCount(mockData.filter(n => !n.isRead).length);
        setLoading(false);
      } catch (error) {
        console.error('获取通知数据失败:', error);
        setLoading(false);
      }
    };

    fetchNotifications();
  }, []);

  // 处理查看通知详情
  const handleViewNotification = (notification: Notification) => {
    setSelectedNotification(notification);
    setDialogAction('view');
    setOpenDialog(true);
  };

  // 处理打开设置对话框
  const handleOpenSettings = () => {
    // 模拟获取用户偏好设置
    const mockPreference: NotificationPreference = {
      id: 'pref-1',
      userId: 'user-1',
      emailEnabled: true,
      emailFrequency: 'immediately',
      pushEnabled: true,
      inAppEnabled: true,
      smsEnabled: false,
      desktopEnabled: true,
      soundEnabled: true,
    };
    setPreference(mockPreference);
    setDialogAction('settings');
    setOpenDialog(true);
  };

  // 处理保存设置
  const handleSaveSettings = async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));
      setOpenDialog(false);
      toast({
        title: "设置保存成功",
        description: "通知偏好设置已更新",
      });
    } catch (error) {
      console.error('保存设置失败:', error);
      toast({
        title: "保存设置失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  // 获取通知类型标签
  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'system': return '系统';
      case 'approval': return '审批';
      case 'document': return '文档';
      case 'task': return '任务';
      case 'mention': return '提及';
      default: return type;
    }
  };

  // 获取通知类型颜色
  const getTypeColor = (type: string) => {
    switch (type) {
      case 'system': return 'default';
      case 'approval': return 'warning';
      case 'document': return 'default';
      case 'task': return 'secondary';
      case 'mention': return 'success';
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
      case 'normal': return 'secondary';
      case 'high': return 'warning';
      case 'urgent': return 'destructive';
      default: return 'default';
    }
  };

  // 获取状态颜色
  const getStatusColor = (isRead: boolean) => {
    return isRead ? 'success' : 'warning';
  };

  // 获取状态文本
  const getStatusText = (isRead: boolean) => {
    return isRead ? '已读' : '未读';
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

  // 处理选择通知
  const handleSelectNotification = (id: string, checked: boolean) => {
    if (checked) {
      setSelectedIds(prev => [...prev, id]);
    } else {
      setSelectedIds(prev => prev.filter(itemId => itemId !== id));
    }
  };

  // 处理全选
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      const filteredNotifications = notifications.filter(n => {
        if (tabValue === 'unread') return !n.isRead; // 未读标签页
        if (tabValue === 'archived') return n.isArchived; // 已归档标签页
        return !n.isArchived; // 全部通知标签页（不包括已归档）
      });
      setSelectedIds(filteredNotifications.map(n => n.id));
    } else {
      setSelectedIds([]);
    }
  };

  // 标记为已读
  const handleMarkAsRead = async (id?: string) => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      if (id) {
        // 标记单个通知为已读
        setNotifications(prev => prev.map(n => 
          n.id === id ? { ...n, isRead: true } : n
        ));
      } else if (selectedIds.length > 0) {
        // 批量标记为已读
        setNotifications(prev => prev.map(n => 
          selectedIds.includes(n.id) ? { ...n, isRead: true } : n
        ));
        setSelectedIds([]);
      }
      
      // 更新未读数量
      setUnreadCount(prev => Math.max(0, prev - (id ? 1 : selectedIds.length)));
      
      toast({
        title: "操作成功",
        description: id ? "通知已标记为已读" : "选中的通知已标记为已读",
      });
    } catch (error) {
      console.error('标记为已读失败:', error);
      toast({
        title: "操作失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  // 标记为未读
  const handleMarkAsUnread = async (id: string) => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      setNotifications(prev => prev.map(n => 
        n.id === id ? { ...n, isRead: false } : n
      ));
      
      // 更新未读数量
      setUnreadCount(prev => prev + 1);
      
      toast({
        title: "操作成功",
        description: "通知已标记为未读",
      });
    } catch (error) {
      console.error('标记为未读失败:', error);
      toast({
        title: "操作失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  // 归档通知
  const handleArchive = async (id?: string) => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      if (id) {
        // 归档单个通知
        setNotifications(prev => prev.map(n => 
          n.id === id ? { ...n, isArchived: true } : n
        ));
      } else if (selectedIds.length > 0) {
        // 批量归档
        setNotifications(prev => prev.map(n => 
          selectedIds.includes(n.id) ? { ...n, isArchived: true } : n
        ));
        setSelectedIds([]);
      }
      
      toast({
        title: "操作成功",
        description: id ? "通知已归档" : "选中的通知已归档",
      });
    } catch (error) {
      console.error('归档失败:', error);
      toast({
        title: "操作失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  // 删除通知
  const handleDelete = async (id?: string) => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      if (id) {
        // 删除单个通知
        setNotifications(prev => prev.filter(n => n.id !== id));
      } else if (selectedIds.length > 0) {
        // 批量删除
        setNotifications(prev => prev.filter(n => !selectedIds.includes(n.id)));
        setSelectedIds([]);
      }
      
      toast({
        title: "操作成功",
        description: id ? "通知已删除" : "选中的通知已删除",
      });
    } catch (error) {
      console.error('删除失败:', error);
      toast({
        title: "操作失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  // 标记所有为已读
  const handleMarkAllAsRead = async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 500));
      
      setNotifications(prev => prev.map(n => ({ ...n, isRead: true })));
      setUnreadCount(0);
      setSelectedIds([]);
      
      toast({
        title: "操作成功",
        description: "所有通知已标记为已读",
      });
    } catch (error) {
      console.error('标记所有为已读失败:', error);
      toast({
        title: "操作失败",
        description: "请重试",
        variant: "destructive",
      });
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Progress className="w-32" />
        <span className="ml-2">加载中...</span>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">通知中心</h1>
          <p className="text-muted-foreground mt-2">
            管理和查看系统通知
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={handleOpenSettings}
          >
            <Settings className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={handleRefresh}
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* 标签页 */}
      <Tabs value={tabValue} onValueChange={setTabValue}>
        <TabsList>
          <TabsTrigger value="all">全部通知 ({notifications.filter(n => !n.isArchived).length})</TabsTrigger>
          <TabsTrigger value="unread">未读 ({unreadCount})</TabsTrigger>
          <TabsTrigger value="archived">已归档 ({notifications.filter(n => n.isArchived).length})</TabsTrigger>
        </TabsList>
        
        <TabsContent value={tabValue} className="mt-6">
          {notifications.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <p>暂无通知</p>
            </div>
          ) : (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-12">
                        <Checkbox
                          checked={selectedIds.length > 0 && selectedIds.length === notifications.filter(n => {
                            if (tabValue === 'unread') return !n.isRead; // 未读标签页
                            if (tabValue === 'archived') return n.isArchived; // 已归档标签页
                            return !n.isArchived; // 全部通知标签页（不包括已归档）
                          }).length}
                          onCheckedChange={(checked) => handleSelectAll(checked as boolean)}
                        />
                      </TableHead>
                      <TableHead>标题</TableHead>
                      <TableHead>类型</TableHead>
                      <TableHead>优先级</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead className="w-16">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {notifications
                      .filter(n => {
                        if (tabValue === 'unread') return !n.isRead; // 未读标签页
                        if (tabValue === 'archived') return n.isArchived; // 已归档标签页
                        return !n.isArchived; // 全部通知标签页（不包括已归档）
                      })
                      .map((notification) => (
                        <TableRow 
                          key={notification.id} 
                          className={selectedIds.includes(notification.id) ? "bg-muted" : ""}
                        >
                          <TableCell>
                            <Checkbox
                              checked={selectedIds.includes(notification.id)}
                              onCheckedChange={(checked) => handleSelectNotification(notification.id, checked as boolean)}
                            />
                          </TableCell>
                          <TableCell>
                            <div 
                              className="font-medium cursor-pointer hover:underline"
                              onClick={() => handleViewNotification(notification)}
                            >
                              {notification.title}
                              {notification.actionRequired && (
                                <Badge variant="destructive" className="ml-2">
                                  需处理
                                </Badge>
                              )}
                            </div>
                            <div 
                              className="text-sm text-muted-foreground mt-1 cursor-pointer hover:underline"
                              onClick={() => handleViewNotification(notification)}
                            >
                              {notification.content}
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge variant={getTypeColor(notification.type)}>
                              {getTypeLabel(notification.type)}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant={getPriorityColor(notification.priority)}>
                              {getPriorityLabel(notification.priority)}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant={getStatusColor(notification.isRead)}>
                              {getStatusText(notification.isRead)}
                            </Badge>
                          </TableCell>
                          <TableCell>{formatDate(notification.createdAt)}</TableCell>
                          <TableCell>
                            <Button
                              size="icon"
                              variant="ghost"
                              onClick={() => handleViewNotification(notification)}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {/* 批量操作按钮 */}
          {selectedIds.length > 0 && (
            <div className="flex items-center gap-2 mt-4">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleMarkAsRead()}
              >
                <Mail className="h-4 w-4 mr-2" />
                标记为已读
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleArchive()}
              >
                <Archive className="h-4 w-4 mr-2" />
                归档
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleDelete()}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                删除
              </Button>
              <div className="text-sm text-muted-foreground ml-2">
                已选择 {selectedIds.length} 项
              </div>
            </div>
          )}
        </TabsContent>
      </Tabs>

      {/* 通知详情和设置对话框 */}
      <Dialog open={openDialog} onOpenChange={setOpenDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {dialogAction === 'view' ? '通知详情' : '通知设置'}
            </DialogTitle>
          </DialogHeader>
          
          <ScrollArea className="max-h-[60vh]">
            {dialogAction === 'view' && selectedNotification && (
              <div className="py-4 space-y-4">
                <div>
                  <h3 className="text-xl font-semibold">{selectedNotification.title}</h3>
                </div>
                <div className="text-muted-foreground">
                  {selectedNotification.content}
                </div>
                <Separator />
                <div className="space-y-2">
                  <Label>类型</Label>
                  <Badge variant={getTypeColor(selectedNotification.type)}>
                    {getTypeLabel(selectedNotification.type)}
                  </Badge>
                </div>
                <div className="space-y-2">
                  <Label>优先级</Label>
                  <Badge variant={getPriorityColor(selectedNotification.priority)}>
                    {getPriorityLabel(selectedNotification.priority)}
                  </Badge>
                </div>
                <div className="space-y-2">
                  <Label>状态</Label>
                  <Badge variant={getStatusColor(selectedNotification.isRead)}>
                    {getStatusText(selectedNotification.isRead)}
                  </Badge>
                </div>
                <div className="space-y-2">
                  <Label>创建时间</Label>
                  <p>{formatDate(selectedNotification.createdAt)}</p>
                </div>
                {selectedNotification.relatedId && (
                  <div className="space-y-2">
                    <Label>关联ID</Label>
                    <p>{selectedNotification.relatedId}</p>
                  </div>
                )}
              </div>
            )}
            
            {dialogAction === 'settings' && preference && (
              <div className="py-4 space-y-4">
                <h3 className="text-xl font-semibold">通知偏好设置</h3>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="inAppEnabled"
                    checked={preference.inAppEnabled}
                    onCheckedChange={(checked) => setPreference({...preference, inAppEnabled: checked as boolean})}
                  />
                  <Label htmlFor="inAppEnabled" className="flex items-center">
                    <Bell className="h-4 w-4 mr-2" />
                    应用内通知
                  </Label>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="emailEnabled"
                    checked={preference.emailEnabled}
                    onCheckedChange={(checked) => setPreference({...preference, emailEnabled: checked as boolean})}
                  />
                  <Label htmlFor="emailEnabled" className="flex items-center">
                    <Mail className="h-4 w-4 mr-2" />
                    邮件通知
                  </Label>
                </div>
                
                {preference.emailEnabled && (
                  <div className="space-y-2 ml-6">
                    <Label htmlFor="emailFrequency">邮件通知频率</Label>
                    <Select
                      value={preference.emailFrequency}
                      onValueChange={(value) => setPreference({...preference, emailFrequency: value as any})}
                    >
                      <SelectTrigger id="emailFrequency">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="immediately">立即发送</SelectItem>
                        <SelectItem value="daily">每日摘要</SelectItem>
                        <SelectItem value="weekly">每周摘要</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                )}
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="pushEnabled"
                    checked={preference.pushEnabled}
                    onCheckedChange={(checked) => setPreference({...preference, pushEnabled: checked as boolean})}
                  />
                  <Label htmlFor="pushEnabled" className="flex items-center">
                    <Smartphone className="h-4 w-4 mr-2" />
                    推送通知
                  </Label>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="desktopEnabled"
                    checked={preference.desktopEnabled}
                    onCheckedChange={(checked) => setPreference({...preference, desktopEnabled: checked as boolean})}
                  />
                  <Label htmlFor="desktopEnabled" className="flex items-center">
                    <Monitor className="h-4 w-4 mr-2" />
                    桌面通知
                  </Label>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="smsEnabled"
                    checked={preference.smsEnabled}
                    onCheckedChange={(checked) => setPreference({...preference, smsEnabled: checked as boolean})}
                  />
                  <Label htmlFor="smsEnabled" className="flex items-center">
                    <Smartphone className="h-4 w-4 mr-2" />
                    短信通知
                  </Label>
                </div>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="soundEnabled"
                    checked={preference.soundEnabled}
                    onCheckedChange={(checked) => setPreference({...preference, soundEnabled: checked as boolean})}
                  />
                  <Label htmlFor="soundEnabled" className="flex items-center">
                    <Volume2 className="h-4 w-4 mr-2" />
                    声音提醒
                  </Label>
                </div>
              </div>
            )}
          </ScrollArea>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpenDialog(false)}>
              关闭
            </Button>
            {dialogAction === 'settings' && (
              <Button onClick={handleSaveSettings}>保存</Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default NotificationCenter;