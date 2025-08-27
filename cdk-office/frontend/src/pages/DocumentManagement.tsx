import React, { useState } from 'react';
import { 
  FileText, 
  FileImage, 
  Folder, 
  Upload, 
  Download, 
  Share, 
  Trash2, 
  Search,
  FileSpreadsheet,
  Image as ImageIcon
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogFooter,
  DialogDescription
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Label } from '@/components/ui/label';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

interface Document {
  id: string;
  name: string;
  type: 'pdf' | 'doc' | 'xls' | 'image' | 'folder';
  size: string;
  lastModified: string;
  owner: string;
  tags: string[];
}

// 模拟文档数据
const mockDocuments: Document[] = [
  {
    id: '1',
    name: '公司年度报告.pdf',
    type: 'pdf',
    size: '2.4 MB',
    lastModified: '2023-10-15',
    owner: '张三',
    tags: ['财务', '年度报告'],
  },
  {
    id: '2',
    name: '产品需求文档.doc',
    type: 'doc',
    size: '1.1 MB',
    lastModified: '2023-10-10',
    owner: '李四',
    tags: ['产品', '需求'],
  },
  {
    id: '3',
    name: '销售数据.xls',
    type: 'xls',
    size: '0.8 MB',
    lastModified: '2023-10-05',
    owner: '王五',
    tags: ['销售', '数据'],
  },
  {
    id: '4',
    name: '项目图片',
    type: 'folder',
    size: '5.2 MB',
    lastModified: '2023-10-01',
    owner: '赵六',
    tags: ['项目', '图片'],
  },
  {
    id: '5',
    name: '公司Logo.png',
    type: 'image',
    size: '0.3 MB',
    lastModified: '2023-09-28',
    owner: '孙七',
    tags: ['品牌', 'Logo'],
  },
];

const DocumentManagement: React.FC = () => {
  const [documents] = useState<Document[]>(mockDocuments);
  const [searchTerm, setSearchTerm] = useState('');
  const [isUploadDialogOpen, setIsUploadDialogOpen] = useState(false);

  const getDocumentIcon = (type: Document['type']) => {
    switch (type) {
      case 'pdf':
        return <FileText className="h-8 w-8 text-red-500" />;
      case 'doc':
        return <FileText className="h-8 w-8 text-blue-500" />;
      case 'xls':
        return <FileSpreadsheet className="h-8 w-8 text-green-500" />;
      case 'image':
        return <ImageIcon className="h-8 w-8 text-blue-400" />;
      case 'folder':
        return <Folder className="h-8 w-8 text-yellow-500" />;
      default:
        return <FileText className="h-8 w-8" />;
    }
  };

  const filteredDocuments = documents.filter(doc => 
    doc.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    doc.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  return (
    <div className="container mx-auto p-6 space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">文档管理</h1>
          <p className="text-muted-foreground mt-2">
            管理和组织您的文档
          </p>
        </div>
        <div className="flex gap-2">
          <div className="relative">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索文档..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-8 w-[200px] md:w-[300px]"
            />
          </div>
          <Button onClick={() => setIsUploadDialogOpen(true)}>
            <Upload className="h-4 w-4 mr-2" />
            上传文档
          </Button>
        </div>
      </div>

      {/* 文档网格 */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredDocuments.map((doc) => (
          <Card key={doc.id} className="flex flex-col">
            <CardHeader className="pb-2">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-muted">
                  {getDocumentIcon(doc.type)}
                </div>
                <div className="flex-1 min-w-0">
                  <CardTitle className="text-lg truncate">{doc.name}</CardTitle>
                  <p className="text-sm text-muted-foreground">
                    {doc.size} • {doc.lastModified}
                  </p>
                </div>
              </div>
            </CardHeader>
            <CardContent className="flex-1 pb-2">
              <div className="flex flex-wrap gap-1 mb-3">
                {doc.tags.map((tag, index) => (
                  <Badge key={index} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
              
              <div className="flex items-center">
                <Avatar className="h-6 w-6">
                  <AvatarFallback className="text-xs">
                    {doc.owner.charAt(0)}
                  </AvatarFallback>
                </Avatar>
                <span className="text-sm text-muted-foreground ml-2">
                  {doc.owner}
                </span>
              </div>
            </CardContent>
            <CardFooter className="flex justify-between p-4 pt-2">
              <Button variant="outline" size="sm">
                <Download className="h-4 w-4 mr-2" />
                下载
              </Button>
              <div className="flex gap-1">
                <Button variant="ghost" size="icon">
                  <Share className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </CardFooter>
          </Card>
        ))}
      </div>

      {/* 上传文档对话框 */}
      <Dialog open={isUploadDialogOpen} onOpenChange={setIsUploadDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>上传文档</DialogTitle>
            <DialogDescription>
              选择要上传的文档文件
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="document-name">文档名称</Label>
              <Input id="document-name" placeholder="输入文档名称" />
            </div>
            
            <div className="space-y-2">
              <Label>文档类型</Label>
              <Select defaultValue="pdf">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pdf">PDF文档</SelectItem>
                  <SelectItem value="doc">Word文档</SelectItem>
                  <SelectItem value="xls">Excel表格</SelectItem>
                  <SelectItem value="image">图片</SelectItem>
                  <SelectItem value="other">其他</SelectItem>
                </SelectContent>
              </Select>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="file-upload">选择文件</Label>
              <div className="flex items-center gap-2">
                <Input id="file-upload" type="file" className="flex-1" />
                <Button variant="outline">
                  <Upload className="h-4 w-4 mr-2" />
                  选择
                </Button>
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="tags">标签</Label>
              <Input id="tags" placeholder="输入标签，用逗号分隔" />
            </div>
          </div>
          
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsUploadDialogOpen(false)}>
              取消
            </Button>
            <Button>
              <Upload className="h-4 w-4 mr-2" />
              上传
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default DocumentManagement;