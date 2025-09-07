// pages/document/detail.js
Page({
  data: {
    document: null,
    versions: [],
    activeTab: 'content',
    loading: true,
    error: null
  },

  onLoad: function (options) {
    const docId = options.id;
    if (docId) {
      this.loadDocumentDetail(docId);
    } else {
      this.setData({
        error: '文档ID不存在',
        loading: false
      });
    }
  },

  // 加载文档详情
  loadDocumentDetail: function (docId) {
    wx.showLoading({ title: '加载中...' });
    
    // 模拟API调用
    // 实际开发中需要替换为真实的API调用
    setTimeout(() => {
      // 模拟文档数据
      const mockDocument = {
        id: docId,
        title: '示例文档标题',
        content: '这是文档的详细内容。在实际应用中，这里会显示从服务器获取的文档内容。',
        category: '技术文档',
        author: '张三',
        createdAt: '2023-05-15',
        updatedAt: '2023-05-20',
        status: 'published',
        version: '1.2'
      };

      // 模拟版本数据
      const mockVersions = [
        { id: '1', version: '1.0', createdAt: '2023-05-15', author: '张三' },
        { id: '2', version: '1.1', createdAt: '2023-05-18', author: '李四' },
        { id: '3', version: '1.2', createdAt: '2023-05-20', author: '王五' }
      ];

      this.setData({
        document: mockDocument,
        versions: mockVersions,
        loading: false
      });

      wx.hideLoading();
    }, 1000);
  },

  // 切换标签页
  switchTab: function (e) {
    const tab = e.currentTarget.dataset.tab;
    this.setData({
      activeTab: tab
    });
  },

  // 下载文档
  downloadDocument: function () {
    wx.showToast({
      title: '下载功能开发中',
      icon: 'none'
    });
  },

  // 分享文档
  shareDocument: function () {
    wx.showToast({
      title: '分享功能开发中',
      icon: 'none'
    });
  },

  // 查看版本历史
  viewVersionHistory: function () {
    wx.showToast({
      title: '版本历史功能开发中',
      icon: 'none'
    });
  }
});