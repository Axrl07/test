import { StrictMode } from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import UserProvider from './authentication/provider';
import "./styles/index.css"

// el div root si existe pero verificamos por seguridad
const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <StrictMode>
      <UserProvider>
      <App />
      </UserProvider>
    </StrictMode>,
  );
} else {
  console.error("Elemento Root no encontrado");
}