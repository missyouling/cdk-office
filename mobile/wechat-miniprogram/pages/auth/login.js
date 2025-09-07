// pages/auth/login.js
Page({
  data: {
    username: '',
    password: '',
    isLoggingIn: false
  },

  onLoad() {
    // 页面加载时检查是否已登录
    const token = wx.getStorageSync('token');
    if (token) {
      // 如果已登录，跳转到首页
      wx.switchTab({
        url: '../index/index'
      });
    }
  },

  // 输入用户名
  onUsernameInput(e) {
    this.setData({
      username: e.detail.value
    });
  },

  // 输入密码
  onPasswordInput(e) {
    this.setData({
      password: e.detail.value
    });
  },

  // 登录按钮点击事件
  onLogin() {
    const { username, password } = this.data;
    
    // 验证输入
    if (!username) {
      wx.showToast({
        title: '请输入用户名',
        icon: 'none'
      });
      return;
    }
    
    if (!password) {
      wx.showToast({
        title: '请输入密码',
        icon: 'none'
      });
      return;
    }
    
    // 显示加载提示
    this.setData({
      isLoggingIn: true
    });
    
    // 调用应用实例的登录方法
    const app = getApp();
    app.login(username, password, (res) => {
      // 登录成功
      this.setData({
        isLoggingIn: false
      });
      
      wx.showToast({
        title: '登录成功',
        icon: 'success'
      });
      
      // 延迟跳转到首页
      setTimeout(() => {
        wx.switchTab({
          url: '../index/index'
        });
      }, 1000);
    }, (error) => {
      // 登录失败
      this.setData({
        isLoggingIn: false
      });
      
      wx.showToast({
        title: error || '登录失败',
        icon: 'none'
      });
    });
  },

  // 微信登录
  onWechatLogin() {
    wx.login({
      success: (res) => {
        if (res.code) {
          // 显示加载提示
          this.setData({
            isLoggingIn: true
          });
          
          // 调用应用实例的微信登录方法
          const app = getApp();
          app.wechatLogin(res.code, (res) => {
            // 登录成功
            this.setData({
              isLoggingIn: false
            });
            
            wx.showToast({
              title: '登录成功',
              icon: 'success'
            });
            
            // 延迟跳转到首页
            setTimeout(() => {
              wx.switchTab({
                url: '../index/index'
              });
            }, 1000);
          }, (error) => {
            // 登录失败
            this.setData({
              isLoggingIn: false
            });
            
            wx.showToast({
              title: error || '微信登录失败',
              icon: 'none'
            });
          });
        } else {
          console.log('登录失败！' + res.errMsg);
          wx.showToast({
            title: '微信登录失败',
            icon: 'none'
          });
        }
      }
    });
  },

  // 跳转到注册页面
  goToRegister() {
    wx.navigateTo({
      url: './register'
    });
  }
})