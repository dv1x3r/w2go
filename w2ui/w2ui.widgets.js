import { w2grid, w2layout, w2sidebar } from './w2ui.es6.min.js'
import { w2fetch, registerSidebarSearch } from './w2ui.helpers.js'

export function createSqlExplorerLayout(opts = {}) {
  const { url } = opts

  const sidebar = new w2sidebar({
    name: 'sqlExplorerSidebar',
    levelPadding: 8,
    topHTML: '<div style="margin-top:2px;padding:3px 5px;height:36px;"><input id="sql-explorer-search" class="w2ui-input" style="width:100%;" placeholder="Search..."></div>',
    onRender: async function(event) {
      await event.complete
      const search = registerSidebarSearch(sidebar)
      const el = document.getElementById('sql-explorer-search')
      el.addEventListener('keyup', e => search(e.target.value))
    }
  })

  async function fetchSchema() {
    const el = document.getElementById('sql-explorer-search')
    if (el) {
      el.value = null
    }
    const result = await w2fetch({ url: url, method: 'GET' })
    sidebar.nodes = result.databases.map((db, dbIndex) => ({
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

  const grid = new w2grid({
    name: 'sqlExplorerGrid',
    show: {
      footer: true,
      toolbar: false,
      lineNumbers: true,
    },
  })

  let abortController = null
  let isRunning = false

  async function executeQuery() {
    if (isRunning) {
      abortController?.abort()
      return
    }

    const toolbar = layout.get('top').toolbar
    const button = toolbar.items[0]
    button.text = 'Cancel (Escape)'
    button.icon = 'fa fa-stop'
    toolbar.refresh()

    isRunning = true
    abortController = new AbortController()

    const textarea = document.getElementById('sql-explorer-query')
    const selection = textarea.value.substring(textarea.selectionStart, textarea.selectionEnd)
    const query = selection || textarea.value
    const startTime = performance.now()

    try {
      const result = await w2fetch({
        owner: grid,
        lock: 'Executing...',
        url: url,
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
        signal: abortController.signal,
      })
      if (result) {
        grid.columns = result.columns.map(column => ({ field: column, text: column, render: 'nullable', min: 100 }))
        grid.records = result.records.map((row, i) => ({ recid: i + 1, ...row }))
        grid.total = result.total
        grid.refresh()
      }
      const elapsed = ((performance.now() - startTime) / 1000).toFixed(3)
      grid.status(`Query Executed ${elapsed} seconds`)
    }
    finally {
      isRunning = false
      abortController = null
      button.text = 'Execute (Alt+Enter)'
      button.icon = 'fa fa-play'
      toolbar.refresh()
    }
  }

  const layout = new w2layout({
    name: 'sqlExplorerInnerLayout',
    panels: [
      {
        type: 'top',
        size: 200,
        style: 'border-left: 1px solid #efefef;',
        resizable: true,
        toolbar: {
          items: [
            {
              type: 'button',
              id: 'execute',
              text: 'Execute (Alt+Enter)',
              tooltip: 'Alt+Enter executes selection or full query<br>Alt+Shift+Enter does the same and refreshes the sidebar schema',
              icon: 'fa fa-play',
              onClick: async function() {
                await executeQuery()
              },
            },
          ],
        },
        html: '<div style="padding: 5px; height:100%;"><textarea id="sql-explorer-query" class="w2ui-input" style="height: 100%; width: 100%; resize:none; font-family: monospace;"></textarea></div>',
      },
      {
        type: 'main',
        html: grid,
      }
    ],
  })

  return new w2layout({
    name: 'sqlExplorerLayout',
    panels: [
      {
        type: 'left',
        size: 200,
        html: sidebar,
      },
      {
        type: 'main',
        html: layout,
      },
    ],
    onRender: async function(event) {
      await event.complete
      await fetchSchema()
      const textarea = document.getElementById('sql-explorer-query')
      textarea.addEventListener('keydown', async e => {
        if (e.key === 'Tab' && !e.shiftKey) {
          e.preventDefault()
          const start = textarea.selectionStart
          const end = textarea.selectionEnd
          textarea.value = textarea.value.substring(0, start) + '    ' + textarea.value.substring(end)
          textarea.selectionStart = textarea.selectionEnd = start + 4
        } else if (e.key === 'Escape') {
          abortController?.abort()
        } else if (e.key === 'Enter' && e.altKey) {
          e.preventDefault()
          await executeQuery()
          if (e.shiftKey) {
            await fetchSchema()
          }
        }
      })
    },
    onDestroy: function() {
      abortController?.abort()
      layout.destroy()
      grid.destroy()
      sidebar.destroy()
    }
  })
}

