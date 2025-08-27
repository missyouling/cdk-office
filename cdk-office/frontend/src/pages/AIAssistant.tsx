import React, { useState, useRef, useEffect } from 'react';
import { 
  Bot, 
  Send, 
  Paperclip, 
  History,
  ThumbsUp,
  ThumbsDown,
  Loader2
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Textarea } from '@/components/ui/textarea';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { ScrollArea } from '@/components/ui/scroll-area';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

// 模拟对话历史
const mockMessages: Message[] = [
  {
    id: '1',
    role: 'user',
    content: '公司年假政策是什么？',
    timestamp: '2023-10-15 10:30',
  },
  {
    id: '2',
    role: 'assistant',
    content: '根据公司政策，员工每年享有15天带薪年假。年假可以在当年内使用，未使用的年假可以累积到下一年，但最多不超过30天。申请年假需要提前一周通过系统提交申请，并得到直属上级的批准。',
    timestamp: '2023-10-15 10:31',
  },
  {
    id: '3',
    role: 'user',
    content: '如何申请年假？',
    timestamp: '2023-10-15 10:32',
  },
  {
    id: '4',
    role: 'assistant',
    content: '申请年假的步骤如下：\n1. 登录内部系统\n2. 进入"假期申请"模块\n3. 填写申请表单，包括请假时间、天数和原因\n4. 提交申请并等待直属上级审批\n5. 审批通过后，系统会自动更新您的假期余额',
    timestamp: '2023-10-15 10:33',
  },
];

const AIAssistant: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>(mockMessages);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 滚动到最新消息
  useEffect(() => {
    scrollToBottom();
  }, [messages, isLoading]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = () => {
    if (!inputValue.trim()) return;

    // 添加用户消息
    const newUserMessage: Message = {
      id: `msg_${Date.now()}`,
      role: 'user',
      content: inputValue,
      timestamp: new Date().toLocaleString('zh-CN'),
    };

    setMessages([...messages, newUserMessage]);
    setInputValue('');
    setIsLoading(true);

    // 模拟AI回复
    setTimeout(() => {
      const aiResponse: Message = {
        id: `msg_${Date.now() + 1}`,
        role: 'assistant',
        content: `这是针对"${inputValue}"的模拟回复。在实际应用中，这将通过Dify AI平台提供智能回答。您可以询问关于公司政策、流程、文档等方面的问题。`,
        timestamp: new Date().toLocaleString('zh-CN'),
      };

      setMessages(prev => [...prev, aiResponse]);
      setIsLoading(false);
    }, 1500);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">AI助手</h1>
          <p className="text-muted-foreground mt-2">
            基于Dify AI平台的智能办公助手
          </p>
        </div>
        <Badge variant="secondary">
          <History className="h-4 w-4 mr-1" />
          对话历史
        </Badge>
      </div>

      {/* AI助手介绍卡片 */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Avatar className="h-12 w-12">
              <AvatarFallback className="bg-primary text-primary-foreground">
                <Bot className="h-6 w-6" />
              </AvatarFallback>
            </Avatar>
            <div>
              <CardTitle>智能办公助手</CardTitle>
              <CardDescription>
                基于Dify AI平台，为您提供企业知识问答服务
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground mb-4">
            您可以询问关于公司政策、流程、文档等方面的问题，AI助手将为您提供准确的信息。
          </p>
          <div className="flex flex-wrap gap-2">
            <Badge variant="outline">年假政策</Badge>
            <Badge variant="outline">报销流程</Badge>
            <Badge variant="outline">入职手续</Badge>
            <Badge variant="outline">会议室预订</Badge>
          </div>
        </CardContent>
      </Card>

      {/* 消息对话区域 */}
      <ScrollArea className="h-[500px] border rounded-lg p-4">
        <div className="space-y-4">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-[80%] rounded-2xl p-4 ${
                  message.role === 'user'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted'
                }`}
              >
                <div className="whitespace-pre-line">{message.content}</div>
                <div className="flex justify-between items-center mt-2">
                  <span className="text-xs opacity-70">{message.timestamp}</span>
                  {message.role === 'assistant' && (
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" className="h-6 w-6">
                        <ThumbsUp className="h-3 w-3" />
                      </Button>
                      <Button variant="ghost" size="icon" className="h-6 w-6">
                        <ThumbsDown className="h-3 w-3" />
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
          
          {isLoading && (
            <div className="flex justify-start">
              <div className="max-w-[80%] bg-muted rounded-2xl p-4 flex items-center">
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                <span>AI助手正在思考中...</span>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </ScrollArea>

      {/* 输入区域 */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Textarea
            placeholder="请输入您的问题..."
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyPress}
            className="min-h-[60px] pr-12"
            disabled={isLoading}
          />
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 top-2"
            disabled={isLoading}
          >
            <Paperclip className="h-4 w-4" />
          </Button>
        </div>
        <Button
          size="lg"
          onClick={handleSendMessage}
          disabled={isLoading || !inputValue.trim()}
        >
          <Send className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
};

export default AIAssistant;