// pages/business/list.js
Page({
  data: {
    modules: [],
    loading: false,
    hasMore: true,
    page: 1,
    pageSize: 10,
    searchKeyword: ''
  },

  onLoad() {
    // 页面加载时获取业务模块列表
    this.loadModules();
  },

  onShow() {
    // 页面显示时刷新数据
    this.loadModules();
  },

  // 加载业务模块列表
  loadModules() {
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
      url: `${app.globalData.apiUrl}/modules`,
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
          const newModules = res.data.items || res.data;
          const hasMore = newModules.length === this.data.pageSize;
          
          this.setData({
            modules: this.data.page === 1 ? newModules : [...this.data.modules, ...newModules],
            hasMore: hasMore,
            page: this.data.page + 1
          });
        } else {
          wx.showToast({
            title: '获取业务模块列表失败',
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
      modules: []
    });
    this.loadModules();
    wx.stopPullDownRefresh();
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadModules();
    }
  },

  // 搜索业务模块
  onSearchInput(e) {
    this.setData({
      searchKeyword: e.detail.value
    });
  },

  // 执行搜索
  onSearch() {
    this.setData({
      page: 1,
      modules: []
    });
    this.loadModules();
  },

  // 跳转到业务模块详情
  goToModuleDetail(e) {
    const moduleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./detail?id=${moduleId}`
    });
  },

  // 跳转到电子合同
  goToContract(e) {
    const moduleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./contract?id=${moduleId}`
    });
  },

  // 跳转到调查问卷
  goToSurvey(e) {
    const moduleId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./survey?id=${moduleId}`
    });
  }
})