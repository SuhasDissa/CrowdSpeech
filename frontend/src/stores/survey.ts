import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export interface SurveyAnswers {
  age_group: string
  gender: string
  country: string
  primary_language: string
  accent: string
  region: string
  education: string
  years_speaking: string
  occupation: string
  speech_condition: string
}

const STORAGE_KEY = 'crowdspeech_survey'

function loadFromStorage(): Partial<SurveyAnswers> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

export const SURVEY_OPTIONS: Record<keyof SurveyAnswers, string[]> = {
  age_group: ['Under 18', '18–24', '25–34', '35–44', '45–54', '55–64', '65+', 'Prefer not to say'],
  gender: ['Male', 'Female', 'Non-binary', 'Prefer not to say'],
  country: [
    'Sri Lanka', 'India', 'Bangladesh', 'Pakistan', 'Nepal', 'Maldives',
    'United Kingdom', 'United States', 'Canada', 'Australia',
    'Germany', 'France', 'Other',
  ],
  primary_language: ['Sinhala', 'Tamil', 'English', 'Hindi', 'Urdu', 'Bengali', 'Other'],
  accent: ['Sri Lankan', 'Indian', 'British', 'American', 'Australian', 'Other'],
  region: [
    'Western Province', 'Central Province', 'Southern Province',
    'Northern Province', 'Eastern Province', 'North Western Province',
    'North Central Province', 'Uva Province', 'Sabaragamuwa Province',
    'Outside Sri Lanka',
  ],
  education: [
    'Primary school', 'Secondary school', 'Diploma / Certificate',
    "Bachelor's degree", "Master's degree", 'Doctorate', 'Prefer not to say',
  ],
  years_speaking: ['Less than 1 year', '1–5 years', '5–10 years', '10–20 years', '20+ years', 'Native speaker'],
  occupation: [
    'Student', 'Professional / Office', 'Technical / Engineering',
    'Academic / Research', 'Service / Hospitality', 'Trade / Skilled work',
    'Homemaker', 'Retired', 'Other',
  ],
  speech_condition: [
    'None', 'Mild stutter / disfluency', 'Stutter', 'Lisp',
    'Hearing impairment', 'Other', 'Prefer not to say',
  ],
}

export const SURVEY_LABELS: Record<keyof SurveyAnswers, string> = {
  age_group: 'Age group',
  gender: 'Gender',
  country: 'Country of residence',
  primary_language: 'Primary language',
  accent: 'Accent',
  region: 'Region (Sri Lanka)',
  education: 'Highest education level',
  years_speaking: 'Years speaking recording language',
  occupation: 'Occupation',
  speech_condition: 'Speech condition',
}

export const useSurveyStore = defineStore('survey', () => {
  const saved = loadFromStorage()

  const answers = ref<SurveyAnswers>({
    age_group: saved.age_group ?? '',
    gender: saved.gender ?? '',
    country: saved.country ?? '',
    primary_language: saved.primary_language ?? '',
    accent: saved.accent ?? '',
    region: saved.region ?? '',
    education: saved.education ?? '',
    years_speaking: saved.years_speaking ?? '',
    occupation: saved.occupation ?? '',
    speech_condition: saved.speech_condition ?? '',
  })

  const completed = ref(
    Object.values(answers.value).every(v => v !== '')
  )

  // Persist to localStorage whenever answers change
  watch(answers, (val) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(val))
    completed.value = Object.values(val).every(v => v !== '')
  }, { deep: true })

  function reset() {
    const keys = Object.keys(answers.value) as (keyof SurveyAnswers)[]
    keys.forEach(k => { answers.value[k] = '' })
    completed.value = false
    localStorage.removeItem(STORAGE_KEY)
  }

  return { answers, completed, reset }
})
