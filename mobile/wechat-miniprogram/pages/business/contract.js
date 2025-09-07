// pages/business/contract.js
Page({
  data: {
    contracts: [],
    loading: false,
    hasMore: true,
    page: 1,
    pageSize: 10,
    searchKeyword: ''
  },

  onLoad: function (options) {
    const moduleId = options.id;
    if (moduleId) {
      // 可以根据moduleId加载特定的合同列表
      console.log('Module ID:', moduleId);
    }
    // 页面加载时获取合同列表
    this.loadContracts();
  },

  onShow() {
    // 页面显示时刷新数据
    this.loadContracts();
  },

  // 加载合同列表
  loadContracts() {
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
    
    // 模拟API调用
    // 实际开发中需要替换为真实的API调用
    setTimeout(() => {
      // 模拟合同数据
      const mockContracts = [
        {
          id: '1',
          title: '劳动合同',
          partyA: '公司名称',
          partyB: '员工姓名',
          status: '已签署',
          createdAt: '2023-05-15',
          updatedAt: '2023-05-20'
        },
        {
          id: '2',
          title: '保密协议',
          partyA: '公司名称',
          partyB: '员工姓名',
          status: '待签署',
          createdAt: '2023-05-10',
          updatedAt: '2023-05-10'
        }
      ];

      const hasMore = mockContracts.length === this.data.pageSize;
      
      this.setData({
        contracts: this.data.page === 1 ? mockContracts : [...this.data.contracts, ...mockContracts],
        hasMore: hasMore,
        page: this.data.page + 1,
        loading: false
      });
    }, 1000);
  },

  // 下拉刷新
  onPullDownRefresh() {
    this.setData({
      page: 1,
      contracts: []
    });
    this.loadContracts();
    wx.stopPullDownRefresh();
  },

  // 上拉加载更多
  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadContracts();
    }
  },

  // 搜索合同
  onSearchInput(e) {
    this.setData({
      searchKeyword: e.detail.value
    });
  },

  // 执行搜索
  onSearch() {
    this.setData({
      page: 1,
      contracts: []
    });
    this.loadContracts();
  },

  // 跳转到合同详情
  goToContractDetail(e) {
    const contractId = e.currentTarget.dataset.id;
    wx.navigateTo({
      url: `./contract-detail?id=${contractId}`
    });
  },

  // 创建新合同
  createContract() {
    wx.navigateTo({
      url: './contract-create'
    });
  }
});