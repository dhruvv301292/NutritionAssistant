import { useRef } from 'react';
import { Animated, PanResponder, StyleSheet, View } from 'react-native';
import { colors } from '../theme';

const REVEAL_WIDTH = 64;

export default function SwipeableGoalRow({ children, toggle }: { children: React.ReactNode; toggle: React.ReactNode }) {
  const translateX = useRef(new Animated.Value(0)).current;
  const openRef = useRef(false);

  const panResponder = useRef(
    PanResponder.create({
      onMoveShouldSetPanResponder: (_, g) => Math.abs(g.dx) > 8 && Math.abs(g.dx) > Math.abs(g.dy),
      onPanResponderMove: (_, g) => {
        const base = openRef.current ? -REVEAL_WIDTH : 0;
        const next = Math.min(0, Math.max(-REVEAL_WIDTH, base + g.dx));
        translateX.setValue(next);
      },
      onPanResponderRelease: (_, g) => {
        const base = openRef.current ? -REVEAL_WIDTH : 0;
        const raw = Math.min(0, Math.max(-REVEAL_WIDTH, base + g.dx));
        const open = raw < -REVEAL_WIDTH / 2;
        openRef.current = open;
        Animated.spring(translateX, {
          toValue: open ? -REVEAL_WIDTH : 0,
          useNativeDriver: true,
          bounciness: 0,
        }).start();
      },
    })
  ).current;

  const toggleTranslateX = Animated.add(translateX, REVEAL_WIDTH);

  return (
    <View style={styles.container}>
      <Animated.View style={[styles.togglePane, { transform: [{ translateX: toggleTranslateX }] }]}>
        {toggle}
      </Animated.View>
      <Animated.View
        style={[styles.foreground, { transform: [{ translateX }] }]}
        {...panResponder.panHandlers}
      >
        {children}
      </Animated.View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { position: 'relative', overflow: 'hidden', width: '100%' },
  togglePane: {
    position: 'absolute',
    right: 0,
    top: 0,
    bottom: 0,
    width: REVEAL_WIDTH,
    alignItems: 'center',
    justifyContent: 'center',
  },
  foreground: { backgroundColor: colors.bg },
});
