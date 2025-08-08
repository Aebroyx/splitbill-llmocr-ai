'use client';

import { Provider } from 'react-redux';
import { Toaster } from 'react-hot-toast';
import { makeStore } from '../lib/store';

const store = makeStore();

export default function ClientProvider({ children }: { children: React.ReactNode }) {
  return (
    <Provider store={store}>
      <Toaster />
      {children}
    </Provider>
  );
}