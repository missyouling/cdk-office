// AI聊天功能的 TypeScript 类型定义

export interface Message {
  id: string;
  type: 'user' | 'ai';
  content: string;
  timestamp: Date;
  sources?: DocumentSource[];
  confidence?: number;
  messageId?: string;
}

export interface DocumentSource {
  id: string;
  name: string;
  snippet: string;
  score: number;
}

export interface ChatRequest {
  question: string;
  context?: {
    source: string;
    session_id: string;
    [key: string]: any;
  };
}

export interface ChatResponse {
  answer: string;
  sources: DocumentSource[];
  confidence: number;
  message_id: string;
  created_at: string;
}

export interface ChatHistoryResponse {
  data: ChatHistoryItem[];
  pagination: {
    page: number;
    size: number;
    total: number;
  };
}

export interface ChatHistoryItem {
  id: string;
  question: string;
  answer: string;
  confidence: number;
  created_at: string;
  sources?: DocumentSource[];
}

export interface FeedbackRequest {
  feedback: string;
}

export interface ChatStats {
  total_chats: number;
  today_chats: number;
  avg_confidence: number;
  top_users: UserActivity[];
  recent_activity: ActivityRecord[];
}

export interface UserActivity {
  user_id: string;
  user_name: string;
  count: number;
}

export interface ActivityRecord {
  date: string;
  count: number;
}

export interface ErrorResponse {
  code: string;
  message: string;
  error?: string;
}

export interface SuccessResponse {
  code: string;
  message: string;
}

// 聊天界面状态
export interface ChatState {
  messages: Message[];
  isLoading: boolean;
  error: string | null;
  inputValue: string;
}

// API客户端类型
export interface APIResponse<T = any> {
  success?: boolean;
  data?: T;
  error?: string;
  message?: string;
  code?: string;
}