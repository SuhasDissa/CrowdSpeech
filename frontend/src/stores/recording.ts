import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Language } from './language'
import { useSurveyStore } from './survey'

export interface Keyword {
  id: string
  text: string      // canonical English label
  text_si: string   // Sinhala display text
  text_ta: string   // Tamil display text
  category: string
  count_en: number
  count_si: number
  count_ta: number
  sample_target: number
}

export type RecordingPhase = 'idle' | 'recording' | 'reviewing' | 'submitting' | 'success' | 'error'

const API_BASE = '/api'

/** Returns the number of recordings already collected for this keyword+language */
export function keywordCount(kw: Keyword, lang: Language): number {
  if (lang === 'si') return kw.count_si
  if (lang === 'ta') return kw.count_ta
  return kw.count_en
}

/** Returns the display text for a keyword in the given language */
export function keywordText(kw: Keyword, lang: Language): string {
  if (lang === 'si') return kw.text_si || kw.text
  if (lang === 'ta') return kw.text_ta || kw.text
  return kw.text
}

/** Returns the count field name for a language */
function countField(lang: Language) {
  return `count_${lang}` as 'count_en' | 'count_si' | 'count_ta'
}

export const useRecordingStore = defineStore('recording', () => {
  const currentKeyword = ref<Keyword | null>(null)
  const phase = ref<RecordingPhase>('idle')
  const audioBlob = ref<Blob | null>(null)
  const audioUrl = ref<string | null>(null)
  const duration = ref(0)
  const errorMessage = ref('')
  const totalContributions = ref(0)
  const sessionCount = ref(0)

  const isRecording = computed(() => phase.value === 'recording')
  const isReviewing = computed(() => phase.value === 'reviewing')
  const isSubmitting = computed(() => phase.value === 'submitting')

  async function fetchNextKeyword(language: Language): Promise<void> {
    try {
      const sortField = countField(language)
      const langFilter = language === 'si'
        ? 'text_si != ""'
        : language === 'ta'
        ? 'text_ta != ""'
        : 'text != ""'

      // Fetch the 20 least-recorded keywords, then pick one at random.
      // This gives even distribution while avoiding everyone seeing the same
      // word simultaneously.
      const resp = await fetch(
        `${API_BASE}/collections/keywords/records` +
        `?filter=${encodeURIComponent(langFilter)}&sort=${sortField}&perPage=20`,
        { headers: { 'Content-Type': 'application/json' } }
      )
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
      const data = await resp.json()
      if (!data.items || data.items.length === 0) {
        throw new Error('No keywords available')
      }

      // Avoid showing the same word twice in a row
      const pool = (data.items as Keyword[]).filter(k => k.id !== currentKeyword.value?.id)
      const candidates = pool.length > 0 ? pool : data.items as Keyword[]
      currentKeyword.value = candidates[Math.floor(Math.random() * candidates.length)]
    } catch (e) {
      errorMessage.value = e instanceof Error ? e.message : 'Failed to fetch keyword'
      phase.value = 'error'
    }
  }

  function setAudio(blob: Blob, durationSec: number) {
    if (audioUrl.value) URL.revokeObjectURL(audioUrl.value)
    audioBlob.value = blob
    audioUrl.value = URL.createObjectURL(blob)
    duration.value = durationSec
    phase.value = 'reviewing'
  }

  function redo() {
    if (audioUrl.value) URL.revokeObjectURL(audioUrl.value)
    audioBlob.value = null
    audioUrl.value = null
    duration.value = 0
    phase.value = 'idle'
    errorMessage.value = ''
  }

  async function submit(language: Language): Promise<void> {
    if (!audioBlob.value || !currentKeyword.value) return

    phase.value = 'submitting'
    errorMessage.value = ''

    try {
      const surveyStore = useSurveyStore()
      const formData = new FormData()
      const ext = audioBlob.value.type.includes('ogg') ? 'ogg' : 'webm'
      formData.append('audio', audioBlob.value, `recording.${ext}`)
      formData.append('language', language)
      formData.append('keyword', currentKeyword.value.id)
      formData.append('duration', String(duration.value))

      // Append survey demographics
      const s = surveyStore.answers
      formData.append('age_group', s.age_group)
      formData.append('gender', s.gender)
      formData.append('country', s.country)
      formData.append('primary_language', s.primary_language)
      formData.append('accent', s.accent)
      formData.append('region', s.region)
      formData.append('education', s.education)
      formData.append('years_speaking', s.years_speaking)
      formData.append('occupation', s.occupation)
      formData.append('speech_condition', s.speech_condition)

      const resp = await fetch(`${API_BASE}/collections/recordings/records`, {
        method: 'POST',
        body: formData,
      })

      if (!resp.ok) {
        const err = await resp.json().catch(() => ({ message: `HTTP ${resp.status}` }))
        throw new Error(err.message || `HTTP ${resp.status}`)
      }

      phase.value = 'success'
      sessionCount.value++
      totalContributions.value++

      await new Promise(r => setTimeout(r, 800))

      if (audioUrl.value) URL.revokeObjectURL(audioUrl.value)
      audioBlob.value = null
      audioUrl.value = null
      duration.value = 0
      phase.value = 'idle'
      errorMessage.value = ''

      await fetchNextKeyword(language)
    } catch (e) {
      errorMessage.value = e instanceof Error ? e.message : 'Submission failed'
      phase.value = 'error'
    }
  }

  async function fetchStats(): Promise<void> {
    try {
      const resp = await fetch(`${API_BASE}/stats`)
      if (!resp.ok) return
      const data = await resp.json()
      totalContributions.value = data.total_recordings ?? 0
    } catch {
      // non-critical
    }
  }

  function reset() {
    redo()
    currentKeyword.value = null
    sessionCount.value = 0
    errorMessage.value = ''
    phase.value = 'idle'
  }

  return {
    currentKeyword,
    phase,
    audioBlob,
    audioUrl,
    duration,
    errorMessage,
    totalContributions,
    sessionCount,
    isRecording,
    isReviewing,
    isSubmitting,
    fetchNextKeyword,
    setAudio,
    redo,
    submit,
    fetchStats,
    reset,
  }
})
