// file: console/src/components/panels/FloatPanel.tsx
import { Panel } from '@/components/ui/panel';
import { formatKey } from '@/utils/textFormat';

type FloatPanelProps = {
  title: string;
  values: Record<string, number>;
  className?: string;
};

export function FloatPanel({ title, values, className }: FloatPanelProps) {
  const entries = Object.entries(values).sort(([keyA], [keyB]) =>
    keyA.localeCompare(keyB)
  );
  if (!entries.length) return null;

  return (
    <Panel title={title} className={className}>
      <table className="w-full">
        <tbody>
          {entries.map(([key, value]) => (
            <tr key={key}>
              <td className="w-3/4">{formatKey(key)}</td>
              <td className="w-1/4 text-right">{value.toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
