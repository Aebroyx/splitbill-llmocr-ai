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
  created_at: string;
  updated_at: string;
}

export interface UploadImageResponse {
  message: string;
  image_url?: string;
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
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Failed to create bill');
    }
  },

  async getBills(): Promise<Bill[]> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills`);
      return response.data as Bill[];
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Failed to fetch bills');
    }
  },

  async getBillById(id: string): Promise<Bill> {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/bills/${id}`);
      return response.data as Bill;
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
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
    } catch (error: any) {
      if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      }
      throw new Error('Failed to upload image');
    }
  },
};
