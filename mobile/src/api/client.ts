import type { ChatMealResponse, DailySummary, Goals, MealLog, NutritionEstimate, Slot } from '../types/api';

// Defaults to the deployed backend so the app works without any local setup.
// For local backend development, set EXPO_PUBLIC_API_BASE_URL to your
// machine's LAN IP (Expo's dev server prints this as "Metro waiting on
// exp://<ip>:8081") — "localhost" on a physical device/simulator refers to
// the phone itself, not the dev machine.
const API_BASE_URL = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'https://nutritionassistant-1r9y.onrender.com';

const CURRENT_USER_ID = 1;

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  });
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(text || `request failed: ${res.status}`);
  }
  return res.json();
}

export function chatMeal(text: string): Promise<ChatMealResponse> {
  return request('/chat/meal', { method: 'POST', body: JSON.stringify({ text }) });
}

export function saveMeal(
  items: { food_name: string; quantity: number; unit: string }[],
  slot: Slot
): Promise<MealLog> {
  return request('/meals', {
    method: 'POST',
    body: JSON.stringify({ user_id: CURRENT_USER_ID, slot, items }),
  });
}

export function dailySummary(date: string): Promise<DailySummary> {
  return request(`/summary/daily?user_id=${CURRENT_USER_ID}&date=${date}`);
}

export function getGoals(): Promise<Goals> {
  return request(`/goals?user_id=${CURRENT_USER_ID}`);
}

export function putGoals(goals: Goals): Promise<Goals> {
  return request('/goals', { method: 'PUT', body: JSON.stringify(goals) });
}

export function estimateFood(name: string): Promise<NutritionEstimate> {
  return request('/foods/estimate', { method: 'POST', body: JSON.stringify({ name }) });
}

export function createFood(estimate: NutritionEstimate): Promise<unknown> {
  return request('/foods', { method: 'POST', body: JSON.stringify(estimate) });
}

export { CURRENT_USER_ID };
