'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { toast } from '@/components/ui/use-toast';
import { Skeleton } from '@/components/ui/skeleton';
import { Plus, Trash2, AlertTriangle } from 'lucide-react';
import { format, isAfter, parseISO } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { TodoItem } from '@/types/dashboard';
import { todoAPI } from '@/lib/api/dashboard';

interface TodoCardProps {
  className?: string;
}

export function TodoCard({ className }: TodoCardProps) {
  const [todos, setTodos] = useState<TodoItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [newTodoTitle, setNewTodoTitle] = useState('');
  const [isAddingTodo, setIsAddingTodo] = useState(false);

  // 加载待办事项
  const loadTodos = async () => {
    try {
      const data = await todoAPI.getAll();
      setTodos(data);
    } catch (error) {
      console.error('获取待办事项失败:', error);
      toast({
        title: '错误',
        description: '获取待办事项失败，请重试',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTodos();
  }, []);

  // 添加新待办事项
  const handleAddTodo = async () => {
    if (!newTodoTitle.trim()) return;

    setIsAddingTodo(true);
    try {
      const newTodo = await todoAPI.create({
        title: newTodoTitle.trim(),
      });
      
      setTodos(prev => [newTodo, ...prev]);
      setNewTodoTitle('');
      
      toast({
        title: '成功',
        description: '待办事项已添加',
      });
    } catch (error) {
      console.error('添加待办事项失败:', error);
      toast({
        title: '错误',
        description: '添加待办事项失败，请重试',
        variant: 'destructive',
      });
    } finally {
      setIsAddingTodo(false);
    }
  };

  // 切换完成状态（乐观更新）
  const handleToggleComplete = async (id: string, completed: boolean) => {
    // 乐观更新 UI
    setTodos(prev => 
      prev.map(todo => 
        todo.id === id ? { ...todo, completed } : todo
      )
    );

    try {
      await todoAPI.update(id, { completed });
      
      toast({
        title: '成功',
        description: completed ? '任务已完成' : '任务已标记为待完成',
      });
    } catch (error) {
      // 发生错误时回滚 UI 状态
      setTodos(prev => 
        prev.map(todo => 
          todo.id === id ? { ...todo, completed: !completed } : todo
        )
      );
      
      console.error('更新待办事项失败:', error);
      toast({
        title: '错误',
        description: '更新待办事项失败，请重试',
        variant: 'destructive',
      });
    }
  };

  // 删除待办事项
  const handleDeleteTodo = async (id: string) => {
    // 乐观更新 UI
    const originalTodos = todos;
    setTodos(prev => prev.filter(todo => todo.id !== id));

    try {
      await todoAPI.delete(id);
      
      toast({
        title: '成功',
        description: '待办事项已删除',
      });
    } catch (error) {
      // 发生错误时回滚 UI 状态
      setTodos(originalTodos);
      
      console.error('删除待办事项失败:', error);
      toast({
        title: '错误',
        description: '删除待办事项失败，请重试',
        variant: 'destructive',
      });
    }
  };

  // 处理回车键添加
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !isAddingTodo) {
      handleAddTodo();
    }
  };

  // 格式化日期
  const formatDueDate = (dueDateStr: string) => {
    try {
      const dueDate = parseISO(dueDateStr);
      return format(dueDate, 'MM月dd日', { locale: zhCN });
    } catch {
      return dueDateStr;
    }
  };

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <span>待办事项</span>
          <Badge variant="secondary" className="text-xs">
            {todos.filter(t => !t.completed).length} 待完成
          </Badge>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 添加新待办事项 */}
        <div className="flex gap-2">
          <Input
            placeholder="添加新的待办事项..."
            value={newTodoTitle}
            onChange={(e) => setNewTodoTitle(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={isAddingTodo}
          />
          <Button 
            onClick={handleAddTodo}
            disabled={!newTodoTitle.trim() || isAddingTodo}
            size="icon"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </div>

        {/* 待办事项列表 */}
        <div className="space-y-2 max-h-80 overflow-y-auto">
          {loading ? (
            // 加载状态
            Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 p-2">
                <Skeleton className="h-4 w-4" />
                <Skeleton className="h-4 flex-1" />
                <Skeleton className="h-4 w-4" />
              </div>
            ))
          ) : todos.length === 0 ? (
            // 空状态
            <div className="text-center py-8 text-muted-foreground">
              <p>暂无待办事项</p>
              <p className="text-sm">添加第一个任务开始管理</p>
            </div>
          ) : (
            // 待办事项列表
            todos.map((todo) => (
              <div
                key={todo.id}
                className={`flex items-center gap-3 p-2 rounded-md hover:bg-muted/50 group ${
                  todo.completed ? 'opacity-60' : ''
                }`}
              >
                <Checkbox
                  checked={todo.completed}
                  onCheckedChange={(checked) =>
                    handleToggleComplete(todo.id, checked as boolean)
                  }
                />
                
                <div className="flex-1 min-w-0">
                  <p
                    className={`text-sm ${
                      todo.completed
                        ? 'line-through text-muted-foreground'
                        : 'text-foreground'
                    }`}
                  >
                    {todo.title}
                  </p>
                  
                  {todo.due_date && (
                    <div className="flex items-center gap-1 mt-1">
                      {todo.is_overdue && !todo.completed && (
                        <AlertTriangle className="h-3 w-3 text-red-500" />
                      )}
                      <span
                        className={`text-xs ${
                          todo.is_overdue && !todo.completed
                            ? 'text-red-500 font-medium'
                            : 'text-muted-foreground'
                        }`}
                      >
                        {formatDueDate(todo.due_date)}
                      </span>
                    </div>
                  )}
                </div>

                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={() => handleDeleteTodo(todo.id)}
                >
                  <Trash2 className="h-3 w-3" />
                </Button>
              </div>
            ))
          )}
        </div>

        {/* 统计信息 */}
        {!loading && todos.length > 0 && (
          <div className="pt-2 border-t text-xs text-muted-foreground flex justify-between">
            <span>
              已完成: {todos.filter(t => t.completed).length}
            </span>
            <span>
              逾期: {todos.filter(t => t.is_overdue && !t.completed).length}
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );
}