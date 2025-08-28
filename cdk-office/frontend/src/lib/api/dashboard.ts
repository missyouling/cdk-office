// Dashboard API 调用函数

import { 
  TodoItem, 
  CalendarEvent, 
  CreateTodoRequest, 
  UpdateTodoRequest, 
  CreateCalendarEventRequest, 
  UpdateCalendarEventRequest,
  DashboardStats,
  APIResponse,
  TodoListResponse,
  CalendarEventListResponse
} from '@/types/dashboard';

const API_BASE = '/api/v1';

// 工具函数：处理API响应
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || `HTTP ${response.status}`);
  }
  return response.json();
}

// 待办事项 API
export const todoAPI = {
  // 获取待办事项列表
  async getAll(completed?: boolean): Promise<TodoItem[]> {
    const params = new URLSearchParams();
    if (completed !== undefined) {
      params.append('completed', completed.toString());
    }
    
    const response = await fetch(`${API_BASE}/todos?${params}`);
    const result: TodoListResponse = await handleResponse(response);
    return result.data;
  },

  // 创建待办事项
  async create(data: CreateTodoRequest): Promise<TodoItem> {
    const response = await fetch(`${API_BASE}/todos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse<TodoItem>(response);
  },

  // 更新待办事项状态
  async update(id: string, data: UpdateTodoRequest): Promise<APIResponse<void>> {
    const response = await fetch(`${API_BASE}/todos/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse<APIResponse<void>>(response);
  },

  // 删除待办事项
  async delete(id: string): Promise<APIResponse<void>> {
    const response = await fetch(`${API_BASE}/todos/${id}`, {
      method: 'DELETE',
    });
    return handleResponse<APIResponse<void>>(response);
  },
};

// 日程事件 API
export const calendarAPI = {
  // 获取日程事件列表
  async getAll(startDate?: string, endDate?: string): Promise<CalendarEvent[]> {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    
    const response = await fetch(`${API_BASE}/calendar-events?${params}`);
    const result: CalendarEventListResponse = await handleResponse(response);
    return result.data;
  },

  // 获取未来7天的日程
  async getUpcoming(): Promise<CalendarEvent[]> {
    const response = await fetch(`${API_BASE}/calendar-events/upcoming`);
    const result: CalendarEventListResponse = await handleResponse(response);
    return result.data;
  },

  // 创建日程事件
  async create(data: CreateCalendarEventRequest): Promise<CalendarEvent> {
    const response = await fetch(`${API_BASE}/calendar-events`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse<CalendarEvent>(response);
  },

  // 更新日程事件
  async update(id: string, data: UpdateCalendarEventRequest): Promise<APIResponse<void>> {
    const response = await fetch(`${API_BASE}/calendar-events/${id}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    return handleResponse<APIResponse<void>>(response);
  },

  // 删除日程事件
  async delete(id: string): Promise<APIResponse<void>> {
    const response = await fetch(`${API_BASE}/calendar-events/${id}`, {
      method: 'DELETE',
    });
    return handleResponse<APIResponse<void>>(response);
  },
};

// Dashboard 统计 API
export const dashboardAPI = {
  // 获取Dashboard统计信息
  async getStats(): Promise<DashboardStats> {
    const response = await fetch(`${API_BASE}/dashboard/stats`);
    return handleResponse<DashboardStats>(response);
  },
};

// 通知 API
export const notificationAPI = {
  // 获取通知列表
  async getAll(limit = 10, unreadOnly = false): Promise<Notification[]> {
    const params = new URLSearchParams();
    params.append('limit', limit.toString());
    if (unreadOnly) {
      params.append('unread_only', 'true');
    }
    
    const response = await fetch(`${API_BASE}/notifications?${params}`);
    const result = await handleResponse<{data: Notification[]}>(response);
    return result.data;
  },

  // 标记通知为已读
  async markAsRead(id: string): Promise<APIResponse<void>> {
    const response = await fetch(`${API_BASE}/notifications/${id}/read`, {
      method: 'PATCH',
    });
    return handleResponse<APIResponse<void>>(response);
  },
};