import { useState, useRef, useEffect } from 'react';
import type { Message, ChatRequest } from '@/types/ai-chat';
import { chatAPI, getErrorMessage } from '@/lib/api/ai-chat';

interface UseAIChatReturn {
  messages: Message[];
  inputValue: string;
  isLoading: boolean;
  error: string | null;
  setInputValue: (value: string) => void;
  sendMessage: () => Promise<void>;
  clearChat: () => void;
  clearError: () => void;
}

export function useAIChat(): UseAIChatReturn {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 发送消息
  const sendMessage = async (): Promise<void> => {
    if (!inputValue.trim() || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputValue.trim(),
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsLoading(true);
    setError(null);

    try {
      const request: ChatRequest = {
        question: userMessage.content,
        context: {
          source: 'web',
          session_id: Date.now().toString()
        }
      };

      const data = await chatAPI.chat(request);

      const aiMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'ai',
        content: data.answer,
        timestamp: new Date(),
        sources: data.sources,
        confidence: data.confidence,
        messageId: data.message_id
      };

      setMessages(prev => [...prev, aiMessage]);
    } catch (err) {
      console.error('Chat API error:', err);
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
      
      // 添加错误消息到聊天界面
      const errorAiMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'ai',
        content: '抱歉，我暂时无法回答您的问题。请稍后再试。',
        timestamp: new Date(),
        confidence: 0
      };
      setMessages(prev => [...prev, errorAiMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  // 清空对话
  const clearChat = (): void => {
    setMessages([]);
    setError(null);
  };

  // 清除错误
  const clearError = (): void => {
    setError(null);
  };

  return {
    messages,
    inputValue,
    isLoading,
    error,
    setInputValue,
    sendMessage,
    clearChat,
    clearError
  };
}

// 自动滚动hook
export function useAutoScroll(dependency: any[], enabled: boolean = true) {
  const scrollAreaRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    if (!enabled || !scrollAreaRef.current) return;
    
    const scrollContainer = scrollAreaRef.current.querySelector('[data-radix-scroll-area-viewport]');
    if (scrollContainer) {
      scrollContainer.scrollTop = scrollContainer.scrollHeight;
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, dependency);

  return { scrollAreaRef, scrollToBottom };
}