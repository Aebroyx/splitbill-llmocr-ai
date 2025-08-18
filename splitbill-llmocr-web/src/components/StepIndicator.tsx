'use client';

import React from 'react';

interface StepIndicatorProps {
  currentStep: number;
  totalSteps: number;
  stepLabels?: string[];
  className?: string;
}

export default function StepIndicator({ 
  currentStep, 
  totalSteps, 
  stepLabels = [],
  className = '' 
}: StepIndicatorProps) {
  return (
    <div className={`w-full ${className}`}>
      {/* Progress Bars - Instagram Stories Style */}
      <div className="flex gap-2 mb-3">
        {Array.from({ length: totalSteps }, (_, index) => (
          <div
            key={index}
            className="flex-1 h-1.5 rounded-full overflow-hidden bg-gray-200"
          >
            <div
              className={`h-full transition-all duration-700 ease-out ${
                index < currentStep 
                  ? 'bg-primary w-full' 
                  : index === currentStep 
                    ? 'bg-secondary animate-pulse' 
                    : 'bg-gray-200 w-0'
              }`}
              style={{
                width: index < currentStep ? '100%' : index === currentStep ? '100%' : '0%'
              }}
            />
          </div>
        ))}
      </div>
      
      {/* Step Labels - Position at start of each progress bar */}
      {stepLabels.length > 0 && (
        <div className="flex gap-2">
          {stepLabels.map((label, index) => (
            <div key={index} className="flex-1">
              <span
                className={`text-xs transition-colors duration-300 ${
                  index < currentStep 
                    ? 'text-primary font-medium' 
                    : index === currentStep 
                      ? 'text-gray-900 font-medium' 
                      : 'text-gray-400'
                }`}
              >
                {label}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
