
import React from 'react';
import { cn } from '@/lib/utils';

interface SystemStatusIndicatorProps {
  statusName: string;
  isActive: boolean;
  className?: string;
}

export const SystemStatusIndicator: React.FC<SystemStatusIndicatorProps> = ({
  statusName,
  isActive,
  className,
}) => {
  const formattedName = statusName.replace('SystemStatusBits.', '');

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div
        className={cn('h-3 w-3 rounded-full', {
          'bg-green-500': isActive,
          'bg-red-500': !isActive,
        })}
      />
      <span className="text-sm font-medium text-muted-foreground">{formattedName}</span>
    </div>
  );
};
