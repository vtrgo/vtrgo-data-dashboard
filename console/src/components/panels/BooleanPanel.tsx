// file: console/src/components/panels/BooleanPanel.tsx
import { Panel } from '@/components/ui/panel';
import { formatKey } from '@/utils/textFormat';

type BooleanPanelProps = {
  title: string;
  values: Record<string, number>;
};

export function BooleanPanel({ title, values }: BooleanPanelProps) {
  const entries = Object.entries(values).sort(([, a], [, b]) => b - a);
  if (!entries.length) return null;

  return (
    <Panel title={title}>
      <table className="w-full">
        <thead>
          <tr>
            <th className="text-left">Name</th>
            <th className="text-right">% True</th>
          </tr>
        </thead>
        <tbody>
          {entries.map(([key, value]) => (
            <tr key={key}>
              <td>{formatKey(key)}</td>
              <td className="text-right">{value.toFixed(1)}%</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Panel>
  );
}
