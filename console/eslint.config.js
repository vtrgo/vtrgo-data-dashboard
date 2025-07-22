import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  // Global ignores. The trailing slash ensures it's treated as a directory.
  { ignores: ['dist/'] },

  // Apply base recommended rules to all files.
  js.configs.recommended,
  ...tseslint.configs.recommended,

  // Configuration specific to React files.
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      // Use the modern recommended rules for React Hooks.
      ...reactHooks.configs['recommended-latest'].rules,

      // Rule for React Refresh.
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],

      // Override the default 'no-unused-vars' to allow underscore-prefixed variables.
      // This is a common convention for intentionally unused function arguments.
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrorsIgnorePattern: '^_',
        },
      ],
    },
  },
)
