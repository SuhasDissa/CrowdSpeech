import { defineStore } from 'pinia'
import { ref } from 'vue'

export type Language = 'en' | 'si' | 'ta'

export const LANGUAGES = {
  en: { code: 'en', label: 'English', native: 'English', dir: 'ltr' },
  si: { code: 'si', label: 'Sinhala', native: 'සිංහල', dir: 'ltr' },
  ta: { code: 'ta', label: 'Tamil', native: 'தமிழ்', dir: 'ltr' },
} as const

export const useLanguageStore = defineStore('language', () => {
  const selected = ref<Language | null>(null)

  function select(lang: Language) {
    selected.value = lang
  }

  function clear() {
    selected.value = null
  }

  return { selected, select, clear }
})
