import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatKey } from "@/utils/textFormat";

type FloatPanelProps = {
  title: string;
  values: Record<string, number>;
};

export function FloatPanel({ title, values }: FloatPanelProps) {
  const entries = Object.entries(values).sort(([keyA], [keyB]) =>
    keyA.localeCompare(keyB)
  );

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
          <tbody>
            {entries.map(([key, value]) => (
              <tr key={key} className="border-b border-dashed border-neutral-200">
                <td className="py-1.5 w-3/4">{formatKey(key)}</td>
                <td className="py-1.5 text-right w-1/4">
                  {(value as number).toFixed(2)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </CardContent>
    </Card>
  );
}
