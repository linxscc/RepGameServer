import { Routes, Route } from 'react-router-dom';
import { LanguageProvider } from './contexts/LanguageContext';
import { CartProvider } from './contexts/CartContext';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import HomePage from './pages/HomePage';
import ProductDetail from './pages/ProductDetail';
import Checkout from './pages/Checkout';
import Orders from './pages/Orders';
import Login from './pages/Login';
import Register from './pages/Register';
import SellerDashboard from './pages/seller/Dashboard';
import NewProduct from './pages/seller/NewProduct';
import MyProducts from './pages/seller/MyProducts';
import './Voyara.css';

export default function VoyaraApp() {
  return (
    <LanguageProvider>
      <CartProvider>
        <div className="voyara">
        <Navbar />
        <main className="vy-main">
          <Routes>
            <Route index element={<HomePage />} />
            <Route path="product/:id" element={<ProductDetail />} />
            <Route path="checkout" element={<Checkout />} />
            <Route path="orders" element={<Orders />} />
            <Route path="login" element={<Login />} />
            <Route path="register" element={<Register />} />
            <Route path="seller" element={<SellerDashboard />} />
            <Route path="seller/products/new" element={<NewProduct />} />
            <Route path="seller/products" element={<MyProducts />} />
          </Routes>
        </main>
        <Footer />
        </div>
      </CartProvider>
    </LanguageProvider>
  );
}
