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

import React, { useState, useEffect, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn } from '@/lib/utils';
import {
  Upload,
  Download,
  FileText,
  Merge,
  Split,
  Compress,
  Shield,
  Type,
  Image,
  Scissors,
  RotateCw,
  Settings,
  History,
  Play,
  Pause,
  Check,
  X,
  AlertCircle,
  Info,
  Loader2,
  Search,
  Filter,
  Trash2,
  Eye,
  Copy,
  Share2,
  MoreHorizontal,
  File,
  FolderOpen,
  Calendar,
  Clock,
  Star,
  Bookmark,
  Zap,
  Target,
  Layers,
} from 'lucide-react';

import {
  PDFOperationResult,
  PDFToolCategory,
  PDFOperation,
  PDFProcessingHistory,
  PDFMergeRequest,
  PDFSplitRequest,
  PDFCompressRequest,
  PDFWatermarkRequest,
  PDFProtectRequest,
  PDFOCRRequest,
  PDFConvertRequest,
  PDFMetadata,
} from '@/types/pdf';
import { pdfService } from '@/services/pdf';

interface FileUploadAreaProps {
  onFilesSelected: (files: File[]) => void;
  accept?: string;
  multiple?: boolean;
  maxSize?: number;
  className?: string;
}

function FileUploadArea({ 
  onFilesSelected, 
  accept = '.pdf', 
  multiple = true, 
  maxSize = 100, 
  className 
}: FileUploadAreaProps) {
  const [dragActive, setDragActive] = useState(false);

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const files = Array.from(e.dataTransfer.files);
    if (files && files.length > 0) {
      onFilesSelected(files);
    }
  }, [onFilesSelected]);

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      onFilesSelected(files);
    }
  }, [onFilesSelected]);

  return (
    <div
      className={cn(
        "relative border-2 border-dashed border-gray-300 rounded-lg p-8 text-center transition-colors",
        dragActive && "border-blue-500 bg-blue-50",
        className
      )}
      onDragEnter={handleDrag}
      onDragLeave={handleDrag}
      onDragOver={handleDrag}
      onDrop={handleDrop}
    >
      <input
        type="file"
        accept={accept}
        multiple={multiple}
        onChange={handleFileSelect}
        className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
      />
      <Upload className="mx-auto h-12 w-12 text-gray-400 mb-4" />
      <h3 className="text-lg font-medium text-gray-900 mb-2">
        {multiple ? '上传PDF文件' : '上传PDF文件'}
      </h3>
      <p className="text-sm text-gray-500 mb-2">
        拖拽文件到此处或点击选择文件
      </p>
      <p className="text-xs text-gray-400">
        支持 PDF 格式，最大 {maxSize}MB
      </p>
    </div>
  );
}

interface OperationCardProps {
  operation: PDFOperation;
  onSelect: (operation: PDFOperation) => void;
  isSelected: boolean;
}

function OperationCard({ operation, onSelect, isSelected }: OperationCardProps) {
  const getOperationIcon = (iconName: string) => {
    const icons: Record<string, React.ReactNode> = {
      merge: <Merge className="h-6 w-6" />,
      split: <Split className="h-6 w-6" />,
      compress: <Compress className="h-6 w-6" />,
      watermark: <Type className="h-6 w-6" />,
      protect: <Shield className="h-6 w-6" />,
      ocr: <Type className="h-6 w-6" />,
      convert: <FileText className="h-6 w-6" />,
      rotate: <RotateCw className="h-6 w-6" />,
      crop: <Scissors className="h-6 w-6" />,
      image: <Image className="h-6 w-6" />,
    };
    return icons[iconName] || <FileText className="h-6 w-6" />;
  };

  return (
    <Card 
      className={cn(
        "cursor-pointer transition-all hover:shadow-md",
        isSelected && "ring-2 ring-blue-500 bg-blue-50"
      )}
      onClick={() => onSelect(operation)}
    >
      <CardHeader className="pb-3">
        <div className="flex items-center space-x-3">
          <div className="text-blue-600">
            {getOperationIcon(operation.icon)}
          </div>
          <div>
            <CardTitle className="text-base">{operation.name}</CardTitle>
            <CardDescription className="text-sm line-clamp-2">
              {operation.description}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="flex flex-wrap gap-1">
          <Badge variant="outline" className="text-xs">
            {operation.input_types.join(', ')}
          </Badge>
          <Badge variant="secondary" className="text-xs">
            → {operation.output_types.join(', ')}
          </Badge>
        </div>
      </CardContent>
    </Card>
  );
}

interface ProcessingStatusProps {
  results: PDFOperationResult[];
  onDownload: (result: PDFOperationResult) => void;
  onRetry: (result: PDFOperationResult) => void;
}

function ProcessingStatus({ results, onDownload, onRetry }: ProcessingStatusProps) {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="h-4 w-4 text-gray-500" />;
      case 'processing':
        return <Loader2 className="h-4 w-4 animate-spin text-blue-500" />;
      case 'completed':
        return <Check className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <X className="h-4 w-4 text-red-500" />;
      default:
        return <Info className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-gray-100 text-gray-700';
      case 'processing':
        return 'bg-blue-100 text-blue-700';
      case 'completed':
        return 'bg-green-100 text-green-700';
      case 'failed':
        return 'bg-red-100 text-red-700';
      default:
        return 'bg-gray-100 text-gray-700';
    }
  };

  if (results.length === 0) {
    return (
      <Card className="p-8 text-center">
        <Target className="h-12 w-12 mx-auto text-gray-400 mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">暂无处理任务</h3>
        <p className="text-gray-500">选择PDF操作并上传文件开始处理</p>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {results.map((result) => (
        <Card key={result.operation_id}>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                {getStatusIcon(result.status)}
                <div>
                  <h4 className="font-medium">{result.operation_type}</h4>
                  <p className="text-sm text-gray-500">
                    {result.input_files.join(', ')}
                  </p>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <Badge className={cn("text-xs", getStatusColor(result.status))}>
                  {result.status}
                </Badge>
                {result.status === 'completed' && (
                  <Button
                    size="sm"
                    onClick={() => onDownload(result)}
                  >
                    <Download className="h-4 w-4 mr-1" />
                    下载
                  </Button>
                )}
                {result.status === 'failed' && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onRetry(result)}
                  >
                    重试
                  </Button>
                )}
              </div>
            </div>
            {result.error_message && (
              <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700">
                <AlertCircle className="h-4 w-4 inline mr-1" />
                {result.error_message}
              </div>
            )}
            {result.processing_time && (
              <div className="mt-2 text-xs text-gray-500">
                处理时间: {result.processing_time}ms
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

const PDFProcessor: React.FC = () => {
  const [activeTab, setActiveTab] = useState('tools');
  const [categories, setCategories] = useState<PDFToolCategory[]>([]);
  const [selectedOperation, setSelectedOperation] = useState<PDFOperation | null>(null);
  const [uploadedFiles, setUploadedFiles] = useState<File[]>([]);
  const [processingResults, setProcessingResults] = useState<PDFOperationResult[]>([]);
  const [processingHistory, setProcessingHistory] = useState<PDFProcessingHistory[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // 加载工具分类
  useEffect(() => {
    loadCategories();
    loadHistory();
  }, []);

  const loadCategories = async () => {
    try {
      const data = await pdfService.getToolCategories();
      setCategories(data);
    } catch (error) {
      console.error('Failed to load categories:', error);
    }
  };

  const loadHistory = async () => {
    try {
      const response = await pdfService.getProcessingHistory(1, 10);
      setProcessingHistory(response.history);
    } catch (error) {
      console.error('Failed to load history:', error);
    }
  };

  const handleFilesSelected = (files: File[]) => {
    setUploadedFiles(files);
  };

  const handleOperationSelect = (operation: PDFOperation) => {
    setSelectedOperation(operation);
  };

  const handleProcessFiles = async () => {
    if (!selectedOperation || uploadedFiles.length === 0) {
      return;
    }

    setLoading(true);
    try {
      let result: PDFOperationResult;

      switch (selectedOperation.id) {
        case 'merge':
          result = await pdfService.mergePDFs({
            files: uploadedFiles,
            output_name: 'merged_document.pdf',
          });
          break;
        case 'split':
          result = await pdfService.splitPDF({
            file: uploadedFiles[0],
            split_type: 'pages',
            pages: [1, 2, 3], // 这里应该从用户输入获取
          });
          break;
        case 'compress':
          result = await pdfService.compressPDF({
            file: uploadedFiles[0],
            quality: 'medium',
            optimize_images: true,
            remove_metadata: false,
          });
          break;
        default:
          result = await pdfService.performOperation({
            operation_type: selectedOperation.id,
            file_data: uploadedFiles,
            options: {},
          });
      }

      setProcessingResults([result, ...processingResults]);
      setUploadedFiles([]);
      setSelectedOperation(null);
    } catch (error) {
      console.error('Processing failed:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async (result: PDFOperationResult) => {
    try {
      const blob = await pdfService.downloadResult(result.operation_id);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = result.output_file || `processed_${result.operation_id}.pdf`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const handleRetry = async (result: PDFOperationResult) => {
    // TODO: 实现重试逻辑
    console.log('Retry operation:', result);
  };

  const filteredOperations = categories.flatMap(category => 
    category.operations.filter(operation =>
      operation.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      operation.description.toLowerCase().includes(searchTerm.toLowerCase())
    )
  );

  return (
    <div className="w-full p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col space-y-4 md:flex-row md:items-center md:justify-between md:space-y-0">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">PDF处理工具</h1>
          <p className="text-muted-foreground">
            强大的PDF文档处理工具，支持50多种操作功能
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline">
            <History className="h-4 w-4 mr-2" />
            处理历史
          </Button>
          <Button variant="outline">
            <Settings className="h-4 w-4 mr-2" />
            设置
          </Button>
        </div>
      </div>

      {/* Main Content */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="tools">
            <Layers className="h-4 w-4 mr-2" />
            工具箱
          </TabsTrigger>
          <TabsTrigger value="batch">
            <Copy className="h-4 w-4 mr-2" />
            批量处理
          </TabsTrigger>
          <TabsTrigger value="workflow">
            <Zap className="h-4 w-4 mr-2" />
            工作流
          </TabsTrigger>
          <TabsTrigger value="history">
            <History className="h-4 w-4 mr-2" />
            历史记录
          </TabsTrigger>
        </TabsList>

        <TabsContent value="tools" className="space-y-6">
          <div className="grid lg:grid-cols-3 gap-6">
            {/* Left Panel: Tool Selection */}
            <div className="lg:col-span-2 space-y-4">
              {/* Search */}
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="搜索PDF工具..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-9"
                />
              </div>

              {/* Operation Categories */}
              {categories.map((category) => {
                const categoryOps = category.operations.filter(op =>
                  op.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                  op.description.toLowerCase().includes(searchTerm.toLowerCase())
                );

                if (categoryOps.length === 0) return null;

                return (
                  <div key={category.id} className="space-y-3">
                    <div className="flex items-center space-x-2">
                      <FolderOpen className="h-5 w-5 text-blue-600" />
                      <h3 className="text-lg font-semibold">{category.name}</h3>
                      <Badge variant="secondary">{categoryOps.length}</Badge>
                    </div>
                    <div className="grid md:grid-cols-2 gap-3">
                      {categoryOps.map((operation) => (
                        <OperationCard
                          key={operation.id}
                          operation={operation}
                          onSelect={handleOperationSelect}
                          isSelected={selectedOperation?.id === operation.id}
                        />
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Right Panel: File Upload & Processing */}
            <div className="space-y-4">
              {/* File Upload */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">文件上传</CardTitle>
                  <CardDescription>
                    {selectedOperation 
                      ? `已选择: ${selectedOperation.name}` 
                      : '请先选择处理操作'
                    }
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <FileUploadArea
                    onFilesSelected={handleFilesSelected}
                    multiple={selectedOperation?.id === 'merge'}
                    accept=".pdf"
                  />
                  {uploadedFiles.length > 0 && (
                    <div className="mt-4 space-y-2">
                      <h4 className="font-medium">已选择文件:</h4>
                      {uploadedFiles.map((file, index) => (
                        <div key={index} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                          <div className="flex items-center space-x-2">
                            <File className="h-4 w-4" />
                            <span className="text-sm">{file.name}</span>
                          </div>
                          <Badge variant="outline" className="text-xs">
                            {(file.size / 1024 / 1024).toFixed(1)}MB
                          </Badge>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
                <CardFooter>
                  <Button
                    className="w-full"
                    onClick={handleProcessFiles}
                    disabled={!selectedOperation || uploadedFiles.length === 0 || loading}
                  >
                    {loading ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        处理中...
                      </>
                    ) : (
                      <>
                        <Play className="h-4 w-4 mr-2" />
                        开始处理
                      </>
                    )}
                  </Button>
                </CardFooter>
              </Card>

              {/* Processing Status */}
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">处理状态</CardTitle>
                </CardHeader>
                <CardContent>
                  <ProcessingStatus
                    results={processingResults}
                    onDownload={handleDownload}
                    onRetry={handleRetry}
                  />
                </CardContent>
              </Card>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="batch" className="space-y-4">
          <Card className="p-8 text-center">
            <Copy className="h-12 w-12 mx-auto text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">批量处理</h3>
            <p className="text-gray-500 mb-4">
              一次性处理多个PDF文件，提高工作效率
            </p>
            <Button>
              <Upload className="h-4 w-4 mr-2" />
              上传多个文件
            </Button>
          </Card>
        </TabsContent>

        <TabsContent value="workflow" className="space-y-4">
          <Card className="p-8 text-center">
            <Zap className="h-12 w-12 mx-auto text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">自定义工作流</h3>
            <p className="text-gray-500 mb-4">
              组合多个PDF操作，创建自定义处理流程
            </p>
            <Button>
              <Settings className="h-4 w-4 mr-2" />
              创建工作流
            </Button>
          </Card>
        </TabsContent>

        <TabsContent value="history" className="space-y-4">
          <div className="space-y-4">
            {processingHistory.map((item) => (
              <Card key={item.id}>
                <CardContent className="p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <File className="h-5 w-5 text-gray-400" />
                      <div>
                        <h4 className="font-medium">{item.operation_type}</h4>
                        <p className="text-sm text-gray-500">
                          {item.input_files.join(', ')}
                        </p>
                        <p className="text-xs text-gray-400">
                          <Calendar className="h-3 w-3 inline mr-1" />
                          {new Date(item.created_at).toLocaleString()}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Badge variant={item.status === 'completed' ? 'default' : 'secondary'}>
                        {item.status}
                      </Badge>
                      <Button variant="ghost" size="sm">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default PDFProcessor;