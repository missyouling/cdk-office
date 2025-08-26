'use client';

import React from 'react';
import { Box, Typography, Grid, Card, CardContent, CardActions, Button } from '@mui/material';
import { 
  People, 
  Description, 
  SmartToy, 
  QrCode, 
  Archive, 
  BarChart,
  CheckCircle,
  Upload,
  Download
} from '@mui/icons-material';
import Link from 'next/link';

export default function Home() {
  const features = [
    {
      title: '员工管理',
      description: '管理员工信息，支持数据导入导出、多选、排序、筛选、行内编辑等功能',
      icon: <People fontSize="large" />,
      link: '/employee-management',
    },
    {
      title: '文档管理',
      description: '管理企业文档，支持版本控制、权限管理、在线预览、标签分类等功能',
      icon: <Description fontSize="large" />,
      link: '/document-management',
    },
    {
      title: 'AI助手',
      description: '集成Dify AI平台，提供智能问答、文档处理和知识管理能力',
      icon: <SmartToy fontSize="large" />,
      link: '/ai-assistant',
    },
    {
      title: '二维码应用',
      description: '支持动态表单、员工签到、在线订餐、问卷调查和访客登记等应用场景',
      icon: <QrCode fontSize="large" />,
      link: '/qrcode',
    },
    {
      title: '审批管理',
      description: '管理文档审批流程，支持自定义审批模板和多级审批',
      icon: <CheckCircle fontSize="large" />,
      link: '/approval-management',
    },
    {
      title: '知识库归档',
      description: '自动归档知识库文档，支持自定义归档规则和资料目录生成',
      icon: <Archive fontSize="large" />,
      link: '/archive',
    },
    {
      title: '数据统计',
      description: '提供全面的数据统计和分析功能，帮助管理者了解企业运营状况',
      icon: <BarChart fontSize="large" />,
      link: '/statistics',
    },
  ];

  return (
    <Box sx={{ flexGrow: 1 }}>
      <Typography variant="h3" component="h1" gutterBottom align="center">
        CDK-Office 企业内容管理平台
      </Typography>
      <Typography variant="h6" component="h2" gutterBottom align="center" color="text.secondary">
        集成Dify AI平台，实现智能文档管理、AI问答和知识库管理功能
      </Typography>
      
      <Grid container spacing={3} sx={{ mt: 4 }}>
        {features.map((feature, index) => (
          <Grid item xs={12} sm={6} md={4} key={index}>
            <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
              <CardContent sx={{ flexGrow: 1 }}>
                <Box sx={{ display: 'flex', justifyContent: 'center', mb: 2 }}>
                  {feature.icon}
                </Box>
                <Typography gutterBottom variant="h5" component="h3" align="center">
                  {feature.title}
                </Typography>
                <Typography align="center" color="text.secondary">
                  {feature.description}
                </Typography>
              </CardContent>
              <CardActions sx={{ justifyContent: 'center', pb: 2 }}>
                <Button 
                  size="small" 
                  variant="contained" 
                  component={Link} 
                  href={feature.link}
                >
                  立即体验
                </Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>
      
      <Box sx={{ mt: 6, textAlign: 'center' }}>
        <Typography variant="h5" gutterBottom>
          快速开始
        </Typography>
        <Box sx={{ display: 'flex', justifyContent: 'center', gap: 2, mt: 2 }}>
          <Button 
            variant="contained" 
            startIcon={<Upload />}
            size="large"
            component={Link}
            href="/employee-management"
          >
            导入员工数据
          </Button>
          <Button 
            variant="outlined" 
            startIcon={<Download />}
            size="large"
            component={Link}
            href="/document-management"
          >
            浏览文档
          </Button>
        </Box>
      </Box>
    </Box>
  );
}