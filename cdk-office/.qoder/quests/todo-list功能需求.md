好的，这是一个专门为在前端仪表板（Dashboard）页面新增用户待办事项和日程提醒功能而设计的、详细的Qoder AI提示词。它将引导AI从设计到实现，完成整个功能的开发。

---

### **Qoder AI 提示词：Dashboard待办事项与日程功能开发**

**角色：** 你是一名资深全栈开发工程师，精通Next.js 15, React 19, TypeScript, Tailwind CSS和Shadcn UI。

**项目背景：** 你正在开发CDK-Office平台的Dashboard页面。需要新增两个核心功能：
1.  **用户待办事项（To-Do List）**：一个允许用户创建、编辑、删除和标记完成个人任务的功能。
2.  **日程提醒（Calendar Reminder）**：一个简单的日程视图和提醒功能，允许用户添加和查看特定日期的日程。

**技术栈要求：**
*   **前端框架:** Next.js 15 (App Router)
*   **UI 组件库:** Shadcn UI (基于Tailwind CSS)
*   **状态管理:** 优先使用 `useState`/`useReducer`，如需跨组件共享可使用 `Zustand`
*   **图标:** Lucide React (`npm install lucide-react`)
*   **日期处理:** date-fns (`npm install date-fns`)
*   **HTTP客户端:** 使用项目中现有的 `api` 工具库（如基于axios或fetch的封装）

---

### **第一步：设计数据模型与API接口**

**提示词：**
“你是一名后端开发工程师。请为‘用户待办事项’和‘日程提醒’功能设计RESTful API和数据模型。

1.  **数据模型设计 (PostgreSQL):** 请提供两个表的SQL定义 (`todo_items` 和 `calendar_events`)。每个表必须包含 `id`, `user_id`, `team_id`（用于数据隔离），以及以下字段：
    *   `todo_items`: `title` (字符串), `completed` (布尔值，默认false), `due_date` (可选的时间戳，用于设置任务截止日), `created_at`, `updated_at`。
    *   `calendar_events`: `title` (字符串), `description` (文本，可选), `start_time` (时间戳), `end_time` (时间戳), `all_day` (布尔值), `created_at`。

2.  **API接口设计:** 为每个模型设计完整的CRUD API端点。
    *   `GET /api/todos` - 获取当前用户的待办列表（支持 `?completed=true/false` 过滤）
    *   `POST /api/todos` - 创建新的待办
    *   `PATCH /api/todos/:id` - 更新待办（如标记完成/未完成）
    *   `DELETE /api/todos/:id` - 删除待办
    *   `GET /api/calendar-events` - 获取用户的日程（支持 `?startDate=...&endDate=...` 查询时间范围）
    *   `POST /api/calendar-events` - 创建新日程
    *   （可选）为 `calendar_events` 实现PATCH和DELETE接口。

3.  **认证与授权:** 确保每个API端点都通过中间件检查，用户只能操作自己所属 `team_id` 下的数据。

请输出Go语言的结构体定义、Gin路由框架代码和SQL迁移脚本。”

---

### **第二步：构建前端UI组件**

**提示词：**
“你是一名前端开发工程师。请在Dashboard页面上创建两个新的UI组件：`<TodoCard />` 和 `<CalendarCard />`。请使用Shadcn UI组件进行构建。

1.  **`<TodoCard />` 组件要求:**
    *   一个清晰的标题，如“待办事项”。
    *   一个输入框和“添加”按钮，用于快速创建新任务。
    *   一个任务列表，每个任务项前有一个复选框，用于标记完成/未完成。已完成的任务应有删除线样式（`line-through text-gray-400`）。
    *   点击复选框即调用API更新状态（乐观更新）。
    *   每个任务项右侧有一个删除图标按钮（使用Lucide的 `Trash2` 图标），点击后删除该任务。
    *   支持显示可选的截止日期（`due_date`），如果已过期，则用红色文字显示。

2.  **`<CalendarCard />` 组件要求:**
    *   一个清晰的标题，如“近期日程”。
    *   显示未来7天的日程事件列表。
    *   每个日程项显示：时间（或“全天”）、标题。
    *   提供一个“+”按钮，点击后弹出一个Shadcn UI的对话框（`<Dialog />`），里面是一个表单，用于创建新的日程事件（包含标题、描述、开始时间、结束时间、是否全天等字段）。

3.  **布局集成:** 将这两个组件优雅地集成到现有的Dashboard网格布局中。考虑使用Shadcn的 `<Card />` 组件来包装它们。

请提供完整的TSX组件代码、必要的TypeScript类型定义（`interface TodoItem {...}`）和API调用函数。”

---

### **第三步：实现数据获取与状态同步**

**提示词：**
“现在，请为上述两个组件实现高效的数据获取和状态管理。

1.  **数据获取:** 使用 `fetch` 或项目现有的 `api` 库，在组件挂载时（`useEffect` 或 Server Component中）从 `/api/todos` 和 `/api/calendar-events` 获取数据。

2.  **状态管理:** 使用 `useState` 来管理待办和日程的列表数据。

3.  **乐观更新:** 为待办事项的“完成/未完成”切换实现乐观更新：
    *   用户点击复选框时，立即更新UI状态（显示为完成状态）。
    *   同时，在后台发起 `PATCH /api/todos/:id` 请求。
    *   如果请求失败，回滚UI状态并显示错误提示（使用Shadcn UI的 `toast`）。

4.  **错误处理:** 为所有API调用添加基本的错误处理，例如使用 `try/catch` 并提示用户。

请提供集成后的完整组件代码，展示数据流和状态是如何管理的。”

---

### **第四步：集成后台提醒功能（可选增强）**

**提示词：**
“作为可选增强功能，请将前端与现有的 `gocron` 后端集成，实现定时提醒。

1.  **后端修改 (Go):** 编写一个后台任务，由 `gocron` 每隔15分钟触发一次。这个任务的功能是：查询 `calendar_events` 表，找出从现在开始15分钟内即将开始的日程，并向对应的用户发送通知（通知方式可以是数据库记录、WebSocket消息或集成邮件服务，此处只需实现数据库记录）。

2.  **前端轮询/通知 (Next.js):** 在前端，使用 `useEffect` 和 `setInterval` 每隔一段时间（如60秒）轮询一次新的提醒（例如调用 `GET /api/notifications`）。如果有新的即将到来的日程，使用Shadcn UI的 `toast` 组件或一个通知铃图标来提醒用户。

请提供后端Go定时任务代码的框架和前端的轮询逻辑代码。”

---

**使用指南：**
1.  **按顺序执行：** 将这四个提示词按顺序提供给Qoder AI。
2.  **审查代码：** 仔细审查每一步生成的代码，确保其符合你的项目结构和编码规范。
3.  **迭代：** 如果某个部分不完美，可以要求AI进行修正（例如：“为这个Todo组件添加加载状态”、“优化日程表单的日期时间选择器”）。
4.  **测试：** 实现后，务必进行手动测试，创建、编辑、删除待办和日程，确保所有功能正常工作。

这套提示词将系统性地引导AI为你创建一个功能完整、UI美观、且与现有技术栈完美融合的待办事项和日程提醒模块。