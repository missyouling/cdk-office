// pages/business/detail.js
Page({
  data: {
    module: null,
    loading: true,
    error: null
  },

  onLoad: function (options) {
    const moduleId = options.id;
    if (moduleId) {
      this.loadModuleDetail(moduleId);
    } else {
      this.setData({
        error: '模块ID不存在',
        loading: false
      });
    }
  },

  // 加载业务模块详情
  loadModuleDetail: function (moduleId) {
    wx.showLoading({ title: '加载中...' });
    
    const app = getApp();
    const token = wx.getStorageSync('token');
    
    if (!token) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      wx.redirectTo({
        url: '../auth/login'
      });
      return;
    }
    
    // 模拟API调用
    // 实际开发中需要替换为真实的API调用
    setTimeout(() => {
      // 模拟业务模块数据
      const mockModule = {
        id: moduleId,
        name: '电子合同',
        description: '电子合同签署平台，支持在线签署和管理合同',
        icon: '../../images/default_module_icon.png',
        type: 'contract',
        version: '1.0.0',
        developer: '业务部',
        createdAt: '2023-05-15',
        updatedAt: '2023-05-20',
        status: 'published'
      };

      this.setData({
        module: mockModule,
        loading: false
      });

      wx.hideLoading();
    }, 1000);
  },

  // 使用模块
  useModule: function () {
    const module = this.data.module;
    if (module.type === 'contract') {
      wx.navigateTo({
        url: `./contract?id=${module.id}`
      });
    } else if (module.type === 'survey') {
      wx.navigateTo({
        url: `./survey?id=${module.id}`
      });
    } else {
      wx.showToast({
        title: '未知模块类型',
        icon: 'none'
      });
    }
  },

  // 编辑模块
  editModule: function () {
    wx.showToast({
      title: '编辑功能开发中',
      icon: 'none'
    });
  },

  // 删除模块
  deleteModule: function () {
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该模块吗？',
      success: (res) => {
        if (res.confirm) {
          wx.showToast({
            title: '删除功能开发中',
            icon: 'none'
          });
        }
      }
    });
  }
});