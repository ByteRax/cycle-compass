export type ThemeMode = 'light' | 'dark'

export interface ThemeColors {
  primary: string
  success: string
  warning: string
  danger: string
}

export const themeColorOptions = [
  {
    name: '经典深蓝',
    primary: '#1e3a8a',
    isCustom: false
  }
]

export interface SystemThemeState {
  theme: ThemeMode
  themeVars: Record<string, string>
}
