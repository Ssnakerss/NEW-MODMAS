import React, { useRef, useEffect } from 'react';

interface Props {
  value: string;
  onChange: (v: string) => void;
  onCommit: () => void;
  onCancel: () => void;
  withTime?: boolean;
}

export const DateEditor: React.FC<Props> = ({ value, onChange, onCommit, onCancel, withTime }) => {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => { ref.current?.focus(); }, []);

  return (
    <input
      ref={ref}
      type={withTime ? 'datetime-local' : 'date'}
      value={value ?? ''}
      onChange={(e) => onChange(e.target.value)}
      onBlur={onCommit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onCommit();
        if (e.key === 'Escape') onCancel();
      }}
      className="w-full h-full px-2 py-1 text-sm border-2 border-blue-500 rounded outline-none"
    />
  );
};