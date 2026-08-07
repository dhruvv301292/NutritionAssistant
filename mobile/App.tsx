import { useRef, useState } from 'react';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider, useSafeAreaInsets } from 'react-native-safe-area-context';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import PagerView from 'react-native-pager-view';
import { Feather } from '@expo/vector-icons';
import { useFonts, Caprasimo_400Regular } from '@expo-google-fonts/caprasimo';
import {
  Figtree_400Regular,
  Figtree_600SemiBold,
  Figtree_700Bold,
} from '@expo-google-fonts/figtree';
import TodayScreen from './src/screens/TodayScreen';
import HistoryScreen from './src/screens/HistoryScreen';
import GoalsScreen from './src/screens/GoalsScreen';
import { colors, fonts } from './src/theme';

const PAGES = [
  { key: 'History', icon: 'clock' as const, Screen: HistoryScreen },
  { key: 'Today', icon: 'home' as const, Screen: TodayScreen },
  { key: 'Goals', icon: 'target' as const, Screen: GoalsScreen },
];
const INITIAL_PAGE = 1; // Today

function TabBar({ activeIndex, onSelect }: { activeIndex: number; onSelect: (i: number) => void }) {
  const insets = useSafeAreaInsets();
  return (
    <View style={[styles.tabBar, { paddingBottom: insets.bottom || 10 }]}>
      {PAGES.map((page, i) => {
        const active = i === activeIndex;
        const color = active ? colors.accent700 : colors.neutral700;
        return (
          <Pressable key={page.key} style={styles.tabItem} onPress={() => onSelect(i)}>
            <Feather name={page.icon} size={22} color={color} />
            <Text style={[styles.tabLabel, { color }]}>{page.key}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

export default function App() {
  const [fontsLoaded] = useFonts({
    Caprasimo_400Regular,
    Figtree_400Regular,
    Figtree_600SemiBold,
    Figtree_700Bold,
  });
  const [activeIndex, setActiveIndex] = useState(INITIAL_PAGE);
  const pagerRef = useRef<PagerView>(null);

  if (!fontsLoaded) {
    return <View style={{ flex: 1, backgroundColor: colors.bg }} />;
  }

  function selectPage(i: number) {
    pagerRef.current?.setPage(i);
  }

  return (
    <SafeAreaProvider>
      <View style={{ flex: 1, backgroundColor: colors.bg }}>
        <PagerView
          ref={pagerRef}
          style={{ flex: 1 }}
          initialPage={INITIAL_PAGE}
          onPageSelected={(e) => setActiveIndex(e.nativeEvent.position)}
        >
          {PAGES.map((page, i) => (
            <View key={page.key} style={{ flex: 1 }}>
              <page.Screen focused={activeIndex === i} />
            </View>
          ))}
        </PagerView>
        <TabBar activeIndex={activeIndex} onSelect={selectPage} />
      </View>
      <StatusBar style="dark" />
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  tabBar: {
    flexDirection: 'row',
    backgroundColor: colors.surface,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.divider,
    paddingTop: 8,
  },
  tabItem: { flex: 1, alignItems: 'center', gap: 2 },
  tabLabel: { fontFamily: fonts.bodySemiBold, fontSize: 10.5 },
});
