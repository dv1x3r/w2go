import { w2form, w2grid, w2layout } from './w2ui.es6.min.js'
import { w2fetch } from './w2ui.helpers.js'

export function createSqlExplorerLayout(opts = {}) {
  const grid = createSqlExplorerGrid()
  const form = createSqlExplorerForm(opts, grid)
  return new w2layout({
    name: 'sqlExplorerLayout',
    padding: 4,
    panels: [
      {
        type: 'left',
        style: 'border: 1px solid #efefef; padding: 5px',
        html: form,
        size: 400,
        resizable: true,
      },
      {
        type: 'main',
        style: 'border: 1px solid #efefef; padding: 5px',
        html: grid,
      },
    ],
    onDestroy: function() {
      form.destroy()
      grid.destroy()
    }
  })
}

async function executeQuery(url, query, grid, signal) {
  if (!query) {
    return
  }
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
}

function createSqlExplorerForm(opts, grid) {
  const { execute, header = '' } = opts
  let abortController = null
  return new w2form({
    name: `sqlExplorerForm`,
    header: header,
    toolbar: {
      items: [
        {
          type: 'button',
          id: 'execute',
          text: 'Execute (Alt+Enter)',
          icon: 'fa fa-play',
          onClick: async function(event) {
            abortController = new AbortController()
            const record = event.owner.owner.record
            await executeQuery(execute, record.query, grid, abortController.signal)
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
      ],
    },
    onRender: async function(event) {
      await event.complete
      event.owner.fields.forEach(x => {
        x.$el.on('keydown', async e => {
          if (e.altKey && e.keyCode === 13) {
            e.preventDefault()
            abortController = new AbortController()
            const record = event.owner.record
            await executeQuery(execute, record.query, grid, abortController.signal)
          }
        })
      })
    },
    fields: [
      {
        field: 'query',
        type: 'textarea',
        html: {
          span: -1,
          label: 'Query',
          attr: 'style="width: 100%; height: 300px; resize: none;"'
        }
      },
    ],
  })
}

function createSqlExplorerGrid() {
  return new w2grid({
    name: 'sqlExplorerGrid',
    show: {
      footer: true,
      toolbar: false,
      lineNumbers: true,
    },
  })
}

