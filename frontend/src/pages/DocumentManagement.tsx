import React, { useState } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  CardActions,
  Button,
  Grid,
  Chip,
  Avatar,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  IconButton,
} from '@mui/material';
import { 
  Description, 
  PictureAsPdf, 
  Article, 
  Image, 
  Folder, 
  Upload, 
  Download, 
  Share, 
  Delete,
  Search,
  Add
} from '@mui/icons-material';

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
        return <PictureAsPdf sx={{ color: '#ff0000' }} />;
      case 'doc':
        return <Article sx={{ color: '#2b579a' }} />;
      case 'xls':
        return <Article sx={{ color: '#217346' }} />;
      case 'image':
        return <Image sx={{ color: '#4285f4' }} />;
      case 'folder':
        return <Folder sx={{ color: '#fdbd00' }} />;
      default:
        return <Description />;
    }
  };

  const filteredDocuments = documents.filter(doc => 
    doc.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    doc.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  return (
    <Box sx={{ width: '100%', p: 2 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          文档管理
        </Typography>
        <Box sx={{ display: 'flex', gap: 2 }}>
          <TextField
            variant="outlined"
            size="small"
            placeholder="搜索文档..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: <Search sx={{ mr: 1, color: 'text.secondary' }} />,
            }}
          />
          <Button
            variant="contained"
            startIcon={<Upload />}
            onClick={() => setIsUploadDialogOpen(true)}
          >
            上传文档
          </Button>
        </Box>
      </Box>

      <Grid container spacing={3}>
        {filteredDocuments.map((doc) => (
          <Grid item xs={12} sm={6} md={4} key={doc.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Avatar sx={{ mr: 2, bgcolor: 'transparent' }}>
                    {getDocumentIcon(doc.type)}
                  </Avatar>
                  <Box>
                    <Typography variant="h6" component="h3" noWrap>
                      {doc.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {doc.size} • {doc.lastModified}
                    </Typography>
                  </Box>
                </Box>
                
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mt: 2 }}>
                  {doc.tags.map((tag, index) => (
                    <Chip key={index} label={tag} size="small" />
                  ))}
                </Box>
                
                <Box sx={{ display: 'flex', alignItems: 'center', mt: 2 }}>
                  <Avatar sx={{ width: 24, height: 24, fontSize: '0.75rem' }}>
                    {doc.owner.charAt(0)}
                  </Avatar>
                  <Typography variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                    {doc.owner}
                  </Typography>
                </Box>
              </CardContent>
              <CardActions sx={{ justifyContent: 'space-between' }}>
                <Button size="small" startIcon={<Download />}>
                  下载
                </Button>
                <Box>
                  <IconButton size="small">
                    <Share />
                  </IconButton>
                  <IconButton size="small">
                    <Delete />
                  </IconButton>
                </Box>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* 上传文档对话框 */}
      <Dialog open={isUploadDialogOpen} onClose={() => setIsUploadDialogOpen(false)}>
        <DialogTitle>上传文档</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: '1rem', minWidth: '400px', mt: '1rem' }}>
            <TextField
              label="文档名称"
              fullWidth
            />
            <FormControl fullWidth>
              <InputLabel>文档类型</InputLabel>
              <Select defaultValue="pdf">
                <MenuItem value="pdf">PDF文档</MenuItem>
                <MenuItem value="doc">Word文档</MenuItem>
                <MenuItem value="xls">Excel表格</MenuItem>
                <MenuItem value="image">图片</MenuItem>
                <MenuItem value="other">其他</MenuItem>
              </Select>
            </FormControl>
            <Button
              variant="outlined"
              component="label"
              startIcon={<Upload />}
            >
              选择文件
              <input type="file" hidden />
            </Button>
            <TextField
              label="标签"
              placeholder="输入标签，用逗号分隔"
              fullWidth
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsUploadDialogOpen(false)}>取消</Button>
          <Button variant="contained">上传</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default DocumentManagement;