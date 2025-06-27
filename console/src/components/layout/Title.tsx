// File: console/src/components/layout/Title.tsx
type Props = {
  text: string;
};

export function Title({ text }: Props) {
  return (
    <h1 className="vtr-title vtr-title text-5xl sm:text-6xl font-extrabold text-white bg-black">
      {text}
    </h1>
  );
}
