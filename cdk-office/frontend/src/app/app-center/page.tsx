'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Users, 
  FileText, 
  Bot, 
  QrCode, 
  Archive, 
  BarChart3,
  CheckCircle,
  Bell,
  ClipboardList,
  HelpCircle,
  Search,
  Star,
  Zap,
  Shield,
  Settings,
  Calendar,
  MessageSquare,
  Globe,
  Database,
  Camera,
  Smartphone
} from 'lucide-react';
import type { AppCenterCard } from '@/types/contract';

export default function AppCenter() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');

  // 应用数据
  const applications: AppCenterCard[] = [
    // 核心应用
    {
      id: 'employee-management',
      title: '员工管理',
      description: '管理员工信息，支持数据导入导出、多选、排序、筛选、行内编辑等功能',
      icon: <Users className="h-8 w-8" />,
      link: '/employee-management',
      category: 'core',
      featured: true,
      status: 'active'
    },
    {
      id: 'document-management',
      title: '文档管理',
      description: '管理企业文档，支持版本控制、权限管理、在线预览、标签分类等功能',
      icon: <FileText className="h-8 w-8" />,
      link: '/document-management',
      category: 'core',
      featured: true,
      status: 'active'
    },
    {
      id: 'approval-management',
      title: '审批管理',
      description: '管理文档审批流程，支持自定义审批模板和多级审批',
      icon: <CheckCircle className="h-8 w-8" />,
      link: '/approval-management',
      category: 'core',
      status: 'active'
    },
    
    // AI 应用
    {
      id: 'ai-chat',
      title: '智能问答',
      description: '集成Dify AI平台，提供智能问答、文档处理和知识管理能力',
      icon: <Bot className="h-8 w-8" />,
      link: '/ai-chat',
      category: 'ai',
      badge: 'HOT',
      badgeVariant: 'destructive',
      featured: true,
      status: 'active'
    },
    {
      id: 'intelligent-analysis',
      title: '智能分析',
      description: '基于AI的数据分析和智能报告生成，帮助决策制定',
      icon: <Zap className="h-8 w-8" />,
      link: '/intelligent-analysis',
      category: 'ai',
      badge: 'NEW',
      badgeVariant: 'success',
      status: 'beta'
    },
    {
      id: 'ocr-recognition',
      title: 'OCR文字识别',
      description: '智能图像文字识别，支持多种文档格式和语言',
      icon: <Camera className="h-8 w-8" />,
      link: '/ocr-recognition',
      category: 'ai',
      status: 'active'
    },
    
    // 业务应用
    {
      id: 'contract-management',
      title: '电子合同',
      description: '强大的电子合同签署平台，支持多方签署、CA证书、区块链存证等功能',
      icon: <ClipboardList className="h-8 w-8" />,
      link: '/contracts',
      category: 'business',
      badge: 'NEW',
      badgeVariant: 'success',
      featured: true,
      status: 'active'
    },
    {
      id: 'survey-system',
      title: '调查问卷',
      description: '采用SurveyJS无缝集成，支持问卷创建、数据收集、智能分析和报告生成',
      icon: <HelpCircle className="h-8 w-8" />,
      link: '/survey',
      category: 'business',
      badge: 'NEW',
      badgeVariant: 'default',
      status: 'active'
    },
    {
      id: 'notification-center',
      title: '通知中心',
      description: '集中管理各类系统通知，支持个性化设置和批量操作',
      icon: <Bell className="h-8 w-8" />,
      link: '/notification-center',
      category: 'business',
      status: 'active'
    },
    {
      id: 'calendar-scheduling',
      title: '日程管理',
      description: '智能日程安排和会议调度，支持团队协作和提醒功能',
      icon: <Calendar className="h-8 w-8" />,
      link: '/calendar',
      category: 'business',
      status: 'active'
    },
    {
      id: 'team-chat',
      title: '团队沟通',
      description: '内部即时通讯工具，支持群聊、文件分享和视频会议',
      icon: <MessageSquare className="h-8 w-8" />,
      link: '/team-chat',
      category: 'business',
      status: 'coming_soon'
    },
    
    // 工具应用
    {
      id: 'qrcode-tools',
      title: '二维码应用',
      description: '支持动态表单、员工签到、在线订餐、问卷调查和访客登记等应用场景',
      icon: <QrCode className="h-8 w-8" />,
      link: '/qrcode',
      category: 'tools',
      status: 'active'
    },
    {
      id: 'knowledge-archive',
      title: '知识库归档',
      description: '自动归档知识库文档，支持自定义归档规则和资料目录生成',
      icon: <Archive className="h-8 w-8" />,
      link: '/archive',
      category: 'tools',
      status: 'active'
    },
    {
      id: 'data-statistics',
      title: '数据统计',
      description: '提供全面的数据统计和分析功能，帮助管理者了解企业运营状况',
      icon: <BarChart3 className="h-8 w-8" />,
      link: '/statistics',
      category: 'tools',
      status: 'active'
    },
    {
      id: 'system-settings',
      title: '系统设置',
      description: '系统配置管理，用户权限设置，安全策略配置',
      icon: <Settings className="h-8 w-8" />,
      link: '/settings',
      category: 'tools',
      status: 'active'
    },
    {
      id: 'data-backup',
      title: '数据备份',
      description: '自动化数据备份和恢复，确保数据安全和业务连续性',
      icon: <Database className="h-8 w-8" />,
      link: '/backup',
      category: 'tools',
      status: 'active'
    },
    {
      id: 'mobile-app',
      title: '移动端应用',
      description: '移动办公应用，随时随地处理工作任务',
      icon: <Smartphone className="h-8 w-8" />,
      link: '/mobile',
      category: 'tools',
      badge: 'SOON',
      badgeVariant: 'outline',
      status: 'coming_soon'
    }
  ];

  // 筛选应用
  const filteredApplications = applications.filter(app => {
    const matchesSearch = app.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         app.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = selectedCategory === 'all' || app.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  // 推荐应用
  const featuredApps = applications.filter(app => app.featured);

  // 获取状态配置
  const getStatusConfig = (status?: string) => {
    switch (status) {
      case 'beta':
        return { color: 'bg-yellow-100 text-yellow-800', label: 'Beta' };
      case 'coming_soon':
        return { color: 'bg-gray-100 text-gray-800', label: '即将推出' };
      default:
        return null;
    }
  };

  const categories = [
    { id: 'all', label: '全部', icon: <Globe className="h-4 w-4" /> },
    { id: 'core', label: '核心应用', icon: <Star className="h-4 w-4" /> },
    { id: 'ai', label: 'AI应用', icon: <Bot className="h-4 w-4" /> },
    { id: 'business', label: '业务应用', icon: <Shield className="h-4 w-4" /> },
    { id: 'tools', label: '工具应用', icon: <Settings className="h-4 w-4" /> }
  ];

  return (
    <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* 页面标题 */}
      <div className="text-center mb-8">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          应用中心
        </h1>
        <p className="text-xl text-gray-600 max-w-3xl mx-auto">
          发现和使用CDK-Office平台提供的丰富应用，提升您的工作效率
        </p>
      </div>

      {/* 搜索栏 */}
      <div className="mb-8">
        <div className="relative max-w-md mx-auto">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <Input
            placeholder="搜索应用..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* 推荐应用 */}
      {searchQuery === '' && (
        <div className="mb-12">
          <h2 className="text-2xl font-semibold text-gray-900 mb-6 flex items-center">
            <Star className="h-6 w-6 mr-2 text-yellow-500" />
            推荐应用
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {featuredApps.map((app) => {
              const statusConfig = getStatusConfig(app.status);
              const isDisabled = app.status === 'coming_soon';
              
              return (
                <Card 
                  key={app.id}
                  className={`relative h-full transition-all duration-200 hover:shadow-lg ${
                    app.badgeVariant === 'success' ? 'border-green-500 shadow-green-100' : ''
                  } ${
                    app.badgeVariant === 'destructive' ? 'border-red-500 shadow-red-100' : ''
                  } ${
                    isDisabled ? 'opacity-60' : ''
                  }`}
                >
                  {/* 状态标签 */}
                  {statusConfig && (
                    <div className="absolute top-3 left-3 z-10">
                      <span className={`px-2 py-1 text-xs font-medium rounded-full ${statusConfig.color}`}>
                        {statusConfig.label}
                      </span>
                    </div>
                  )}
                  
                  {/* 特性标签 */}
                  {app.badge && (
                    <div className="absolute top-3 right-3 z-10">
                      <Badge variant={app.badgeVariant || 'default'}>
                        {app.badge}
                      </Badge>
                    </div>
                  )}
                  
                  <CardHeader className="text-center">
                    <div className="flex justify-center mb-3 text-blue-600">
                      {app.icon}
                    </div>
                    <CardTitle className="text-lg">{app.title}</CardTitle>
                  </CardHeader>
                  <CardContent className="flex-1 flex flex-col">
                    <p className="text-sm text-gray-600 mb-4 flex-1">
                      {app.description}
                    </p>
                    <Button 
                      asChild={!isDisabled}
                      disabled={isDisabled}
                      className={`w-full ${
                        app.badgeVariant === 'success' ? 'bg-green-600 hover:bg-green-700' : ''
                      } ${
                        app.badgeVariant === 'destructive' ? 'bg-red-600 hover:bg-red-700' : ''
                      }`}
                      variant={app.badgeVariant === 'success' || app.badgeVariant === 'destructive' ? 'default' : 'default'}
                    >
                      {!isDisabled ? (
                        <Link href={app.link}>
                          立即使用
                        </Link>
                      ) : (
                        <span>即将推出</span>
                      )}
                    </Button>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      )}

      {/* 应用分类 */}
      <div className="mb-8">
        <Tabs value={selectedCategory} onValueChange={setSelectedCategory}>
          <TabsList className="grid w-full grid-cols-5">
            {categories.map((category) => (
              <TabsTrigger key={category.id} value={category.id} className="flex items-center gap-2">
                {category.icon}
                <span className="hidden sm:inline">{category.label}</span>
              </TabsTrigger>
            ))}
          </TabsList>
          
          {categories.map((category) => (
            <TabsContent key={category.id} value={category.id} className="mt-8">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {filteredApplications
                  .filter(app => category.id === 'all' || app.category === category.id)
                  .map((app) => {
                    const statusConfig = getStatusConfig(app.status);
                    const isDisabled = app.status === 'coming_soon';
                    
                    return (
                      <Card 
                        key={app.id}
                        className={`relative h-full transition-all duration-200 hover:shadow-lg ${
                          isDisabled ? 'opacity-60' : ''
                        }`}
                      >
                        {/* 状态标签 */}
                        {statusConfig && (
                          <div className="absolute top-3 left-3 z-10">
                            <span className={`px-2 py-1 text-xs font-medium rounded-full ${statusConfig.color}`}>
                              {statusConfig.label}
                            </span>
                          </div>
                        )}
                        
                        {/* 特性标签 */}
                        {app.badge && (
                          <div className="absolute top-3 right-3 z-10">
                            <Badge variant={app.badgeVariant || 'default'}>
                              {app.badge}
                            </Badge>
                          </div>
                        )}
                        
                        <CardHeader className="text-center">
                          <div className="flex justify-center mb-3 text-blue-600">
                            {app.icon}
                          </div>
                          <CardTitle className="text-lg">{app.title}</CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 flex flex-col">
                          <p className="text-sm text-gray-600 mb-4 flex-1">
                            {app.description}
                          </p>
                          <Button 
                            asChild={!isDisabled}
                            disabled={isDisabled}
                            className="w-full"
                            variant="default"
                          >
                            {!isDisabled ? (
                              <Link href={app.link}>
                                立即使用
                              </Link>
                            ) : (
                              <span>即将推出</span>
                            )}
                          </Button>
                        </CardContent>
                      </Card>
                    );
                  })}
              </div>
              
              {/* 无搜索结果 */}
              {filteredApplications.filter(app => category.id === 'all' || app.category === category.id).length === 0 && (
                <div className="text-center py-12">
                  <Search className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">未找到相关应用</h3>
                  <p className="text-gray-600">
                    {searchQuery ? `没有找到包含"${searchQuery}"的应用` : '该分类下暂无应用'}
                  </p>
                </div>
              )}
            </TabsContent>
          ))}
        </Tabs>
      </div>

      {/* 统计信息 */}
      <div className="mt-12 pt-8 border-t border-gray-200">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
          <div>
            <div className="text-3xl font-bold text-blue-600">{applications.length}</div>
            <div className="text-sm text-gray-600">总应用数</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-green-600">
              {applications.filter(app => app.status === 'active').length}
            </div>
            <div className="text-sm text-gray-600">可用应用</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-yellow-600">
              {applications.filter(app => app.status === 'beta').length}
            </div>
            <div className="text-sm text-gray-600">测试版应用</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-gray-600">
              {applications.filter(app => app.status === 'coming_soon').length}
            </div>
            <div className="text-sm text-gray-600">即将推出</div>
          </div>
        </div>
      </div>
    </div>
  );
}