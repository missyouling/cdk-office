import React, { useState } from 'react';
import {
  Box,
  Typography,
  TextField,
  Button,
  Card,
  CardContent,
  CardHeader,
  Avatar,
  IconButton,
  Chip,
  CircularProgress,
} from '@mui/material';
import { 
  SmartToy, 
  Send, 
  AttachFile, 
  History,
  ThumbUp,
  ThumbDown
} from '@mui/icons-material';

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
    <Box sx={{ width: '100%', p: 2 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          AI助手
        </Typography>
        <Chip icon={<History />} label="对话历史" />
      </Box>

      <Card sx={{ mb: 3 }}>
        <CardHeader
          avatar={
            <Avatar sx={{ bgcolor: 'primary.main' }}>
              <SmartToy />
            </Avatar>
          }
          title="智能办公助手"
          subheader="基于Dify AI平台，为您提供企业知识问答服务"
        />
        <CardContent>
          <Typography variant="body1" color="text.secondary">
            您可以询问关于公司政策、流程、文档等方面的问题，AI助手将为您提供准确的信息。
          </Typography>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 2 }}>
            <Chip label="年假政策" variant="outlined" />
            <Chip label="报销流程" variant="outlined" />
            <Chip label="入职手续" variant="outlined" />
            <Chip label="会议室预订" variant="outlined" />
          </Box>
        </CardContent>
      </Card>

      <Box sx={{ 
        height: '500px', 
        overflowY: 'auto', 
        border: '1px solid #e0e0e0', 
        borderRadius: 1, 
        mb: 2,
        p: 2
      }}>
        {messages.map((message) => (
          <Box 
            key={message.id} 
            sx={{ 
              display: 'flex', 
              justifyContent: message.role === 'user' ? 'flex-end' : 'flex-start',
              mb: 2
            }}
          >
            <Box sx={{ 
              maxWidth: '80%', 
              bgcolor: message.role === 'user' ? 'primary.light' : 'grey.100',
              borderRadius: 2,
              p: 2
            }}>
              <Typography variant="body1" sx={{ whiteSpace: 'pre-line' }}>
                {message.content}
              </Typography>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 1 }}>
                <Typography variant="caption" color="text.secondary">
                  {message.timestamp}
                </Typography>
                {message.role === 'assistant' && (
                  <Box>
                    <IconButton size="small">
                      <ThumbUp fontSize="small" />
                    </IconButton>
                    <IconButton size="small">
                      <ThumbDown fontSize="small" />
                    </IconButton>
                  </Box>
                )}
              </Box>
            </Box>
          </Box>
        ))}
        
        {isLoading && (
          <Box sx={{ display: 'flex', justifyContent: 'flex-start', mb: 2 }}>
            <Box sx={{ 
              maxWidth: '80%', 
              bgcolor: 'grey.100',
              borderRadius: 2,
              p: 2,
              display: 'flex',
              alignItems: 'center'
            }}>
              <CircularProgress size={20} sx={{ mr: 1 }} />
              <Typography variant="body1">
                AI助手正在思考中...
              </Typography>
            </Box>
          </Box>
        )}
      </Box>

      <Box sx={{ display: 'flex', gap: 2 }}>
        <TextField
          fullWidth
          multiline
          maxRows={4}
          variant="outlined"
          placeholder="请输入您的问题..."
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyPress={handleKeyPress}
          InputProps={{
            startAdornment: (
              <IconButton sx={{ mr: 1 }}>
                <AttachFile />
              </IconButton>
            ),
          }}
        />
        <Button
          variant="contained"
          size="large"
          endIcon={<Send />}
          onClick={handleSendMessage}
          disabled={isLoading || !inputValue.trim()}
        >
          发送
        </Button>
      </Box>
    </Box>
  );
};

export default AIAssistant;