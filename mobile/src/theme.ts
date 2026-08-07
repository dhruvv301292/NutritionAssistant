// Ported from the "Organic" Claude Design system's styles.css tokens —
// React Native has no CSS custom properties, so this is the same palette
// as a plain object. Keep in sync by hand if the design system changes.
export const colors = {
  bg: '#f5ead8',
  surface: '#ebddc5',
  text: '#201e1d',
  accent: '#c67139',
  accent2: '#7a8a5e',
  divider: 'rgba(32,30,29,0.16)',

  neutral100: '#f9f4ed',
  neutral200: '#eee7db',
  neutral300: '#dcd3c4',
  neutral400: '#c0b6a5',
  neutral500: '#a19786',
  neutral600: '#82796a',
  neutral700: '#645c50',
  neutral800: '#474238',
  neutral900: '#2e2b25',

  accent100: '#fff2eb',
  accent200: '#ffe1d0',
  accent300: '#ffc6a5',
  accent400: '#f6a06b',
  accent500: '#d67f48',
  accent600: '#b2622d',
  accent700: '#8c491a',
  accent800: '#643312',
  accent900: '#402310',

  accent2_100: '#f0fae1',
  accent2_500: '#8fa073',
  accent2_700: '#56633f',
  accent2_800: '#3d472b',
} as const;

export const fonts = {
  heading: 'Caprasimo_400Regular',
  body: 'Figtree_400Regular',
  bodySemiBold: 'Figtree_600SemiBold',
  bodyBold: 'Figtree_700Bold',
} as const;

export const radius = { sm: 8, md: 16, lg: 28, pill: 999 } as const;

export const space = { 1: 4, 2: 9, 3: 13, 4: 18, 6: 26, 8: 35 } as const;

export const shadow = {
  sm: { shadowColor: colors.neutral900, shadowOpacity: 0.14, shadowRadius: 2, shadowOffset: { width: 0, height: 1 }, elevation: 2 },
  md: { shadowColor: colors.neutral900, shadowOpacity: 0.16, shadowRadius: 10, shadowOffset: { width: 0, height: 3 }, elevation: 4 },
  lg: { shadowColor: colors.neutral900, shadowOpacity: 0.22, shadowRadius: 32, shadowOffset: { width: 0, height: 12 }, elevation: 8 },
} as const;
