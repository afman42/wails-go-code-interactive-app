<script lang="ts">
  import { CheckFileExecutable } from '../wailsjs/go/main/App.js'
  import { RunFileExecutable } from '../wailsjs/go/main/App.js'
  import { main } from '../wailsjs/go/models.js'
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
  const toast = useToast()
  // Language configuration map
  const languageConfigs = {
    node: javascript(),
    php: php(),
    go: go()
  } as Record<string, LanguageSupport>
  // Reactive states
  let view: EditorView | null = null
  let editorContainer: HTMLDivElement
  let currentLang = $state(languageConfigs.node) as LanguageSupport
  let editorValue = $state('')
  let prevLang = $state(langState.value)
  let prevType = $state(langState.type)
  let allLang: string[] = $state([])
  const init = async (arrayFile: string[]) => await CheckFileExecutable(arrayFile)
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
  async function checkAllFileExecutable(arrayFile: string[]) {
    await init(arrayFile).then((result: string[]) => result.forEach((v) => allLang.push(v)))
  }
  // Initialize editor on mount
  onMount(async () => {
    await checkAllFileExecutable(['node', 'php', 'go'])
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
    if (allLang.length == 0) {
      toast.warning('Something Went Wrong', 4000)
      return
    }
    toast.info('Waiting Response', 3000)
    disabled = true
    let data = new main.Data()
    data.txt = langState.sampleDataLang[langState.type][langState.value]
    data.lang = langState.value
    data.type = langState.type

    try {
      const result = await RunFileExecutable(data)
      if (result.meta == undefined || result.data == undefined) {
        disabled = false
        toast.warning('Something Went Wrong', 3000)
        return
      }
      if (result.meta?.status_code == 200) {
        disabled = false
        stderr =
          result.data?.errout.trim().length > 0
            ? result.data?.errout
            : 'Nothing'
        stdout =
          result.data?.out.trim().length > 0 && stderr == 'Nothing'
            ? result.data?.out
            : 'Nothing'
        if (langState.type == 'stq') {
          stderr =
            result.data?.errout.trim().length > 0
              ? result.data?.errout
              : 'Nothing'
          stdout =
            result.data?.out.trim().length > 0
              ? JSON.parse(result.data?.out.trim())
              : 'Nothing'
        }
        if (stderr != 'Nothing') {
          toast.warning('Something Went Wrong', 1000)
        }
        if (stdout != 'Nothing') {
          toast.success('Success Response', 1000)
        }
      }
      if (result.meta?.status_code == 400) {
        toast.warning(result.meta?.message,3000)
        disabled = false
      }
    } catch (error: any) {
      if (error.meta.status_code == 400)
        toast.error(error.meta.message, 1000)
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
    <div class="flex w-full border-black border-2 border-solid mb-2 h-80">
      <div bind:this={editorContainer} class="w-full h-full"></div>
    </div>
    <div class="flex items-center">
      <button
        class="bg-red-500 flex py-2.5 px-3 text-white rounded-lg mr-1"
        {disabled}
        onclick={send}
        type="button">Send</button
      >
      <div class="flex gap-1">
        <div class="flex gap-2 items-center">
          {#each allLang as lang}
            <label for={lang} class="gap-2">
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
            <b>--- Proses Code ---</b>
          {/if}
        </div>
      </div>
    </div>
  </div>

  <div class="flex flex-col ml-4">
    <div class="flex mt-3 flex-col w-full">
      {#if allLang.length == 0}
        <h3>Please Add Golang,Node JS,PHP Installation</h3>
        <button
          class="bg-blue-500 text-white w-full h-20"
          type="button"
          onclick={async (e) => {
            e.preventDefault()
            await checkAllFileExecutable(['node', 'php', 'go'])
            toast.info(
              `Installation executable ${allLang.length == 0 ? `none` : allLang.join(' , ')}`,
              1000
            )
          }}>Restart</button
        >
      {:else}
        {#if langState.type == 'repl'}
          <h6 class="">StdOut</h6>
          <blockquote class="border-l-4 border-gray-500 my-2 py-4 pl-4">
            {stdout}
          </blockquote>
        {/if}
        {#if langState.type == 'stq'}
          <h6 class="">Simple Test Question : change integer to string</h6>
          <h6 class="">Result</h6>
          <blockquote
            class="flex gap-1 flex-start border-l-4 border-gray-500 my-2 py-4 pl-4"
          >
            <input type="checkbox" value="stq1" checked={!stdout} disabled /> Check
            change after int to string
          </blockquote>
        {/if}
        <h6 class="min-sm:text-sm md:text-sm">StdErr</h6>
        <blockquote class="border-l-4 border-gray-500 my-2 py-4 pl-4">
          {stderr}
        </blockquote>
      {/if}
    </div>
  </div>
</div>

<ToasterContainer />
