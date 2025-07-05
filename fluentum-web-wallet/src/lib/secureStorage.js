import CryptoJS from 'crypto-js'

const SECRET_KEY = import.meta.env.VITE_STORAGE_KEY

export const secureStorage = {
  set(key, value) {
    const encrypted = CryptoJS.AES.encrypt(value, SECRET_KEY).toString()
    localStorage.setItem(key, encrypted)
  },
  get(key) {
    const encrypted = localStorage.getItem(key)
    if (!encrypted) return null
    return CryptoJS.AES.decrypt(encrypted, SECRET_KEY).toString(CryptoJS.enc.Utf8)
  }
} 