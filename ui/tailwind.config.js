/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: {
          DEFAULT: '#0c0c0d',
          surface: '#141416',
          raised: '#1c1c1f',
          border: '#2b2b2f',
          subtle: '#222225',
        },
        cream: {
          DEFAULT: '#edeae2',
          dim: '#9a9299',
          faint: '#57535e',
        },
        gold: {
          DEFAULT: '#c4913a',
          light: '#d4a44e',
          glow: 'rgba(196,145,58,0.1)',
        },
        jade: '#4daa74',
        ember: '#d95b5b',
      },
      fontFamily: {
        serif: ['JetBrains Mono', 'Menlo', 'Consolas', 'monospace'],
        sans: ['JetBrains Mono', 'Menlo', 'Consolas', 'monospace'],
        mono: ['JetBrains Mono', 'Menlo', 'Consolas', 'monospace'],
      },
      spacing: { '13': '3.25rem' },
      animation: {
        'fade-up': 'fade-up 0.35s ease-out both',
        'fade-in': 'fade-in 0.2s ease-out both',
        'slide-up': 'slide-up 0.25s cubic-bezier(0.16, 1, 0.3, 1) both',
        'pulse-dot': 'pulse-dot 2.4s ease-in-out infinite',
        spin: 'spin 0.8s linear infinite',
      },
      keyframes: {
        'fade-up': {
          from: { opacity: '0', transform: 'translateY(10px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        'fade-in': {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        'slide-up': {
          from: { opacity: '0', transform: 'translateY(20px) scale(0.97)' },
          to: { opacity: '1', transform: 'translateY(0) scale(1)' },
        },
        'pulse-dot': {
          '0%, 100%': { opacity: '1', transform: 'scale(1)' },
          '50%': { opacity: '0.35', transform: 'scale(0.75)' },
        },
        spin: {
          from: { transform: 'rotate(0deg)' },
          to: { transform: 'rotate(360deg)' },
        },
      },
    },
  },
  plugins: [],
}
