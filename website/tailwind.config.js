module.exports = {
  future: {
    removeDeprecatedGapUtilities: true,
    purgeLayersByDefault: true,
  },
  purge: ['./components/**/*.{js,ts,jsx,tsx}', './pages/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        'black': '#1c1c1c',
        'black2': "#242424",
        'white2': "#F9F9F9",
        'white3': '#f4f4f4',
        'tgrey-black': "#cccccc",
        'tgrey-white': "#5e5e5e",
        'main': '#00d1b1',
        'ensemble': '#00d1b1',
      },
      /*
      inset: {
        '16': '4rem !important'
      },
      */
      backgroundImage: theme => ({
        'use-cases': "url('/texture-squares.svg')",
      }),
    },
    boxShadow: {
      'use-cases': '7px 7px 0 0 rgba(50,50,50,.11)'
    },
    fontFamily: {
      sans: ['"Mada"', 'sans-serif']
    }
  },
  variants: {},
  plugins: []
}
