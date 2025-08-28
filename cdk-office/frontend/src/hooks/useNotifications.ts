// 通知轮询 Hook

import { useState, useEffect, useCallback } from 'react';
import { Notification } from '@/types/dashboard';
import { notificationAPI } from '@/lib/api/dashboard';
import { toast } from '@/components/ui/use-toast';

interface UseNotificationsOptions {
  pollingInterval?: number; // 轮询间隔，默认60秒
  limit?: number; // 获取通知数量限制，默认10条
  unreadOnly?: boolean; // 是否只获取未读通知，默认false
  autoMarkAsRead?: boolean; // 是否自动标记已读，默认false
}

export function useNotifications(options: UseNotificationsOptions = {}) {
  const {
    pollingInterval = 60000, // 60秒
    limit = 10,
    unreadOnly = false,
    autoMarkAsRead = false,
  } = options;

  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [unreadCount, setUnreadCount] = useState(0);

  // 获取通知列表
  const fetchNotifications = useCallback(async () => {
    try {
      setError(null);
      const data = await notificationAPI.getAll(limit, unreadOnly);
      setNotifications(data);
      
      // 计算未读通知数量
      const unreadNotifications = data.filter(n => !n.is_read);
      setUnreadCount(unreadNotifications.length);
      
      // 检查是否有新的日程提醒
      const calendarReminders = data.filter(
        n => n.type === 'calendar_reminder' && !n.is_read
      );
      
      // 为日程提醒显示 Toast 通知
      calendarReminders.forEach(notification => {
        toast({
          title: notification.title,
          description: notification.content,
          duration: 5000, // 5秒后自动消失
        });
        
        // 如果开启自动标记已读，则标记该通知为已读
        if (autoMarkAsRead) {
          markAsRead(notification.id);
        }
      });
      
    } catch (err) {
      console.error('获取通知失败:', err);
      setError(err instanceof Error ? err.message : '获取通知失败');
    } finally {
      setLoading(false);
    }
  }, [limit, unreadOnly, autoMarkAsRead]);

  // 标记通知为已读
  const markAsRead = useCallback(async (notificationId: string) => {
    try {
      await notificationAPI.markAsRead(notificationId);
      
      // 更新本地状态
      setNotifications(prev => 
        prev.map(notification => 
          notification.id === notificationId 
            ? { ...notification, is_read: true, read_at: new Date().toISOString() }
            : notification
        )
      );
      
      // 更新未读数量
      setUnreadCount(prev => Math.max(0, prev - 1));
      
    } catch (err) {
      console.error('标记通知已读失败:', err);
      toast({
        title: '错误',
        description: '标记通知已读失败，请重试',
        variant: 'destructive',
      });
    }
  }, []);

  // 批量标记已读
  const markAllAsRead = useCallback(async () => {
    const unreadNotifications = notifications.filter(n => !n.is_read);
    
    try {
      // 并发标记所有未读通知为已读
      await Promise.all(
        unreadNotifications.map(notification => 
          notificationAPI.markAsRead(notification.id)
        )
      );
      
      // 更新本地状态
      setNotifications(prev => 
        prev.map(notification => ({ 
          ...notification, 
          is_read: true, 
          read_at: new Date().toISOString() 
        }))
      );
      
      setUnreadCount(0);
      
      toast({
        title: '成功',
        description: '已标记所有通知为已读',
      });
      
    } catch (err) {
      console.error('批量标记已读失败:', err);
      toast({
        title: '错误',
        description: '批量标记已读失败，请重试',
        variant: 'destructive',
      });
    }
  }, [notifications]);

  // 手动刷新通知
  const refresh = useCallback(() => {
    setLoading(true);
    fetchNotifications();
  }, [fetchNotifications]);

  // 设置轮询
  useEffect(() => {
    // 立即获取一次
    fetchNotifications();

    // 设置定时轮询
    const interval = setInterval(fetchNotifications, pollingInterval);

    // 清理定时器
    return () => clearInterval(interval);
  }, [fetchNotifications, pollingInterval]);

  return {
    notifications,
    loading,
    error,
    unreadCount,
    markAsRead,
    markAllAsRead,
    refresh,
  };
}