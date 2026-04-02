import React, { useRef, useEffect } from 'react';
import type { SelectOption } from '../../types';

interface Props {
  value: string | string[];
  options: SelectOption[];
  isMulti?: boolean;
  onChange: (v: string | string[]) => void;
  onCommit: () => void;
  onCancel: () => void;
}

export const SelectEditor: React.FC<Props> = ({ value, options, isMulti, onChange, onCommit, onCancel }) => {
  const ref = useRef<HTMLSelectElement>(null);
  useEffect(() => { ref.current?.focus(); }, []);

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    if (isMulti) {
      const selected = Array.from(e.target.selectedOptions).map(o => o.value);
      onChange(selected);
    } else {
      onChange(e.target.value);
      onCommit();
    }
  };

  return (
    <select
      ref={ref}
      multiple={isMulti}
      value={isMulti ? (value as string[]) : (value as string)}
      onChange={handleChange}
      onBlur={onCommit}
      onKeyDown={(e) => { if (e.key === 'Escape') onCancel(); }}
      className="w-full h-full px-2 py-1 text-sm border-2 border-blue-500 rounded outline-none bg-white"
    >
      {!isMulti && <option value="">— не выбрано —</option>}
      {options.map(opt => (
        <option key={opt.value} value={opt.value}>{opt.label}</option>
      ))}
    </select>
  );
};