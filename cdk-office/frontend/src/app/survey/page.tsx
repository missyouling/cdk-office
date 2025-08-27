'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
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
  Badge 
} from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { 
  Input 
} from '@/components/ui/input';
import { 
  toast 
} from '@/components/ui/use-toast';

import { 
  Plus, 
  MoreHorizontal, 
  Edit, 
  Trash2, 
  Share, 
  BarChart3, 
  Eye, 
  Play, 
  Square, 
  Copy,
  FileText
} from 'lucide-react';

interface Survey {
  id: string;
  title: string;
  description: string;
  status: 'draft' | 'active' | 'closed' | 'archived';
  responseCount: number;
  viewCount: number;
  createdAt: string;
  updatedAt: string;
}

export default function SurveyPage() {
  const [surveys, setSurveys] = useState<Survey[]>([]);
  const [selectedTab, setSelectedTab] = useState('all');
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedSurvey, setSelectedSurvey] = useState<Survey | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [shareDialogOpen, setShareDialogOpen] = useState(false);
  const [shareUrl, setShareUrl] = useState('');

  // 模拟数据
  useEffect(() => {
    const mockSurveys: Survey[] = [
      {
        id: '1',
        title: '员工满意度调查',
        description: '了解员工对公司的满意度和建议',
        status: 'active',
        responseCount: 42,
        viewCount: 156,
        createdAt: '2025-01-01',
        updatedAt: '2025-01-15'
      },
      {
        id: '2',
        title: '产品需求调研',
        description: '收集用户对新产品功能的需求和意见',
        status: 'draft',
        responseCount: 0,
        viewCount: 8,
        createdAt: '2025-01-10',
        updatedAt: '2025-01-10'
      },
      {
        id: '3',
        title: '培训效果评估',
        description: '评估近期培训课程的效果和改进建议',
        status: 'closed',
        responseCount: 28,
        viewCount: 89,
        createdAt: '2024-12-15',
        updatedAt: '2025-01-05'
      }
    ];
    setSurveys(mockSurveys);
  }, []);

  const tabLabels = [
    { id: 'all', label: '全部' },
    { id: 'active', label: '进行中' },
    { id: 'draft', label: '草稿' },
    { id: 'closed', label: '已结束' },
    { id: 'archived', label: '已归档' }
  ];

  const filteredSurveys = surveys.filter(survey => {
    if (selectedTab === 'all') return true;
    return survey.status === selectedTab;
  });

  const getStatusCounts = () => {
    return {
      all: surveys.length,
      active: surveys.filter(s => s.status === 'active').length,
      draft: surveys.filter(s => s.status === 'draft').length,
      closed: surveys.filter(s => s.status === 'closed').length,
      archived: surveys.filter(s => s.status === 'archived').length
    };
  };

  const statusCounts = getStatusCounts();

  const handleMenuOpen = (survey: Survey) => {
    setSelectedSurvey(survey);
  };

  const handleMenuClose = () => {
    setSelectedSurvey(null);
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
    handleMenuClose();
  };

  const handleShareClick = () => {
    if (selectedSurvey) {
      setShareUrl(`${window.location.origin}/survey/take/${selectedSurvey.id}`);
      setShareDialogOpen(true);
    }
    handleMenuClose();
  };

  const handleCopyShareUrl = () => {
    navigator.clipboard.writeText(shareUrl);
    toast({
      title: "复制成功",
      description: "问卷链接已复制到剪贴板",
    });
    setShareDialogOpen(false);
  };

  const handleDeleteConfirm = () => {
    if (selectedSurvey) {
      setSurveys(surveys.filter(s => s.id !== selectedSurvey.id));
      toast({
        title: "删除成功",
        description: "问卷已删除",
      });
    }
    setDeleteDialogOpen(false);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success';
      case 'draft': return 'secondary';
      case 'closed': return 'warning';
      case 'archived': return 'destructive';
      default: return 'secondary';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '进行中';
      case 'draft': return '草稿';
      case 'closed': return '已结束';
      case 'archived': return '已归档';
      default: return status;
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <Play className="h-4 w-4" />;
      case 'draft': return <FileText className="h-4 w-4" />;
      case 'closed': return <Square className="h-4 w-4" />;
      case 'archived': return <FileText className="h-4 w-4" />;
      default: return <FileText className="h-4 w-4" />;
    }
  };

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">问卷调查</h1>
          <p className="text-muted-foreground mt-2">
            创建和管理问卷调查
          </p>
        </div>
        <Button asChild>
          <Link href="/survey/create">
            <Plus className="h-4 w-4 mr-2" />
            创建问卷
          </Link>
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-blue-100 p-2 mr-3">
                <FileText className="h-5 w-5 text-blue-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">总问卷数</p>
                <p className="text-2xl font-bold">{statusCounts.all}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-green-100 p-2 mr-3">
                <Play className="h-5 w-5 text-green-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">进行中</p>
                <p className="text-2xl font-bold">{statusCounts.active}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-gray-100 p-2 mr-3">
                <FileText className="h-5 w-5 text-gray-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">草稿</p>
                <p className="text-2xl font-bold">{statusCounts.draft}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-yellow-100 p-2 mr-3">
                <Square className="h-5 w-5 text-yellow-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">已结束</p>
                <p className="text-2xl font-bold">{statusCounts.closed}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-red-100 p-2 mr-3">
                <FileText className="h-5 w-5 text-red-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">已归档</p>
                <p className="text-2xl font-bold">{statusCounts.archived}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 标签页 */}
      <Tabs value={selectedTab} onValueChange={setSelectedTab}>
        <TabsList>
          {tabLabels.map((tab) => (
            <TabsTrigger key={tab.id} value={tab.id}>
              {tab.label} {tab.id !== 'all' && `(${statusCounts[tab.id as keyof typeof statusCounts]})`}
            </TabsTrigger>
          ))}
        </TabsList>
        
        <TabsContent value={selectedTab} className="mt-6">
          {filteredSurveys.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <FileText className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <p>暂无问卷数据</p>
              <Button className="mt-4" asChild>
                <Link href="/survey/create">
                  <Plus className="h-4 w-4 mr-2" />
                  创建第一个问卷
                </Link>
              </Button>
            </div>
          ) : (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>问卷标题</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>响应数</TableHead>
                      <TableHead>浏览量</TableHead>
                      <TableHead>更新时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredSurveys.map((survey) => (
                      <TableRow key={survey.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{survey.title}</div>
                            <div className="text-sm text-muted-foreground">{survey.description}</div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getStatusColor(survey.status)}>
                            {getStatusIcon(survey.status)}
                            <span className="ml-1">{getStatusText(survey.status)}</span>
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">{survey.responseCount}</div>
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">{survey.viewCount}</div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">{survey.updatedAt}</div>
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem asChild>
                                <Link href={`/survey/take/${survey.id}`}>
                                  <Eye className="h-4 w-4 mr-2" />
                                  查看
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem asChild>
                                <Link href={`/survey/edit/${survey.id}`}>
                                  <Edit className="h-4 w-4 mr-2" />
                                  编辑
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={handleShareClick}>
                                <Share className="h-4 w-4 mr-2" />
                                分享
                              </DropdownMenuItem>
                              <DropdownMenuItem asChild>
                                <Link href={`/survey/analytics/${survey.id}`}>
                                  <BarChart3 className="h-4 w-4 mr-2" />
                                  分析
                                </Link>
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={handleDeleteClick}>
                                <Trash2 className="h-4 w-4 mr-2" />
                                删除
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>

      {/* 删除确认对话框 */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除问卷</DialogTitle>
            <DialogDescription>
              确定要删除问卷 "{selectedSurvey?.title}" 吗？此操作无法撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm}>
              确认删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 分享对话框 */}
      <Dialog open={shareDialogOpen} onOpenChange={setShareDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>分享问卷</DialogTitle>
            <DialogDescription>
              复制以下链接分享问卷
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <Input value={shareUrl} readOnly />
            <Button onClick={handleCopyShareUrl}>
              <Copy className="h-4 w-4 mr-2" />
              复制链接
            </Button>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShareDialogOpen(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}