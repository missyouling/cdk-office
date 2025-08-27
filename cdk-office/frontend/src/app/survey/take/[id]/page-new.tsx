'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle,
  CardDescription
} from '@/components/ui/card';
import { 
  Button 
} from '@/components/ui/button';
import { 
  Input 
} from '@/components/ui/input';
import { 
  Label 
} from '@/components/ui/label';
import { 
  Textarea 
} from '@/components/ui/textarea';
import { 
  Progress 
} from '@/components/ui/progress';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { 
  Badge 
} from '@/components/ui/badge';
import { 
  toast 
} from '@/components/ui/use-toast';

import { 
  CheckCircle, 
  Timer, 
  Share, 
  ArrowLeft,
  FileText
} from 'lucide-react';

// 模拟 SurveyJS Survey 组件
const SurveyComponent = ({ 
  json, 
  onComplete, 
  onProgress 
}: { 
  json: any,
  onComplete: (result: any) => void,
  onProgress: (progress: number) => void
}) => {
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [answers, setAnswers] = useState<{[key: string]: any}>({});
  const [startTime] = useState(Date.now());
  
  const questions = [
    {
      name: "name",
      title: "您的姓名",
      type: "text",
      required: true
    },
    {
      name: "satisfaction",
      title: "总体满意度如何？",
      type: "radiogroup",
      choices: ["非常满意", "满意", "一般", "不满意", "非常不满意"]
    },
    {
      name: "feedback",
      title: "您还有什么建议吗？",
      type: "comment"
    }
  ];

  const progress = ((currentQuestion + 1) / questions.length) * 100;

  useEffect(() => {
    onProgress(progress);
  }, [progress, onProgress]);

  const handleNext = () => {
    if (currentQuestion < questions.length - 1) {
      setCurrentQuestion(currentQuestion + 1);
    } else {
      // 完成问卷
      const timeSpent = Math.floor((Date.now() - startTime) / 1000);
      onComplete({
        data: answers,
        timeSpent: timeSpent
      });
    }
  };

  const handlePrev = () => {
    if (currentQuestion > 0) {
      setCurrentQuestion(currentQuestion - 1);
    }
  };

  const handleAnswerChange = (value: any) => {
    const question = questions[currentQuestion];
    setAnswers({ ...answers, [question.name]: value });
  };

  const currentQ = questions[currentQuestion];
  const currentAnswer = answers[currentQ.name];
  const isRequired = currentQ.required;
  const canNext = !isRequired || (currentAnswer && currentAnswer !== '');

  const renderQuestion = () => {
    switch (currentQ.type) {
      case 'text':
        return (
          <div className="mt-4">
            <Input
              value={currentAnswer || ''}
              onChange={(e) => handleAnswerChange(e.target.value)}
              placeholder="请输入您的回答"
            />
          </div>
        );
      
      case 'radiogroup':
        return (
          <div className="mt-4 space-y-2">
            {currentQ.choices?.map((choice: string, index: number) => (
              <div
                key={index}
                className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                  currentAnswer === choice 
                    ? 'border-primary bg-primary/10' 
                    : 'border-border hover:bg-muted'
                }`}
                onClick={() => handleAnswerChange(choice)}
              >
                {choice}
              </div>
            ))}
          </div>
        );
      
      case 'comment':
        return (
          <div className="mt-4">
            <Textarea
              value={currentAnswer || ''}
              onChange={(e) => handleAnswerChange(e.target.value)}
              placeholder="请输入您的建议..."
              rows={4}
            />
          </div>
        );
      
      default:
        return null;
    }
  };

  return (
    <div>
      {/* 进度条 */}
      <div className="mb-6">
        <div className="flex justify-between mb-2">
          <span className="text-sm text-muted-foreground">
            问题 {currentQuestion + 1} / {questions.length}
          </span>
          <span className="text-sm text-muted-foreground">
            {Math.round(progress)}% 完成
          </span>
        </div>
        <Progress value={progress} className="h-2" />
      </div>

      {/* 当前问题 */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <h3 className="text-xl font-semibold">
            {currentQ.title}
            {currentQ.required && <span className="text-destructive"> *</span>}
          </h3>
          
          {renderQuestion()}
        </CardContent>
      </Card>

      {/* 导航按钮 */}
      <div className="flex justify-between">
        <Button
          onClick={handlePrev}
          disabled={currentQuestion === 0}
          variant="outline"
        >
          上一题
        </Button>
        
        <Button
          onClick={handleNext}
          disabled={!canNext}
        >
          {currentQuestion === questions.length - 1 ? '提交问卷' : '下一题'}
        </Button>
      </div>
    </div>
  );
};

export default function TakeSurveyPage() {
  const router = useRouter();
  const params = useParams();
  const surveyId = params.id as string;
  
  const [survey, setSurvey] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [progress, setProgress] = useState(0);
  const [completed, setCompleted] = useState(false);
  const [timeSpent, setTimeSpent] = useState(0);
  const [submitDialogOpen, setSubmitDialogOpen] = useState(false);

  // 模拟加载问卷数据
  useEffect(() => {
    setTimeout(() => {
      setSurvey({
        id: surveyId,
        title: '员工满意度调查',
        description: '了解员工对公司的满意度和建议',
        createdAt: '2025-01-01',
        expiresAt: '2025-02-01'
      });
      setLoading(false);
    }, 1000);
  }, [surveyId]);

  const handleProgress = (progress: number) => {
    setProgress(progress);
  };

  const handleComplete = (result: any) => {
    setTimeSpent(result.timeSpent);
    setCompleted(true);
    setSubmitDialogOpen(true);
    
    // 模拟提交数据
    console.log('Survey result:', result);
    
    toast({
      title: "提交成功",
      description: "感谢您参与本次问卷调查",
    });
  };

  const handleConfirmSubmit = () => {
    setSubmitDialogOpen(false);
    // 实际项目中这里会提交数据到服务器
    router.push('/survey');
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <span className="ml-2">正在加载问卷...</span>
      </div>
    );
  }

  if (!survey) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px]">
        <FileText className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-semibold">问卷未找到</h3>
        <p className="text-muted-foreground">无法找到指定的问卷</p>
        <Button className="mt-4" onClick={() => router.back()}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          返回
        </Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和返回按钮 */}
      <div className="flex items-center space-x-4">
        <Button variant="outline" size="icon" onClick={() => router.back()}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">{survey.title}</h1>
          <p className="text-muted-foreground">{survey.description}</p>
        </div>
        <div className="ml-auto flex items-center space-x-2">
          <Badge variant="secondary">
            <Timer className="h-3 w-3 mr-1" />
            剩余时间: 14天
          </Badge>
        </div>
      </div>

      {/* 问卷内容 */}
      {completed ? (
        <Card>
          <CardContent className="p-8 text-center">
            <CheckCircle className="h-16 w-16 text-green-500 mx-auto mb-4" />
            <h2 className="text-2xl font-bold mb-2">问卷提交成功</h2>
            <p className="text-muted-foreground mb-6">
              感谢您参与本次问卷调查，您的反馈对我们非常重要。
            </p>
            <div className="flex justify-center space-x-4">
              <Button onClick={() => router.push('/survey')}>
                返回问卷列表
              </Button>
              <Button variant="outline" onClick={() => router.back()}>
                关闭
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <SurveyComponent 
          json={{}}
          onProgress={handleProgress}
          onComplete={handleComplete}
        />
      )}

      {/* 提交确认对话框 */}
      <Dialog open={submitDialogOpen} onOpenChange={setSubmitDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认提交问卷</DialogTitle>
            <DialogDescription>
              确认提交问卷 "{survey.title}" 吗？
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="flex items-center space-x-2 rounded-lg border p-4">
              <FileText className="h-5 w-5 text-muted-foreground" />
              <div>
                <div className="font-medium">{survey.title}</div>
                <div className="text-sm text-muted-foreground">
                  用时 {Math.floor(timeSpent / 60)}分{timeSpent % 60}秒
                </div>
              </div>
            </div>
            <div className="text-sm text-muted-foreground">
              提交后将无法修改答案，请确认所有问题都已回答。
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSubmitDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleConfirmSubmit}>
              确认提交
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}