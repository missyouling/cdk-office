// Dashboard 待办事项和日程的 TypeScript 类型定义

export interface TodoItem {
  id: string;
  title: string;
  completed: boolean;
  due_date?: string;
  created_at: string;
  updated_at: string;
  is_overdue: boolean;
}

export interface CalendarEvent {
  id: string;
  title: string;
  description?: string;
  start_time: string;
  end_time: string;
  all_day: boolean;
  created_at: string;
}

export interface CreateTodoRequest {
  title: string;
  due_date?: string;
}

export interface UpdateTodoRequest {
  completed: boolean;
}

export interface CreateCalendarEventRequest {
  title: string;
  description?: string;
  start_time: string;
  end_time: string;
  all_day: boolean;
}

export interface UpdateCalendarEventRequest {
  title?: string;
  description?: string;
  start_time?: string;
  end_time?: string;
  all_day?: boolean;
}

export interface DashboardStats {
  total_todos: number;
  completed_todos: number;
  overdue_todos: number;
  pending_todos: number;
  today_events: number;
}

export interface Notification {
  id: string;
  title: string;
  content: string;
  type: string;
  category: string;
  priority: string;
  is_read: boolean;
  read_at?: string;
  related_id?: string;
  related_type?: string;
  action_required: boolean;
  action_taken: boolean;
  created_at: string;
}

export interface APIResponse<T> {
  data?: T;
  message?: string;
  error?: string;
}

export interface TodoListResponse {
  data: TodoItem[];
}

export interface CalendarEventListResponse {
  data: CalendarEvent[];
}