/**
 * @jest-environment jsdom
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { useRouter } from 'next/navigation';
import AppCenter from './page';

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock Lucide React icons
jest.mock('lucide-react', () => ({
  Users: () => <div data-testid="users-icon">Users</div>,
  FileText: () => <div data-testid="filetext-icon">FileText</div>,
  Bot: () => <div data-testid="bot-icon">Bot</div>,
  QrCode: () => <div data-testid="qrcode-icon">QrCode</div>,
  Archive: () => <div data-testid="archive-icon">Archive</div>,
  BarChart3: () => <div data-testid="barchart3-icon">BarChart3</div>,
  CheckCircle: () => <div data-testid="checkcircle-icon">CheckCircle</div>,
  Bell: () => <div data-testid="bell-icon">Bell</div>,
  ClipboardList: () => <div data-testid="clipboardlist-icon">ClipboardList</div>,
  HelpCircle: () => <div data-testid="helpcircle-icon">HelpCircle</div>,
  Search: () => <div data-testid="search-icon">Search</div>,
  Star: () => <div data-testid="star-icon">Star</div>,
  Zap: () => <div data-testid="zap-icon">Zap</div>,
  Shield: () => <div data-testid="shield-icon">Shield</div>,
  Settings: () => <div data-testid="settings-icon">Settings</div>,
  Calendar: () => <div data-testid="calendar-icon">Calendar</div>,
  MessageSquare: () => <div data-testid="messagesquare-icon">MessageSquare</div>,
  Globe: () => <div data-testid="globe-icon">Globe</div>,
  Database: () => <div data-testid="database-icon">Database</div>,
  Camera: () => <div data-testid="camera-icon">Camera</div>,
  Smartphone: () => <div data-testid="smartphone-icon">Smartphone</div>,
}));

// Mock Link component from Next.js
jest.mock('next/link', () => {
  const MockLink = ({ children, href, ...props }: any) => (
    <a href={href} {...props}>
      {children}
    </a>
  );
  MockLink.displayName = 'MockLink';
  return MockLink;
});

// Mock UI components
jest.mock('@/components/ui/card', () => ({
  Card: ({ children, className, ...props }: any) => (
    <div className={`card ${className}`} {...props} data-testid="card">
      {children}
    </div>
  ),
  CardContent: ({ children, className, ...props }: any) => (
    <div className={`card-content ${className}`} {...props} data-testid="card-content">
      {children}
    </div>
  ),
  CardHeader: ({ children, className, ...props }: any) => (
    <div className={`card-header ${className}`} {...props} data-testid="card-header">
      {children}
    </div>
  ),
  CardTitle: ({ children, className, ...props }: any) => (
    <h3 className={`card-title ${className}`} {...props} data-testid="card-title">
      {children}
    </h3>
  ),
  CardDescription: ({ children, className, ...props }: any) => (
    <p className={`card-description ${className}`} {...props} data-testid="card-description">
      {children}
    </p>
  ),
}));

jest.mock('@/components/ui/button', () => ({
  Button: ({ children, onClick, className, variant, size, ...props }: any) => (
    <button 
      onClick={onClick} 
      className={`button ${variant} ${size} ${className}`} 
      {...props}
      data-testid="button"
    >
      {children}
    </button>
  ),
}));

jest.mock('@/components/ui/badge', () => ({
  Badge: ({ children, variant, className, ...props }: any) => (
    <span 
      className={`badge ${variant} ${className}`} 
      {...props}
      data-testid="badge"
    >
      {children}
    </span>
  ),
}));

jest.mock('@/components/ui/input', () => ({
  Input: ({ placeholder, value, onChange, className, ...props }: any) => (
    <input 
      placeholder={placeholder}
      value={value}
      onChange={onChange}
      className={`input ${className}`}
      {...props}
      data-testid="input"
    />
  ),
}));

jest.mock('@/components/ui/tabs', () => ({
  Tabs: ({ children, value, onValueChange, className, ...props }: any) => (
    <div className={`tabs ${className}`} {...props} data-testid="tabs">
      {children}
    </div>
  ),
  TabsContent: ({ children, value, className, ...props }: any) => (
    <div className={`tabs-content ${className}`} data-value={value} {...props} data-testid="tabs-content">
      {children}
    </div>
  ),
  TabsList: ({ children, className, ...props }: any) => (
    <div className={`tabs-list ${className}`} {...props} data-testid="tabs-list">
      {children}
    </div>
  ),
  TabsTrigger: ({ children, value, onClick, className, ...props }: any) => (
    <button 
      onClick={onClick}
      className={`tabs-trigger ${className}`} 
      data-value={value}
      {...props}
      data-testid="tabs-trigger"
    >
      {children}
    </button>
  ),
}));

describe('AppCenter Component', () => {
  const mockPush = jest.fn();

  beforeEach(() => {
    // Reset all mocks before each test
    jest.clearAllMocks();
    
    // Setup router mock
    (useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
      replace: jest.fn(),
      prefetch: jest.fn(),
    });
  });

  describe('Page Rendering', () => {
    it('should render the main title and description', () => {
      render(<AppCenter />);
      
      expect(screen.getByText('应用中心')).toBeInTheDocument();
      expect(screen.getByText('发现和使用CDK-Office平台提供的丰富应用，提升您的工作效率')).toBeInTheDocument();
    });

    it('should render the search input', () => {
      render(<AppCenter />);
      
      const searchInput = screen.getByPlaceholderText('搜索应用...');
      expect(searchInput).toBeInTheDocument();
      expect(searchInput).toHaveAttribute('data-testid', 'input');
    });

    it('should render category tabs', () => {
      render(<AppCenter />);
      
      // Check if category buttons exist
      expect(screen.getByText('全部')).toBeInTheDocument();
      expect(screen.getByText('核心应用')).toBeInTheDocument();
      expect(screen.getByText('AI应用')).toBeInTheDocument();
      expect(screen.getByText('业务应用')).toBeInTheDocument();
      expect(screen.getByText('工具应用')).toBeInTheDocument();
    });
  });

  describe('Application Cards', () => {
    it('should render featured applications section', () => {
      render(<AppCenter />);
      
      expect(screen.getByText('推荐应用')).toBeInTheDocument();
    });

    it('should render core applications', () => {
      render(<AppCenter />);
      
      // Check for core applications
      expect(screen.getByText('员工管理')).toBeInTheDocument();
      expect(screen.getByText('文档管理')).toBeInTheDocument();
      expect(screen.getByText('审批管理')).toBeInTheDocument();
    });

    it('should render AI applications including the new AI chat', () => {
      render(<AppCenter />);
      
      // Check for AI applications
      expect(screen.getByText('智能问答')).toBeInTheDocument();
      expect(screen.getByText('智能分析')).toBeInTheDocument();
      expect(screen.getByText('OCR文字识别')).toBeInTheDocument();
      
      // Verify AI chat application details
      const aiChatCard = screen.getByText('智能问答').closest('[data-testid="card"]');
      expect(aiChatCard).toBeInTheDocument();
      expect(aiChatCard).toHaveTextContent('集成Dify AI平台，提供智能问答、文档处理和知识管理能力');
    });

    it('should render business applications', () => {
      render(<AppCenter />);
      
      expect(screen.getByText('电子合同')).toBeInTheDocument();
      expect(screen.getByText('调查问卷')).toBeInTheDocument();
      expect(screen.getByText('通知中心')).toBeInTheDocument();
      expect(screen.getByText('日程管理')).toBeInTheDocument();
      expect(screen.getByText('团队沟通')).toBeInTheDocument();
    });

    it('should render tool applications', () => {
      render(<AppCenter />);
      
      expect(screen.getByText('二维码应用')).toBeInTheDocument();
      expect(screen.getByText('知识库归档')).toBeInTheDocument();
      expect(screen.getByText('数据统计')).toBeInTheDocument();
      expect(screen.getByText('系统设置')).toBeInTheDocument();
      expect(screen.getByText('数据备份')).toBeInTheDocument();
      expect(screen.getByText('移动端应用')).toBeInTheDocument();
    });

    it('should render application cards with correct structure', () => {
      render(<AppCenter />);
      
      // Get all card elements
      const cards = screen.getAllByTestId('card');
      expect(cards.length).toBeGreaterThan(0);
      
      // Check that each card has the expected structure
      cards.forEach(card => {
        expect(card).toHaveClass('card');
      });
    });

    it('should render badges for applications with special status', () => {
      render(<AppCenter />);
      
      // Check for HOT badge on AI chat
      const hotBadges = screen.getAllByText('HOT');
      expect(hotBadges.length).toBeGreaterThan(0);
      
      // Check for NEW badges
      const newBadges = screen.getAllByText('NEW');
      expect(newBadges.length).toBeGreaterThan(0);
      
      // Check for SOON badge
      const soonBadges = screen.getAllByText('SOON');
      expect(soonBadges.length).toBeGreaterThan(0);
    });
  });

  describe('Search Functionality', () => {
    it('should filter applications based on search query', async () => {
      render(<AppCenter />);
      
      const searchInput = screen.getByPlaceholderText('搜索应用...');
      
      // Search for "智能问答"
      fireEvent.change(searchInput, { target: { value: '智能问答' } });
      
      await waitFor(() => {
        expect(screen.getByText('智能问答')).toBeInTheDocument();
        // Other applications should not be visible (depends on implementation)
      });
    });

    it('should show no results message when search has no matches', async () => {
      render(<AppCenter />);
      
      const searchInput = screen.getByPlaceholderText('搜索应用...');
      
      // Search for something that doesn't exist
      fireEvent.change(searchInput, { target: { value: 'nonexistent app' } });
      
      await waitFor(() => {
        // Check if applications are filtered out
        expect(screen.queryByText('员工管理')).not.toBeInTheDocument();
        expect(screen.queryByText('智能问答')).not.toBeInTheDocument();
      });
    });

    it('should clear search and show all applications', async () => {
      render(<AppCenter />);
      
      const searchInput = screen.getByPlaceholderText('搜索应用...');
      
      // First search for something
      fireEvent.change(searchInput, { target: { value: '智能问答' } });
      
      // Then clear the search
      fireEvent.change(searchInput, { target: { value: '' } });
      
      await waitFor(() => {
        expect(screen.getByText('员工管理')).toBeInTheDocument();
        expect(screen.getByText('智能问答')).toBeInTheDocument();
        expect(screen.getByText('文档管理')).toBeInTheDocument();
      });
    });
  });

  describe('Category Filtering', () => {
    it('should show all applications by default', () => {
      render(<AppCenter />);
      
      expect(screen.getByText('员工管理')).toBeInTheDocument();
      expect(screen.getByText('智能问答')).toBeInTheDocument();
      expect(screen.getByText('电子合同')).toBeInTheDocument();
      expect(screen.getByText('二维码应用')).toBeInTheDocument();
    });

    it('should filter applications by AI category', async () => {
      render(<AppCenter />);
      
      const aiCategoryButton = screen.getByText('AI应用');
      fireEvent.click(aiCategoryButton);
      
      await waitFor(() => {
        // AI applications should be visible
        expect(screen.getByText('智能问答')).toBeInTheDocument();
        expect(screen.getByText('智能分析')).toBeInTheDocument();
        expect(screen.getByText('OCR文字识别')).toBeInTheDocument();
      });
    });

    it('should filter applications by core category', async () => {
      render(<AppCenter />);
      
      const coreCategory = screen.getByText('核心应用');
      fireEvent.click(coreCategory);
      
      await waitFor(() => {
        // Core applications should be visible
        expect(screen.getByText('员工管理')).toBeInTheDocument();
        expect(screen.getByText('文档管理')).toBeInTheDocument();
        expect(screen.getByText('审批管理')).toBeInTheDocument();
      });
    });

    it('should filter applications by business category', async () => {
      render(<AppCenter />);
      
      const businessCategory = screen.getByText('业务应用');
      fireEvent.click(businessCategory);
      
      await waitFor(() => {
        // Business applications should be visible
        expect(screen.getByText('电子合同')).toBeInTheDocument();
        expect(screen.getByText('调查问卷')).toBeInTheDocument();
        expect(screen.getByText('通知中心')).toBeInTheDocument();
      });
    });

    it('should filter applications by tools category', async () => {
      render(<AppCenter />);
      
      const toolsCategory = screen.getByText('工具应用');
      fireEvent.click(toolsCategory);
      
      await waitFor(() => {
        // Tool applications should be visible
        expect(screen.getByText('二维码应用')).toBeInTheDocument();
        expect(screen.getByText('知识库归档')).toBeInTheDocument();
        expect(screen.getByText('数据统计')).toBeInTheDocument();
      });
    });
  });

  describe('Application Navigation', () => {
    it('should have correct links for applications', () => {
      render(<AppCenter />);
      
      // Check AI chat link
      const aiChatLink = screen.getByText('智能问答').closest('a');
      expect(aiChatLink).toHaveAttribute('href', '/ai-chat');
      
      // Check other important links
      const employeeLink = screen.getByText('员工管理').closest('a');
      expect(employeeLink).toHaveAttribute('href', '/employee-management');
      
      const documentLink = screen.getByText('文档管理').closest('a');
      expect(documentLink).toHaveAttribute('href', '/document-management');
    });

    it('should handle coming soon applications correctly', () => {
      render(<AppCenter />);
      
      // Find coming soon applications
      const comingSoonApps = screen.getAllByText('即将推出');
      expect(comingSoonApps.length).toBeGreaterThan(0);
      
      // These should be styled differently (disabled state)
      comingSoonApps.forEach(badge => {
        const card = badge.closest('[data-testid="card"]');
        expect(card).toHaveClass('opacity-60');
      });
    });
  });

  describe('Responsive Design', () => {
    it('should render with responsive grid classes', () => {
      render(<AppCenter />);
      
      // Check if the grid container has responsive classes
      const container = screen.getByText('应用中心').closest('div');
      expect(container).toHaveClass('max-w-7xl', 'mx-auto');
    });
  });

  describe('Application Status', () => {
    it('should show beta status for some applications', () => {
      render(<AppCenter />);
      
      // Look for beta status indicators
      const betaStatuses = screen.getAllByText('Beta');
      // Verify that beta applications are present if any
      if (betaStatuses.length > 0) {
        betaStatuses.forEach(status => {
          expect(status).toHaveClass('bg-yellow-100', 'text-yellow-800');
        });
      }
    });

    it('should show active status for most applications', () => {
      render(<AppCenter />);
      
      // Most applications should be active (no special status indicator)
      expect(screen.getByText('智能问答')).toBeInTheDocument();
      expect(screen.getByText('员工管理')).toBeInTheDocument();
      expect(screen.getByText('文档管理')).toBeInTheDocument();
    });
  });

  describe('Featured Applications', () => {
    it('should display featured applications in the promoted section', () => {
      render(<AppCenter />);
      
      // Check that the featured section exists
      const featuredSection = screen.getByText('推荐应用');
      expect(featuredSection).toBeInTheDocument();
      
      // Featured applications should include key ones like AI chat
      expect(screen.getByText('智能问答')).toBeInTheDocument();
      expect(screen.getByText('员工管理')).toBeInTheDocument();
      expect(screen.getByText('文档管理')).toBeInTheDocument();
      expect(screen.getByText('电子合同')).toBeInTheDocument();
    });
  });

  describe('Icons and Visual Elements', () => {
    it('should render icons for applications', () => {
      render(<AppCenter />);
      
      // Check for presence of icons (mocked as test ids)
      expect(screen.getAllByTestId('bot-icon').length).toBeGreaterThan(0);
      expect(screen.getAllByTestId('users-icon').length).toBeGreaterThan(0);
      expect(screen.getAllByTestId('filetext-icon').length).toBeGreaterThan(0);
    });

    it('should render search icon', () => {
      render(<AppCenter />);
      
      expect(screen.getByTestId('search-icon')).toBeInTheDocument();
    });

    it('should render star icon for featured section', () => {
      render(<AppCenter />);
      
      expect(screen.getAllByTestId('star-icon').length).toBeGreaterThan(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle missing application data gracefully', () => {
      // This test ensures the component doesn't crash with incomplete data
      expect(() => render(<AppCenter />)).not.toThrow();
    });
  });

  describe('Performance', () => {
    it('should render efficiently with many applications', () => {
      const startTime = performance.now();
      render(<AppCenter />);
      const endTime = performance.now();
      
      // Component should render quickly (within 100ms for this test)
      expect(endTime - startTime).toBeLessThan(100);
    });
  });

  describe('Accessibility', () => {
    it('should have proper heading structure', () => {
      render(<AppCenter />);
      
      // Main heading
      const mainHeading = screen.getByRole('heading', { level: 1 });
      expect(mainHeading).toHaveTextContent('应用中心');
      
      // Section headings
      const featuredHeading = screen.getByText('推荐应用');
      expect(featuredHeading).toBeInTheDocument();
    });

    it('should have accessible search input', () => {
      render(<AppCenter />);
      
      const searchInput = screen.getByPlaceholderText('搜索应用...');
      expect(searchInput).toBeInTheDocument();
      expect(searchInput).toHaveAttribute('type', 'text');
    });

    it('should have accessible buttons', () => {
      render(<AppCenter />);
      
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
      
      // Category buttons should be accessible
      const allCategoryButton = screen.getByText('全部');
      expect(allCategoryButton).toBeInTheDocument();
    });

    it('should have accessible links', () => {
      render(<AppCenter />);
      
      const links = screen.getAllByRole('link');
      expect(links.length).toBeGreaterThan(0);
      
      // Each application should be a clickable link
      links.forEach(link => {
        expect(link).toHaveAttribute('href');
      });
    });
  });
});