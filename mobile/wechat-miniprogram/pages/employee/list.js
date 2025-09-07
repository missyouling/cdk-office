// pages/employee/list.js
Page({
  data: {
    employees: [],
    loading: false,
    hasMore: true,
    page: 1,
    pageSize: 10,
    searchKeyword: ''
  },

  onLoad() {
    // 页面加载时获取员工列表
    this.loadEmployees();
  },

  onShow() {
    // 页面显示时刷新数据
    this.loadEmployees();
  },

  // 加载员工列表
  loadEmployees() {
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
      url: `${app.globalData.apiUrl}/employees`,
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
          const newEmployees = res.data.items || res.data;
          const hasMore = newEmployees.length === this.data.pageSize;
          
          this.setData({
            employees: this.data.page === 1 ? newEmployees : [...this.data.employees, ...newEmployees],
            hasMore: hasMore,
            page: this.data.page + 1
          });
        } else {
          wx.showToast({
            title: '获取员工列表失败',
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
      employees: []
    });
    this.loadEmployees();
    wx.stopPullDownRefresh();
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadEmployees();
    }
  },

  // 搜索员工
  onSearchInput(e) {
    this.setData({
      searchKeyword: e.detail.value
    });
  },

  // 执行搜索
  onSearch() {
    this.setData({
      page: 1,
      employees: []
    });
    this.loadEmployees();
  },

  // 跳转到员工详情
  goToEmployeeDetail(e) {
    const employeeId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./detail?id=${employeeId}`
    });
  }
})