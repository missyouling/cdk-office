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
 * 个人知识库类型定义
 */

export interface PersonalKnowledge {
  id: string;
  user_id: string;
  title: string;
  description?: string;
  content: string;
  content_type: 'markdown' | 'text' | 'html';
  tags: string[];
  category?: string;
  privacy: 'private' | 'shared' | 'public';
  source_type: 'manual' | 'wechat' | 'upload' | 'scan';
  source_data?: Record<string, any>;
  is_shared: boolean;
  shared_at?: string;
  created_at: string;
  updated_at: string;
}

export interface WeChatRecord {
  id: string;
  user_id: string;
  session_name: string;
  message_type: 'text' | 'image' | 'voice' | 'video' | 'file';
  message_id: string;
  sender_name: string;
  sender_id: string;
  content: string;
  message_time: string;
  file_data?: string;
  file_name?: string;
  original_file?: string;
  processed_file?: string;
  ocr_text?: string;
  extracted_info?: string;
  is_archived: boolean;
  archived_to?: string;
  process_status: 'pending' | 'processing' | 'completed' | 'failed';
  created_at: string;
  updated_at: string;
}

export interface KnowledgeStatistics {
  total_knowledge: number;
  shared_knowledge: number;
  weekly_added: number;
  by_category: Array<{
    category: string;
    count: number;
  }>;
  by_source: Array<{
    source_type: string;
    count: number;
  }>;
}

export interface CreateKnowledgeRequest {
  title: string;
  description?: string;
  content: string;
  content_type?: string;
  tags?: string[];
  category?: string;
  privacy?: string;
  source_type?: string;
  source_data?: Record<string, any>;
}

export interface UpdateKnowledgeRequest {
  title?: string;
  description?: string;
  content?: string;
  content_type?: string;
  tags?: string[];
  category?: string;
  privacy?: string;
}

export interface ListKnowledgeRequest {
  page?: number;
  page_size?: number;
  category?: string;
  privacy?: string;
  source_type?: string;
  keyword?: string;
  tags?: string[];
  sort_by?: string;
}

export interface ListKnowledgeResponse {
  knowledge: PersonalKnowledge[];
  total: number;
  page: number;
  page_size: number;
}

export interface SearchKnowledgeRequest {
  query: string;
  tags?: string[];
  category?: string;
  source_type?: string;
  page?: number;
  page_size?: number;
}

export interface SearchKnowledgeResponse {
  results: PersonalKnowledge[];
  total: number;
  page: number;
  page_size: number;
  query: string;
}

export interface WeChatUploadRequest {
  session_name: string;
  records: Array<{
    message_id?: string;
    message_type: string;
    sender_name?: string;
    sender_id?: string;
    content: string;
    message_time?: string;
    file_data?: string;
    file_name?: string;
    extra_data?: Record<string, any>;
  }>;
  process_config?: {
    enable_ocr?: boolean;
    enable_auto_archive?: boolean;
    filter_message_types?: string[];
    extract_keywords?: boolean;
    group_by_session?: boolean;
  };
}

export interface ShareToTeamRequest {
  knowledge_id: string;
  team_id: string;
  share_reason: string;
}

export interface BatchOperationRequest {
  knowledge_ids: string[];
  updates?: Record<string, any>;
}

export interface BatchOperationResponse {
  success_count: number;
  failed_count: number;
  failed_ids?: string[];
}

export interface TagStat {
  tag: string;
  count: number;
}

export interface KnowledgeFilter {
  category?: string;
  privacy?: string;
  source_type?: string;
  tags?: string[];
  search?: string;
  sort_by?: 'created_at' | 'updated_at' | 'title';
  sort_order?: 'asc' | 'desc';
}