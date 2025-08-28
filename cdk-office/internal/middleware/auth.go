/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireSuperAdmin 要求超级管理员权限的中间件
func RequireSuperAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 这里应该检查用户的JWT token和角色
		// 为了示例，我们简化处理

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "需要授权",
				"message": "请提供有效的授权token",
			})
			c.Abort()
			return
		}

		// TODO: 实际项目中应该验证JWT token并检查用户角色
		// 这里为了演示，我们简单检查header是否存在
		if authHeader != "Bearer admin-token" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "权限不足",
				"message": "需要超级管理员权限",
			})
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set("user_id", "admin-user-id")
		c.Set("user_role", "super_admin")

		c.Next()
	})
}

// RequireAuth 基础认证中间件
func RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "需要授权",
				"message": "请提供有效的授权token",
			})
			c.Abort()
			return
		}

		// TODO: 验证JWT token
		c.Set("user_id", "authenticated-user-id")
		c.Next()
	})
}
