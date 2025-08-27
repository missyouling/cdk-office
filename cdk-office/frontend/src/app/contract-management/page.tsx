'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Card, 
  CardContent, 
  CardHeader, 
  CardTitle 
} from '@/components/ui/card';
import { 
  Button 
} from '@/components/ui/button';
import { 
  Badge 
} from '@/components/ui/badge';
import { 
  Progress 
} from '@/components/ui/progress';
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
  Tooltip, 
  TooltipContent, 
  TooltipProvider, 
  TooltipTrigger 
} from '@/components/ui/tooltip';
import { 
  Avatar, 
  AvatarFallback 
} from '@/components/ui/avatar';
import { 
  toast 
} from '@/components/ui/use-toast';

import { 
  Plus, 
  MoreHorizontal, 
  Edit, 
  Trash2, 
  Send, 
  Download, 
  Eye, 
  Clock, 
  CheckCircle, 
  XCircle, 
  Calendar, 
  TrendingUp, 
  Users, 
  FileText,
  FileSignature
} from 'lucide-react';

interface Contract {
  id: string;
  title: string;
  description: string;
  status: 'draft' | 'pending' | 'signing' | 'completed' | 'rejected' | 'cancelled' | 'expired';
  progress: number;
  createdAt: string;
  expireTime: string;
  createdBy: string;
  signers: Array<{
    id: string;
    name: string;
    signerType: 'person' | 'company';
    status: 'pending' | 'signed' | 'rejected';
    signTime?: string;
  }>;
}

const CONTRACT_STATUS_CONFIG = {
  draft: { label: '草稿', color: 'secondary', icon: <Edit className="h-4 w-4" /> },
  pending: { label: '待发送', color: 'warning', icon: <Send className="h-4 w-4" /> },
  signing: { label: '签署中', color: 'info', icon: <Clock className="h-4 w-4" /> },
  completed: { label: '已完成', color: 'success', icon: <CheckCircle className="h-4 w-4" /> },
  rejected: { label: '已拒绝', color: 'destructive', icon: <XCircle className="h-4 w-4" /> },
  cancelled: { label: '已取消', color: 'secondary', icon: <XCircle className="h-4 w-4" /> },
  expired: { label: '已过期', color: 'destructive', icon: <Clock className="h-4 w-4" /> }
};

export default function ContractManagement() {
  const router = useRouter();
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedTab, setSelectedTab] = useState('all');
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedContract, setSelectedContract] = useState<Contract | null>(null);

  // 模拟数据
  useEffect(() => {
    setTimeout(() => {
      setContracts([
        {
          id: '1',
          title: '员工劳动合同',
          description: '与张三签署的劳动合同',
          status: 'completed',
          progress: 100,
          createdAt: '2025-01-15',
          expireTime: '2026-01-15',
          createdBy: '管理员',
          signers: [
            { id: '1', name: '张三', signerType: 'person', status: 'signed', signTime: '2025-01-16' },
            { id: '2', name: '公司HR', signerType: 'company', status: 'signed', signTime: '2025-01-17' }
          ]
        },
        {
          id: '2',
          title: '供应商合作协议',
          description: '与ABC公司的供应商合作协议',
          status: 'signing',
          progress: 50,
          createdAt: '2025-01-20',
          expireTime: '2025-02-20',
          createdBy: '采购部',
          signers: [
            { id: '3', name: '采购经理', signerType: 'person', status: 'signed', signTime: '2025-01-21' },
            { id: '4', name: 'ABC公司', signerType: 'company', status: 'pending' }
          ]
        },
        {
          id: '3',
          title: '保密协议',
          description: '技术部门保密协议',
          status: 'draft',
          progress: 0,
          createdAt: '2025-01-22',
          expireTime: '2025-03-22',
          createdBy: '技术部',
          signers: [
            { id: '5', name: '技术总监', signerType: 'person', status: 'pending' },
            { id: '6', name: '员工代表', signerType: 'person', status: 'pending' }
          ]
        }
      ]);
      setLoading(false);
    }, 1000);
  }, []);

  const tabLabels = [
    { id: 'all', label: '全部' },
    { id: 'draft', label: '草稿' },
    { id: 'signing', label: '签署中' },
    { id: 'completed', label: '已完成' },
    { id: 'rejected', label: '已拒绝' }
  ];

  const filteredContracts = contracts.filter(contract => {
    if (selectedTab === 'all') return true;
    return contract.status === selectedTab;
  });

  const getStatusCounts = () => {
    return {
      all: contracts.length,
      draft: contracts.filter(c => c.status === 'draft').length,
      signing: contracts.filter(c => c.status === 'signing').length,
      completed: contracts.filter(c => c.status === 'completed').length,
      rejected: contracts.filter(c => c.status === 'rejected').length
    };
  };

  const statusCounts = getStatusCounts();

  const handleCreateContract = () => {
    router.push('/contract-management/create');
  };

  const handleContractAction = (action: string, contract: Contract) => {
    switch (action) {
      case 'view':
        router.push(`/contract-management/${contract.id}`);
        break;
      case 'edit':
        router.push(`/contract-management/${contract.id}/edit`);
        break;
      case 'delete':
        // 实现删除逻辑
        toast({
          title: "删除成功",
          description: `合同 "${contract.title}" 已删除`,
        });
        break;
      case 'send':
        // 实现发送逻辑
        toast({
          title: "发送成功",
          description: `合同 "${contract.title}" 已发送`,
        });
        break;
      case 'download':
        // 实现下载逻辑
        toast({
          title: "下载成功",
          description: `合同 "${contract.title}" 已开始下载`,
        });
        break;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'draft': return 'secondary';
      case 'pending': return 'warning';
      case 'signing': return 'info';
      case 'completed': return 'success';
      case 'rejected': return 'destructive';
      case 'cancelled': return 'secondary';
      case 'expired': return 'destructive';
      default: return 'secondary';
    }
  };

  const getSignerStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'warning';
      case 'signed': return 'success';
      case 'rejected': return 'destructive';
      default: return 'secondary';
    }
  };

  const getSignerStatusText = (status: string) => {
    switch (status) {
      case 'pending': return '待签署';
      case 'signed': return '已签署';
      case 'rejected': return '已拒绝';
      default: return status;
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] space-y-4">
        <Progress className="w-32" />
        <span>正在加载合同数据...</span>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">电子合同管理</h1>
          <p className="text-muted-foreground mt-2">
            管理和跟踪电子合同的签署状态
          </p>
        </div>
        <Button onClick={handleCreateContract}>
          <Plus className="h-4 w-4 mr-2" />
          创建合同
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
                <p className="text-sm text-muted-foreground">总合同数</p>
                <p className="text-2xl font-bold">{statusCounts.all}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-yellow-100 p-2 mr-3">
                <Clock className="h-5 w-5 text-yellow-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">签署中</p>
                <p className="text-2xl font-bold">{statusCounts.signing}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-green-100 p-2 mr-3">
                <CheckCircle className="h-5 w-5 text-green-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">已完成</p>
                <p className="text-2xl font-bold">{statusCounts.completed}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center">
              <div className="rounded-full bg-gray-100 p-2 mr-3">
                <Edit className="h-5 w-5 text-gray-500" />
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
              <div className="rounded-full bg-red-100 p-2 mr-3">
                <XCircle className="h-5 w-5 text-red-500" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">已拒绝</p>
                <p className="text-2xl font-bold">{statusCounts.rejected}</p>
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
          {filteredContracts.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <FileSignature className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
              <p>暂无合同数据</p>
              <Button className="mt-4" onClick={handleCreateContract}>
                <Plus className="h-4 w-4 mr-2" />
                创建第一个合同
              </Button>
            </div>
          ) : (
            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>合同标题</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>进度</TableHead>
                      <TableHead>签署人</TableHead>
                      <TableHead>创建时间</TableHead>
                      <TableHead>操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredContracts.map((contract) => (
                      <TableRow key={contract.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{contract.title}</div>
                            <div className="text-sm text-muted-foreground">{contract.description}</div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={getStatusColor(contract.status)}>
                            {CONTRACT_STATUS_CONFIG[contract.status].icon}
                            <span className="ml-1">{CONTRACT_STATUS_CONFIG[contract.status].label}</span>
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="space-y-1">
                            <Progress value={contract.progress} className="h-2" />
                            <div className="text-xs text-muted-foreground">{contract.progress}%</div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex -space-x-2">
                            {contract.signers.map((signer, index) => (
                              <TooltipProvider key={signer.id}>
                                <Tooltip>
                                  <TooltipTrigger>
                                    <Avatar className="border-2 border-background">
                                      <AvatarFallback>{signer.name.substring(0, 2)}</AvatarFallback>
                                    </Avatar>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <div className="flex items-center">
                                      <span className="mr-2">{signer.name}</span>
                                      <Badge variant={getSignerStatusColor(signer.status)} className="text-xs">
                                        {getSignerStatusText(signer.status)}
                                      </Badge>
                                    </div>
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">{contract.createdAt}</div>
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon">
                                <MoreHorizontal className="h-4 w-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuItem onClick={() => handleContractAction('view', contract)}>
                                <Eye className="h-4 w-4 mr-2" />
                                查看
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleContractAction('edit', contract)}>
                                <Edit className="h-4 w-4 mr-2" />
                                编辑
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleContractAction('send', contract)}>
                                <Send className="h-4 w-4 mr-2" />
                                发送
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleContractAction('download', contract)}>
                                <Download className="h-4 w-4 mr-2" />
                                下载
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleContractAction('delete', contract)}>
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
    </div>
  );
}