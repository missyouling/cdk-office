// pages/document/list.js
Page({
  data: {
    documents: [],
    loading: false,
    hasMore: true,
    page: 1,
    pageSize: 10,
    searchKeyword: ''
  },

  onLoad() {
    // 页面加载时获取文档列表
    this.loadDocuments();
  },

  onShow() {
    // 页面显示时刷新数据
    this.loadDocuments();
  },

  // 加载文档列表
  loadDocuments() {
    if (this.data.loading || !this.data.hasMore) {
      return;
    }
    
    this.setData({
      loading: true
    });
    
    const app = getApp();
    const token = wx.getStorageSync('token');
    
    if (!token) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      wx.switchTab({
        url: '../auth/login'
      });
      return;
    }
    
    wx.request({
      url: `${app.globalData.apiUrl}/documents`,
      method: 'GET',
      header: {
        'Authorization': `Bearer ${token}`
      },
      data: {
        page: this.data.page,
        size: this.data.pageSize,
        keyword: this.data.searchKeyword
      },
      success: (res) => {
        if (res.statusCode === 200) {
          const newDocuments = res.data.items || res.data;
          const hasMore = newDocuments.length === this.data.pageSize;
          
          this.setData({
            documents: this.data.page === 1 ? newDocuments : [...this.data.documents, ...newDocuments],
            hasMore: hasMore,
            page: this.data.page + 1
          });
        } else {
          wx.showToast({
            title: '获取文档列表失败',
            icon: 'none'
          });
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '网络错误',
          icon: 'none'
        });
      },
      complete: () => {
        this.setData({
          loading: false
        });
      }
    });
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.setData({
      page: 1,
      documents: []
    });
    this.loadDocuments();
    wx.stopPullDownRefresh();
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadDocuments();
    }
  },

  // 搜索文档
  onSearchInput(e) {
    this.setData({
      searchKeyword: e.detail.value
    });
  },

  // 执行搜索
  onSearch() {
    this.setData({
      page: 1,
      documents: []
    });
    this.loadDocuments();
  },

  // 跳转到文档详情
  goToDocumentDetail(e) {
    const documentId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./detail?id=${documentId}`
    });
  },

  // 格式化文件大小
  formatFileSize(size) {
    if (size < 1024) {
      return size + ' B';
    } else if (size < 1024 * 1024) {
      return (size / 1024).toFixed(2) + ' KB';
    } else if (size < 1024 * 1024 * 1024) {
      return (size / (1024 * 1024)).toFixed(2) + ' MB';
    } else {
      return (size / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    }
  }
})