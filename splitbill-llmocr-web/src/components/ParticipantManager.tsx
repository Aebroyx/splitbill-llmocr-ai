'use client';

import { useState, useEffect } from 'react';
import { PlusIcon, UserIcon, CheckIcon, PencilIcon, TrashIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline';
import { BillItem, BillParticipant, billService } from '../lib/services/billService';
import toast from 'react-hot-toast';

interface ParticipantManagerProps {
  billId: string;
  items: BillItem[];
  participants: BillParticipant[];
  itemAssignments: {itemId: number, participantId: number}[];
  bill: { tax_amount: number; tip_amount: number } | null;
  onParticipantsChange: (participants: BillParticipant[]) => void;
  onItemAssignmentsChange: (assignments: {itemId: number, participantId: number}[]) => void;
}

export default function ParticipantManager({
  billId,
  items,
  participants,
  itemAssignments: initialItemAssignments,
  bill,
  onParticipantsChange,
  onItemAssignmentsChange
}: ParticipantManagerProps) {
  const [newParticipantName, setNewParticipantName] = useState('');
  const [isAddingParticipant, setIsAddingParticipant] = useState(false);
  const [itemAssignments, setItemAssignments] = useState<{itemId: number, participantId: number}[]>(initialItemAssignments);
  const [editingParticipant, setEditingParticipant] = useState<number | null>(null);
  const [editName, setEditName] = useState('');

  // Load initial item assignments from props
  useEffect(() => {
    if (initialItemAssignments) {
      setItemAssignments(initialItemAssignments);
    }
  }, [initialItemAssignments]);

  // Sync local itemAssignments with prop changes
  useEffect(() => {
    if (initialItemAssignments) {
      setItemAssignments(initialItemAssignments);
    }
  }, [initialItemAssignments]);



  const handleAddParticipant = async () => {
    if (!newParticipantName.trim()) {
      toast.error('Please enter a participant name');
      return;
    }

    try {
      const newParticipant = await billService.addParticipant(billId, {
        name: newParticipantName.trim(),
        payment_status: 'unpaid',
        share_of_common_costs: 0
      });
      
      if (newParticipant) {
        onParticipantsChange([...participants, newParticipant]);
        setNewParticipantName('');
        setIsAddingParticipant(false);
        toast.success('Participant added successfully!');
      }
    } catch (error) {
      console.error('Error adding participant:', error);
      toast.error('Failed to add participant');
    }
  };

  const handleEditParticipant = (participantId: number, currentName: string) => {
    setEditingParticipant(participantId);
    setEditName(currentName);
  };

  const handleSaveEdit = (participantId: number) => {
    if (!editName.trim()) {
      toast.error('Please enter a participant name');
      return;
    }

    const updatedParticipants = participants.map(p => 
      p.id === participantId ? { ...p, name: editName.trim() } : p
    );
    onParticipantsChange(updatedParticipants);
    setEditingParticipant(null);
    setEditName('');
    toast.success('Participant name updated!');
  };

  const handleRemoveParticipant = async (participantId: number) => {
    try {
      await billService.deleteParticipant(billId, participantId);
      
      // Remove participant from local state
      onParticipantsChange(participants.filter(p => p.id !== participantId));
      
      // Remove all item assignments for this participant
      onItemAssignmentsChange(itemAssignments.filter(assignment => assignment.participantId !== participantId));
      
      toast.success('Participant removed!');
    } catch (error) {
      console.error('Error removing participant:', error);
      toast.error('Failed to remove participant');
    }
  };

  const handleItemAssignment = async (itemId: number, participantId: number) => {
    try {
      const existingAssignment = itemAssignments.find(
        a => a.itemId === itemId && a.participantId === participantId
      );

      if (existingAssignment) {
        // Remove assignment
        await billService.deleteItemAssignment(billId, itemId, participantId);
        const updatedAssignments = itemAssignments.filter(
          a => !(a.itemId === itemId && a.participantId === participantId)
        );
        setItemAssignments(updatedAssignments);
        onItemAssignmentsChange(updatedAssignments);
        toast.success('Item assignment removed');
      } else {
        // Add assignment
        await billService.assignItemToParticipant(billId, itemId, participantId);
        const newAssignment = { itemId, participantId };
        const updatedAssignments = [...itemAssignments, newAssignment];
        setItemAssignments(updatedAssignments);
        onItemAssignmentsChange(updatedAssignments);
        toast.success('Item assigned successfully!');
      }
    } catch (error) {
      console.error('Error assigning item:', error);
      toast.error('Failed to assign item');
    }
  };

  const isItemAssignedToParticipant = (itemId: number, participantId: number) => {
    const isAssigned = itemAssignments.some(
      a => a.itemId === itemId && a.participantId === participantId
    );
    return isAssigned;
  };

  const getParticipantTotal = (participantId: number) => {
    let itemTotal = 0;
    
    // Calculate total from assigned items
    items.forEach(item => {
      if (isItemAssignedToParticipant(item.id, participantId)) {
        // Count how many participants are sharing this item
        const participantsSharingThisItem = itemAssignments.filter(
          assignment => assignment.itemId === item.id
        ).length;
        
        if (participantsSharingThisItem > 0) {
          // Split the item cost equally among all participants sharing it
          const itemCostPerPerson = (item.price * item.quantity) / participantsSharingThisItem;
          itemTotal += itemCostPerPerson;
        }
      }
    });
    
    return itemTotal;
  };

  const getTotalAssigned = () => {
    return participants.reduce((total, participant) => {
      return total + getParticipantTotal(participant.id);
    }, 0);
  };

  const getTotalItems = () => {
    return items.reduce((total, item) => total + (item.price * item.quantity), 0);
  };



  const getParticipantTaxTipShare = (participantId: number) => {
    if (!bill) return 0;
    
    const taxAmount = bill.tax_amount || 0;
    const tipAmount = bill.tip_amount || 0;
    const totalTaxTip = taxAmount + tipAmount;
    
    if (totalTaxTip === 0) return 0;
    
    // Calculate participant's share of items
    const participantItemTotal = getParticipantTotal(participantId);
    const totalAssignedItems = getTotalAssigned();
    
    if (totalAssignedItems === 0) return 0;
    
    // Calculate tax and tip proportion based on items purchased
    const participantRatio = participantItemTotal / totalAssignedItems;
    return totalTaxTip * participantRatio;
  };

  const getParticipantTotalWithTaxTip = (participantId: number) => {
    const itemTotal = getParticipantTotal(participantId);
    const taxTipShare = getParticipantTaxTipShare(participantId);
    return itemTotal + taxTipShare;
  };

  const getTotalBillWithTaxTip = () => {
    const itemsTotal = getTotalItems();
    const taxAmount = bill?.tax_amount || 0;
    const tipAmount = bill?.tip_amount || 0;
    return itemsTotal + taxAmount + tipAmount;
  };

  // Calculate unassigned items total
  const getUnassignedItemsTotal = () => {
    const assignedTotal = getTotalAssigned();
    const itemsTotal = getTotalItems();
    return itemsTotal - assignedTotal;
  };



  return (
    <div className="space-y-6">
      {/* Share Link Section */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4 sm:p-6">
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">Share this bill</h3>
          
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1 bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 min-w-0">
              <span className="text-xs sm:text-sm text-gray-600 font-mono break-all block">
                {typeof window !== 'undefined' ? window.location.href : ''}
              </span>
            </div>
            <button
              onClick={async () => {
                try {
                  await navigator.clipboard.writeText(window.location.href);
                  toast.success('Link copied to clipboard!');
                } catch {
                  // Fallback for older browsers
                  const textArea = document.createElement('textarea');
                  textArea.value = window.location.href;
                  document.body.appendChild(textArea);
                  textArea.select();
                  document.execCommand('copy');
                  document.body.removeChild(textArea);
                  toast.success('Link copied to clipboard!');
                }
              }}
              className="w-full sm:w-auto px-3 sm:px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2 text-sm sm:text-base whitespace-nowrap"
            >
              <ClipboardDocumentIcon className="w-4 h-4" />
              <span className="hidden sm:inline">Copy Link</span>
              <span className="sm:hidden">Copy</span>
            </button>
          </div>
          
          <p className="text-xs text-gray-500">
            Share this link with others so they can view and participate in splitting the bill
          </p>
        </div>
      </div>

      {/* Participants Section */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-xl font-semibold text-gray-900">Participants</h3>
          <button
            onClick={() => setIsAddingParticipant(true)}
            className="px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-lg font-medium transition-colors flex items-center gap-2"
          >
            <PlusIcon className="w-4 h-4" />
            Add Participant
          </button>
        </div>

        {/* Add New Participant */}
        {isAddingParticipant && (
          <div className="mb-4 p-4 bg-gray-50 rounded-lg">
            <div className="flex flex-col sm:flex-row gap-3">
              <input
                type="text"
                value={newParticipantName}
                onChange={(e) => setNewParticipantName(e.target.value)}
                placeholder="Enter participant name"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary text-gray-900"
                onKeyPress={(e) => e.key === 'Enter' && handleAddParticipant()}
                autoFocus
              />
              <div className="flex gap-2 sm:gap-3">
                <button
                  onClick={handleAddParticipant}
                  className="flex-1 sm:flex-none px-3 sm:px-4 py-2 bg-primary hover:bg-primary-dark text-white rounded-lg font-medium transition-colors flex items-center justify-center gap-2 text-sm sm:text-base"
                >
                  Add
                </button>
                <button
                  onClick={() => {
                    setIsAddingParticipant(false);
                    setNewParticipantName('');
                  }}
                  className="flex-1 sm:flex-none px-3 sm:px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-900 rounded-lg font-medium transition-colors flex items-center justify-center gap-2 text-sm sm:text-base"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Participants List */}
        <div className="space-y-3">
          {participants.map((participant) => (
            <div key={participant.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-3">
                <UserIcon className="w-5 h-5 text-gray-500" />
                {editingParticipant === participant.id ? (
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      className="px-2 py-1 border border-gray-300 rounded focus:border-primary focus:ring-primary text-gray-900"
                      onKeyPress={(e) => e.key === 'Enter' && handleSaveEdit(participant.id)}
                      autoFocus
                    />
                    <button
                      onClick={() => handleSaveEdit(participant.id)}
                      className="text-green-600 hover:text-green-700"
                    >
                      <CheckIcon className="w-4 h-4" />
                    </button>
                  </div>
                ) : (
                  <span className="font-medium text-gray-900">{participant.name}</span>
                )}
              </div>
              
              <div className="flex items-center gap-3">
                <span className="text-sm text-gray-500 font-bold text-l">
                  Total:
                </span>
                <span className="text-gray-900 font-bold text-xl">
                  ${getParticipantTotalWithTaxTip(participant.id).toFixed(2)}
                </span>
                <button
                  onClick={() => handleEditParticipant(participant.id, participant.name)}
                  className="text-primary hover:text-primary-dark hover:bg-indigo-50 p-1.5 rounded-lg transition-colors flex items-center justify-center"
                  title="Edit participant name"
                >
                  <PencilIcon className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleRemoveParticipant(participant.id)}
                  className="text-red-600 hover:text-red-700 hover:bg-red-50 p-1.5 rounded-lg transition-colors flex items-center justify-center"
                  title="Delete participant"
                >
                  <TrashIcon className="w-4 h-4" />
                </button>
              </div>
            </div>
          ))}
          
          {participants.length === 0 && (
            <div className="text-center py-6 text-gray-500">
              <UserIcon className="w-12 h-12 mx-auto mb-2 text-gray-300" />
              <p>No participants added yet</p>
              <p className="text-sm">Add participants to start assigning items</p>
            </div>
          )}
        </div>
      </div>

      {/* Item Assignment Section */}
      {participants.length > 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="text-xl font-semibold text-gray-900 mb-4">Assign Items to Participants</h3>
          
          <div className="space-y-4">
            {items.map((item) => (
              <div key={item.id} className="border border-gray-200 rounded-lg p-4">
                <div className="flex items-center justify-between mb-3">
                  <div>
                    <h4 className="font-medium text-gray-900">{item.name}</h4>
                    <p className="text-sm text-gray-500">
                      ${item.price.toFixed(2)} × {item.quantity} = ${(item.price * item.quantity).toFixed(2)}
                    </p>
                  </div>
                </div>
                
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                  {participants.map((participant) => (
                    <button
                      key={participant.id}
                      onClick={() => handleItemAssignment(item.id, participant.id)}
                      className={`p-2 rounded-lg border text-sm font-medium transition-colors ${
                        isItemAssignedToParticipant(item.id, participant.id)
                          ? 'bg-primary text-white border-primary'
                          : 'bg-gray-50 text-gray-700 border-gray-200 hover:bg-gray-100'
                      }`}
                    >
                      {participant.name}
                    </button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Summary Section */}
      {participants.length > 0 && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h3 className="text-xl font-semibold text-gray-900 mb-4">Bill Summary</h3>
          
          <div className="space-y-3">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Items Total:</span>
              <span className="font-medium text-gray-900">${getTotalItems().toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Assigned Items:</span>
              <span className="font-medium text-gray-900">${getTotalAssigned().toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Unassigned Items:</span>
              <span className="font-medium text-gray-900">${getUnassignedItemsTotal().toFixed(2)}</span>
            </div>
            
            {/* Tax and Tip Breakdown */}
            {bill && (bill.tax_amount > 0 || bill.tip_amount > 0) && (
              <>
                <div className="pt-3 border-t border-gray-200">
                  <h4 className="font-medium text-gray-900 mb-2">Tax & Tip Distribution</h4>
                  <div className="space-y-2 text-sm">
                    {bill.tax_amount > 0 && (
                      <div className="flex justify-between items-center">
                        <span className="text-gray-600">Tax Amount:</span>
                        <span className="font-medium text-gray-900">${bill.tax_amount.toFixed(2)}</span>
                      </div>
                    )}
                    {bill.tip_amount > 0 && (
                      <div className="flex justify-between items-center">
                        <span className="text-gray-600">Tip / Service Amount:</span>
                        <span className="font-medium text-gray-900">${bill.tip_amount.toFixed(2)}</span>
                      </div>
                    )}
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">Total Tax + Tip/Service:</span>
                      <span className="font-medium text-gray-900">${((bill.tax_amount || 0) + (bill.tip_amount || 0)).toFixed(2)}</span>
                    </div>
                    <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded">
                      Tax and tip/service are distributed proportionally based on each participant&apos;s share of items purchased
                    </div>
                  </div>
                </div>
                
                {/* Individual Tax & Tip Breakdown */}
                <div className="pt-3 border-t border-gray-200">
                  <h4 className="font-medium text-gray-900 mb-2">Individual Tax & Tip/Service Breakdown</h4>
                  <div className="space-y-2 text-sm">
                    {participants.map((participant) => {
                      const itemTotal = getParticipantTotal(participant.id);
                      const taxTipShare = getParticipantTaxTipShare(participant.id);
                      const totalAssignedItems = getTotalAssigned();
                      const participantRatio = totalAssignedItems > 0 ? (itemTotal / totalAssignedItems) * 100 : 0;
                      
                      return (
                        <div key={participant.id} className="flex justify-between items-center">
                          <div className="text-left">
                            <div className="font-medium text-gray-900">{participant.name}</div>
                            <div className="text-xs text-gray-500">
                              {participantRatio.toFixed(1)}% of items
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="text-gray-900">${taxTipShare.toFixed(2)}</div>
                            <div className="text-xs text-gray-500">
                              Tax: ${((bill?.tax_amount || 0) * (itemTotal / totalAssignedItems)).toFixed(2)} + 
                              Tip: ${((bill?.tip_amount || 0) * (itemTotal / totalAssignedItems)).toFixed(2)}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </>
            )}

            {/* Item Sharing Breakdown */}
            <div className="pt-3 border-t border-gray-200">
              <h4 className="font-medium text-gray-900 mb-2">Item Sharing Breakdown</h4>
              <div className="space-y-2 text-sm">
                {items.map(item => {
                  const participantsSharingThisItem = itemAssignments.filter(
                    assignment => assignment.itemId === item.id
                  ).length;
                  
                  if (participantsSharingThisItem === 0) {
                    return (
                      <div key={item.id} className="flex justify-between items-center text-gray-400">
                        <span>{item.name}</span>
                        <span>Unassigned</span>
                      </div>
                    );
                  }
                  
                  const itemTotal = item.price * item.quantity;
                  const costPerPerson = itemTotal / participantsSharingThisItem;
                  const participantsNames = itemAssignments
                    .filter(a => a.itemId === item.id)
                    .map(a => participants.find(p => p.id === a.participantId)?.name)
                    .filter(Boolean)
                    .join(', ');
                  
                  return (
                    <div key={item.id} className="flex justify-between items-center">
                      <div className="text-left">
                        <div className="font-medium text-gray-900">{item.name}</div>
                        <div className="text-xs text-gray-500">
                          Shared by: {participantsNames}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-gray-900">${itemTotal.toFixed(2)}</div>
                        <div className="text-xs text-gray-500">
                          ${costPerPerson.toFixed(2)} each
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Per-Participant Summary */}
            <div className="pt-3 border-t border-gray-200">
              <h4 className="font-medium text-gray-900 mb-2">Per-Participant Summary</h4>
              <div className="space-y-2">
                {participants.map((participant) => (
                  <div key={participant.id} className="flex justify-between items-center text-sm">
                    <span className="text-gray-600">{participant.name}:</span>
                    <div className="text-right">
                      <div className="font-medium text-gray-900">${getParticipantTotalWithTaxTip(participant.id).toFixed(2)}</div>
                      <div className="text-xs text-gray-500">
                        Items: ${getParticipantTotal(participant.id).toFixed(2)} + 
                        Tax/Tip: ${getParticipantTaxTipShare(participant.id).toFixed(2)}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="pt-3 border-t border-gray-200 mt-3">
              <div className="flex justify-between items-center font-medium">
                <span className="text-gray-900">Total Bill:</span>
                <span className="text-gray-900">${getTotalBillWithTaxTip().toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
