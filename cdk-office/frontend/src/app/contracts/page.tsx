'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import {
  useReactTable,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  ColumnDef,
  SortingState,
  ColumnFiltersState,
  VisibilityState,
  flexRender,
} from '@tanstack/react-table';

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
  DropdownMenuCheckboxItem,
} from '@/components/ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Separator } from '@/components/ui/separator';
import { toast } from '@/components/ui/use-toast';

import {
  Plus,
  Search,
  MoreHorizontal,
  Eye,
  Edit,
  Trash2,
  Send,
  Download,
  Filter,
  SortAsc,
  SortDesc,
  ArrowUpDown,
  FileSignature,
  Users,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Calendar,
  TrendingUp,
  FileText,
  Settings,
  Columns
} from 'lucide-react';

import type { Contract, ContractStatus, ContractSigner } from '@/types/contract';

export default function ContractsPage() {
  const router = useRouter();
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [loading, setLoading] = useState(true);
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});
  const [globalFilter, setGlobalFilter] = useState('');
  const [activeTab, setActiveTab] = useState('all');

  // 状态配置
  const statusConfig = {
    draft: { label: '草稿', color: 'secondary', icon: <Edit className="h-4 w-4" /> },
    pending: { label: '待发送', color: 'warning', icon: <Send className="h-4 w-4" /> },
    signing: { label: '签署中', color: 'info', icon: <Clock className="h-4 w-4" /> },
    completed: { label: '已完成', color: 'success', icon: <CheckCircle className="h-4 w-4" /> },
    rejected: { label: '已拒绝', color: 'destructive', icon: <XCircle className="h-4 w-4" /> },
    cancelled: { label: '已取消', color: 'secondary', icon: <XCircle className="h-4 w-4" /> },
    expired: { label: '已过期', color: 'destructive', icon: <AlertCircle className="h-4 w-4" /> }
  };

  // 定义表格列
  const columns = useMemo<ColumnDef<Contract>[]>(() => [
    {
      accessorKey: 'title',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-auto p-0 font-semibold"
        >
          合同名称
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <div className="flex flex-col">
          <span className="font-medium">{row.getValue('title')}</span>
          {row.original.description && (
            <span className="text-sm text-muted-foreground">
              {row.original.description}
            </span>
          )}
        </div>
      ),
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-auto p-0 font-semibold"
        >
          状态
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const status = row.getValue('status') as ContractStatus;
        const config = statusConfig[status];
        return (
          <Badge 
            variant={config.color as any} 
            className="flex items-center gap-1 w-fit"
          >
            {config.icon}
            {config.label}
          </Badge>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      accessorKey: 'progress',
      header: '签署进度',
      cell: ({ row }) => {
        const progress = row.getValue('progress') as number;
        const status = row.original.status;
        return (
          <div className="flex items-center space-x-2 min-w-[120px]">
            <Progress value={progress} className="flex-1" />
            <span className="text-sm text-muted-foreground">{progress}%</span>
          </div>
        );
      },
    },
    {
      accessorKey: 'signers',
      header: '签署方',
      cell: ({ row }) => {
        const signers = row.getValue('signers') as ContractSigner[];
        return (
          <div className="flex -space-x-2">
            {signers.slice(0, 3).map((signer, index) => (
              <Avatar key={signer.id} className="h-8 w-8 border-2 border-background">
                <AvatarImage src={`/avatars/${signer.id}.jpg`} />
                <AvatarFallback className="text-xs">
                  {signer.name.slice(0, 2)}
                </AvatarFallback>
              </Avatar>
            ))}
            {signers.length > 3 && (
              <div className="h-8 w-8 rounded-full border-2 border-background bg-muted flex items-center justify-center text-xs">
                +{signers.length - 3}
              </div>
            )}
          </div>
        );
      },
    },
    {
      accessorKey: 'createdAt',
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          className="h-auto p-0 font-semibold"
        >
          创建时间
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue('createdAt'));
        return <span className="text-sm">{date.toLocaleDateString('zh-CN')}</span>;
      },
    },
    {
      accessorKey: 'expireTime',
      header: '过期时间',
      cell: ({ row }) => {
        const date = new Date(row.getValue('expireTime'));
        const isExpiringSoon = (date.getTime() - Date.now()) < 7 * 24 * 60 * 60 * 1000; // 7天内过期
        return (
          <span className={`text-sm ${isExpiringSoon ? 'text-orange-600 font-medium' : ''}`}>
            {date.toLocaleDateString('zh-CN')}
          </span>
        );
      },
    },
    {
      accessorKey: 'createdByName',
      header: '创建者',
      cell: ({ row }) => (
        <span className="text-sm">{row.getValue('createdByName')}</span>
      ),
    },
    {
      id: 'actions',
      header: '操作',
      cell: ({ row }) => {
        const contract = row.original;
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">打开菜单</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => handleView(contract.id)}>
                <Eye className="mr-2 h-4 w-4" />
                查看详情
              </DropdownMenuItem>
              {contract.status === 'draft' && (
                <DropdownMenuItem onClick={() => handleEdit(contract.id)}>
                  <Edit className="mr-2 h-4 w-4" />
                  编辑
                </DropdownMenuItem>
              )}
              {(contract.status === 'draft' || contract.status === 'pending') && (
                <DropdownMenuItem onClick={() => handleSend(contract.id)}>
                  <Send className="mr-2 h-4 w-4" />
                  发送签署
                </DropdownMenuItem>
              )}
              {contract.status === 'signing' && canUserSign(contract) && (
                <DropdownMenuItem onClick={() => handleSign(contract.id)}>
                  <FileSignature className="mr-2 h-4 w-4" />
                  立即签署
                </DropdownMenuItem>
              )}
              <DropdownMenuItem onClick={() => handleDownload(contract.id)}>
                <Download className="mr-2 h-4 w-4" />
                下载合同
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              {(contract.status === 'draft' || contract.status === 'pending') && (
                <DropdownMenuItem 
                  onClick={() => handleDelete(contract.id)}
                  className="text-red-600"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  删除
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ], []);

  // 创建表格实例
  const table = useReactTable({
    data: contracts,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: 'includesString',
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
      globalFilter,
    },
  });

  // 模拟数据加载
  useEffect(() => {
    const loadContracts = async () => {
      // 模拟API调用
      setTimeout(() => {
        const mockContracts: Contract[] = [
          {
            id: '1',
            title: '员工劳动合同',
            description: '与张三签署的劳动合同',
            status: 'completed',
            progress: 100,
            createdAt: '2025-01-15T08:00:00Z',
            updatedAt: '2025-01-17T10:30:00Z',
            expireTime: '2026-01-15T23:59:59Z',
            completedAt: '2025-01-17T10:30:00Z',
            createdBy: 'admin',
            createdByName: '管理员',
            teamId: 'team-1',
            currentSignerIndex: 2,
            signers: [
              { id: '1', name: '张三', signerType: 'person', status: 'signed', signTime: '2025-01-16T09:15:00Z' },
              { id: '2', name: '公司HR', signerType: 'company', status: 'signed', signTime: '2025-01-17T10:30:00Z' }
            ]
          },
          {
            id: '2',
            title: '供应商合作协议',
            description: '与ABC公司的供应商合作协议',
            status: 'signing',
            progress: 50,
            createdAt: '2025-01-20T14:00:00Z',
            updatedAt: '2025-01-21T09:00:00Z',
            expireTime: '2025-02-20T23:59:59Z',
            createdBy: 'procurement',
            createdByName: '采购部',
            teamId: 'team-1',
            currentSignerIndex: 1,
            signers: [
              { id: '3', name: '采购经理', signerType: 'person', status: 'signed', signTime: '2025-01-21T09:00:00Z' },
              { id: '4', name: 'ABC公司', signerType: 'company', status: 'pending' }
            ]
          },
          {
            id: '3',
            title: '保密协议',
            description: '技术部门保密协议',
            status: 'draft',
            progress: 0,
            createdAt: '2025-01-22T16:30:00Z',
            updatedAt: '2025-01-22T16:30:00Z',
            expireTime: '2025-03-22T23:59:59Z',
            createdBy: 'tech',
            createdByName: '技术部',
            teamId: 'team-1',
            currentSignerIndex: 0,
            signers: [
              { id: '5', name: '技术总监', signerType: 'person', status: 'pending' },
              { id: '6', name: '员工代表', signerType: 'person', status: 'pending' }
            ]
          },
          {
            id: '4',
            title: '租赁合同',
            description: '办公室租赁合同',
            status: 'pending',
            progress: 0,
            createdAt: '2025-01-23T10:00:00Z',
            updatedAt: '2025-01-23T10:00:00Z',
            expireTime: '2025-02-23T23:59:59Z',
            createdBy: 'admin',
            createdByName: '行政部',
            teamId: 'team-1',
            currentSignerIndex: 0,
            signers: [
              { id: '7', name: '行政经理', signerType: 'person', status: 'pending' },
              { id: '8', name: '物业公司', signerType: 'company', status: 'pending' }
            ]
          }
        ];
        setContracts(mockContracts);
        setLoading(false);
      }, 1000);
    };

    loadContracts();
  }, []);

  // 操作处理函数
  const handleView = (contractId: string) => {
    router.push(`/contracts/${contractId}`);
  };

  const handleEdit = (contractId: string) => {
    router.push(`/contracts/${contractId}/edit`);
  };

  const handleSend = (contractId: string) => {
    toast({
      title: "发送成功",
      description: "合同已发送给签署方",
    });
  };

  const handleSign = (contractId: string) => {
    router.push(`/contracts/${contractId}/sign`);
  };

  const handleDownload = (contractId: string) => {
    toast({
      title: "下载开始",
      description: "合同文件正在下载...",
    });
  };

  const handleDelete = (contractId: string) => {
    setContracts(prev => prev.filter(c => c.id !== contractId));
    toast({
      title: "删除成功",
      description: "合同已删除",
    });
  };

  const canUserSign = (contract: Contract): boolean => {
    // 简化的逻辑：检查当前用户是否为当前应签署人
    return contract.status === 'signing' && contract.currentSignerIndex < contract.signers.length;
  };

  // 获取统计数据
  const getStatusCounts = () => {
    return {
      all: contracts.length,
      draft: contracts.filter(c => c.status === 'draft').length,
      pending: contracts.filter(c => c.status === 'pending').length,
      signing: contracts.filter(c => c.status === 'signing').length,
      completed: contracts.filter(c => c.status === 'completed').length,
      rejected: contracts.filter(c => c.status === 'rejected').length,
    };
  };

  const statusCounts = getStatusCounts();

  // 筛选数据
  const getFilteredData = () => {
    if (activeTab === 'all') return contracts;
    return contracts.filter(contract => contract.status === activeTab);
  };

  const filteredData = getFilteredData();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">电子合同管理</h1>
          <p className="text-gray-600 mt-2">
            管理和跟踪所有电子合同的签署状态
          </p>
        </div>
        <Button onClick={() => router.push('/contracts/create')} className="flex items-center gap-2">
          <Plus className="h-4 w-4" />
          新建合同
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-6 gap-4 mb-8">
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-blue-600">{statusCounts.all}</div>
            <div className="text-sm text-gray-600">总计</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-gray-600">{statusCounts.draft}</div>
            <div className="text-sm text-gray-600">草稿</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-orange-600">{statusCounts.pending}</div>
            <div className="text-sm text-gray-600">待发送</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-blue-600">{statusCounts.signing}</div>
            <div className="text-sm text-gray-600">签署中</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-green-600">{statusCounts.completed}</div>
            <div className="text-sm text-gray-600">已完成</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-red-600">{statusCounts.rejected}</div>
            <div className="text-sm text-gray-600">已拒绝</div>
          </CardContent>
        </Card>
      </div>

      {/* 搜索和筛选 */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-4 flex-1">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="搜索合同..."
              value={globalFilter ?? ''}
              onChange={(e) => setGlobalFilter(e.target.value)}
              className="pl-10"
            />
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="flex items-center gap-2">
                <Filter className="h-4 w-4" />
                筛选
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-48">
              <div className="p-2">
                <p className="text-sm font-medium mb-2">状态筛选</p>
                {Object.entries(statusConfig).map(([status, config]) => (
                  <DropdownMenuCheckboxItem
                    key={status}
                    checked={!table.getColumn('status')?.getFilterValue() || 
                            (table.getColumn('status')?.getFilterValue() as string[])?.includes(status)}
                    onCheckedChange={(checked) => {
                      const currentFilter = table.getColumn('status')?.getFilterValue() as string[] || [];
                      if (checked) {
                        table.getColumn('status')?.setFilterValue([...currentFilter, status]);
                      } else {
                        table.getColumn('status')?.setFilterValue(
                          currentFilter.filter(s => s !== status)
                        );
                      }
                    }}
                  >
                    <Badge variant={config.color as any} className="mr-2">
                      {config.label}
                    </Badge>
                  </DropdownMenuCheckboxItem>
                ))}
              </div>
            </DropdownMenuContent>
          </DropdownMenu>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="flex items-center gap-2">
                <Columns className="h-4 w-4" />
                列显示
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {table
                .getAllColumns()
                .filter((column) => column.getCanHide())
                .map((column) => {
                  return (
                    <DropdownMenuCheckboxItem
                      key={column.id}
                      className="capitalize"
                      checked={column.getIsVisible()}
                      onCheckedChange={(value) =>
                        column.toggleVisibility(!!value)
                      }
                    >
                      {column.id}
                    </DropdownMenuCheckboxItem>
                  )
                })}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* 状态标签页 */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="mb-6">
        <TabsList>
          <TabsTrigger value="all">全部 ({statusCounts.all})</TabsTrigger>
          <TabsTrigger value="draft">草稿 ({statusCounts.draft})</TabsTrigger>
          <TabsTrigger value="pending">待发送 ({statusCounts.pending})</TabsTrigger>
          <TabsTrigger value="signing">签署中 ({statusCounts.signing})</TabsTrigger>
          <TabsTrigger value="completed">已完成 ({statusCounts.completed})</TabsTrigger>
          <TabsTrigger value="rejected">已拒绝 ({statusCounts.rejected})</TabsTrigger>
        </TabsList>
      </Tabs>

      {/* 数据表格 */}
      <Card>
        <CardContent className="p-0">
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      data-state={row.getIsSelected() && "selected"}
                      className="cursor-pointer hover:bg-gray-50"
                      onClick={() => handleView(row.original.id)}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id} onClick={(e) => {
                          if (cell.column.id === 'actions') {
                            e.stopPropagation();
                          }
                        }}>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={columns.length} className="h-24 text-center">
                      暂无数据
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>

          {/* 分页控件 */}
          <div className="flex items-center justify-between px-4 py-4">
            <div className="text-sm text-gray-600">
              共 {table.getFilteredRowModel().rows.length} 条记录
            </div>
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
              >
                上一页
              </Button>
              <div className="text-sm text-gray-600">
                第 {table.getState().pagination.pageIndex + 1} 页，共{' '}
                {table.getPageCount()} 页
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
              >
                下一页
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}