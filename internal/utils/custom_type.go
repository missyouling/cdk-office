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

package utils

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSON 自定义JSON类型
type JSON json.RawMessage

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}

	*j = JSON(bytes)
	return nil
}

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// String 转换为字符串
func (j JSON) String() string {
	return string(j)
}

// MarshalJSON 实现json.Marshaler接口
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSON: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Timestamp 自定义时间戳类型
type Timestamp time.Time

// Scan 实现sql.Scanner接口
func (t *Timestamp) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*t = Timestamp(v)
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		*t = Timestamp(parsed)
	case []byte:
		parsed, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return err
		}
		*t = Timestamp(parsed)
	default:
		return fmt.Errorf("cannot scan %T into Timestamp", value)
	}

	return nil
}

// Value 实现driver.Valuer接口
func (t Timestamp) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// MarshalJSON 实现json.Marshaler接口
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format("2006-01-02 15:04:05"))
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return err
	}

	*t = Timestamp(parsed)
	return nil
}

// String 转换为字符串
func (t Timestamp) String() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

// Unix 返回Unix时间戳
func (t Timestamp) Unix() int64 {
	return time.Time(t).Unix()
}

// UUID 自定义UUID类型
type UUID string

// IsValid 检查UUID是否有效
func (u UUID) IsValid() bool {
	// 简单检查UUID格式
	s := string(u)
	if len(s) != 36 {
		return false
	}

	// 检查基本格式
	// 这里可以添加更严格的UUID格式验证
	return true
}

// String 转换为字符串
func (u UUID) String() string {
	return string(u)
}

// IsEmpty 检查是否为空
func (u UUID) IsEmpty() bool {
	return string(u) == ""
}
