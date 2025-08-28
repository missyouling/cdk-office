import React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import {
  Users,
  FileText,
  Bot,
  QrCode,
  Archive,
  BarChart3,
  CheckCircle,
  Bell,
  Menu,
  GitBranch,
  ListTodo,
  Settings,
  ClipboardList,
  User,
  Grid3X3
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@/components/ui/navigation-menu';
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from '@/components/ui/sheet';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';

const Navigation: React.FC = () => {
  const router = useRouter();
  
  const isActive = (path: string) => {
    return router.pathname.startsWith(path);
  };

  const navigationItems = [
    { href: '/app-center', label: '应用中心', icon: Grid3X3 },
    { href: '/employee-management', label: '员工管理', icon: Users },
    { href: '/document-management', label: '文档管理', icon: FileText },
    { href: '/contracts', label: '电子合同', icon: ClipboardList },
    { href: '/ai-assistant', label: 'AI助手', icon: Bot },
    { href: '/approval-management', label: '审批管理', icon: CheckCircle },
    { href: '/workflow-management', label: '工作流', icon: GitBranch },
    { href: '/task-management', label: '我的任务', icon: ListTodo },
    { href: '/notification-center', label: '通知中心', icon: Bell },
  ];

  const mobileMenuItems = [
    ...navigationItems,
    { href: '/survey', label: '调查问卷', icon: ClipboardList },
    { href: '/workflow-instances', label: '工作流实例', icon: ClipboardList },
    { href: '/workflow-designer', label: '工作流设计器', icon: Settings },
    { href: '/qrcode', label: '二维码应用', icon: QrCode },
    { href: '/archive', label: '知识库归档', icon: Archive },
    { href: '/statistics', label: '数据统计', icon: BarChart3 },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center">
        {/* Logo */}
        <div className="mr-4 flex">
          <Link href="/" className="mr-6 flex items-center space-x-2">
            <span className="hidden font-bold sm:inline-block">
              CDK-Office
            </span>
          </Link>
        </div>

        {/* Desktop Navigation */}
        <div className="flex flex-1 items-center justify-between space-x-2 md:justify-end">
          <NavigationMenu className="hidden md:flex">
            <NavigationMenuList>
              {navigationItems.map((item) => {
                const IconComponent = item.icon;
                return (
                  <NavigationMenuItem key={item.href}>
                    <Link href={item.href} legacyBehavior passHref>
                      <Button
                        variant={isActive(item.href) ? "secondary" : "ghost"}
                        className="flex items-center space-x-2"
                        asChild
                      >
                        <a>
                          <IconComponent className="h-4 w-4" />
                          <span>{item.label}</span>
                        </a>
                      </Button>
                    </Link>
                  </NavigationMenuItem>
                );
              })}
            </NavigationMenuList>
          </NavigationMenu>

          {/* Mobile Navigation */}
          <Sheet>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                className="mr-2 px-0 text-base hover:bg-transparent focus-visible:bg-transparent focus-visible:ring-0 focus-visible:ring-offset-0 md:hidden"
              >
                <Menu className="h-5 w-5" />
                <span className="sr-only">Toggle Menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="pr-0">
              <div className="px-6">
                <Link href="/" className="flex items-center">
                  <span className="font-bold">CDK-Office</span>
                </Link>
              </div>
              <div className="my-4 h-[calc(100vh-8rem)] pb-10 pl-6">
                <div className="flex flex-col space-y-3">
                  {mobileMenuItems.map((item) => {
                    const IconComponent = item.icon;
                    return (
                      <Link key={item.href} href={item.href}>
                        <Button
                          variant={isActive(item.href) ? "secondary" : "ghost"}
                          className="w-full justify-start"
                        >
                          <IconComponent className="mr-2 h-4 w-4" />
                          {item.label}
                        </Button>
                      </Link>
                    );
                  })}
                </div>
              </div>
            </SheetContent>
          </Sheet>

          {/* User Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                <User className="h-4 w-4" />
                <span className="sr-only">Open user menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56" align="end" forceMount>
              <DropdownMenuItem>
                <User className="mr-2 h-4 w-4" />
                <span>个人资料</span>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Settings className="mr-2 h-4 w-4" />
                <span>设置</span>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <span>退出登录</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  );
};

export default Navigation;