import React from 'react';
import { 
  AppBar, 
  Toolbar, 
  Typography, 
  Button, 
  Box,
  IconButton,
  Menu,
  MenuItem,
  Divider
} from '@mui/material';
import { 
  AccountCircle, 
  People, 
  Description, 
  SmartToy, 
  QrCode, 
  Archive, 
  BarChart,
  CheckCircle,
  Menu as MenuIcon
} from '@mui/icons-material';
import Link from 'next/link';
import { useRouter } from 'next/router';

const Navigation: React.FC = () => {
  const router = useRouter();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  
  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };
  
  const handleMenuClose = () => {
    setAnchorEl(null);
  };
  
  const isActive = (path: string) => {
    return router.pathname.startsWith(path);
  };

  return (
    <AppBar position="static">
      <Toolbar>
        <Typography 
          variant="h6" 
          component="div" 
          sx={{ flexGrow: 1, cursor: 'pointer' }}
          onClick={() => router.push('/')}
        >
          CDK-Office
        </Typography>
        
        {/* 桌面端导航 */}
        <Box sx={{ display: { xs: 'none', md: 'flex' }, gap: 1 }}>
          <Button 
            color="inherit" 
            component={Link} 
            href="/employee-management"
            sx={{ 
              bgcolor: isActive('/employee-management') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
              '&:hover': {
                bgcolor: 'rgba(255, 255, 255, 0.2)'
              }
            }}
          >
            <People sx={{ mr: 1 }} />
            员工管理
          </Button>
          <Button 
            color="inherit" 
            component={Link} 
            href="/document-management"
            sx={{ 
              bgcolor: isActive('/document-management') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
              '&:hover': {
                bgcolor: 'rgba(255, 255, 255, 0.2)'
              }
            }}
          >
            <Description sx={{ mr: 1 }} />
            文档管理
          </Button>
          <Button 
            color="inherit" 
            component={Link} 
            href="/ai-assistant"
            sx={{ 
              bgcolor: isActive('/ai-assistant') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
              '&:hover': {
                bgcolor: 'rgba(255, 255, 255, 0.2)'
              }
            }}
          >
            <SmartToy sx={{ mr: 1 }} />
            AI助手
          </Button>
          <Button 
            color="inherit" 
            component={Link} 
            href="/approval-management"
            sx={{ 
              bgcolor: isActive('/approval-management') ? 'rgba(255, 255, 255, 0.1)' : 'transparent',
              '&:hover': {
                bgcolor: 'rgba(255, 255, 255, 0.2)'
              }
            }}
          >
            <CheckCircle sx={{ mr: 1 }} />
            审批管理
          </Button>
        </Box>
        
        {/* 移动端导航菜单 */}
        <IconButton
          size="large"
          edge="start"
          color="inherit"
          aria-label="menu"
          sx={{ mr: 2, display: { xs: 'flex', md: 'none' } }}
          onClick={handleMenuOpen}
        >
          <MenuIcon />
        </IconButton>
        <Menu
          id="menu-appbar"
          anchorEl={anchorEl}
          anchorOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
          keepMounted
          transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
          open={Boolean(anchorEl)}
          onClose={handleMenuClose}
        >
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/employee-management"
            selected={isActive('/employee-management')}
          >
            <People sx={{ mr: 1 }} />
            员工管理
          </MenuItem>
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/document-management"
            selected={isActive('/document-management')}
          >
            <Description sx={{ mr: 1 }} />
            文档管理
          </MenuItem>
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/ai-assistant"
            selected={isActive('/ai-assistant')}
          >
            <SmartToy sx={{ mr: 1 }} />
            AI助手
          </MenuItem>
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/approval-management"
            selected={isActive('/approval-management')}
          >
            <CheckCircle sx={{ mr: 1 }} />
            审批管理
          </MenuItem>
          <Divider />
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/qrcode"
            selected={isActive('/qrcode')}
          >
            <QrCode sx={{ mr: 1 }} />
            二维码应用
          </MenuItem>
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/archive"
            selected={isActive('/archive')}
          >
            <Archive sx={{ mr: 1 }} />
            知识库归档
          </MenuItem>
          <MenuItem 
            onClick={handleMenuClose}
            component={Link} 
            href="/statistics"
            selected={isActive('/statistics')}
          >
            <BarChart sx={{ mr: 1 }} />
            数据统计
          </MenuItem>
        </Menu>
        
        {/* 用户菜单 */}
        <IconButton
          size="large"
          aria-label="account of current user"
          aria-controls="menu-appbar"
          aria-haspopup="true"
          color="inherit"
        >
          <AccountCircle />
        </IconButton>
      </Toolbar>
    </AppBar>
  );
};

export default Navigation;