import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { voyaraApi } from '../api/client';

interface CartContextType {
  count: number;
  refreshCount: () => void;
}

const CartContext = createContext<CartContextType>({ count: 0, refreshCount: () => {} });

export function CartProvider({ children }: { children: ReactNode }) {
  const [count, setCount] = useState(0);

  const refreshCount = async () => {
    try {
      const token = localStorage.getItem('voyara_token');
      if (!token) { setCount(0); return; }
      const res = await voyaraApi.get<{ count: number }>('/cart?countOnly=1');
      setCount(res.count);
    } catch {
      setCount(0);
    }
  };

  useEffect(() => { refreshCount(); }, []);

  return (
    <CartContext.Provider value={{ count, refreshCount }}>
      {children}
    </CartContext.Provider>
  );
}

export const useCart = () => useContext(CartContext);
