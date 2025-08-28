'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { toast } from '@/components/ui/use-toast';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Plus, Calendar, Clock, CalendarDays } from 'lucide-react';
import { format, parseISO, startOfDay, addDays } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { CalendarEvent, CreateCalendarEventRequest } from '@/types/dashboard';
import { calendarAPI } from '@/lib/api/dashboard';

interface CalendarCardProps {
  className?: string;
}

export function CalendarCard({ className }: CalendarCardProps) {
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);

  // 表单状态
  const [formData, setFormData] = useState<CreateCalendarEventRequest>({
    title: '',
    description: '',
    start_time: '',
    end_time: '',
    all_day: false,
  });

  // 加载未来7天的日程
  const loadEvents = async () => {
    try {
      const data = await calendarAPI.getUpcoming();
      setEvents(data);
    } catch (error) {
      console.error('获取日程失败:', error);
      toast({
        title: '错误',
        description: '获取日程失败，请重试',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadEvents();
  }, []);

  // 重置表单
  const resetForm = () => {
    const now = new Date();
    const oneHourLater = new Date(now.getTime() + 60 * 60 * 1000);
    
    setFormData({
      title: '',
      description: '',
      start_time: format(now, "yyyy-MM-dd'T'HH:mm"),
      end_time: format(oneHourLater, "yyyy-MM-dd'T'HH:mm"),
      all_day: false,
    });
  };

  // 打开对话框
  const handleOpenDialog = () => {
    resetForm();
    setIsDialogOpen(true);
  };

  // 创建日程事件
  const handleCreateEvent = async () => {
    if (!formData.title.trim()) {
      toast({
        title: '错误',
        description: '请输入日程标题',
        variant: 'destructive',
      });
      return;
    }

    if (!formData.start_time || !formData.end_time) {
      toast({
        title: '错误',
        description: '请选择开始和结束时间',
        variant: 'destructive',
      });
      return;
    }

    const startDate = new Date(formData.start_time);
    const endDate = new Date(formData.end_time);

    if (endDate <= startDate) {
      toast({
        title: '错误',
        description: '结束时间必须晚于开始时间',
        variant: 'destructive',
      });
      return;
    }

    setIsCreating(true);
    try {
      // 转换为 ISO 格式
      const eventData: CreateCalendarEventRequest = {
        ...formData,
        title: formData.title.trim(),
        start_time: startDate.toISOString(),
        end_time: endDate.toISOString(),
      };

      const newEvent = await calendarAPI.create(eventData);
      setEvents(prev => [newEvent, ...prev].sort((a, b) => 
        new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
      ));
      
      setIsDialogOpen(false);
      toast({
        title: '成功',
        description: '日程已创建',
      });
    } catch (error) {
      console.error('创建日程失败:', error);
      toast({
        title: '错误',
        description: '创建日程失败，请重试',
        variant: 'destructive',
      });
    } finally {
      setIsCreating(false);
    }
  };

  // 格式化日期时间
  const formatEventTime = (event: CalendarEvent) => {
    if (event.all_day) {
      return '全天';
    }

    try {
      const startTime = parseISO(event.start_time);
      const endTime = parseISO(event.end_time);
      
      return `${format(startTime, 'HH:mm')} - ${format(endTime, 'HH:mm')}`;
    } catch {
      return '时间格式错误';
    }
  };

  // 格式化日期
  const formatEventDate = (dateStr: string) => {
    try {
      const date = parseISO(dateStr);
      const today = startOfDay(new Date());
      const tomorrow = addDays(today, 1);
      const eventDate = startOfDay(date);

      if (eventDate.getTime() === today.getTime()) {
        return '今天';
      } else if (eventDate.getTime() === tomorrow.getTime()) {
        return '明天';
      } else {
        return format(date, 'MM月dd日', { locale: zhCN });
      }
    } catch {
      return dateStr;
    }
  };

  // 按日期分组事件
  const groupEventsByDate = (events: CalendarEvent[]) => {
    const grouped: { [key: string]: CalendarEvent[] } = {};
    
    events.forEach(event => {
      const dateKey = format(parseISO(event.start_time), 'yyyy-MM-dd');
      if (!grouped[dateKey]) {
        grouped[dateKey] = [];
      }
      grouped[dateKey].push(event);
    });

    return grouped;
  };

  const groupedEvents = groupEventsByDate(events);

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span>近期日程</span>
            <Badge variant="secondary" className="text-xs">
              {events.length} 个事件
            </Badge>
          </div>
          
          <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
            <DialogTrigger asChild>
              <Button size="sm" onClick={handleOpenDialog}>
                <Plus className="h-4 w-4 mr-1" />
                添加
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
              <DialogHeader>
                <DialogTitle>创建新日程</DialogTitle>
                <DialogDescription>
                  添加一个新的日程事件到您的日历中
                </DialogDescription>
              </DialogHeader>
              
              <div className="grid gap-4 py-4">
                <div className="grid gap-2">
                  <Label htmlFor="title">标题 *</Label>
                  <Input
                    id="title"
                    placeholder="输入日程标题"
                    value={formData.title}
                    onChange={(e) => setFormData(prev => ({ ...prev, title: e.target.value }))}
                  />
                </div>
                
                <div className="grid gap-2">
                  <Label htmlFor="description">描述</Label>
                  <Textarea
                    id="description"
                    placeholder="输入日程描述（可选）"
                    value={formData.description}
                    onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                    rows={3}
                  />
                </div>
                
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="all_day"
                    checked={formData.all_day}
                    onCheckedChange={(checked) => 
                      setFormData(prev => ({ ...prev, all_day: checked as boolean }))
                    }
                  />
                  <Label htmlFor="all_day">全天事件</Label>
                </div>
                
                {!formData.all_day && (
                  <>
                    <div className="grid gap-2">
                      <Label htmlFor="start_time">开始时间 *</Label>
                      <Input
                        id="start_time"
                        type="datetime-local"
                        value={formData.start_time}
                        onChange={(e) => setFormData(prev => ({ ...prev, start_time: e.target.value }))}
                      />
                    </div>
                    
                    <div className="grid gap-2">
                      <Label htmlFor="end_time">结束时间 *</Label>
                      <Input
                        id="end_time"
                        type="datetime-local"
                        value={formData.end_time}
                        onChange={(e) => setFormData(prev => ({ ...prev, end_time: e.target.value }))}
                      />
                    </div>
                  </>
                )}
              </div>
              
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                  取消
                </Button>
                <Button onClick={handleCreateEvent} disabled={isCreating}>
                  {isCreating ? '创建中...' : '创建日程'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardTitle>
      </CardHeader>
      
      <CardContent>
        <div className="space-y-3 max-h-80 overflow-y-auto">
          {loading ? (
            // 加载状态
            Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-start gap-3 p-3 rounded-md border">
                <Skeleton className="h-4 w-4 mt-1" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
            ))
          ) : events.length === 0 ? (
            // 空状态
            <div className="text-center py-8 text-muted-foreground">
              <Calendar className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>暂无即将到来的日程</p>
              <p className="text-sm">点击"添加"创建新日程</p>
            </div>
          ) : (
            // 按日期分组显示事件
            Object.entries(groupedEvents).map(([dateKey, dayEvents]) => (
              <div key={dateKey} className="space-y-2">
                <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                  <CalendarDays className="h-4 w-4" />
                  {formatEventDate(dayEvents[0].start_time)}
                </div>
                
                {dayEvents.map((event) => (
                  <div
                    key={event.id}
                    className="flex items-start gap-3 p-3 rounded-md border hover:bg-muted/50 transition-colors"
                  >
                    <div className="mt-1">
                      {event.all_day ? (
                        <CalendarDays className="h-4 w-4 text-blue-500" />
                      ) : (
                        <Clock className="h-4 w-4 text-green-500" />
                      )}
                    </div>
                    
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-sm">{event.title}</p>
                      
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs text-muted-foreground">
                          {formatEventTime(event)}
                        </span>
                      </div>
                      
                      {event.description && (
                        <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                          {event.description}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}