'use client';

import React from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  ArrowLeft, 
  Download, 
  Edit, 
  Send, 
  FileSignature,
  Users,
  Calendar,
  Clock,
  CheckCircle
} from 'lucide-react';

export default function ContractDetailPage() {
  const params = useParams();
  const router = useRouter();
  const contractId = params.id as string;

  return (
    <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
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
            <h1 className="text-3xl font-bold text-gray-900">合同详情</h1>
            <p className="text-gray-600 mt-1">合同ID: {contractId}</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" className="flex items-center gap-2">
            <Download className="h-4 w-4" />
            下载
          </Button>
          <Button variant="outline" className="flex items-center gap-2">
            <Edit className="h-4 w-4" />
            编辑
          </Button>
          <Button className="flex items-center gap-2">
            <FileSignature className="h-4 w-4" />
            签署
          </Button>
        </div>
      </div>

      {/* 合同信息卡片 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>合同信息</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <h3 className="text-lg font-semibold">员工劳动合同</h3>
                  <p className="text-gray-600">与张三签署的劳动合同</p>
                </div>
                <Separator />
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="font-medium">创建时间：</span>
                    <span className="text-gray-600">2025-01-15</span>
                  </div>
                  <div>
                    <span className="font-medium">过期时间：</span>
                    <span className="text-gray-600">2026-01-15</span>
                  </div>
                  <div>
                    <span className="font-medium">创建者：</span>
                    <span className="text-gray-600">管理员</span>
                  </div>
                  <div>
                    <span className="font-medium">状态：</span>
                    <Badge variant="success" className="ml-2">
                      <CheckCircle className="h-3 w-3 mr-1" />
                      已完成
                    </Badge>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="h-5 w-5" />
                签署方
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 border rounded-lg">
                  <div>
                    <p className="font-medium">张三</p>
                    <p className="text-sm text-gray-600">个人签署</p>
                  </div>
                  <Badge variant="success">已签署</Badge>
                </div>
                <div className="flex items-center justify-between p-3 border rounded-lg">
                  <div>
                    <p className="font-medium">公司HR</p>
                    <p className="text-sm text-gray-600">公司签署</p>
                  </div>
                  <Badge variant="success">已签署</Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* 占位符内容 */}
      <div className="mt-8 text-center py-12 bg-gray-50 rounded-lg">
        <FileSignature className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">合同详情页面</h3>
        <p className="text-gray-600 mb-4">
          这里将显示完整的合同内容、签署历史和相关操作
        </p>
        <p className="text-sm text-gray-500">
          功能开发中，敬请期待...
        </p>
      </div>
    </div>
  );
}