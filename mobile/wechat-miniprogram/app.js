App({
  onLaunch() {
    // 小程序初始化时执行
    console.log('CDK-Office 小程序启动');
    
    // 检查登录状态
    this.checkLoginStatus();
  },
  
  onShow() {
    // 小程序启动，或从后台进入前台显示时执行
    console.log('CDK-Office 小程序显示');
  },
  
  onHide() {
    // 小程序从前台进入后台时执行
    console.log('CDK-Office 小程序隐藏');
  },
  
  onError(msg) {
    // 小程序发生脚本错误，或者 api 调用失败时触发
    console.log('CDK-Office 小程序错误:', msg);
  },
  
  globalData: {
    userInfo: null,
    token: null,
    apiUrl: 'http://localhost:8080/api/v1' // 后端API地址
  },
  
  // 检查登录状态
  checkLoginStatus() {
    const token = wx.getStorageSync('token');
    if (token) {
      this.globalData.token = token;
      // 验证token有效性
      this.verifyToken(token);
    }
  },
  
  // 验证token有效性
  verifyToken(token) {
    wx.request({
      url: `${this.globalData.apiUrl}/auth/user/profile`,
      method: 'GET',
      header: {
        'Authorization': `Bearer ${token}`
      },
      success: (res) => {
        if (res.statusCode === 200) {
          this.globalData.userInfo = res.data;
        } else {
          // token无效，清除本地存储
          wx.removeStorageSync('token');
          this.globalData.token = null;
          this.globalData.userInfo = null;
        }
      },
      fail: () => {
        // 网络错误，保留现有状态
        console.log('网络错误，无法验证token');
      }
    });
  },
  
  // 登录方法
  login(username, password, successCallback, errorCallback) {
    wx.request({
      url: `${this.globalData.apiUrl}/auth/login`,
      method: 'POST',
      data: {
        username: username,
        password: password
      },
      success: (res) => {
        if (res.statusCode === 200) {
          // 保存token到本地存储
          wx.setStorageSync('token', res.data.token);
          this.globalData.token = res.data.token;
          this.globalData.userInfo = res.data.user;
          successCallback(res.data);
        } else {
          errorCallback(res.data.error);
        }
      },
      fail: (err) => {
        errorCallback('网络错误');
      }
    });
  },
  
  // 微信登录方法
  wechatLogin(code, successCallback, errorCallback) {
    wx.request({
      url: `${this.globalData.apiUrl}/auth/wechat/login`,
      method: 'POST',
      data: {
        code: code
      },
      success: (res) => {
        if (res.statusCode === 200) {
          // 保存token到本地存储
          wx.setStorageSync('token', res.data.token);
          this.globalData.token = res.data.token;
          this.globalData.userInfo = res.data.user;
          successCallback(res.data);
        } else {
          errorCallback(res.data.error);
        }
      },
      fail: (err) => {
        errorCallback('网络错误');
      }
    });
  }
})