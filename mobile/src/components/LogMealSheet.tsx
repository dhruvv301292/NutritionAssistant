import { useRef, useState } from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';
import {
  ExpoSpeechRecognitionModule,
  useSpeechRecognitionEvent,
} from 'expo-speech-recognition';
import { Feather } from '@expo/vector-icons';
import { chatMeal, saveMeal } from '../api/client';
import { colors, fonts, radius, shadow } from '../theme';
import { SLOT_LABEL, SLOT_ORDER, suggestedSlot } from '../slots';
import type { ChatMealResponse, Slot } from '../types/api';
import NutritionTags from './NutritionTags';
import EstimateFoodForm, { needsEstimate } from './EstimateFoodForm';

type ChatEntry =
  | { role: 'user'; text: string }
  | { role: 'assistant'; response: ChatMealResponse; saved: boolean; sourceText: string }
  | { role: 'error'; text: string };

type Props = {
  visible: boolean;
  onClose: () => void;
  onMealSaved: () => void;
};

export default function LogMealSheet({ visible, onClose, onMealSaved }: Props) {
  const [input, setInput] = useState('');
  const [entries, setEntries] = useState<ChatEntry[]>([]);
  const [sending, setSending] = useState(false);
  const [listening, setListening] = useState(false);
  const [slot, setSlot] = useState<Slot>(() => suggestedSlot(new Date()));
  const textInputRef = useRef<TextInput>(null);

  useSpeechRecognitionEvent('start', () => setListening(true));
  useSpeechRecognitionEvent('end', () => setListening(false));
  useSpeechRecognitionEvent('result', (event) => {
    const transcript = event.results[0]?.transcript;
    if (transcript) setInput(transcript);
  });
  useSpeechRecognitionEvent('error', () => setListening(false));

  async function handleMicPress() {
    if (listening) {
      ExpoSpeechRecognitionModule.stop();
      return;
    }
    const result = await ExpoSpeechRecognitionModule.requestPermissionsAsync();
    if (!result.granted) return;
    ExpoSpeechRecognitionModule.start({ lang: 'en-US', interimResults: true, continuous: false });
  }

  async function sendText(text: string, { echoAsUserMessage }: { echoAsUserMessage: boolean }) {
    if (!text || sending) return;
    if (echoAsUserMessage) {
      setEntries((prev) => [...prev, { role: 'user', text }]);
    }
    setSending(true);
    try {
      const response = await chatMeal(text);
      setEntries((prev) => [...prev, { role: 'assistant', response, saved: false, sourceText: text }]);
    } catch {
      setEntries((prev) => [...prev, { role: 'error', text: 'Could not reach the server.' }]);
    } finally {
      setSending(false);
    }
  }

  async function handleSend() {
    const text = input.trim();
    setInput('');
    await sendText(text, { echoAsUserMessage: true });
  }

  // After the user saves a food via the AI-estimate flow, re-run the same
  // original message so the newly-saved food resolves this time — mirrors
  // ChatMeal.tsx's web behavior.
  async function handleRetry(sourceText: string) {
    await sendText(sourceText, { echoAsUserMessage: false });
  }

  async function handleLogMeal(index: number, response: ChatMealResponse) {
    const items = response.result.items
      .filter((item) => item.matched_food && !item.ambiguous && !item.error)
      .map((item) => ({ food_name: item.matched_food!.name, quantity: item.quantity, unit: item.unit }));
    if (items.length === 0) return;
    try {
      await saveMeal(items, slot);
      setEntries((prev) =>
        prev.map((e, i) => (i === index && e.role === 'assistant' ? { ...e, saved: true } : e))
      );
      onMealSaved();
    } catch {
      setEntries((prev) => [...prev, { role: 'error', text: 'Could not save the meal.' }]);
    }
  }

  function handleClose() {
    setEntries([]);
    setInput('');
    setSlot(suggestedSlot(new Date()));
    onClose();
  }

  return (
    <Modal
      visible={visible}
      animationType="slide"
      transparent
      onRequestClose={handleClose}
      onShow={() => textInputRef.current?.focus()}
    >
      <Pressable style={styles.backdrop} onPress={handleClose} />
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.sheet}
      >
        <View style={styles.header}>
          <Text style={styles.headerTitle}>Log a meal</Text>
          <Pressable onPress={handleClose} style={styles.closeButton}>
            <Text style={styles.closeButtonText}>✕</Text>
          </Pressable>
        </View>

        <View style={styles.slotRow}>
          {SLOT_ORDER.map((s) => {
            const active = s === slot;
            return (
              <Pressable
                key={s}
                onPress={() => setSlot(s)}
                style={[styles.slotChip, { backgroundColor: active ? colors.accent : colors.neutral200 }]}
              >
                <Text style={[styles.slotChipText, { color: active ? colors.bg : colors.text }]}>
                  {SLOT_LABEL[s]}
                </Text>
              </Pressable>
            );
          })}
        </View>

        <ScrollView style={styles.chatArea} contentContainerStyle={{ gap: 10, paddingBottom: 10 }}>
          {entries.map((entry, i) => {
            if (entry.role === 'user') {
              return (
                <View key={i} style={[styles.bubble, styles.userBubble]}>
                  <Text style={styles.userBubbleText}>{entry.text}</Text>
                </View>
              );
            }
            if (entry.role === 'error') {
              return (
                <View key={i} style={[styles.bubble, styles.assistantBubble]}>
                  <Text style={styles.errorText}>{entry.text}</Text>
                </View>
              );
            }
            const { response, saved, sourceText } = entry;
            return (
              <View key={i} style={[styles.bubble, styles.assistantBubble, { maxWidth: '100%' }]}>
                <Text style={styles.assistantIntro}>
                  {response.needs_clarification
                    ? 'I need a bit more info on some of these — resolved items can still be logged:'
                    : "Here's what I found:"}
                </Text>
                <View style={{ gap: 8, marginTop: 8 }}>
                  {response.result.items.map((item, j) => (
                    <View key={j} style={styles.itemRow}>
                      <Text style={styles.itemName}>{item.matched_food?.name ?? item.food_name}</Text>
                      {item.unconfirmed_food && (
                        <EstimateFoodForm
                          foodName={item.food_name}
                          onSaved={() => handleRetry(sourceText)}
                          externalMatch={item.unconfirmed_food}
                        />
                      )}
                      {!item.unconfirmed_food && item.error && needsEstimate(item.error) && (
                        <EstimateFoodForm foodName={item.food_name} onSaved={() => handleRetry(sourceText)} />
                      )}
                      {!item.unconfirmed_food && item.error && !needsEstimate(item.error) && (
                        <Text style={styles.errorText}>{item.error}</Text>
                      )}
                      {item.ambiguous && (
                        <Text style={styles.warningText}>
                          did you mean: {item.candidates?.map((c) => c.name).join(', ')}?
                        </Text>
                      )}
                      {item.nutrition && <NutritionTags nutrition={item.nutrition} />}
                    </View>
                  ))}
                </View>
                <Text style={styles.totalText}>
                  Total: {Math.round(response.result.total.calories)} cal,{' '}
                  {response.result.total.protein.toFixed(1)}g protein
                </Text>
                {saved ? (
                  <Text style={styles.savedText}>meal logged</Text>
                ) : (
                  <Pressable style={styles.logButton} onPress={() => handleLogMeal(i, response)}>
                    <Text style={styles.logButtonText}>log this meal</Text>
                  </Pressable>
                )}
              </View>
            );
          })}
          {sending && <ActivityIndicator color={colors.accent700} style={{ alignSelf: 'flex-start' }} />}
        </ScrollView>

        <View style={styles.inputRow}>
          <Pressable
            onPress={handleMicPress}
            style={[styles.micButton, { backgroundColor: listening ? colors.accent700 : colors.accent100 }]}
          >
            <Feather name="mic" size={17} color={listening ? colors.bg : colors.accent700} />
          </Pressable>
          <TextInput
            ref={textInputRef}
            style={styles.textInput}
            value={input}
            onChangeText={setInput}
            placeholder="Describe what you ate…"
            onSubmitEditing={handleSend}
            returnKeyType="send"
          />
          <Pressable
            onPress={handleSend}
            disabled={input.trim().length === 0}
            style={[styles.sendButton, { opacity: input.trim().length === 0 ? 0.4 : 1 }]}
          >
            <Feather name="arrow-up" size={17} color={colors.bg} />
          </Pressable>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: { flex: 1, backgroundColor: 'rgba(32,30,29,0.4)' },
  sheet: {
    position: 'absolute',
    left: 0,
    right: 0,
    bottom: 0,
    maxHeight: '85%',
    backgroundColor: colors.bg,
    borderTopLeftRadius: 28,
    borderTopRightRadius: 28,
    ...shadow.lg,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: 18,
    paddingBottom: 10,
  },
  headerTitle: { fontFamily: fonts.heading, fontSize: 17, color: colors.text },
  closeButton: {
    width: 30,
    height: 30,
    borderRadius: radius.pill,
    backgroundColor: colors.neutral200,
    alignItems: 'center',
    justifyContent: 'center',
  },
  closeButtonText: { fontFamily: fonts.body, fontSize: 14, color: colors.text },
  slotRow: { flexDirection: 'row', gap: 6, paddingHorizontal: 18, paddingBottom: 10 },
  slotChip: { flex: 1, alignItems: 'center', paddingVertical: 7, borderRadius: radius.pill },
  slotChipText: { fontFamily: fonts.bodySemiBold, fontSize: 12.5 },
  chatArea: { paddingHorizontal: 18, minHeight: 120 },
  bubble: { maxWidth: '78%', padding: 10, borderRadius: radius.lg, borderTopLeftRadius: 8 },
  userBubble: { alignSelf: 'flex-end', backgroundColor: colors.accent, borderTopLeftRadius: radius.lg, borderTopRightRadius: 8 },
  userBubbleText: { fontFamily: fonts.body, color: colors.bg, fontSize: 14 },
  assistantBubble: { alignSelf: 'flex-start', backgroundColor: colors.neutral200 },
  assistantIntro: { fontFamily: fonts.body, fontSize: 14, color: colors.text },
  itemRow: { gap: 4 },
  itemName: { fontFamily: fonts.heading, fontSize: 14, color: colors.text },
  errorText: { fontFamily: fonts.body, fontSize: 12, color: '#c0392b' },
  warningText: { fontFamily: fonts.body, fontSize: 12, color: '#d68910' },
  totalText: { fontFamily: fonts.bodyBold, marginTop: 8, fontSize: 13, color: colors.text },
  savedText: { fontFamily: fonts.bodySemiBold, color: colors.accent2_700, marginTop: 8, fontSize: 13 },
  logButton: {
    marginTop: 8,
    backgroundColor: colors.accent,
    borderRadius: radius.pill,
    paddingVertical: 8,
    alignItems: 'center',
  },
  logButtonText: { color: colors.bg, fontSize: 13, fontFamily: fonts.heading },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: colors.surface,
    borderRadius: radius.pill,
    padding: 6,
    margin: 16,
    ...shadow.md,
  },
  micButton: { width: 40, height: 40, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center' },
  textInput: { fontFamily: fonts.body, flex: 1, fontSize: 14, color: colors.text, paddingHorizontal: 8 },
  sendButton: {
    width: 40,
    height: 40,
    borderRadius: radius.pill,
    backgroundColor: colors.accent,
    alignItems: 'center',
    justifyContent: 'center',
  },
});
