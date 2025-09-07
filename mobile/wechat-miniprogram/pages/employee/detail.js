// pages/employee/detail.js
Page({
  data: {
    employee: null,
    loading: true,
    error: null
  },

  onLoad: function (options) {
    const employeeId = options.id;
    if (employeeId) {
      this.loadEmployeeDetail(employeeId);
    } else {
      this.setData({
        error: '员工ID不存在',
        loading: false
      });
    }
  },

  // 加载员工详情
  loadEmployeeDetail: function (employeeId) {
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
      // 模拟员工数据
      const mockEmployee = {
        id: employeeId,
        realName: '张三',
        username: 'zhangsan',
        email: 'zhangsan@example.com',
        phone: '13800138000',
        department: '技术部',
        position: '高级工程师',
        hireDate: '2020-01-15',
        status: '在职',
        idCard: '110101199001011234',
        address: '北京市朝阳区某某街道',
        emergencyContact: '李四',
        emergencyPhone: '13900139000'
      };

      this.setData({
        employee: mockEmployee,
        loading: false
      });

      wx.hideLoading();
    }, 1000);
  },

  // 编辑员工信息
  editEmployee: function () {
    wx.showToast({
      title: '编辑功能开发中',
      icon: 'none'
    });
  },

  // 离职操作
  terminateEmployee: function () {
    wx.showModal({
      title: '确认离职',
      content: '确定要为该员工办理离职手续吗？',
      success: (res) => {
        if (res.confirm) {
          wx.showToast({
            title: '离职功能开发中',
            icon: 'none'
          });
        }
        }
    });
  }
});