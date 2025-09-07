// pages/auth/register.js
Page({
  data: {
    username: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
    isRegistering: false
  },

  // 输入用户名
  onUsernameInput(e) {
    this.setData({
      username: e.detail.value
    });
  },

  // 输入邮箱
  onEmailInput(e) {
    this.setData({
      email: e.detail.value
    });
  },

  // 输入手机号
  onPhoneInput(e) {
    this.setData({
      phone: e.detail.value
    });
  },

  // 输入密码
  onPasswordInput(e) {
    this.setData({
      password: e.detail.value
    });
  },

  // 输入确认密码
  onConfirmPasswordInput(e) {
    this.setData({
      confirmPassword: e.detail.value
    });
  },

  // 注册按钮点击事件
  onRegister() {
    const { username, email, phone, password, confirmPassword } = this.data;
    
    // 验证输入
    if (!username) {
      wx.showToast({
        title: '请输入用户名',
        icon: 'none'
      });
      return;
    }
    
    if (!email) {
      wx.showToast({
        title: '请输入邮箱',
        icon: 'none'
      });
      return;
    }
    
    // 简单的邮箱格式验证
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      wx.showToast({
        title: '请输入正确的邮箱格式',
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
    
    if (password.length < 6) {
      wx.showToast({
        title: '密码长度至少6位',
        icon: 'none'
      });
      return;
    }
    
    if (password !== confirmPassword) {
      wx.showToast({
        title: '两次输入的密码不一致',
        icon: 'none'
      });
      return;
    }
    
    // 显示加载提示
    this.setData({
      isRegistering: true
    });
    
    // 调用后端注册接口
    const app = getApp();
    wx.request({
      url: `${app.globalData.apiUrl}/auth/register`,
      method: 'POST',
      data: {
        username: username,
        email: email,
        phone: phone,
        password: password
      },
      success: (res) => {
        if (res.statusCode === 200) {
          // 注册成功
          this.setData({
            isRegistering: false
          });
          
          wx.showToast({
            title: '注册成功',
            icon: 'success'
          });
          
          // 延迟跳转到登录页面
          setTimeout(() => {
            wx.navigateBack();
          }, 1000);
        } else {
          // 注册失败
          this.setData({
            isRegistering: false
          });
          
          wx.showToast({
            title: res.data.error || '注册失败',
            icon: 'none'
          });
        }
      },
      fail: (err) => {
        // 网络错误
        this.setData({
          isRegistering: false
        });
        
        wx.showToast({
          title: '网络错误',
          icon: 'none'
        });
      }
    });
  },

  // 跳转到登录页面
  goToLogin() {
    wx.navigateBack();
  }
})