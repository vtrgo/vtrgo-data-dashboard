import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatFaultKey } from "@/utils/textFormat";
import { describeTimeRange } from "@/utils/describeTimeRange";

type FaultPanelProps = {
  faults: Record<string, number>;
  start: string;
  stop: string;
};

export function FaultPanel({ faults, start, stop }: FaultPanelProps) {
  const entries = Object.entries(faults)
    .sort(([, a], [, b]) => (b as number) - (a as number))
    .filter(([k]) => k.startsWith("FaultBits."));

  if (!entries.length) return null;

  return (
    <section className="mt-10 border-t border-neutral-400 pt-6 font-serif">
      <h2 className="text-xl uppercase tracking-widest text-neutral-500 mb-4 italic">
        Fault Counts
      </h2>
      <div className="grid gap-6">
<Card className="bg-[url('/textures/paper-fiber.png')] border border-neutral-300 shadow-inner font-serif">
  <CardHeader className="p-5 pb-2 border-b border-dashed border-neutral-300">
    <CardTitle className="text-base font-black uppercase tracking-widest text-black">
      Fault Counts for the {describeTimeRange(start, stop).label}
    </CardTitle>
  </CardHeader>
  <CardContent className="p-5 pt-3">
    <table className="w-full text-[0.95rem] text-black font-medium table-fixed">
      <thead>
        <tr className="uppercase text-xs text-neutral-500 border-b border-neutral-300">
          <th className="pb-1 text-left w-3/4">Fault Name</th>
          <th className="pb-1 text-right w-1/4">Count</th>
        </tr>
      </thead>
      <tbody>
        {entries.map(([key, value]) => (
          <tr key={key} className="border-b border-dashed border-neutral-200">
            <td className="py-1.5">{formatFaultKey(key)}</td>
            <td className="py-1.5 text-right">{value}</td>
          </tr>
        ))}
      </tbody>
    </table>
  </CardContent>
</Card>
      </div>
    </section>
  );
}
