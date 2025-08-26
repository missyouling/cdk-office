'use client';

import React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import ApprovalDetailPage from '../../pages/ApprovalDetail';

const ApprovalDetailRoute: React.FC = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const approvalId = searchParams.get('id') || '';

  if (!approvalId) {
    return <div>缺少审批ID参数</div>;
  }

  return <ApprovalDetailPage approvalId={approvalId} />;
};

export default ApprovalDetailRoute;