'use client';

import { useState } from 'react';
import { PlusIcon, XMarkIcon, UserIcon, CheckIcon } from '@heroicons/react/24/outline';
import { BillItem, BillParticipant, billService } from '../lib/services/billService';
import toast from 'react-hot-toast';

interface ParticipantManagerProps {
  billId: string;
  items: BillItem[];
  participants: BillParticipant[];
  onParticipantsChange: (participants: BillParticipant[]) => void;
  onItemAssignmentsChange: (assignments: ItemAssignment[]) => void;
}

interface ItemAssignment {
  itemId: number;
  participantId: number;
}

export default function ParticipantManager({
  billId,
  items,
  participants,
  onParticipantsChange,
  onItemAssignmentsChange
}: ParticipantManagerProps) {
  const [newParticipantName, setNewParticipantName] = useState('');
  const [isAddingParticipant, setIsAddingParticipant] = useState(false);
  const [itemAssignments, setItemAssignments] = useState<ItemAssignment[]>([]);
  const [editingParticipant, setEditingParticipant] = useState<number | null>(null);
  const [editName, setEditName] = useState('');

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
      
      onParticipantsChange([...participants, newParticipant]);
      setNewParticipantName('');
      setIsAddingParticipant(false);
      toast.success('Participant added successfully!');
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

  const handleRemoveParticipant = (participantId: number) => {
    const updatedParticipants = participants.filter(p => p.id !== participantId);
    onParticipantsChange(updatedParticipants);
    
    // Remove item assignments for this participant
    const updatedAssignments = itemAssignments.filter(a => a.participantId !== participantId);
    setItemAssignments(updatedAssignments);
    onItemAssignmentsChange(updatedAssignments);
    
    toast.success('Participant removed!');
  };

  const handleItemAssignment = async (itemId: number, participantId: number) => {
    const existingAssignment = itemAssignments.find(
      a => a.itemId === itemId && a.participantId === participantId
    );

    if (existingAssignment) {
      // Remove assignment
      const updatedAssignments = itemAssignments.filter(
        a => !(a.itemId === itemId && a.participantId === participantId)
      );
      setItemAssignments(updatedAssignments);
      onItemAssignmentsChange(updatedAssignments);
      toast.success('Item assignment removed');
    } else {
      // Add assignment
      try {
        await billService.assignItemToParticipant(itemId, participantId);
        const newAssignment = { itemId, participantId };
        const updatedAssignments = [...itemAssignments, newAssignment];
        setItemAssignments(updatedAssignments);
        onItemAssignmentsChange(updatedAssignments);
        toast.success('Item assigned successfully!');
      } catch (error) {
        console.error('Error assigning item:', error);
        toast.error('Failed to assign item');
      }
    }
  };

  const isItemAssignedToParticipant = (itemId: number, participantId: number) => {
    return itemAssignments.some(
      a => a.itemId === itemId && a.participantId === participantId
    );
  };

  const getParticipantTotal = (participantId: number) => {
    return items.reduce((total, item) => {
      if (isItemAssignedToParticipant(item.id, participantId)) {
        return total + (item.price * item.quantity);
      }
      return total;
    }, 0);
  };

  const getTotalAssigned = () => {
    return participants.reduce((total, participant) => {
      return total + getParticipantTotal(participant.id);
    }, 0);
  };

  const getTotalItems = () => {
    return items.reduce((total, item) => total + (item.price * item.quantity), 0);
  };

  return (
    <div className="space-y-6">
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
            <div className="flex gap-3">
              <input
                type="text"
                value={newParticipantName}
                onChange={(e) => setNewParticipantName(e.target.value)}
                placeholder="Enter participant name"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:border-primary focus:ring-primary"
                onKeyPress={(e) => e.key === 'Enter' && handleAddParticipant()}
                autoFocus
              />
              <button
                onClick={handleAddParticipant}
                className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg font-medium transition-colors"
              >
                Add
              </button>
              <button
                onClick={() => {
                  setIsAddingParticipant(false);
                  setNewParticipantName('');
                }}
                className="px-4 py-2 bg-gray-500 hover:bg-gray-600 text-white rounded-lg font-medium transition-colors"
              >
                Cancel
              </button>
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
                      className="px-2 py-1 border border-gray-300 rounded focus:border-primary focus:ring-primary"
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
                <span className="text-sm text-gray-600">
                  Total: ${getParticipantTotal(participant.id).toFixed(2)}
                </span>
                <button
                  onClick={() => handleEditParticipant(participant.id, participant.name)}
                  className="text-blue-600 hover:text-blue-700 text-sm"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleRemoveParticipant(participant.id)}
                  className="text-red-600 hover:text-red-700"
                >
                  <XMarkIcon className="w-4 h-4" />
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
              <span className="font-medium">${getTotalItems().toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Assigned Items:</span>
              <span className="font-medium">${getTotalAssigned().toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600">Unassigned Items:</span>
              <span className="font-medium">${(getTotalItems() - getTotalAssigned()).toFixed(2)}</span>
            </div>
            
            <div className="pt-3 border-t border-gray-200">
              <div className="space-y-2">
                {participants.map((participant) => (
                  <div key={participant.id} className="flex justify-between items-center text-sm">
                    <span className="text-gray-600">{participant.name}:</span>
                    <span className="font-medium">${getParticipantTotal(participant.id).toFixed(2)}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
