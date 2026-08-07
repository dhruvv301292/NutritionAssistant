import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { View } from 'react-native';
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
import { colors } from './src/theme';

const Tab = createBottomTabNavigator();

const TAB_ICON: Record<string, keyof typeof Feather.glyphMap> = {
  History: 'clock',
  Today: 'home',
  Goals: 'target',
};

export default function App() {
  const [fontsLoaded] = useFonts({
    Caprasimo_400Regular,
    Figtree_400Regular,
    Figtree_600SemiBold,
    Figtree_700Bold,
  });

  if (!fontsLoaded) {
    return <View style={{ flex: 1, backgroundColor: colors.bg }} />;
  }

  return (
    <SafeAreaProvider>
      <NavigationContainer>
        <Tab.Navigator
          initialRouteName="Today"
          screenOptions={({ route }) => ({
            headerShown: false,
            tabBarActiveTintColor: colors.accent700,
            tabBarInactiveTintColor: colors.neutral700,
            tabBarStyle: { backgroundColor: colors.surface, borderTopColor: colors.divider },
            tabBarLabelStyle: { fontFamily: 'Figtree_600SemiBold', fontSize: 10.5 },
            tabBarIcon: ({ color, size }) => <Feather name={TAB_ICON[route.name]} size={size} color={color} />,
          })}
        >
          <Tab.Screen name="History" component={HistoryScreen} />
          <Tab.Screen name="Today" component={TodayScreen} />
          <Tab.Screen name="Goals" component={GoalsScreen} />
        </Tab.Navigator>
      </NavigationContainer>
      <StatusBar style="dark" />
    </SafeAreaProvider>
  );
}
