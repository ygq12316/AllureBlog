// 头像圆形裁剪:Settings 与 VisitorSetup 共用(此前 60 行级复制两份)。
// pickImage 进裁剪模式;applyCrop 上传并返回 URL;cancelCrop 取消。
import { ref, watch } from 'vue'
import { uploadFile } from '../api/upload'

export function useAvatarCrop() {
  const cropMode = ref(false)
  const cropSrc = ref('')
  const cropImgStyle = ref({})
  const cropZoom = ref(1)
  let cropImg = null, dragging = false, dragStart = { x: 0, y: 0 }, imgPos = { x: 0, y: 0 }

  // 缩放变化即重绘(原 Settings 版缺此联动,拖动前缩放不生效)
  watch(cropZoom, () => updateCropStyle())

  async function pickImage({ file }) {
    cropSrc.value = URL.createObjectURL(file.file)
    cropImg = new Image()
    cropImg.src = cropSrc.value
    await new Promise(r => cropImg.onload = r)
    imgPos = { x: 0, y: 0 }; cropZoom.value = 1
    updateCropStyle()
    cropMode.value = true
  }

  function updateCropStyle() {
    if (!cropImg) return
    const size = 200
    const scale = (size / Math.min(cropImg.width, cropImg.height)) * cropZoom.value
    cropImgStyle.value = {
      width: (cropImg.width * scale) + 'px',
      height: (cropImg.height * scale) + 'px',
      transform: `translate(${imgPos.x}px, ${imgPos.y}px)`,
    }
  }

  function onWheel(e) {
    cropZoom.value = Math.max(0.5, Math.min(3, cropZoom.value + (e.deltaY > 0 ? -0.1 : 0.1)))
  }
  function startDrag(e) {
    dragging = true
    const ev = e.touches ? e.touches[0] : e
    dragStart = { x: ev.clientX - imgPos.x, y: ev.clientY - imgPos.y }
  }
  function onDrag(e) {
    if (!dragging) return
    const ev = e.touches ? e.touches[0] : e
    imgPos.x = ev.clientX - dragStart.x
    imgPos.y = ev.clientY - dragStart.y
    updateCropStyle()
  }
  function endDrag() { dragging = false }

  async function applyCrop() {
    if (!cropImg) return null
    const size = 200
    const cv = document.createElement('canvas')
    cv.width = size; cv.height = size
    const ctx = cv.getContext('2d')
    ctx.beginPath(); ctx.arc(size / 2, size / 2, size / 2, 0, Math.PI * 2); ctx.clip()
    const scale = (size / Math.min(cropImg.width, cropImg.height)) * cropZoom.value
    const sx = -imgPos.x / scale, sy = -imgPos.y / scale, sw = size / scale
    ctx.drawImage(cropImg, sx, sy, sw, sw, 0, 0, size, size)
    const blob = await new Promise(r => cv.toBlob(r, 'image/png'))
    const fd = new FormData(); fd.append('file', blob, 'avatar.png')
    const { url } = await uploadFile(fd)
    cropMode.value = false
    URL.revokeObjectURL(cropSrc.value)
    return url
  }

  function cancelCrop() {
    cropMode.value = false
    if (cropSrc.value) URL.revokeObjectURL(cropSrc.value)
  }

  return { cropMode, cropSrc, cropImgStyle, cropZoom, pickImage, onWheel, startDrag, onDrag, endDrag, applyCrop, cancelCrop }
}
