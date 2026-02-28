import { w2ui, w2tooltip, w2utils } from './w2ui.es6.min.js'

export function w2init() {
  window.w2ui = w2ui
  window.w2tooltip = w2tooltip
  w2utils.settings.dataType = 'JSON'
  w2utils.formatters['text'] = (_, extra) => w2utils.encodeTags(String(extra.value ?? ''))
  w2utils.formatters['dropdown'] = (_, extra) => w2utils.encodeTags(String(extra.value?.text ?? ''))
  w2utils.formatters['tooltip-text'] = (_, extra) => {
    const text = w2utils.encodeTags(String(extra.value ?? ''))
    const encodedBase64 = btoa(encodeURIComponent(text))
    return `<span onmouseenter="w2tooltip.show(this, {'html': decodeURIComponent(atob(('${encodedBase64}'))), 'name': 'tooltip'})" onmouseleave="w2tooltip.hide('tooltip')">${text}</span>`
  }
  w2utils.formatters['tooltip-dropdown'] = (_, extra) => {
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
}

export async function w2fetch(opts = {}) {
  const { owner, reload, lock, url, method, headers, body, timeout = 5000 } = opts
  if (owner && lock) {
    owner.lock({ spinner: true, msg: lock })
  }
  try {
    const res = await fetch(url, {
      method: method,
      headers: headers,
      body: body,
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
  }
  catch (err) {
    if (owner) {
      owner.message(err.toString())
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

export function registerSidebarSearch(sidebar) {
  return function(value) {
    // Normalize the string to ensure consistent comparison
    const normalizeString = str => str.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    sidebar.expandAll()
    sidebar.search(value, (str, node) => {
      const str1 = normalizeString(str.toLowerCase())
      const str2 = normalizeString(node.text.toLowerCase())
      return str2.indexOf(str1) != -1
    })
  }
}

export function boolOptions() {
  return { items: [{ id: '1', text: 'True' }, { id: '0', text: 'False' }] }
}

export function remoteListOptions(url) {
  return {
    url: url,
    type: 'list',
    recId: 'id',
    match: 'contains',
    align: 'left',
    cacheMax: 5000,
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

