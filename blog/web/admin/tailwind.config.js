export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        paper: '#fdfaf5',
        tea: '#786951',
        gold: '#b8944c',
        soup: '#f3ede3',
        gray: '#998b6e',
        border: '#d9cdb3',
        card: '#faf7f2',
        'card-border': '#e8ddd0',
      },
      fontFamily: {
        serif: ['"LXGW WenKai"', '"KaiTi"', 'serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
    },
  },
  plugins: [],
}
