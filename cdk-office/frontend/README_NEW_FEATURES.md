# 应用中心和电子合同模块

## 功能概述

本次开发完成了以下两个核心模块：

### 1. 应用中心页面 (`/app-center`)

**位置：** `src/app/app-center/page.tsx`

**功能特点：**
- 🎯 **网格布局展示**：使用 Shadcn UI 卡片组件展示所有应用
- 🔍 **搜索功能**：支持按应用名称和描述搜索
- 📂 **分类筛选**：核心应用、AI应用、业务应用、工具应用
- ⭐ **推荐应用**：突出显示重要应用
- 🏷️ **状态标签**：NEW、HOT、BETA、即将推出等状态
- 📊 **统计信息**：显示应用总数、可用数量等

**技术实现：**
- Next.js 15 + TypeScript
- Shadcn UI 组件库
- Tailwind CSS 样式
- 响应式设计
- 严格遵循项目UI设计规范

### 2. 电子合同模块 (`/contracts`)

**位置：** `src/app/contracts/page.tsx`

**功能特点：**
- 📋 **数据表格**：使用 `@tanstack/react-table` 实现高性能表格
- 📊 **状态管理**：草稿、待发送、签署中、已完成等状态
- 👥 **签署方显示**：头像组件展示多个签署方
- 📈 **进度条**：签署进度可视化
- 🔍 **搜索筛选**：全局搜索 + 状态筛选
- 📄 **分页功能**：表格分页显示
- ⚡ **操作菜单**：查看、编辑、签署、下载、删除等操作
- 📊 **统计卡片**：各状态合同数量统计

**表格列配置：**
- 合同名称（支持排序）
- 状态标签（支持筛选）
- 签署进度条
- 签署方头像
- 创建时间（支持排序）
- 过期时间
- 创建者
- 操作菜单

## 文件结构

```
src/
├── app/
│   ├── app-center/
│   │   └── page.tsx                 # 应用中心主页
│   ├── contracts/
│   │   ├── page.tsx                 # 合同列表页面
│   │   ├── [id]/
│   │   │   └── page.tsx             # 合同详情页面（占位符）
│   │   └── create/
│   │       └── page.tsx             # 创建合同页面（占位符）
│   └── page.tsx                     # 主页（已更新链接）
├── components/
│   ├── Navigation.tsx               # 导航组件（已更新）
│   └── ui/                          # Shadcn UI 组件
└── types/
    └── contract.ts                  # 电子合同类型定义
```

## 类型定义

### 电子合同核心类型

```typescript
// 合同状态
type ContractStatus = 'draft' | 'pending' | 'signing' | 'completed' | 'rejected' | 'cancelled' | 'expired';

// 签署方类型
type SignerType = 'person' | 'company';

// 签署状态
type SignerStatus = 'pending' | 'signed' | 'rejected';

// 合同接口
interface Contract {
  id: string;
  title: string;
  description?: string;
  status: ContractStatus;
  progress: number;
  signers: ContractSigner[];
  createdAt: string;
  expireTime: string;
  // ... 其他字段
}
```

### 应用中心类型

```typescript
interface AppCenterCard {
  id: string;
  title: string;
  description: string;
  icon: any;
  link: string;
  badge?: string;
  category: 'core' | 'ai' | 'business' | 'tools';
  status?: 'active' | 'beta' | 'coming_soon';
}
```

## 导航集成

### 桌面端导航
- ✅ 添加了"应用中心"入口
- ✅ 更新了"电子合同"链接路径为 `/contracts`

### 移动端导航
- ✅ 完整的应用列表
- ✅ 响应式设计适配

### 主页面更新
- ✅ 添加"浏览全部应用"按钮
- ✅ 更新电子合同卡片链接

## 技术栈

- **框架：** Next.js 15
- **语言：** TypeScript
- **UI库：** Shadcn UI
- **样式：** Tailwind CSS
- **表格：** @tanstack/react-table ^8.11.8
- **图标：** Lucide React
- **组件：** Radix UI

## Shadcn UI 组件使用

严格遵循项目的 Shadcn UI 设计规范，使用的组件包括：

- `Card, CardContent, CardHeader, CardTitle`
- `Button`
- `Badge`
- `Input`
- `Table, TableBody, TableCell, TableHead, TableHeader, TableRow`
- `Tabs, TabsContent, TabsList, TabsTrigger`
- `DropdownMenu`
- `Avatar, AvatarFallback, AvatarImage`
- `Progress`
- `Separator`
- `Alert Dialog`
- `Toast`

## 功能特性

### 🎨 设计一致性
- 严格遵循 Shadcn UI 设计系统
- 统一的颜色方案和视觉风格
- 响应式设计，适配各种屏幕尺寸

### ⚡ 性能优化
- 使用 `@tanstack/react-table` 实现高性能表格
- 虚拟化渲染支持大量数据
- 客户端排序、筛选、分页

### 🔍 用户体验
- 直观的搜索和筛选功能
- 清晰的状态指示和进度显示
- 便捷的操作入口和导航

### 🛡️ 类型安全
- 完整的 TypeScript 类型定义
- 接口和组件的类型保护
- 编译时错误检查

## 后续扩展

### 占位符页面
已创建但需要进一步开发的页面：
- `/contracts/[id]` - 合同详情页面
- `/contracts/create` - 创建合同页面
- `/contracts/[id]/edit` - 编辑合同页面
- `/contracts/[id]/sign` - 签署合同页面

### 建议的后续功能
1. **合同编辑器**：集成富文本编辑器
2. **电子签名**：集成签名画板
3. **文件上传**：支持合同文件上传
4. **模板系统**：合同模板管理
5. **通知系统**：签署提醒和状态通知
6. **审批流程**：多级审批工作流
7. **权限控制**：基于角色的访问控制

## 使用说明

1. **访问应用中心**：导航至 `/app-center` 或点击主页的"浏览全部应用"按钮
2. **查看合同列表**：导航至 `/contracts` 或点击导航栏的"电子合同"
3. **创建新合同**：在合同列表页点击"新建合同"按钮
4. **操作合同**：使用表格行末的操作菜单进行各种操作

## 注意事项

- 所有功能使用模拟数据，实际使用时需要集成后端API
- 部分页面为占位符实现，需要根据具体需求完善
- 已确保与现有项目架构和设计规范的完全兼容