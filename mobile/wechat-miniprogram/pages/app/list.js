// pages/app/list.js
Page({
  data: {
    apps: [],
    loading: false,
    hasMore: true,
    page: 1,
    pageSize: 10,
    searchKeyword: ''
  },

  onLoad() {
    // 页面加载时获取应用列表
    this.loadApps();
  },

  onShow() {
    // 页面显示时刷新数据
    this.loadApps();
  },

  // 加载应用列表
  loadApps() {
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
      url: `${app.globalData.apiUrl}/apps`,
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
          const newApps = res.data.items || res.data;
          const hasMore = newApps.length === this.data.pageSize;
          
          this.setData({
            apps: this.data.page === 1 ? newApps : [...this.data.apps, ...newApps],
            hasMore: hasMore,
            page: this.data.page + 1
          });
        } else {
          wx.showToast({
            title: '获取应用列表失败',
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
      apps: []
    });
    this.loadApps();
    wx.stopPullDownRefresh();
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadApps();
    }
  },

  // 搜索应用
  onSearchInput(e) {
    this.setData({
      searchKeyword: e.detail.value
    });
  },

  // 执行搜索
  onSearch() {
    this.setData({
      page: 1,
      apps: []
    });
    this.loadApps();
  },

  // 跳转到应用详情
  goToAppDetail(e) {
    const appId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./detail?id=${appId}`
    });
  },

  // 跳转到二维码页面
  goToQRCode(e) {
    const appId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./qrcode?id=${appId}`
    });
  },

  // 跳转到表单页面
  goToForm(e) {
    const appId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./form?id=${appId}`
    });
  }
})