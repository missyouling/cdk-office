// pages/app/detail.js
Page({
  data: {
    app: null,
    loading: true,
    error: null
  },

  onLoad: function (options) {
    const appId = options.id;
    if (appId) {
      this.loadAppDetail(appId);
    } else {
      this.setData({
        error: '应用ID不存在',
        loading: false
      });
    }
  },

  // 加载应用详情
  loadAppDetail: function (appId) {
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
      // 模拟应用数据
      const mockApp = {
        id: appId,
        name: '批量二维码生成',
        description: '批量生成二维码应用，支持自定义内容和样式',
        icon: '../../images/default_app_icon.png',
        type: 'qrcode',
        version: '1.0.0',
        developer: '技术部',
        createdAt: '2023-05-15',
        updatedAt: '2023-05-20',
        status: 'published'
      };

      this.setData({
        app: mockApp,
        loading: false
      });

      wx.hideLoading();
    }, 1000);
  },

  // 使用应用
  useApp: function () {
    const app = this.data.app;
    if (app.type === 'qrcode') {
      wx.navigateTo({
        url: `./qrcode?id=${app.id}`
      });
    } else if (app.type === 'form') {
      wx.navigateTo({
        url: `./form?id=${app.id}`
      });
    } else if (app.type === 'data-collection') {
      wx.navigateTo({
        url: `./data-collection?id=${app.id}`
      });
    } else {
      wx.showToast({
        title: '未知应用类型',
        icon: 'none'
      });
    }
  },

  // 编辑应用
  editApp: function () {
    wx.showToast({
      title: '编辑功能开发中',
      icon: 'none'
    });
  },

  // 删除应用
  deleteApp: function () {
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该应用吗？',
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