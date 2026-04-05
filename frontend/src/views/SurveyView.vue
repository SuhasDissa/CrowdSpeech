<template>
  <div class="survey-view">
    <header class="survey-header">
      <div class="survey-header__inner">
        <button class="back-btn" @click="router.push('/')" aria-label="Back to home">
          <span class="mdi mdi-arrow-left" />
        </button>
        <span class="home-logo">CROWD<strong>SPEECH</strong></span>
        <span class="lang-badge">{{ LANGUAGES[language as Language]?.label ?? language }}</span>
      </div>
    </header>

    <main class="survey-main">
      <h1 class="cs-title survey-title">Tell us about yourself</h1>
      <p class="survey-desc">
        These answers help improve dataset quality. They are saved locally and submitted
        anonymously with each recording.
      </p>

      <form class="survey-form" @submit.prevent="onSubmit">
        <div
          v-for="field in SURVEY_FIELDS"
          :key="field"
          class="survey-field"
          :class="{ 'survey-field--error': submitted && !store.answers[field] }"
        >
          <label :for="`field-${field}`" class="survey-label">
            {{ SURVEY_LABELS[field] }}
            <span class="required-star" aria-hidden="true">*</span>
          </label>
          <div class="select-wrapper">
            <select
              :id="`field-${field}`"
              v-model="store.answers[field]"
              class="survey-select"
              :aria-required="true"
            >
              <option value="" disabled>Select…</option>
              <option
                v-for="opt in SURVEY_OPTIONS[field]"
                :key="opt"
                :value="opt"
              >{{ opt }}</option>
            </select>
            <span class="mdi mdi-chevron-down select-arrow" aria-hidden="true" />
          </div>
          <p v-if="submitted && !store.answers[field]" class="field-error" role="alert">
            Please select an option.
          </p>
        </div>

        <div class="survey-actions">
          <button type="submit" class="cs-btn-primary survey-submit">
            Continue to recording
            <span class="mdi mdi-arrow-right" style="margin-left: 0.5rem" />
          </button>
        </div>
      </form>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { LANGUAGES, type Language } from '@/stores/language'
import { useSurveyStore, SURVEY_OPTIONS, SURVEY_LABELS, markCompletedForLanguage, type SurveyAnswers } from '@/stores/survey'

const props = defineProps<{ language: string }>()
const router = useRouter()
const store = useSurveyStore()

const SURVEY_FIELDS = Object.keys(SURVEY_LABELS) as (keyof SurveyAnswers)[]
const submitted = ref(false)

function onSubmit() {
  submitted.value = true
  if (!store.completed) return
  markCompletedForLanguage(props.language)
  router.push({ name: 'contribute', params: { language: props.language as Language } })
}
</script>

<style scoped>
.survey-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  color: #000000;
}

.survey-header {
  border-bottom: 2px solid #000;
  padding: 1rem 1.5rem;
}

.survey-header__inner {
  max-width: 700px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.back-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.25rem;
  padding: 0;
  line-height: 1;
  color: #000;
}
.back-btn:hover { opacity: 0.6; }

.home-logo {
  flex: 1;
  font-size: 1.25rem;
  letter-spacing: -0.02em;
}
.home-logo strong { font-weight: 900; }

.lang-badge {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  border: 2px solid #000;
  padding: 0.2rem 0.6rem;
}

.survey-main {
  flex: 1;
  max-width: 700px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem;
}

.survey-title {
  margin-bottom: 0.75rem;
}

.survey-desc {
  font-size: 0.95rem;
  color: #555;
  line-height: 1.6;
  margin-bottom: 2.5rem;
  max-width: 520px;
}

.survey-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.survey-field {
  border-top: 1px solid #e0e0e0;
  padding: 1.25rem 0;
}
.survey-field:last-of-type {
  border-bottom: 1px solid #e0e0e0;
}
.survey-field--error .survey-label {
  color: #c00;
}
.survey-field--error .survey-select {
  border-color: #c00;
}

.survey-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: #000;
}

.required-star {
  color: #c00;
  margin-left: 0.15rem;
}

.select-wrapper {
  position: relative;
  display: inline-block;
  width: 100%;
  max-width: 380px;
}

.survey-select {
  width: 100%;
  appearance: none;
  -webkit-appearance: none;
  background: #fff;
  border: 2px solid #000;
  padding: 0.625rem 2.5rem 0.625rem 0.875rem;
  font-size: 1rem;
  font-family: inherit;
  color: #000;
  cursor: pointer;
  outline: none;
}
.survey-select:focus {
  outline: 3px solid #000;
  outline-offset: 2px;
}

.select-arrow {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  font-size: 1.25rem;
}

.field-error {
  font-size: 0.8rem;
  color: #c00;
  margin-top: 0.375rem;
}

.survey-actions {
  margin-top: 2rem;
}

.cs-btn-primary {
  display: inline-flex;
  align-items: center;
  background: #000;
  color: #fff;
  border: 2px solid #000;
  padding: 0.875rem 2rem;
  font-size: 1rem;
  font-family: inherit;
  font-weight: 700;
  cursor: pointer;
  letter-spacing: 0.02em;
  transition: background 0.1s, color 0.1s;
}
.cs-btn-primary:hover {
  background: #fff;
  color: #000;
}

@media (max-width: 600px) {
  .select-wrapper { max-width: 100%; }
}
</style>
