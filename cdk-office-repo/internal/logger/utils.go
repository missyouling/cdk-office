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

package logger

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// LogWithTrace 记录带链路追踪信息的日志
func LogWithTrace(ctx context.Context, level, format string, v ...interface{}) {
	// 获取链路追踪ID
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()
	
	// 获取调用者信息
	_, file, line, _ := runtime.Caller(1)
	
	// 构建日志消息
	message := fmt.Sprintf("[%s] [trace_id=%s] [span_id=%s] [file=%s:%d] %s",
		level,
		traceID.String(),
		spanID.String(),
		file,
		line,
		fmt.Sprintf(format, v...),
	)
	
	// 记录日志
	Logger.Println(message)
}

// LogWithTraceID 记录带指定追踪ID的日志
func LogWithTraceID(traceID, spanID, level, format string, v ...interface{}) {
	// 获取调用者信息
	_, file, line, _ := runtime.Caller(1)
	
	// 构建日志消息
	message := fmt.Sprintf("[%s] [trace_id=%s] [span_id=%s] [file=%s:%d] %s",
		level,
		traceID,
		spanID,
		file,
		line,
		fmt.Sprintf(format, v...),
	)
	
	// 记录日志
	Logger.Println(message)
}

// LogWithDuration 记录带执行时间的日志
func LogWithDuration(start time.Time, operation string) {
	duration := time.Since(start)
	Logger.Printf("[INFO] [duration=%v] %s completed", duration, operation)
}