import { useState } from 'react'
import type { ChatMealResponse, ItemResult } from './types'
import EstimateFoodForm, { needsEstimate } from './EstimateFoodForm'

const CURRENT_USER_ID = 1

type ChatMessage =
  | { role: 'user', text: string }
  | { role: 'assistant', response: ChatMealResponse, saved: boolean, sourceText: string }
  | { role: 'assistant-error', text: string }

type Props = {
  onLogged: () => void,
}

function NutritionCard({ item, onFoodAdded }: { item: ItemResult, onFoodAdded: () => void }) {
  if (item.unconfirmed_food) {
    return (
      <div className="nutrition-card nutrition-card-warning">
        <strong>{item.food_name}</strong>
        <EstimateFoodForm foodName={item.food_name} onSaved={onFoodAdded} externalMatch={item.unconfirmed_food} />
      </div>
    )
  }
  if (item.error && needsEstimate(item.error)) {
    return (
      <div className="nutrition-card nutrition-card-error">
        <strong>{item.food_name}</strong>
        <EstimateFoodForm foodName={item.food_name} onSaved={onFoodAdded} />
      </div>
    )
  }
  if (item.error) {
    return (
      <div className="nutrition-card nutrition-card-error">
        <strong>{item.food_name}</strong>
        <span className="error">{item.error}</span>
      </div>
    )
  }
  if (item.ambiguous) {
    return (
      <div className="nutrition-card nutrition-card-warning">
        <strong>{item.food_name}</strong>
        <span className="warning">
          did you mean: {item.candidates?.map(c => c.name).join(', ')}?
        </span>
      </div>
    )
  }
  return (
    <div className="nutrition-card">
      <strong>{item.matched_food?.name ?? item.food_name}</strong>
      <span>{item.quantity} {item.unit}</span>
      {item.nutrition && (
        <span>
          {item.nutrition.calories.toFixed(0)} cal · {item.nutrition.protein.toFixed(1)}g protein ·{' '}
          {item.nutrition.carbs.toFixed(1)}g carbs · {item.nutrition.fat.toFixed(1)}g fat
        </span>
      )}
    </div>
  )
}

export default function ChatMeal({ onLogged }: Props) {
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [loading, setLoading] = useState(false)

  async function sendText(text: string, { echoAsUserMessage }: { echoAsUserMessage: boolean }) {
    if (text === '' || loading) return

    if (echoAsUserMessage) {
      setMessages(prev => [...prev, { role: 'user', text }])
    }
    setLoading(true)

    try {
      const res = await fetch('/api/chat/meal', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text }),
      })
      if (!res.ok) {
        const errText = await res.text()
        setMessages(prev => [...prev, { role: 'assistant-error', text: errText || 'Something went wrong.' }])
        return
      }
      const data: ChatMealResponse = await res.json()
      setMessages(prev => [...prev, { role: 'assistant', response: data, saved: false, sourceText: text }])
    } catch {
      setMessages(prev => [...prev, { role: 'assistant-error', text: 'Could not reach the server.' }])
    } finally {
      setLoading(false)
    }
  }

  async function handleSend() {
    const text = input.trim()
    setInput('')
    await sendText(text, { echoAsUserMessage: true })
  }

  // After the user saves a food via the AI-estimate flow, re-run the same
  // original message — the newly-saved food should now resolve from
  // Postgres. Shown as a fresh assistant turn rather than trying to patch
  // just the one item, since other items in the same message may also
  // depend on re-resolution.
  async function handleRetry(sourceText: string) {
    await sendText(sourceText, { echoAsUserMessage: false })
  }

  async function handleLogMeal(messageIndex: number, response: ChatMealResponse) {
    const items = response.result.items
      .filter(item => item.matched_food && !item.ambiguous && !item.error)
      .map(item => ({ food_name: item.matched_food!.name, quantity: item.quantity, unit: item.unit }))
    if (items.length === 0) return

    const res = await fetch('/api/meals', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: CURRENT_USER_ID, items }),
    })
    if (!res.ok) return

    setMessages(prev => prev.map((m, i) =>
      i === messageIndex && m.role === 'assistant' ? { ...m, saved: true } : m
    ))
    onLogged()
  }

  return (
    <div className="chat-meal">
      <h2>Log a Meal (natural language)</h2>
      <div className="chat-history">
        {messages.map((message, i) => {
          if (message.role === 'user') {
            return <div key={i} className="chat-message chat-message-user">{message.text}</div>
          }
          if (message.role === 'assistant-error') {
            return <div key={i} className="chat-message chat-message-assistant error">{message.text}</div>
          }
          const { response, saved, sourceText } = message
          const hasUnresolvedFood = response.result.items.some(item => needsEstimate(item.error) || item.unconfirmed_food)
          return (
            <div key={i} className="chat-message chat-message-assistant">
              {response.needs_clarification ? (
                <p className="warning">I need a bit more info on some of these — resolved items can still be logged below:</p>
              ) : (
                <p>Here's what I found:</p>
              )}
              <div className="nutrition-cards">
                {response.result.items.map((item, j) => (
                  <NutritionCard key={j} item={item} onFoodAdded={() => handleRetry(sourceText)} />
                ))}
              </div>
              <div className="chat-meal-total">
                <strong>Total:</strong> {response.result.total.calories.toFixed(0)} cal,{' '}
                {response.result.total.protein.toFixed(1)}g protein,{' '}
                {response.result.total.carbs.toFixed(1)}g carbs,{' '}
                {response.result.total.fat.toFixed(1)}g fat
              </div>
              {saved ? (
                <p className="success">meal logged</p>
              ) : (
                <button type="button" onClick={() => handleLogMeal(i, response)}>
                  log this meal
                </button>
              )}
              {hasUnresolvedFood && !saved && (
                <p className="warning">Add the food above, then it'll be included automatically next time.</p>
              )}
            </div>
          )
        })}
        {loading && <div className="chat-message chat-message-assistant">thinking…</div>}
      </div>
      <div className="chat-input-row">
        <input
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter') handleSend() }}
          placeholder="I had 230g chicken thighs and a Quest protein shake…"
        />
        <button type="button" onClick={handleSend} disabled={loading}>send</button>
      </div>
    </div>
  )
}
