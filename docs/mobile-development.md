# 微信小程序开发文档

## 项目结构

```
mobile/
└── wechat-miniprogram/
    ├── app.js                 # 小程序全局逻辑
    ├── app.json               # 小程序全局配置
    ├── app.wxss               # 小程序全局样式
    ├── images/                # 图片资源
    ├── pages/                 # 页面目录
    │   ├── index/             # 首页
    │   ├── auth/              # 认证相关页面
    │   ├── document/          # 文档相关页面
    │   ├── employee/          # 员工相关页面
    │   ├── app/               # 应用中心相关页面
    │   └── business/          # 业务中心相关页面
    └── utils/                 # 工具类
```

## 已完成的页面

### 1. 首页 (index)
- 功能：展示系统主要功能模块入口
- 文件：
  - `pages/index/index.js` - 页面逻辑
  - `pages/index/index.wxml` - 页面结构
  - `pages/index/index.wxss` - 页面样式
  - `pages/index/index.json` - 页面配置

### 2. 认证页面 (auth)
- 登录页面：
  - `pages/auth/login.js` - 登录逻辑
  - `pages/auth/login.wxml` - 登录页面结构
  - `pages/auth/login.wxss` - 登录页面样式
  - `pages/auth/login.json` - 登录页面配置
- 注册页面：
  - `pages/auth/register.js` - 注册逻辑
  - `pages/auth/register.wxml` - 注册页面结构
  - `pages/auth/register.wxss` - 注册页面样式
  - `pages/auth/register.json` - 注册页面配置

### 3. 文档管理页面 (document)
- 文档列表页面：
  - `pages/document/list.js` - 文档列表逻辑
  - `pages/document/list.wxml` - 文档列表页面结构
  - `pages/document/list.wxss` - 文档列表页面样式
  - `pages/document/list.json` - 文档列表页面配置
- 文档详情页面：
  - `pages/document/detail.js` - 文档详情逻辑
  - `pages/document/detail.wxml` - 文档详情页面结构
  - `pages/document/detail.wxss` - 文档详情页面样式
  - `pages/document/detail.json` - 文档详情页面配置

### 4. 员工管理页面 (employee)
- 员工列表页面：
  - `pages/employee/list.js` - 员工列表逻辑
  - `pages/employee/list.wxml` - 员工列表页面结构
  - `pages/employee/list.wxss` - 员工列表页面样式
  - `pages/employee/list.json` - 员工列表页面配置
- 员工详情页面：
  - `pages/employee/detail.js` - 员工详情逻辑
  - `pages/employee/detail.wxml` - 员工详情页面结构
  - `pages/employee/detail.wxss` - 员工详情页面样式
  - `pages/employee/detail.json` - 员工详情页面配置

### 5. 应用中心页面 (app)
- 应用列表页面：
  - `pages/app/list.js` - 应用列表逻辑
  - `pages/app/list.wxml` - 应用列表页面结构
  - `pages/app/list.wxss` - 应用列表页面样式
  - `pages/app/list.json` - 应用列表页面配置
- 应用详情页面：
  - `pages/app/detail.js` - 应用详情逻辑
  - `pages/app/detail.wxml` - 应用详情页面结构
  - `pages/app/detail.wxss` - 应用详情页面样式
  - `pages/app/detail.json` - 应用详情页面配置
- 二维码生成页面：
  - `pages/app/qrcode.js` - 二维码生成逻辑
  - `pages/app/qrcode.wxml` - 二维码生成页面结构
  - `pages/app/qrcode.wxss` - 二维码生成页面样式
  - `pages/app/qrcode.json` - 二维码生成页面配置

### 6. 业务中心页面 (business)
- 业务模块列表页面：
  - `pages/business/list.js` - 业务模块列表逻辑
  - `pages/business/list.wxml` - 业务模块列表页面结构
  - `pages/business/list.wxss` - 业务模块列表页面样式
  - `pages/business/list.json` - 业务模块列表页面配置
- 业务模块详情页面：
  - `pages/business/detail.js` - 业务模块详情逻辑
  - `pages/business/detail.wxml` - 业务模块详情页面结构
  - `pages/business/detail.wxss` - 业务模块详情页面样式
  - `pages/business/detail.json` - 业务模块详情页面配置
- 电子合同页面：
  - `pages/business/contract.js` - 电子合同逻辑
  - `pages/business/contract.wxml` - 电子合同页面结构
  - `pages/business/contract.wxss` - 电子合同页面样式
  - `pages/business/contract.json` - 电子合同页面配置

## 技术实现

### 1. 技术栈
- 微信小程序原生开发
- 使用微信小程序框架和组件
- 采用模块化开发方式

### 2. 页面跳转
- 使用 `wx.navigateTo` 进行页面跳转
- 使用 `wx.switchTab` 进行 tabBar 页面切换
- 使用 `wx.redirectTo` 进行页面重定向

### 3. 数据请求
- 使用 `wx.request` 发起网络请求
- 统一在 `app.js` 中配置 API 基础地址
- 使用 Token 进行身份验证

### 4. 数据存储
- 使用 `wx.setStorageSync` 和 `wx.getStorageSync` 进行本地数据存储
- 使用 `wx.setStorage` 和 `wx.getStorage` 进行异步数据存储

### 5. UI 组件
- 使用微信小程序原生组件
- 自定义样式组件
- 响应式布局设计

## 待完成功能

### 1. 表单设计器页面
- 创建表单设计器页面
- 实现拖拽式表单设计功能
- 支持多种表单控件

### 2. 数据收集页面
- 创建数据收集页面
- 实现数据录入功能
- 支持数据导出功能

### 3. 调查问卷页面
- 创建调查问卷页面
- 实现问卷设计和发布功能
- 支持问卷填写和统计

### 4. 移动端与后端API集成
- 完善所有页面与后端API的对接
- 实现完整的数据交互功能
- 添加错误处理和加载状态

### 5. 微信登录集成
- 完善微信登录功能
- 实现微信授权登录流程
- 处理微信登录回调

### 6. 性能优化
- 优化页面加载速度
- 减少网络请求次数
- 实现图片懒加载

### 7. 测试和调试
- 进行功能测试
- 进行兼容性测试
- 进行性能测试

## 部署说明

### 1. 开发环境
- 安装微信开发者工具
- 配置小程序 AppID
- 启动本地开发服务器

### 2. 生产环境
- 使用微信开发者工具上传代码
- 在微信公众平台提交审核
- 审核通过后发布上线

## 注意事项

1. 所有页面都已实现基础功能，但与后端API的对接需要进一步完善
2. 部分功能如二维码生成、表单设计等需要调用后端服务
3. 需要根据实际的API接口调整数据请求部分
4. 需要根据实际需求调整UI设计和交互逻辑