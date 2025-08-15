'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { PlusIcon } from '@heroicons/react/24/outline';
import { billService, CreateBillPayload } from '../lib/services/billService';
import toast from 'react-hot-toast';

import { Bill } from '../lib/services/billService';

interface BillCreatorProps {
  onCreateBill?: (bill: Bill) => void;
}

export default function BillCreator({ onCreateBill }: BillCreatorProps) {
  const router = useRouter();
  const [billName, setBillName] = useState('');
  const [taxAmount, setTaxAmount] = useState('');
  const [tipAmount, setTipAmount] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!billName.trim()) {
      toast.error('Bill name is required');
      return;
    }

    // Validate tax and tip amounts
    const tax = parseFloat(taxAmount) || 0;
    const tip = parseFloat(tipAmount) || 0;
    
    if (tax < 0 || tip < 0) {
      toast.error('Tax and tip amounts cannot be negative');
      return;
    }

    setIsSubmitting(true);
    
    try {
      const payload: CreateBillPayload = {
        name: billName.trim(),
        tax_amount: tax,
        tip_amount: tip,
      };

      const newBill = await billService.createBill(payload);
      
      toast.success('Bill created successfully!');
      onCreateBill?.(newBill);
      
      // Redirect to the new bill page for image upload
      router.push(`/${newBill.id}`);
    } catch (error) {
      console.error('Error creating bill:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to create bill');
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setBillName('');
    setTaxAmount('');
    setTipAmount('');
    setIsCreating(false);
  };

  if (!isCreating) {
    return (
      <div className="w-full max-w-md mx-auto">
        <button
          onClick={() => setIsCreating(true)}
          className="w-full bg-primary hover:bg-primary-dark text-white font-medium py-4 px-6 rounded-xl shadow-lg transition-all duration-200 flex items-center justify-center gap-3 text-lg"
        >
          <PlusIcon className="w-6 h-6" />
          Create New Bill
        </button>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto bg-white rounded-xl shadow-lg border border-gray-200 p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-4 text-center">
        Create New Bill
      </h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="billName" className="block text-sm font-medium text-gray-700 mb-2">
            Bill Name *
          </label>
          <input
            type="text"
            id="billName"
            value={billName}
            onChange={(e) => setBillName(e.target.value)}
            placeholder="Enter bill name..."
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary transition-colors duration-200 text-base text-gray-900 placeholder-gray-500"
            autoFocus
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="taxAmount" className="block text-sm font-medium text-gray-700 mb-2">
              Tax Amount
            </label>
            <input
              type="number"
              id="taxAmount"
              value={taxAmount}
              onChange={(e) => setTaxAmount(e.target.value)}
              placeholder="0.00"
              step="0.01"
              min="0"
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary transition-colors duration-200 text-base text-gray-900 placeholder-gray-500"
            />
          </div>

          <div>
            <label htmlFor="tipAmount" className="block text-sm font-medium text-gray-700 mb-2">
              Tip Amount
            </label>
            <input
              type="number"
              id="tipAmount"
              value={tipAmount}
              onChange={(e) => setTipAmount(e.target.value)}
              placeholder="0.00"
              step="0.01"
              min="0"
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary transition-colors duration-200 text-base text-gray-900 placeholder-gray-500"
            />
          </div>
        </div>
        
        <div className="flex gap-3 pt-2">
          <button
            type="button"
            onClick={handleCancel}
            disabled={isSubmitting}
            className="flex-1 px-4 py-3 text-gray-700 bg-gray-100 hover:bg-gray-200 disabled:bg-gray-300 disabled:cursor-not-allowed rounded-lg font-medium transition-colors duration-200"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!billName.trim() || isSubmitting}
            className="flex-1 px-4 py-3 bg-primary hover:bg-primary-dark disabled:bg-gray-400 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-colors duration-200"
          >
            {isSubmitting ? 'Creating...' : 'Create Bill'}
          </button>
        </div>
      </form>
    </div>
  );
}
