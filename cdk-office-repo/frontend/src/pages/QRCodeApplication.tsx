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
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Avatar,
} from '@mui/material';
import { 
  QrCode, 
  QrCode2, 
  Add, 
  Edit, 
  Delete, 
  FileDownload,
  Share,
  Preview
} from '@mui/icons-material';

interface QRCodeForm {
  id: string;
  name: string;
  type: 'survey' | 'registration' | 'feedback' | 'checkin';
  createdAt: string;
  createdBy: string;
}

// 模拟二维码表单数据
const mockForms: QRCodeForm[] = [
  {
    id: '1',
    name: '员工签到表单',
    type: 'checkin',
    createdAt: '2023-10-15',
    createdBy: '张三',
  },
  {
    id: '2',
    name: '会议室预订表单',
    type: 'registration',
    createdAt: '2023-10-10',
    createdBy: '李四',
  },
  {
    id: '3',
    name: '员工满意度调查',
    type: 'survey',
    createdAt: '2023-10-05',
    createdBy: '王五',
  },
  {
    id: '4',
    name: '访客登记表单',
    type: 'registration',
    createdAt: '2023-09-28',
    createdBy: '赵六',
  },
];

const QRCodeApplication: React.FC = () => {
  const [forms] = useState<QRCodeForm[]>(mockForms);
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [selectedForm, setSelectedForm] = useState<QRCodeForm | null>(null);

  const getFormTypeLabel = (type: QRCodeForm['type']) => {
    switch (type) {
      case 'survey': return '调查问卷';
      case 'registration': return '登记表';
      case 'feedback': return '反馈表';
      case 'checkin': return '签到表';
      default: return '未知类型';
    }
  };

  const getFormTypeColor = (type: QRCodeForm['type']) => {
    switch (type) {
      case 'survey': return 'primary';
      case 'registration': return 'secondary';
      case 'feedback': return 'success';
      case 'checkin': return 'warning';
      default: return 'default';
    }
  };

  return (
    <Box sx={{ width: '100%', p: 2 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          二维码应用
        </Typography>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={() => setIsCreateDialogOpen(true)}
        >
          创建表单
        </Button>
      </Box>

      <Grid container spacing={3}>
        {forms.map((form) => (
          <Grid item xs={12} sm={6} md={4} key={form.id}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Avatar sx={{ mr: 2, bgcolor: 'primary.main' }}>
                    <QrCode2 />
                  </Avatar>
                  <Box>
                    <Typography variant="h6" component="h3" noWrap>
                      {form.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {form.createdAt} • {form.createdBy}
                    </Typography>
                  </Box>
                </Box>
                
                <Box sx={{ display: 'flex', justifyContent: 'center', my: 2 }}>
                  <QrCode sx={{ width: 100, height: 100, color: 'grey.500' }} />
                </Box>
                
                <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                  <Chip 
                    label={getFormTypeLabel(form.type)} 
                    color={getFormTypeColor(form.type) as any}
                    size="small"
                  />
                </Box>
              </CardContent>
              <CardActions sx={{ justifyContent: 'space-between' }}>
                <Button size="small" startIcon={<Preview />}>
                  预览
                </Button>
                <Box>
                  <IconButton size="small" onClick={() => setSelectedForm(form)}>
                    <Edit />
                  </IconButton>
                  <IconButton size="small">
                    <Share />
                  </IconButton>
                  <IconButton size="small">
                    <FileDownload />
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

      {/* 创建表单对话框 */}
      <Dialog open={isCreateDialogOpen} onClose={() => setIsCreateDialogOpen(false)}>
        <DialogTitle>创建二维码表单</DialogTitle>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: '1rem', minWidth: '400px', mt: '1rem' }}>
            <TextField
              label="表单名称"
              fullWidth
            />
            <FormControl fullWidth>
              <InputLabel>表单类型</InputLabel>
              <Select defaultValue="checkin">
                <MenuItem value="checkin">签到表</MenuItem>
                <MenuItem value="registration">登记表</MenuItem>
                <MenuItem value="survey">调查问卷</MenuItem>
                <MenuItem value="feedback">反馈表</MenuItem>
              </Select>
            </FormControl>
            <TextField
              label="表单描述"
              multiline
              rows={3}
              fullWidth
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsCreateDialogOpen(false)}>取消</Button>
          <Button variant="contained">创建</Button>
        </DialogActions>
      </Dialog>

      {/* 编辑表单对话框 */}
      <Dialog open={!!selectedForm} onClose={() => setSelectedForm(null)}>
        <DialogTitle>编辑二维码表单</DialogTitle>
        <DialogContent>
          {selectedForm && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: '1rem', minWidth: '400px', mt: '1rem' }}>
              <TextField
                label="表单名称"
                defaultValue={selectedForm.name}
                fullWidth
              />
              <FormControl fullWidth>
                <InputLabel>表单类型</InputLabel>
                <Select defaultValue={selectedForm.type}>
                  <MenuItem value="checkin">签到表</MenuItem>
                  <MenuItem value="registration">登记表</MenuItem>
                  <MenuItem value="survey">调查问卷</MenuItem>
                  <MenuItem value="feedback">反馈表</MenuItem>
                </Select>
              </FormControl>
              <TextField
                label="表单描述"
                multiline
                rows={3}
                fullWidth
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelectedForm(null)}>取消</Button>
          <Button variant="contained">保存</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default QRCodeApplication;