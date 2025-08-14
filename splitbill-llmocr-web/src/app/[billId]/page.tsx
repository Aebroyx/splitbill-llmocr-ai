'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { billService, Bill, BillWithItems, BillParticipant } from '../../lib/services/billService';
import { CameraIcon, PhotoIcon, XMarkIcon, ArrowLeftIcon, ExclamationTriangleIcon, UserGroupIcon, DocumentTextIcon, XCircleIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
import { useRef } from 'react';
import { useBillStatus } from '../../lib/hooks/useBillStatus';
import BillStatusIndicator from '../../components/BillStatusIndicator';
import ParticipantManager from '../../components/ParticipantManager';


export default function BillPage() {
  const params = useParams();
  const router = useRouter();
  const billId = params.billId as string;
  
  const [bill, setBill] = useState<Bill | null>(null);
  const [billWithItems, setBillWithItems] = useState<BillWithItems | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [participants, setParticipants] = useState<BillParticipant[]>([]);
  const [itemAssignments, setItemAssignments] = useState<{itemId: number, participantId: number}[]>([]);
  const [activeTab, setActiveTab] = useState<'overview' | 'participants' | 'upload'>('overview');
  const [editingItem, setEditingItem] = useState<number | null>(null);
  const [editItemData, setEditItemData] = useState<{name: string, price: number, quantity: number}>({
    name: '',
    price: 0,
    quantity: 1
  });
  const [editingTaxTip, setEditingTaxTip] = useState(false);
  const [editTaxTipData, setEditTaxTipData] = useState<{tax_amount: number, tip_amount: number}>({
    tax_amount: 0,
    tip_amount: 0
  });
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  // Use the status polling hook
  const { status, isPolling, error: statusError, lastUpdated } = useBillStatus({
    billId,
    pollInterval: 2000,
    onStatusChange: (newStatus) => {
      console.log('Status changed to:', newStatus);
    },
    onComplete: async (billStatus) => {
      console.log('Processing completed with status:', billStatus.status);
      if (billStatus.status === 'completed') {
        // Load the bill with items and participants
        try {
          const updatedBill = await billService.getBillWithItems(billId);
          setBillWithItems(updatedBill);
          setParticipants(updatedBill.participants || []);
          toast.success('Bill processing completed! Items have been extracted.');
        } catch (error) {
          console.error('Error loading bill with items:', error);
          toast.error('Failed to load extracted items');
        }
      } else if (billStatus.status === 'failed') {
        toast.error('Bill processing failed. Please try uploading the image again.');
      }
    },
    onError: (error) => {
      console.error('Status polling error:', error);
      toast.error('Failed to check bill status');
    }
  });

  useEffect(() => {
    if (billId) {
      loadBill();
    }
  }, [billId]);

  const loadBill = async () => {
    try {
      const billData = await billService.getBillById(billId);
      setBill(billData);
      
      // If bill is already completed, load items and participants
      if (billData.status === 'completed') {
        const billWithItemsData = await billService.getBillWithItems(billId);
        setBillWithItems(billWithItemsData);
        setParticipants(billWithItemsData.participants || []);
      }
    } catch (error) {
      console.error('Error loading bill:', error);
      toast.error('Failed to load bill');
      router.push('/');
    } finally {
      setLoading(false);
    }
  };

  const handleImageUpload = (file: File) => {
    if (file && file.type.startsWith('image/')) {
      setSelectedImage(file);
      const reader = new FileReader();
      reader.onload = (e) => {
        setImagePreview(e.target?.result as string);
      };
      reader.readAsDataURL(file);
    } else {
      toast.error('Please select a valid image file');
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleImageUpload(file);
    }
  };

  const startCamera = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
      }
    } catch (error) {
      toast.error('Unable to access camera');
      console.error('Camera error:', error);
    }
  };

  const stopCamera = () => {
    if (videoRef.current?.srcObject) {
      const stream = videoRef.current.srcObject as MediaStream;
      stream.getTracks().forEach(track => track.stop());
      videoRef.current.srcObject = null;
    }
  };

  const capturePhoto = () => {
    if (videoRef.current && canvasRef.current) {
      const context = canvasRef.current.getContext('2d');
      if (context) {
        canvasRef.current.width = videoRef.current.videoWidth;
        canvasRef.current.height = videoRef.current.videoHeight;
        context.drawImage(videoRef.current, 0, 0);
        
        canvasRef.current.toBlob((blob) => {
          if (blob) {
            const file = new File([blob], 'camera-photo.jpg', { type: 'image/jpeg' });
            handleImageUpload(file);
            stopCamera();
          }
        }, 'image/jpeg');
      }
    }
  };

  const handleImageSubmit = async () => {
    if (!selectedImage) {
      toast.error('Please select or capture an image');
      return;
    }

    setIsUploading(true);
    
    try {
      const result = await billService.uploadBillImage(billId, selectedImage);
      toast.success(result.message || 'Image uploaded successfully!');
      
      // Reset image selection
      setSelectedImage(null);
      setImagePreview(null);
      
      // The status polling will automatically detect the status change
    } catch (error) {
      console.error('Error uploading image:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to upload image');
    } finally {
      setIsUploading(false);
    }
  };

  const handleParticipantsChange = (newParticipants: BillParticipant[]) => {
    setParticipants(newParticipants);
  };

  const handleItemAssignmentsChange = (newAssignments: {itemId: number, participantId: number}[]) => {
    setItemAssignments(newAssignments);
  };

  const handleRetryUpload = () => {
    // Reset status and allow user to upload again
    setActiveTab('upload');
    // Clear any previous image selection
    setSelectedImage(null);
    setImagePreview(null);
  };

  const startEditingItem = (item: any) => {
    setEditingItem(item.id);
    setEditItemData({
      name: item.name,
      price: item.price,
      quantity: item.quantity
    });
  };

  const cancelEditingItem = () => {
    setEditingItem(null);
    setEditItemData({ name: '', price: 0, quantity: 1 });
  };

  const saveItemEdit = async (itemId: number) => {
    if (!billWithItems) return;

    // Validate the input
    if (!editItemData.name.trim()) {
      toast.error('Item name cannot be empty');
      return;
    }

    if (editItemData.price < 0) {
      toast.error('Price cannot be negative');
      return;
    }

    if (editItemData.quantity < 1) {
      toast.error('Quantity must be at least 1');
      return;
    }

    try {
      // Update the item via API
      await billService.updateItem(itemId, {
        name: editItemData.name.trim(),
        price: editItemData.price,
        quantity: editItemData.quantity
      });

      // Update the item in the local state for immediate UI feedback
      const updatedItems = billWithItems.items?.map(item => 
        item.id === itemId 
          ? { ...item, ...editItemData }
          : item
      ) || [];

      const updatedBillWithItems = {
        ...billWithItems,
        items: updatedItems
      };

      setBillWithItems(updatedBillWithItems);

      setEditingItem(null);
      setEditItemData({ name: '', price: 0, quantity: 1 });
      toast.success('Item updated successfully!');
    } catch (error) {
      console.error('Error updating item:', error);
      toast.error('Failed to update item');
    }
  };

  const startEditingTaxTip = () => {
    setEditingTaxTip(true);
    setEditTaxTipData({
      tax_amount: bill?.tax_amount || 0,
      tip_amount: bill?.tip_amount || 0
    });
  };

  const cancelEditingTaxTip = () => {
    setEditingTaxTip(false);
    setEditTaxTipData({ tax_amount: 0, tip_amount: 0 });
  };

  const saveTaxTipEdit = async () => {
    if (!bill) return;

    // Validate the input
    if (editTaxTipData.tax_amount < 0) {
      toast.error('Tax amount cannot be negative');
      return;
    }

    if (editTaxTipData.tip_amount < 0) {
      toast.error('Tip amount cannot be negative');
      return;
    }

    try {
      // Update the bill via API
      await billService.updateBill(bill.id, {
        tax_amount: editTaxTipData.tax_amount,
        tip_amount: editTaxTipData.tip_amount
      });

      // Update the bill in the local state for immediate UI feedback
      const updatedBill = {
        ...bill,
        tax_amount: editTaxTipData.tax_amount,
        tip_amount: editTaxTipData.tip_amount
      };

      setBill(updatedBill);

      setEditingTaxTip(false);
      setEditTaxTipData({ tax_amount: 0, tip_amount: 0 });
      toast.success('Tax and tip updated successfully!');
    } catch (error) {
      console.error('Error updating tax and tip:', error);
      toast.error('Failed to update tax and tip');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-gray-600">Loading bill...</p>
        </div>
      </div>
    );
  }

  if (!bill) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600 mb-4">Bill not found</p>
          <button
            onClick={() => router.push('/')}
            className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-dark transition-colors"
          >
            Go Home
          </button>
        </div>
      </div>
    );
  }

  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return (
          <>
            {/* Bill Info */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-4">{bill?.name}</h2>
              
              {editingTaxTip ? (
                // Edit Mode for Tax and Tip
                <div className="space-y-4 mb-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-medium text-gray-900">Editing Tax & Tip</h3>
                    <div className="flex gap-2">
                      <button
                        onClick={saveTaxTipEdit}
                        className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded text-sm font-medium transition-colors"
                      >
                        Save
                      </button>
                      <button
                        onClick={cancelEditingTaxTip}
                        className="px-3 py-1 bg-gray-500 hover:bg-gray-600 text-white rounded text-sm font-medium transition-colors"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Tax Amount
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        value={editTaxTipData.tax_amount}
                        onChange={(e) => setEditTaxTipData({...editTaxTipData, tax_amount: parseFloat(e.target.value) || 0})}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Tip Amount
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        value={editTaxTipData.tip_amount}
                        onChange={(e) => setEditTaxTipData({...editTaxTipData, tip_amount: parseFloat(e.target.value) || 0})}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary"
                      />
                    </div>
                  </div>
                </div>
              ) : (
                // Display Mode for Tax and Tip
                <div className="grid grid-cols-2 gap-4 text-sm mb-4">
                  <div className="flex items-center justify-between">
                    <span className="text-gray-500">Tax Amount:</span>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-gray-900">
                        ${bill?.tax_amount ? bill.tax_amount.toFixed(2) : '0.00'}
                      </span>
                      <button
                        onClick={startEditingTaxTip}
                        className="text-blue-600 hover:text-blue-700 text-xs"
                      >
                        Edit
                      </button>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-gray-500">Tip Amount:</span>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-gray-900">
                        ${bill?.tip_amount ? bill.tip_amount.toFixed(2) : '0.00'}
                      </span>
                      <button
                        onClick={startEditingTaxTip}
                        className="text-blue-600 hover:text-blue-700 text-xs"
                      >
                        Edit
                      </button>
                    </div>
                  </div>
                </div>
              )}
              
              <div className="text-xs text-gray-400 mb-4">
                Created: {bill && new Date(bill.created_at).toLocaleDateString()}
              </div>
              
              {/* Status Indicator */}
              <div className="mb-4">
                <BillStatusIndicator 
                  status={status} 
                  isPolling={isPolling} 
                  lastUpdated={lastUpdated}
                  onRetry={status === 'failed' ? handleRetryUpload : undefined}
                />
              </div>
            </div>

            {/* Extracted Items Section */}
            {billWithItems && billWithItems.items && billWithItems.items.length > 0 && (
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-8">
                <h3 className="text-xl font-semibold text-gray-900 mb-4">
                  Extracted Items
                </h3>
                <p className="text-sm text-gray-600 mb-4">
                  Click on any item to edit the quantity or price if the AI made a mistake
                </p>
                <div className="space-y-3">
                  {billWithItems.items.map((item) => (
                    <div key={item.id} className="border border-gray-200 rounded-lg p-4">
                      {editingItem === item.id ? (
                        // Edit Mode
                        <div className="space-y-3">
                          <div className="flex items-center justify-between">
                            <h4 className="font-medium text-gray-900">Editing: {item.name}</h4>
                            <div className="flex gap-2">
                              <button
                                onClick={() => saveItemEdit(item.id)}
                                className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded text-sm font-medium transition-colors"
                              >
                                Save
                              </button>
                              <button
                                onClick={cancelEditingItem}
                                className="px-3 py-1 bg-gray-500 hover:bg-gray-600 text-white rounded text-sm font-medium transition-colors"
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                          
                          <div className="grid grid-cols-3 gap-3">
                            <div>
                              <label className="block text-xs font-medium text-gray-700 mb-1">
                                Name
                              </label>
                              <input
                                type="text"
                                value={editItemData.name}
                                onChange={(e) => setEditItemData({...editItemData, name: e.target.value})}
                                className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:border-primary focus:ring-primary"
                              />
                            </div>
                            <div>
                              <label className="block text-xs font-medium text-gray-700 mb-1">
                                Price
                              </label>
                              <input
                                type="number"
                                step="0.01"
                                min="0"
                                value={editItemData.price}
                                onChange={(e) => setEditItemData({...editItemData, price: parseFloat(e.target.value) || 0})}
                                className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:border-primary focus:ring-primary"
                              />
                            </div>
                            <div>
                              <label className="block text-xs font-medium text-gray-700 mb-1">
                                Quantity
                              </label>
                              <input
                                type="number"
                                min="1"
                                value={editItemData.quantity}
                                onChange={(e) => setEditItemData({...editItemData, quantity: parseInt(e.target.value) || 1})}
                                className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:border-primary focus:ring-primary"
                              />
                            </div>
                          </div>
                          
                          <div className="text-right text-sm text-gray-600">
                            Total: ${(editItemData.price * editItemData.quantity).toFixed(2)}
                          </div>
                        </div>
                      ) : (
                        // Display Mode
                        <div className="flex items-center justify-between">
                          <div className="flex-1">
                            <h4 className="font-medium text-gray-900">{item.name}</h4>
                            <p className="text-sm text-gray-500">
                              Price: ${item.price.toFixed(2)} × Quantity: {item.quantity}
                            </p>
                          </div>
                          <div className="flex items-center gap-3">
                            <span className="font-semibold text-gray-900">
                              ${(item.price * item.quantity).toFixed(2)}
                            </span>
                            <button
                              onClick={() => startEditingItem(item)}
                              className="px-3 py-1 text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded text-sm font-medium transition-colors"
                            >
                              Edit
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                  <div className="pt-3 border-t border-gray-200 gap-4">
                    <div className="flex justify-between items-center font-semibold gap-4">
                      <span className="text-gray-900">Subtotal:</span>
                      <span className="text-gray-900">${billWithItems.items.reduce((sum, item) => sum + (item.price * item.quantity), 0).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between items-center text-sm text-gray-600 gap-4">
                      <span className="text-gray-900 mt-1">Tax:</span>
                      <span className="text-gray-900 mt-1">${bill?.tax_amount ? bill.tax_amount.toFixed(2) : '0.00'}</span>
                    </div>
                    <div className="flex justify-between items-center text-sm text-gray-600 gap-4 mb-4">
                      <span className="text-gray-900 mt-1">Tip:</span>
                      <span className="text-gray-900 mt-1">${bill?.tip_amount ? bill.tip_amount.toFixed(2) : '0.00'}</span>
                    </div>
                    <div className="flex justify-between items-center font-semibold text-lg pt-2 border-t border-gray-200 gap-4">
                      <span className="text-gray-900">Total:</span>
                      <span className="text-gray-900">${(billWithItems.items.reduce((sum, item) => sum + (item.price * item.quantity), 0) + (bill?.tax_amount || 0) + (bill?.tip_amount || 0)).toFixed(2)}</span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </>
        );

      case 'participants':
        return billWithItems ? (
          <ParticipantManager
            billId={billId}
            items={billWithItems.items || []}
            participants={participants}
            onParticipantsChange={handleParticipantsChange}
            onItemAssignmentsChange={handleItemAssignmentsChange}
          />
        ) : (
          <div className="text-center py-12 text-gray-500">
            <DocumentTextIcon className="w-16 h-16 mx-auto mb-4 text-gray-300" />
            <p className="text-lg font-medium">Items not yet processed</p>
            <p className="text-sm">Please wait for the AI to finish processing your bill image</p>
          </div>
        );

      case 'upload':
        return (
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="text-xl font-semibold text-gray-900 mb-4 text-center">
              Upload Bill Image
            </h3>
            
            {status === 'failed' && (
              <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
                <div className="flex items-center gap-3">
                  <XCircleIcon className="w-6 h-6 text-red-600" />
                  <div>
                    <h4 className="font-medium text-red-800">Previous upload failed</h4>
                    <p className="text-sm text-red-700">
                      The AI processing service encountered an error. You can try uploading the image again.
                    </p>
                  </div>
                </div>
              </div>
            )}
            
            <div className="space-y-4">
              {/* Image Preview */}
              {imagePreview && (
                <div className="relative">
                  <img 
                    src={imagePreview} 
                    alt="Bill preview" 
                    className="w-full h-48 object-cover rounded-lg border border-gray-300"
                  />
                  <button
                    onClick={() => {
                      setSelectedImage(null);
                      setImagePreview(null);
                    }}
                    className="absolute top-2 right-2 bg-red-500 text-white rounded-full p-1 hover:bg-red-600 transition-colors"
                  >
                    <XMarkIcon className="w-4 h-4" />
                  </button>
                </div>
              )}

              {/* Camera Section */}
              {!imagePreview && (
                <div className="space-y-3">
                  <div className="relative">
                    <video
                      ref={videoRef}
                      autoPlay
                      playsInline
                      className="w-full h-48 object-cover rounded-lg border border-gray-300 bg-gray-100"
                    />
                    <canvas ref={canvasRef} className="hidden" />
                  </div>
                  
                  <div className="flex gap-2">
                    <button
                      onClick={startCamera}
                      className="flex-1 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
                    >
                      <CameraIcon className="w-4 h-4" />
                      Start Camera
                    </button>
                    <button
                      onClick={capturePhoto}
                      className="flex-1 px-4 py-2 bg-green-500 hover:bg-green-600 text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2"
                    >
                      <PhotoIcon className="w-4 h-4" />
                      Capture
                    </button>
                  </div>
                </div>
              )}

              {/* File Upload */}
              {!imagePreview && (
                <div className="space-y-3">
                  <label className="block text-sm font-medium text-gray-700">
                    Or upload from device
                  </label>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={handleFileInputChange}
                    className="hidden"
                  />
                  <button
                    onClick={() => fileInputRef.current?.click()}
                    className="w-full px-4 py-3 border-2 border-dashed border-gray-300 rounded-lg hover:border-primary hover:bg-gray-50 transition-colors text-gray-600 hover:text-gray-900"
                  >
                    <PhotoIcon className="w-6 h-6 mx-auto mb-2" />
                    <span>Click to select image</span>
                  </button>
                </div>
              )}

              {/* Upload Button */}
              {selectedImage && (
                <button
                  onClick={handleImageSubmit}
                  disabled={isUploading}
                  className="w-full px-4 py-3 bg-primary hover:bg-primary-dark disabled:bg-gray-400 disabled:cursor-not-allowed text-white rounded-lg font-medium transition-colors duration-200"
                >
                  {isUploading ? 'Uploading...' : 'Upload Image'}
                </button>
              )}
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center h-16">
            <button
              onClick={() => router.push('/')}
              className="mr-4 p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <ArrowLeftIcon className="w-5 h-5" />
            </button>
            <h1 className="text-xl font-bold text-gray-900">SplitBill</h1>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="max-w-4xl mx-auto">
          {/* Tabs */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-1 mb-8">
            <div className="flex space-x-1">
              <button
                onClick={() => setActiveTab('overview')}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                  activeTab === 'overview'
                    ? 'bg-primary text-white'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                }`}
              >
                Overview
              </button>
              <button
                onClick={() => setActiveTab('upload')}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                  activeTab === 'upload'
                    ? 'bg-primary text-white'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                }`}
                disabled={status === 'completed'}
              >
                Upload Image
              </button>
              <button
                onClick={() => setActiveTab('participants')}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                  activeTab === 'participants'
                    ? 'bg-primary text-white'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                }`}
                disabled={!billWithItems || !billWithItems.items || billWithItems.items.length === 0}
              >
                <UserGroupIcon className="w-4 h-4 inline mr-2" />
                Participants
              </button>
            </div>
          </div>

          {/* Tab Content */}
          {renderTabContent()}

          {/* Error Display */}
          {statusError && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 mt-4">
              <div className="flex items-center gap-2">
                <ExclamationTriangleIcon className="w-5 h-5 text-red-600" />
                <span className="text-red-800 font-medium">Error checking status:</span>
              </div>
              <p className="text-red-700 mt-1">{statusError.message}</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
