import { getStoredToken } from '../auth/tokenStorage';
import type { ChatMealResponse, DailySummary, Food, Goals, MealLog, NutritionEstimate, Slot, User } from '../types/api';

// Defaults to the deployed backend so the app works without any local setup.
// For local backend development, set EXPO_PUBLIC_API_BASE_URL to your
// machine's LAN IP (Expo's dev server prints this as "Metro waiting on
// exp://<ip>:8081") — "localhost" on a physical device/simulator refers to
// the phone itself, not the dev machine.
const API_BASE_URL = process.env.EXPO_PUBLIC_API_BASE_URL ?? 'https://nutritionassistant-1r9y.onrender.com';

// AuthProvider registers itself here on mount so this module (outside React)
// can trigger a sign-out when any request comes back 401 — e.g. an expired
// session token found during normal use, not just at cold start.
let onUnauthorized: (() => void) | null = null;
export function setUnauthorizedHandler(handler: () => void): void {
  onUnauthorized = handler;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = await getStoredToken();
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
  if (res.status === 401) {
    onUnauthorized?.();
    throw new Error('session expired');
  }
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(text || `request failed: ${res.status}`);
  }
  return res.json();
}

export function googleLogin(idToken: string): Promise<{ token: string; user: User }> {
  return request('/auth/google', { method: 'POST', body: JSON.stringify({ id_token: idToken }) });
}

export function appleLogin(idToken: string, name: string): Promise<{ token: string; user: User }> {
  return request('/auth/apple', { method: 'POST', body: JSON.stringify({ id_token: idToken, name }) });
}

export function getMe(): Promise<User> {
  return request('/me');
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
    body: JSON.stringify({ slot, items }),
  });
}

export function dailySummary(date: string): Promise<DailySummary> {
  return request(`/summary/daily?date=${date}`);
}

export function getGoals(): Promise<Goals> {
  return request('/goals');
}

export function putGoals(goals: Goals): Promise<Goals> {
  return request('/goals', { method: 'PUT', body: JSON.stringify(goals) });
}

export function estimateFood(name: string): Promise<NutritionEstimate> {
  return request('/foods/estimate', { method: 'POST', body: JSON.stringify({ name }) });
}

export function createFood(estimate: NutritionEstimate): Promise<Food> {
  return request('/foods', { method: 'POST', body: JSON.stringify(estimate) });
}
