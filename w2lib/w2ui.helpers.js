import { w2ui, w2tooltip, w2utils } from './w2ui.es6.min.js'

const darkThemeStorageKey = 'w2ui-theme'
const darkThemeChangeEvent = 'w2ui:dark-theme-change'

export const localeWeekStartsStorageKey = 'w2ui-locale-week-starts'
export const localeDateFormatStorageKey = 'w2ui-locale-date-format'
export const localeDatetimeFormatStorageKey = 'w2ui-locale-datetime-format'
export const localeTimeFormatStorageKey = 'w2ui-locale-time-format'

export function getStorageItem(key) {
  try {
    return localStorage.getItem(key)
  } catch (_err) {
    return null
  }
}

export function isDarkTheme() {
  return document.documentElement.classList.contains('dark')
}

export function setDarkTheme(isDark) {
  document.documentElement.classList.toggle('dark', isDark)
  window.dispatchEvent(new CustomEvent(darkThemeChangeEvent, { detail: { isDark } }))
  try {
    localStorage.setItem(darkThemeStorageKey, isDark ? 'dark' : 'light')
  } catch (_err) { }
}

export function onDarkThemeChange(listener) {
  const handleThemeChange = isDark => listener(Boolean(isDark))
  const handleCustomThemeChange = event => handleThemeChange(event.detail?.isDark)
  window.addEventListener(darkThemeChangeEvent, handleCustomThemeChange)
  return () => {
    window.removeEventListener(darkThemeChangeEvent, handleCustomThemeChange)
  }
}

export function w2init() {
  window.w2ui = w2ui
  window.w2tooltip = w2tooltip
  w2utils.settings.dataType = 'JSON'
  w2utils.formatters['text'] = (_, extra) => w2utils.encodeTags(String(extra.value ?? ''))
  w2utils.formatters['dropdown'] = (_, extra) => w2utils.encodeTags(String(extra.value?.text ?? ''))
  w2utils.formatters['nullable'] = (row, extra) => {
    const value = row[extra.field] // nullable
    return value == null ? `<span style="font-style: italic; color: darkgrey;">NULL</span>` : w2utils.encodeTags(String(extra.value))
  }
  w2utils.formatters['text-tooltip'] = (_, extra) => {
    const text = w2utils.encodeTags(String(extra.value ?? ''))
    const encodedBase64 = btoa(encodeURIComponent(text))
    return `<span onmouseenter="w2tooltip.show(this, {'html': decodeURIComponent(atob(('${encodedBase64}'))), 'name': 'tooltip'})" onmouseleave="w2tooltip.hide('tooltip')">${text}</span>`
  }
  w2utils.formatters['dropdown-tooltip'] = (_, extra) => {
    const text = w2utils.encodeTags(String(extra.value?.text ?? ''))
    const encodedBase64 = btoa(encodeURIComponent(text))
    return extra.value?.text == null ? null : `<span onmouseenter="w2tooltip.show(this, {'html': decodeURIComponent(atob(('${encodedBase64}'))), 'name': 'tooltip'})" onmouseleave="w2tooltip.hide('tooltip')">${text}</span>`
  }
  w2utils.formatters['dropdown-multi'] = (_, extra) => {
    let result = ''
    for (const el of extra.value) {
      const text = w2utils.encodeTags(String(el.text ?? ''))
      result += `<span style="color: var(--w2-text); background-color: var(--w2-surface-muted, #eff3f5); border: 1px solid var(--w2-border, #b4d0de); border-radius: 15px; margin: 0 3px 0 3px; padding: 2px 12px; font-size: 11px;">${text}</span>`
    }
    return result
  }
  w2utils.formatters['icon'] = (_, extra) => {
    const src = w2utils.encodeTags(extra.value)
    return extra.value == '' ? null : `<img src="${src}" style="display:block; max-width: 24px; max-height: 24px; margin: auto;"/>`
  }
  w2utils.formatters['icon-sm'] = (_, extra) => {
    const src = w2utils.encodeTags(extra.value)
    return extra.value == '' ? null : `<img src="${src}" style="display:block; max-width: 16px; max-height: 16px; margin: auto;"/>`
  }
  w2utils.formatters['icon-lg'] = (_, extra) => {
    const src = w2utils.encodeTags(extra.value)
    return extra.value == '' ? null : `<img src="${src}" style="display:block; max-width: 72px; max-height: 72px; margin: auto;"/>`
  }
  w2utils.formatters['datetime-local'] = (_, extra) => {
    const d = new Date(extra.value)
    if (Number.isNaN(d.getTime())) {
      return extra.value
    }
    return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`
  }
  w2utils.formatters['datetime-local-ms'] = (_, extra) => {
    const d = new Date(extra.value)
    if (Number.isNaN(d.getTime())) {
      return extra.value
    }
    const pad = (n, l = 2) => String(n).padStart(l, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
      `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}.` +
      `${pad(d.getMilliseconds(), 3)}`
  }
  setDarkTheme(getStorageItem(darkThemeStorageKey) == 'dark')
}

export function w2initLocale(opts = {}) {
  const { defaultWeekStarts, defaultDateFormat, defaultDatetimeFormat, defaultTimeFormat } = opts
  w2utils.locale({
    weekStarts: getStorageItem(localeWeekStartsStorageKey) ?? defaultWeekStarts ?? 'S',
    dateFormat: getStorageItem(localeDateFormatStorageKey) ?? defaultDateFormat ?? 'yyyy-MM-dd',
    datetimeFormat: getStorageItem(localeDatetimeFormatStorageKey) ?? defaultDatetimeFormat ?? 'yyyy-MM-dd hh24:mi:ss',
    timeFormat: getStorageItem(localeTimeFormatStorageKey) ?? defaultTimeFormat ?? 'hh24:mi:ss',
  })
}

export async function w2fetch(opts = {}) {
  const { owner, reload, lock, url, method, headers, body, signal } = opts
  if (owner && lock) {
    owner.lock({ spinner: true, msg: lock })
  }
  try {
    const res = await fetch(url, {
      method: method,
      headers: headers,
      body: body,
      signal: signal,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => {
        return { message: res.statusText }
      })
      throw new Error(err.message)
    }
    const result = await res.json()
    if (owner) {
      if (result.message) {
        owner.message(result.message)
      }
      if (reload) {
        owner.reload()
      }
    }
    return result
  }
  catch (err) {
    if (owner) {
      owner.message(err.toString())
    } else {
      throw err
    }
  }
  finally {
    if (owner && lock) {
      owner.unlock()
    }
  }
}

export async function w2download(opts = {}) {
  const { owner, lock, url, name, method, headers, body, signal } = opts
  if (owner && lock) {
    owner.lock({ spinner: true, msg: lock })
  }
  try {
    const res = await fetch(url, {
      method: method,
      headers: headers,
      body: body,
      signal: signal,
    })
    if (!res.ok) {
      const err = await res.json().catch(() => {
        return { message: res.statusText }
      })
      throw new Error(err.message)
    }
    const blob = await res.blob()
    const objectUrl = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = objectUrl
    a.download = name
    a.click()
    URL.revokeObjectURL(objectUrl)
  }
  catch (err) {
    if (owner) {
      owner.message(err.toString())
    } else {
      throw err
    }
  }
  finally {
    if (owner && lock) {
      owner.unlock()
    }
  }
}

export function w2upload(opts = {}) {
  const { accept, multiple } = opts
  const input = document.createElement('input')
  input.type = 'file'
  if (accept) {
    input.accept = accept
  }
  if (multiple) {
    input.multiple = true
  }
  input.onchange = async event => {
    const body = new FormData()
    for (const file of event.target.files) {
      body.append('files[]', file)
    }
    await w2fetch({ ...opts, body })
  }
  input.click()
}

export async function w2reorder(event, opts = {}) {
  const result = await w2fetch({
    ...opts,
    owner: event.owner,
    lock: 'Reordering...',
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(event.detail),
  })
  if (!result) {
    event.owner.reload()
  }
}

export function registerSidebarSearch(sidebar) {
  // Normalize the string to ensure consistent comparison (garumzīmes)
  const normalize = str => str?.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase()
  return function(value) {
    const search = normalize(value)
    sidebar.expandAll()
    sidebar.search(value, (_, node) => {
      const text = normalize(node.text)
      const parentText = normalize(node.parent.text)
      return text.includes(search) || parentText?.includes(search)
    })
  }
}

export function boolOptions() {
  return { items: [{ id: '1', text: 'True' }, { id: '0', text: 'False' }] }
}

export function remoteListOptions(url, opts = {}) {
  const {
    type = 'list',
    recId = 'id',
    match = 'contains',
    align = 'left',
    cacheMax = 500,
    minLength = 0,
    openOnFocus = true,
  } = opts
  return {
    url: url,
    type: type,
    recId: recId,
    match: match,
    align: align,
    cacheMax: cacheMax,
    minLength: minLength,
    openOnFocus: openOnFocus,
    renderDrop: value => w2utils.encodeTags(value?.text),
  }
}

export function reloadOnSuccess(event) {
  event.onComplete = () => {
    if (event.detail.data?.status == 'success') {
      event.owner.reload()
    }
  }
}

export function searchAllFilter(event) {
  if (event.detail.searchField == 'all') {
    const fields = event.owner.columns.filter(x => x.searchAll).map(x => x.field)
    event.detail.searchData = event.detail.searchData.filter(x => fields.includes(x.field))
  }
}

export function doubleClickNonEditable(event, fn) {
  const isEditable = column => Object.keys(column.editable).length > 0
  if (!isEditable(event.owner.columns[event.detail.column])) {
    fn(event)
  }
}

export function colorizeGridRows(event, fn) {
  event.onComplete = () => {
    event.detail.data?.records?.forEach(x => {
      const color = fn(x)
      if (color != null) {
        const row = event.owner.get(x.id)
        row.w2ui = { style: `color: ${color} !important;` }
        event.owner.refreshRow(x.id)
      }
    })
  }
}

export function setFormRecordFromGridSelection(event, form) {
  event.onComplete = () => {
    const selection = event.owner.getSelection()
    const id = selection.length == 1 ? selection[0] : null
    if (id == null) {
      form.recid = null
      form.clear()
    } else {
      const record = event.owner.get(id)
      form.recid = id
      form.record = record
      form.refresh()
    }
  }
}

