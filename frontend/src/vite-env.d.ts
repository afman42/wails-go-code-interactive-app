/// <reference types="svelte" />
/// <reference types="vite/client" />

import { main } from '../wailsjs/go/models'

type ResultData = main.ResponseData & {
  meta: {
    status_code: number
    message: string
  }
  data: {
    txt: string
    out: string
    errout: string
    lang: string
    type: string
  }
}
