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
  PDFOperationRequest,
  PDFOperationResult,
  PDFMergeRequest,
  PDFSplitRequest,
  PDFCompressRequest,
  PDFWatermarkRequest,
  PDFProtectRequest,
  PDFOCRRequest,
  PDFConvertRequest,
  PDFSignRequest,
  PDFRepairRequest,
  PDFOptimizeRequest,
  PDFMetadata,
  PDFBatchRequest,
  PDFWorkflowRequest,
  PDFProcessingHistory,
  PDFToolCategory,
} from '@/types/pdf';

class PDFService {
  private readonly baseURL = '/api/pdf';

  // 创建FormData的辅助方法
  private createFormData(data: any, files?: File[]): FormData {
    const formData = new FormData();
    
    // 添加JSON数据
    formData.append('data', JSON.stringify(data));
    
    // 添加文件
    if (files) {
      files.forEach((file, index) => {
        formData.append(`file_${index}`, file);
      });
    }
    
    return formData;
  }

  // 通用PDF操作
  async performOperation(request: PDFOperationRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: request.operation_type,
        options: request.options,
        output_format: request.output_format,
        output_name: request.output_name,
      },
      request.file_data
    );

    const response = await axios.post(`${this.baseURL}/operation`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // PDF合并
  async mergePDFs(request: PDFMergeRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'merge',
        output_name: request.output_name,
        bookmark_levels: request.bookmark_levels,
      },
      request.files
    );

    const response = await axios.post(`${this.baseURL}/merge`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // PDF拆分
  async splitPDF(request: PDFSplitRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'split',
        split_type: request.split_type,
        pages: request.pages,
        ranges: request.ranges,
        max_size_mb: request.max_size_mb,
        output_prefix: request.output_prefix,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/split`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // PDF压缩
  async compressPDF(request: PDFCompressRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'compress',
        quality: request.quality,
        optimize_images: request.optimize_images,
        remove_metadata: request.remove_metadata,
        output_name: request.output_name,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/compress`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 添加水印
  async addWatermark(request: PDFWatermarkRequest): Promise<PDFOperationResult> {
    const files = [request.file];
    if (request.watermark_image) {
      files.push(request.watermark_image);
    }

    const formData = this.createFormData(
      {
        operation_type: 'watermark',
        watermark_type: request.watermark_type,
        watermark_content: request.watermark_content,
        position: request.position,
        opacity: request.opacity,
        rotation: request.rotation,
        font_size: request.font_size,
        font_color: request.font_color,
        output_name: request.output_name,
      },
      files
    );

    const response = await axios.post(`${this.baseURL}/watermark`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 密码保护
  async protectPDF(request: PDFProtectRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'protect',
        owner_password: request.owner_password,
        user_password: request.user_password,
        permissions: request.permissions,
        output_name: request.output_name,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/protect`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // OCR识别
  async performOCR(request: PDFOCRRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'ocr',
        languages: request.languages,
        ocr_type: request.ocr_type,
        output_format: request.output_format,
        dpi: request.dpi,
        output_name: request.output_name,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/ocr`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 格式转换
  async convertPDF(request: PDFConvertRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'convert',
        target_format: request.target_format,
        options: request.options,
        output_name: request.output_name,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/convert`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // PDF签名
  async signPDF(request: PDFSignRequest): Promise<PDFOperationResult> {
    const files = [request.file, request.certificate, request.private_key];
    if (request.signature_image) {
      files.push(request.signature_image);
    }

    const formData = this.createFormData(
      {
        operation_type: 'sign',
        password: request.password,
        signature_text: request.signature_text,
        position: request.position,
        output_name: request.output_name,
      },
      files
    );

    const response = await axios.post(`${this.baseURL}/sign`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // PDF修复
  async repairPDF(request: PDFRepairRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'repair',
        repair_type: request.repair_type,
        output_name: request.output_name,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/repair`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // PDF优化
  async optimizePDF(request: PDFOptimizeRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        operation_type: 'optimize',
        optimization_level: request.optimization_level,
        options: request.options,
        output_name: request.output_name,
      },
      [request.file]
    );

    const response = await axios.post(`${this.baseURL}/optimize`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 获取PDF元数据
  async getPDFMetadata(file: File): Promise<PDFMetadata> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await axios.post(`${this.baseURL}/metadata`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 批量处理
  async batchProcess(request: PDFBatchRequest): Promise<PDFOperationResult[]> {
    const formData = this.createFormData(
      {
        operation_type: request.operation_type,
        options: request.options,
        output_format: request.output_format,
        output_prefix: request.output_prefix,
      },
      request.files
    );

    const response = await axios.post(`${this.baseURL}/batch`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 工作流处理
  async processWorkflow(request: PDFWorkflowRequest): Promise<PDFOperationResult> {
    const formData = this.createFormData(
      {
        workflow_steps: request.workflow_steps,
        output_name: request.output_name,
      },
      request.files
    );

    const response = await axios.post(`${this.baseURL}/workflow`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  }

  // 获取操作结果
  async getOperationResult(operationId: string): Promise<PDFOperationResult> {
    const response = await axios.get(`${this.baseURL}/operation/${operationId}`);
    return response.data;
  }

  // 下载处理结果
  async downloadResult(operationId: string): Promise<Blob> {
    const response = await axios.get(`${this.baseURL}/download/${operationId}`, {
      responseType: 'blob',
    });
    return response.data;
  }

  // 获取处理历史
  async getProcessingHistory(page: number = 1, pageSize: number = 20): Promise<{
    history: PDFProcessingHistory[];
    total: number;
    page: number;
    page_size: number;
  }> {
    const response = await axios.get(`${this.baseURL}/history`, {
      params: { page, page_size: pageSize },
    });
    return response.data;
  }

  // 获取工具分类
  async getToolCategories(): Promise<PDFToolCategory[]> {
    const response = await axios.get(`${this.baseURL}/categories`);
    return response.data;
  }

  // 删除处理历史
  async deleteHistory(historyId: string): Promise<void> {
    await axios.delete(`${this.baseURL}/history/${historyId}`);
  }

  // 清理临时文件
  async cleanupTempFiles(): Promise<void> {
    await axios.post(`${this.baseURL}/cleanup`);
  }
}

export const pdfService = new PDFService();
export default pdfService;