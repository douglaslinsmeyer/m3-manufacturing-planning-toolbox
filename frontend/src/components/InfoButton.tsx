import { useState, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';

interface InfoButtonProps {
  title: string;
  children: React.ReactNode;
  className?: string;
}

function InfoIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="10" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 16v-4m0-4h.01" />
    </svg>
  );
}

export function InfoButton({ title, children, className = '' }: InfoButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const popoverRef = useRef<HTMLDivElement>(null);
  const [position, setPosition] = useState({ top: 0, left: 0 });

  // Calculate popover position
  useEffect(() => {
    if (isOpen && buttonRef.current) {
      const buttonRect = buttonRef.current.getBoundingClientRect();
      const popoverWidth = 384; // max-w-sm = 384px
      const popoverHeight = 200; // estimated

      // Calculate initial position (below and to the right of button)
      let top = buttonRect.bottom + 8;
      let left = buttonRect.left;

      // Adjust if popover would go off-screen to the right
      if (left + popoverWidth > window.innerWidth) {
        left = window.innerWidth - popoverWidth - 16;
      }

      // Adjust if popover would go off-screen at the bottom
      if (top + popoverHeight > window.innerHeight) {
        top = buttonRect.top - popoverHeight - 8;
      }

      setPosition({ top, left });
    }
  }, [isOpen]);

  // Handle click outside to close popover
  useEffect(() => {
    if (!isOpen) return;

    function handleClickOutside(event: MouseEvent) {
      if (
        popoverRef.current &&
        buttonRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  // Handle ESC key to close popover
  useEffect(() => {
    if (!isOpen) return;

    function handleEscKey(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    document.addEventListener('keydown', handleEscKey);
    return () => {
      document.removeEventListener('keydown', handleEscKey);
    };
  }, [isOpen]);

  // Handle Enter/Space key to open popover
  function handleKeyDown(event: React.KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      setIsOpen(!isOpen);
    }
  }

  return (
    <>
      <button
        ref={buttonRef}
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={handleKeyDown}
        className={`inline-flex items-center justify-center text-blue-600 hover:text-blue-700 transition-colors ${className}`}
        aria-label="More information"
        aria-expanded={isOpen}
      >
        <InfoIcon className="h-4 w-4" />
      </button>

      {isOpen &&
        createPortal(
          <div
            ref={popoverRef}
            className="fixed z-50 bg-white shadow-lg rounded-lg ring-1 ring-blue-200 p-4 max-w-sm"
            style={{
              top: `${position.top}px`,
              left: `${position.left}px`,
            }}
          >
            <h3 className="text-blue-900 font-semibold mb-2">{title}</h3>
            <div className="text-slate-600 text-sm">{children}</div>
          </div>,
          document.body
        )}
    </>
  );
}
