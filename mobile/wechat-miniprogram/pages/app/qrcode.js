// pages/app/qrcode.js
Page({
  data: {
    content: '',
    qrCodeUrl: '',
    loading: false,
    error: null
  },

  onLoad: function (options) {
    const appId = options.id;
    if (appId) {
      // 可以根据appId加载特定的二维码配置
      console.log('App ID:', appId);
    }
  },

  // 输入内容
  onContentInput: function (e) {
    this.setData({
      content: e.detail.value
    });
  },

  // 生成二维码
  generateQRCode: function () {
    if (!this.data.content.trim()) {
      wx.showToast({
        title: '请输入内容',
        icon: 'none'
      });
      return;
    }

    this.setData({
      loading: true,
      error: null
    });

    // 模拟生成二维码
    // 实际开发中需要调用后端API生成二维码
    setTimeout(() => {
      // 模拟二维码URL
      const mockQRCodeUrl = 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=' + encodeURIComponent(this.data.content);
      
      this.setData({
        qrCodeUrl: mockQRCodeUrl,
        loading: false
      });

      wx.showToast({
        title: '二维码生成成功',
        icon: 'success'
      });
    }, 1000);
  },

  // 保存二维码
  saveQRCode: function () {
    if (!this.data.qrCodeUrl) {
      wx.showToast({
        title: '请先生成二维码',
        icon: 'none'
      });
      return;
    }

    wx.downloadFile({
      url: this.data.qrCodeUrl,
      success: (res) => {
        if (res.statusCode === 200) {
          wx.saveImageToPhotosAlbum({
            filePath: res.tempFilePath,
            success: () => {
              wx.showToast({
                title: '保存成功',
                icon: 'success'
              });
            },
            fail: (err) => {
              wx.showToast({
                title: '保存失败',
                icon: 'none'
              });
            }
          });
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '下载失败',
          icon: 'none'
        });
      }
    });
  },

  // 分享二维码
  shareQRCode: function () {
    if (!this.data.qrCodeUrl) {
      wx.showToast({
        title: '请先生成二维码',
        icon: 'none'
      });
      return;
    }

    wx.downloadFile({
      url: this.data.qrCodeUrl,
      success: (res) => {
        if (res.statusCode === 200) {
          wx.shareAppMessage({
            title: '二维码分享',
            imageUrl: res.tempFilePath
          });
        }
      },
      fail: (err) => {
        wx.showToast({
          title: '分享失败',
          icon: 'none'
        });
      }
    });
  }
});