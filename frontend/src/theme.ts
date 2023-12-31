import { createTheme } from '@mantine/core'
import { themeToVars } from '@mantine/vanilla-extract'

export const theme = createTheme({
  fontFamily: 'san-serif',
})

export const vars = themeToVars(theme)
