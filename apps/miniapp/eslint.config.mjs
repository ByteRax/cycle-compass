import uni from '@uni-helper/eslint-config'
import prettier from 'eslint-config-prettier'

export default uni(
  {
    vue: true,
    markdown: false,
    ignores: [
      'node_modules',
      '.nuxt',
      'dist',
      'unpackage',
      '**/uni_modules/',
      '**/nativeplugins/',
      'auto-import.d.ts',
      '**/uni-pages.d.ts',
      '**/async-import.d.ts',
      '**/async-component.d.ts',
      'src/pages.json',
      'src/manifest.json',
      'src/service/**',
      'index.html'
    ]
  },
  {
    files: ['src/**/*.vue', 'src/**/*.{ts,js}'],
    rules: {
      '@typescript-eslint/ban-ts-comment': ['error', { 'ts-ignore': 'allow-with-description' }],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/explicit-module-boundary-types': 'off',
      '@typescript-eslint/no-empty-function': ['warn', { allow: ['arrowFunctions', 'functions', 'methods'] }],
      '@typescript-eslint/no-non-null-asserted-optional-chain': 'error',
      '@typescript-eslint/no-non-null-assertion': 'warn',
      '@typescript-eslint/no-redeclare': 'error',
      '@typescript-eslint/no-unused-vars': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      'dot-notation': 'off',

      'vue/block-order': [
        'error',
        {
          order: ['script', 'template', 'style']
        }
      ],
      'vue/component-api-style': ['error', ['script-setup']],
      'vue/component-name-in-template-casing': ['error', 'PascalCase'],
      'vue/custom-event-name-casing': ['error', 'camelCase'],
      'vue/define-emits-declaration': ['error', 'type-based'],
      'vue/define-props-declaration': ['error', 'type-based'],
      'vue/multi-word-component-names': 'off',
      'vue/no-v-html': 'warn',

      'no-console': 'off',
      'no-debugger': 'error',
      'prefer-const': 'error'
    }
  },
  prettier
)
