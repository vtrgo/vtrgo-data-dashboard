// File: console/src/components/layout/Title.tsx
import React from "react";

type Props = {
  text: string;
  className?: string;
  style?: React.CSSProperties;
};

export function Title({ text, className = "", style }: Props) {
  return (
    <h1
      className={`vtr-title text-5xl sm:text-6xl
        px-[var(--title-px,1.5rem)] py-[var(--title-py,1rem)] ${className}`}
      style={style}
    >
      {text}
    </h1>
  );
}

// Usage example (not included in file):
// <Title text="Hello" style={{ "--title-px": "2rem", "--title-py": "2rem" } as React.CSSProperties} />
