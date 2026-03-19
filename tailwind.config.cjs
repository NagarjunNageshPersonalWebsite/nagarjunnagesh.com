module.exports = {
  content: [
    "./frontend/**/*.html",
    "./frontend/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        'yc-orange': '#F26522',
        'yc-dark': '#1A1A1A',
        'yc-gray': '#9CA3AF',
        'yc-border': '#E5E7EB'
      }
    },
  },
  safelist: [
    // ensure custom project tokens are always generated
    { pattern: /^(bg|text|border|hover:text|hover:border|group-hover:bg|marker:text)-yc(-[a-z0-9]+)?$/i }
  ],
  plugins: [],
}
