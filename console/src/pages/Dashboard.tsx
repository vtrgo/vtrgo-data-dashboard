import { useState } from "react";
import { useStats } from "@/hooks/useStats";
import { PanelGrid } from "@/components/layout/PanelGrid";
import { BooleanPanel } from "@/components/panels/BooleanPanel";
import { FaultPanel } from "@/components/panels/FaultPanel";
import { FloatPanel } from "@/components/panels/FloatPanel";
import { describeTimeRange } from "@/utils/describeTimeRange";
import { inspirationalQuotes } from "@/utils/quotes";

function formatSectionTitle(name: string) {
  return name
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/([A-Z])([A-Z][a-z])/g, "$1 $2")
    .trim();
}

function groupFloatsBySection(floats: Record<string, number>) {
  const grouped: Record<string, { title: string; values: Record<string, number> }> = {};

  for (const fullKey in floats) {
    const parts = fullKey.split(".");
    // Expects keys like "Floats.SectionName.FieldName"
    if (parts.length !== 3 || parts[0] !== "Floats") continue;

    const [, section, subField] = parts;

    if (!grouped[section]) {
      grouped[section] = {
        title: formatSectionTitle(section),
        values: {},
      };
    }
    grouped[section].values[subField] = floats[fullKey];
  }

  return grouped;
}

export default function Dashboard() {
  const [timeRange, setTimeRange] = useState({ start: "-1h", stop: "now()" });
  const { data, loading } = useStats(timeRange);
  const randomQuote = inspirationalQuotes[Math.floor(Math.random() * inspirationalQuotes.length)];
  const timeRangeLabel = describeTimeRange(timeRange.start, timeRange.stop).label;
  const groupedFloats = data ? groupFloatsBySection(data.float_averages || {}) : {};

  return (
    <div className="min-h-screen bg-[url('/textures/paper-fiber.png')] bg-repeat">
      {data && (
        <div className="sticky top-0 z-60 bg-white text-black border-y-4 border-gray-300 bg-[url('/textures/paper-fiber.png')] px-6 py-6 shadow-md">
          <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
            <div className="text-center sm:text-left">
              <h1 className="text-5xl font-serif font-bold tracking-wide">The VTR Times</h1>
              <div className="text-sm text-muted-foreground italic mt-1">{randomQuote}</div>
            </div>
            <div className="flex flex-wrap gap-4 items-center text-sm">
              <select
                className="bg-[url('/textures/paper-fiber.png')] border px-2 py-2 text-sm bg-background rounded-md shadow-sm hover:border-primary focus:outline-none focus:ring-2 focus:ring-ring"
                onChange={(e) => setTimeRange({ start: e.target.value, stop: "now()" })}
                value={timeRange.start}
              >
                <option value="-1h">Last 1 hour</option>
                <option value="-3h">Last 3 hours</option>
                <option value="-6h">Last 6 hours</option>
                <option value="-12h">Last 12 hours</option>
                <option value="-1d">Last 1 day</option>
                <option value="-2d">Last 2 days</option>
                <option value="-3d">Last 3 days</option>
                <option value="-1w">Last 1 week</option>
                <option value="-2w">Last 2 weeks</option>
                <option value="-3w">Last 3 weeks</option>
                <option value="-1mo">Last 1 month</option>
              </select>
            </div>
          </div>
        </div>
      )}

      {loading && <div className="p-4 text-center text-muted-foreground">Loading...</div>}
      {!data && !loading && <div className="p-4 text-center text-muted-foreground">No data</div>}

      {data && (
        <main className="p-6 space-y-10">
          <section className="font-serif">
            <h2 className="text-xl uppercase tracking-widest text-neutral-500 mb-4 italic">
              System Status
            </h2>
            <div className="grid gap-6">
              <BooleanPanel booleans={data.boolean_percentages || {}} timeRangeLabel={timeRangeLabel} />
            </div>
          </section>

          <section className="font-serif">
            <h2 className="text-xl uppercase tracking-widest text-neutral-500 mb-4 italic">
              Float Averages
            </h2>
            <PanelGrid>
              {Object.entries(groupedFloats).map(([key, { title, values }]) => (
                <FloatPanel key={key} title={title} values={values} />
              ))}
            </PanelGrid>
          </section>

          <FaultPanel faults={data.fault_counts || {}} start={timeRange.start} stop={timeRange.stop} />
        </main>
      )}
    </div>
  );
}
