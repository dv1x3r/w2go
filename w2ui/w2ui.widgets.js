import { w2grid, w2layout, w2sidebar } from './w2ui.es6.min.js'
import { w2fetch, registerSidebarSearch } from './w2ui.helpers.js'

async function executeQuery(url, query, grid, signal) {
  const startTime = performance.now()
  const result = await w2fetch({
    owner: grid,
    lock: 'Executing...',
    url: url,
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query }),
    signal: signal,
  })
  if (result) {
    grid.columns = result.columns.map(column => ({ field: column, text: column, render: 'text', min: 100 }))
    grid.records = result.records.map((row, i) => ({ recid: i + 1, ...row }))
    grid.total = result.total
    grid.refresh()
  }
  const elapsed = ((performance.now() - startTime) / 1000).toFixed(3)
  grid.status(`Query Executed ${elapsed} seconds`)
}

async function fetchSchema(url) {
  const result = await w2fetch({ url: url, method: 'GET' })
  return result.databases.map((db, dbIndex) => ({
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
          expanded: true,
          nodes: table.columns.map((col, colIndex) => ({
            id: `db-${dbIndex}-table-${tableIndex}-col-${colIndex}`,
            text: `${col.name} (${col.type})`,
            icon: col.pk ? 'fa fa-key' : 'fa',
          })),
        })),
      },
    ],
  }))
}

export function createSqlExplorerLayout(opts = {}) {
  const { execute, schema } = opts
  const sidebar = new w2sidebar({
    name: 'sqlExplorerSidebar',
    levelPadding: 8,
    topHTML: '<div style="height:36px;padding:3px 5px;"><input id="sqlExplorerSearch" style="width:100%;" class="w2ui-input" placeholder="Search..."></div>',
    onRender: async function(event) {
      await event.complete
      const search = registerSidebarSearch(sidebar)
      const el = document.getElementById('sqlExplorerSearch')
      el.addEventListener('keyup', e => search(e.target.value))
    }
  })
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
  return new w2layout({
    name: 'sqlExplorerLayout',
    padding: 4,
    panels: [
      {
        type: 'left',
        size: 200,
        resizable: true,
        html: sidebar,
      },
      {
        type: 'main',
        toolbar: {
          height: '20px',
          items: [
            {
              type: 'button',
              id: 'execute',
              text: 'Execute (Alt+Enter)',
              tooltip: 'Alt+Enter executes selection or full query',
              icon: 'fa fa-play',
              onClick: async function() {
                if (isRunning) {
                  return
                }
                isRunning = true
                abortController = new AbortController()
                const textarea = document.getElementById('sqlExplorerQuery')
                try {
                  await executeQuery(execute, textarea.value, grid, abortController.signal)
                } finally {
                  isRunning = false
                  abortController = null
                }
              },
            },
            {
              type: 'button',
              id: 'cancel',
              text: 'Cancel',
              icon: 'fa fa-stop',
              onClick: function() {
                abortController?.abort()
              },
            },
            {
              type: 'button',
              id: 'refresh',
              text: 'Schema',
              icon: 'fa fa-refresh',
              onClick: async function() {
                document.getElementById('sqlExplorerSearch').value = ''
                sidebar.nodes = await fetchSchema(schema)
                sidebar.refresh()
              },
            },
          ],
        },
        html: '<div style="padding:5px;width:100%;height:100%;"><textarea id="sqlExplorerQuery" class="w2ui-input" style="width:100%;height:100%;resize:none;"></textarea></div>',
      },
      {
        type: 'right',
        size: -350,
        resizable: true,
        html: grid,
      }
    ],
    onRender: async function(event) {
      await event.complete
      sidebar.nodes = await fetchSchema(schema)
      sidebar.refresh()
      const textarea = document.getElementById('sqlExplorerQuery')
      textarea.addEventListener('keydown', async e => {
        if (e.key === 'Tab') {
          e.preventDefault()
          const start = textarea.selectionStart
          const end = textarea.selectionEnd
          textarea.value = textarea.value.substring(0, start) + '    ' + textarea.value.substring(end)
          textarea.selectionStart = textarea.selectionEnd = start + 4
        } else if (e.key === 'Enter' && e.altKey) {
          e.preventDefault()
          if (isRunning) {
            return
          }
          isRunning = true
          abortController = new AbortController()
          const selection = textarea.value.substring(textarea.selectionStart, textarea.selectionEnd)
          const query = selection || textarea.value
          try {
            await executeQuery(execute, query, grid, abortController.signal)
          } finally {
            isRunning = false
            abortController = null
          }
        }
      })
    },
    onDestroy: function() {
      sidebar.destroy()
      grid.destroy()
    }
  })
}

