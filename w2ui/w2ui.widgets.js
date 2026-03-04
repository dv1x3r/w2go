import { w2grid, w2layout, w2sidebar } from './w2ui.es6.min.js'
import { w2fetch, registerSidebarSearch } from './w2ui.helpers.js'

export function createSqlExplorerLayout(opts = {}) {
  const { execute, schema } = opts

  const sidebar = new w2sidebar({
    name: 'sqlExplorerSidebar',
    levelPadding: 8,
    topHTML: '<div style="margin-top:2px;padding:3px 5px;height:36px;"><input id="sqlExplorerSearch" class="w2ui-input" style="width:100%;" placeholder="Search..."></div>',
    onRender: async function(event) {
      await event.complete
      const search = registerSidebarSearch(sidebar)
      const el = document.getElementById('sqlExplorerSearch')
      el.addEventListener('keyup', e => search(e.target.value))
    }
  })

  async function fetchSchema() {
    const result = await w2fetch({ url: schema, method: 'GET' })
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
              text: `${col.name} (${col.type})`,
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

    const toolbar = layout.get('main').toolbar
    const button = toolbar.items[0]
    button.text = 'Cancel (Escape)'
    button.icon = 'fa fa-stop'
    toolbar.refresh()

    isRunning = true
    abortController = new AbortController()

    const textarea = document.getElementById('sqlExplorerQuery')
    const selection = textarea.value.substring(textarea.selectionStart, textarea.selectionEnd)
    const query = selection || textarea.value
    const startTime = performance.now()

    try {
      const result = await w2fetch({
        owner: grid,
        lock: 'Executing...',
        url: execute,
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
    name: 'sqlExplorerLayout',
    panels: [
      {
        type: 'left',
        size: 200,
        resizable: true,
        html: sidebar,
      },
      {
        type: 'main',
        style: 'border-left: 1px solid #efefef;',
        toolbar: {
          items: [
            {
              type: 'button',
              id: 'execute',
              text: 'Execute (Alt+Enter)',
              tooltip: 'Alt+Enter executes selection or full query',
              icon: 'fa fa-play',
              onClick: async function() {
                await executeQuery()
              },
            },
          ],
        },
        html: '<div style="padding: 5px; height:100%;"><textarea id="sqlExplorerQuery" class="w2ui-input" style="height: 100%; width: 100%; resize:none;"></textarea></div>',
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
      await fetchSchema()
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
          await executeQuery()
        } else if (e.key === 'Escape') {
          abortController?.abort()
        }
      })
    },
    onDestroy: function() {
      abortController?.abort()
      sidebar.destroy()
      grid.destroy()
    }
  })

  return layout
}

