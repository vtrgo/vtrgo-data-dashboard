import { useState, useEffect, useMemo } from "react";
import { useStats } from "@/hooks/useStats";
import { PanelGrid } from "@/components/layout/PanelGrid";
import { Title } from "@/components/layout/Title";
import { BooleanPanel } from "@/components/panels/BooleanPanel";
import { FloatPanel } from "@/components/panels/FloatPanel";
import { FloatAreaChartPanel } from "@/components/panels/FloatAreaChartPanel";
import { FaultBarChartPanel } from "@/components/panels/FaultBarChartPanel";
import { getRandomTitle } from "@/utils/titles";
import { getAvailableThemes } from "@/utils/getThemes";
import { describeTimeRange } from "@/utils/describeTimeRange";
import { inspirationalQuotes } from "@/utils/quotes";
import { formatKey } from "@/utils/textFormat"; // Ensure formatKey is imported for dropdown
import { Settings2 } from "lucide-react";
import { Button } from "@/components/ui/button"; // Import Button
import { ConfigDrawer } from "@/components/layout/ConfigDrawer"; // Import ConfigDrawer
import { ThemeProvider } from "@/components/ui/theme-provider";
import { Quote } from "@/components/layout/Quote";

function formatSectionTitle(name: string) {
  return name
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/([A-Z])([A-Z][a-z])/g, "$1 $2")
    .trim();
}

function groupBooleansBySection(booleans: Record<string, number>) {
  const grouped: Record<string, { title: string; values: Record<string, number> }> = {};

  for (const fullKey in booleans) {
    if (fullKey.startsWith("FaultBits.")) continue;
    const parts = fullKey.split(".");
    if (parts.length < 2) continue;
    const section = parts[0];
    const subField = parts.slice(1).join(".");
    if (!grouped[section]) {
      grouped[section] = { title: formatSectionTitle(section), values: {} };
    }
    grouped[section].values[subField] = booleans[fullKey];
  }
  return grouped;
}

function groupFloatsBySection(floats: Record<string, number>) {
  const grouped: Record<string, { title: string; values: Record<string, number> }> = {};
  for (const fullKey in floats) {
    const parts = fullKey.split(".");
    if (parts.length !== 3 || parts[0] !== "Floats") continue;
    const [, section, subField] = parts;
    if (!grouped[section]) {
      grouped[section] = { title: formatSectionTitle(section), values: {} };
    }
    grouped[section].values[subField] = floats[fullKey];
  }
  return grouped;
}

export default function Dashboard() {
  const [timeRange, setTimeRange] = useState({ start: "-1h", stop: "now()" });
  const [selectedFloatField, setSelectedFloatField] = useState<string | null>(null);
  const { data, loading, error } = useStats(timeRange, 60000);
  const [showConfig, setShowConfig] = useState(false);
  const [randomTitle, setRandomTitle] = useState("");
  const randomQuote = inspirationalQuotes[Math.floor(Math.random() * inspirationalQuotes.length)];
  const timeRangeLabel = describeTimeRange(timeRange.start, timeRange.stop).label;
  const [themes, setThemes] = useState<string[]>([]);
  const [themeIndex, setThemeIndexState] = useState(0);
  const [enableTheming, setEnableTheming] = useState(true);
  const currentTheme = enableTheming && themes.length > 0 ? themes[themeIndex] : "default";

  useEffect(() => {
    if (
      data &&
      data.float_averages &&
      Object.keys(data.float_averages).length > 0 &&
      selectedFloatField === null
    ) {
      setSelectedFloatField(Object.keys(data.float_averages)[0]);
    }
  }, [data, selectedFloatField]);

  useEffect(() => {
    setRandomTitle(getRandomTitle());
  }, []);

  useEffect(() => {
    const loaded = getAvailableThemes();
    if (loaded.length > 0) setThemes(loaded);

    // Load themeIndex from localStorage
    const storedThemeIndex = localStorage.getItem("vtr-title-theme-index");
    if (storedThemeIndex !== null && !isNaN(Number(storedThemeIndex))) {
      setThemeIndexState(Number(storedThemeIndex));
    }
  }, []);

  // Custom setter to persist themeIndex
  const setThemeIndex = (idx: number) => {
    setThemeIndexState(idx);
    localStorage.setItem("vtr-title-theme-index", String(idx));
  };

  const groupedFloats = useMemo(() => (data ? groupFloatsBySection(data.float_averages || {}) : {}), [data]);
  const groupedBooleans = useMemo(() => (data ? groupBooleansBySection(data.boolean_percentages || {}) : {}), [data]);

  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
    <div className={`theme-${currentTheme} min-h-screen bg-background relative`}>
      <div className="sticky top-0 z-60 bg-background text-foreground border-b shadow px-6 py-6">
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 w-full">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" onClick={() => setShowConfig(true)}>
              <Settings2 className="h-5 w-5" />
              <span className="sr-only">Open Configuration</span>
            </Button>
            <div className="text-center sm:text-left">
              <Title text={randomTitle} />
              {/* <div className="text-sm text-muted-foreground italic mt-1">{randomQuote}</div> */}
              <Quote text={randomQuote} className="mt-1" />
            </div>
          </div>
          <div className="flex flex-wrap gap-4 items-center text-sm">
            <select className="border px-2 py-2 text-sm bg-background rounded-md shadow-sm hover:border-primary focus:outline-none focus:ring-2 focus:ring-ring" onChange={(e) => setTimeRange({ start: e.target.value, stop: "now()" })} value={timeRange.start}>
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
            {data && data.float_averages && Object.keys(data.float_averages).length > 0 && (
              <select className="border px-2 py-2 text-sm bg-background rounded-md shadow-sm hover:border-primary focus:outline-none focus:ring-2 focus:ring-ring" onChange={(e) => setSelectedFloatField(e.target.value)} value={selectedFloatField || ''}>
                <option value="" disabled>Select a float field</option>
                {Object.keys(data.float_averages).map((fieldKey) => (
                  <option key={fieldKey} value={fieldKey}>{formatKey(fieldKey)}</option>
                ))}
              </select>
            )}
          </div>
        </div>
      </div>

      <div className="absolute left-0 bottom-0 top-[144px] z-50">
        <ConfigDrawer
          open={showConfig}
          onOpenChange={setShowConfig}
          themes={themes}
          themeIndex={themeIndex}
          setThemeIndex={setThemeIndex}
          enableTheming={enableTheming}
          setEnableTheming={setEnableTheming}
        />
      </div>

      {loading && <div className="p-4 text-center text-muted-foreground">Loading dashboard data...</div>}
      {error && <div className="p-4 text-center text-red-500">Error loading dashboard data: {error.message}</div>}
      {!data && !loading && <div className="p-4 text-center text-muted-foreground">No dashboard data available.</div>}

      {selectedFloatField && (
        <section className="font-serif">
          <h2 className="pl-9 pt-3 text-xl uppercase tracking-widest text-muted-foreground mb-4 italic">Performance Data</h2>
          <PanelGrid>
            <FloatAreaChartPanel field={selectedFloatField} start={timeRange.start} stop={timeRange.stop} intervalMs={60000} />
            <FaultBarChartPanel faults={data.fault_counts || {}} />
          </PanelGrid>
        </section>
      )}

      {data && (
        <main className="p-6 space-y-10">
          <section className="font-serif">
            <h2 className="text-xl uppercase tracking-widest text-muted-foreground mb-4 italic">System Status (% True over {timeRangeLabel})</h2>
            <PanelGrid>
              {Object.entries(groupedBooleans).map(([key, { title, values }]) => (
                <BooleanPanel key={key} title={title} values={values} />
              ))}
            </PanelGrid>
          </section>

          <section className="font-serif">
            <h2 className="text-xl uppercase tracking-widest text-muted-foreground mb-4 italic">Float Averages</h2>
            <PanelGrid>
              {Object.entries(groupedFloats).map(([key, { title, values }]) => (
                <FloatPanel key={key} title={title} values={values} />
              ))}
            </PanelGrid>
          </section>
        </main>
      )}
    </div>
    </ThemeProvider>
  );
}
