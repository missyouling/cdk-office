// pages/index/index.js
Page({
  data: {
    userInfo: null,
    hasUserInfo: false,
    canIUseGetUserProfile: false,
    modules: [
      {
        id: 'document',
        name: '文档管理',
        icon: '../../images/icon_document.png',
        description: '查看和管理企业文档'
      },
      {
        id: 'employee',
        name: '员工管理',
        icon: '../../images/icon_employee.png',
        description: '查看组织架构和员工信息'
      },
      {
        id: 'application',
        name: '应用中心',
        icon: '../../images/icon_app.png',
        description: '使用企业应用和工具'
      }
    ]
  },

  onLoad() {
    // 检查是否可以使用 getUserProfile
    if (wx.canIUse('getUserProfile')) {
      this.setData({
        canIUseGetUserProfile: true
      })
    }
    
    // 检查用户登录状态
    this.checkLoginStatus();
  },

  onShow() {
    // 页面显示时检查登录状态
    this.checkLoginStatus();
  },

  // 检查登录状态
  checkLoginStatus() {
    const app = getApp();
    if (app.globalData.userInfo) {
      this.setData({
        userInfo: app.globalData.userInfo,
        hasUserInfo: true
      })
    } else if (wx.getStorageSync('token')) {
      // 从本地存储获取用户信息
      this.getUserInfo();
    }
  },

  // 获取用户信息
  getUserInfo() {
    const app = getApp();
    const token = wx.getStorageSync('token');
    
    if (!token) {
      return;
    }
    
    wx.request({
      url: `${app.globalData.apiUrl}/auth/user/profile`,
      method: 'GET',
      header: {
        'Authorization': `Bearer ${token}`
      },
      success: (res) => {
        if (res.statusCode === 200) {
          app.globalData.userInfo = res.data;
          this.setData({
            userInfo: res.data,
            hasUserInfo: true
          })
        }
      }
    })
  },

  // 获取用户授权信息
  getUserProfile(e) {
    // 推荐使用wx.getUserProfile获取用户信息，开发者每次通过该接口获取用户个人信息均需用户确认
    wx.getUserProfile({
      desc: '用于完善会员资料',
      success: (res) => {
        console.log('用户授权信息:', res);
        this.setData({
          userInfo: res.userInfo,
          hasUserInfo: true
        })
        getApp().globalData.userInfo = res.userInfo
      }
    })
  },

  // 跳转到登录页面
  goToLogin() {
    wx.navigateTo({
      url: '../auth/login'
    })
  },

  // 跳转到模块页面
  goToModule(e) {
    const moduleId = e.currentTarget.dataset.id;
    
    // 检查登录状态
    if (!this.data.hasUserInfo) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      this.goToLogin();
      return;
    }
    
    switch (moduleId) {
      case 'document':
        wx.navigateTo({
          url: '../document/list'
        })
        break;
      case 'employee':
        wx.navigateTo({
          url: '../employee/list'
        })
        break;
      case 'application':
        wx.showToast({
          title: '应用中心开发中',
          icon: 'none'
        })
        break;
      default:
        wx.showToast({
          title: '功能开发中',
          icon: 'none'
        })
    }
  },

  // 退出登录
  logout() {
    wx.showModal({
      title: '确认退出',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          // 清除本地存储
          wx.removeStorageSync('token');
          getApp().globalData.token = null;
          getApp().globalData.userInfo = null;
          
          // 更新页面数据
          this.setData({
            userInfo: null,
            hasUserInfo: false
          })
          
          wx.showToast({
            title: '已退出登录',
            icon: 'success'
          })
        }
      }
    })
  }
})