<template>
  <div ref="editorContainer" :style="{ height: height, width: '100%', border: '1px solid #dcdfe6', borderRadius: '4px' }"></div>
</template>

<script>
import * as monaco from 'monaco-editor'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

self.MonacoEnvironment = {
  getWorker() {
    return new editorWorker()
  }
}

export default {
  name: 'MonacoEditor',
  props: {
    modelValue: {
      type: String,
      default: ''
    },
    language: {
      type: String,
      default: 'shell'
    },
    height: {
      type: String,
      default: '200px'
    },
    readOnly: {
      type: Boolean,
      default: false
    }
  },
  emits: ['update:modelValue'],
  data() {
    return {
      editor: null
    }
  },
  watch: {
    modelValue(newVal) {
      if (this.editor && this.editor.getValue() !== newVal) {
        this.editor.setValue(newVal || '')
      }
    },
    language(newVal) {
      if (this.editor) {
        const model = this.editor.getModel()
        if (model) {
          monaco.editor.setModelLanguage(model, newVal)
        }
      }
    },
    readOnly(newVal) {
      if (this.editor) {
        this.editor.updateOptions({ readOnly: newVal })
      }
    }
  },
  mounted() {
    this.initEditor()
  },
  beforeUnmount() {
    if (this.editor) {
      this.editor.dispose()
    }
  },
  methods: {
    initEditor() {
      this.editor = monaco.editor.create(this.$refs.editorContainer, {
        value: this.modelValue || '',
        language: this.language,
        theme: 'vs',
        minimap: { enabled: false },
        lineNumbers: 'on',
        wordWrap: 'on',
        fontSize: 14,
        scrollBeyondLastLine: false,
        automaticLayout: true,
        tabSize: 2,
        readOnly: this.readOnly,
        scrollbar: {
          verticalScrollbarSize: 8,
          horizontalScrollbarSize: 8
        },
        overviewRulerLanes: 0,
        hideCursorInOverviewRuler: true,
        overviewRulerBorder: false,
        renderLineHighlight: 'line'
      })

      this.editor.onDidChangeModelContent(() => {
        this.$emit('update:modelValue', this.editor.getValue())
      })
    }
  }
}
</script>
