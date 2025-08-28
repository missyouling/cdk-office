'use client';

import React, { useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Avatar } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Send, 
  Bot, 
  User, 
  Loader2, 
  MessageSquare,
  AlertCircle,
  FileText,
  Star
} from 'lucide-react';
import type {
  Message,
  DocumentSource
} from '@/types/ai-chat';
import { useAIChat, useAutoScroll } from '@/hooks/useAIChat';

export default function AIChatInterface() {
  const {
    messages,
    inputValue,
    isLoading,
    error,
    setInputValue,
    sendMessage,
    clearChat
  } = useAIChat();
  
  const { scrollAreaRef } = useAutoScroll([messages, isLoading]);
  const inputRef = useRef<HTMLInputElement>(null);

  // 处理回车键发送
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  // 渲染消息气泡
  const renderMessage = (message: Message) => {
    const isUser = message.type === 'user';
    
    return (
      <div
        key={message.id}
        className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}
      >
        <div className={`flex ${isUser ? 'flex-row-reverse' : 'flex-row'} items-start space-x-2 max-w-[80%]`}>
          {/* 头像 */}
          <Avatar className="h-8 w-8 flex-shrink-0">
            <div className={`h-full w-full flex items-center justify-center ${isUser ? 'bg-blue-500' : 'bg-green-500'}`}>
              {isUser ? <User className="h-4 w-4 text-white" /> : <Bot className="h-4 w-4 text-white" />}
            </div>
          </Avatar>

          {/* 消息内容 */}
          <div className={`${isUser ? 'mr-2' : 'ml-2'}`}>
            <div
              className={`rounded-lg px-4 py-2 ${
                isUser 
                  ? 'bg-blue-500 text-white' 
                  : 'bg-gray-100 text-gray-900 border'
              }`}
            >
              <p className="text-sm whitespace-pre-wrap">{message.content}</p>
            </div>

            {/* AI消息的额外信息 */}
            {!isUser && (
              <div className="mt-2 space-y-2">
                {/* 置信度 */}
                {message.confidence !== undefined && (
                  <div className="flex items-center space-x-2">
                    <Badge variant={message.confidence > 0.8 ? 'default' : message.confidence > 0.5 ? 'secondary' : 'outline'}>
                      <Star className="h-3 w-3 mr-1" />
                      置信度: {(message.confidence * 100).toFixed(0)}%
                    </Badge>
                  </div>
                )}

                {/* 来源文档 */}
                {message.sources && message.sources.length > 0 && (
                  <div className="space-y-1">
                    <p className="text-xs text-gray-500">参考来源:</p>
                    {message.sources.map((source, index) => (
                      <div key={index} className="bg-gray-50 rounded p-2 text-xs">
                        <div className="flex items-center space-x-2">
                          <FileText className="h-3 w-3 text-gray-400" />
                          <span className="font-medium">{source.name}</span>
                          <Badge variant="outline" className="text-xs">
                            {(source.score * 100).toFixed(0)}%
                          </Badge>
                        </div>
                        {source.snippet && (
                          <p className="mt-1 text-gray-600 line-clamp-2">{source.snippet}</p>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* 时间戳 */}
            <p className={`text-xs mt-1 ${isUser ? 'text-right text-blue-300' : 'text-gray-500'}`}>
              {message.timestamp.toLocaleTimeString('zh-CN', { 
                hour: '2-digit', 
                minute: '2-digit' 
              })}
            </p>
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* 头部 */}
      <Card className="rounded-none border-l-0 border-r-0 border-t-0">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="h-10 w-10 rounded-full bg-gradient-to-r from-blue-500 to-green-500 flex items-center justify-center">
                <MessageSquare className="h-5 w-5 text-white" />
              </div>
              <div>
                <CardTitle className="text-lg">AI 智能问答</CardTitle>
                <p className="text-sm text-gray-500">基于 Dify AI 平台的智能助手</p>
              </div>
            </div>
            <Button variant="outline" size="sm" onClick={clearChat}>
              清空对话
            </Button>
          </div>
        </CardHeader>
      </Card>

      {/* 错误提示 */}
      {error && (
        <Alert className="rounded-none border-l-0 border-r-0 border-t-0">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 聊天区域 */}
      <div className="flex-1 flex flex-col">
        <ScrollArea className="flex-1 p-4" ref={scrollAreaRef}>
          {messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-center space-y-4">
              <div className="h-16 w-16 rounded-full bg-gradient-to-r from-blue-500 to-green-500 flex items-center justify-center">
                <Bot className="h-8 w-8 text-white" />
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900">欢迎使用 AI 智能问答</h3>
                <p className="text-gray-500 mt-1">请输入您的问题，我会尽力为您提供帮助</p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2 max-w-md w-full">
                {[
                  '什么是 CDK-Office？',
                  '如何使用员工管理功能？',
                  '如何上传和管理文档？',
                  '系统有哪些核心功能？'
                ].map((suggestion, index) => (
                  <Button
                    key={index}
                    variant="outline"
                    size="sm"
                    className="text-left justify-start h-auto py-2 px-3"
                    onClick={() => setInputValue(suggestion)}
                  >
                    {suggestion}
                  </Button>
                ))}
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map(renderMessage)}
              {/* 加载指示器 */}
              {isLoading && (
                <div className="flex justify-start mb-4">
                  <div className="flex items-start space-x-2">
                    <Avatar className="h-8 w-8 flex-shrink-0">
                      <div className="h-full w-full flex items-center justify-center bg-green-500">
                        <Bot className="h-4 w-4 text-white" />
                      </div>
                    </Avatar>
                    <div className="ml-2">
                      <div className="bg-gray-100 rounded-lg px-4 py-2 border">
                        <div className="flex items-center space-x-2">
                          <Loader2 className="h-4 w-4 animate-spin" />
                          <span className="text-sm text-gray-600">AI 正在思考中...</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </ScrollArea>

        <Separator />

        {/* 输入区域 */}
        <CardContent className="p-4 bg-white">
          <div className="flex space-x-2">
            <Input
              ref={inputRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="输入您的问题..."
              className="flex-1"
              disabled={isLoading}
            />
            <Button 
              onClick={sendMessage} 
              disabled={!inputValue.trim() || isLoading}
              size="icon"
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </Button>
          </div>
          <p className="text-xs text-gray-500 mt-2">
            按 Enter 发送消息，Shift + Enter 换行
          </p>
        </CardContent>
      </div>
    </div>
  );
}