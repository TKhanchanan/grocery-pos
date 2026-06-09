import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { applyDocumentPreferences, readStoredLanguage, readStoredTheme } from './i18n'
import router from './router'
import './assets/main.css'

applyDocumentPreferences(readStoredLanguage(), readStoredTheme())

const storedTextSize = localStorage.getItem('app_text_size')
if (storedTextSize === 'sm' || storedTextSize === 'base' || storedTextSize === 'lg' || storedTextSize === 'xl') {
  document.documentElement.dataset.textSize = storedTextSize
}

createApp(App).use(createPinia()).use(router).mount('#app')
