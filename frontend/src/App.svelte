<script lang="ts">
  import {
    ListBundledRuntimes,
    ListLanguageAvailability,
    RunFileExecutable
  } from '../wailsjs/go/main/App.js'
  import { main } from '../wailsjs/go/models'
  //Component
  import ToasterContainer from './component/ToasterContainer.svelte'
  import { onMount, onDestroy } from 'svelte'
  import { EditorState, type Extension } from '@codemirror/state'
  import {
    EditorView,
    keymap,
    highlightSpecialChars,
    drawSelection,
    highlightActiveLine,
    dropCursor,
    rectangularSelection,
    crosshairCursor,
    lineNumbers,
    highlightActiveLineGutter
  } from '@codemirror/view'
  import {
    defaultHighlightStyle,
    syntaxHighlighting,
    indentOnInput,
    bracketMatching,
    foldGutter,
    foldKeymap,
    type LanguageSupport
  } from '@codemirror/language'
  import {
    defaultKeymap,
    history,
    historyKeymap,
    indentWithTab
  } from '@codemirror/commands'
  import { searchKeymap, highlightSelectionMatches } from '@codemirror/search'
  import {
    autocompletion,
    completionKeymap,
    closeBrackets,
    closeBracketsKeymap,
    snippetCompletion,
    type CompletionContext,
    type CompletionResult
  } from '@codemirror/autocomplete'
  import { lintKeymap } from '@codemirror/lint'
  import { javascript } from '@codemirror/lang-javascript'
  import { go } from '@codemirror/lang-go'
  import { php } from '@codemirror/lang-php'
  // lib state
  import { langState } from './utils/lang-state.svelte'
  import { useToast } from './utils/toast.svelte'
  let stdout: string = $state('Nothing')
  let stderr: string = $state('Nothing')
  let disabled: boolean = $state(false)
  let execMode: string = $state('default')
  let preferBundled: boolean = $state(false)
  let bundledRuntimes: string[] = $state([])
  let selectedRuntime: string = $state('')
  let customExecutable: string = $state('')
  let customWorkingDir: string = $state('')
  let systemLangs: string[] = $state([])
  let bundledLangs: string[] = $state([])
  let displayedLangs: string[] = $state([])
  const toast = useToast()
  // Language configuration map
  const languageConfigs = {
    node: javascript(),
    php: php(),
    go: go()
  } as Record<string, LanguageSupport>
  const supportedLanguages = ['node', 'php', 'go']
  // Reactive states
  let view: EditorView | null = null
  let editorContainer: HTMLDivElement
  let currentLang = $state(languageConfigs.node) as LanguageSupport
  let editorValue = $state('')
  let prevLang = $state(langState.value)
  let prevType = $state(langState.type)
  const loadStoredLangs = () => {
    const raw = window.localStorage.getItem('allLang')
    if (!raw) return []
    try {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) {
        return parsed.filter(
          (value): value is string => typeof value === 'string'
        )
      }
    } catch {
      return []
    }
    return []
  }

  const toUniqueList = (values: string[]) => {
    return [...new Set(values.map((value) => (typeof value === 'string' ? value.trim() : '')).filter((value) => value.length > 0))]
  }

  let arrLang: string[] = $state(loadStoredLangs())

  const persistLangs = (values: string[]) => {
    const unique = toUniqueList(values)
    arrLang = [...unique]
    window.localStorage.setItem('allLang', JSON.stringify(arrLang))
  }

  const updateDisplayedLanguages = () => {
    let next: string[] = []
    if (execMode === 'bundled') {
      next = bundledLangs
    } else if (execMode === 'custom') {
      next = systemLangs
    } else {
      next = systemLangs.length > 0 ? systemLangs : bundledLangs
    }
    const unique = toUniqueList(next)
    displayedLangs = unique
    if (unique.length > 0 && !unique.includes(langState.value)) {
      langState.value = unique[0]
    }
  }

  const init = async (arrayFile: string[]) => {
    let system: string[] = []
    let bundled: string[] = []
    try {
      const availability = await ListLanguageAvailability(arrayFile)
      system = toUniqueList(availability?.system ?? [])
      bundled = toUniqueList(availability?.bundled ?? [])
    } catch {
      system = [...systemLangs]
      bundled = [...bundledLangs]
    }

    systemLangs = [...system]
    bundledLangs = [...bundled]

    const merged = toUniqueList([...system, ...bundled])
    persistLangs(merged)
    if (merged.length > 0 && !merged.includes(langState.value)) {
      langState.value = merged[0]
    }

    await refreshBundledRuntimes()
    updateDisplayedLanguages()
  }

  const refreshBundledRuntimes = async () => {
    try {
      const result = await ListBundledRuntimes()
      const filtered = result
        .map((value) => value?.trim())
        .filter(
          (value): value is string =>
            typeof value === 'string' && value.length > 0
        )
      const unique = [...new Set(filtered)]
      bundledRuntimes = unique
      if (unique.length === 0) {
        selectedRuntime = ''
      } else if (!unique.includes(selectedRuntime)) {
        selectedRuntime = unique[0]
      }
    } catch {
      bundledRuntimes = []
      selectedRuntime = ''
    }
  }

  function onChangeExecMode(event: Event) {
    const value = (event.target as HTMLInputElement).value
    execMode = value
    if (value === 'bundled') {
      preferBundled = false
      if (bundledLangs.length === 0) {
        void init(supportedLanguages)
      } else {
        void refreshBundledRuntimes()
      }
    } else if (value === 'custom') {
      preferBundled = false
    }
    updateDisplayedLanguages()
  }

  function onTogglePreferBundled(event: Event) {
    preferBundled = (event.target as HTMLInputElement).checked
    updateDisplayedLanguages()
  }

  function onSelectBundledRuntime(event: Event) {
    selectedRuntime = (event.target as HTMLSelectElement).value
  }

  function onInputCustomExecutable(event: Event) {
    customExecutable = (event.target as HTMLInputElement).value
  }

  function onInputCustomWorkingDir(event: Event) {
    customWorkingDir = (event.target as HTMLInputElement).value
  }
  // Custom autocompletion for PHP
  function phpCompletions(context: CompletionContext): CompletionResult | null {
    const word = context.matchBefore(/\w*/)
    if (!word || (word.from === word.to && !context.explicit)) return null

    const phpKeywords = [
      'echo',
      'else',
      'foreach',
      'function',
      'return',
      'class',
      'public',
      'private',
      'protected',
      'namespace',
      'use'
    ]
    const phpFunctions = [
      'array',
      'strlen',
      'str_replace',
      'explode',
      'implode',
      'isset'
    ]

    return {
      from: word.from,
      options: [
        ...phpKeywords.map((kw) => ({ label: kw, type: 'keyword' })),
        ...phpFunctions.map((fn) => ({ label: fn, type: 'function' })),
        snippetCompletion(
          'for (let ${i} = 0; ${i} < ${len}; ${i}++) {\n\t${}\n}',
          { label: 'for', detail: 'loop' }
        ) // Example Snippet Completion
      ]
    }
  }

  // Custom autocompletion for Go
  function goCompletions(context: CompletionContext): CompletionResult | null {
    const word = context.matchBefore(/\w*/)
    if (!word || (word.from === word.to && !context.explicit)) return null

    const goKeywords = [
      'func',
      'var',
      'const',
      'else',
      'for',
      'range',
      'return',
      'struct',
      'interface',
      'package',
      'import',
      'type'
    ]
    const goBuiltins = [
      'println',
      'print',
      'len',
      'cap',
      'make',
      'new',
      'append'
    ]

    return {
      from: word.from,
      options: [
        ...goKeywords.map((kw) => ({ label: kw, type: 'keyword' })),
        ...goBuiltins.map((fn) => ({ label: fn, type: 'function' })),
        snippetCompletion('if ${} {\n\t${}\n}', {
          label: 'if',
          detail: 'if ${i} block'
        }) // Example Snippet Completion
      ]
    }
  }

  // Language-specific completion extensions
  const completionExtensions = {
    node: autocompletion(), // Built-in for JavaScript
    php: autocompletion({ override: [phpCompletions] }),
    go: autocompletion({ override: [goCompletions] })
  } as Record<string, Extension>

  // Base extensions (shared across all configurations)
  const baseExtensions: Extension[] = [
    lineNumbers(),
    foldGutter(),
    highlightSpecialChars(),
    history(),
    drawSelection(),
    dropCursor(),
    EditorState.allowMultipleSelections.of(true),
    indentOnInput(),
    syntaxHighlighting(defaultHighlightStyle),
    bracketMatching(),
    closeBrackets(),
    autocompletion(),
    rectangularSelection(),
    crosshairCursor(),
    highlightActiveLine(),
    highlightActiveLineGutter(),
    highlightSelectionMatches(),
    keymap.of([
      indentWithTab,
      ...closeBracketsKeymap,
      ...defaultKeymap,
      ...searchKeymap,
      ...historyKeymap,
      ...foldKeymap,
      ...completionKeymap,
      ...lintKeymap
    ])
  ]
  // Initialize editor on mount
  onMount(async () => {
    await init(supportedLanguages)
    currentLang = languageConfigs[langState.value] || languageConfigs.node
    const initialState = EditorState.create({
      doc: langState.sampleDataLang[langState.type][langState.value] || '',
      extensions: [
        ...baseExtensions,
        currentLang,
        completionExtensions[langState.value]
      ]
    })

    view = new EditorView({
      state: initialState,
      parent: editorContainer,
      dispatch: (tr) => {
        view?.update([tr])
        if (tr.docChanged) {
          editorValue = view?.state.doc.toString() || ''
          langState.sampleDataLang[langState.type][langState.value] =
            editorValue
        }
      }
    })

    editorValue = view.state.doc.toString()
  })

  // Effect for language/type changes
  $effect(() => {
    if (langState.value !== prevLang || langState.type !== prevType) {
      currentLang = languageConfigs[langState.value] || languageConfigs.node
      editorValue =
        langState.sampleDataLang[langState.type][langState.value] || ''

      if (view) {
        const newState = EditorState.create({
          doc: editorValue,
          extensions: [
            ...baseExtensions,
            currentLang,
            completionExtensions[langState.value]
          ]
        })
        view.setState(newState)
      }

      prevLang = langState.value
      prevType = langState.type
    }
  })
  async function send() {
    if (displayedLangs.length == 0) {
      toast.warning('Something Went Wrong', 4000)
      return
    }
    toast.info('Waiting Response', 3000)
    disabled = true
    let data = new main.Data()
    data.txt = langState.sampleDataLang[langState.type][langState.value]
    data.lang = langState.value
    data.type = langState.type

    const normalizedMode = execMode.trim().toLowerCase()
    const trimmedRuntime = selectedRuntime.trim()
    const trimmedExecutable = customExecutable.trim()
    const trimmedWorkingDir = customWorkingDir.trim()

    if (normalizedMode === 'bundled' && trimmedRuntime.length === 0) {
      disabled = false
      toast.warning('Select a bundled runtime before running the code', 4000)
      return
    }

    if (normalizedMode === 'custom' && trimmedExecutable.length === 0) {
      disabled = false
      toast.warning('Provide the custom executable path before running the code', 4000)
      return
    }

    if (normalizedMode === 'bundled') {
      data.execMode = 'bundled'
      data.bundledRuntime = trimmedRuntime
      data.preferBundled = false
      data.customExecutable = ''
      data.customWorkingDir = ''
    } else if (normalizedMode === 'custom') {
      data.execMode = 'custom'
      data.customExecutable = trimmedExecutable
      data.customWorkingDir = trimmedWorkingDir
      data.preferBundled = false
      data.bundledRuntime = ''
    } else {
      data.execMode = 'default'
      data.preferBundled = preferBundled
      data.bundledRuntime = trimmedRuntime
      data.customExecutable = ''
      data.customWorkingDir = ''
    }

    try {
      const result = await RunFileExecutable(data)

      if (result == undefined) {
        disabled = false
        toast.warning('Something Went Wrong', 3000)
        return
      }
      disabled = false
      stderr = result.errout.trim().length > 0 ? result.errout : 'Nothing'
      stdout =
        result.out.trim().length > 0 && stderr == 'Nothing'
          ? result.out
          : 'Nothing'
      if (langState.type == 'stq') {
        stderr = result.errout.trim().length > 0 ? result.errout : 'Nothing'
        stdout =
          result.out.trim().length > 0
            ? JSON.parse(result.out.trim())
            : 'Nothing'
      }
      if (stderr != 'Nothing') {
        toast.warning('Something Went Wrong', 1000)
      }
      if (stdout != 'Nothing') {
        toast.success('Success Response', 1000)
      }
    } catch (error: any) {
      toast.error(error, 5000)
      disabled = false
      stdout = 'Nothing'
      stderr = 'Nothing'
    }
  }
  function onChangeRadio(event: Event) {
    langState.value = (event.target as HTMLInputElement).value
    stdout = 'Nothing'
    stderr = 'Nothing'
  }
  function onChangeType(event: Event) {
    langState.type = (event.target as HTMLInputElement).value
    stdout = 'Nothing'
    stderr = 'Nothing'
  }

  // Clean up
  onDestroy(() => {
    view?.destroy()
    view = null
  })
</script>

<div class="flex flex-col">
  <div class="flex flex-col m-4 h-full">
    <div
      class="flex w-full border-black border-2 border-solid mb-2 h-80 overflow-y-scroll overflow-x-hidden"
    >
      <div bind:this={editorContainer} class="w-full h-ful"></div>
    </div>
    <div class="flex items-center">
      <button
        class="bg-red-500 flex py-2.5 px-3 text-white rounded-lg mr-1"
        {disabled}
        onclick={send}
        type="button">Send</button
      >
      <div class="flex gap-1 w-full">
        <div class="flex gap-2 items-center w-full">
          {#each displayedLangs as lang}
            <label for={lang} class="flex items-center">
              <input
                type="radio"
                value={lang}
                class="mr-0.5"
                onchange={onChangeRadio}
                checked={langState.value == lang}
              />{lang}
            </label>
          {:else}
            <label for="empty">Empty</label>
          {/each}
          <b> || </b>
          <label for="repl" class="">
            <input
              type="radio"
              value="repl"
              class="mr-0.5"
              onchange={onChangeType}
              checked={langState.type == 'repl'}
            />REPL
          </label>
          <label for="stq" class="">
            <input
              type="radio"
              value="stq"
              class="mr-0.5"
              onchange={onChangeType}
              checked={langState.type == 'stq'}
            />Simple Test Question
          </label>
          {#if disabled}
            <b>--- Process Code ---</b>
          {/if}
        </div>
        <div class="flex flex-col gap-2 w-full mt-2">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-semibold">Execution mode:</span>
            <label class="flex items-center gap-1">
              <input
                type="radio"
                value="default"
                class="mr-0.5"
                onchange={onChangeExecMode}
                checked={execMode == 'default'}
              />Default
            </label>
            <label class="flex items-center gap-1">
              <input
                type="radio"
                value="bundled"
                class="mr-0.5"
                onchange={onChangeExecMode}
                checked={execMode == 'bundled'}
              />Bundled
            </label>
            <label class="flex items-center gap-1">
              <input
                type="radio"
                value="custom"
                class="mr-0.5"
                onchange={onChangeExecMode}
                checked={execMode == 'custom'}
              />Custom
            </label>
            {#if execMode == 'default'}
              <label class="flex items-center gap-1">
                <input
                  type="checkbox"
                  onchange={onTogglePreferBundled}
                  checked={preferBundled}
                />
                Prefer bundled fallback
              </label>
            {/if}
          </div>
          {#if execMode == 'bundled'}
            <div class="flex flex-wrap items-center gap-2">
              <label class="font-medium" for="bundled-runtime">Runtime:</label>
              <select
                class="border rounded px-2 py-1"
                onchange={onSelectBundledRuntime}
                id="bundled-runtime"
              >
                <option value="" selected={selectedRuntime == ''} disabled>
                  Select runtime
                </option>
                {#each bundledRuntimes as runtime}
                  <option value={runtime} selected={runtime == selectedRuntime}>
                    {runtime}
                  </option>
                {/each}
              </select>
              <button
                class="bg-blue-500 text-white px-2 py-1 rounded"
                type="button"
                onclick={async (e) => {
                  e.preventDefault()
                  await init(supportedLanguages)
                  toast.info('Bundled runtimes refreshed', 1500)
                }}
              >
                Refresh
              </button>
            </div>
            {#if bundledRuntimes.length == 0}
              <p class="text-sm text-red-600">
                No bundled runtimes detected for this build.
              </p>
            {/if}
          {/if}
          {#if execMode == 'custom'}
            <div class="flex flex-col gap-2">
              <label class="flex flex-col">
                <span class="text-sm font-medium">Executable</span>
                <input
                  type="text"
                  class="border rounded px-2 py-1"
                  placeholder="Absolute or relative executable path"
                  value={customExecutable}
                  oninput={onInputCustomExecutable}
                />
              </label>
              <label class="flex flex-col">
                <span class="text-sm font-medium">Working directory (optional)</span>
                <input
                  type="text"
                  class="border rounded px-2 py-1"
                  placeholder="Working directory"
                  value={customWorkingDir}
                  oninput={onInputCustomWorkingDir}
                />
              </label>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>

  <div class="flex flex-col mx-4">
    <div class="flex mt-3 flex-col w-full">
      {#if arrLang.length == 0}
        <h3>Please Add Golang,Node JS,PHP Installation</h3>
        <button
          class="bg-blue-500 text-white w-full h-20"
          type="button"
          onclick={async (e) => {
            e.preventDefault()
            await init(supportedLanguages)
            toast.info(
              `Installation executable ${arrLang.length == 0 ? `none` : arrLang.join(' , ')}`,
              1000
            )
          }}>Restart</button
        >
      {:else}
        {#if langState.type == 'repl'}
          <h6>StdOut</h6>
          <blockquote class="border-l-4 border-gray-500 my-2 py-4 pl-4">
            {stdout}
          </blockquote>
        {/if}
        {#if langState.type == 'stq'}
          <h6>Simple Test Question : change integer to string</h6>
          <h6>Result</h6>
          <blockquote
            class="flex gap-1 flex-start border-l-4 border-gray-500 my-2 py-4 pl-4 items-center"
          >
            <input type="checkbox" value="stq1" checked={!stdout} disabled /> Check
            change after int to string
          </blockquote>
        {/if}
        <h6>StdErr</h6>
        <blockquote class="border-l-4 border-gray-500 my-2 py-4 pl-4">
          {stderr}
        </blockquote>
      {/if}
    </div>
  </div>
</div>

<ToasterContainer />
