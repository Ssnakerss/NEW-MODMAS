import React, { useState } from 'react';
import type { SelectOption } from '../../types';

const COLORS = ['#ef4444','#f97316','#eab308','#22c55e','#3b82f6','#8b5cf6','#ec4899','#6b7280'];

interface Props {
  choices: SelectOption[];
  onChange: (choices: SelectOption[]) => void;
}

export const SelectOptionsEditor: React.FC<Props> = ({ choices, onChange }) => {
  const [newLabel, setNewLabel] = useState('');
  const [selectedColor, setSelectedColor] = useState(COLORS[0]);

  const addChoice = () => {
    if (!newLabel.trim()) return;
    const choice: SelectOption = {
      value: newLabel.trim().toLowerCase().replace(/\s+/g, '_'),
      label: newLabel.trim(),
      color: selectedColor,
    };
    onChange([...choices, choice]);
    setNewLabel('');
  };

  const removeChoice = (value: string) => {
    onChange(choices.filter(c => c.value !== value));
  };

  const updateColor = (value: string, color: string) => {
    onChange(choices.map(c => c.value === value ? { ...c, color } : c));
  };

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-gray-700">Варианты</label>

      <div className="space-y-1 max-h-40 overflow-y-auto">
        {choices.map(choice => (
          <div key={choice.value} className="flex items-center gap-2 group">
            <div className="flex gap-1">
              {COLORS.map(c => (
                <button
                  key={c}
                  type="button"
                  onClick={() => updateColor(choice.value, c)}
                  className={`w-4 h-4 rounded-full transition-transform \${choice.color === c ? 'scale-125 ring-2 ring-offset-1 ring-gray-400' : ''}`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
            <span
              className="flex-1 px-2 py-0.5 rounded-full text-xs font-medium"
              style={{ backgroundColor: (choice.color ?? '#6b7280') + '22', color: choice.color ?? '#6b7280' }}
            >
              {choice.label}
            </span>
            <button
              type="button"
              onClick={() => removeChoice(choice.value)}
              className="text-gray-300 hover:text-red-500 opacity-0 group-hover:opacity-100"
            >
              ✕
            </button>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-2">
        <div className="flex gap-1">
          {COLORS.map(c => (
            <button
              key={c}
              type="button"
              onClick={() => setSelectedColor(c)}
              className={`w-4 h-4 rounded-full \${selectedColor === c ? 'ring-2 ring-offset-1 ring-gray-400 scale-125' : ''}`}
              style={{ backgroundColor: c }}
            />
          ))}
        </div>
        <input
          value={newLabel}
          onChange={(e) => setNewLabel(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addChoice(); } }}
          placeholder="Новый вариант..."
          className="flex-1 border border-gray-300 rounded px-2 py-1 text-sm outline-none focus:ring-1 focus:ring-blue-500"
        />
        <button
          type="button"
          onClick={addChoice}
          className="px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded text-sm font-medium"
        >
          +
        </button>
      </div>
    </div>
  );
};