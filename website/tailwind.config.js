module.exports = {
  future: {
    removeDeprecatedGapUtilities: true,
    purgeLayersByDefault: true,
  },
  theme: {
    extend: {
      colors: {
        'black': '#1c1c1c',
        'black2': "#242424",
        'white2': "#F9F9F9",
        'white3': '#f4f4f4',
        'tgrey-black': "#cccccc",
        'tgrey-white': "#5e5e5e",
        'main': 'var(--color-primary)',
      },
    },
    fontFamily: {
      sans: ['"Mada"', 'sans-serif']
    }
  },
  variants: {},
  plugins: []
}
