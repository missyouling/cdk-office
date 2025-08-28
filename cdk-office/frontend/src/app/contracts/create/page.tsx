'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { 
  ArrowLeft, 
  Save, 
  Send, 
  Plus,
  Trash2,
  FileText,
  Users,
  Calendar
} from 'lucide-react';

export default function CreateContractPage() {
  const router = useRouter();
  const [contractData, setContractData] = useState({
    title: '',
    description: '',
    expireTime: '',
    signers: [{ name: '', email: '', signerType: 'person' }]
  });

  const handleAddSigner = () => {
    setContractData(prev => ({
      ...prev,
      signers: [...prev.signers, { name: '', email: '', signerType: 'person' }]
    }));
  };

  const handleRemoveSigner = (index: number) => {
    setContractData(prev => ({
      ...prev,
      signers: prev.signers.filter((_, i) => i !== index)
    }));
  };

  const handleSave = () => {
    // TODO: 实现保存逻辑
    router.push('/contracts');
  };

  const handleSend = () => {
    // TODO: 实现发送逻辑
    router.push('/contracts');
  };

  return (
    <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* 页面头部 */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center space-x-4">
          <Button 
            variant="outline" 
            size="sm"
            onClick={() => router.back()}
            className="flex items-center gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            返回
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">创建合同</h1>
            <p className="text-gray-600 mt-1">填写合同基本信息和签署方信息</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" onClick={handleSave} className="flex items-center gap-2">
            <Save className="h-4 w-4" />
            保存草稿
          </Button>
          <Button onClick={handleSend} className="flex items-center gap-2">
            <Send className="h-4 w-4" />
            发送签署
          </Button>
        </div>
      </div>

      <div className="space-y-6">
        {/* 基本信息 */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              基本信息
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="title">合同标题 *</Label>
              <Input 
                id="title"
                placeholder="请输入合同标题"
                value={contractData.title}
                onChange={(e) => setContractData(prev => ({ ...prev, title: e.target.value }))}
              />
            </div>
            <div>
              <Label htmlFor="description">合同描述</Label>
              <Textarea 
                id="description"
                placeholder="请输入合同描述"
                value={contractData.description}
                onChange={(e) => setContractData(prev => ({ ...prev, description: e.target.value }))}
              />
            </div>
            <div>
              <Label htmlFor="expireTime">过期时间 *</Label>
              <Input 
                id="expireTime"
                type="datetime-local"
                value={contractData.expireTime}
                onChange={(e) => setContractData(prev => ({ ...prev, expireTime: e.target.value }))}
              />
            </div>
          </CardContent>
        </Card>

        {/* 签署方信息 */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              签署方信息
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {contractData.signers.map((signer, index) => (
              <div key={index} className="p-4 border rounded-lg">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="font-medium">签署方 {index + 1}</h4>
                  {contractData.signers.length > 1 && (
                    <Button 
                      variant="outline" 
                      size="sm"
                      onClick={() => handleRemoveSigner(index)}
                      className="text-red-600 hover:text-red-700"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor={`signer-name-${index}`}>姓名 *</Label>
                    <Input 
                      id={`signer-name-${index}`}
                      placeholder="请输入签署方姓名"
                      value={signer.name}
                      onChange={(e) => {
                        const newSigners = [...contractData.signers];
                        newSigners[index].name = e.target.value;
                        setContractData(prev => ({ ...prev, signers: newSigners }));
                      }}
                    />
                  </div>
                  <div>
                    <Label htmlFor={`signer-email-${index}`}>邮箱</Label>
                    <Input 
                      id={`signer-email-${index}`}
                      type="email"
                      placeholder="请输入邮箱地址"
                      value={signer.email}
                      onChange={(e) => {
                        const newSigners = [...contractData.signers];
                        newSigners[index].email = e.target.value;
                        setContractData(prev => ({ ...prev, signers: newSigners }));
                      }}
                    />
                  </div>
                </div>
              </div>
            ))}
            
            <Button 
              variant="outline" 
              onClick={handleAddSigner}
              className="w-full flex items-center gap-2"
            >
              <Plus className="h-4 w-4" />
              添加签署方
            </Button>
          </CardContent>
        </Card>

        {/* 合同内容 */}
        <Card>
          <CardHeader>
            <CardTitle>合同内容</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-12 bg-gray-50 rounded-lg">
              <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">合同编辑器</h3>
              <p className="text-gray-600 mb-4">
                这里将集成富文本编辑器或模板选择器
              </p>
              <p className="text-sm text-gray-500">
                功能开发中，敬请期待...
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}