import { useState } from 'react'
import type { Food, NutritionEstimate } from './types'

type Props = {
  foodName: string,
  onSaved: () => void,
  // When set, this is an unconfirmed match from an external food database
  // (USDA/FatSecret) rather than a total miss — pre-fills the edit form
  // instead of requiring a separate "suggest values with AI" step.
  externalMatch?: Food,
}

function toEstimate(food: Food): NutritionEstimate {
  return {
    name: food.name,
    calories: food.calories,
    protein: food.protein,
    carbs: food.carbs,
    fat: food.fat,
    fiber: food.fiber,
    sodium: food.sodium,
    unit: food.unit,
    unitquantity: food.unitquantity,
  }
}

export default function EstimateFoodForm({ foodName, onSaved, externalMatch }: Props) {
  const [estimate, setEstimate] = useState<NutritionEstimate | null>(externalMatch ? toEstimate(externalMatch) : null)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleEstimate() {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch('/api/foods/estimate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: foodName }),
      })
      if (!res.ok) {
        setError('Could not get an estimate. Try again.')
        return
      }
      setEstimate(await res.json())
    } catch {
      setError('Could not reach the server.')
    } finally {
      setLoading(false)
    }
  }

  function updateField(field: keyof NutritionEstimate, value: string) {
    setEstimate(prev => prev ? { ...prev, [field]: field === 'name' || field === 'unit' ? value : parseFloat(value) || 0 } : prev)
  }

  async function handleSave() {
    if (!estimate) return
    setSaving(true)
    setError(null)
    try {
      const res = await fetch('/api/foods', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(estimate),
      })
      if (!res.ok) {
        setError('Could not save this food. Try again.')
        return
      }
      onSaved()
    } catch {
      setError('Could not reach the server.')
    } finally {
      setSaving(false)
    }
  }

  if (!estimate) {
    return (
      <div className="estimate-form">
        <p className="warning">"{foodName}" isn't in our database yet.</p>
        <button type="button" onClick={handleEstimate} disabled={loading}>
          {loading ? 'estimating…' : 'suggest values with AI'}
        </button>
        {error && <p className="error">{error}</p>}
      </div>
    )
  }

  return (
    <div className="estimate-form">
      <p className="warning">
        {externalMatch
          ? 'Found in an external food database — review and edit before saving:'
          : 'AI-suggested values — review and edit before saving:'}
      </p>
      <div className="estimate-fields">
        <label>
          Name
          <input value={estimate.name} onChange={e => updateField('name', e.target.value)} />
        </label>
        <label>
          Unit
          <select value={estimate.unit} onChange={e => updateField('unit', e.target.value)}>
            <option value="grams">grams</option>
            <option value="count">count</option>
          </select>
        </label>
        <label>
          Per (unit quantity)
          <input type="number" value={estimate.unitquantity} onChange={e => updateField('unitquantity', e.target.value)} />
        </label>
        <label>
          Calories
          <input type="number" value={estimate.calories} onChange={e => updateField('calories', e.target.value)} />
        </label>
        <label>
          Protein (g)
          <input type="number" value={estimate.protein} onChange={e => updateField('protein', e.target.value)} />
        </label>
        <label>
          Carbs (g)
          <input type="number" value={estimate.carbs} onChange={e => updateField('carbs', e.target.value)} />
        </label>
        <label>
          Fat (g)
          <input type="number" value={estimate.fat} onChange={e => updateField('fat', e.target.value)} />
        </label>
        <label>
          Fiber (g)
          <input type="number" value={estimate.fiber} onChange={e => updateField('fiber', e.target.value)} />
        </label>
      </div>
      <button type="button" onClick={handleSave} disabled={saving}>
        {saving ? 'saving…' : 'save & add to database'}
      </button>
      {error && <p className="error">{error}</p>}
    </div>
  )
}
