// AI聊天API客户端

import type {
  ChatRequest,
  ChatResponse,
  ChatHistoryResponse,
  FeedbackRequest,
  SuccessResponse,
  ErrorResponse,
  APIResponse
} from '@/types/ai-chat';

// API基础配置
const API_BASE_URL = '/api';

// 通用错误处理
class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message);
    this.name = 'APIError';
  }
}

// 通用请求处理器
async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  
  const defaultHeaders = {
    'Content-Type': 'application/json',
  };

  const config: RequestInit = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
  };

  try {
    const response = await fetch(url, config);
    
    if (!response.ok) {
      let errorMessage = '请求失败';
      let errorCode = 'UNKNOWN_ERROR';
      
      try {
        const errorData: ErrorResponse = await response.json();
        errorMessage = errorData.message || errorData.error || errorMessage;
        errorCode = errorData.code || errorCode;
      } catch {
        errorMessage = `HTTP ${response.status}: ${response.statusText}`;
      }
      
      throw new APIError(errorMessage, response.status, errorCode);
    }

    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      return response.json();
    }
    
    return response.text() as T;
  } catch (error) {
    if (error instanceof APIError) {
      throw error;
    }
    
    // 网络错误或其他错误
    if (error instanceof TypeError && error.message.includes('fetch')) {
      throw new APIError('网络连接失败，请检查网络连接', 0, 'NETWORK_ERROR');
    }
    
    throw new APIError(
      error instanceof Error ? error.message : '未知错误',
      0,
      'UNKNOWN_ERROR'
    );
  }
}

// AI聊天API客户端
export class AIChatAPI {
  // 发送聊天消息
  static async chat(request: ChatRequest): Promise<ChatResponse> {
    return apiRequest<ChatResponse>('/ai/chat', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // 获取聊天历史
  static async getChatHistory(
    page: number = 1,
    size: number = 20
  ): Promise<ChatHistoryResponse> {
    return apiRequest<ChatHistoryResponse>(
      `/ai/chat/history?page=${page}&size=${size}`,
      {
        method: 'GET',
      }
    );
  }

  // 提交反馈
  static async submitFeedback(
    messageId: string,
    feedback: FeedbackRequest
  ): Promise<SuccessResponse> {
    return apiRequest<SuccessResponse>(`/ai/chat/${messageId}/feedback`, {
      method: 'PATCH',
      body: JSON.stringify(feedback),
    });
  }

  // 获取聊天统计
  static async getChatStats(): Promise<any> {
    return apiRequest<any>('/ai/chat/stats', {
      method: 'GET',
    });
  }
}

// 导出便捷方法
export const chatAPI = AIChatAPI;

// 错误处理工具
export const isAPIError = (error: unknown): error is APIError => {
  return error instanceof APIError;
};

export const getErrorMessage = (error: unknown): string => {
  if (isAPIError(error)) {
    return error.message;
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return '发生了未知错误';
};

export const getErrorCode = (error: unknown): string | undefined => {
  if (isAPIError(error)) {
    return error.code;
  }
  
  return undefined;
};