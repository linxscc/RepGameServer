import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { LanguageProvider } from './contexts/LanguageContext';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { CartProvider } from './contexts/CartContext';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import HomePage from './pages/HomePage';
import ProductDetail from './pages/ProductDetail';
import Checkout from './pages/Checkout';
import Orders from './pages/Orders';
import OrderDetail from './pages/OrderDetail';
import CartPage from './pages/CartPage';
import PaymentPage from './pages/PaymentPage';
import Login from './pages/Login';
import Register from './pages/Register';
import ForgotPassword from './pages/ForgotPassword';
import AccountPage from './pages/AccountPage';
import SellerDashboard from './pages/seller/Dashboard';
import NewProduct from './pages/seller/NewProduct';
import MyProducts from './pages/seller/MyProducts';
import PendingProducts from './pages/admin/PendingProducts';
import './Voyara.css';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();
  if (isLoading) {
    return <div className="vy-section" style={{ textAlign: 'center', padding: '4rem 0' }}><div className="vy-spinner" /></div>;
  }
  if (!isAuthenticated) {
    return <Navigate to="/voyara/login" state={{ from: location.pathname + location.search }} replace />;
  }
  return <>{children}</>;
}

export default function VoyaraApp() {
  return (
    <LanguageProvider>
      <AuthProvider>
        <CartProvider>
          <div className="voyara">
          <Navbar />
          <main className="vy-main">
            <Routes>
              {/* Public */}
              <Route index element={<HomePage />} />
              <Route path="product/:id" element={<ProductDetail />} />
              <Route path="login" element={<Login />} />
              <Route path="register" element={<Register />} />
              <Route path="forgot-password" element={<ForgotPassword />} />

              {/* Protected */}
              <Route path="cart" element={<ProtectedRoute><CartPage /></ProtectedRoute>} />
              <Route path="checkout" element={<ProtectedRoute><Checkout /></ProtectedRoute>} />
              <Route path="payment" element={<ProtectedRoute><PaymentPage /></ProtectedRoute>} />
              <Route path="orders" element={<ProtectedRoute><Orders /></ProtectedRoute>} />
              <Route path="order/:id" element={<ProtectedRoute><OrderDetail /></ProtectedRoute>} />
              <Route path="account" element={<ProtectedRoute><AccountPage /></ProtectedRoute>} />

              {/* Seller (also protected) */}
              <Route path="seller" element={<ProtectedRoute><SellerDashboard /></ProtectedRoute>} />
              <Route path="seller/products/new" element={<ProtectedRoute><NewProduct /></ProtectedRoute>} />
              <Route path="seller/products" element={<ProtectedRoute><MyProducts /></ProtectedRoute>} />

              {/* Admin (protected) */}
              <Route path="admin/products" element={<ProtectedRoute><PendingProducts /></ProtectedRoute>} />
            </Routes>
          </main>
          <Footer />
          </div>
        </CartProvider>
      </AuthProvider>
    </LanguageProvider>
  );
}
