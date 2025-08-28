/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 */

'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn } from '@/lib/utils';
import {
  Camera,
  ScanLine,
  Upload,
  Settings,
  Eye,
  Grid3X3,
  FlashOn,
  FlashOff,
  Loader2,
  Archive,
  X,
} from 'lucide-react';

import {
  ScannedDocument,
  ScanSession,
  CapturedImage,
  DocumentProcessOptions,
  MobilePermission,
  CameraConfig,
} from '@/types/scanner';
import { scannerService } from '@/services/scanner';

// 相机组件
function CameraView({ onImageCapture, cameraConfig, onConfigChange }: any) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isStreaming, setIsStreaming] = useState(false);
  const [flashEnabled, setFlashEnabled] = useState(false);

  useEffect(() => {
    startCamera();
    return () => stopCamera();
  }, []);

  const startCamera = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'environment' }
      });
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
        setIsStreaming(true);
      }
    } catch (error) {
      console.error('Camera access failed:', error);
    }
  };

  const stopCamera = () => {
    if (videoRef.current?.srcObject) {
      const stream = videoRef.current.srcObject as MediaStream;
      stream.getTracks().forEach(track => track.stop());
    }
  };

  const captureImage = () => {
    if (!videoRef.current || !canvasRef.current) return;

    const canvas = canvasRef.current;
    const video = videoRef.current;
    const context = canvas.getContext('2d');

    if (!context) return;

    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    context.drawImage(video, 0, 0, canvas.width, canvas.height);

    canvas.toBlob((blob) => {
      if (blob) {
        const file = new File([blob], `scan_${Date.now()}.jpg`, { type: 'image/jpeg' });
        onImageCapture(file);
      }
    }, 'image/jpeg', 0.9);
  };

  return (
    <div className="relative w-full h-full bg-black rounded-lg overflow-hidden">
      <video ref={videoRef} autoPlay playsInline muted className="w-full h-full object-cover" />
      <canvas ref={canvasRef} className="hidden" />
      
      {/* 文档边框 */}
      <div className="absolute inset-4 border-2 border-blue-500/50 rounded-lg pointer-events-none">
        <div className="absolute top-0 left-0 w-6 h-6 border-t-2 border-l-2 border-blue-500" />
        <div className="absolute top-0 right-0 w-6 h-6 border-t-2 border-r-2 border-blue-500" />
        <div className="absolute bottom-0 left-0 w-6 h-6 border-b-2 border-l-2 border-blue-500" />
        <div className="absolute bottom-0 right-0 w-6 h-6 border-b-2 border-r-2 border-blue-500" />
      </div>
      
      {/* 控制按钮 */}
      <div className="absolute top-4 left-4 flex space-x-2">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => setFlashEnabled(!flashEnabled)}
          className="bg-black/50 text-white"
        >
          {flashEnabled ? <FlashOn className="h-4 w-4" /> : <FlashOff className="h-4 w-4" />}
        </Button>
        <Button
          variant="secondary"
          size="sm"
          className="bg-black/50 text-white"
        >
          <Grid3X3 className="h-4 w-4" />
        </Button>
      </div>
      
      {/* 拍摄按钮 */}
      <div className="absolute bottom-8 left-1/2 transform -translate-x-1/2">
        <Button
          size="lg"
          onClick={captureImage}
          disabled={!isStreaming}
          className="w-16 h-16 rounded-full bg-white"
        >
          <Camera className="h-6 w-6 text-gray-800" />
        </Button>
      </div>
    </div>
  );
}

// 图像预览组件
function ImagePreview({ images, onImageSelect, onImageDelete, selectedImage }: any) {
  if (images.length === 0) {
    return (
      <Card className="p-8 text-center">
        <ScanLine className="h-12 w-12 mx-auto text-gray-400 mb-4" />
        <h3 className="text-lg font-medium">暂无扫描图像</h3>
        <p className="text-gray-500">开始拍摄文档图像</p>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {selectedImage && (
        <Card>
          <CardContent className="p-4">
            <img
              src={selectedImage.processed_uri || selectedImage.image_uri}
              alt="Selected"
              className="w-full h-64 object-contain rounded-lg bg-gray-100"
            />
          </CardContent>
        </Card>
      )}
      
      <div className="grid grid-cols-4 gap-2">
        {images.map((image: any, index: number) => (
          <div
            key={image.id}
            className={cn(
              "relative cursor-pointer rounded-lg overflow-hidden border-2",
              selectedImage?.id === image.id ? "border-blue-500" : "border-gray-200"
            )}
            onClick={() => onImageSelect(image)}
          >
            <img src={image.thumbnail_uri} alt={`Scan ${index + 1}`} className="w-full h-16 object-cover" />
            <Badge variant="secondary" className="absolute top-1 left-1 text-xs py-0 px-1">
              {index + 1}
            </Badge>
            <Button
              variant="destructive"
              size="sm"
              className="absolute top-1 right-1 w-5 h-5 p-0"
              onClick={(e) => {
                e.stopPropagation();
                onImageDelete(image.id);
              }}
            >
              <X className="h-3 w-3" />
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}

// 主组件
const DocumentScanner: React.FC = () => {
  const [activeTab, setActiveTab] = useState('camera');
  const [currentSession, setCurrentSession] = useState<ScanSession | null>(null);
  const [capturedImages, setCapturedImages] = useState<CapturedImage[]>([]);
  const [selectedImage, setSelectedImage] = useState<CapturedImage | undefined>();
  const [permissions, setPermissions] = useState<MobilePermission | null>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadToKB, setUploadToKB] = useState(false);
  
  const [processingOptions, setProcessingOptions] = useState<DocumentProcessOptions>({
    perspective_correction: true,
    brightness_adjustment: true,
    contrast_enhancement: true,
    noise_reduction: false,
    text_enhancement: true,
    auto_crop: true,
    deskew: true,
    shadow_removal: false,
  });

  const [cameraConfig, setCameraConfig] = useState<CameraConfig>({
    resolution: 'high',
    flash_mode: 'auto',
    focus_mode: 'auto',
    enable_grid: true,
    enable_edge_detection: true,
    capture_format: 'jpeg',
    quality: 90,
  });

  useEffect(() => {
    loadPermissions();
    createNewSession();
  }, []);

  const loadPermissions = async () => {
    try {
      const perms = await scannerService.checkPermissions();
      setPermissions(perms);
      setUploadToKB(perms.has_personal_kb_access);
    } catch (error) {
      console.error('Failed to load permissions:', error);
    }
  };

  const createNewSession = async () => {
    try {
      const session = await scannerService.createScanSession(
        `扫描会话_${new Date().toLocaleString()}`
      );
      setCurrentSession(session);
      setCapturedImages([]);
    } catch (error) {
      console.error('Failed to create session:', error);
    }
  };

  const handleImageCapture = async (imageFile: File) => {
    if (!currentSession) return;

    try {
      const capturedImage = await scannerService.addImageToSession(
        currentSession.id,
        imageFile
      );
      setCapturedImages([...capturedImages, capturedImage]);
      setSelectedImage(capturedImage);
      setActiveTab('preview');
    } catch (error) {
      console.error('Failed to add image:', error);
    }
  };

  const handleImageDelete = (imageId: string) => {
    setCapturedImages(capturedImages.filter(img => img.id !== imageId));
    if (selectedImage?.id === imageId) {
      setSelectedImage(capturedImages[0]);
    }
  };

  const handleProcessAndUpload = async () => {
    if (!currentSession || capturedImages.length === 0) return;

    setUploading(true);
    try {
      const task = await scannerService.processBatchScan({
        session_id: currentSession.id,
        processing_options: processingOptions,
        output_format: 'pdf',
        merge_to_single_pdf: true,
        document_name: `扫描文档_${new Date().toLocaleString()}`,
        to_personal_kb: uploadToKB && (permissions?.has_personal_kb_access || false),
      });
      
      alert(uploadToKB ? '文档已处理并上传到个人知识库' : '文档处理完成');
      createNewSession();
      setActiveTab('camera');
    } catch (error) {
      console.error('Failed to process document:', error);
      alert('文档处理失败');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="w-full h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <ScanLine className="h-6 w-6 text-blue-600" />
            <h1 className="text-xl font-bold">文档扫描</h1>
          </div>
          {permissions && (
            <Badge variant="outline">
              剩余 {permissions.daily_scan_limit - permissions.used_scan_count} 次
            </Badge>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
          <TabsList className="grid w-full grid-cols-3 mx-4 mt-2">
            <TabsTrigger value="camera">
              <Camera className="h-4 w-4 mr-2" />
              拍摄
            </TabsTrigger>
            <TabsTrigger value="preview">
              <Eye className="h-4 w-4 mr-2" />
              预览 ({capturedImages.length})
            </TabsTrigger>
            <TabsTrigger value="process">
              <Settings className="h-4 w-4 mr-2" />
              处理
            </TabsTrigger>
          </TabsList>

          <div className="flex-1 p-4">
            <TabsContent value="camera" className="h-full">
              <CameraView
                onImageCapture={handleImageCapture}
                cameraConfig={cameraConfig}
                onConfigChange={setCameraConfig}
              />
            </TabsContent>

            <TabsContent value="preview">
              <ImagePreview
                images={capturedImages}
                onImageSelect={setSelectedImage}
                onImageDelete={handleImageDelete}
                selectedImage={selectedImage}
              />
            </TabsContent>

            <TabsContent value="process" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>处理选项</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    {Object.entries(processingOptions).map(([key, value]) => (
                      <div key={key} className="flex items-center space-x-2">
                        <input
                          type="checkbox"
                          id={key}
                          checked={value}
                          onChange={(e) => setProcessingOptions({
                            ...processingOptions,
                            [key]: e.target.checked
                          })}
                          className="rounded"
                        />
                        <label htmlFor={key} className="text-sm">
                          {key.replace(/_/g, ' ')}
                        </label>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>上传选项</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="upload_to_kb"
                      checked={uploadToKB}
                      onChange={(e) => setUploadToKB(e.target.checked)}
                      disabled={!permissions?.has_personal_kb_access}
                    />
                    <label htmlFor="upload_to_kb" className="text-sm">
                      上传到个人知识库
                    </label>
                  </div>
                </CardContent>
              </Card>

              <Button
                className="w-full"
                onClick={handleProcessAndUpload}
                disabled={capturedImages.length === 0 || uploading}
              >
                {uploading ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    处理中...
                  </>
                ) : (
                  <>
                    <Upload className="h-4 w-4 mr-2" />
                    处理并上传 ({capturedImages.length} 张)
                  </>
                )}
              </Button>
            </TabsContent>
          </div>
        </Tabs>
      </div>
    </div>
  );
};

export default DocumentScanner;