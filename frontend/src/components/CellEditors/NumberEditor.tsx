import React, { useRef, useEffect } from 'react';

interface Props {
  value: number | null;
  onChange: (v: number | null) => void;
  onCommit: () => void;
  onCancel: () => void;
  isDecimal?: boolean;
}

export const NumberEditor: React.FC<Props> = ({ value, onChange, onCommit, onCancel, isDecimal }) => {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => { ref.current?.focus(); ref.current?.select(); }, []);

  return (
    <input
      ref={ref}
      type="number"
      step={isDecimal ? '0.000001' : '1'}
      value={value ?? ''}
      onChange={(e) => {
        const v = e.target.value === '' ? null : Number(e.target.value);
        onChange(v);
      }}
      onBlur={onCommit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onCommit();
        if (e.key === 'Escape') onCancel();
      }}
      className="w-full h-full px-2 py-1 text-sm text-right border-2 border-blue-500 rounded outline-none"
    />
  );
};