'use client';

import React, { useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Users, 
  FileText, 
  Bot, 
  QrCode, 
  Archive, 
  BarChart3,
  CheckCircle,
  Bell,
  Upload,
  Download,
  ClipboardList,
  HelpCircle
} from 'lucide-react';
import Link from 'next/link';
import { TodoCard } from '@/components/dashboard/TodoCard';
import { CalendarCard } from '@/components/dashboard/CalendarCard';
import { useNotifications } from '@/hooks/useNotifications';

export default function Home() {
  // 启用通知轮询，仅获取日程提醒并自动标记已读
  const { unreadCount } = useNotifications({
    pollingInterval: 60000, // 每60秒轮询一次
    limit: 20,
    unreadOnly: true, // 只获取未读通知
    autoMarkAsRead: true, // 自动标记已读
  });
  const features = [
    {
      title: '员工管理',
      description: '管理员工信息，支持数据导入导出、多选、排序、筛选、行内编辑等功能',
      icon: <Users className="h-8 w-8" />,
      link: '/employee-management',
    },
    {
      title: '文档管理',
      description: '管理企业文档，支持版本控制、权限管理、在线预览、标签分类等功能',
      icon: <FileText className="h-8 w-8" />,
      link: '/document-management',
    },
    {
      title: '电子合同',
      description: '强大的电子合同签署平台，支持多方签署、CA证书、区块链存证等功能',
      icon: <ClipboardList className="h-8 w-8" />,
      link: '/contracts',
      badge: 'NEW',
      badgeVariant: 'success' as const
    },
    {
      title: 'AI助手',
      description: '集成Dify AI平台，提供智能问答、文档处理和知识管理能力',
      icon: <Bot className="h-8 w-8" />,
      link: '/ai-assistant',
    },
    {
      title: '调查问卷',
      description: '采用SurveyJS无缝集成，支持问卷创建、数据收集、智能分析和报告生成',
      icon: <HelpCircle className="h-8 w-8" />,
      link: '/survey',
      badge: 'NEW',
      badgeVariant: 'default' as const
    },
    {
      title: '二维码应用',
      description: '支持动态表单、员工签到、在线订餐、问卷调查和访客登记等应用场景',
      icon: <QrCode className="h-8 w-8" />,
      link: '/qrcode',
    },
    {
      title: '审批管理',
      description: '管理文档审批流程，支持自定义审批模板和多级审批',
      icon: <CheckCircle className="h-8 w-8" />,
      link: '/approval-management',
    },
    {
      title: '通知中心',
      description: '集中管理各类系统通知，支持个性化设置和批量操作',
      icon: <Bell className="h-8 w-8" />,
      link: '/notification-center',
    },
    {
      title: '知识库归档',
      description: '自动归档知识库文档，支持自定义归档规则和资料目录生成',
      icon: <Archive className="h-8 w-8" />,
      link: '/archive',
    },
    {
      title: '数据统计',
      description: '提供全面的数据统计和分析功能，帮助管理者了解企业运营状况',
      icon: <BarChart3 className="h-8 w-8" />,
      link: '/statistics',
    },
  ];

  return (
    <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          CDK-Office 企业内容管理平台
        </h1>
        <p className="text-xl text-gray-600 max-w-3xl mx-auto mb-6">
          集成Dify AI平台，实现智能文档管理、AI问答和知识库管理功能
        </p>
        <Button asChild size="lg" className="mb-4">
          <Link href="/app-center">
            <BarChart3 className="mr-2 h-5 w-5" />
            浏览全部应用
          </Link>
        </Button>
      </div>
      
      {/* 待办事项和日程卡片 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <TodoCard />
        <CalendarCard />
      </div>
      
      {/* 功能应用卡片 */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-12">
        {features.map((feature, index) => (
          <Card 
            key={index}
            className={`relative h-full transition-all duration-200 hover:shadow-lg ${
              feature.badgeVariant === 'success' ? 'border-green-500 shadow-green-100' : ''
            }`}
          >
            {feature.badge && (
              <div className="absolute top-3 right-3 z-10">
                <Badge variant={feature.badgeVariant || 'default'}>
                  {feature.badge}
                </Badge>
              </div>
            )}
            <CardHeader className="text-center">
              <div className="flex justify-center mb-3 text-blue-600">
                {feature.icon}
              </div>
              <CardTitle className="text-lg">{feature.title}</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col">
              <p className="text-sm text-gray-600 mb-4 flex-1">
                {feature.description}
              </p>
              <Button 
                asChild 
                className={`w-full ${
                  feature.badgeVariant === 'success' ? 'bg-green-600 hover:bg-green-700' : ''
                }`}
                variant={feature.badgeVariant === 'success' ? 'default' : 'default'}
              >
                <Link href={feature.link}>
                  立即体验
                </Link>
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
      
      <div className="text-center">
        <h2 className="text-2xl font-semibold text-gray-900 mb-6">
          快速开始
        </h2>
        <div className="flex flex-col sm:flex-row justify-center gap-4">
          <Button asChild size="lg" className="min-w-40">
            <Link href="/employee-management">
              <Upload className="mr-2 h-4 w-4" />
              导入员工数据
            </Link>
          </Button>
          <Button asChild variant="outline" size="lg" className="min-w-40">
            <Link href="/document-management">
              <Download className="mr-2 h-4 w-4" />
              浏览文档
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}