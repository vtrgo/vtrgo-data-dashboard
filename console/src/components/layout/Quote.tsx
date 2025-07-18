// File: console/src/components/layout/Quote.tsx
import React from "react";

type Props = {
  text: string;
  className?: string;
  style?: React.CSSProperties;
};

export function Quote({ text, className = "", style }: Props) {
  return (
    <blockquote
      className={`vtr-quote text-sm text-accent-foreground italic mt-1 px-[var(--quote-px,1.5rem)] py-[var(--quote-py,1rem)] ${className}`}
      style={style}
    >
      {text}
    </blockquote>
  );
}