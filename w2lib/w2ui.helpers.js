import { w2ui, w2tooltip, w2utils } from './w2ui.es6.min.js'

const darkThemeStorageKey = 'w2ui-theme'
const darkThemeChangeEvent = 'w2ui:dark-theme-change'

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
  w2utils.formatters['icon-small'] = (_, extra) => {
    const src = w2utils.encodeTags(extra.value)
    return extra.value == '' ? null : `<img src="${src}" style="max-width: 24px; max-height: 24px; margin: auto;"/>`
  }
  w2utils.formatters['icon-normal'] = (_, extra) => {
    const src = w2utils.encodeTags(extra.value)
    return extra.value == '' ? null : `<img src="${src}" style="max-width: 72px; max-height: 72px; margin: auto;"/>`
  }
  setDarkTheme(localStorage.getItem(darkThemeStorageKey) == 'dark')
}

export async function w2fetch(opts = {}) {
  const { owner, reload, lock, url, method, headers, body, signal, timeout = 5000 } = opts
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
    if (result.message) {
      w2utils.notify(result.message, { timeout })
    }
    if (owner && reload) {
      owner.reload()
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

export function remoteListOptions(url, cacheMax = 500) {
  return {
    url: url,
    type: 'list',
    recId: 'id',
    match: 'contains',
    align: 'left',
    cacheMax: cacheMax,
    minLength: 0,
    openOnFocus: true,
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

