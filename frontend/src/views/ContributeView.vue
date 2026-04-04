<template>
  <div class="contribute-view">
    <!-- Top bar -->
    <header class="contribute-header">
      <button class="back-btn" @click="goHome" aria-label="Back to language selection">
        <span class="mdi mdi-arrow-left" />
        <span>{{ LANGUAGES[language as Language].native }}</span>
      </button>
      <div class="session-info">
        <span class="contribution-counter" v-if="store.sessionCount > 0">
          <span class="mdi mdi-check" style="font-size:0.85em"/>
          {{ store.sessionCount }} this session
        </span>
      </div>
    </header>

    <!-- Main content -->
    <main class="contribute-main">
      <!-- Error state -->
      <div v-if="store.phase === 'error'" class="error-panel">
        <p class="error-msg">{{ store.errorMessage }}</p>
        <button class="cs-btn cs-btn--outline" @click="retryLoad">Try again</button>
      </div>

      <!-- Loading keyword -->
      <div v-else-if="!store.currentKeyword" class="loading-panel">
        <div class="loading-dots">
          <span /><span /><span />
        </div>
        <p class="cs-subtitle" style="margin-top:1rem">Loading keyword…</p>
      </div>

      <!-- Active recording UI -->
      <template v-else>
        <!-- Keyword display -->
        <div class="keyword-display">
          {{ keywordText(store.currentKeyword, language) }}
        </div>

        <!-- Category pill -->
        <div class="category-pill" v-if="store.currentKeyword.category">
          {{ store.currentKeyword.category }}
        </div>

        <!-- Waveform -->
        <div class="waveform-wrap">
          <WaveformDisplay
            :analyser="analyserNode"
            :audio-url="store.phase === 'reviewing' || store.phase === 'submitting' ? store.audioUrl : null"
            :is-live="store.phase === 'recording'"
            :playback-progress="playbackProgress"
          />
        </div>

        <!-- Recording progress bar -->
        <div class="cs-progress" role="progressbar" :aria-valuenow="recordProgress" aria-valuemin="0" aria-valuemax="100">
          <div class="cs-progress__fill" :style="{ width: recordProgress + '%' }" />
        </div>

        <!-- Timer text -->
        <div class="timer-text" v-if="store.phase === 'recording'">
          {{ elapsed.toFixed(1) }}s / 10s
        </div>

        <!-- ── IDLE: big record button ── -->
        <div v-if="store.phase === 'idle'" class="record-area">
          <button
            class="record-btn record-btn--idle"
            @click="startRecording"
            @mousedown="startRecording"
            @touchstart.prevent="startRecording"
            aria-label="Start recording (hold Space or hold button)"
          >
            <span class="mdi mdi-microphone" style="font-size:2rem; display:block; margin-bottom:0.5rem" />
            RECORD
          </button>
          <p class="kbd-hint">Hold <kbd>Space</kbd> to record</p>
        </div>

        <!-- ── RECORDING: stop button ── -->
        <div v-else-if="store.phase === 'recording'" class="record-area">
          <button
            class="record-btn record-btn--recording"
            @mouseup="stopRecording"
            @touchend.prevent="stopRecording"
            aria-label="Stop recording (release to stop)"
          >
            <span class="mdi mdi-stop" style="font-size:2rem; display:block; margin-bottom:0.5rem" />
            STOP
          </button>
          <p class="kbd-hint">Release <kbd>Space</kbd> to stop</p>
        </div>

        <!-- ── REVIEWING: play/redo/submit ── -->
        <div v-else-if="store.phase === 'reviewing' || store.phase === 'submitting'" class="review-area">
          <div class="playback-controls">
            <button class="cs-btn cs-btn--icon" @click="togglePlayback" :aria-label="isPlaying ? 'Pause' : 'Play'">
              <span :class="isPlaying ? 'mdi mdi-pause' : 'mdi mdi-play'" />
            </button>
            <span class="playback-time">{{ formatTime(playbackCurrentTime) }} / {{ formatTime(store.duration) }}</span>
          </div>
          <div class="action-row" style="margin-top: 1.5rem">
            <button
              class="cs-btn cs-btn--outline"
              @click="store.redo(); stopPlayback()"
              :disabled="store.phase === 'submitting'"
            >
              <span class="mdi mdi-refresh" />
              Redo
            </button>
            <button
              class="cs-btn cs-btn--filled"
              @click="submitRecording"
              :disabled="store.phase === 'submitting'"
            >
              <span v-if="store.phase === 'submitting'" class="mdi mdi-loading cs-spin" />
              <span v-else class="mdi mdi-check" />
              Submit
            </button>
          </div>
        </div>

        <!-- ── SUCCESS flash ── -->
        <div v-else-if="store.phase === 'success'" class="success-flash">
          <span class="mdi mdi-check-circle" style="font-size:3rem" />
          <p>Saved!</p>
        </div>
      </template>
    </main>

    <!-- Validation snackbar -->
    <v-snackbar v-model="validationSnack" location="bottom" color="black" :timeout="3000">
      {{ validationMessage }}
    </v-snackbar>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { LANGUAGES, type Language } from '@/stores/language'
import { useRecordingStore, keywordText } from '@/stores/recording'
import WaveformDisplay from '@/components/WaveformDisplay.vue'

const props = defineProps<{ language: string }>()
const router = useRouter()
const store = useRecordingStore()

const language = props.language as Language

// Recording state
const mediaRecorder = ref<MediaRecorder | null>(null)
const audioChunks = ref<Blob[]>([])
const audioContext = ref<AudioContext | null>(null)
const analyserNode = ref<AnalyserNode | null>(null)
const recordingStart = ref(0)
const elapsed = ref(0)
let elapsedInterval = 0
let autoStopTimeout = 0

const recordProgress = computed(() =>
  store.phase === 'recording' ? Math.min((elapsed.value / 10) * 100, 100) : 0
)

// Playback state
const audioElement = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const playbackCurrentTime = ref(0)
const playbackProgress = ref(0)
let playbackInterval = 0

// Validation snackbar
const validationSnack = ref(false)
const validationMessage = ref('')

function showValidation(msg: string) {
  validationMessage.value = msg
  validationSnack.value = true
}

// ── Navigation ──────────────────────────────────────────────────────────────
function goHome() {
  cleanup()
  store.reset()
  router.push({ name: 'home' })
}

function retryLoad() {
  store.phase = 'idle'
  store.errorMessage = ''
  store.fetchNextKeyword(language)
}

// ── Recording ───────────────────────────────────────────────────────────────
async function startRecording() {
  if (store.phase !== 'idle') return

  let stream: MediaStream
  try {
    stream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false })
  } catch {
    showValidation('Microphone access denied. Please allow microphone use.')
    return
  }

  // Set up Web Audio analyser for waveform
  audioContext.value = new AudioContext()
  const source = audioContext.value.createMediaStreamSource(stream)
  analyserNode.value = audioContext.value.createAnalyser()
  analyserNode.value.fftSize = 2048
  source.connect(analyserNode.value)

  audioChunks.value = []

  const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
    ? 'audio/webm;codecs=opus'
    : MediaRecorder.isTypeSupported('audio/ogg;codecs=opus')
    ? 'audio/ogg;codecs=opus'
    : 'audio/webm'

  mediaRecorder.value = new MediaRecorder(stream, { mimeType })
  mediaRecorder.value.ondataavailable = e => {
    if (e.data.size > 0) audioChunks.value.push(e.data)
  }
  mediaRecorder.value.onstop = onRecordingStop

  mediaRecorder.value.start(100) // collect every 100ms
  store.phase = 'recording'
  recordingStart.value = Date.now()
  elapsed.value = 0

  elapsedInterval = window.setInterval(() => {
    elapsed.value = (Date.now() - recordingStart.value) / 1000
  }, 100)

  // Auto-stop after 10 seconds
  autoStopTimeout = window.setTimeout(() => {
    stopRecording()
  }, 10000)
}

function stopRecording() {
  if (!mediaRecorder.value || store.phase !== 'recording') return
  clearInterval(elapsedInterval)
  clearTimeout(autoStopTimeout)
  mediaRecorder.value.stop()
  mediaRecorder.value.stream.getTracks().forEach(t => t.stop())
}

function onRecordingStop() {
  const dur = (Date.now() - recordingStart.value) / 1000
  elapsed.value = dur

  // Validation: minimum 0.5s, minimum loudness check
  if (dur < 0.5) {
    showValidation('Recording too short. Please hold the button longer.')
    store.phase = 'idle'
    cleanupAudioContext()
    return
  }

  const blob = new Blob(audioChunks.value, { type: mediaRecorder.value?.mimeType ?? 'audio/webm' })

  // Rough size check as proxy for silence (< 1KB for >0.5s is likely silent)
  if (blob.size < 1024) {
    showValidation('Recording seems too quiet. Please speak louder.')
    store.phase = 'idle'
    cleanupAudioContext()
    return
  }

  cleanupAudioContext()
  store.setAudio(blob, dur)
  setupPlayback()
}

// ── Playback ────────────────────────────────────────────────────────────────
function setupPlayback() {
  if (audioElement.value) {
    audioElement.value.pause()
    audioElement.value.src = ''
  }
  if (!store.audioUrl) return

  const audio = new Audio(store.audioUrl)
  audioElement.value = audio

  audio.ontimeupdate = () => {
    playbackCurrentTime.value = audio.currentTime
    playbackProgress.value = audio.duration ? audio.currentTime / audio.duration : 0
  }
  audio.onended = () => {
    isPlaying.value = false
    playbackCurrentTime.value = 0
    playbackProgress.value = 0
  }
}

function togglePlayback() {
  const audio = audioElement.value
  if (!audio) return
  if (isPlaying.value) {
    audio.pause()
    isPlaying.value = false
  } else {
    audio.currentTime = 0
    audio.play()
    isPlaying.value = true
  }
}

function stopPlayback() {
  if (audioElement.value) {
    audioElement.value.pause()
    audioElement.value.currentTime = 0
  }
  isPlaying.value = false
  playbackCurrentTime.value = 0
  playbackProgress.value = 0
}

// ── Submit ───────────────────────────────────────────────────────────────────
async function submitRecording() {
  stopPlayback()
  await store.submit(language)
}

// ── Cleanup ──────────────────────────────────────────────────────────────────
function cleanupAudioContext() {
  analyserNode.value = null
  if (audioContext.value) {
    audioContext.value.close()
    audioContext.value = null
  }
}

function cleanup() {
  clearInterval(elapsedInterval)
  clearTimeout(autoStopTimeout)
  clearInterval(playbackInterval)
  stopPlayback()
  if (mediaRecorder.value?.state !== 'inactive') {
    mediaRecorder.value?.stop()
    mediaRecorder.value?.stream.getTracks().forEach(t => t.stop())
  }
  cleanupAudioContext()
}

// ── Keyboard shortcuts ────────────────────────────────────────────────────────
function onKeydown(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.code === 'Space') {
    e.preventDefault()
    if (e.repeat) return  // ignore key-repeat events while held
    if (store.phase === 'idle') startRecording()
    else if (store.phase === 'reviewing') togglePlayback()
  }
}

function onKeyup(e: KeyboardEvent) {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
  if (e.code === 'Space') {
    e.preventDefault()
    if (store.phase === 'recording') stopRecording()
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function formatTime(sec: number) {
  const s = Math.floor(sec)
  const ms = Math.floor((sec - s) * 10)
  return `${s}.${ms}s`
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
onMounted(() => {
  if (!Object.keys(LANGUAGES).includes(props.language)) {
    router.push({ name: 'home' })
    return
  }
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('keyup', onKeyup)
  store.reset()
  store.fetchNextKeyword(language)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('keyup', onKeyup)
  cleanup()
})

// When store resets (after submit), reset playback UI
watch(() => store.phase, (p) => {
  if (p === 'idle') {
    stopPlayback()
  } else if (p === 'reviewing') {
    setupPlayback()
  }
})
</script>

<style scoped>
.contribute-view {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  color: #000000;
}

/* Header */
.contribute-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 2px solid #000;
}

.back-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
  font-size: 1rem;
  font-weight: 700;
  color: #000;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0;
  text-decoration: none;
}
.back-btn:hover { text-decoration: underline; }

/* Main */
.contribute-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.5rem 1rem 3rem;
  max-width: 640px;
  width: 100%;
  margin: 0 auto;
}

/* Keyword */
.keyword-display {
  font-size: clamp(2.5rem, 12vw, 5rem);
  font-weight: 700;
  line-height: 1.1;
  text-align: center;
  padding: 2rem 1rem 1rem;
  min-height: 130px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.category-pill {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: #666;
  border: 1px solid #ccc;
  padding: 0.2rem 0.6rem;
  margin-bottom: 1.5rem;
}

/* Waveform */
.waveform-wrap {
  width: 100%;
  margin-bottom: 0;
  border-top: 1px solid #eee;
  border-bottom: 1px solid #eee;
  overflow: hidden;
}

/* Timer */
.timer-text {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: #666;
  margin: 0.25rem 0 1rem;
  align-self: flex-end;
}

/* Record area */
.record-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  margin-top: 2rem;
}

.record-btn {
  width: min(260px, 75vw);
  height: min(260px, 75vw);
  border: 2px solid #000 !important;
  font-family: inherit;
  font-size: 1.25rem;
  font-weight: 900;
  letter-spacing: 0.15em;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transition: background 0.1s, color 0.1s;
}

.record-btn--idle {
  background: #fff;
  color: #000;
}
.record-btn--idle:hover {
  background: #000;
  color: #fff;
}

.record-btn--recording {
  background: #000;
  color: #fff;
  animation: pulse-border 0.8s infinite;
}

@keyframes pulse-border {
  0%, 100% { border-color: #000; }
  50% { border-color: #555; }
}

kbd {
  font-family: monospace;
  font-size: 0.8em;
  border: 1px solid #000;
  padding: 0.1em 0.3em;
}

/* Review area */
.review-area {
  width: 100%;
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.playback-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.playback-time {
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: #666;
}

/* Action row */
.action-row {
  display: flex;
  width: 100%;
  max-width: 400px;
}

/* Generic buttons */
.cs-btn {
  flex: 1;
  height: 52px;
  border: none;
  font-family: inherit;
  font-size: 0.9rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition: background 0.1s, color 0.1s;
}

.cs-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cs-btn--outline {
  background: #fff;
  color: #000;
  border: 2px solid #000 !important;
}
.cs-btn--outline:hover:not(:disabled) {
  background: #eee;
}

.cs-btn--filled {
  background: #000;
  color: #fff;
  border: 2px solid #000 !important;
}
.cs-btn--filled:hover:not(:disabled) {
  background: #333;
}

.cs-btn--icon {
  width: 52px;
  height: 52px;
  flex: none;
  background: #fff;
  border: 2px solid #000 !important;
  font-size: 1.5rem;
  color: #000;
}
.cs-btn--icon:hover { background: #000; color: #fff; }

/* Success flash */
.success-flash {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  margin-top: 3rem;
  font-weight: 900;
  font-size: 1.25rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

/* Error/Loading */
.error-panel, .loading-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  margin-top: 4rem;
  text-align: center;
}

.error-msg {
  font-size: 1rem;
  color: #000;
  border: 2px solid #000;
  padding: 1rem 1.5rem;
  max-width: 360px;
}

/* Spinner */
.cs-spin {
  animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Loading dots */
.loading-dots {
  display: flex;
  gap: 0.5rem;
}
.loading-dots span {
  width: 10px;
  height: 10px;
  background: #000;
  display: block;
  animation: dot-blink 1s infinite;
}
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dot-blink {
  0%, 100% { opacity: 0.2; }
  50% { opacity: 1; }
}
</style>
