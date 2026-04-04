<template>
  <canvas
    ref="canvas"
    class="waveform-canvas"
    :width="canvasWidth"
    :height="canvasHeight"
    aria-hidden="true"
  />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'

const props = withDefaults(defineProps<{
  analyser?: AnalyserNode | null
  audioUrl?: string | null
  isLive?: boolean       // true = live mic feed, false = static decoded waveform
  playbackProgress?: number  // 0–1 for playback position
  height?: number
}>(), {
  analyser: null,
  audioUrl: null,
  isLive: false,
  playbackProgress: 0,
  height: 80,
})

const canvas = ref<HTMLCanvasElement | null>(null)
const canvasWidth = ref(600)
const canvasHeight = props.height

// Live animation
let animationId = 0

// Static decoded waveform data
const staticWaveform = ref<Float32Array | null>(null)

// Decode audio buffer for static display
async function decodeAudio(url: string) {
  try {
    const resp = await fetch(url)
    const arrayBuf = await resp.arrayBuffer()
    const ctx = new AudioContext()
    const decoded = await ctx.decodeAudioData(arrayBuf)
    // Downsample to canvas width
    const channelData = decoded.getChannelData(0)
    const samples = canvasWidth.value
    const blockSize = Math.floor(channelData.length / samples)
    const reduced = new Float32Array(samples)
    for (let i = 0; i < samples; i++) {
      let sum = 0
      for (let j = 0; j < blockSize; j++) {
        sum += Math.abs(channelData[i * blockSize + j])
      }
      reduced[i] = sum / blockSize
    }
    staticWaveform.value = reduced
    ctx.close()
    drawStatic()
  } catch {
    // non-critical
  }
}

function drawStatic() {
  const el = canvas.value
  if (!el || !staticWaveform.value) return
  const ctx = el.getContext('2d')
  if (!ctx) return

  const w = canvasWidth.value
  const h = canvasHeight
  const data = staticWaveform.value
  const max = Math.max(...Array.from(data), 0.001)

  ctx.clearRect(0, 0, w, h)

  const barW = Math.max(1, w / data.length - 0.5)
  const progress = props.playbackProgress

  for (let i = 0; i < data.length; i++) {
    const barH = (data[i] / max) * (h * 0.85)
    const x = (i / data.length) * w
    const y = (h - barH) / 2

    // Played portion = black, unplayed = light gray
    ctx.fillStyle = i / data.length <= progress ? '#000000' : '#cccccc'
    ctx.fillRect(x, y, barW, barH)
  }

  // Progress line
  if (progress > 0) {
    ctx.fillStyle = '#000000'
    ctx.fillRect(progress * w - 1, 0, 2, h)
  }
}

function drawLive() {
  const el = canvas.value
  if (!el || !props.analyser) {
    animationId = requestAnimationFrame(drawLive)
    return
  }
  const ctx = el.getContext('2d')
  if (!ctx) return

  const w = canvasWidth.value
  const h = canvasHeight
  const bufLen = props.analyser.frequencyBinCount
  const dataArr = new Uint8Array(bufLen)
  props.analyser.getByteTimeDomainData(dataArr)

  ctx.clearRect(0, 0, w, h)
  ctx.strokeStyle = '#000000'
  ctx.lineWidth = 2
  ctx.beginPath()

  const sliceW = w / bufLen
  let x = 0
  for (let i = 0; i < bufLen; i++) {
    const v = dataArr[i] / 128.0
    const y = (v * h) / 2
    if (i === 0) ctx.moveTo(x, y)
    else ctx.lineTo(x, y)
    x += sliceW
  }
  ctx.lineTo(w, h / 2)
  ctx.stroke()

  animationId = requestAnimationFrame(drawLive)
}

function startLive() {
  if (animationId) cancelAnimationFrame(animationId)
  animationId = requestAnimationFrame(drawLive)
}

function stopLive() {
  if (animationId) {
    cancelAnimationFrame(animationId)
    animationId = 0
  }
}

function clearCanvas() {
  const el = canvas.value
  if (!el) return
  const ctx = el.getContext('2d')
  if (!ctx) return
  ctx.clearRect(0, 0, canvasWidth.value, canvasHeight)
  // Draw flat line
  ctx.strokeStyle = '#eeeeee'
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(0, canvasHeight / 2)
  ctx.lineTo(canvasWidth.value, canvasHeight / 2)
  ctx.stroke()
}

// Resize observer
let ro: ResizeObserver | null = null

onMounted(() => {
  const el = canvas.value
  if (!el) return
  canvasWidth.value = el.offsetWidth || 600
  clearCanvas()

  ro = new ResizeObserver(entries => {
    for (const entry of entries) {
      canvasWidth.value = Math.floor(entry.contentRect.width)
      if (staticWaveform.value) drawStatic()
    }
  })
  ro.observe(el.parentElement ?? el)
})

onUnmounted(() => {
  stopLive()
  ro?.disconnect()
})

// React to prop changes
watch(() => props.isLive, (live) => {
  if (live) {
    staticWaveform.value = null
    startLive()
  } else {
    stopLive()
  }
}, { immediate: true })

watch(() => props.audioUrl, (url) => {
  if (url) {
    stopLive()
    decodeAudio(url)
  } else {
    staticWaveform.value = null
    clearCanvas()
  }
})

watch(() => props.playbackProgress, () => {
  if (staticWaveform.value) drawStatic()
})
</script>
