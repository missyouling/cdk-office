'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
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
  Input 
} from '@/components/ui/input';
import { 
  Label 
} from '@/components/ui/label';
import { 
  Textarea 
} from '@/components/ui/textarea';
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
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';

import { 
  ArrowLeft, 
  Edit, 
  Send, 
  Download, 
  Share, 
  XCircle, 
  CheckCircle, 
  Clock, 
  FileText, 
  User, 
  Building, 
  Shield, 
  History, 
  Eye, 
  File,
  Link,
  MapPin,
  Calendar,
  Hash,
  Server
} from 'lucide-react';

interface ContractDetail {
  id: string;
  title: string;
  description: string;
  content: string;
  status: string;
  progress: number;
  createdAt: string;
  startTime: string;
  expireTime: string;
  completedAt?: string;
  createdBy: string;
  requireCA: boolean;
  requireBlockchain: boolean;
  originalFileURL: string;
  finalFileURL?: string;
  evidenceURL?: string;
  fileHash: string;
  blockchainTxHash?: string;
  signers: Array<{
    id: string;
    name: string;
    email: string;
    phone: string;
    signerType: 'person' | 'company';
    signOrder: number;
    status: 'pending' | 'signed' | 'rejected';
    signTime?: string;
    signIP?: string;
    signLocation?: string;
    certificateID?: string;
  }>;
  logs: Array<{
    id: string;
    action: string;
    description: string;
    operatorName: string;
    ipAddress: string;
    createdAt: string;
  }>;
}

const STATUS_CONFIG = {
  draft: { label: '草稿', color: 'secondary', icon: <Edit className="h-4 w-4" /> },
  pending: { label: '待发送', color: 'warning', icon: <Send className="h-4 w-4" /> },
  signing: { label: '签署中', color: 'info', icon: <Clock className="h-4 w-4" /> },
  completed: { label: '已完成', color: 'success', icon: <CheckCircle className="h-4 w-4" /> },
  rejected: { label: '已拒绝', color: 'destructive', icon: <XCircle className="h-4 w-4" /> },
  cancelled: { label: '已取消', color: 'secondary', icon: <XCircle className="h-4 w-4" /> },
  expired: { label: '已过期', color: 'destructive', icon: <Clock className="h-4 w-4" /> }
};

export default function ContractDetail() {
  const router = useRouter();
  const params = useParams();
  const contractId = params.id as string;

  const [contract, setContract] = useState<ContractDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [signDialogOpen, setSignDialogOpen] = useState(false);
  const [rejectDialogOpen, setRejectDialogOpen] = useState(false);
  const [rejectReason, setRejectReason] = useState('');
  const [historyOpen, setHistoryOpen] = useState(false);

  // 模拟数据加载
  useEffect(() => {
    setTimeout(() => {
      setContract({
        id: contractId,
        title: '员工劳动合同',
        description: '与张三签署的劳动合同，包含薪资、福利、工作职责等条款',
        content: '这里是合同的详细内容...',
        status: 'signing',
        progress: 50,
        createdAt: '2025-01-15',
        startTime: '2025-01-16',
        expireTime: '2026-01-15',
        createdBy: '人事部-李经理',
        requireCA: true,
        requireBlockchain: true,
        originalFileURL: '/contracts/contract-001-original.pdf',
        fileHash: 'sha256:abcd1234...',
        signers: [
          {
            id: '1',
            name: '张三',
            email: 'zhangsan@example.com',
            phone: '13800138001',
            signerType: 'person',
            signOrder: 1,
            status: 'signed',
            signTime: '2025-01-16 14:30:00',
            signIP: '192.168.1.100',
            signLocation: '北京市朝阳区',
            certificateID: 'CA-001-20250116'
          },
          {
            id: '2',
            name: '公司HR部门',
            email: 'hr@company.com',
            phone: '010-12345678',
            signerType: 'company',
            signOrder: 2,
            status: 'pending'
          }
        ],
        logs: [
          {
            id: '1',
            action: 'create',
            description: '创建合同',
            operatorName: '李经理',
            ipAddress: '192.168.1.50',
            createdAt: '2025-01-15 10:00:00'
          },
          {
            id: '2',
            action: 'send',
            description: '发送合同给签署人',
            operatorName: '李经理',
            ipAddress: '192.168.1.50',
            createdAt: '2025-01-16 09:00:00'
          },
          {
            id: '3',
            action: 'sign',
            description: '张三完成签署',
            operatorName: '张三',
            ipAddress: '192.168.1.100',
            createdAt: '2025-01-16 14:30:00'
          }
        ]
      });
      setLoading(false);
    }, 1000);
  }, [contractId]);

  const handleSign = () => {
    // 实现签署逻辑
    setSignDialogOpen(false);
    toast({
      title: "签署成功",
      description: "合同已成功签署",
    });
  };

  const handleReject = () => {
    // 实现拒绝逻辑
    setRejectDialogOpen(false);
    setRejectReason('');
    toast({
      title: "合同已拒绝",
      description: "合同已被拒绝，拒绝原因已记录",
    });
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

  const getActionIcon = (action: string) => {
    switch (action) {
      case 'create': return <FileText className="h-4 w-4" />;
      case 'send': return <Send className="h-4 w-4" />;
      case 'sign': return <CheckCircle className="h-4 w-4" />;
      case 'reject': return <XCircle className="h-4 w-4" />;
      default: return <History className="h-4 w-4" />;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <span className="ml-2">正在加载合同详情...</span>
      </div>
    );
  }

  if (!contract) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px]">
        <FileText className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="text-lg font-semibold">合同未找到</h3>
        <p className="text-muted-foreground">无法找到指定的合同</p>
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
          <h1 className="text-3xl font-bold">{contract.title}</h1>
          <p className="text-muted-foreground">{contract.description}</p>
        </div>
        <div className="ml-auto flex space-x-2">
          <Button variant="outline" size="icon">
            <Share className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="icon">
            <Download className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="icon">
            <Eye className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* 合同状态卡片 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex justify-between items-start">
              <div>
                <CardTitle>合同状态</CardTitle>
                <CardDescription>当前合同的签署状态和进度</CardDescription>
              </div>
              <Badge variant={getStatusColor(contract.status)}>
                {STATUS_CONFIG[contract.status as keyof typeof STATUS_CONFIG].icon}
                <span className="ml-1">{STATUS_CONFIG[contract.status as keyof typeof STATUS_CONFIG].label}</span>
              </Badge>
            </div>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">签署进度</span>
                <span>{contract.progress}%</span>
              </div>
              <Progress value={contract.progress} className="h-2" />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="flex items-center text-sm">
                  <Calendar className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">创建时间:</span>
                  <span className="ml-2">{contract.createdAt}</span>
                </div>
                <div className="flex items-center text-sm">
                  <Calendar className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">开始时间:</span>
                  <span className="ml-2">{contract.startTime}</span>
                </div>
                <div className="flex items-center text-sm">
                  <Calendar className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">到期时间:</span>
                  <span className="ml-2">{contract.expireTime}</span>
                </div>
              </div>
              <div className="space-y-2">
                <div className="flex items-center text-sm">
                  <User className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">创建人:</span>
                  <span className="ml-2">{contract.createdBy}</span>
                </div>
                <div className="flex items-center text-sm">
                  <Shield className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">数字证书:</span>
                  <span className="ml-2">{contract.requireCA ? '需要' : '不需要'}</span>
                </div>
                <div className="flex items-center text-sm">
                  <Server className="h-4 w-4 mr-2 text-muted-foreground" />
                  <span className="text-muted-foreground">区块链存证:</span>
                  <span className="ml-2">{contract.requireBlockchain ? '需要' : '不需要'}</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 签署人信息卡片 */}
        <Card>
          <CardHeader>
            <CardTitle>签署人信息</CardTitle>
            <CardDescription>合同签署人的状态和信息</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {contract.signers.map((signer) => (
              <div key={signer.id} className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center space-x-3">
                  <Avatar>
                    <AvatarFallback>{signer.name.substring(0, 2)}</AvatarFallback>
                  </Avatar>
                  <div>
                    <div className="font-medium">{signer.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {signer.signerType === 'person' ? (
                        <User className="h-3 w-3 inline mr-1" />
                      ) : (
                        <Building className="h-3 w-3 inline mr-1" />
                      )}
                      {signer.signerType === 'person' ? '个人' : '企业'}
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <Badge variant={getSignerStatusColor(signer.status)}>
                    {getSignerStatusText(signer.status)}
                  </Badge>
                  {signer.signTime && (
                    <div className="text-xs text-muted-foreground mt-1">
                      {signer.signTime}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* 操作按钮 */}
      <div className="flex justify-center space-x-4">
        <Button onClick={() => setSignDialogOpen(true)}>
          <CheckCircle className="h-4 w-4 mr-2" />
          签署合同
        </Button>
        <Button variant="outline" onClick={() => setRejectDialogOpen(true)}>
          <XCircle className="h-4 w-4 mr-2" />
          拒绝合同
        </Button>
        <Sheet open={historyOpen} onOpenChange={setHistoryOpen}>
          <SheetTrigger asChild>
            <Button variant="outline">
              <History className="h-4 w-4 mr-2" />
              操作历史
            </Button>
          </SheetTrigger>
          <SheetContent>
            <SheetHeader>
              <SheetTitle>操作历史</SheetTitle>
              <SheetDescription>
                合同的所有操作记录
              </SheetDescription>
            </SheetHeader>
            <div className="mt-6 space-y-4">
              {contract.logs.map((log) => (
                <div key={log.id} className="flex items-start space-x-3">
                  <div className="mt-1">
                    {getActionIcon(log.action)}
                  </div>
                  <div className="flex-1">
                    <div className="font-medium">{log.operatorName}</div>
                    <div className="text-sm text-muted-foreground">{log.description}</div>
                    <div className="flex items-center text-xs text-muted-foreground mt-1">
                      <MapPin className="h-3 w-3 mr-1" />
                      {log.ipAddress}
                      <span className="mx-2">•</span>
                      {log.createdAt}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </SheetContent>
        </Sheet>
      </div>

      {/* 合同内容 */}
      <Card>
        <CardHeader>
          <CardTitle>合同内容</CardTitle>
          <CardDescription>合同的详细条款和内容</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="prose max-w-none">
            <p>{contract.content}</p>
          </div>
        </CardContent>
      </Card>

      {/* 存证信息 */}
      <Card>
        <CardHeader>
          <CardTitle>存证信息</CardTitle>
          <CardDescription>合同的数字存证和区块链信息</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div className="flex items-center">
              <Hash className="h-4 w-4 mr-2 text-muted-foreground" />
              <span className="text-muted-foreground">文件哈希:</span>
            </div>
            <code className="text-sm">{contract.fileHash}</code>
          </div>
          {contract.blockchainTxHash && (
            <div className="flex items-center justify-between p-3 border rounded-lg">
              <div className="flex items-center">
                <Server className="h-4 w-4 mr-2 text-muted-foreground" />
                <span className="text-muted-foreground">区块链交易:</span>
              </div>
              <code className="text-sm">{contract.blockchainTxHash}</code>
            </div>
          )}
          <div className="flex space-x-2">
            <Button variant="outline" size="sm">
              <File className="h-4 w-4 mr-2" />
              查看原始文件
            </Button>
            {contract.finalFileURL && (
              <Button variant="outline" size="sm">
                <File className="h-4 w-4 mr-2" />
                查看最终文件
              </Button>
            )}
            {contract.evidenceURL && (
              <Button variant="outline" size="sm">
                <File className="h-4 w-4 mr-2" />
                查看存证证明
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 签署对话框 */}
      <Dialog open={signDialogOpen} onOpenChange={setSignDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>签署合同</DialogTitle>
            <DialogDescription>
              确认签署合同 "{contract.title}" 吗？
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="flex items-center space-x-2 rounded-lg border p-4">
              <FileText className="h-5 w-5 text-muted-foreground" />
              <div>
                <div className="font-medium">{contract.title}</div>
                <div className="text-sm text-muted-foreground">请确认合同内容无误后签署</div>
              </div>
            </div>
            <div className="text-sm text-muted-foreground">
              签署后将无法修改合同内容，请仔细核对。
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSignDialogOpen(false)}>
              取消
            </Button>
            <Button onClick={handleSign}>
              确认签署
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 拒绝对话框 */}
      <Dialog open={rejectDialogOpen} onOpenChange={setRejectDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>拒绝合同</DialogTitle>
            <DialogDescription>
              请填写拒绝签署合同 "{contract.title}" 的原因
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="rejectReason">拒绝原因</Label>
              <Textarea
                id="rejectReason"
                placeholder="请输入拒绝原因..."
                value={rejectReason}
                onChange={(e) => setRejectReason(e.target.value)}
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRejectDialogOpen(false)}>
              取消
            </Button>
            <Button variant="destructive" onClick={handleReject}>
              确认拒绝
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}