import { useState } from 'react'
import type { ItemRequest, CalculateResponse } from './types'

const UNITS = ['grams', 'ounces', 'count']
const CURRENT_USER_ID = 1

type Props = {
  onSaved: () => void,
}

function emptyItem(): ItemRequest {
  return { food_name: '', quantity: 0, unit: 'grams' }
}

export default function MealLogger({ onSaved }: Props) {
  const [items, setItems] = useState<ItemRequest[]>([emptyItem()])
  const [preview, setPreview] = useState<CalculateResponse | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  function updateItem(index: number, patch: Partial<ItemRequest>) {
    setItems(prev => prev.map((item, i) => (i === index ? { ...item, ...patch } : item)))
    setPreview(null)
    setPreviewError(null)
    setSaveError(null)
  }

  function addItem() {
    setItems(prev => [...prev, emptyItem()])
  }

  function removeItem(index: number) {
    setItems(prev => prev.filter((_, i) => i !== index))
    setPreview(null)
  }

  async function handlePreview() {
    const validItems = items.filter(i => i.food_name.trim() !== '' && i.quantity > 0)
    if (validItems.length === 0) return
    setPreviewLoading(true)
    setPreviewError(null)
    try {
      const res = await fetch('/api/meals/calculate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ items: validItems }),
      })
      if (!res.ok) {
        setPreviewError('Could not calculate nutrition. Try again.')
        return
      }
      const data = await res.json()
      setPreview(data)
    } catch {
      setPreviewError('Could not reach the server.')
    } finally {
      setPreviewLoading(false)
    }
  }

  async function handleSave() {
    const validItems = items.filter(i => i.food_name.trim() !== '' && i.quantity > 0)
    if (validItems.length === 0) return
    setSaving(true)
    setSaveError(null)
    try {
      const res = await fetch('/api/meals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: CURRENT_USER_ID, items: validItems }),
      })
      if (!res.ok) {
        const data = await res.json().catch(() => null)
        setSaveError(data?.error ?? 'failed to save meal')
        return
      }
      setItems([emptyItem()])
      setPreview(null)
      onSaved()
    } catch {
      setSaveError('Could not reach the server.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="meal-logger">
      <h2>Log a Meal</h2>
      {items.map((item, index) => (
        <div className="meal-item-row" key={index}>
          <input
            placeholder="food name"
            value={item.food_name}
            onChange={e => updateItem(index, { food_name: e.target.value })}
          />
          <input
            type="number"
            placeholder="quantity"
            value={item.quantity || ''}
            onChange={e => updateItem(index, { quantity: parseFloat(e.target.value) || 0 })}
          />
          <select value={item.unit} onChange={e => updateItem(index, { unit: e.target.value })}>
            {UNITS.map(u => <option key={u} value={u}>{u}</option>)}
          </select>
          {items.length > 1 && (
            <button type="button" onClick={() => removeItem(index)}>remove</button>
          )}
        </div>
      ))}

      <div className="meal-logger-actions">
        <button type="button" onClick={addItem}>+ add ingredient</button>
        <button type="button" onClick={handlePreview} disabled={previewLoading}>
          {previewLoading ? 'calculating…' : 'preview'}
        </button>
        <button type="button" onClick={handleSave} disabled={saving}>
          {saving ? 'saving…' : 'save meal'}
        </button>
      </div>

      {previewError && <p className="error">{previewError}</p>}
      {saveError && <p className="error">{saveError}</p>}

      {preview && (
        <div className="meal-preview">
          {preview.items.map((result, i) => (
            <div key={i} className="meal-preview-item">
              <strong>{result.food_name}</strong>{' '}
              {result.error && <span className="error">{result.error}</span>}
              {result.ambiguous && (
                <span className="warning">
                  ambiguous — matches: {result.candidates?.map(c => c.name).join(', ')}
                </span>
              )}
              {result.nutrition && (
                <span>{result.nutrition.calories.toFixed(0)} cal, {result.nutrition.protein.toFixed(1)}g protein</span>
              )}
            </div>
          ))}
          <div className="meal-preview-total">
            <strong>Total:</strong> {preview.total.calories.toFixed(0)} cal,{' '}
            {preview.total.protein.toFixed(1)}g protein, {preview.total.carbs.toFixed(1)}g carbs,{' '}
            {preview.total.fat.toFixed(1)}g fat, {preview.total.fiber.toFixed(1)}g fiber
          </div>
        </div>
      )}
    </div>
  )
}
