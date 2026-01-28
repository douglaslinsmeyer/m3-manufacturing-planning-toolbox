import React, { useState, useRef, useEffect } from 'react';
import { ChevronDown } from 'lucide-react';

export interface ActionMenuItem {
  id: string;
  label: string;
  icon?: React.ReactNode;
  variant: 'default' | 'danger' | 'warning' | 'success' | 'info';
  disabled?: boolean;
  disabledReason?: string;
}

export interface ActionMenuButtonProps {
  label: string;
  icon?: React.ReactNode;
  actions: ActionMenuItem[];
  onActionSelect: (actionId: string) => void;
  disabled?: boolean;
}

export const ActionMenuButton: React.FC<ActionMenuButtonProps> = ({
  label,
  icon,
  actions,
  onActionSelect,
  disabled = false,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const [openUpward, setOpenUpward] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const menuItemsRef = useRef<(HTMLButtonElement | null)[]>([]);

  // Calculate positioning when opening
  useEffect(() => {
    if (isOpen && buttonRef.current) {
      const buttonRect = buttonRef.current.getBoundingClientRect();
      const viewportHeight = window.innerHeight;
      const spaceBelow = viewportHeight - buttonRect.bottom;
      const estimatedDropdownHeight = Math.min(actions.length * 44 + 8, 320);

      // Open upward if not enough space below
      setOpenUpward(spaceBelow < estimatedDropdownHeight && buttonRect.top > estimatedDropdownHeight);
    }
  }, [isOpen, actions.length]);

  // Click outside detection
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
        setFocusedIndex(-1);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isOpen]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!isOpen) return;

      switch (event.key) {
        case 'ArrowDown':
          event.preventDefault();
          setFocusedIndex((prev) => {
            const nextIndex = prev < actions.length - 1 ? prev + 1 : 0;
            menuItemsRef.current[nextIndex]?.focus();
            return nextIndex;
          });
          break;

        case 'ArrowUp':
          event.preventDefault();
          setFocusedIndex((prev) => {
            const prevIndex = prev > 0 ? prev - 1 : actions.length - 1;
            menuItemsRef.current[prevIndex]?.focus();
            return prevIndex;
          });
          break;

        case 'Escape':
          event.preventDefault();
          setIsOpen(false);
          setFocusedIndex(-1);
          buttonRef.current?.focus();
          break;

        case 'Enter':
          if (focusedIndex >= 0 && !actions[focusedIndex].disabled) {
            event.preventDefault();
            handleActionClick(actions[focusedIndex].id);
          }
          break;

        case 'Tab':
          // Allow tabbing out to close
          setIsOpen(false);
          setFocusedIndex(-1);
          break;
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      return () => document.removeEventListener('keydown', handleKeyDown);
    }
  }, [isOpen, focusedIndex, actions]);

  const handleButtonClick = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
      if (!isOpen) {
        setFocusedIndex(-1);
      }
    }
  };

  const handleActionClick = (actionId: string) => {
    setIsOpen(false);
    setFocusedIndex(-1);
    onActionSelect(actionId);
    // Return focus to button after action
    setTimeout(() => buttonRef.current?.focus(), 0);
  };

  const getVariantClasses = (variant: ActionMenuItem['variant'], isDisabled: boolean) => {
    if (isDisabled) {
      return 'text-gray-400 cursor-not-allowed';
    }

    switch (variant) {
      case 'danger':
        return 'text-red-700 hover:bg-red-50';
      case 'warning':
        return 'text-orange-700 hover:bg-orange-50';
      case 'success':
        return 'text-green-700 hover:bg-green-50';
      case 'info':
        return 'text-blue-700 hover:bg-blue-50';
      default:
        return 'text-gray-700 hover:bg-gray-50';
    }
  };

  return (
    <div className="relative inline-block">
      <button
        ref={buttonRef}
        type="button"
        onClick={handleButtonClick}
        disabled={disabled}
        className={`
          inline-flex items-center gap-2 px-3 py-1.5 border text-sm font-medium rounded-md
          focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500
          ${
            disabled
              ? 'border-gray-300 text-gray-400 bg-gray-50 cursor-not-allowed'
              : 'border-primary-600 text-primary-700 bg-primary-50 hover:bg-primary-100'
          }
        `}
        aria-haspopup="menu"
        aria-expanded={isOpen}
      >
        {icon}
        <span>{label}</span>
        <ChevronDown
          className={`h-3 w-3 transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div
          ref={dropdownRef}
          role="menu"
          aria-orientation="vertical"
          className={`
            absolute right-0 z-50 mt-1 min-w-[200px] bg-white rounded-lg shadow-lg border border-gray-200
            ${openUpward ? 'bottom-full mb-1' : 'top-full mt-1'}
          `}
          style={{ maxHeight: '320px', overflowY: 'auto' }}
        >
          <div className="py-1">
            {actions.map((action, index) => (
              <button
                key={action.id}
                ref={(el) => (menuItemsRef.current[index] = el)}
                type="button"
                role="menuitem"
                disabled={action.disabled}
                onClick={() => !action.disabled && handleActionClick(action.id)}
                onMouseEnter={() => setFocusedIndex(index)}
                className={`
                  w-full flex items-center gap-3 px-4 py-2.5 text-left text-sm transition-colors
                  focus:outline-none focus:bg-gray-100
                  ${getVariantClasses(action.variant, action.disabled || false)}
                  ${focusedIndex === index ? 'bg-gray-100' : ''}
                `}
                title={action.disabled ? action.disabledReason : undefined}
                style={{ minHeight: '44px' }}
              >
                {action.icon && (
                  <span className="flex-shrink-0" aria-hidden="true">
                    {action.icon}
                  </span>
                )}
                <span className="flex-1">{action.label}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
