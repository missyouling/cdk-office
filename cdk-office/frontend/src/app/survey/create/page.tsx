'use client';

import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { 
  Switch 
} from '@/components/ui/switch';
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
  Save, 
  Eye, 
  ArrowLeft, 
  Send, 
  FileText,
  Calendar,
  Hash,
  Users,
  Globe
} from 'lucide-react';

// SurveyJS Creator 组件（模拟实现）
const SurveyCreatorComponent = ({ 
  json, 
  onSurveyChange, 
  onSave 
}: { 
  json: any, 
  onSurveyChange: (survey: any) => void,
  onSave: (saveNo: number, callback: (saveNo: number, isSuccess: boolean) => void) => void
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // 在实际项目中，这里会初始化 SurveyJS Creator
    // 模拟 SurveyJS Creator 的界面
    console.log('Initializing SurveyJS Creator with:', json);
  }, [json]);

  return (
    <div 
      ref={containerRef} 
      className="w-full h-[600px] border rounded-lg flex items-center justify-center bg-muted"
    >
      <div className="text-center">
        <h3 className="text-lg font-medium text-muted-foreground mb-2">
          SurveyJS 问卷创建器
        </h3>
        <p className="text-sm text-muted-foreground mb-4">
          在实际项目中，这里将展示完整的 SurveyJS Creator 界面
        </p>
        <Button 
          variant="outline"
          onClick={() => {
            // 模拟问卷设计变化
            const mockSurvey = {
              title: "示例问卷",
              pages: [{
                name: "page1",
                elements: [{
                  type: "text",
                  name: "question1",
                  title: "您的姓名是？"
                }]
              }]
            };
            onSurveyChange(mockSurvey);
          }}
        >
          模拟添加问题
        </Button>
      </div>
    </div>
  );
};

export default function CreateSurveyPage() {
  const router = useRouter();
  const [activeStep, setActiveStep] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  
  // 问卷基本信息
  const [surveyInfo, setSurveyInfo] = useState({
    title: '',
    description: '',
    tags: [] as string[],
    isPublic: false,
    maxResponses: 0,
    startTime: '',
    endTime: ''
  });
  
  // SurveyJS JSON 定义
  const [surveyJson, setSurveyJson] = useState({
    title: "新建问卷",
    pages: [{
      name: "page1",
      elements: []
    }]
  });
  
  // 对话框状态
  const [previewOpen, setPreviewOpen] = useState(false);
  const [templateOpen, setTemplateOpen] = useState(false);
  const [publishOpen, setPublishOpen] = useState(false);
  
  // 模板数据
  const [templates] = useState([
    { id: '1', name: '员工满意度调查', category: '人事管理' },
    { id: '2', name: '产品反馈问卷', category: '市场调研' },
    { id: '3', name: '培训效果评估', category: '教育培训' },
    { id: '4', name: '客户服务评价', category: '服务质量' }
  ]);

  const steps = ['基本信息', '问卷设计', '发布设置'];

  const handleNext = () => {
    if (activeStep === 0 && !surveyInfo.title) {
      toast({
        title: "缺少标题",
        description: "请填写问卷标题",
        variant: "destructive",
      });
      return;
    }
    setActiveStep((prevActiveStep) => prevActiveStep + 1);
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
  };

  const handleSurveyChange = (survey: any) => {
    setSurveyJson(survey);
  };

  const handleSave = async (saveNo: number, callback: (saveNo: number, isSuccess: boolean) => void) => {
    setIsLoading(true);
    try {
      // 模拟保存操作
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      console.log('Saving survey:', {
        ...surveyInfo,
        jsonDefinition: surveyJson
      });
      
      toast({
        title: "保存成功",
        description: "问卷已保存为草稿",
      });
      
      callback(saveNo, true);
    } catch (error) {
      toast({
        title: "保存失败",
        description: "请重试",
        variant: "destructive",
      });
      callback(saveNo, false);
    } finally {
      setIsLoading(false);
    }
  };

  const handlePreview = () => {
    setPreviewOpen(true);
  };

  const handlePublish = async () => {
    setIsLoading(true);
    try {
      // 模拟发布操作
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      console.log('Publishing survey:', {
        ...surveyInfo,
        jsonDefinition: surveyJson,
        status: 'active'
      });
      
      toast({
        title: "发布成功",
        description: "问卷已成功发布",
      });
      
      router.push('/survey');
    } catch (error) {
      console.error('Failed to publish survey:', error);
      toast({
        title: "发布失败",
        description: "请重试",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
      setPublishOpen(false);
    }
  };

  const renderBasicInfo = () => (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="title">问卷标题 *</Label>
        <Input
          id="title"
          value={surveyInfo.title}
          onChange={(e) => setSurveyInfo({ ...surveyInfo, title: e.target.value })}
          placeholder="请输入问卷标题"
        />
      </div>
      
      <div className="space-y-2">
        <Label htmlFor="description">问卷描述</Label>
        <Textarea
          id="description"
          value={surveyInfo.description}
          onChange={(e) => setSurveyInfo({ ...surveyInfo, description: e.target.value })}
          placeholder="请输入问卷描述"
          rows={3}
        />
      </div>
      
      <div className="space-y-2">
        <Label>标签</Label>
        <div className="flex flex-wrap gap-2">
          {surveyInfo.tags.map((tag, index) => (
            <Badge key={index} variant="secondary">
              {tag}
              <button 
                className="ml-2"
                onClick={() => {
                  const newTags = [...surveyInfo.tags];
                  newTags.splice(index, 1);
                  setSurveyInfo({ ...surveyInfo, tags: newTags });
                }}
              >
                ×
              </button>
            </Badge>
          ))}
          <Input
            placeholder="添加标签"
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                const target = e.target as HTMLInputElement;
                if (target.value.trim()) {
                  setSurveyInfo({ 
                    ...surveyInfo, 
                    tags: [...surveyInfo.tags, target.value.trim()] 
                  });
                  target.value = '';
                }
              }
            }}
          />
        </div>
      </div>
    </div>
  );

  const renderSurveyDesign = () => (
    <div className="space-y-6">
      <SurveyCreatorComponent 
        json={surveyJson}
        onSurveyChange={handleSurveyChange}
        onSave={handleSave}
      />
      
      <div className="flex justify-end space-x-2">
        <Button variant="outline" onClick={handlePreview}>
          <Eye className="h-4 w-4 mr-2" />
          预览问卷
        </Button>
        <Button onClick={() => handleSave(1, () => {})}>
          <Save className="h-4 w-4 mr-2" />
          保存草稿
        </Button>
      </div>
    </div>
  );

  const renderPublishSettings = () => (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Label htmlFor="isPublic">公开问卷</Label>
          <Switch
            id="isPublic"
            checked={surveyInfo.isPublic}
            onCheckedChange={(checked) => setSurveyInfo({ ...surveyInfo, isPublic: checked })}
          />
        </div>
        <p className="text-sm text-muted-foreground">
          公开问卷允许任何人访问，非公开问卷需要邀请码才能访问
        </p>
      </div>
      
      <div className="space-y-2">
        <Label htmlFor="maxResponses">最大响应数</Label>
        <Input
          id="maxResponses"
          type="number"
          value={surveyInfo.maxResponses || ''}
          onChange={(e) => setSurveyInfo({ ...surveyInfo, maxResponses: parseInt(e.target.value) || 0 })}
          placeholder="0表示无限制"
        />
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="startTime">开始时间</Label>
          <div className="flex items-center space-x-2">
            <Input
              id="startTime"
              type="datetime-local"
              value={surveyInfo.startTime}
              onChange={(e) => setSurveyInfo({ ...surveyInfo, startTime: e.target.value })}
            />
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="endTime">结束时间</Label>
          <div className="flex items-center space-x-2">
            <Input
              id="endTime"
              type="datetime-local"
              value={surveyInfo.endTime}
              onChange={(e) => setSurveyInfo({ ...surveyInfo, endTime: e.target.value })}
            />
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </div>
        </div>
      </div>
    </div>
  );

  const getCurrentStepContent = () => {
    switch (activeStep) {
      case 0: return renderBasicInfo();
      case 1: return renderSurveyDesign();
      case 2: return renderPublishSettings();
      default: return null;
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和返回按钮 */}
      <div className="flex items-center space-x-4">
        <Button variant="outline" size="icon" onClick={() => router.back()}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-3xl font-bold">创建问卷</h1>
          <p className="text-muted-foreground">通过向导创建新的问卷调查</p>
        </div>
      </div>

      {/* 步骤指示器 */}
      <div className="flex justify-center">
        <div className="flex items-center space-x-4">
          {steps.map((step, index) => (
            <div key={index} className="flex items-center">
              <div className={`flex items-center justify-center w-8 h-8 rounded-full ${
                index === activeStep 
                  ? 'bg-primary text-primary-foreground' 
                  : index < activeStep 
                    ? 'bg-secondary text-secondary-foreground' 
                    : 'bg-muted text-muted-foreground'
              }`}>
                {index < activeStep ? '✓' : index + 1}
              </div>
              <span className={`ml-2 ${index === activeStep ? 'font-medium' : 'text-muted-foreground'}`}>
                {step}
              </span>
              {index < steps.length - 1 && (
                <div className="w-8 h-0.5 bg-muted mx-2"></div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* 步骤内容 */}
      <Card>
        <CardHeader>
          <CardTitle>{steps[activeStep]}</CardTitle>
          <CardDescription>
            {activeStep === 0 && '填写问卷的基本信息'}
            {activeStep === 1 && '设计问卷内容和问题'}
            {activeStep === 2 && '设置问卷的发布选项'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {getCurrentStepContent()}
        </CardContent>
      </Card>

      {/* 操作按钮 */}
      <div className="flex justify-between">
        <div>
          {activeStep > 0 && (
            <Button variant="outline" onClick={handleBack}>
              上一步
            </Button>
          )}
        </div>
        <div className="space-x-2">
          {activeStep < steps.length - 1 ? (
            <Button onClick={handleNext}>
              下一步
              <ArrowLeft className="h-4 w-4 mr-2 rotate-180" />
            </Button>
          ) : (
            <Button onClick={() => setPublishOpen(true)}>
              <Send className="h-4 w-4 mr-2" />
              发布问卷
            </Button>
          )}
        </div>
      </div>

      {/* 预览对话框 */}
      <Dialog open={previewOpen} onOpenChange={setPreviewOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>问卷预览</DialogTitle>
            <DialogDescription>
              预览问卷内容和样式
            </DialogDescription>
          </DialogHeader>
          <div className="overflow-y-auto max-h-[60vh]">
            <div className="prose max-w-none p-4">
              <h2>{surveyInfo.title || '问卷标题'}</h2>
              <p>{surveyInfo.description || '问卷描述'}</p>
              <div className="bg-muted p-4 rounded-lg my-4">
                <h3>问卷内容</h3>
                <p>在实际项目中，这里将显示问卷的完整预览</p>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setPreviewOpen(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 发布确认对话框 */}
      <Dialog open={publishOpen} onOpenChange={setPublishOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认发布问卷</DialogTitle>
            <DialogDescription>
              确认发布问卷 "{surveyInfo.title}" 吗？
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="flex items-center space-x-2 rounded-lg border p-4">
              <FileText className="h-5 w-5 text-muted-foreground" />
              <div>
                <div className="font-medium">{surveyInfo.title || '问卷标题'}</div>
                <div className="text-sm text-muted-foreground">
                  {surveyJson.pages[0]?.elements?.length || 0} 个问题
                </div>
              </div>
            </div>
            <div className="text-sm text-muted-foreground">
              发布后问卷将对用户可见，您可以在问卷管理页面查看响应数据。
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPublishOpen(false)}>
              取消
            </Button>
            <Button onClick={handlePublish} disabled={isLoading}>
              {isLoading ? '发布中...' : '确认发布'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}