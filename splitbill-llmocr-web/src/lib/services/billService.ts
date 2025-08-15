import axios from 'axios';

export interface CreateBillPayload {
  name: string;
  tax_amount: number;
  tip_amount: number;
}

export interface Bill {
  id: string;
  name: string;
  tax_amount: number;
  tip_amount: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface BillWithItems extends Bill {
  items?: BillItem[];
  participants?: BillParticipant[];
}

export interface BillItem {
  id: number;
  bill_id: string;
  name: string;
  price: number;
  quantity: number;
  created_at: string;
}

export interface BillParticipant {
  id: number;
  bill_id: string;
  name: string;
  payment_status: string;
  share_of_common_costs: number;
  created_at: string;
}

export interface BillStatus {
  bill_id: string;
  status: string;
}

export interface UploadImageResponse {
  message: string;
  image_url?: string;
}

export interface ItemAssignment {
  item_id: number;
  participant_id: number;
}

export interface UpdateItemPayload {
  name?: string;
  price?: number;
  quantity?: number;
}

export interface UpdateBillPayload {
  tax_amount?: number;
  tip_amount?: number;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const billService = {
  async createBill(payload: CreateBillPayload): Promise<Bill> {
    try {
      const response = await axios.post(`${API_BASE_URL}/api/bills/`, payload, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      return response.data as Bill;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to create bill');
    }
  },

  async getBills(): Promise<Bill[]> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills`);
      return response.data as Bill[];
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to fetch bills');
    }
  },

  async getBillById(id: string): Promise<Bill> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${id}`);
      return response.data as Bill;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to fetch bill');
    }
  },

  async uploadBillImage(billId: string, imageFile: File): Promise<UploadImageResponse> {
    try {
      const formData = new FormData();
      formData.append('image', imageFile);
      
      const response = await axios.post(`${API_BASE_URL}/api/bills/${billId}/image`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data as UploadImageResponse;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to upload image');
    }
  },

  async getBillStatus(billId: string): Promise<BillStatus> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${billId}/status`);
      return response.data as BillStatus;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to fetch bill status');
    }
  },

  async getBillWithItems(billId: string): Promise<BillWithItems> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${billId}`);
      return response.data as BillWithItems;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to fetch bill with items');
    }
  },

  async getParticipants(billId: string): Promise<BillParticipant[]> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${billId}/participants`);
      return response.data as BillParticipant[];
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to fetch participants');
    }
  },

  async getItemAssignments(billId: string): Promise<ItemAssignment[]> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${billId}/item-assignments`);
      
      if (response.data && Array.isArray(response.data)) {
        return response.data.map((assignment: { item_id: number; participant_id: number }) => ({
          item_id: assignment.item_id,
          participant_id: assignment.participant_id
        }));
      }
      
      return [];
    } catch (error) {
      console.error('Error fetching item assignments:', error);
      return [];
    }
  },

  async addParticipant(billId: string, participant: Omit<BillParticipant, 'id' | 'bill_id' | 'created_at'>): Promise<BillParticipant> {
    try {
      const response = await axios.post(`${API_BASE_URL}/api/bills/${billId}/participants`, participant, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      return response.data as BillParticipant;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to add participant');
    }
  },

  async assignItemToParticipant(billId: string, itemId: number, participantId: number): Promise<ItemAssignment> {
    try {
      const response = await axios.post(`${API_BASE_URL}/api/bills/${billId}/assign-items`, {
        item_id: itemId,
        participant_id: participantId,
      }, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      return response.data as ItemAssignment;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to assign item to participant');
    }
  },

  async deleteItemAssignment(billId: string, itemId: number, participantId: number): Promise<void> {
    try {
      await axios({
        method: 'delete',
        url: `${API_BASE_URL}/api/bills/${billId}/assign-items`,
        data: {
          item_id: itemId,
          participant_id: participantId,
        },
        headers: {
          'Content-Type': 'application/json',
        },
      });
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to remove item assignment');
    }
  },

  async deleteParticipant(billId: string, participantId: number): Promise<void> {
    try {
      await axios.delete(`${API_BASE_URL}/api/bills/${billId}/participants/${participantId}`);
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to delete participant');
    }
  },

  async updateItem(itemId: number, updates: UpdateItemPayload): Promise<BillItem> {
    try {
      const response = await axios.put(`${API_BASE_URL}/api/items/${itemId}`, updates, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      return response.data as BillItem;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to update item');
    }
  },

  async updateBill(billId: string, updates: UpdateBillPayload): Promise<Bill> {
    try {
      const response = await axios.put(`${API_BASE_URL}/api/bills/${billId}`, updates, {
        headers: {
          'Content-Type': 'application/json',
        },
      });
      return response.data as Bill;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to update bill');
    }
  },

  async getBillSummary(billId: string): Promise<unknown> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${billId}/summary`);
      return response.data;
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'response' in error && error.response && typeof error.response === 'object' && 'data' in error.response && error.response.data && typeof error.response.data === 'object' && 'message' in error.response.data) {
        throw new Error((error.response.data as { message: string }).message);
      }
      throw new Error('Failed to fetch bill summary');
    }
  },
};
