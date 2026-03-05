import { w2grid, w2layout, w2sidebar } from './w2ui.es6.min.js'
import { w2fetch, registerSidebarSearch } from './w2ui.helpers.js'

export function createSqlExplorerLayout(opts = {}) {
  const { url } = opts

  let abortController = null
  let isRunning = false
  let editor = null

  const grid = new w2grid({
    name: 'sqlExplorerGrid',
    show: {
      footer: true,
      toolbar: false,
      lineNumbers: true,
    },
  })

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

  async function executeQuery() {
    if (isRunning) {
      return
    }

    isRunning = true
    abortController = new AbortController()

    const toolbar = editorLayout.get('top').toolbar
    toolbar.disable('run')
    toolbar.enable('cancel')

    const query = editor.getSelection() || editor.getValue()
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
        grid.columns = result.columns.map(column => ({ field: column, text: column, render: 'nullable', sortable: true, editable: {} }))
        grid.records = result.records.map((row, i) => ({ recid: i + 1, ...row }))
        grid.total = result.total
        grid.sortData = []
        grid.refresh()
        grid.columnAutoSize()
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
    name: 'sqlEditorLayout',
    panels: [
      {
        type: 'top',
        size: 200,
        resizable: true,
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
              type: 'check',
              id: 'vim',
              text: 'Vim Mode',
              checked: false,
              onClick: async function(event) {
                await event.complete
                const checked = event.detail.item.checked
                editor.setOption('keyMap', checked ? 'vim' : 'default')
              },
            },
          ],
        },
        html: `<style>.CodeMirror-hints{ z-index: 9999 !important; }</style><div id="sql-explorer-editor" style="height:100%;"></div>`,
      },
      {
        type: 'main',
        html: grid,
      }
    ],
    onRender: async function(event) {
      await event.complete
      CodeMirror.Vim.defineEx('write', 'w', async () => {
        await executeQuery()
      })
      editor = CodeMirror(document.getElementById('sql-explorer-editor'), {
        lineNumbers: true,
        mode: 'text/x-sql',
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
            const schema = await w2fetch({ url: url, method: 'GET' })
            setSchemaSidebar(schema)
            setSchemaAutocomplete(schema)
          },
          'Shift-Esc': () => {
            abortController?.abort('The query has been cancelled')
          },
        },
      })
      editor.setSize('100%', '100%')
    }
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
        html: editorLayout,
      },
    ],
    onRender: async function(event) {
      await event.complete
      const schema = await w2fetch({ url: url, method: 'GET' })
      setSchemaSidebar(schema)
      setSchemaAutocomplete(schema)
    },
    onDestroy: function() {
      abortController?.abort()
      editorLayout.destroy()
      sidebar.destroy()
      grid.destroy()
    }
  })
}

