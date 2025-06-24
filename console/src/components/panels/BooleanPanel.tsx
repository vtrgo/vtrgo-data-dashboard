import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatKey } from "@/utils/textFormat";

type BooleanPanelProps = {
  title: string;
  values: Record<string, number>;
};

export function BooleanPanel({ title, values }: BooleanPanelProps) {
  const entries = Object.entries(values).sort(([, a], [, b]) => (b as number) - (a as number));

  if (!entries.length) return null;

  return (
    <Card className="bg-[url('/textures/paper-fiber.png')] border border-neutral-300 shadow-inner font-serif">
      <CardHeader className="p-5 pb-2 border-b border-dashed border-neutral-300">
        <CardTitle className="text-base font-black uppercase tracking-widest text-black">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-5 pt-3">
        <table className="w-full text-[0.95rem] text-black font-medium table-fixed">
          <thead>
            <tr className="uppercase text-xs text-neutral-500 border-b border-neutral-300">
              <th className="pb-1 text-left w-3/4">Name</th>
              <th className="pb-1 text-right w-1/4">% True</th>
            </tr>
          </thead>
          <tbody>
            {entries.map(([key, value]) => (
              <tr key={key} className="border-b border-dashed border-neutral-200">
                <td className="py-1.5">{formatKey(key)}</td>
                <td className="py-1.5 text-right">{(value as number).toFixed(1)}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </CardContent>
    </Card>
  );
}
