import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

const crowdSpeechTheme = {
  dark: false,
  colors: {
    background: '#ffffff',
    surface: '#ffffff',
    primary: '#000000',
    'primary-darken-1': '#111111',
    secondary: '#111111',
    'secondary-darken-1': '#000000',
    error: '#000000',
    info: '#000000',
    success: '#000000',
    warning: '#000000',
    'on-background': '#000000',
    'on-surface': '#000000',
    'on-primary': '#ffffff',
    'on-secondary': '#ffffff',
  },
}

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'crowdSpeechTheme',
    themes: { crowdSpeechTheme },
  },
  defaults: {
    global: {
      rounded: 0,
      elevation: 0,
    },
    VBtn: {
      rounded: 0,
      elevation: 0,
      variant: 'flat',
    },
    VCard: {
      rounded: 0,
      elevation: 0,
    },
    VTextField: {
      rounded: 0,
      variant: 'outlined',
    },
    VSelect: {
      rounded: 0,
      variant: 'outlined',
    },
    VSnackbar: {
      rounded: 0,
    },
    VDialog: {
      rounded: 0,
    },
    VChip: {
      rounded: 0,
    },
    VAlert: {
      rounded: 0,
    },
  },
  icons: {
    defaultSet: 'mdi',
  },
})
