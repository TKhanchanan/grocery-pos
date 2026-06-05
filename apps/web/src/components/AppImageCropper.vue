<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppButton from './AppButton.vue'

const props = withDefaults(defineProps<{ file?: File | null; aspectRatio?: number; outputType?: string; quality?: number }>(), {
  file: null,
  aspectRatio: 1,
  outputType: 'image/webp',
  quality: 0.9,
})

const emit = defineEmits<{ cropped: [file: File]; cancel: [] }>()
const imageURL = ref('')
const zoom = ref(1)
const positionX = ref(50)
const positionY = ref(50)

const objectPosition = computed(() => `${positionX.value}% ${positionY.value}%`)

watch(() => props.file, (file) => {
  if (imageURL.value) URL.revokeObjectURL(imageURL.value)
  imageURL.value = file ? URL.createObjectURL(file) : ''
  zoom.value = 1
  positionX.value = 50
  positionY.value = 50
}, { immediate: true })

async function crop() {
  if (!props.file || !imageURL.value) return
  const image = new Image()
  image.src = imageURL.value
  await image.decode()

  const outputSize = 720
  const canvas = document.createElement('canvas')
  canvas.width = outputSize
  canvas.height = Math.round(outputSize / props.aspectRatio)
  const context = canvas.getContext('2d')
  if (!context) return

  const scale = Math.max(canvas.width / image.width, canvas.height / image.height) * zoom.value
  const drawWidth = image.width * scale
  const drawHeight = image.height * scale
  const x = (canvas.width - drawWidth) * (positionX.value / 100)
  const y = (canvas.height - drawHeight) * (positionY.value / 100)
  context.drawImage(image, x, y, drawWidth, drawHeight)

  canvas.toBlob((blob) => {
    if (!blob || !props.file) return
    const name = props.file.name.replace(/\.[^.]+$/, '.webp')
    emit('cropped', new File([blob], name, { type: props.outputType }))
  }, props.outputType, props.quality)
}
</script>

<template>
  <div v-if="imageURL" class="grid gap-4">
    <div class="overflow-hidden rounded-2xl border border-slate-200 bg-slate-100 dark:border-slate-700 dark:bg-slate-950" :style="{ aspectRatio }">
      <img class="h-full w-full object-cover" :src="imageURL" alt="" :style="{ transform: `scale(${zoom})`, objectPosition }" />
    </div>
    <div class="grid gap-3 text-sm">
      <label class="grid gap-1.5 font-semibold text-slate-700 dark:text-slate-200">
        ซูม
        <input v-model.number="zoom" class="accent-brand-600" type="range" min="1" max="2.4" step="0.05" />
      </label>
      <label class="grid gap-1.5 font-semibold text-slate-700 dark:text-slate-200">
        ตำแหน่งแนวนอน
        <input v-model.number="positionX" class="accent-brand-600" type="range" min="0" max="100" step="1" />
      </label>
      <label class="grid gap-1.5 font-semibold text-slate-700 dark:text-slate-200">
        ตำแหน่งแนวตั้ง
        <input v-model.number="positionY" class="accent-brand-600" type="range" min="0" max="100" step="1" />
      </label>
    </div>
    <div class="flex justify-center gap-2">
      <AppButton variant="secondary" @click="$emit('cancel')">ยกเลิก</AppButton>
      <AppButton icon="check-circle" @click="crop">ใช้รูปนี้</AppButton>
    </div>
  </div>
</template>
