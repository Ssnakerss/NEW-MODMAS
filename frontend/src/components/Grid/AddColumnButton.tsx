import React from 'react';

export const AddColumnButton: React.FC<{ onClick: () => void }> = ({ onClick }) => (
  <button
    onClick={onClick}
    className="flex items-center justify-center w-12 h-9 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors shrink-0"
    title="Добавить поле"
  >
    <span className="text-xl leading-none">+</span>
  </button>
);