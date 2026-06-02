import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { applyDocumentPreferences, readStoredLanguage, readStoredTheme } from './i18n'
import router from './router'
import './assets/main.css'

applyDocumentPreferences(readStoredLanguage(), readStoredTheme())

createApp(App).use(createPinia()).use(router).mount('#app')
