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
const SESSION_DONE_PREFIX = 'crowdspeech_survey_done_'

function loadFromStorage(): Partial<SurveyAnswers> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

export function isCompletedForLanguage(language: string): boolean {
  return sessionStorage.getItem(SESSION_DONE_PREFIX + language) === '1'
}

export function markCompletedForLanguage(language: string): void {
  sessionStorage.setItem(SESSION_DONE_PREFIX + language, '1')
}

export const SURVEY_OPTIONS: Record<keyof SurveyAnswers, string[]> = {
  age_group: ['Under 18', '18–24', '25–34', '35–44', '45–54', '55–64', '65+', 'Prefer not to say'],
  gender: ['Male', 'Female'],
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

/** Only these fields are shown and required on the survey form. */
export const VISIBLE_SURVEY_FIELDS: (keyof SurveyAnswers)[] = ['gender', 'primary_language']

const HIDDEN_DEFAULTS: Pick<SurveyAnswers, 'age_group' | 'country' | 'accent'> = {
  age_group: '18–24',
  country: 'Sri Lanka',
  accent: 'Sri Lankan',
}

function pickVisible(saved: Partial<SurveyAnswers>, field: keyof SurveyAnswers): string {
  const value = saved[field] ?? ''
  return SURVEY_OPTIONS[field].includes(value) ? value : ''
}

function withHiddenDefaults(visible: Pick<SurveyAnswers, 'gender' | 'primary_language'>): SurveyAnswers {
  return {
    ...HIDDEN_DEFAULTS,
    gender: visible.gender,
    primary_language: visible.primary_language,
    region: '',
    education: '',
    years_speaking: '',
    occupation: '',
    speech_condition: '',
  }
}

function isComplete(answers: SurveyAnswers): boolean {
  return VISIBLE_SURVEY_FIELDS.every(field => answers[field] !== '')
}

export const useSurveyStore = defineStore('survey', () => {
  const saved = loadFromStorage()

  const answers = ref<SurveyAnswers>(withHiddenDefaults({
    gender: pickVisible(saved, 'gender'),
    primary_language: pickVisible(saved, 'primary_language'),
  }))

  const completed = ref(isComplete(answers.value))

  // Persist to localStorage whenever answers change
  watch(answers, (val) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(val))
    completed.value = isComplete(val)
  }, { deep: true })

  function reset() {
    Object.assign(answers.value, withHiddenDefaults({ gender: '', primary_language: '' }))
    completed.value = false
    localStorage.removeItem(STORAGE_KEY)
  }

  return { answers, completed, reset }
})
