/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

'use client';

import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn } from '@/lib/utils';
import {
  Search,
  Plus,
  Filter,
  Share2,
  Trash2,
  Edit,
  Eye,
  Lock,
  Users,
  Globe,
  MessageSquare,
  ScanLine,
  Upload,
  Calendar,
  Tag,
  FolderOpen,
  MoreHorizontal,
  FileText,
  Zap,
  TrendingUp,
  Database,
  Download,
  Heart,
  BookmarkPlus,
} from 'lucide-react';

import { 
  PersonalKnowledge, 
  WeChatRecord, 
  KnowledgeStatistics,
  CreateKnowledgeRequest,
  KnowledgeFilter,
} from '@/types/knowledge';
import { knowledgeService } from '@/services/knowledge';

interface StatCardProps {
  title: string;
  value: number;
  icon: React.ReactNode;
  trend?: number;
  color?: string;
}

function StatCard({ title, value, icon, trend, color = "blue" }: StatCardProps) {
  const colorClasses = {
    blue: "bg-blue-50 border-blue-200",
    green: "bg-green-50 border-green-200", 
    purple: "bg-purple-50 border-purple-200",
    orange: "bg-orange-50 border-orange-200",
  };

  const iconColors = {
    blue: "text-blue-600",
    green: "text-green-600",
    purple: "text-purple-600", 
    orange: "text-orange-600",
  };

  return (
    <Card className={cn("", colorClasses[color as keyof typeof colorClasses])}>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            <p className="text-2xl font-bold">{value}</p>
            {trend && (
              <p className={cn("text-xs flex items-center mt-1", 
                trend > 0 ? "text-green-600" : "text-red-600"
              )}>
                <TrendingUp className="h-3 w-3 mr-1" />
                {trend > 0 ? '+' : ''}{trend}%
              </p>
            )}
          </div>
          <div className={cn("h-8 w-8", iconColors[color as keyof typeof iconColors])}>
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

interface KnowledgeCardProps {
  knowledge: PersonalKnowledge;
  onEdit: (knowledge: PersonalKnowledge) => void;
  onDelete: (id: string) => void;
  onShare: (knowledge: PersonalKnowledge) => void;
  isSelected: boolean;
  onSelect: (id: string, selected: boolean) => void;
}

function KnowledgeCard({ 
  knowledge, 
  onEdit, 
  onDelete, 
  onShare, 
  isSelected, 
  onSelect 
}: KnowledgeCardProps) {
  const getPrivacyIcon = (privacy: string) => {
    switch (privacy) {
      case 'private':
        return <Lock className="h-4 w-4" />;
      case 'shared':
        return <Users className="h-4 w-4" />;
      case 'public':
        return <Globe className="h-4 w-4" />;
      default:
        return <Lock className="h-4 w-4" />;
    }
  };

  const getSourceBadge = (sourceType: string) => {
    const variants = {
      manual: { variant: "default" as const, label: "手动创建", icon: <Edit className="h-3 w-3 mr-1" /> },
      wechat: { variant: "secondary" as const, label: "微信", icon: <MessageSquare className="h-3 w-3 mr-1" /> },
      upload: { variant: "outline" as const, label: "上传", icon: <Upload className="h-3 w-3 mr-1" /> },
      scan: { variant: "secondary" as const, label: "扫描", icon: <ScanLine className="h-3 w-3 mr-1" /> },
    };
    
    const config = variants[sourceType as keyof typeof variants] || variants.manual;
    return (
      <Badge variant={config.variant} className="text-xs">
        {config.icon}
        {config.label}
      </Badge>
    );
  };

  return (
    <Card className={cn("transition-all hover:shadow-md", isSelected && "ring-2 ring-blue-500")}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={isSelected}
              onChange={(e) => onSelect(knowledge.id, e.target.checked)}
              className="rounded border-gray-300"
            />
            <div className="flex items-center space-x-2">
              {getPrivacyIcon(knowledge.privacy)}
              <CardTitle className="text-lg line-clamp-1">{knowledge.title}</CardTitle>
            </div>
          </div>
          <Button variant="ghost" size="sm">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </div>
        {knowledge.description && (
          <CardDescription className="line-clamp-2">
            {knowledge.description}
          </CardDescription>
        )}
      </CardHeader>
      <CardContent className="pt-0">
        <div className="space-y-3">
          <div className="flex flex-wrap gap-1">
            {knowledge.tags.slice(0, 3).map((tag, index) => (
              <Badge key={index} variant="outline" className="text-xs">
                <Tag className="h-3 w-3 mr-1" />
                {tag}
              </Badge>
            ))}
            {knowledge.tags.length > 3 && (
              <Badge variant="outline" className="text-xs">
                +{knowledge.tags.length - 3}
              </Badge>
            )}
          </div>
          
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <div className="flex items-center space-x-4">
              {getSourceBadge(knowledge.source_type)}
              {knowledge.category && (
                <div className="flex items-center">
                  <FolderOpen className="h-3 w-3 mr-1" />
                  {knowledge.category}
                </div>
              )}
            </div>
            <div className="flex items-center">
              <Calendar className="h-3 w-3 mr-1" />
              {new Date(knowledge.updated_at).toLocaleDateString()}
            </div>
          </div>
        </div>
      </CardContent>
      <CardFooter className="pt-0">
        <div className="flex justify-between w-full">
          <div className="flex space-x-1">
            <Button variant="ghost" size="sm" onClick={() => onEdit(knowledge)}>
              <Edit className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm">
              <Eye className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="sm">
              <Download className="h-4 w-4" />
            </Button>
          </div>
          <div className="flex space-x-1">
            <Button variant="ghost" size="sm" onClick={() => onShare(knowledge)}>
              <Share2 className="h-4 w-4" />
            </Button>
            <Button 
              variant="ghost" 
              size="sm" 
              onClick={() => onDelete(knowledge.id)}
              className="hover:text-red-600"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardFooter>
    </Card>
  );
}

const PersonalKnowledgeBase: React.FC = () => {
  const [knowledgeList, setKnowledgeList] = useState<PersonalKnowledge[]>([]);
  const [wechatRecords, setWechatRecords] = useState<WeChatRecord[]>([]);
  const [statistics, setStatistics] = useState<KnowledgeStatistics | null>(null);
  const [loading, setLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [filter, setFilter] = useState<KnowledgeFilter>({});
  const [activeTab, setActiveTab] = useState('knowledge');

  // Load data on component mount
  useEffect(() => {
    loadKnowledgeList();
    loadStatistics();
  }, [filter]);

  const loadKnowledgeList = async () => {
    setLoading(true);
    try {
      const response = await knowledgeService.listKnowledge({
        page: 1,
        page_size: 50,
        ...filter,
        keyword: searchTerm,
      });
      setKnowledgeList(response.knowledge);
    } catch (error) {
      console.error('Failed to load knowledge list:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadStatistics = async () => {
    try {
      const stats = await knowledgeService.getStatistics();
      setStatistics(stats);
    } catch (error) {
      console.error('Failed to load statistics:', error);
    }
  };

  const loadWechatRecords = async () => {
    try {
      const records = await knowledgeService.getWechatRecords({ page: 1, page_size: 50 });
      setWechatRecords(records.records);
    } catch (error) {
      console.error('Failed to load wechat records:', error);
    }
  };

  const handleEdit = (knowledge: PersonalKnowledge) => {
    // TODO: 实现编辑功能
    console.log('Edit knowledge:', knowledge);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除这个知识吗？')) return;
    
    try {
      await knowledgeService.deleteKnowledge(id);
      loadKnowledgeList();
    } catch (error) {
      console.error('Failed to delete knowledge:', error);
    }
  };

  const handleShare = (knowledge: PersonalKnowledge) => {
    // TODO: 实现分享功能
    console.log('Share knowledge:', knowledge);
  };

  const handleSelectItem = (id: string, selected: boolean) => {
    if (selected) {
      setSelectedItems([...selectedItems, id]);
    } else {
      setSelectedItems(selectedItems.filter(item => item !== id));
    }
  };

  const handleSelectAll = () => {
    if (selectedItems.length === knowledgeList.length) {
      setSelectedItems([]);
    } else {
      setSelectedItems(knowledgeList.map(item => item.id));
    }
  };

  const filteredKnowledge = knowledgeList.filter(item =>
    item.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
    item.content.toLowerCase().includes(searchTerm.toLowerCase()) ||
    item.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  return (
    <div className="w-full p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col space-y-4 md:flex-row md:items-center md:justify-between md:space-y-0">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">个人知识库</h1>
          <p className="text-muted-foreground">
            管理您的个人知识文档，支持微信聊天记录导入和团队分享
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="搜索知识..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-9 md:w-[300px]"
            />
          </div>
          <Button variant="outline" size="sm">
            <Filter className="h-4 w-4 mr-2" />
            筛选
          </Button>
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            新建知识
          </Button>
        </div>
      </div>

      {/* Statistics Cards */}
      {statistics && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="总知识数量"
            value={statistics.total_knowledge}
            icon={<Database className="h-6 w-6" />}
            trend={12}
            color="blue"
          />
          <StatCard
            title="已分享知识"
            value={statistics.shared_knowledge}
            icon={<Share2 className="h-6 w-6" />}
            trend={8}
            color="green"
          />
          <StatCard
            title="本周新增"
            value={statistics.weekly_added}
            icon={<Zap className="h-6 w-6" />}
            trend={-2}
            color="purple"
          />
          <StatCard
            title="分类数量"
            value={statistics.by_category.length}
            icon={<FolderOpen className="h-6 w-6" />}
            color="orange"
          />
        </div>
      )}

      {/* Main Content */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="knowledge">
            <FileText className="h-4 w-4 mr-2" />
            知识文档
          </TabsTrigger>
          <TabsTrigger value="wechat">
            <MessageSquare className="h-4 w-4 mr-2" />
            微信记录
          </TabsTrigger>
          <TabsTrigger value="analytics">
            <TrendingUp className="h-4 w-4 mr-2" />
            统计分析
          </TabsTrigger>
        </TabsList>

        <TabsContent value="knowledge" className="space-y-4">
          {/* Batch Operations */}
          {selectedItems.length > 0 && (
            <Card className="p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-muted-foreground">
                    已选择 {selectedItems.length} 个项目
                  </span>
                  <Button variant="outline" size="sm" onClick={handleSelectAll}>
                    {selectedItems.length === knowledgeList.length ? '取消全选' : '全选'}
                  </Button>
                </div>
                <div className="flex items-center space-x-2">
                  <Button variant="outline" size="sm">
                    <Share2 className="h-4 w-4 mr-2" />
                    批量分享
                  </Button>
                  <Button variant="outline" size="sm">
                    <Download className="h-4 w-4 mr-2" />
                    批量导出
                  </Button>
                  <Button variant="destructive" size="sm">
                    <Trash2 className="h-4 w-4 mr-2" />
                    批量删除
                  </Button>
                </div>
              </div>
            </Card>
          )}

          {/* Knowledge Grid */}
          {loading ? (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {[...Array(6)].map((_, i) => (
                <Card key={i} className="animate-pulse">
                  <CardHeader className="space-y-2">
                    <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                    <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      <div className="h-3 bg-gray-200 rounded"></div>
                      <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {filteredKnowledge.map((knowledge) => (
                <KnowledgeCard
                  key={knowledge.id}
                  knowledge={knowledge}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  onShare={handleShare}
                  isSelected={selectedItems.includes(knowledge.id)}
                  onSelect={handleSelectItem}
                />
              ))}
            </div>
          )}

          {filteredKnowledge.length === 0 && !loading && (
            <Card className="p-12 text-center">
              <FileText className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold mb-2">暂无知识文档</h3>
              <p className="text-muted-foreground mb-4">
                开始创建您的第一个知识文档，或者导入微信聊天记录
              </p>
              <div className="flex justify-center space-x-2">
                <Button>
                  <Plus className="h-4 w-4 mr-2" />
                  新建知识
                </Button>
                <Button variant="outline">
                  <MessageSquare className="h-4 w-4 mr-2" />
                  导入微信记录
                </Button>
              </div>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="wechat" className="space-y-4">
          <Card className="p-6">
            <CardHeader className="px-0 pt-0">
              <CardTitle>微信聊天记录</CardTitle>
              <CardDescription>
                导入和管理微信聊天记录，自动转换为知识文档
              </CardDescription>
            </CardHeader>
            <CardContent className="px-0">
              <div className="space-y-4">
                <Button onClick={loadWechatRecords}>
                  <Upload className="h-4 w-4 mr-2" />
                  导入微信记录
                </Button>
                
                {wechatRecords.length > 0 ? (
                  <div className="space-y-2">
                    {wechatRecords.map((record) => (
                      <Card key={record.id} className="p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-medium">{record.session_name}</h4>
                            <p className="text-sm text-muted-foreground">
                              {record.sender_name} • {new Date(record.message_time).toLocaleString()}
                            </p>
                          </div>
                          <Badge variant={record.process_status === 'completed' ? 'default' : 'secondary'}>
                            {record.process_status}
                          </Badge>
                        </div>
                        <p className="mt-2 text-sm line-clamp-2">{record.content}</p>
                      </Card>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <MessageSquare className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                    <p className="text-muted-foreground">暂无微信记录</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="analytics" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card className="p-6">
              <CardHeader className="px-0 pt-0">
                <CardTitle>分类统计</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                {statistics?.by_category.map((cat, index) => (
                  <div key={index} className="flex justify-between items-center py-2">
                    <span>{cat.category || '未分类'}</span>
                    <Badge variant="outline">{cat.count}</Badge>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card className="p-6">
              <CardHeader className="px-0 pt-0">
                <CardTitle>来源统计</CardTitle>
              </CardHeader>
              <CardContent className="px-0">
                {statistics?.by_source.map((source, index) => (
                  <div key={index} className="flex justify-between items-center py-2">
                    <span>{source.source_type}</span>
                    <Badge variant="outline">{source.count}</Badge>
                  </div>
                ))}
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default PersonalKnowledgeBase;