/**
 * 应用根渲染
 * 1. 初始化 i18n 资源
 * 2. 挂载根组件到 #root
 */
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './i18n';
import './index.css';
import App from './App.jsx';

const ROOT_ELEMENT_ID = 'root';

createRoot(document.getElementById(ROOT_ELEMENT_ID)).render(
  <StrictMode>
    <App />
  </StrictMode>
);
