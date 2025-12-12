/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'Outfit', 'system-ui', 'sans-serif'],
      },
      colors: {
        // NeuroTrade 风格配色
        primary: {
          50: '#FFFBEB',
          100: '#FEF3C7',
          200: '#FDE68A',
          300: '#FCD34D',
          400: '#FBBF24',
          500: '#F59E0B',
          600: '#D97706',
          700: '#B45309',
          800: '#92400E',
          900: '#78350F',
        },
        accent: {
          gold: '#D4AF37',
          amber: '#F6C343',
          cyan: '#00C8FF',
        },
        surface: {
          light: '#FFFFFF',
          muted: '#F8F7F4',
          warm: '#F5F0E8',
        },
        card: {
          dark: '#1C1C1E',
          darker: '#0D0D0D',
        }
      },
      backgroundImage: {
        'gradient-warm': 'linear-gradient(135deg, #F5F0E8 0%, #E8DFD0 50%, #D4C4A8 100%)',
        'gradient-gold': 'linear-gradient(135deg, #F6C343 0%, #D4AF37 100%)',
        'gradient-card': 'linear-gradient(180deg, #2A2A2E 0%, #1C1C1E 100%)',
      },
      borderRadius: {
        '2xl': '1rem',
        '3xl': '1.5rem',
        '4xl': '2rem',
      },
      boxShadow: {
        'card': '0 4px 24px rgba(0, 0, 0, 0.08)',
        'card-lg': '0 8px 40px rgba(0, 0, 0, 0.12)',
        'card-dark': '0 8px 32px rgba(0, 0, 0, 0.4)',
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'slide-up': {
          '0%': { opacity: '0', transform: 'translateY(20px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'pulse-slow': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.7' },
        }
      },
      animation: {
        'fade-in': 'fade-in 0.4s ease-out forwards',
        'slide-up': 'slide-up 0.5s ease-out forwards',
        'pulse-slow': 'pulse-slow 3s ease-in-out infinite',
      }
    },
  },
  plugins: [],
}
