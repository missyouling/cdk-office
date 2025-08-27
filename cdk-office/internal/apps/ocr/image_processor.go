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

package ocr

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"math"
)

// ImageProcessor 图像处理器
type ImageProcessor struct {
	config *ImageProcessorConfig
}

// ImageProcessorConfig 图像处理器配置
type ImageProcessorConfig struct {
	EnableDenoising             bool    `json:"enable_denoising"`              // 启用降噪
	EnableSharpening            bool    `json:"enable_sharpening"`             // 启用锐化
	EnableContrastAdjust        bool    `json:"enable_contrast_adjust"`        // 启用对比度调整
	EnableBrightnessAdjust      bool    `json:"enable_brightness_adjust"`      // 启用亮度调整
	EnablePerspectiveCorrection bool    `json:"enable_perspective_correction"` // 启用透视矫正
	AutoDetectEdges             bool    `json:"auto_detect_edges"`             // 自动检测边缘
	ContrastFactor              float64 `json:"contrast_factor"`               // 对比度因子 (0.5-2.0)
	BrightnessFactor            float64 `json:"brightness_factor"`             // 亮度因子 (-100 到 100)
	SharpnessFactor             float64 `json:"sharpness_factor"`              // 锐化因子 (0.0-2.0)
	DenoisingStrength           float64 `json:"denoising_strength"`            // 降噪强度 (0.0-1.0)
}

// NewImageProcessor 创建图像处理器
func NewImageProcessor(config *ImageProcessorConfig) *ImageProcessor {
	if config == nil {
		config = &ImageProcessorConfig{
			EnableDenoising:             true,
			EnableSharpening:            true,
			EnableContrastAdjust:        true,
			EnableBrightnessAdjust:      true,
			EnablePerspectiveCorrection: true,
			AutoDetectEdges:             true,
			ContrastFactor:              1.2,
			BrightnessFactor:            10,
			SharpnessFactor:             1.1,
			DenoisingStrength:           0.3,
		}
	}
	return &ImageProcessor{config: config}
}

// ProcessImage 处理图像
func (p *ImageProcessor) ProcessImage(ctx context.Context, imageData []byte) ([]byte, *ImageProcessingResult, error) {
	// 解码图像
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode image: %w", err)
	}

	result := &ImageProcessingResult{
		OriginalFormat:  format,
		ProcessingSteps: []ProcessingStep{},
	}

	// 应用处理步骤
	processedImg := img

	// 1. 降噪
	if p.config.EnableDenoising {
		processedImg = p.applyDenoising(processedImg)
		result.ProcessingSteps = append(result.ProcessingSteps, ProcessingStep{
			Name:    "denoising",
			Applied: true,
			Parameters: map[string]interface{}{
				"strength": p.config.DenoisingStrength,
			},
		})
	}

	// 2. 亮度调整
	if p.config.EnableBrightnessAdjust {
		processedImg = p.adjustBrightness(processedImg, p.config.BrightnessFactor)
		result.ProcessingSteps = append(result.ProcessingSteps, ProcessingStep{
			Name:    "brightness_adjust",
			Applied: true,
			Parameters: map[string]interface{}{
				"factor": p.config.BrightnessFactor,
			},
		})
	}

	// 3. 对比度调整
	if p.config.EnableContrastAdjust {
		processedImg = p.adjustContrast(processedImg, p.config.ContrastFactor)
		result.ProcessingSteps = append(result.ProcessingSteps, ProcessingStep{
			Name:    "contrast_adjust",
			Applied: true,
			Parameters: map[string]interface{}{
				"factor": p.config.ContrastFactor,
			},
		})
	}

	// 4. 透视矫正
	if p.config.EnablePerspectiveCorrection {
		if correctedImg, applied := p.correctPerspective(processedImg); applied {
			processedImg = correctedImg
			result.ProcessingSteps = append(result.ProcessingSteps, ProcessingStep{
				Name:    "perspective_correction",
				Applied: true,
				Parameters: map[string]interface{}{
					"auto_detect": p.config.AutoDetectEdges,
				},
			})
		}
	}

	// 5. 锐化
	if p.config.EnableSharpening {
		processedImg = p.applySharpen(processedImg, p.config.SharpnessFactor)
		result.ProcessingSteps = append(result.ProcessingSteps, ProcessingStep{
			Name:    "sharpening",
			Applied: true,
			Parameters: map[string]interface{}{
				"factor": p.config.SharpnessFactor,
			},
		})
	}

	// 编码输出图像
	var outputBuf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&outputBuf, processedImg, &jpeg.Options{Quality: 95})
	case "png":
		err = png.Encode(&outputBuf, processedImg)
	default:
		// 默认使用JPEG
		err = jpeg.Encode(&outputBuf, processedImg, &jpeg.Options{Quality: 95})
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	result.OutputFormat = format
	result.InputSize = len(imageData)
	result.OutputSize = outputBuf.Len()
	result.CompressionRatio = float64(result.InputSize) / float64(result.OutputSize)

	log.Printf("Image processing completed: %d steps applied, size %d -> %d",
		len(result.ProcessingSteps), result.InputSize, result.OutputSize)

	return outputBuf.Bytes(), result, nil
}

// applyDenoising 应用降噪
func (p *ImageProcessor) applyDenoising(img image.Image) image.Image {
	// 简化的高斯模糊降噪实现
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建输出图像
	outputImg := image.NewRGBA(bounds)

	// 高斯核（3x3）
	kernel := [][]float64{
		{0.077847, 0.123317, 0.077847},
		{0.123317, 0.195346, 0.123317},
		{0.077847, 0.123317, 0.077847},
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var r, g, b float64

			// 应用卷积核
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					px := img.At(x+kx, y+ky)
					pr, pg, pb, _ := px.RGBA()
					weight := kernel[ky+1][kx+1]

					r += float64(pr>>8) * weight
					g += float64(pg>>8) * weight
					b += float64(pb>>8) * weight
				}
			}

			// 设置像素
			outputImg.Set(x, y, image.RGBA{
				R: uint8(math.Max(0, math.Min(255, r))),
				G: uint8(math.Max(0, math.Min(255, g))),
				B: uint8(math.Max(0, math.Min(255, b))),
				A: 255,
			})
		}
	}

	return outputImg
}

// adjustBrightness 调整亮度
func (p *ImageProcessor) adjustBrightness(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	outputImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 调整亮度
			newR := float64(r>>8) + factor
			newG := float64(g>>8) + factor
			newB := float64(b>>8) + factor

			// 限制在有效范围内
			newR = math.Max(0, math.Min(255, newR))
			newG = math.Max(0, math.Min(255, newG))
			newB = math.Max(0, math.Min(255, newB))

			outputImg.Set(x, y, image.RGBA{
				R: uint8(newR),
				G: uint8(newG),
				B: uint8(newB),
				A: uint8(a >> 8),
			})
		}
	}

	return outputImg
}

// adjustContrast 调整对比度
func (p *ImageProcessor) adjustContrast(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	outputImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 调整对比度 (以128为中心点)
			newR := 128 + (float64(r>>8)-128)*factor
			newG := 128 + (float64(g>>8)-128)*factor
			newB := 128 + (float64(b>>8)-128)*factor

			// 限制在有效范围内
			newR = math.Max(0, math.Min(255, newR))
			newG = math.Max(0, math.Min(255, newG))
			newB = math.Max(0, math.Min(255, newB))

			outputImg.Set(x, y, image.RGBA{
				R: uint8(newR),
				G: uint8(newG),
				B: uint8(newB),
				A: uint8(a >> 8),
			})
		}
	}

	return outputImg
}

// correctPerspective 透视矫正
func (p *ImageProcessor) correctPerspective(img image.Image) (image.Image, bool) {
	// 简化的透视矫正实现
	// 在实际应用中，这里会使用更复杂的算法，如Hough变换检测直线

	if !p.config.AutoDetectEdges {
		return img, false
	}

	// 检测文档边缘
	edges := p.detectDocumentEdges(img)
	if edges == nil {
		return img, false
	}

	// 应用透视变换
	correctedImg := p.applyPerspectiveTransform(img, edges)
	return correctedImg, true
}

// detectDocumentEdges 检测文档边缘
func (p *ImageProcessor) detectDocumentEdges(img image.Image) []image.Point {
	// 简化实现：假设图像中心区域是文档
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算边缘点（简化为矩形）
	margin := 0.05 // 5%边距
	leftMargin := int(float64(width) * margin)
	topMargin := int(float64(height) * margin)
	rightMargin := width - leftMargin
	bottomMargin := height - topMargin

	// 返回四个角点
	return []image.Point{
		{leftMargin, topMargin},     // 左上
		{rightMargin, topMargin},    // 右上
		{rightMargin, bottomMargin}, // 右下
		{leftMargin, bottomMargin},  // 左下
	}
}

// applyPerspectiveTransform 应用透视变换
func (p *ImageProcessor) applyPerspectiveTransform(img image.Image, corners []image.Point) image.Image {
	// 简化实现：这里只做简单的裁剪
	bounds := img.Bounds()

	if len(corners) != 4 {
		return img
	}

	// 计算最小边界矩形
	minX, minY := corners[0].X, corners[0].Y
	maxX, maxY := corners[0].X, corners[0].Y

	for _, corner := range corners[1:] {
		if corner.X < minX {
			minX = corner.X
		}
		if corner.X > maxX {
			maxX = corner.X
		}
		if corner.Y < minY {
			minY = corner.Y
		}
		if corner.Y > maxY {
			maxY = corner.Y
		}
	}

	// 裁剪图像
	cropBounds := image.Rect(minX, minY, maxX, maxY)
	croppedImg := image.NewRGBA(cropBounds)

	for y := cropBounds.Min.Y; y < cropBounds.Max.Y; y++ {
		for x := cropBounds.Min.X; x < cropBounds.Max.X; x++ {
			if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
				croppedImg.Set(x, y, img.At(x, y))
			}
		}
	}

	return croppedImg
}

// applySharpen 应用锐化
func (p *ImageProcessor) applySharpen(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	outputImg := image.NewRGBA(bounds)

	// 锐化卷积核
	kernel := [][]float64{
		{0, -1 * factor, 0},
		{-1 * factor, 1 + 4*factor, -1 * factor},
		{0, -1 * factor, 0},
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var r, g, b float64

			// 应用卷积核
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					px := img.At(x+kx, y+ky)
					pr, pg, pb, _ := px.RGBA()
					weight := kernel[ky+1][kx+1]

					r += float64(pr>>8) * weight
					g += float64(pg>>8) * weight
					b += float64(pb>>8) * weight
				}
			}

			// 限制在有效范围内
			r = math.Max(0, math.Min(255, r))
			g = math.Max(0, math.Min(255, g))
			b = math.Max(0, math.Min(255, b))

			outputImg.Set(x, y, image.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: 255,
			})
		}
	}

	return outputImg
}

// AnalyzeImageQuality 分析图像质量
func (p *ImageProcessor) AnalyzeImageQuality(imageData []byte) (*ImageQualityAnalysis, error) {
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	analysis := &ImageQualityAnalysis{
		ImageFormat: format,
		Width:       width,
		Height:      height,
		AspectRatio: float64(width) / float64(height),
		FileSize:    len(imageData),
		Metrics:     make(map[string]float64),
		Suggestions: []string{},
	}

	// 计算亮度
	brightness := p.calculateBrightness(img)
	analysis.Metrics["brightness"] = brightness

	// 计算对比度
	contrast := p.calculateContrast(img)
	analysis.Metrics["contrast"] = contrast

	// 计算清晰度
	sharpness := p.calculateSharpness(img)
	analysis.Metrics["sharpness"] = sharpness

	// 计算噪声水平
	noiseLevel := p.calculateNoiseLevel(img)
	analysis.Metrics["noise_level"] = noiseLevel

	// 生成建议
	if brightness < 80 {
		analysis.Suggestions = append(analysis.Suggestions, "图像亮度较低，建议增加亮度")
	}
	if brightness > 200 {
		analysis.Suggestions = append(analysis.Suggestions, "图像亮度过高，建议降低亮度")
	}
	if contrast < 50 {
		analysis.Suggestions = append(analysis.Suggestions, "图像对比度较低，建议增加对比度")
	}
	if sharpness < 0.3 {
		analysis.Suggestions = append(analysis.Suggestions, "图像模糊，建议应用锐化")
	}
	if noiseLevel > 0.6 {
		analysis.Suggestions = append(analysis.Suggestions, "图像噪声较多，建议应用降噪")
	}

	// 计算总体质量评分
	analysis.OverallScore = p.calculateOverallScore(analysis.Metrics)

	return analysis, nil
}

// calculateBrightness 计算平均亮度
func (p *ImageProcessor) calculateBrightness(img image.Image) float64 {
	bounds := img.Bounds()
	var totalBrightness float64
	var pixelCount int

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// 使用感知亮度公式
			brightness := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			totalBrightness += brightness
			pixelCount++
		}
	}

	return totalBrightness / float64(pixelCount)
}

// calculateContrast 计算对比度
func (p *ImageProcessor) calculateContrast(img image.Image) float64 {
	bounds := img.Bounds()
	var values []float64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			values = append(values, gray)
		}
	}

	// 计算标准差作为对比度指标
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))

	return math.Sqrt(variance)
}

// calculateSharpness 计算清晰度
func (p *ImageProcessor) calculateSharpness(img image.Image) float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var totalVariance float64
	var edgeCount int

	// 使用Sobel算子检测边缘
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// 计算灰度值
			gray := func(px, py int) float64 {
				r, g, b, _ := img.At(px, py).RGBA()
				return 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			}

			// Sobel X方向
			gx := -1*gray(x-1, y-1) + 0*gray(x, y-1) + 1*gray(x+1, y-1) +
				-2*gray(x-1, y) + 0*gray(x, y) + 2*gray(x+1, y) +
				-1*gray(x-1, y+1) + 0*gray(x, y+1) + 1*gray(x+1, y+1)

			// Sobel Y方向
			gy := -1*gray(x-1, y-1) + -2*gray(x, y-1) + -1*gray(x+1, y-1) +
				0*gray(x-1, y) + 0*gray(x, y) + 0*gray(x+1, y) +
				1*gray(x-1, y+1) + 2*gray(x, y+1) + 1*gray(x+1, y+1)

			// 计算梯度幅值
			magnitude := math.Sqrt(gx*gx + gy*gy)
			totalVariance += magnitude
			edgeCount++
		}
	}

	if edgeCount == 0 {
		return 0
	}

	return totalVariance / float64(edgeCount) / 255.0 // 归一化到0-1
}

// calculateNoiseLevel 计算噪声水平
func (p *ImageProcessor) calculateNoiseLevel(img image.Image) float64 {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var totalVariance float64
	var regionCount int

	// 分析局部区域的方差来估计噪声
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// 计算3x3区域的方差
			var values []float64
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					r, g, b, _ := img.At(x+dx, y+dy).RGBA()
					gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
					values = append(values, gray)
				}
			}

			// 计算局部方差
			mean := 0.0
			for _, v := range values {
				mean += v
			}
			mean /= float64(len(values))

			variance := 0.0
			for _, v := range values {
				variance += (v - mean) * (v - mean)
			}
			variance /= float64(len(values))

			totalVariance += variance
			regionCount++
		}
	}

	if regionCount == 0 {
		return 0
	}

	// 归一化噪声水平
	noiseLevel := totalVariance / float64(regionCount) / (255.0 * 255.0)
	return math.Min(1.0, noiseLevel)
}

// calculateOverallScore 计算总体质量评分
func (p *ImageProcessor) calculateOverallScore(metrics map[string]float64) float64 {
	// 加权计算总体评分
	weights := map[string]float64{
		"brightness":  0.2,
		"contrast":    0.3,
		"sharpness":   0.3,
		"noise_level": 0.2,
	}

	score := 0.0
	totalWeight := 0.0

	for metric, weight := range weights {
		if value, exists := metrics[metric]; exists {
			normalizedValue := value

			// 根据指标类型进行归一化
			switch metric {
			case "brightness":
				// 亮度最佳范围是100-150
				if value >= 100 && value <= 150 {
					normalizedValue = 1.0
				} else {
					normalizedValue = 1.0 - math.Abs(value-125)/125
				}
			case "contrast":
				// 对比度越高越好，但要有上限
				normalizedValue = math.Min(1.0, value/100)
			case "sharpness":
				// 清晰度直接使用
				normalizedValue = value
			case "noise_level":
				// 噪声越低越好
				normalizedValue = 1.0 - value
			}

			score += normalizedValue * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return math.Max(0, math.Min(1, score/totalWeight)) * 100 // 转换为0-100分
}

// ImageProcessingResult 图像处理结果
type ImageProcessingResult struct {
	OriginalFormat   string           `json:"original_format"`
	OutputFormat     string           `json:"output_format"`
	InputSize        int              `json:"input_size"`
	OutputSize       int              `json:"output_size"`
	CompressionRatio float64          `json:"compression_ratio"`
	ProcessingSteps  []ProcessingStep `json:"processing_steps"`
}

// ProcessingStep 处理步骤
type ProcessingStep struct {
	Name       string                 `json:"name"`
	Applied    bool                   `json:"applied"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ImageQualityAnalysis 图像质量分析
type ImageQualityAnalysis struct {
	ImageFormat  string             `json:"image_format"`
	Width        int                `json:"width"`
	Height       int                `json:"height"`
	AspectRatio  float64            `json:"aspect_ratio"`
	FileSize     int                `json:"file_size"`
	OverallScore float64            `json:"overall_score"` // 0-100
	Metrics      map[string]float64 `json:"metrics"`
	Suggestions  []string           `json:"suggestions"`
}
