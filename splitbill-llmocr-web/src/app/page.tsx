'use client';

// src/app/page.tsx
import BillCreator from '../components/BillCreator';
import { Toaster } from 'react-hot-toast';

export default function Home() {
  const handleCreateBill = () => {
    // Bill creation handled by redirect to bill page
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Toast Notifications */}
      <Toaster 
        position="top-center"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#363636',
            color: '#fff',
          },
          success: {
            duration: 3000,
            iconTheme: {
              primary: '#10B981',
              secondary: '#fff',
            },
          },
          error: {
            duration: 4000,
            iconTheme: {
              primary: '#EF4444',
              secondary: '#fff',
            },
          },
        }}
      />

      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <h1 className="text-xl font-bold text-gray-900">SplitBill</h1>
            <div className="flex items-center space-x-4">
              {/* Add user profile or settings icon here later */}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="max-w-2xl mx-auto">
          {/* Welcome Section */}
          <div className="text-center mb-8">
            <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-2">
              Welcome to SplitBill
            </h2>
            <p className="text-gray-600 text-base sm:text-lg">
              Create and manage your bills with ease
            </p>
          </div>

          {/* Bill Creator */}
          <div className="mb-8">
            <BillCreator onCreateBill={handleCreateBill} />
          </div>

          {/* Recent Bills Section */}
          {/* <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Recent Bills
            </h3>
            {bills.length === 0 ? (
              <div className="text-center py-8">
                <p className="text-gray-500 text-sm">
                  No bills created yet. Start by creating your first bill above!
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {bills.map((bill) => (
                  <div key={bill.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                    <div>
                      <h4 className="font-medium text-gray-900">{bill.name}</h4>
                      <p className="text-sm text-gray-500">
                        Tax: ${bill.tax_amount.toFixed(2)} | Tip: ${bill.tip_amount.toFixed(2)}
                      </p>
                    </div>
                    <span className="text-xs text-gray-400">
                      {new Date(bill.created_at).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div> */}
        </div>
      </main>
    </div>
  );
}