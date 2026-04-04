<template>
  <div class="home-view">
    <!-- Header -->
    <header class="home-header">
      <div class="home-header__inner">
        <span class="home-logo">CROWD<strong>SPEECH</strong></span>
        <span v-if="totalContributions > 0" class="contribution-counter">
          <span class="mdi mdi-microphone" style="font-size:0.9em;"/>
          {{ totalContributions.toLocaleString() }} voices
        </span>
      </div>
    </header>

    <!-- Phone nudge banner (shown on non-mobile or when Bluetooth audio detected) -->
    <div v-if="showPhoneNudge && !nudgeDismissed" class="phone-nudge" role="alert">
      <span class="mdi mdi-cellphone-sound phone-nudge__icon" aria-hidden="true" />
      <span class="phone-nudge__text">
        <strong>Use your phone</strong> for best results —
        laptop mics and Bluetooth headsets compress audio heavily,
        which hurts dataset quality.
        <a :href="pageUrl" class="phone-nudge__qr">Open on phone ↗</a>
      </span>
      <button class="phone-nudge__dismiss" @click="nudgeDismissed = true" aria-label="Dismiss">
        <span class="mdi mdi-close" />
      </button>
    </div>

    <!-- Hero -->
    <main class="home-main">
      <div class="home-hero">
        <h1 class="cs-title">Speak.<br>Record.<br>Contribute.</h1>
        <p class="home-desc">
          Help build open voice datasets for Sinhala, Tamil and English.
          No account needed. Just pick a language and start recording.
        </p>
      </div>

      <!-- Language Selection -->
      <section class="lang-section">
        <p class="cs-subtitle" style="margin-bottom: 1rem;">Select your language</p>
        <div class="lang-grid">
          <button
            v-for="lang in languages"
            :key="lang.code"
            class="lang-card"
            :aria-label="`Record in ${lang.label}`"
            @click="selectLanguage(lang.code)"
          >
            <span class="lang-card__native">{{ lang.native }}</span>
            <span class="lang-card__label cs-subtitle" style="margin-top:0.5rem">{{ lang.label }}</span>
          </button>
        </div>
      </section>

      <!-- How it works -->
      <section class="how-it-works">
        <p class="cs-subtitle" style="margin-bottom: 1.5rem">How it works</p>
        <ol class="how-steps">
          <li v-for="(step, i) in steps" :key="i" class="how-step">
            <span class="how-step__num">{{ i + 1 }}</span>
            <span class="how-step__text">{{ step }}</span>
          </li>
        </ol>
      </section>
    </main>

    <!-- Footer -->
    <footer class="home-footer">
      <span>CrowdSpeech — open voice data for South Asian languages</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { LANGUAGES, type Language } from '@/stores/language'
import { useRecordingStore } from '@/stores/recording'

const router = useRouter()
const recordingStore = useRecordingStore()

const languages = Object.values(LANGUAGES)
const totalContributions = recordingStore.totalContributions
const nudgeDismissed = ref(false)
const pageUrl = window.location.href

// Show nudge when on a non-touch desktop or when a Bluetooth audio device is active
const showPhoneNudge = computed(() => {
  // Already on a touch/mobile device — no nudge needed
  const isMobile = navigator.maxTouchPoints > 0 && /Mobi|Android|iPhone|iPad/i.test(navigator.userAgent)
  if (isMobile) return false
  return true
})

const steps = [
  'Choose your language above',
  'Read the product name shown on screen',
  'Press the big square button (or Space) to record',
  'Listen back, then Submit — or Redo',
  'Automatically advance to the next word',
]

function selectLanguage(code: string) {
  router.push({ name: 'contribute', params: { language: code as Language } })
}

onMounted(() => {
  recordingStore.fetchStats()
})
</script>

<style scoped>
.home-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  color: #000000;
}

.home-header {
  border-bottom: 2px solid #000000;
  padding: 1rem 1.5rem;
}

.home-header__inner {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.home-logo {
  font-size: 1.25rem;
  letter-spacing: -0.02em;
  color: #000;
}
.home-logo strong {
  font-weight: 900;
}

.home-main {
  flex: 1;
  max-width: 900px;
  width: 100%;
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem;
}

.home-hero {
  margin-bottom: 3rem;
}

.home-desc {
  font-size: 1.1rem;
  line-height: 1.6;
  color: #333;
  max-width: 540px;
  margin-top: 1.5rem;
}

.lang-section {
  margin-bottom: 4rem;
}

.lang-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0;
  border: 2px solid #000000;
}

.lang-card {
  border: none !important;
  border-right: 2px solid #000000 !important;
  background: none;
  color: #000;
  cursor: pointer;
  padding: 2rem 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 160px;
  transition: background 0.1s, color 0.1s;
  font-family: inherit;
}
.lang-card:last-child {
  border-right: none !important;
}
.lang-card:hover, .lang-card:focus-visible {
  background: #000000;
  color: #ffffff;
  outline: none;
}
.lang-card:hover .lang-card__label {
  color: #eeeeee;
}

.lang-card__native {
  font-size: clamp(2rem, 5vw, 3rem);
  font-weight: 700;
  line-height: 1;
}

.lang-card__label {
  font-size: 0.75rem;
  color: #666;
  margin-top: 0.5rem;
}

.how-it-works {
  border-top: 2px solid #000;
  padding-top: 2rem;
}

.how-steps {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.how-step {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.875rem 0;
  border-bottom: 1px solid #eeeeee;
}

.how-step__num {
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.1em;
  color: #666;
  min-width: 1.5rem;
  padding-top: 0.1em;
}

.how-step__text {
  font-size: 1rem;
  line-height: 1.5;
  color: #000;
}

/* Phone nudge banner */
.phone-nudge {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.875rem 1.25rem;
  background: #000;
  color: #fff;
  font-size: 0.875rem;
  line-height: 1.5;
}

.phone-nudge__icon {
  font-size: 1.25rem;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.phone-nudge__text {
  flex: 1;
}

.phone-nudge__qr {
  color: #fff;
  text-underline-offset: 2px;
  margin-left: 0.25rem;
  white-space: nowrap;
}

.phone-nudge__dismiss {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  font-size: 1.1rem;
  padding: 0;
  flex-shrink: 0;
  opacity: 0.7;
  line-height: 1;
  margin-top: 0.1rem;
}
.phone-nudge__dismiss:hover { opacity: 1; }

.home-footer {
  border-top: 1px solid #eeeeee;
  padding: 1rem 1.5rem;
  font-size: 0.75rem;
  color: #999;
  text-align: center;
}

@media (max-width: 600px) {
  .lang-grid {
    grid-template-columns: 1fr;
  }
  .lang-card {
    border-right: none !important;
    border-bottom: 2px solid #000 !important;
  }
  .lang-card:last-child {
    border-bottom: none !important;
  }
}
</style>
