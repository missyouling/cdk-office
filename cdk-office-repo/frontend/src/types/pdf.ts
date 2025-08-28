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

/**
 * PDF处理工具类型定义
 */

// PDF操作请求
export interface PDFOperationRequest {
  operation_type: string;
  file_data?: File[];
  options?: Record<string, any>;
  output_format?: string;
  output_name?: string;
}

// PDF操作结果
export interface PDFOperationResult {
  operation_id: string;
  operation_type: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  input_files: string[];
  output_file?: string;
  download_url?: string;
  error_message?: string;
  processing_time?: number;
  created_at: string;
  completed_at?: string;
}

// PDF合并请求
export interface PDFMergeRequest {
  files: File[];
  output_name?: string;
  bookmark_levels?: number[];
}

// PDF拆分请求
export interface PDFSplitRequest {
  file: File;
  split_type: 'pages' | 'range' | 'size';
  pages?: number[];
  ranges?: string[];
  max_size_mb?: number;
  output_prefix?: string;
}

// PDF压缩请求
export interface PDFCompressRequest {
  file: File;
  quality: 'low' | 'medium' | 'high' | 'maximum';
  optimize_images: boolean;
  remove_metadata: boolean;
  output_name?: string;
}

// PDF水印请求
export interface PDFWatermarkRequest {
  file: File;
  watermark_type: 'text' | 'image';
  watermark_content: string;
  watermark_image?: File;
  position: 'center' | 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
  opacity: number;
  rotation: number;
  font_size?: number;
  font_color?: string;
  output_name?: string;
}

// PDF密码保护请求
export interface PDFProtectRequest {
  file: File;
  owner_password: string;
  user_password?: string;
  permissions: {
    print: boolean;
    copy: boolean;
    modify: boolean;
    annotate: boolean;
    fill_forms: boolean;
    extract_text: boolean;
    assemble: boolean;
    high_quality_print: boolean;
  };
  output_name?: string;
}

// PDF OCR请求
export interface PDFOCRRequest {
  file: File;
  languages: string[];
  ocr_type: 'force' | 'skip_text' | 'only_images';
  output_format: 'pdf' | 'pdf_searchable' | 'text' | 'hocr';
  dpi: number;
  output_name?: string;
}

// PDF转换请求
export interface PDFConvertRequest {
  file: File;
  target_format: 'docx' | 'xlsx' | 'pptx' | 'html' | 'txt' | 'xml' | 'json';
  options?: {
    extract_images?: boolean;
    preserve_layout?: boolean;
    include_metadata?: boolean;
  };
  output_name?: string;
}

// PDF元数据
export interface PDFMetadata {
  title?: string;
  author?: string;
  subject?: string;
  keywords?: string;
  creator?: string;
  producer?: string;
  creation_date?: string;
  modification_date?: string;
  page_count: number;
  file_size: number;
  version: string;
  encrypted: boolean;
  permissions?: {
    print: boolean;
    copy: boolean;
    modify: boolean;
    annotate: boolean;
  };
}

// PDF签名请求
export interface PDFSignRequest {
  file: File;
  certificate: File;
  private_key: File;
  password: string;
  signature_text?: string;
  signature_image?: File;
  position: {
    page: number;
    x: number;
    y: number;
    width: number;
    height: number;
  };
  output_name?: string;
}

// PDF修复请求
export interface PDFRepairRequest {
  file: File;
  repair_type: 'basic' | 'advanced' | 'force';
  output_name?: string;
}

// PDF优化请求
export interface PDFOptimizeRequest {
  file: File;
  optimization_level: 'basic' | 'standard' | 'maximum';
  options: {
    compress_images: boolean;
    remove_unused_objects: boolean;
    flatten_forms: boolean;
    remove_metadata: boolean;
    linearize: boolean;
  };
  output_name?: string;
}

// PDF工具操作分类
export interface PDFToolCategory {
  id: string;
  name: string;
  description: string;
  icon: string;
  operations: PDFOperation[];
}

// PDF操作定义
export interface PDFOperation {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: string;
  input_types: string[];
  output_types: string[];
  options?: PDFOperationOption[];
  examples?: string[];
}

// PDF操作选项
export interface PDFOperationOption {
  key: string;
  label: string;
  type: 'text' | 'number' | 'select' | 'checkbox' | 'file' | 'range';
  required: boolean;
  default_value?: any;
  options?: Array<{ label: string; value: any }>;
  min?: number;
  max?: number;
  step?: number;
  description?: string;
}

// PDF批量操作请求
export interface PDFBatchRequest {
  files: File[];
  operation_type: string;
  options: Record<string, any>;
  output_format?: string;
  output_prefix?: string;
}

// PDF工作流请求
export interface PDFWorkflowRequest {
  files: File[];
  workflow_steps: Array<{
    operation_type: string;
    options: Record<string, any>;
  }>;
  output_name?: string;
}

// PDF处理历史
export interface PDFProcessingHistory {
  id: string;
  user_id: string;
  operation_type: string;
  input_files: string[];
  output_file?: string;
  status: string;
  error_message?: string;
  processing_time?: number;
  file_size_before: number;
  file_size_after?: number;
  created_at: string;
  completed_at?: string;
}