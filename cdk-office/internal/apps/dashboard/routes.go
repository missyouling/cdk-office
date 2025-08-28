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

package dashboard

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册Dashboard路由
func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	// Dashboard统计
	r.GET("/dashboard/stats", handler.GetDashboardStats)

	// 待办事项路由
	todos := r.Group("/todos")
	{
		todos.GET("", handler.GetTodoItems)
		todos.POST("", handler.CreateTodoItem)
		todos.PATCH("/:id", handler.UpdateTodoItem)
		todos.DELETE("/:id", handler.DeleteTodoItem)
	}

	// 日程事件路由
	events := r.Group("/calendar-events")
	{
		events.GET("", handler.GetCalendarEvents)
		events.GET("/upcoming", handler.GetUpcomingEvents)
		events.POST("", handler.CreateCalendarEvent)
		events.PATCH("/:id", handler.UpdateCalendarEvent)
		events.DELETE("/:id", handler.DeleteCalendarEvent)
	}

	// 通知路由
	notifications := r.Group("/notifications")
	{
		notifications.GET("", handler.GetNotifications)
		notifications.PATCH("/:id/read", handler.MarkNotificationAsRead)
	}
}
