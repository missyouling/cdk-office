# AI智能问答功能使用指南

## 快速开始

### 1. 确保依赖已安装

```bash
# 在frontend目录下
npm install
# 或者
yarn install
```

### 2. 启动开发服务器

```bash
npm run dev
# 或者
yarn dev
```

### 3. 访问功能

1. 打开浏览器访问 `http://localhost:3000`
2. 进入"应用中心"页面
3. 点击"智能问答"卡片
4. 开始使用AI智能问答功能

## 功能测试

### 测试步骤

1. **基础对话测试**
   - 输入简单问题："什么是CDK-Office？"
   - 观察AI回答和界面响应

2. **预设问题测试**
   - 点击预设问题按钮
   - 验证问题自动填入输入框

3. **加载状态测试**
   - 发送问题后观察加载指示器
   - 确认加载期间输入框被禁用

4. **错误处理测试**
   - 断开网络连接后发送问题
   - 观察错误提示显示

5. **UI响应测试**
   - 测试不同屏幕尺寸下的界面适配
   - 验证自动滚动功能

### 测试用例

#### 正常流程
```
用户输入：什么是CDK-Office？
期望结果：
- 显示用户消息气泡
- 显示加载指示器
- 显示AI回答
- 显示置信度（如果有）
- 显示来源文档（如果有）
```

#### 错误流程
```
网络错误场景：
期望结果：
- 显示错误提示
- 显示友好的错误AI消息
- 用户可以重新尝试
```

## 自定义配置

### API端点配置

如果需要修改API端点，编辑 `src/lib/api/ai-chat.ts`：

```typescript
// 修改API基础URL
const API_BASE_URL = '/your-api-base';
```

### UI主题定制

如果需要自定义界面样式，编辑 `src/components/ai/AIChatInterface.tsx`：

```typescript
// 修改颜色方案
const userMessageStyle = "bg-blue-500 text-white";
const aiMessageStyle = "bg-gray-100 text-gray-900";
```

### 预设问题配置

修改预设问题列表：

```typescript
const suggestions = [
  '你的自定义问题1',
  '你的自定义问题2',
  // ...
];
```

## 常见问题

### Q: API调用失败怎么办？
A: 检查以下几点：
1. 后端服务是否正常运行
2. API端点配置是否正确
3. 认证信息是否配置正确
4. 网络连接是否正常

### Q: 界面显示异常怎么办？
A: 确认：
1. Shadcn UI组件是否正确安装
2. CSS样式是否正确加载
3. 浏览器是否支持相关特性

### Q: 如何添加新功能？
A: 可以：
1. 在`useAIChat.ts`中添加新的状态管理
2. 在`AIChatInterface.tsx`中添加新的UI组件
3. 在`ai-chat.ts`中添加新的API调用

## 性能优化建议

1. **懒加载** - 考虑对聊天历史进行分页加载
2. **防抖** - 为输入添加防抖处理避免频繁API调用
3. **缓存** - 对常见问题的回答进行客户端缓存
4. **虚拟滚动** - 长对话列表使用虚拟滚动优化性能

## 部署注意事项

### 生产环境配置

1. **API配置**
   ```typescript
   const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api';
   ```

2. **错误监控**
   - 集成错误监控服务（如Sentry）
   - 添加用户行为分析

3. **安全考虑**
   - 确保API认证配置正确
   - 添加输入内容过滤
   - 防范XSS攻击

### Docker部署

如果使用Docker部署，确保环境变量正确配置：

```dockerfile
ENV NEXT_PUBLIC_API_URL=your-api-url
```

## 贡献指南

如果你想为这个功能做贡献：

1. Fork项目仓库
2. 创建功能分支
3. 提交你的更改
4. 创建Pull Request

请确保：
- 代码符合项目的代码规范
- 添加适当的类型定义
- 包含必要的测试
- 更新相关文档