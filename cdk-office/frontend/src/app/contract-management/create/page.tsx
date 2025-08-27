'use client';

import React, { useState } from 'react';
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
  Checkbox 
} from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { 
  Avatar, 
  AvatarFallback 
} from '@/components/ui/avatar';
import { 
  toast 
} from '@/components/ui/use-toast';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

import { 
  ArrowLeft, 
  Plus, 
  Trash2, 
  User, 
  Building, 
  Upload, 
  Eye, 
  Save, 
  Send,
  Calendar,
  Shield,
  Server
} from 'lucide-react';

interface Signer {
  id: string;
  signerType: 'person' | 'company';
  name: string;
  email: string;
  phone: string;
  idCard?: string;
  companyName?: string;
  unifiedCode?: string;
  signOrder: number;
  signType: 'signature' | 'seal';
}

interface ContractForm {
  title: string;
  description: string;
  templateId: string;
  content: string;
  signMode: 'sequential' | 'parallel';
  requireCA: boolean;
  requireBlockchain: boolean;
  expireTime: Date | null;
  signers: Signer[];
}

const CONTRACT_TEMPLATES = [
  { id: '1', name: '劳动合同模板', category: '人事合同' },
  { id: '2', name: '供应商协议模板', category: '商务合同' },
  { id: '3', name: '保密协议模板', category: '法务合同' },
  { id: '4', name: '租赁合同模板', category: '房产合同' },
  { id: '5', name: '服务合同模板', category: '服务合同' }
];

export default function CreateContract() {
  const router = useRouter();
  const [activeStep, setActiveStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [previewDialogOpen, setPreviewDialogOpen] = useState(false);

  const [formData, setFormData] = useState<ContractForm>({
    title: '',
    description: '',
    templateId: '',
    content: '',
    signMode: 'sequential',
    requireCA: true,
    requireBlockchain: false,
    expireTime: null,
    signers: []
  });

  const steps = ['基本信息', '合同内容', '签署人设置', '高级设置', '预览提交'];

  // 添加签署人
  const addSigner = () => {
    const newSigner: Signer = {
      id: Date.now().toString(),
      signerType: 'person',
      name: '',
      email: '',
      phone: '',
      signOrder: formData.signers.length + 1,
      signType: 'signature'
    };
    setFormData({
      ...formData,
      signers: [...formData.signers, newSigner]
    });
  };

  // 删除签署人
  const removeSigner = (id: string) => {
    const updatedSigners = formData.signers
      .filter(signer => signer.id !== id)
      .map((signer, index) => ({ ...signer, signOrder: index + 1 }));
    
    setFormData({
      ...formData,
      signers: updatedSigners
    });
  };

  // 更新签署人信息
  const updateSigner = (id: string, field: keyof Signer, value: any) => {
    const updatedSigners = formData.signers.map(signer =>
      signer.id === id ? { ...signer, [field]: value } : signer
    );
    setFormData({
      ...formData,
      signers: updatedSigners
    });
  };

  // 下一步
  const handleNext = () => {
    if (validateCurrentStep()) {
      setActiveStep(prev => prev + 1);
    }
  };

  // 上一步
  const handleBack = () => {
    setActiveStep(prev => prev - 1);
  };

  // 验证当前步骤
  const validateCurrentStep = () => {
    switch (activeStep) {
      case 0:
        return formData.title.trim() && formData.description.trim();
      case 1:
        return formData.content.trim() || formData.templateId;
      case 2:
        return formData.signers.length > 0 && formData.signers.every(s => s.name.trim() && s.email.trim());
      case 3:
        return formData.expireTime;
      default:
        return true;
    }
  };

  // 保存草稿
  const handleSaveDraft = async () => {
    setLoading(true);
    // 实现保存草稿逻辑
    setTimeout(() => {
      setLoading(false);
      toast({
        title: "保存成功",
        description: "合同草稿已保存",
      });
      router.push('/contract-management');
    }, 1000);
  };

  // 发送签署
  const handleSendContract = async () => {
    setLoading(true);
    // 实现发送合同逻辑
    setTimeout(() => {
      setLoading(false);
      toast({
        title: "发送成功",
        description: "合同已发送给签署人",
      });
      router.push('/contract-management');
    }, 1000);
  };

  const renderBasicInfo = () => (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="title">合同标题 *</Label>
        <Input
          id="title"
          value={formData.title}
          onChange={(e) => setFormData({ ...formData, title: e.target.value })}
          placeholder="请输入合同标题"
        />
      </div>
      
      <div className="space-y-2">
        <Label htmlFor="description">合同描述</Label>
        <Textarea
          id="description"
          value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          placeholder="请输入合同描述"
          rows={3}
        />
      </div>
      
      <div className="space-y-2">
        <Label>合同模板</Label>
        <Select 
          value={formData.templateId} 
          onValueChange={(value) => setFormData({ ...formData, templateId: value })}
        >
          <SelectTrigger>
            <SelectValue placeholder="选择合同模板" />
          </SelectTrigger>
          <SelectContent>
            {CONTRACT_TEMPLATES.map((template) => (
              <SelectItem key={template.id} value={template.id}>
                {template.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-sm text-muted-foreground">选择预设模板可快速创建合同</p>
      </div>
    </div>
  );

  const renderContractContent = () => (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="content">合同内容</Label>
        <Textarea
          id="content"
          value={formData.content}
          onChange={(e) => setFormData({ ...formData, content: e.target.value })}
          placeholder="请输入合同内容或使用模板"
          rows={10}
        />
      </div>
      
      <div className="flex items-center space-x-2">
        <Button variant="outline">
          <Upload className="h-4 w-4 mr-2" />
          上传文件
        </Button>
        <Button variant="outline">
          <Eye className="h-4 w-4 mr-2" />
          预览模板
        </Button>
      </div>
    </div>
  );

  const renderSigners = () => (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-lg font-medium">签署人列表</h3>
          <p className="text-sm text-muted-foreground">添加需要签署合同的人员或企业</p>
        </div>
        <Button onClick={addSigner}>
          <Plus className="h-4 w-4 mr-2" />
          添加签署人
        </Button>
      </div>
      
      <div className="space-y-4">
        {formData.signers.map((signer, index) => (
          <Card key={signer.id}>
            <CardContent className="p-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center space-x-2">
                  <Avatar>
                    <AvatarFallback>{index + 1}</AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium">签署人 {index + 1}</div>
                    <div className="text-sm text-muted-foreground">
                      {signer.signerType === 'person' ? '个人签署' : '企业签署'}
                    </div>
                  </div>
                </div>
                <Button variant="ghost" size="icon" onClick={() => removeSigner(signer.id)}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>签署人类型</Label>
                  <Select 
                    value={signer.signerType} 
                    onValueChange={(value) => updateSigner(signer.id, 'signerType', value)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="person">
                        <div className="flex items-center">
                          <User className="h-4 w-4 mr-2" />
                          个人
                        </div>
                      </SelectItem>
                      <SelectItem value="company">
                        <div className="flex items-center">
                          <Building className="h-4 w-4 mr-2" />
                          企业
                        </div>
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                
                <div className="space-y-2">
                  <Label>签署方式</Label>
                  <Select 
                    value={signer.signType} 
                    onValueChange={(value) => updateSigner(signer.id, 'signType', value)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="signature">电子签名</SelectItem>
                      <SelectItem value="seal">电子印章</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                
                {signer.signerType === 'person' ? (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor={`name-${signer.id}`}>姓名 *</Label>
                      <Input
                        id={`name-${signer.id}`}
                        value={signer.name}
                        onChange={(e) => updateSigner(signer.id, 'name', e.target.value)}
                        placeholder="请输入姓名"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor={`idCard-${signer.id}`}>身份证号</Label>
                      <Input
                        id={`idCard-${signer.id}`}
                        value={signer.idCard || ''}
                        onChange={(e) => updateSigner(signer.id, 'idCard', e.target.value)}
                        placeholder="请输入身份证号"
                      />
                    </div>
                  </>
                ) : (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor={`companyName-${signer.id}`}>企业名称 *</Label>
                      <Input
                        id={`companyName-${signer.id}`}
                        value={signer.companyName || ''}
                        onChange={(e) => updateSigner(signer.id, 'companyName', e.target.value)}
                        placeholder="请输入企业名称"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor={`unifiedCode-${signer.id}`}>统一社会信用代码</Label>
                      <Input
                        id={`unifiedCode-${signer.id}`}
                        value={signer.unifiedCode || ''}
                        onChange={(e) => updateSigner(signer.id, 'unifiedCode', e.target.value)}
                        placeholder="请输入统一社会信用代码"
                      />
                    </div>
                  </>
                )}
                
                <div className="space-y-2">
                  <Label htmlFor={`email-${signer.id}`}>邮箱 *</Label>
                  <Input
                    id={`email-${signer.id}`}
                    type="email"
                    value={signer.email}
                    onChange={(e) => updateSigner(signer.id, 'email', e.target.value)}
                    placeholder="请输入邮箱"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor={`phone-${signer.id}`}>手机号</Label>
                  <Input
                    id={`phone-${signer.id}`}
                    type="tel"
                    value={signer.phone}
                    onChange={(e) => updateSigner(signer.id, 'phone', e.target.value)}
                    placeholder="请输入手机号"
                  />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
      
      {formData.signers.length === 0 && (
        <div className="text-center py-8 text-muted-foreground">
          <User className="h-12 w-12 mx-auto mb-4" />
          <p>暂无签署人</p>
          <Button className="mt-4" onClick={addSigner}>
            <Plus className="h-4 w-4 mr-2" />
            添加第一个签署人
          </Button>
        </div>
      )}
    </div>
  );

  const renderAdvancedSettings = () => (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center space-x-2">
          <Checkbox
            id="requireCA"
            checked={formData.requireCA}
            onCheckedChange={(checked) => setFormData({ ...formData, requireCA: checked as boolean })}
          />
          <Label htmlFor="requireCA" className="flex items-center">
            <Shield className="h-4 w-4 mr-2" />
            需要数字证书认证
          </Label>
        </div>
        <p className="text-sm text-muted-foreground ml-6">
          启用数字证书认证可提高合同法律效力
        </p>
      </div>
      
      <div className="space-y-4">
        <div className="flex items-center space-x-2">
          <Checkbox
            id="requireBlockchain"
            checked={formData.requireBlockchain}
            onCheckedChange={(checked) => setFormData({ ...formData, requireBlockchain: checked as boolean })}
          />
          <Label htmlFor="requireBlockchain" className="flex items-center">
            <Server className="h-4 w-4 mr-2" />
            需要区块链存证
          </Label>
        </div>
        <p className="text-sm text-muted-foreground ml-6">
          启用区块链存证可确保合同不可篡改
        </p>
      </div>
      
      <div className="space-y-2">
        <Label>签署顺序</Label>
        <Select 
          value={formData.signMode} 
          onValueChange={(value) => setFormData({ ...formData, signMode: value as 'sequential' | 'parallel' })}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="sequential">顺序签署</SelectItem>
            <SelectItem value="parallel">并行签署</SelectItem>
          </SelectContent>
        </Select>
      </div>
      
      <div className="space-y-2">
        <Label htmlFor="expireTime">合同到期时间</Label>
        <div className="flex items-center space-x-2">
          <Input
            id="expireTime"
            type="datetime-local"
            value={formData.expireTime ? formData.expireTime.toISOString().slice(0, 16) : ''}
            onChange={(e) => setFormData({ 
              ...formData, 
              expireTime: e.target.value ? new Date(e.target.value) : null 
            })}
          />
          <Calendar className="h-4 w-4 text-muted-foreground" />
        </div>
        <p className="text-sm text-muted-foreground">
          合同将在指定时间后自动过期
        </p>
      </div>
    </div>
  );

  const renderPreview = () => (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>合同信息预览</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label className="text-muted-foreground">合同标题</Label>
              <div className="font-medium">{formData.title || '未填写'}</div>
            </div>
            <div>
              <Label className="text-muted-foreground">合同描述</Label>
              <div>{formData.description || '未填写'}</div>
            </div>
            <div>
              <Label className="text-muted-foreground">签署方式</Label>
              <div>{formData.signMode === 'sequential' ? '顺序签署' : '并行签署'}</div>
            </div>
            <div>
              <Label className="text-muted-foreground">数字证书</Label>
              <div>{formData.requireCA ? '需要' : '不需要'}</div>
            </div>
            <div>
              <Label className="text-muted-foreground">区块链存证</Label>
              <div>{formData.requireBlockchain ? '需要' : '不需要'}</div>
            </div>
            <div>
              <Label className="text-muted-foreground">到期时间</Label>
              <div>{formData.expireTime ? formData.expireTime.toLocaleString() : '未设置'}</div>
            </div>
          </div>
        </CardContent>
      </Card>
      
      <Card>
        <CardHeader>
          <CardTitle>签署人信息 ({formData.signers.length}人)</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {formData.signers.map((signer, index) => (
            <div key={signer.id} className="flex items-center justify-between p-3 border rounded-lg">
              <div className="flex items-center space-x-3">
                <Avatar>
                  <AvatarFallback>{index + 1}</AvatarFallback>
                </Avatar>
                <div>
                  <div className="font-medium">{signer.name || signer.companyName || '未填写'}</div>
                  <div className="text-sm text-muted-foreground">
                    {signer.signerType === 'person' ? '个人' : '企业'} • {signer.email}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="text-sm">{signer.signType === 'signature' ? '电子签名' : '电子印章'}</div>
                <div className="text-xs text-muted-foreground">
                  签署顺序: {signer.signOrder}
                </div>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );

  const getCurrentStepContent = () => {
    switch (activeStep) {
      case 0: return renderBasicInfo();
      case 1: return renderContractContent();
      case 2: return renderSigners();
      case 3: return renderAdvancedSettings();
      case 4: return renderPreview();
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
          <h1 className="text-3xl font-bold">创建电子合同</h1>
          <p className="text-muted-foreground">通过向导创建新的电子合同</p>
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
            {activeStep === 0 && '填写合同的基本信息'}
            {activeStep === 1 && '编辑合同内容或选择模板'}
            {activeStep === 2 && '添加需要签署合同的人员或企业'}
            {activeStep === 3 && '设置合同的高级选项'}
            {activeStep === 4 && '预览合同信息并提交'}
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
            <>
              <Button variant="outline" onClick={handleSaveDraft} disabled={loading}>
                <Save className="h-4 w-4 mr-2" />
                保存草稿
              </Button>
              <Button onClick={handleNext} disabled={loading}>
                下一步
                <ArrowLeft className="h-4 w-4 mr-2 rotate-180" />
              </Button>
            </>
          ) : (
            <>
              <Button variant="outline" onClick={() => setPreviewDialogOpen(true)}>
                <Eye className="h-4 w-4 mr-2" />
                预览合同
              </Button>
              <Button onClick={handleSendContract} disabled={loading}>
                <Send className="h-4 w-4 mr-2" />
                发送签署
              </Button>
            </>
          )}
        </div>
      </div>

      {/* 预览对话框 */}
      <Dialog open={previewDialogOpen} onOpenChange={setPreviewDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>合同预览</DialogTitle>
            <DialogDescription>
              预览合同内容和签署人信息
            </DialogDescription>
          </DialogHeader>
          <div className="overflow-y-auto max-h-[60vh]">
            <div className="prose max-w-none p-4">
              <h2>{formData.title}</h2>
              <p>{formData.description}</p>
              <div className="bg-muted p-4 rounded-lg my-4">
                <h3>合同内容</h3>
                <p>{formData.content || '合同内容将在此处显示'}</p>
              </div>
              <h3>签署人信息</h3>
              <ul>
                {formData.signers.map((signer, index) => (
                  <li key={signer.id}>
                    {index + 1}. {signer.name || signer.companyName} ({signer.email})
                  </li>
                ))}
              </ul>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPreviewDialogOpen(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}