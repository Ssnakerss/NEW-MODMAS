import React, { useEffect, useRef, useState } from 'react';
import { clsx } from 'clsx';

type DropdownSize = 'sm' | 'md' | 'lg';
type DropdownAlign = 'left' | 'right';

export interface DropdownItem {
  label: string;
  value: string;
  icon?: React.ReactNode;
  danger?: boolean;
  disabled?: boolean;
  divider?: boolean;
}

interface DropdownProps {
  trigger?: React.ReactNode;
  items: DropdownItem[];
  onSelect: (item: DropdownItem) => void;
  size?: DropdownSize;
  align?: DropdownAlign;
  disabled?: boolean;
  className?: string;
  // For controlled usage
  open?: boolean;
  onClose?: () => void;
}

const sizeClass: Record<DropdownSize, string> = {
  sm: 'text-xs py-1',
  md: 'text-sm py-1.5',
  lg: 'text-base py-2',
};

const itemSizeClass: Record<DropdownSize, string> = {
  sm: 'px-3 py-1 text-xs',
  md: 'px-4 py-2 text-sm',
  lg: 'px-5 py-2.5 text-base',
};

const alignClass: Record<DropdownAlign, string> = {
  left:  'left-0',
  right: 'right-0',
};

export const Dropdown: React.FC<DropdownProps> = ({
  trigger,
  items,
  onSelect,
  size = 'md',
  align = 'left',
  disabled = false,
  className,
  open: controlledOpen,
  onClose,
}) => {
  const [internalOpen, setInternalOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const isControlled = controlledOpen !== undefined;
  const open = isControlled ? controlledOpen : internalOpen;

  const setOpen = (value: boolean) => {
    if (!isControlled) {
      setInternalOpen(value);
    }
    if (!value && onClose) {
      onClose();
    }
  };

  // Закрытие по клику вне компонента
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Закрытие по Escape
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleSelect = (item: DropdownItem) => {
    if (item.disabled) return;
    onSelect(item);
    setOpen(false);
  };

  return (
    <div ref={ref} className={clsx('relative inline-block', className)}>
      {/* Триггер */}
      {trigger && (
        <div
          onClick={() => !disabled && setOpen(!open)}
          className={clsx(disabled && 'opacity-50 cursor-not-allowed')}
        >
          {trigger}
        </div>
      )}

      {/* Меню */}
      {open && (
        <div
          className={clsx(
            'absolute z-50 mt-1 min-w-[10rem] bg-white border border-gray-200',
            'rounded-lg shadow-lg overflow-hidden',
            sizeClass[size],
            alignClass[align],
          )}
        >
          {items.map((item, index) => (
            <React.Fragment key={item.value ?? index}>
              {/* Разделитель */}
              {item.divider && (
                <div className="my-1 border-t border-gray-100" />
              )}

              {/* Пункт меню */}
              <button
                type="button"
                disabled={item.disabled}
                onClick={() => handleSelect(item)}
                className={clsx(
                  'w-full flex items-center gap-2 text-left transition-colors',
                  'focus:outline-none focus:bg-gray-50',
                  itemSizeClass[size],
                  item.danger
                    ? 'text-red-600 hover:bg-red-50'
                    : 'text-gray-700 hover:bg-gray-100',
                  item.disabled && 'opacity-40 cursor-not-allowed pointer-events-none',
                )}
              >
                {item.icon && (
                  <span className="shrink-0 w-4 h-4">{item.icon}</span>
                )}
                {item.label}
              </button>
            </React.Fragment>
          ))}
        </div>
      )}
    </div>
  );
};