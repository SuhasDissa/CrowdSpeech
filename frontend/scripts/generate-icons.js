#!/usr/bin/env node
// Generates minimal black-background microphone PWA icons
// Run: node scripts/generate-icons.js
// Requires: npm install --save-dev sharp (or just use an SVG icon tool)

import { createCanvas } from 'canvas'
import { writeFileSync, mkdirSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const publicDir = join(__dirname, '..', 'public', 'icons')
mkdirSync(publicDir, { recursive: true })

function drawIcon(size) {
  const canvas = createCanvas(size, size)
  const ctx = canvas.getContext('2d')
  const s = size / 32 // scale factor

  // Background
  ctx.fillStyle = '#000000'
  ctx.fillRect(0, 0, size, size)

  // Microphone body
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(13 * s, 4 * s, 6 * s, 14 * s)

  // Arc
  ctx.strokeStyle = '#ffffff'
  ctx.lineWidth = 2 * s
  ctx.beginPath()
  ctx.arc(16 * s, 16 * s, 8 * s, Math.PI, 0)
  ctx.stroke()

  // Stem
  ctx.beginPath()
  ctx.moveTo(16 * s, 24 * s)
  ctx.lineTo(16 * s, 28 * s)
  ctx.stroke()

  // Base
  ctx.beginPath()
  ctx.moveTo(10 * s, 28 * s)
  ctx.lineTo(22 * s, 28 * s)
  ctx.stroke()

  return canvas.toBuffer('image/png')
}

try {
  writeFileSync(join(publicDir, 'icon-192.png'), drawIcon(192))
  writeFileSync(join(publicDir, 'icon-512.png'), drawIcon(512))
  console.log('Icons generated: icons/icon-192.png, icons/icon-512.png')
} catch (e) {
  console.error('Failed to generate icons (canvas not available). Use any 192x192 and 512x512 PNG files.')
  console.error('Place them at: frontend/public/icons/icon-192.png and icon-512.png')
}
