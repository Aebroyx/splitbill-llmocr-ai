import { CheckCircleIcon, ClockIcon, XCircleIcon, ArrowPathIcon } from '@heroicons/react/24/outline';

interface BillStatusIndicatorProps {
  status: string;
  isPolling?: boolean;
  lastUpdated?: Date | null;
  onRetry?: () => void;
}

export default function BillStatusIndicator({ status, isPolling, lastUpdated, onRetry }: BillStatusIndicatorProps) {
  const getStatusConfig = () => {
    switch (status) {
      case 'pending':
        return {
          icon: ClockIcon,
          color: 'text-gray-500',
          bgColor: 'bg-gray-100',
          text: 'Pending',
          description: 'Waiting to start processing'
        };
      case 'processing':
        return {
          icon: ClockIcon,
          color: 'text-blue-600',
          bgColor: 'bg-blue-100',
          text: 'Processing',
          description: 'AI is analyzing your bill image'
        };
      case 'completed':
        return {
          icon: CheckCircleIcon,
          color: 'text-green-600',
          bgColor: 'bg-green-100',
          text: 'Completed',
          description: 'Bill items extracted successfully'
        };
      case 'failed':
        return {
          icon: XCircleIcon,
          color: 'text-red-600',
          bgColor: 'bg-red-100',
          text: 'Failed',
          description: 'Failed to process bill image'
        };
      default:
        return {
          icon: ClockIcon,
          color: 'text-gray-500',
          bgColor: 'bg-gray-100',
          text: 'Unknown',
          description: 'Unknown status'
        };
    }
  };

  const config = getStatusConfig();
  const IconComponent = config.icon;

  return (
    <div className={`w-full ${config.bgColor} rounded-lg p-4`}>
      <div className="flex flex-col sm:flex-row sm:items-center gap-3">
        {/* Status Icon and Text */}
        <div className="flex items-center gap-2 flex-shrink-0">
          <IconComponent className={`w-5 h-5 ${config.color}`} />
          <span className={`font-medium ${config.color}`}>{config.text}</span>
        </div>
        
        {/* Processing Spinner */}
        {isPolling && status === 'processing' && (
          <div className="flex items-center gap-2 flex-shrink-0">
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
            <span className="text-sm text-blue-600">Processing...</span>
          </div>
        )}
        
        {/* Retry Button */}
        {status === 'failed' && onRetry && (
          <button
            onClick={onRetry}
            className="flex items-center gap-2 px-3 py-1 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors flex-shrink-0"
          >
            <ArrowPathIcon className="w-4 h-4" />
            Try Again
          </button>
        )}
        
        {/* Last Updated Timestamp */}
        {lastUpdated && (
          <span className="text-xs text-gray-500 flex-shrink-0">
            Last updated: {lastUpdated.toLocaleTimeString()}
          </span>
        )}
      </div>
      
      {/* Description - Always on new line for better readability */}
      <div className="text-sm text-gray-600 mt-2 break-words">
        {config.description}
      </div>
    </div>
  );
}
