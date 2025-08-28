/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import axios from 'axios';
import {
  ScannedDocument,
  ScanUploadRequest,
  ScanTask,
  ImageProcessingResult,
  MobilePermission,
  ScanSession,
  CapturedImage,
  BatchScanRequest,
  MobileOCRRequest,
  MobileOCRResult,
  DocumentPreview,
  ScanHistory,
  QualityAnalysis,
  ScanStatistics,
  DocumentProcessOptions,
} from '@/types/scanner';

class ScannerService {
  private readonly baseURL = '/api/mobile/scanner';

  // 检查用户权限
  async checkPermissions(): Promise<MobilePermission> {
    const response = await axios.get(`${this.baseURL}/permissions`);
    return response.data;
  }

  // 创建扫描会话
  async createScanSession(sessionName: string): Promise<ScanSession> {
    const response = await axios.post(`${this.baseURL}/session`, {
      session_name: sessionName,
    });
    return response.data;
  }

  // 获取扫描会话
  async getScanSession(sessionId: string): Promise<ScanSession> {
    const response = await axios.get(`${this.baseURL}/session/${sessionId}`);
    return response.data;
  }

  // 更新扫描会话
  async updateScanSession(
    sessionId: string, 
    updates: Partial<ScanSession>
  ): Promise<ScanSession> {
    const response = await axios.put(`${this.baseURL}/session/${sessionId}`, updates);
    return response.data;
  }

  // 上传扫描图像
  async uploadScanImages(request: ScanUploadRequest): Promise<ScanTask> {
    const formData = new FormData();
    
    // 添加图像文件
    request.images.forEach((image, index) => {
      formData.append(`image_${index}`, image);
    });
    
    // 添加其他数据
    formData.append('to_personal_kb', request.to_personal_kb.toString());
    formData.append('processing_options', JSON.stringify(request.processing_options));
    
    if (request.document_name) {
      formData.append('document_name', request.document_name);
    }
    
    if (request.tags) {
      formData.append('tags', JSON.stringify(request.tags));
    }
    
    if (request.category) {
      formData.append('category', request.category);
    }

    const response = await axios.post(`${this.baseURL}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 添加图像到会话
  async addImageToSession(
    sessionId: string, 
    imageFile: File,
    metadata?: any
  ): Promise<CapturedImage> {
    const formData = new FormData();
    formData.append('image', imageFile);
    formData.append('session_id', sessionId);
    
    if (metadata) {
      formData.append('metadata', JSON.stringify(metadata));
    }

    const response = await axios.post(`${this.baseURL}/session/${sessionId}/images`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 处理单个图像
  async processImage(
    imageId: string,
    options: DocumentProcessOptions
  ): Promise<ImageProcessingResult> {
    const response = await axios.post(`${this.baseURL}/process-image/${imageId}`, {
      processing_options: options,
    });
    return response.data;
  }

  // 批量处理会话中的所有图像
  async processBatchScan(request: BatchScanRequest): Promise<ScanTask> {
    const response = await axios.post(`${this.baseURL}/batch-process`, request);
    return response.data;
  }

  // 获取扫描任务状态
  async getScanTaskStatus(taskId: string): Promise<ScanTask> {
    const response = await axios.get(`${this.baseURL}/task/${taskId}`);
    return response.data;
  }

  // OCR识别
  async performOCR(request: MobileOCRRequest): Promise<MobileOCRResult> {
    const response = await axios.post(`${this.baseURL}/ocr`, request);
    return response.data;
  }

  // 质量分析
  async analyzeImageQuality(imageUri: string): Promise<QualityAnalysis> {
    const response = await axios.post(`${this.baseURL}/analyze-quality`, {
      image_uri: imageUri,
    });
    return response.data;
  }

  // 获取文档预览
  async getDocumentPreview(documentId: string): Promise<DocumentPreview> {
    const response = await axios.get(`${this.baseURL}/preview/${documentId}`);
    return response.data;
  }

  // 下载扫描结果
  async downloadScannedDocument(documentId: string): Promise<Blob> {
    const response = await axios.get(`${this.baseURL}/download/${documentId}`, {
      responseType: 'blob',
    });
    return response.data;
  }

  // 删除扫描文档
  async deleteScannedDocument(documentId: string): Promise<void> {
    await axios.delete(`${this.baseURL}/document/${documentId}`);
  }

  // 获取扫描历史
  async getScanHistory(page: number = 1, pageSize: number = 20): Promise<{
    history: ScanHistory[];
    total: number;
    page: number;
    page_size: number;
  }> {
    const response = await axios.get(`${this.baseURL}/history`, {
      params: { page, page_size: pageSize },
    });
    return response.data;
  }

  // 获取扫描统计
  async getScanStatistics(): Promise<ScanStatistics> {
    const response = await axios.get(`${this.baseURL}/statistics`);
    return response.data;
  }

  // 分享到个人知识库
  async shareToPersonalKB(documentId: string): Promise<void> {
    await axios.post(`${this.baseURL}/share-to-kb/${documentId}`);
  }

  // 分享到团队知识库
  async shareToTeamKB(documentId: string, teamId: string, reason: string): Promise<void> {
    await axios.post(`${this.baseURL}/share-to-team/${documentId}`, {
      team_id: teamId,
      reason: reason,
    });
  }

  // 更新文档标签
  async updateDocumentTags(documentId: string, tags: string[]): Promise<void> {
    await axios.put(`${this.baseURL}/document/${documentId}/tags`, {
      tags: tags,
    });
  }

  // 重新处理文档
  async reprocessDocument(
    documentId: string,
    options: DocumentProcessOptions
  ): Promise<ScanTask> {
    const response = await axios.post(`${this.baseURL}/reprocess/${documentId}`, {
      processing_options: options,
    });
    return response.data;
  }

  // 获取处理进度
  async getProcessingProgress(taskId: string): Promise<{
    progress: number;
    current_step: string;
    total_steps: number;
    estimated_time_remaining: number;
  }> {
    const response = await axios.get(`${this.baseURL}/progress/${taskId}`);
    return response.data;
  }

  // 取消处理任务
  async cancelProcessingTask(taskId: string): Promise<void> {
    await axios.post(`${this.baseURL}/cancel/${taskId}`);
  }

  // 清理临时文件
  async cleanupTempFiles(sessionId: string): Promise<void> {
    await axios.post(`${this.baseURL}/cleanup/${sessionId}`);
  }

  // 获取用户配置
  async getUserConfig(): Promise<{
    default_processing_options: DocumentProcessOptions;
    auto_upload_to_kb: boolean;
    image_quality: string;
    ocr_language: string;
  }> {
    const response = await axios.get(`${this.baseURL}/config`);
    return response.data;
  }

  // 更新用户配置
  async updateUserConfig(config: {
    default_processing_options?: DocumentProcessOptions;
    auto_upload_to_kb?: boolean;
    image_quality?: string;
    ocr_language?: string;
  }): Promise<void> {
    await axios.put(`${this.baseURL}/config`, config);
  }

  // 搜索扫描文档
  async searchDocuments(query: string, filters?: {
    date_range?: [string, string];
    tags?: string[];
    shared_only?: boolean;
  }): Promise<{
    documents: ScannedDocument[];
    total: number;
  }> {
    const response = await axios.get(`${this.baseURL}/search`, {
      params: {
        q: query,
        ...filters,
      },
    });
    return response.data;
  }

  // 批量操作
  async batchOperation(
    documentIds: string[],
    operation: 'delete' | 'share_to_kb' | 'update_tags',
    options?: any
  ): Promise<{
    success_count: number;
    failed_count: number;
    failed_ids: string[];
  }> {
    const response = await axios.post(`${this.baseURL}/batch-operation`, {
      document_ids: documentIds,
      operation: operation,
      options: options,
    });
    return response.data;
  }
}

export const scannerService = new ScannerService();
export default scannerService;