import React, { forwardRef } from 'react';
import { clsx } from 'clsx';

type SelectSize = 'sm' | 'md' | 'lg';
type SelectVariant = 'default' | 'error' | 'success';

export interface SelectOption {
  label: string;
  value: string;
  disabled?: boolean;
}

export interface SelectGroup {
  label: string;
  options: SelectOption[];
}

interface SelectProps extends Omit<React.SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label?: string;
  hint?: string;
  error?: string;
  size?: SelectSize;
  variant?: SelectVariant;
  options?: SelectOption[];
  groups?: SelectGroup[];
  placeholder?: string;
  loading?: boolean;
}

const sizeClass: Record<SelectSize, string> = {
  sm: 'px-2.5 py-1.5 text-sm',
  md: 'px-3 py-2 text-sm',
  lg: 'px-4 py-2.5 text-base',
};

const variantClass: Record<SelectVariant, string> = {
  default: 'border-gray-300 focus:border-blue-500 focus:ring-blue-500',
  error:   'border-red-400 focus:border-red-500 focus:ring-red-500',
  success: 'border-green-400 focus:border-green-500 focus:ring-green-500',
};

export const Select = forwardRef<HTMLSelectElement, SelectProps>(({
  label,
  hint,
  error,
  size = 'md',
  variant = 'default',
  options = [],
  groups = [],
  placeholder,
  loading = false,
  className,
  disabled,
  id,
  ...props
}, ref) => {
  const selectId = id ?? label?.toLowerCase().replace(/\s+/g, '-');
  const resolvedVariant: SelectVariant = error ? 'error' : variant;

  return (
    <div className="flex flex-col gap-1 w-full">

      {/* Label */}
      {label && (
        <label
          htmlFor={selectId}
          className="text-sm font-medium text-gray-700"
        >
          {label}
          {props.required && (
            <span className="ml-1 text-red-500">*</span>
          )}
        </label>
      )}

      {/* Select wrapper */}
      <div className="relative flex items-center">

        {/* Select */}
        <select
          {...props}
          ref={ref}
          id={selectId}
          disabled={disabled || loading}
          className={clsx(
            'w-full rounded-lg border bg-white outline-none transition-colors appearance-none',
            'focus:ring-2 focus:ring-offset-0',
            'disabled:bg-gray-50 disabled:text-gray-400 disabled:cursor-not-allowed',
            sizeClass[size],
            variantClass[resolvedVariant],
            'pr-9',
            className,
          )}
        >
          {/* Placeholder */}
          {placeholder && (
            <option value="" disabled hidden>
              {placeholder}
            </option>
          )}

          {/* Flat options */}
          {options.map(option => (
            <option
              key={option.value}
              value={option.value}
              disabled={option.disabled}
            >
              {option.label}
            </option>
          ))}

          {/* Grouped options */}
          {groups.map(group => (
            <optgroup key={group.label} label={group.label}>
              {group.options.map(option => (
                <option
                  key={option.value}
                  value={option.value}
                  disabled={option.disabled}
                >
                  {option.label}
                </option>
              ))}
            </optgroup>
          ))}
        </select>

        {/* Chevron / spinner */}
        <span className="absolute right-3 flex items-center text-gray-400 pointer-events-none">
          {loading ? (
            <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle
                className="opacity-25"
                cx="12" cy="12" r="10"
                stroke="currentColor" strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8v8H4z"
              />
            </svg>
          ) : (
            <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M5.22 8.22a.75.75 0 011.06 0L10 11.94l3.72-3.72a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.22 9.28a.75.75 0 010-1.06z" clipRule="evenodd" />
            </svg>
          )}
        </span>
      </div>

      {/* Error message */}
      {error && (
        <p className="text-xs text-red-500 flex items-center gap-1">
          <svg className="w-3.5 h-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-4.5A.75.75 0 0110 5zm0 10a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
          </svg>
          {error}
        </p>
      )}

      {/* Hint */}
      {hint && !error && (
        <p className="text-xs text-gray-400">{hint}</p>
      )}

    </div>
  );
});

Select.displayName = 'Select';