<template>
  <div class="code-editor-wrapper">
    <div class="line-numbers" ref="lineNumbers">
      <span v-for="n in lineCount" :key="n">{{ n }}</span>
    </div>
    <textarea
      ref="textarea"
      class="code-textarea"
      :value="modelValue"
      :readonly="readOnly"
      :style="{ height: height }"
      spellcheck="false"
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      @input="onInput"
      @scroll="syncScroll"
      @keydown="onKeydown"
    ></textarea>
  </div>
</template>

<script>
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
  computed: {
    lineCount() {
      if (!this.modelValue) return 1
      return this.modelValue.split('\n').length
    }
  },
  methods: {
    onInput(e) {
      this.$emit('update:modelValue', e.target.value)
    },
    syncScroll() {
      if (this.$refs.lineNumbers && this.$refs.textarea) {
        this.$refs.lineNumbers.scrollTop = this.$refs.textarea.scrollTop
      }
    },
    onKeydown(e) {
      // Tab 键插入两个空格
      if (e.key === 'Tab') {
        e.preventDefault()
        const textarea = e.target
        const start = textarea.selectionStart
        const end = textarea.selectionEnd
        const value = textarea.value
        textarea.value = value.substring(0, start) + '  ' + value.substring(end)
        textarea.selectionStart = textarea.selectionEnd = start + 2
        this.$emit('update:modelValue', textarea.value)
      }
    }
  }
}
</script>

<style scoped>
.code-editor-wrapper {
  display: flex;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  background: #1e1e1e;
}

.line-numbers {
  padding: 10px 0;
  background: #2d2d2d;
  color: #858585;
  font-family: 'Menlo', 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  text-align: right;
  user-select: none;
  overflow: hidden;
  min-width: 40px;
}

.line-numbers span {
  display: block;
  padding: 0 8px;
}

.code-textarea {
  flex: 1;
  padding: 10px 12px;
  border: none;
  outline: none;
  resize: none;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Menlo', 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  tab-size: 2;
  white-space: pre;
  overflow: auto;
}

.code-textarea::placeholder {
  color: #5a5a5a;
}

.code-textarea:focus {
  outline: none;
}

.code-editor-wrapper:focus-within {
  border-color: #409eff;
}
</style>
