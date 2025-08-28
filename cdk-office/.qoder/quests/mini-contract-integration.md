# Mini-Contract 电子合同集成设计

## Overview

基于CDK-Office企业内容管理平台现有架构，集成Mini-Contract v2.1.0电子合同系统。通过Docker Compose编排部署，在应用中心添加电子合同卡片，实现完整的电子签署服务。

## Technology Stack & Dependencies

### 集成技术栈

| 组件 | 技术 | 版本 | 用途 |
|------|------|------|------|
| **CDK-Office前端** | Next.js + React | 15 + 19 | 应用中心界面 |
| **CDK-Office后端** | Go + Gin | 1.24+ | API网关代理 |
| **Mini-Contract** | Vue 2.0 + uni-app | v2.1.0 | 电子合同服务 |
| **统一数据库** | PostgreSQL | 13+ | 共享数据存储 |
| **统一缓存** | Redis | 6+ | 共享会话管理 |
| **统一文件存储** | CDK-Office FileService | - | 共享文件管理 |
| **权限** | Casbin | - | 统一访问控制 |
| **知识库集成** | Dify Platform | - | 文档审核与管理 |

### 外部服务

| 服务 | 提供商 | 功能 |
|------|--------|------|
| **CA证书** | 法大大/e签宝/君子签 | 数字签名 |
| **区块链存证** | 蚂蚁司法链 | 司法存证 |
| **短信通知** | 阿里云/腾讯云 | 签署通知 |
| **实名认证** | 第三方服务 | 身份验证 |

## Architecture

### 整体系统架构

```mermaid
graph TB
    subgraph "CDK-Office统一平台"
        CDKWeb[CDK-Office Web<br/>Next.js + React]
        AppCenter[应用中心]
        ContractCard[电子合同卡片]
        CDKGateway[API网关<br/>Go + Gin]
        AuthMiddleware[Casbin权限]
        KnowledgeBase[知识库<br/>Dify Platform]
        ApprovalFlow[审批流程<br/>go-workflows]
        
        subgraph "统一基础设施"
            SharedDB[PostgreSQL<br/>统一数据库]
            SharedRedis[Redis<br/>统一缓存]
            SharedFileService[CDK-Office FileService<br/>统一文件存储]
        end
    end
    
    subgraph "Mini-Contract集成模块"
        ContractWeb[合同Web端<br/>Vue.js]
        ContractService[合同服务<br/>Go服务]
        ContractEngine[签署引擎]
        TemplateEngine[模板引擎]
    end
    
    subgraph "第三方服务"
        CAService[CA证书服务]
        SMSService[短信服务]
        BlockchainService[区块链存证]
    end
    
    CDKWeb --> AppCenter
    AppCenter --> ContractCard
    ContractCard --> CDKGateway
    CDKGateway --> AuthMiddleware
    AuthMiddleware --> ContractService
    
    ContractService --> SharedDB
    ContractService --> SharedRedis
    ContractService --> SharedFileService
    ContractService --> ContractEngine
    ContractEngine --> TemplateEngine
    
    ContractService --> CAService
    ContractService --> SMSService
    ContractService --> BlockchainService
    
    ContractService --> ApprovalFlow
    ApprovalFlow --> KnowledgeBase
    KnowledgeBase --> SharedDB
```

### Docker Compose架构

```mermaid
graph TB
    subgraph "Docker统一服务编排"
        subgraph "CDK-Office核心组"
            CDKOfficeAPI[cdk-office-api<br/>集成合同服务]
            CDKOfficeFrontend[cdk-office-frontend<br/>包含合同界面]
            SharedRedis[shared-redis<br/>统一缓存]
            SharedDB[shared-postgres<br/>统一数据库]
            DifyService[dify-platform<br/>知识库服务]
        end
        
        subgraph "工具服务组"
            StirlingPDF[stirling-pdf<br/>PDF处理]
            NginxProxy[nginx<br/>统一代理]
        end
    end
    
    CDKOfficeAPI -.-> SharedDB
    CDKOfficeAPI -.-> SharedRedis
    CDKOfficeAPI -.-> DifyService
    
    NginxProxy -.-> CDKOfficeAPI
    NginxProxy -.-> CDKOfficeFrontend
    NginxProxy -.-> StirlingPDF
```

## Business Logic Layer (Architecture of each feature)

### 合同签署完成后的知识库集成流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant ContractUI as 合同界面
    participant ContractService as 合同服务
    participant WorkflowEngine as 工作流引擎
    participant ApprovalService as 审批服务
    participant KnowledgeBase as 知识库
    participant NotifyService as 通知服务
    participant DifyAPI as Dify API
    
    User->>ContractUI: 完成最后一方签署
    ContractUI->>ContractService: 提交签署完成
    ContractService->>ContractService: 更新合同状态为已完成
    ContractService->>WorkflowEngine: 触发合同完成事件
    
    WorkflowEngine->>WorkflowEngine: 检查工作流配置
    alt 自动提交到知识库
        WorkflowEngine->>KnowledgeBase: 创建文档记录
        KnowledgeBase->>DifyAPI: 调用文档处理API
        DifyAPI-->>KnowledgeBase: 返回处理结果
        
        alt 需要审批
            WorkflowEngine->>ApprovalService: 创建审批任务
            ApprovalService->>NotifyService: 发送审批通知
            NotifyService-->>User: 通知相关审批人
        else 自动通过
            WorkflowEngine->>KnowledgeBase: 直接发布到知识库
            KnowledgeBase->>NotifyService: 发送完成通知
        end
    end
    
    ContractService->>NotifyService: 发送签署完成通知
    NotifyService-->>User: 通知所有签署方
```

### 合同服务核心业务逻辑

```go
// ContractService 合同核心服务
type ContractService struct {
    db              *gorm.DB
    redis           *redis.Client
    fileService     *FileService    // 使用CDK-Office文件服务
    workflowEngine  *WorkflowEngine
    knowledgeBase   *KnowledgeBaseService
    notifyService   *NotificationService
    difyClient      *DifyClient
}

// CompleteContract 完成合同签署
func (s *ContractService) CompleteContract(contractID string, signerID string, signData SignData) error {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // 1. 更新签署状态
    if err := s.updateSignerStatus(tx, contractID, signerID, signData); err != nil {
        tx.Rollback()
        return err
    }
    
    // 2. 检查是否所有人都已签署
    allSigned, err := s.checkAllSigned(tx, contractID)
    if err != nil {
        tx.Rollback()
        return err
    }
    
    if allSigned {
        // 3. 更新合同状态为已完成
        if err := s.updateContractStatus(tx, contractID, "completed"); err != nil {
            tx.Rollback()
            return err
        }
        
        // 4. 生成最终合同文件
        finalFileURL, err := s.generateFinalContract(contractID)
        if err != nil {
            tx.Rollback()
            return err
        }
        
        // 5. 使用CDK-Office文件服务保存文件
        fileID, err := s.fileService.SaveContractFile(contractID, finalFileURL)
        if err != nil {
            tx.Rollback()
            return err
        }
        
        // 6. 区块链存证
        evidenceURL, err := s.createBlockchainEvidence(contractID, finalFileURL)
        if err != nil {
            tx.Rollback()
            return err
        }
        
        tx.Commit()
        
        // 7. 异步处理：触发工作流
        go s.triggerContractCompletedWorkflow(contractID, finalFileURL, evidenceURL)
    } else {
        tx.Commit()
        // 通知下一个签署人
        go s.notifyNextSigner(contractID)
    }
    
    return nil
}

// triggerContractCompletedWorkflow 触发合同完成工作流
func (s *ContractService) triggerContractCompletedWorkflow(contractID, fileURL, evidenceURL string) {
    contract, err := s.GetContract(contractID)
    if err != nil {
        log.Errorf("获取合同信息失败: %v", err)
        return
    }
    
    // 获取团队工作流配置
    workflow, err := s.getTeamContractWorkflow(contract.TeamID)
    if err != nil {
        log.Errorf("获取工作流配置失败: %v", err)
        return
    }
    
    if workflow.AutoSubmitKnowledge {
        // 自动提交到知识库
        s.submitToKnowledgeBase(contract, fileURL, evidenceURL, workflow)
    }
    
    // 发送完成通知
    s.sendCompletionNotifications(contract)
}

// submitToKnowledgeBase 提交到知识库
func (s *ContractService) submitToKnowledgeBase(contract *Contract, fileURL, evidenceURL string, workflow *ContractWorkflow) {
    // 1. 创建知识库文档记录
    document := &Document{
        ID:          uuid.New().String(),
        Title:       fmt.Sprintf("合同：%s", contract.Title),
        Content:     s.generateDocumentContent(contract),
        FileURL:     fileURL,
        FileType:    "contract",
        TeamID:      contract.TeamID,
        CreatorID:   contract.CreatorID,
        Status:      "pending_review",
        Visibility:  "team",
        Tags:        []string{"合同", "已签署", contract.Status},
        Metadata: map[string]interface{}{
            "contract_id":      contract.ID,
            "signers_count":    len(contract.Signers),
            "evidence_url":     evidenceURL,
            "blockchain_hash":  contract.FileHash,
            "completed_at":     contract.CompletedAt,
        },
    }
    
    // 2. 保存文档到数据库
    if err := s.db.Create(document).Error; err != nil {
        log.Errorf("创建知识库文档失败: %v", err)
        return
    }
    
    // 3. 创建知识库提交记录
    submission := &KnowledgeSubmission{
        ID:                 uuid.New().String(),
        ContractID:         contract.ID,
        DocumentID:         document.ID,
        SubmissionType:     "auto",
        Status:             "pending",
        AutoProcessing:     true,
        DifyWorkflowStatus: "pending",
    }
    
    if err := s.db.Create(submission).Error; err != nil {
        log.Errorf("创建知识库提交记录失败: %v", err)
        return
    }
    
    // 4. 判断是否需要审批
    if workflow.RequireApproval {
        s.createApprovalTask(contract, document, workflow)
    } else {
        // 直接处理
        s.processKnowledgeSubmission(submission)
    }
}

// processKnowledgeSubmission 处理知识库提交
func (s *ContractService) processKnowledgeSubmission(submission *KnowledgeSubmission) {
    // 1. 调用Dify API处理文档
    if submission.AutoProcessing {
        go s.processByDify(submission)
    }
    
    // 2. 更新文档状态
    s.db.Model(&Document{}).Where("id = ?", submission.DocumentID).Updates(map[string]interface{}{
        "status":     "published",
        "updated_at": time.Now(),
    })
    
    // 3. 更新提交状态
    s.db.Model(submission).Updates(map[string]interface{}{
        "status":     "completed",
        "updated_at": time.Now(),
    })
    
    // 4. 发送通知
    s.notifyKnowledgeBaseUpdate(submission)
}

// processByDify 通过Dify处理文档
func (s *ContractService) processByDify(submission *KnowledgeSubmission) {
    document, err := s.getDocument(submission.DocumentID)
    if err != nil {
        log.Errorf("获取文档失败: %v", err)
        return
    }
    
    // 调用Dify API进行文档向量化和知识抽取
    result, err := s.difyClient.ProcessDocument(DifyProcessRequest{
        DocumentID:   document.ID,
        Title:        document.Title,
        Content:      document.Content,
        FileURL:      document.FileURL,
        DocumentType: "contract",
        Metadata:     document.Metadata,
    })
    
    if err != nil {
        log.Errorf("Dify处理文档失败: %v", err)
        s.db.Model(submission).Updates(map[string]interface{}{
            "dify_workflow_status": "failed",
            "updated_at":           time.Now(),
        })
        return
    }
    
    // 更新处理结果
    s.db.Model(submission).Updates(map[string]interface{}{
        "dify_workflow_status": "completed",
        "updated_at":           time.Now(),
    })
    
    // 更新文档的AI处理结果
    s.db.Model(document).Updates(map[string]interface{}{
        "ai_processed":  true,
        "ai_summary":    result.Summary,
        "ai_keywords":   result.Keywords,
        "vector_status": "completed",
        "updated_at":    time.Now(),
    })
}
```

### 审批流程集成

```go
// createApprovalTask 创建审批任务
func (s *ContractService) createApprovalTask(contract *Contract, document *Document, workflow *ContractWorkflow) {
    // 1. 创建工作流实例
    workflowInstance := &WorkflowInstance{
        ID:            uuid.New().String(),
        WorkflowDefID: workflow.ID,
        Status:        "running",
        InputData:     s.serializeWorkflowInput(contract, document),
        CreatedBy:     contract.CreatorID,
    }
    
    s.db.Create(workflowInstance)
    
    // 2. 创建审批任务
    for _, role := range workflow.ApprovalRoles {
        approvers := s.getApproversByRole(contract.TeamID, role)
        for _, approver := range approvers {
            approvalTask := &ApprovalTask{
                ID:             uuid.New().String(),
                WorkflowInstID: workflowInstance.ID,
                Name:           fmt.Sprintf("合同知识库审批：%s", contract.Title),
                Description:    fmt.Sprintf("合同《%s》已完成签署，需要审批是否加入知识库", contract.Title),
                Assignee:       approver.ID,
                Status:         "pending",
                DueDate:        time.Now().Add(24 * time.Hour), // 24小时内完成
            }
            
            s.db.Create(approvalTask)
            
            // 发送审批通知
            s.notifyService.SendApprovalNotification(approver, approvalTask, contract)
        }
    }
}

// ApproveKnowledgeSubmission 审批知识库提交
func (s *ContractService) ApproveKnowledgeSubmission(taskID, approverID, comments string, approved bool) error {
    task := &ApprovalTask{}
    if err := s.db.Where("id = ? AND assignee = ?", taskID, approverID).First(task).Error; err != nil {
        return fmt.Errorf("审批任务不存在或无权限")
    }
    
    // 更新审批任务状态
    status := "rejected"
    if approved {
        status = "approved"
    }
    
    s.db.Model(task).Updates(map[string]interface{}{
        "status":    status,
        "comments":  comments,
        "updated_at": time.Now(),
    })
    
    // 检查所有审批任务是否完成
    if s.checkAllApprovalsCompleted(task.WorkflowInstID) {
        if approved {
            // 所有审批通过，处理知识库提交
            submission := s.getSubmissionByWorkflow(task.WorkflowInstID)
            s.processKnowledgeSubmission(submission)
        } else {
            // 审批被拒绝
            s.rejectKnowledgeSubmission(task.WorkflowInstID, comments)
        }
    }
    
    return nil
}
```

### CDK-Office应用中心集成

#### 前端组件设计

```typescript
// components/app-center/ContractCard.tsx
import React from 'react';
import { Card, CardContent, Typography, Button } from '@mui/material';
import { Assignment } from '@mui/icons-material';

interface ContractCardProps {
  onLaunch: () => void;
}

const ContractCard: React.FC<ContractCardProps> = ({ onLaunch }) => {
  return (
    <Card className="h-full flex flex-col">
      <div className="h-40 bg-green-100 flex items-center justify-center">
        <Assignment className="text-green-500" style={{ fontSize: 80 }} />
      </div>
      <CardContent className="flex-grow flex flex-col">
        <Typography gutterBottom variant="h5" component="div">
          电子合同
        </Typography>
        <Typography variant="body2" color="text.secondary" className="flex-grow">
          强大的电子合同签署平台，支持多方签署、CA证书、区块链存证等功能
        </Typography>
        <Button 
          variant="contained" 
          color="primary" 
          onClick={onLaunch}
          className="mt-4"
        >
          立即使用
        </Button>
      </CardContent>
    </Card>
  );
};

// components/contract/ContractModal.tsx
const ContractModal: React.FC<{ isOpen: boolean; onClose: () => void }> = ({ 
  isOpen, 
  onClose 
}) => {
  const [contractUrl, setContractUrl] = useState<string>('');
  
  useEffect(() => {
    if (isOpen) {
      // 获取SSO登录URL
      fetchContractAccessUrl().then(setContractUrl);
    }
  }, [isOpen]);
  
  return (
    <Dialog open={isOpen} onClose={onClose} maxWidth="lg" fullWidth fullScreen>
      <div className="h-full flex flex-col">
        <div className="flex items-center justify-between p-4 border-b">
          <Typography variant="h6">电子合同系统</Typography>
          <IconButton onClick={onClose}>
            <Close />
          </IconButton>
        </div>
        <div className="flex-grow">
          {contractUrl && (
            <iframe 
              src={contractUrl} 
              width="100%" 
              height="100%" 
              frameBorder="0"
            />
          )}
        </div>
      </div>
    </Dialog>
  );
};
```

#### 后端API代理服务

```go
// internal/apps/contract/controller.go
package contract

import (
    "github.com/gin-gonic/gin"
    "github.com/linux-do/cdk-office/internal/auth"
)

type ContractController struct {
    contractService *ContractService
    authService     *auth.Service
}

// 代理合同API请求
func (c *ContractController) ProxyContractAPI(ctx *gin.Context) {
    // 验证用户权限
    user, err := c.authService.GetCurrentUser(ctx)
    if err != nil {
        ctx.JSON(401, gin.H{"error": "未授权"})
        return
    }
    
    // 检查合同权限
    if !c.checkContractPermission(user, ctx.Request.Method, ctx.Param("path")) {
        ctx.JSON(403, gin.H{"error": "权限不足"})
        return
    }
    
    // 代理请求到Mini-Contract
    err = c.contractService.ProxyRequest(ctx, user.TeamID)
    if err != nil {
        ctx.JSON(500, gin.H{"error": "代理请求失败"})
        return
    }
}

// 生成SSO登录URL
func (c *ContractController) GenerateAccessURL(ctx *gin.Context) {
    user, _ := c.authService.GetCurrentUser(ctx)
    
    accessURL, token, err := c.contractService.GenerateSSO(user)
    if err != nil {
        ctx.JSON(500, gin.H{"error": "生成访问URL失败"})
        return
    }
    
    ctx.JSON(200, gin.H{
        "access_url": accessURL,
        "token": token,
        "expires_at": time.Now().Add(time.Hour),
    })
}
```

## Component Architecture

### CDK-Office应用中心集成

#### 前端组件设计（直接集成到现有界面）

```typescript
// components/app-center/ContractCard.tsx
import React from 'react';
import { Card, CardContent, Typography, Button } from '@mui/material';
import { Assignment } from '@mui/icons-material';
import { useAuth } from '@/hooks/use-auth';
import { useRouter } from 'next/navigation';

const ContractCard: React.FC = () => {
  const { user } = useAuth();
  const router = useRouter();
  
  const handleLaunch = () => {
    // 直接跳转到合同管理页面，无需SSO
    router.push('/contracts');
  };
  
  // 检查用户是否有合同权限
  const hasContractPermission = user?.permissions?.includes('contract:access');
  
  if (!hasContractPermission) {
    return null; // 无权限时不显示卡片
  }
  
  return (
    <Card className="h-full flex flex-col">
      <div className="h-40 bg-green-100 flex items-center justify-center">
        <Assignment className="text-green-500" style={{ fontSize: 80 }} />
      </div>
      <CardContent className="flex-grow flex flex-col">
        <Typography gutterBottom variant="h5" component="div">
          电子合同
        </Typography>
        <Typography variant="body2" color="text.secondary" className="flex-grow">
          强大的电子合同签署平台，支持多方签署、CA证书、区块链存证，已签署合同自动归档到知识库
        </Typography>
        <Button 
          variant="contained" 
          color="primary" 
          onClick={handleLaunch}
          className="mt-4"
        >
          立即使用
        </Button>
      </CardContent>
    </Card>
  );
};

export default ContractCard;

// app/(main)/contracts/page.tsx - 合同管理主页面
import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Typography, 
  Button, 
  Table, 
  TableBody, 
  TableCell, 
  TableContainer, 
  TableHead, 
  TableRow,
  Paper,
  Chip,
  Dialog
} from '@mui/material';
import { Add, Visibility, Edit, Delete } from '@mui/icons-material';
import { useAuth } from '@/hooks/use-auth';
import { contractService } from '@/lib/services';

const ContractsPage: React.FC = () => {
  const { user } = useAuth();
  const [contracts, setContracts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  
  useEffect(() => {
    loadContracts();
  }, [user?.currentTeamId]);
  
  const loadContracts = async () => {
    try {
      setLoading(true);
      const response = await contractService.getContracts({
        team_id: user?.currentTeamId,
        page: 1,
        size: 20
      });
      setContracts(response.data);
    } catch (error) {
      console.error('加载合同列表失败:', error);
    } finally {
      setLoading(false);
    }
  };
  
  const getStatusColor = (status: string) => {
    const colors = {
      'draft': 'default',
      'signing': 'warning',
      'completed': 'success',
      'rejected': 'error',
      'expired': 'error'
    };
    return colors[status] || 'default';
  };
  
  const getStatusText = (status: string) => {
    const texts = {
      'draft': '草稿',
      'signing': '签署中',
      'completed': '已完成',
      'rejected': '已拒绝',
      'expired': '已过期'
    };
    return texts[status] || status;
  };
  
  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">电子合同管理</Typography>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={() => setShowCreateDialog(true)}
        >
          创建合同
        </Button>
      </Box>
      
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>合同标题</TableCell>
              <TableCell>状态</TableCell>
              <TableCell>创建时间</TableCell>
              <TableCell>签署方数量</TableCell>
              <TableCell>知识库状态</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {contracts.map((contract) => (
              <TableRow key={contract.id}>
                <TableCell>{contract.title}</TableCell>
                <TableCell>
                  <Chip 
                    label={getStatusText(contract.status)} 
                    color={getStatusColor(contract.status)}
                    size="small"
                  />
                </TableCell>
                <TableCell>{new Date(contract.created_at).toLocaleDateString()}</TableCell>
                <TableCell>{contract.signers?.length || 0}</TableCell>
                <TableCell>
                  {contract.document_id ? (
                    <Chip label="已归档" color="success" size="small" />
                  ) : (
                    <Chip label="未归档" color="default" size="small" />
                  )}
                </TableCell>
                <TableCell>
                  <Button size="small" startIcon={<Visibility />}>
                    查看
                  </Button>
                  {contract.status === 'draft' && (
                    <Button size="small" startIcon={<Edit />}>
                      编辑
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      
      {/* 创建合同对话框 */}
      <CreateContractDialog 
        open={showCreateDialog}
        onClose={() => setShowCreateDialog(false)}
        onSuccess={loadContracts}
      />
    </Box>
  );
};

export default ContractsPage;
```

## Data Models & ORM Mapping

### CDK-Office文件服务集成

```go
// FileService CDK-Office文件服务
type FileService struct {
    config      *config.Config
    uploadPath  string
    baseURL     string
}

// NewFileService 创建文件服务实例
func NewFileService(cfg *config.Config) *FileService {
    return &FileService{
        config:     cfg,
        uploadPath: cfg.FileStorage.Path,
        baseURL:    cfg.FileStorage.BaseURL,
    }
}

// SaveContractFile 保存合同文件
func (f *FileService) SaveContractFile(contractID string, fileData []byte) (string, error) {
    // 1. 生成文件名
    fileName := fmt.Sprintf("contract_%s_%d.pdf", contractID, time.Now().Unix())
    
    // 2. 创建目录
    contractDir := filepath.Join(f.uploadPath, "contracts", contractID)
    if err := os.MkdirAll(contractDir, 0755); err != nil {
        return "", fmt.Errorf("创建目录失败: %v", err)
    }
    
    // 3. 保存文件
    filePath := filepath.Join(contractDir, fileName)
    if err := ioutil.WriteFile(filePath, fileData, 0644); err != nil {
        return "", fmt.Errorf("保存文件失败: %v", err)
    }
    
    // 4. 返回访问URL
    fileURL := fmt.Sprintf("%s/uploads/contracts/%s/%s", f.baseURL, contractID, fileName)
    return fileURL, nil
}

// GetContractFile 获取合同文件
func (f *FileService) GetContractFile(contractID, fileName string) ([]byte, error) {
    filePath := filepath.Join(f.uploadPath, "contracts", contractID, fileName)
    return ioutil.ReadFile(filePath)
}

// DeleteContractFile 删除合同文件
func (f *FileService) DeleteContractFile(contractID, fileName string) error {
    filePath := filepath.Join(f.uploadPath, "contracts", contractID, fileName)
    return os.Remove(filePath)
}

// UploadFile 通用文件上传
func (f *FileService) UploadFile(file multipart.File, header *multipart.FileHeader, category string) (string, error) {
    // 1. 验证文件类型
    if err := f.validateFileType(header.Filename); err != nil {
        return "", err
    }
    
    // 2. 生成唯一文件名
    ext := filepath.Ext(header.Filename)
    fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
    
    // 3. 创建目录
    uploadDir := filepath.Join(f.uploadPath, category)
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        return "", fmt.Errorf("创建目录失败: %v", err)
    }
    
    // 4. 保存文件
    filePath := filepath.Join(uploadDir, fileName)
    dst, err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("创建文件失败: %v", err)
    }
    defer dst.Close()
    
    if _, err := io.Copy(dst, file); err != nil {
        return "", fmt.Errorf("保存文件失败: %v", err)
    }
    
    // 5. 返回访问URL
    fileURL := fmt.Sprintf("%s/uploads/%s/%s", f.baseURL, category, fileName)
    return fileURL, nil
}

// validateFileType 验证文件类型
func (f *FileService) validateFileType(filename string) error {
    ext := strings.ToLower(filepath.Ext(filename))
    allowedTypes := []string{".pdf", ".doc", ".docx", ".jpg", ".jpeg", ".png"}
    
    for _, allowedExt := range allowedTypes {
        if ext == allowedExt {
            return nil
        }
    }
    
    return fmt.Errorf("不支持的文件类型: %s", ext)
}
```

### 文件配置模型

```go
// config/model.go 添加文件存储配置
type FileStorageConfig struct {
    Type    string `yaml:"type"`     // local, oss, s3
    Path    string `yaml:"path"`     // 本地存储路径
    BaseURL string `yaml:"base_url"` // 文件访问基础URL
    MaxSize int64  `yaml:"max_size"` // 最大文件大小(MB)
}

type Config struct {
    // ... 其他配置
    FileStorage FileStorageConfig `yaml:"file_storage"`
}
```

### 文件API控制器

```go
// internal/apps/file/controller.go
package file

import (
    "net/http"
    "path/filepath"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/linux-do/cdk-office/internal/auth"
)

type FileController struct {
    fileService *FileService
    authService *auth.Service
}

// UploadFile 文件上传
func (c *FileController) UploadFile(ctx *gin.Context) {
    user := c.authService.GetCurrentUser(ctx)
    if user == nil {
        ctx.JSON(401, gin.H{"error": "未授权"})
        return
    }
    
    // 获取上传的文件
    file, header, err := ctx.Request.FormFile("file")
    if err != nil {
        ctx.JSON(400, gin.H{"error": "获取文件失败"})
        return
    }
    defer file.Close()
    
    // 获取文件分类
    category := ctx.PostForm("category")
    if category == "" {
        category = "general"
    }
    
    // 上传文件
    fileURL, err := c.fileService.UploadFile(file, header, category)
    if err != nil {
        ctx.JSON(500, gin.H{"error": fmt.Sprintf("上传失败: %v", err)})
        return
    }
    
    ctx.JSON(200, gin.H{
        "file_url":  fileURL,
        "file_name": header.Filename,
        "file_size": header.Size,
    })
}

// DownloadFile 文件下载
func (c *FileController) DownloadFile(ctx *gin.Context) {
    filePath := ctx.Param("filepath")
    
    // 安全检查：防止路径遍历攻击
    if filepath.IsAbs(filePath) || filepath.Clean(filePath) != filePath {
        ctx.JSON(400, gin.H{"error": "非法文件路径"})
        return
    }
    
    fullPath := filepath.Join(c.fileService.uploadPath, filePath)
    
    // 检查文件是否存在
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        ctx.JSON(404, gin.H{"error": "文件不存在"})
        return
    }
    
    ctx.File(fullPath)
}

// GetFileInfo 获取文件信息
func (c *FileController) GetFileInfo(ctx *gin.Context) {
    filePath := ctx.Param("filepath")
    
    fullPath := filepath.Join(c.fileService.uploadPath, filePath)
    
    info, err := os.Stat(fullPath)
    if err != nil {
        ctx.JSON(404, gin.H{"error": "文件不存在"})
        return
    }
    
    ctx.JSON(200, gin.H{
        "name":         info.Name(),
        "size":         info.Size(),
        "modified_at":  info.ModTime(),
        "is_directory": info.IsDir(),
    })
}
```

```go
// Contract 合同主表（使用PostgreSQL）
type Contract struct {
    ID              string    `gorm:"primaryKey;type:uuid" json:"id"`
    Title           string    `gorm:"size:200;not null" json:"title"`
    Content         string    `gorm:"type:text" json:"content"`
    TemplateID      string    `gorm:"type:uuid;index" json:"template_id"`
    CreatorID       string    `gorm:"type:uuid;not null;index" json:"creator_id"`
    TeamID          string    `gorm:"type:uuid;not null;index" json:"team_id"`
    Status          string    `gorm:"size:20;default:draft" json:"status"`
    EvidenceMode    string    `gorm:"size:20;default:blockchain" json:"evidence_mode"`
    FileURL         string    `gorm:"size:500" json:"file_url"`
    FileHash        string    `gorm:"size:64" json:"file_hash"`
    DocumentID      string    `gorm:"type:uuid;index" json:"document_id"` // 关联知识库文档
    ApprovalStatus  string    `gorm:"size:20;default:pending" json:"approval_status"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    CompletedAt     *time.Time `json:"completed_at"`
    
    // 关联关系
    Signers         []ContractSigner `gorm:"foreignKey:ContractID" json:"signers"`
    Document        Document        `gorm:"foreignKey:DocumentID" json:"document"`
    Creator         User           `gorm:"foreignKey:CreatorID" json:"creator"`
    Team            Team           `gorm:"foreignKey:TeamID" json:"team"`
}

// ContractSigner 签署方
type ContractSigner struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    ContractID   string    `gorm:"type:uuid;not null;index" json:"contract_id"`
    UserID       string    `gorm:"type:uuid;index" json:"user_id"`
    Name         string    `gorm:"size:100;not null" json:"name"`
    Mobile       string    `gorm:"size:20" json:"mobile"`
    Email        string    `gorm:"size:100" json:"email"`
    SignerType   string    `gorm:"size:20;not null" json:"signer_type"`
    Status       string    `gorm:"size:20;default:pending" json:"status"`
    SignOrder    int       `gorm:"default:1" json:"sign_order"`
    SignTime     *time.Time `json:"sign_time"`
    SignMethod   string    `gorm:"size:50" json:"sign_method"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// ContractTemplate 合同模板
type ContractTemplate struct {
    ID          string    `gorm:"primaryKey;type:uuid" json:"id"`
    Name        string    `gorm:"size:100;not null" json:"name"`
    Category    string    `gorm:"size:50" json:"category"`
    Content     string    `gorm:"type:text" json:"content"`
    Variables   JSON      `gorm:"type:json" json:"variables"`
    CreatorID   string    `gorm:"type:uuid;not null" json:"creator_id"`
    TeamID      string    `gorm:"type:uuid;index" json:"team_id"`
    IsPublic    bool      `gorm:"default:false" json:"is_public"`
    Status      string    `gorm:"size:20;default:active" json:"status"`
    UsageCount  int       `gorm:"default:0" json:"usage_count"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// ContractApproval 合同审批记录（集成审批流程）
type ContractApproval struct {
    ID              string    `gorm:"primaryKey;type:uuid" json:"id"`
    ContractID      string    `gorm:"type:uuid;not null;index" json:"contract_id"`
    WorkflowID      string    `gorm:"type:uuid;not null" json:"workflow_id"`
    ApprovalType    string    `gorm:"size:50;not null" json:"approval_type"` // knowledge_submit, document_archive
    Status          string    `gorm:"size:20;default:pending" json:"status"`
    ApproverID      string    `gorm:"type:uuid" json:"approver_id"`
    Comments        string    `gorm:"type:text" json:"comments"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    ApprovedAt      *time.Time `json:"approved_at"`
}
```

### 知识库集成数据模型

```go
// ContractDocument 合同文档（集成到知识库）
type ContractDocument struct {
    Document    // 继承基础文档模型
    
    ContractID      string    `gorm:"type:uuid;unique;not null" json:"contract_id"`
    ContractType    string    `gorm:"size:50" json:"contract_type"`
    SignersCount    int       `gorm:"default:0" json:"signers_count"`
    CompletedAt     *time.Time `json:"completed_at"`
    EvidenceURL     string    `gorm:"size:500" json:"evidence_url"`
    BlockchainHash  string    `gorm:"size:66" json:"blockchain_hash"`
}

// KnowledgeSubmission 知识库提交记录
type KnowledgeSubmission struct {
    ID                  string    `gorm:"primaryKey;type:uuid" json:"id"`
    ContractID          string    `gorm:"type:uuid;not null;index" json:"contract_id"`
    DocumentID          string    `gorm:"type:uuid;index" json:"document_id"`
    SubmissionType      string    `gorm:"size:50;not null" json:"submission_type"` // auto, manual
    Status              string    `gorm:"size:20;default:pending" json:"status"`
    ReviewerID          string    `gorm:"type:uuid" json:"reviewer_id"`
    ReviewComments      string    `gorm:"type:text" json:"review_comments"`
    AutoProcessing      bool      `gorm:"default:true" json:"auto_processing"`
    DifyWorkflowStatus  string    `gorm:"size:20" json:"dify_workflow_status"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
    ReviewedAt         *time.Time `json:"reviewed_at"`
}

// ContractWorkflow 合同工作流配置
type ContractWorkflow struct {
    ID                  string    `gorm:"primaryKey;type:uuid" json:"id"`
    TeamID              string    `gorm:"type:uuid;not null;index" json:"team_id"`
    WorkflowName        string    `gorm:"size:100;not null" json:"workflow_name"`
    TriggerEvent        string    `gorm:"size:50;not null" json:"trigger_event"` // contract_completed, contract_signed
    AutoSubmitKnowledge bool      `gorm:"default:true" json:"auto_submit_knowledge"`
    RequireApproval     bool      `gorm:"default:false" json:"require_approval"`
    ApprovalRoles       JSON      `gorm:"type:json" json:"approval_roles"`
    DifyIntegration     bool      `gorm:"default:true" json:"dify_integration"`
    NotificationConfig  JSON      `gorm:"type:json" json:"notification_config"`
    IsActive           bool      `gorm:"default:true" json:"is_active"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}
```

## API Endpoints Reference

### CDK-Office统一API接口

```yaml
# 合同管理接口（集成到CDK-Office）
GET /api/v1/contracts
  获取合同列表（使用CDK-Office用户权限）

POST /api/v1/contracts
  创建合同（自动关联当前用户和团队）

GET /api/v1/contracts/{id}
  获取合同详情

PUT /api/v1/contracts/{id}
  更新合同

DELETE /api/v1/contracts/{id}
  删除合同

# 合同签署接口
POST /api/v1/contracts/{id}/sign
  签署合同

POST /api/v1/contracts/{id}/send
  发送签署通知

POST /api/v1/contracts/{id}/reject
  拒绝签署

# 知识库集成接口
POST /api/v1/contracts/{id}/submit-knowledge
  手动提交合同到知识库

GET /api/v1/knowledge/submissions
  获取知识库提交记录

POST /api/v1/knowledge/submissions/{id}/approve
  审批知识库提交

# 工作流配置接口
GET /api/v1/teams/{team_id}/contract-workflows
  获取团队合同工作流配置

PUT /api/v1/teams/{team_id}/contract-workflows
  更新团队合同工作流配置

# 应用中心接口
GET /api/v1/apps/contracts
  获取电子合同应用信息

POST /api/v1/apps/contracts/launch
  启动电子合同应用（直接使用当前用户身份）
```

### 统一用户认证接口

```yaml
# 用户已通过CDK-Office登录，合同模块直接使用当前用户上下文
GET /api/v1/auth/current-user
  获取当前登录用户信息

GET /api/v1/auth/permissions
  获取当前用户的合同相关权限

POST /api/v1/auth/switch-team
  切换用户当前团队上下文
```

## Docker Compose Configuration

```yaml
version: '3.8'

services:
  # CDK-Office 统一服务（集成合同功能）
  cdk-office-api:
    build: ./cdk-office
    container_name: cdk-office-api
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql://cdkoffice:password123@shared-postgres:5432/cdk_office
      - REDIS_URL=redis://shared-redis:6379/0
      - DIFY_API_URL=http://dify-api:5001
      - DIFY_API_KEY=${DIFY_API_KEY}
      - CA_SERVICE_URL=${CA_SERVICE_URL}
      - SMS_SERVICE_URL=${SMS_SERVICE_URL}
      - BLOCKCHAIN_SERVICE_URL=${BLOCKCHAIN_SERVICE_URL}
      - CONTRACT_MODULE_ENABLED=true
      - AUTO_KNOWLEDGE_SUBMIT=true
      - FILE_STORAGE_TYPE=local  # 使用CDK-Office本地文件存储
      - FILE_STORAGE_PATH=/app/uploads
    volumes:
      - ./cdk-office/config:/app/config
      - shared-logs:/app/logs
      - shared-uploads:/app/uploads  # 使用本地文件存储
    depends_on:
      - shared-postgres
      - shared-redis
      - dify-api
    networks:
      - app-network
    restart: unless-stopped

  cdk-office-frontend:
    build: ./cdk-office/frontend
    container_name: cdk-office-frontend
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://localhost:8000
      - NEXT_PUBLIC_CONTRACT_ENABLED=true
    depends_on:
      - cdk-office-api
    networks:
      - app-network
    restart: unless-stopped

  # 统一数据库服务
  shared-postgres:
    image: postgres:15-alpine
    container_name: shared-postgres
    environment:
      - POSTGRES_DB=cdk_office
      - POSTGRES_USER=cdkoffice
      - POSTGRES_PASSWORD=password123
      - POSTGRES_MULTIPLE_EXTENSIONS=uuid-ossp,pg_trgm
    volumes:
      - shared-postgres-data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    networks:
      - app-network
    restart: unless-stopped

  # 统一缓存服务
  shared-redis:
    image: redis:7-alpine
    container_name: shared-redis
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    volumes:
      - shared-redis-data:/data
    ports:
      - "6379:6379"
    networks:
      - app-network
    restart: unless-stopped

  # 统一文件存储服务
  shared-minio:
    image: minio/minio
    container_name: shared-minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin123
    command: server /data --console-address ":9001"
    volumes:
      - shared-minio-data:/data
    networks:
      - app-network
    restart: unless-stopped

  # Dify AI平台服务
  dify-api:
    image: langgenius/dify-api:latest
    container_name: dify-api
    environment:
      - DATABASE_URL=postgresql://cdkoffice:password123@shared-postgres:5432/cdk_office
      - REDIS_URL=redis://shared-redis:6379/1
      - STORAGE_TYPE=local  # 使用本地存储
      - STORAGE_LOCAL_PATH=/app/storage
    volumes:
      - ./dify/api:/app
      - shared-uploads:/app/storage  # 共享文件存储目录
    depends_on:
      - shared-postgres
      - shared-redis
    networks:
      - app-network
    restart: unless-stopped

  # Nginx 反向代理
  nginx:
    image: nginx:alpine
    container_name: contract-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - cdk-office-api
      - mini-contract-api
    networks:
      - app-network

volumes:
  shared-postgres-data:
  shared-redis-data:
  shared-minio-data:
  shared-logs:
  dify-data:

networks:
  app-network:
    driver: bridge
```

```yaml
# config.yaml - CDK-Office配置文件
app:
  name: "cdk-office"
  port: 8000
  debug: false

# 数据库配置
database:
  type: "postgres"
  host: "shared-postgres"
  port: 5432
  username: "cdkoffice"
  password: "password123"
  database: "cdk_office"
  sslmode: "disable"

# Redis配置
redis:
  host: "shared-redis"
  port: 6379
  password: ""
  db: 0

# 文件存储配置
file_storage:
  type: "local"                    # 使用本地存储
  path: "/app/uploads"             # 存储路径
  base_url: "http://localhost:8000" # 访问基础URL
  max_size: 100                     # 最大100MB

# 合同模块配置
contract:
  enabled: true
  auto_knowledge_submit: true
  evidence_mode: "blockchain"       # blockchain, simple
  ca_service_url: "${CA_SERVICE_URL}"
  sms_service_url: "${SMS_SERVICE_URL}"
  blockchain_service_url: "${BLOCKCHAIN_SERVICE_URL}"

# Dify集成配置
dify:
  api_url: "http://dify-api:5001"
  api_key: "${DIFY_API_KEY}"
  knowledge_base_id: "${DIFY_KNOWLEDGE_BASE_ID}"
  
# 知识库工作流配置
knowledge_workflow:
  auto_submit: true
  require_approval: false
  approval_roles: ["admin", "manager"]
  dify_processing: true
```

```nginx
# nginx.conf
upstream cdk-office {
    server cdk-office-api:8000;
}

upstream mini-contract {
    server mini-contract-api:8080;
}

server {
    listen 80;
    server_name localhost;

    # CDK-Office 主应用
    location / {
        proxy_pass http://cdk-office;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # Mini-Contract 代理
    location /contract/ {
        proxy_pass http://mini-contract/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # 文件上传大小限制
    client_max_body_size 100M;
}
```

## Testing

### 单元测试

```go
// internal/apps/contract/service_test.go
func TestContractService_ProxyRequest(t *testing.T) {
    service := NewContractService("http://localhost:8080")
    
    // 测试代理GET请求
    req := httptest.NewRequest("GET", "/api/v1/contracts", nil)
    w := httptest.NewRecorder()
    
    err := service.ProxyRequest(gin.Context{Request: req, Writer: w}, "team-001")
    assert.NoError(t, err)
    assert.Equal(t, 200, w.Code)
}

func TestContractService_GenerateSSO(t *testing.T) {
    service := NewContractService("http://localhost:8080")
    user := &auth.User{ID: "user-001", TeamID: "team-001"}
    
    accessURL, token, err := service.GenerateSSO(user)
    assert.NoError(t, err)
    assert.NotEmpty(t, accessURL)
    assert.NotEmpty(t, token)
}
```