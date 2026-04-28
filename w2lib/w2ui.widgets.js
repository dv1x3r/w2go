import { w2form, w2grid, w2layout, w2popup, w2sidebar, w2utils } from './w2ui.es6.min.js'
import * as helpers from './w2ui.helpers.js'

const sqlQueryStorageKey = 'w2ui-sql-explorer-query'
const vimModeStorageKey = 'w2ui-sql-explorer-vim-mode'

export function openLocalePopup() {
  const form = new w2form({
    name: `w2localeForm`,
    fields: [
      {
        field: 'week',
        type: 'text',
        html: {
          label: 'Week Starts',
          attr: 'style="width:100%;" placeholder="S"',
          span: 5,
        },
      },
      {
        field: 'date',
        type: 'text',
        html: {
          label: 'Date Format',
          attr: 'style="width:100%;" placeholder="yyyy-MM-dd"',
          span: 5,
        },
      },
      {
        field: 'datetime',
        type: 'text',
        html: {
          label: 'Datetime Format',
          attr: 'style="width:100%;" placeholder="yyyy-MM-dd hh24:mi:ss"',
          span: 5,
        },
      },
      {
        field: 'time',
        type: 'text',
        html: {
          label: 'Time Format',
          attr: 'style="width:100%;" placeholder="hh24:mi:ss"',
          span: 5,
        },
      },
    ],
    record: {
      week: helpers.getStorageItem(helpers.localeWeekStartsStorageKey),
      date: helpers.getStorageItem(helpers.localeDateFormatStorageKey),
      datetime: helpers.getStorageItem(helpers.localeDatetimeFormatStorageKey),
      time: helpers.getStorageItem(helpers.localeTimeFormatStorageKey),
    },
    actions: {
      Save() {
        try {
          if (this.record.week) {
            localStorage.setItem(helpers.localeWeekStartsStorageKey, this.record.week)
          } else {
            localStorage.removeItem(helpers.localeWeekStartsStorageKey)
          }
          if (this.record.date) {
            localStorage.setItem(helpers.localeDateFormatStorageKey, this.record.date)
          } else {
            localStorage.removeItem(helpers.localeDateFormatStorageKey)
          }
          if (this.record.datetime) {
            localStorage.setItem(helpers.localeDatetimeFormatStorageKey, this.record.datetime)
          } else {
            localStorage.removeItem(helpers.localeDatetimeFormatStorageKey)
          }
          if (this.record.time) {
            localStorage.setItem(helpers.localeTimeFormatStorageKey, this.record.time)
          } else {
            localStorage.removeItem(helpers.localeTimeFormatStorageKey)
          }
          helpers.w2initLocale()
        } catch (_err) { }
        w2popup.close()
      },
      Cancel() { w2popup.close() },
    },
  })

  w2popup.open({
    title: 'Locale Settings',
    body: '<div id="w2locale-form" style="width: 100%; height: 100%;"></div>',
    width: 400, height: 300, showMax: false, resizable: false,
  })
    .then(() => form.render('#w2locale-form'))
    .close(() => form.destroy())
}

export function createSqlExplorerLayout(opts = {}) {
  const { url, darkTheme = 'dracula', initialQuery = '' } = opts

  let abortController = null
  let isRunning = false
  let editor = null
  let stopWatchingTheme = null

  const grid = new w2grid({
    name: 'sqlExplorerGrid-' + Date.now(),
    selectType: 'cell',
    recordHeight: 28,
    show: {
      footer: true,
      toolbar: false,
      lineNumbers: true,
    },
    onDelete: function(event) {
      event.preventDefault()
    },
  })

  const sidebar = new w2sidebar({
    name: 'sqlExplorerSidebar-' + Date.now(),
    levelPadding: 8,
    topHTML: '<div style="margin-top:2px;padding:3px 5px;height:36px;"><input id="sql-explorer-search" class="w2ui-input" style="width:100%;" placeholder="Search..."></div>',
    onContextMenu: function(event) {
      const isTableNode = event.object?.query != null
      this.menu = isTableNode ? [{
        id: 'select-1000-rows',
        text: 'Select Top 1000 Rows',
        icon: 'fa fa-arrow-pointer',
      }] : []
    },
    onMenuClick: async function(event) {
      if (event.detail.item?.id == 'select-1000-rows') {
        const node = this.get(event.target)
        const query = node.query
        editor.setValue(query)
        editor.focus()
        await executeQuery(query)
      }
    },
    onRender: async function(event) {
      await event.complete
      const search = helpers.registerSidebarSearch(sidebar)
      const el = document.getElementById('sql-explorer-search')
      el.addEventListener('keyup', e => search(e.target.value))
    }
  })

  function isVimModeEnabled() {
    try {
      return localStorage.getItem(vimModeStorageKey) == 'true'
    }
    catch (_err) {
      return false
    }
  }

  function setVimMode(enabled) {
    editor?.setOption('keyMap', enabled ? 'vim' : 'default')
    try {
      localStorage.setItem(vimModeStorageKey, enabled ? 'true' : 'false')
    }
    catch (_err) {
    }
  }

  function buildSelectRowsQuery(dbName, dbTable, dbColumns) {
    const quote = value => `"${String(value).replaceAll('"', '""')}"`
    const database = quote(dbName)
    const table = quote(dbTable)
    const columns = dbColumns.map(col => `\t${quote(col.name)}`).join(',\n')
    return `SELECT\n${columns}\nFROM ${database}.${table}\nLIMIT 1000;`
  }

  function setSchemaSidebar(schema) {
    sidebar.nodes = schema.databases.map((db, dbIndex) => ({
      id: `db-${dbIndex}`,
      text: db.name,
      icon: 'fa fa-database',
      expanded: true,
      nodes: [
        {
          id: `db-${dbIndex}-tables`,
          text: 'tables',
          icon: 'fa fa-table-list',
          expanded: true,
          nodes: db.tables.map((table, tableIndex) => ({
            id: `db-${dbIndex}-table-${tableIndex}`,
            text: table.name,
            icon: 'fa fa-table',
            expanded: false,
            query: buildSelectRowsQuery(db.name, table.name, table.columns),
            nodes: table.columns.map((col, colIndex) => ({
              id: `db-${dbIndex}-table-${tableIndex}-col-${colIndex}`,
              text: `${col.name}${col.type ? ` (${col.type})` : ''}`,
              icon: col.pk ? 'fa fa-key' : 'fa',
            })),
          })),
        },
      ],
    }))
    sidebar.refresh()
  }

  function setSchemaAutocomplete(schema) {
    const tables = {}
    schema.databases.forEach(db => {
      db.tables.forEach(table => {
        if (!tables[table.name]) {
          tables[table.name] = []
        }
        table.columns.forEach(col => {
          if (!tables[table.name].includes(col.name)) {
            tables[table.name].push(col.name)
          }
        })
      })
    })
    editor.setOption('hintOptions', { tables })
  }

  async function executeQuery(queryOverride = null) {
    if (isRunning) {
      return
    }

    isRunning = true
    abortController = new AbortController()

    const toolbar = editorLayout.get('main').toolbar
    toolbar.disable('run')
    toolbar.enable('cancel')

    const query = queryOverride ?? (editor.getSelection() || editor.getValue())
    try {
      localStorage.setItem(sqlQueryStorageKey, query)
    } catch (_err) { }

    try {
      const startTime = performance.now()
      const result = await helpers.w2fetch({
        owner: grid,
        lock: 'Executing...',
        url: url,
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
        signal: abortController.signal,
      })
      if (result) {
        grid.lock({ spinner: true, msg: 'Processing...' })
        grid.columns = result.columns
          .filter(col => col.toLowerCase() !== 'recid')
          .map(col => ({ field: col, text: w2utils.encodeTags(col), render: 'nullable', min: 80, sortable: true, editable: {} }))
        grid.records = result.records.map((row, i) => {
          const { recid, ...rest } = row;
          return { recid: i + 1, ...rest };
        })
        grid.total = result.total
        grid.sortData = []
        // grid.refresh()
        grid.reset()
        grid.selectNone()
        grid.columnAutoSize()
        grid.unlock()
      }
      const elapsed = ((performance.now() - startTime) / 1000).toFixed(3)
      grid.status(`Query Executed ${elapsed} seconds`)
    }
    finally {
      isRunning = false
      abortController = null
      toolbar.enable('run')
      toolbar.disable('cancel')
    }
  }

  const editorLayout = new w2layout({
    name: 'sqlEditorLayout-' + Date.now(),
    panels: [
      {
        type: 'left',
        size: 250,
        html: sidebar,
      },
      {
        type: 'main',
        toolbar: {
          items: [
            {
              type: 'button',
              id: 'run',
              text: 'Run',
              tooltip: 'Shift-Enter executes selection or full query<br>Shift-Alt-Enter does the same and refreshes the sidebar schema',
              icon: 'fa fa-play',
              onClick: async function() {
                await executeQuery()
              },
            },
            {
              type: 'button',
              id: 'cancel',
              text: 'Cancel',
              tooltip: 'Shift-Esc cancels a running query',
              icon: 'fa fa-stop',
              disabled: true,
              onClick: function() {
                abortController?.abort('The query has been cancelled')
              },
            },
            { type: 'spacer' },
            {
              type: 'button',
              id: 'reopen',
              text: 'Reopen last query',
              onClick: function() {
                editor.setValue(localStorage.getItem(sqlQueryStorageKey) ?? '')
              },
            },
            {
              type: 'check',
              id: 'vim',
              text: 'Vim Mode',
              tooltip: 'Ctrl-` for maximum efficiency',
              icon: 'fa-brands fa-vim',
              checked: isVimModeEnabled(),
              onClick: async function(event) {
                await event.complete
                setVimMode(event.detail.item.checked)
              },
            },
          ],
        },
        html: `<style>.CodeMirror-hints{ z-index: 9999 !important; }</style><div id="sql-explorer-editor" style="height:100%;"></div>`,
      },
    ],
    onRender: async function(event) {
      await event.complete
      CodeMirror.Vim.defineEx('write', 'w', async () => {
        await executeQuery()
      })
      editor = CodeMirror(document.getElementById('sql-explorer-editor'), {
        lineNumbers: true,
        keyMap: isVimModeEnabled() ? 'vim' : 'default',
        mode: 'text/x-sql',
        theme: helpers.isDarkTheme() ? darkTheme : 'default',
        value: initialQuery,
        extraKeys: {
          "Ctrl-Space": cm => {
            CodeMirror.commands.autocomplete(cm, null, { completeSingle: false })
          },
          'Shift-Enter': async () => {
            await executeQuery()
          },
          'Shift-Alt-Enter': async () => {
            await executeQuery()
            document.getElementById('sql-explorer-search').value = ''
            const schema = await helpers.w2fetch({ url: url, method: 'GET' })
            setSchemaSidebar(schema)
            setSchemaAutocomplete(schema)
          },
          'Shift-Esc': () => {
            abortController?.abort('The query has been cancelled')
          },
          'Ctrl-`': () => {
            const toolbar = editorLayout.get('main').toolbar
            toolbar.click('vim')
            toolbar.disable('vim')
            toolbar.enable('vim')
          }
        },
      })
      stopWatchingTheme = helpers.onDarkThemeChange(isDark => {
        editor.setOption('theme', isDark ? darkTheme : 'default')
      })
      editor.setSize('100%', '100%')
      const schema = await helpers.w2fetch({ url: url, method: 'GET' })
      setSchemaSidebar(schema)
      setSchemaAutocomplete(schema)
    }
  })

  return new w2layout({
    name: 'sqlExplorerLayout-' + Date.now(),
    panels: [
      {
        type: 'top',
        size: '50%',
        resizable: true,
        html: editorLayout,
      },
      {
        type: 'main',
        html: grid,
      },
    ],
    onDestroy: function() {
      stopWatchingTheme?.()
      abortController?.abort()
      editorLayout.destroy()
      sidebar.destroy()
      grid.destroy()
    }
  })
}

