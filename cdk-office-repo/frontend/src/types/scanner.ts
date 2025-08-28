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
 * 移动端文档扫描类型定义
 */

// 扫描文档结果
export interface ScannedDocument {
  id: string;
  user_id: string;
  team_id?: string;
  file_name: string;
  file_path: string;
  file_type: 'pdf' | 'image';
  file_size: number;
  ocr_text?: string;
  processed: boolean;
  tags: string[];
  created_at: string;
  original_images: string[];
  processed_images: string[];
  processing_options: DocumentProcessOptions;
}

// 文档处理选项
export interface DocumentProcessOptions {
  perspective_correction: boolean;  // 透视矫正
  brightness_adjustment: boolean;   // 亮度调整
  contrast_enhancement: boolean;    // 对比度增强
  noise_reduction: boolean;         // 降噪
  text_enhancement: boolean;        // 文本增强
  auto_crop: boolean;              // 自动裁剪
  deskew: boolean;                 // 去倾斜
  shadow_removal: boolean;         // 阴影去除
}

// 扫描上传请求
export interface ScanUploadRequest {
  images: File[];
  to_personal_kb: boolean;
  processing_options: DocumentProcessOptions;
  document_name?: string;
  tags?: string[];
  category?: string;
}

// 扫描任务
export interface ScanTask {
  id: string;
  user_id: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  images_count: number;
  processed_count: number;
  processing_options: DocumentProcessOptions;
  result_document?: ScannedDocument;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

// 图像处理结果
export interface ImageProcessingResult {
  original_image: string;
  processed_image: string;
  processing_time: number;
  quality_score: number;
  detected_text_regions: TextRegion[];
  perspective_corrected: boolean;
  enhancement_applied: DocumentProcessOptions;
}

// 文本区域
export interface TextRegion {
  x: number;
  y: number;
  width: number;
  height: number;
  confidence: number;
  text?: string;
}

// 移动端权限检查
export interface MobilePermission {
  user_id: string;
  has_personal_kb_access: boolean;
  has_camera_access: boolean;
  has_storage_access: boolean;
  daily_scan_limit: number;
  used_scan_count: number;
  subscription_type: 'free' | 'basic' | 'premium';
}

// 相机配置
export interface CameraConfig {
  resolution: 'low' | 'medium' | 'high' | 'ultra';
  flash_mode: 'auto' | 'on' | 'off';
  focus_mode: 'auto' | 'manual';
  enable_grid: boolean;
  enable_edge_detection: boolean;
  capture_format: 'jpeg' | 'png';
  quality: number;
}

// 扫描会话
export interface ScanSession {
  id: string;
  user_id: string;
  session_name: string;
  captured_images: CapturedImage[];
  processing_options: DocumentProcessOptions;
  status: 'active' | 'completed' | 'cancelled';
  created_at: string;
  updated_at: string;
}

// 拍摄的图像
export interface CapturedImage {
  id: string;
  session_id: string;
  image_uri: string;
  thumbnail_uri: string;
  capture_time: string;
  camera_settings: CameraConfig;
  image_metadata: ImageMetadata;
  processing_status: 'pending' | 'processing' | 'completed' | 'failed';
  processed_uri?: string;
}

// 图像元数据
export interface ImageMetadata {
  width: number;
  height: number;
  file_size: number;
  format: string;
  exif_data?: Record<string, any>;
  location?: {
    latitude: number;
    longitude: number;
  };
  device_info?: {
    model: string;
    os_version: string;
    app_version: string;
  };
}

// 批量扫描请求
export interface BatchScanRequest {
  session_id: string;
  processing_options: DocumentProcessOptions;
  output_format: 'pdf' | 'images' | 'both';
  merge_to_single_pdf: boolean;
  document_name?: string;
  tags?: string[];
  to_personal_kb: boolean;
}

// OCR识别请求
export interface MobileOCRRequest {
  image_uri: string;
  language: string;
  recognition_type: 'text' | 'table' | 'handwriting';
  enhance_image: boolean;
}

// OCR识别结果
export interface MobileOCRResult {
  recognized_text: string;
  confidence: number;
  text_regions: TextRegion[];
  processing_time: number;
  language_detected: string;
  word_count: number;
}

// 文档预览
export interface DocumentPreview {
  document_id: string;
  preview_images: string[];
  thumbnail: string;
  ocr_preview: string;
  page_count: number;
  estimated_word_count: number;
  file_size: number;
  quality_score: number;
}

// 扫描历史
export interface ScanHistory {
  id: string;
  user_id: string;
  document_name: string;
  scan_date: string;
  page_count: number;
  file_size: number;
  status: string;
  thumbnail: string;
  tags: string[];
  shared_to_kb: boolean;
}

// 质量分析结果
export interface QualityAnalysis {
  overall_score: number;
  sharpness_score: number;
  brightness_score: number;
  contrast_score: number;
  text_clarity_score: number;
  suggestions: string[];
  auto_enhancement_recommended: boolean;
}

// 扫描统计
export interface ScanStatistics {
  total_scans: number;
  this_month_scans: number;
  total_pages: number;
  average_quality_score: number;
  most_used_tags: string[];
  scan_frequency: Array<{
    date: string;
    count: number;
  }>;
}