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
  PersonalKnowledge,
  WeChatRecord,
  KnowledgeStatistics,
  CreateKnowledgeRequest,
  UpdateKnowledgeRequest,
  ListKnowledgeRequest,
  ListKnowledgeResponse,
  SearchKnowledgeRequest,
  SearchKnowledgeResponse,
  WeChatUploadRequest,
  ShareToTeamRequest,
  BatchOperationRequest,
  BatchOperationResponse,
  TagStat,
} from '@/types/knowledge';

/**
 * 个人知识库API服务
 */
class KnowledgeService {
  private readonly baseURL = '/api/knowledge';

  /**
   * 创建个人知识
   */
  async createKnowledge(data: CreateKnowledgeRequest): Promise<PersonalKnowledge> {
    const response = await axios.post(this.baseURL, data);
    return response.data;
  }

  /**
   * 获取个人知识详情
   */
  async getKnowledge(id: string): Promise<PersonalKnowledge> {
    const response = await axios.get(`${this.baseURL}/${id}`);
    return response.data;
  }

  /**
   * 更新个人知识
   */
  async updateKnowledge(id: string, data: UpdateKnowledgeRequest): Promise<PersonalKnowledge> {
    const response = await axios.put(`${this.baseURL}/${id}`, data);
    return response.data;
  }

  /**
   * 删除个人知识
   */
  async deleteKnowledge(id: string): Promise<void> {
    await axios.delete(`${this.baseURL}/${id}`);
  }

  /**
   * 列出个人知识
   */
  async listKnowledge(params: ListKnowledgeRequest = {}): Promise<ListKnowledgeResponse> {
    const response = await axios.get(this.baseURL, { params });
    return response.data;
  }

  /**
   * 搜索个人知识
   */
  async searchKnowledge(data: SearchKnowledgeRequest): Promise<SearchKnowledgeResponse> {
    const response = await axios.post(`${this.baseURL}/search`, data);
    return response.data;
  }

  /**
   * 获取知识库统计信息
   */
  async getStatistics(): Promise<KnowledgeStatistics> {
    const response = await axios.get(`${this.baseURL}/statistics`);
    return response.data;
  }

  /**
   * 获取热门标签
   */
  async getPopularTags(limit = 10): Promise<TagStat[]> {
    const response = await axios.get(`${this.baseURL}/tags/popular`, {
      params: { limit }
    });
    return response.data;
  }

  /**
   * 分享知识到团队
   */
  async shareToTeam(id: string, data: Omit<ShareToTeamRequest, 'knowledge_id'>): Promise<void> {
    await axios.post(`${this.baseURL}/${id}/share`, {
      ...data,
      knowledge_id: id,
    });
  }

  /**
   * 获取分享状态
   */
  async getShareStatus(id: string): Promise<any> {
    const response = await axios.get(`${this.baseURL}/${id}/share-status`);
    return response.data;
  }

  /**
   * 批量删除知识
   */
  async batchDelete(knowledgeIds: string[]): Promise<BatchOperationResponse> {
    const response = await axios.post(`${this.baseURL}/batch/delete`, {
      knowledge_ids: knowledgeIds,
    });
    return response.data;
  }

  /**
   * 批量更新知识
   */
  async batchUpdate(data: BatchOperationRequest): Promise<BatchOperationResponse> {
    const response = await axios.post(`${this.baseURL}/batch/update`, data);
    return response.data;
  }

  /**
   * 导入知识
   */
  async importKnowledge(format: string, data: string, options?: any): Promise<any> {
    const response = await axios.post(`${this.baseURL}/import`, {
      format,
      data,
      options,
    });
    return response.data;
  }

  /**
   * 导出知识
   */
  async exportKnowledge(format: string, knowledgeIds?: string[], options?: any): Promise<any> {
    const response = await axios.post(`${this.baseURL}/export`, {
      format,
      knowledge_ids: knowledgeIds,
      include_private: true,
      options,
    });
    return response.data;
  }

  /**
   * 上传微信聊天记录
   */
  async uploadWeChatRecords(data: WeChatUploadRequest): Promise<any> {
    const response = await axios.post(`${this.baseURL}/wechat/upload`, data);
    return response.data;
  }

  /**
   * 列出微信聊天记录
   */
  async listWeChatRecords(params: any = {}): Promise<any> {
    const response = await axios.get(`${this.baseURL}/wechat/records`, { params });
    return response.data;
  }

  /**
   * 获取微信聊天记录详情
   */
  async getWeChatRecord(id: string): Promise<WeChatRecord> {
    const response = await axios.get(`${this.baseURL}/wechat/records/${id}`);
    return response.data;
  }

  /**
   * 删除微信聊天记录
   */
  async deleteWeChatRecord(id: string): Promise<void> {
    await axios.delete(`${this.baseURL}/wechat/records/${id}`);
  }

  /**
   * 归档微信聊天记录到知识库
   */
  async archiveWeChatRecord(id: string, data: {
    title: string;
    description?: string;
    tags?: string[];
    category?: string;
  }): Promise<PersonalKnowledge> {
    const response = await axios.post(`${this.baseURL}/wechat/records/${id}/archive`, data);
    return response.data;
  }
}

export const knowledgeService = new KnowledgeService();