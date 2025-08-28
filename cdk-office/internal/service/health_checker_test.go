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

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/linux-do/cdk-office/internal/models"
)

// MockDB 模拟数据库
type MockDB struct {
	mock.Mock
}

func (m *MockDB) WithContext(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB)
}

// HealthCheckerTestSuite 健康检查器测试套件
type HealthCheckerTestSuite struct {
	suite.Suite
	db           *gorm.DB
	logger       *logrus.Logger
	checker      *ServiceHealthChecker
	mockHTTPResp string
}

// SetupSuite 套件初始化
func (suite *HealthCheckerTestSuite) SetupSuite() {
	// 创建内存数据库用于测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移模型
	err = db.AutoMigrate(&models.ServiceHealthStatus{})
	suite.Require().NoError(err)

	suite.db = db

	// 创建测试日志器
	suite.logger = logrus.New()
	suite.logger.SetLevel(logrus.ErrorLevel) // 减少测试输出

	// 创建健康检查器实例
	suite.checker = NewServiceHealthChecker(db, suite.logger)
}

// SetupTest 每个测试前的设置
func (suite *HealthCheckerTestSuite) SetupTest() {
	// 清理数据库
	suite.db.Exec("DELETE FROM service_statuses")
}

// TestNewServiceHealthChecker 测试创建健康检查器实例
func (suite *HealthCheckerTestSuite) TestNewServiceHealthChecker() {
	checker := NewServiceHealthChecker(suite.db, suite.logger)

	assert.NotNil(suite.T(), checker)
	assert.Equal(suite.T(), suite.db, checker.db)
	assert.Equal(suite.T(), suite.logger, checker.logger)
	assert.NotNil(suite.T(), checker.client)
}

// TestGetServiceConfigs 测试获取服务配置
func (suite *HealthCheckerTestSuite) TestGetServiceConfigs() {
	configs := suite.checker.GetServiceConfigs()

	assert.Greater(suite.T(), len(configs), 0)

	// 验证配置结构
	for _, config := range configs {
		assert.NotEmpty(suite.T(), config.Name)
		assert.NotEmpty(suite.T(), config.Type)
		assert.NotEmpty(suite.T(), config.Endpoint)
		assert.Greater(suite.T(), config.Timeout, time.Duration(0))
	}

	// 验证特定服务存在
	foundDatabase := false
	foundRedis := false
	foundAIService := false

	for _, config := range configs {
		switch config.Type {
		case "database":
			foundDatabase = true
			assert.True(suite.T(), config.Critical)
		case "redis":
			foundRedis = true
			assert.True(suite.T(), config.Critical)
		case "ai_service":
			foundAIService = true
		}
	}

	assert.True(suite.T(), foundDatabase, "应该包含数据库服务配置")
	assert.True(suite.T(), foundRedis, "应该包含Redis服务配置")
	assert.True(suite.T(), foundAIService, "应该包含AI服务配置")
}

// TestPerformHealthCheckDatabase 测试数据库健康检查
func (suite *HealthCheckerTestSuite) TestPerformHealthCheckDatabase() {
	ctx := context.Background()
	config := ServiceConfig{
		Name:     "test_database",
		Type:     "database",
		Endpoint: "localhost:5432",
		Timeout:  5 * time.Second,
		Critical: true,
	}

	result := suite.checker.performHealthCheck(ctx, config)

	// 由于使用的是SQLite内存数据库，查询应该成功
	assert.Equal(suite.T(), "test_database", result.ServiceName)
	assert.Equal(suite.T(), "healthy", result.Status)
	assert.Equal(suite.T(), 200, result.StatusCode)
	assert.Empty(suite.T(), result.ErrorMessage)
	assert.True(suite.T(), result.ResponseTime > 0)
}

// TestPerformHealthCheckRedis 测试Redis健康检查
func (suite *HealthCheckerTestSuite) TestPerformHealthCheckRedis() {
	ctx := context.Background()
	config := ServiceConfig{
		Name:     "test_redis",
		Type:     "redis",
		Endpoint: "localhost:6379",
		Timeout:  3 * time.Second,
		Critical: true,
	}

	result := suite.checker.performHealthCheck(ctx, config)

	// 验证Redis检查结果
	assert.Equal(suite.T(), "test_redis", result.ServiceName)
	assert.NotEmpty(suite.T(), result.Status)
	assert.True(suite.T(), result.ResponseTime > 0)
}

// TestPerformHealthCheckUnsupportedType 测试不支持的服务类型
func (suite *HealthCheckerTestSuite) TestPerformHealthCheckUnsupportedType() {
	ctx := context.Background()
	config := ServiceConfig{
		Name:     "test_unknown",
		Type:     "unknown_type",
		Endpoint: "localhost:8080",
		Timeout:  5 * time.Second,
	}

	result := suite.checker.performHealthCheck(ctx, config)

	assert.Equal(suite.T(), "test_unknown", result.ServiceName)
	assert.Equal(suite.T(), "unhealthy", result.Status)
	assert.Contains(suite.T(), result.ErrorMessage, "不支持的服务类型")
}

// TestCheckAllServices 测试检查所有服务
func (suite *HealthCheckerTestSuite) TestCheckAllServices() {
	ctx := context.Background()

	results, err := suite.checker.CheckAllServices(ctx)

	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(results), 0)

	// 验证结果结构
	for _, result := range results {
		assert.NotEmpty(suite.T(), result.ServiceName)
		assert.NotEmpty(suite.T(), result.Status)
		assert.True(suite.T(), result.ResponseTime > 0)
		assert.False(suite.T(), result.LastChecked.IsZero())
	}

	// 验证数据库中保存了结果
	var count int64
	suite.db.Model(&models.ServiceHealthStatus{}).Count(&count)
	assert.Equal(suite.T(), int64(len(results)), count)
}

// TestGetServiceStatus 测试获取服务状态
func (suite *HealthCheckerTestSuite) TestGetServiceStatus() {
	ctx := context.Background()

	// 首先插入一些测试数据
	testStatus := &models.ServiceHealthStatus{
		ServiceName:  "test_service",
		Status:       "healthy",
		ResponseTime: 100,
		StatusCode:   200,
		CheckedAt:    time.Now(),
	}

	err := suite.db.Create(testStatus).Error
	assert.NoError(suite.T(), err)

	// 测试获取特定服务状态
	status, err := suite.checker.GetServiceStatus(ctx, "test_service")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), status)
	assert.Equal(suite.T(), "test_service", status.ServiceName)
	assert.Equal(suite.T(), "healthy", status.Status)

	// 测试获取不存在的服务状态
	_, err = suite.checker.GetServiceStatus(ctx, "nonexistent_service")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "服务状态记录未找到")
}

// TestGetAllServiceStatuses 测试获取所有服务状态
func (suite *HealthCheckerTestSuite) TestGetAllServiceStatuses() {
	ctx := context.Background()

	// 插入测试数据
	testStatuses := []*models.ServiceHealthStatus{
		{
			ServiceName:  "service1",
			Status:       "healthy",
			ResponseTime: 100,
			StatusCode:   200,
			CheckedAt:    time.Now(),
		},
		{
			ServiceName:  "service2",
			Status:       "degraded",
			ResponseTime: 600,
			StatusCode:   200,
			CheckedAt:    time.Now(),
		},
		{
			ServiceName:  "service1",
			Status:       "healthy",
			ResponseTime: 120,
			StatusCode:   200,
			CheckedAt:    time.Now().Add(-1 * time.Hour), // 旧记录
		},
	}

	for _, status := range testStatuses {
		err := suite.db.Create(status).Error
		assert.NoError(suite.T(), err)
	}

	// 获取所有最新状态
	statuses, err := suite.checker.GetAllServiceStatuses(ctx)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), statuses, 2) // 应该只返回每个服务的最新状态

	// 验证结果
	serviceNames := make(map[string]bool)
	for _, status := range statuses {
		serviceNames[status.ServiceName] = true
	}
	assert.True(suite.T(), serviceNames["service1"])
	assert.True(suite.T(), serviceNames["service2"])
}

// TestCleanupOldRecords 测试清理旧记录
func (suite *HealthCheckerTestSuite) TestCleanupOldRecords() {
	ctx := context.Background()

	// 插入新旧测试数据
	oldTime := time.Now().AddDate(0, 0, -40) // 40天前
	newTime := time.Now().AddDate(0, 0, -10) // 10天前

	testStatuses := []*models.ServiceHealthStatus{
		{
			ServiceName: "service1",
			Status:      "healthy",
			CheckedAt:   oldTime,
		},
		{
			ServiceName: "service2",
			Status:      "healthy",
			CheckedAt:   newTime,
		},
	}

	for _, status := range testStatuses {
		err := suite.db.Create(status).Error
		assert.NoError(suite.T(), err)
	}

	// 清理30天前的记录
	err := suite.checker.CleanupOldRecords(ctx, 30)
	assert.NoError(suite.T(), err)

	// 验证只剩下新记录
	var count int64
	suite.db.Model(&models.ServiceHealthStatus{}).Count(&count)
	assert.Equal(suite.T(), int64(1), count)

	// 验证剩下的是新记录
	var remaining models.ServiceHealthStatus
	err = suite.db.First(&remaining).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "service2", remaining.ServiceName)
}

// TestCalculateOverallHealth 测试计算总体健康状态
func (suite *HealthCheckerTestSuite) TestCalculateOverallHealth() {
	// 测试全部健康
	healthyStatuses := []models.ServiceHealthStatus{
		{ServiceName: "service1", Status: "healthy", ResponseTime: 100},
		{ServiceName: "service2", Status: "healthy", ResponseTime: 150},
	}

	summary := suite.checker.calculateOverallHealth(healthyStatuses)
	assert.Equal(suite.T(), "healthy", summary["overall_status"])
	assert.Equal(suite.T(), 2, summary["healthy_count"])
	assert.Equal(suite.T(), 0, summary["degraded_count"])
	assert.Equal(suite.T(), 0, summary["unhealthy_count"])

	// 测试有降级服务
	mixedStatuses := []models.ServiceHealthStatus{
		{ServiceName: "service1", Status: "healthy", ResponseTime: 100},
		{ServiceName: "service2", Status: "degraded", ResponseTime: 600},
	}

	summary = suite.checker.calculateOverallHealth(mixedStatuses)
	assert.Equal(suite.T(), "warning", summary["overall_status"])
	assert.Equal(suite.T(), 1, summary["healthy_count"])
	assert.Equal(suite.T(), 1, summary["degraded_count"])

	// 测试有关键服务不可用
	criticalDownStatuses := []models.ServiceHealthStatus{
		{ServiceName: "postgresql_database", Status: "unhealthy", ResponseTime: 0},
		{ServiceName: "service2", Status: "healthy", ResponseTime: 100},
	}

	summary = suite.checker.calculateOverallHealth(criticalDownStatuses)
	assert.Equal(suite.T(), "critical", summary["overall_status"])
	criticalDown := summary["critical_services_down"].([]string)
	assert.Contains(suite.T(), criticalDown, "postgresql_database")
}

// TestStartPeriodicHealthCheck 测试定期健康检查
func (suite *HealthCheckerTestSuite) TestStartPeriodicHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 启动定期检查（间隔很短用于测试）
	go suite.checker.StartPeriodicHealthCheck(ctx, 50*time.Millisecond)

	// 等待一段时间让检查执行
	time.Sleep(150 * time.Millisecond)

	// 验证有健康检查记录被创建
	var count int64
	suite.db.Model(&models.ServiceHealthStatus{}).Count(&count)
	assert.Greater(suite.T(), count, int64(0))
}

// TestConcurrentHealthChecks 测试并发健康检查
func (suite *HealthCheckerTestSuite) TestConcurrentHealthChecks() {
	ctx := context.Background()

	// 并发执行多次健康检查
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := suite.checker.CheckAllServices(ctx)
			assert.NoError(suite.T(), err)
			done <- true
		}()
	}

	// 等待所有并发检查完成
	for i := 0; i < 3; i++ {
		<-done
	}

	// 验证数据库中有记录（可能有重复，但应该有数据）
	var count int64
	suite.db.Model(&models.ServiceHealthStatus{}).Count(&count)
	assert.Greater(suite.T(), count, int64(0))
}

// TestErrorHandling 测试错误处理
func (suite *HealthCheckerTestSuite) TestErrorHandling() {
	ctx := context.Background()

	// 测试HTTP服务健康检查的错误处理
	config := ServiceConfig{
		Name:     "invalid_service",
		Type:     "ai_service",
		Endpoint: "http://nonexistent-domain-12345.com",
		Timeout:  1 * time.Second,
	}

	result := suite.checker.performHealthCheck(ctx, config)

	assert.Equal(suite.T(), "invalid_service", result.ServiceName)
	assert.Equal(suite.T(), "unhealthy", result.Status)
	assert.NotEmpty(suite.T(), result.ErrorMessage)
}

// TestResponseTimeThresholds 测试响应时间阈值
func (suite *HealthCheckerTestSuite) TestResponseTimeThresholds() {
	ctx := context.Background()

	// 测试数据库响应时间判断
	config := ServiceConfig{
		Name:     "test_db",
		Type:     "database",
		Endpoint: "localhost:5432",
		Timeout:  5 * time.Second,
	}

	result := suite.checker.performHealthCheck(ctx, config)

	// 内存数据库应该很快，所以应该是healthy
	if result.ResponseTime <= 500*time.Millisecond {
		assert.Equal(suite.T(), "healthy", result.Status)
	} else {
		assert.Equal(suite.T(), "degraded", result.Status)
	}
}

// 运行测试套件
func TestHealthCheckerTestSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckerTestSuite))
}

// 单独的基准测试
func BenchmarkCheckAllServices(b *testing.B) {
	// 创建测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	err = db.AutoMigrate(&models.ServiceHealthStatus{})
	if err != nil {
		b.Fatal(err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	checker := NewServiceHealthChecker(db, logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := checker.CheckAllServices(ctx)
		if err != nil {
			b.Error(err)
		}
	}
}

// 测试HTTP客户端配置
func (suite *HealthCheckerTestSuite) TestHTTPClientConfiguration() {
	assert.NotNil(suite.T(), suite.checker.client)
	assert.Equal(suite.T(), 30*time.Second, suite.checker.client.Timeout)

	// 验证Transport配置
	transport := suite.checker.client.Transport.(*http.Transport)
	assert.Equal(suite.T(), 10, transport.MaxIdleConns)
	assert.Equal(suite.T(), 30*time.Second, transport.IdleConnTimeout)
	assert.True(suite.T(), transport.DisableCompression)
	assert.Equal(suite.T(), 5, transport.MaxIdleConnsPerHost)
}
