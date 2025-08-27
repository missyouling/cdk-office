'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  CardDescription
} from '@/components/ui/card';
import { 
  Button 
} from '@/components/ui/button';
import { 
  Badge 
} from '@/components/ui/badge';
import { 
  toast 
} from '@/components/ui/use-toast';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';

import { 
  ArrowLeft, 
  Mail, 
  MailOpen, 
  Archive, 
  Trash2,
  Bell,
  FileText,
  CheckCircle,
  Clock,
  AlertTriangle
} from 'lucide-react';

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

const NotificationDetail: React.FC = () => {
  const router = useRouter();
  const [notification, setNotification] = useState<Notification | null>(null);
  const [loading, setLoading] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [archiveDialogOpen, setArchiveDialogOpen] = useState(false);

  // 模拟获取通知详情
  useEffect(() => {
    // 从URL参数获取ID，这里简化处理
    const id = '1'; // 实际项目中应该从router.query获取
    
    if (id) {
      const fetchNotification = async () => {
        try {
          // 模拟API调用
          await new Promise(resolve => setTimeout(resolve, 500));
          
          // 模拟数据
          const mockNotification: Notification = {
            id: id,
            title: '审批请求',
            content: '您有一个新的文档审批请求需要处理。文档名称：2023年Q4财务报告.pdf，申请人：张三，提交时间：2023-12-01 10:30。',
            type: 'approval',
            category: 'important',
            priority: 'high',
            isRead: false,
            isArchived: false,
            createdAt: '2023-12-01T10:30:00Z',
            relatedId: 'approval-1',
            relatedType: 'approval',
            actionRequired: true,
          };
          
          setNotification(mockNotification);
          setLoading(false);
        } catch (error) {
          console.error('获取通知详情失败:', error);
          setLoading(false);
        }
      };

      fetchNotification();
    }
  }, []);

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

  // 格式化日期
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('zh-CN');
  };

  // 处理返回
  const handleBack = () => {
    router.push('/notification-center');
  };

  // 标记为已读
  const handleMarkAsRead = async () => {
    if (!notification) return;
    
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      setNotification({ ...notification, isRead: true });
      toast({
        title: "操作成功",
        description: "已标记为已读",
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
  const handleMarkAsUnread = async () => {
    if (!notification) return;
    
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      setNotification({ ...notification, isRead: false });
      toast({
        title: "操作成功",
        description: "已标记为未读",
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
  const handleArchive = async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      toast({
        title: "操作成功",
        description: "通知已归档",
      });
      router.push('/notification-center');
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
  const handleDelete = async () => {
    try {
      // 模拟API调用
      await new Promise(resolve => setTimeout(resolve, 300));
      
      toast({
        title: "操作成功",
        description: "通知已删除",
      });
      router.push('/notification-center');
    } catch (error) {
      console.error('删除失败:', error);
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
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <span className="ml-2">加载中...</span>
      </div>
    );
  }

  if (!notification) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px]">
        <Bell className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-semibold">通知未找到</h3>
        <p className="text-muted-foreground">无法找到指定的通知</p>
        <Button className="mt-4" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          返回通知中心
        </Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和返回按钮 */}
      <div className="flex items-center space-x-4">
        <Button variant="outline" size="icon" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">通知详情</h1>
          <p className="text-muted-foreground">查看和管理通知详情</p>
        </div>
        <div className="ml-auto flex items-center space-x-2">
          {!notification.isRead ? (
            <Button variant="outline" size="sm" onClick={handleMarkAsRead}>
              <MailOpen className="h-4 w-4 mr-2" />
              标记为已读
            </Button>
          ) : (
            <Button variant="outline" size="sm" onClick={handleMarkAsUnread}>
              <Mail className="h-4 w-4 mr-2" />
              标记为未读
            </Button>
          )}
        </div>
      </div>

      {/* 通知详情卡片 */}
      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <div>
              <CardTitle className="flex items-center">
                {notification.title}
                {notification.actionRequired && (
                  <Badge variant="destructive" className="ml-2">
                    需处理
                  </Badge>
                )}
              </CardTitle>
              <CardDescription className="mt-2">
                {notification.content}
              </CardDescription>
            </div>
            <Badge variant={notification.isRead ? "success" : "warning"}>
              {notification.isRead ? (
                <MailOpen className="h-3 w-3 mr-1" />
              ) : (
                <Mail className="h-3 w-3 mr-1" />
              )}
              {notification.isRead ? '已读' : '未读'}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center text-sm">
                <Bell className="h-4 w-4 mr-2 text-muted-foreground" />
                <span className="text-muted-foreground">类型:</span>
                <Badge variant={getTypeColor(notification.type)} className="ml-2">
                  {getTypeLabel(notification.type)}
                </Badge>
              </div>
              <div className="flex items-center text-sm">
                <AlertTriangle className="h-4 w-4 mr-2 text-muted-foreground" />
                <span className="text-muted-foreground">优先级:</span>
                <Badge variant={getPriorityColor(notification.priority)} className="ml-2">
                  {getPriorityLabel(notification.priority)}
                </Badge>
              </div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center text-sm">
                <Clock className="h-4 w-4 mr-2 text-muted-foreground" />
                <span className="text-muted-foreground">创建时间:</span>
                <span className="ml-2">{formatDate(notification.createdAt)}</span>
              </div>
              {notification.relatedId && (
                <div className="flex items-center text-sm">
                  <FileText className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">关联ID:</span>
                  <span className="ml-2">{notification.relatedId}</span>
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-end space-x-2">
            <Button variant="outline" onClick={() => setArchiveDialogOpen(true)}>
              <Archive className="h-4 w-4 mr-2" />
              归档
            </Button>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(true)}>
              <Trash2 className="h-4 w-4 mr-2" />
              删除
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 删除确认对话框 */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除通知</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除通知 "{notification.title}" 吗？此操作无法撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>确认删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 归档确认对话框 */}
      <AlertDialog open={archiveDialogOpen} onOpenChange={setArchiveDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认归档通知</AlertDialogTitle>
            <AlertDialogDescription>
              确定要归档通知 "{notification.title}" 吗？归档后的通知可以在归档标签页中查看。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleArchive}>确认归档</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default NotificationDetail;