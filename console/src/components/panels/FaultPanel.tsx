// file: console/src/components/panels/FaultPanel.tsx
import { Panel } from '@/components/ui/panel';
import { formatFaultKey } from '@/utils/textFormat';
import { describeTimeRange } from '@/utils/describeTimeRange';

type FaultPanelProps = {
  faults: Record<string, number>;
  start: string;
  stop: string;
};

export function FaultPanel({ faults, start, stop }: FaultPanelProps) {
  const entries = Object.entries(faults)
    .sort(([, a], [, b]) => b - a)
    .filter(([k]) => k.startsWith('FaultBits.'));

  if (!entries.length) return null;

  const { label } = describeTimeRange(start, stop);

  return (
    <section className="mt-10 border-t border-neutral-400 pt-6">
      <h2 className="text-xl uppercase tracking-widest text-neutral-500 mb-4 italic">
        Fault Counts
      </h2>
      <div className="grid gap-6">
        <Panel title={`Fault Counts for the ${label}`}>
          <table className="w-full">
            <thead>
              <tr>
                <th className="text-left">Fault Name</th>
                <th className="text-right">Count</th>
              </tr>
            </thead>
            <tbody>
              {entries.map(([key, value]) => (
                <tr key={key}>
                  <td>{formatFaultKey(key)}</td>
                  <td className="text-right">{value}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Panel>
      </div>
    </section>
  );
}
