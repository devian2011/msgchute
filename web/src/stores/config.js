import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useConfigStore = defineStore('config', () => {
    const host = ref('')
    const apiKey = ref('')
    const apiHeader = ref('Authorization')

    function setConfig(newHost, newHeader, newKey) {
        host.value = newHost
        apiKey.value = newKey
        apiHeader.value = newHeader
    }

    function clearConfig() {
        host.value = ''
        apiKey.value = ''
        apiHeader.value = ''
    }

    return { host, apiKey, apiHeader, setConfig, clearConfig }
}, {
    persist: true
})
