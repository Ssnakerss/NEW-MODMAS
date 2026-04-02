import React from 'react';

interface Props {
  value: boolean;
  onChange: (v: boolean) => void;
  onCommit: () => void;
}

export const BooleanEditor: React.FC<Props> = ({ value, onChange, onCommit }) => (
  <div className="flex items-center justify-center w-full h-full">
    <input
      type="checkbox"
      checked={Boolean(value)}
      onChange={(e) => { onChange(e.target.checked); onCommit(); }}
      className="w-4 h-4 text-blue-600 cursor-pointer"
      autoFocus
    />
  </div>
);