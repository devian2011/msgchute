import { createApp } from 'vue'
import { createBootstrap } from 'bootstrap-vue-next'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

import App from './App.vue'
import router from './router'

import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue-next/dist/bootstrap-vue-next.css'

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(createBootstrap({
    components: true,
    directives: true,
}))
app.use(pinia)
app.use(router)

app.mount('#app')
