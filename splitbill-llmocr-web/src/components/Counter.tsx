'use client';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '../lib/store';
import { increment, decrement } from '../lib/reducers/counterReducer';
import toast from 'react-hot-toast';

export default function Counter() {
  const count = useSelector((state: RootState) => state.counter.value);
  const dispatch = useDispatch();

  const handleIncrement = () => {
    dispatch(increment());
    toast.success('Incremented!');
  };

  const handleDecrement = () => {
    dispatch(decrement());
    toast.error('Decremented!');
  };

  return (
    <div className="flex flex-col items-center p-4 rounded-lg shadow-md">
      <div className="text-2xl font-bold mb-4">{count}</div>
      <div className="flex space-x-4">
         <button
          onClick={handleDecrement}
          className="px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
        >
          Decrement
        </button>
        <button
          onClick={handleIncrement}
          className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          Increment
        </button>
      </div>
    </div>
  );
}