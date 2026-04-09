export const WAILS_BACKEND_ERR = 'WAILS_BACKEND'

export function wailsBackendUserMessage() {
  return '请在桌面应用中运行：在项目根目录执行 wails dev，或运行已打包程序；不要单独用浏览器打开前端开发地址。'
}

function goApp() {
  if (typeof window === 'undefined') return null
  return window.go?.main?.App ?? null
}

const defaultConfigFallback = () => ({
  delayMin: 1,
  delayMax: 4,
  timeout: 30,
  imageTimeout: 60,
  maxRetries: 3,
  concurrentDownloads: 1,
  downloadMode: 'lake',
  failOnImageError: false
})

export async function GetDefaultConfig() {
  const app = goApp()
  if (app) return app.GetDefaultConfig()
  return defaultConfigFallback()
}

export async function SelectDirectory() {
  const app = goApp()
  if (app) return app.SelectDirectory()
  const path =
    typeof window !== 'undefined'
      ? window.prompt('开发模式：请输入输出目录的完整路径', '')
      : ''
  return path || ''
}

export async function GetAllTasks() {
  const app = goApp()
  if (app) return app.GetAllTasks()
  return []
}

export async function AddTask(arg1, arg2, arg3, arg4) {
  const app = goApp()
  if (!app) throw new Error(WAILS_BACKEND_ERR)
  return app.AddTask(arg1, arg2, arg3, arg4)
}

export async function RemoveTask(arg1) {
  const app = goApp()
  if (!app) throw new Error(WAILS_BACKEND_ERR)
  return app.RemoveTask(arg1)
}

export async function StartTask(arg1) {
  const app = goApp()
  if (!app) throw new Error(WAILS_BACKEND_ERR)
  return app.StartTask(arg1)
}

export async function CancelTask(arg1) {
  const app = goApp()
  if (!app) throw new Error(WAILS_BACKEND_ERR)
  return app.CancelTask(arg1)
}

export async function StartAllPendingTasks() {
  const app = goApp()
  if (!app) throw new Error(WAILS_BACKEND_ERR)
  return app.StartAllPendingTasks()
}

export async function ClearCompletedTasks() {
  const app = goApp()
  if (!app) throw new Error(WAILS_BACKEND_ERR)
  return app.ClearCompletedTasks()
}

export async function ValidateURL(arg1) {
  const app = goApp()
  if (app) return app.ValidateURL(arg1)
  return (
    typeof arg1 === 'string' &&
    arg1.length > 0 &&
    (arg1.includes('yuque.com') || arg1.includes('www.yuque.com'))
  )
}

/**
 * @param {string} eventName
 * @param {(data: unknown) => void} callback
 */
export function EventsOn(eventName, callback) {
  if (typeof window !== 'undefined' && window.runtime?.EventsOnMultiple) {
    return window.runtime.EventsOnMultiple(eventName, callback, -1)
  }
  return () => {}
}
