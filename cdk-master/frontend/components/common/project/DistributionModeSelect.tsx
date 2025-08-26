'use client';

import {Label} from '@/components/ui/label';
import {User, Lock} from 'lucide-react';
import {DistributionType} from '@/lib/services/project/types';

interface DistributionModeSelectProps {
  distributionType: DistributionType;
  onDistributionTypeChange: (type: DistributionType) => void;
}

export function DistributionModeSelect({
  distributionType,
  onDistributionTypeChange,
}: DistributionModeSelectProps) {
  return (
    <div className="space-y-3">
      <Label>分发模式</Label>
      <div className="grid grid-cols-2 gap-3">
        <div
          className={`relative cursor-pointer rounded-lg border-2 p-4 transition-all ${
            distributionType === DistributionType.ONE_FOR_EACH ?
              'border-primary bg-primary/5' :
              'border-border hover:border-primary/50'
          }`}
          onClick={() => onDistributionTypeChange(DistributionType.ONE_FOR_EACH)}
        >
          <div className="flex items-center gap-3">
            <div
              className={`rounded-full p-2 ${
                distributionType === DistributionType.ONE_FOR_EACH ?
                  'bg-primary text-primary-foreground' :
                  'bg-muted text-muted-foreground'
              }`}
            >
              <User className="h-4 w-4" />
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-sm">一码一用</h3>
              <p className="text-muted-foreground text-xs mt-0.5">
                邀请码仅供一人使用
              </p>
            </div>
            {distributionType === DistributionType.ONE_FOR_EACH && (
              <div className="h-2 w-2 rounded-full bg-primary"></div>
            )}
          </div>
        </div>

        <div
          className={`relative cursor-pointer rounded-lg border-2 p-4 transition-all ${
            distributionType === DistributionType.INVITE ?
              'border-primary bg-primary/5' :
              'border-border hover:border-primary/50'
          }`}
          onClick={() => onDistributionTypeChange(DistributionType.INVITE)}
        >
          <div className="flex items-center gap-3">
            <div
              className={`rounded-full p-2 ${
                distributionType === DistributionType.INVITE ?
                  'bg-primary text-primary-foreground' :
                  'bg-muted text-muted-foreground'
              }`}
            >
              <Lock className="h-4 w-4" />
            </div>
            <div className="flex-1">
              <h3 className="font-medium text-sm">手动邀请</h3>
              <p className="text-muted-foreground text-xs mt-0.5">
                用户填写内容后发放
              </p>
            </div>
            {distributionType === DistributionType.INVITE && (
              <div className="h-2 w-2 rounded-full bg-primary"></div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
