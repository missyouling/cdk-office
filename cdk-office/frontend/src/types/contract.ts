// 电子合同功能的 TypeScript 类型定义

export type ContractStatus = 
  | 'draft'        // 草稿
  | 'pending'      // 待发送
  | 'signing'      // 签署中
  | 'completed'    // 已完成
  | 'rejected'     // 已拒绝
  | 'cancelled'    // 已取消
  | 'expired';     // 已过期

export type SignerType = 'person' | 'company';

export type SignerStatus = 'pending' | 'signed' | 'rejected';

export interface ContractSigner {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  signerType: SignerType;
  status: SignerStatus;
  signTime?: string;
  rejectReason?: string;
  position?: string;          // 职位（个人签署者）
  companyName?: string;       // 公司名称（公司签署者）
  signatureImage?: string;    // 签名图片
  certificateId?: string;     // CA证书ID
}

export interface Contract {
  id: string;
  title: string;
  description?: string;
  status: ContractStatus;
  progress: number;           // 签署进度百分比
  templateId?: string;        // 合同模板ID
  content?: string;           // 合同内容
  fileUrl?: string;           // 合同文件URL
  
  // 时间信息
  createdAt: string;
  updatedAt: string;
  expireTime: string;         // 过期时间
  completedAt?: string;       // 完成时间
  
  // 创建者信息
  createdBy: string;
  createdByName: string;
  teamId: string;
  
  // 签署者信息
  signers: ContractSigner[];
  currentSignerIndex: number; // 当前应签署人索引
  
  // 额外信息
  tags?: string[];            // 标签
  priority?: 'low' | 'medium' | 'high'; // 优先级
  reminderEnabled?: boolean;  // 是否启用提醒
  blockchainHash?: string;    // 区块链存证哈希
}

export interface ContractTemplate {
  id: string;
  name: string;
  description?: string;
  category: string;
  content: string;
  variables: ContractVariable[]; // 模板变量
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

export interface ContractVariable {
  key: string;
  label: string;
  type: 'text' | 'number' | 'date' | 'select';
  required: boolean;
  defaultValue?: string;
  options?: string[];         // select类型的选项
}

export interface ContractHistory {
  id: string;
  contractId: string;
  action: string;            // 操作类型
  description: string;       // 操作描述
  operatorId: string;
  operatorName: string;
  timestamp: string;
  details?: Record<string, any>; // 操作详情
}

// API 请求响应类型
export interface CreateContractRequest {
  title: string;
  description?: string;
  templateId?: string;
  content?: string;
  expireTime: string;
  signers: Omit<ContractSigner, 'id' | 'status' | 'signTime'>[];
  tags?: string[];
  priority?: 'low' | 'medium' | 'high';
  reminderEnabled?: boolean;
}

export interface UpdateContractRequest {
  title?: string;
  description?: string;
  content?: string;
  expireTime?: string;
  tags?: string[];
  priority?: 'low' | 'medium' | 'high';
  reminderEnabled?: boolean;
}

export interface ContractListResponse {
  data: Contract[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ContractResponse {
  data: Contract;
}

export interface ContractStatsResponse {
  total: number;
  draft: number;
  pending: number;
  signing: number;
  completed: number;
  rejected: number;
  cancelled: number;
  expired: number;
}

// 表格列配置
export interface ContractTableColumn {
  key: keyof Contract | 'actions';
  title: string;
  width?: number;
  sortable?: boolean;
  filterable?: boolean;
}

// 筛选器配置
export interface ContractFilter {
  status?: ContractStatus[];
  priority?: ('low' | 'medium' | 'high')[];
  createdBy?: string[];
  dateRange?: {
    start: string;
    end: string;
  };
  tags?: string[];
}

// 批量操作类型
export type BatchAction = 'delete' | 'send' | 'cancel' | 'remind';

export interface BatchOperationRequest {
  contractIds: string[];
  action: BatchAction;
  data?: Record<string, any>;
}

// 签署相关类型
export interface SignContractRequest {
  contractId: string;
  signerId: string;
  signatureImage?: string;
  verificationCode?: string; // 短信验证码
  certificateId?: string;    // CA证书ID
}

export interface RejectContractRequest {
  contractId: string;
  signerId: string;
  reason: string;
}

// 电子签名相关
export interface DigitalSignature {
  id: string;
  contractId: string;
  signerId: string;
  signatureData: string;     // 签名数据
  timestamp: string;
  certificateId?: string;    // CA证书ID
  verified: boolean;         // 是否已验证
  algorithm: string;         // 签名算法
  hashValue: string;         // 文档哈希值
}

// CA证书相关
export interface CACertificate {
  id: string;
  userId: string;
  certificateData: string;   // 证书数据
  issuer: string;            // 发行者
  subject: string;           // 主体
  serialNumber: string;      // 序列号
  validFrom: string;         // 有效期开始
  validTo: string;           // 有效期结束
  revoked: boolean;          // 是否已撤销
  createdAt: string;
}

// 工作流相关（如果需要审批流程）
export interface ContractApprovalFlow {
  id: string;
  contractId: string;
  approvers: ContractApprover[];
  currentApproverIndex: number;
  status: 'pending' | 'approved' | 'rejected';
  createdAt: string;
  completedAt?: string;
}

export interface ContractApprover {
  id: string;
  userId: string;
  userName: string;
  status: 'pending' | 'approved' | 'rejected';
  comments?: string;
  approvedAt?: string;
  order: number;            // 审批顺序
}

// 通知设置
export interface ContractNotification {
  id: string;
  contractId: string;
  type: 'reminder' | 'status_change' | 'expiry_warning';
  recipients: string[];     // 接收者用户ID列表
  message: string;
  scheduledAt?: string;     // 定时发送时间
  sentAt?: string;          // 实际发送时间
  status: 'pending' | 'sent' | 'failed';
}

// 应用中心卡片类型
export interface AppCenterCard {
  id: string;
  title: string;
  description: string;
  icon: any; // React.ReactNode
  link: string;
  badge?: string;
  badgeVariant?: 'default' | 'secondary' | 'destructive' | 'outline' | 'success';
  category: 'core' | 'ai' | 'business' | 'tools';
  featured?: boolean;       // 是否为推荐应用
  status?: 'active' | 'beta' | 'coming_soon';
}