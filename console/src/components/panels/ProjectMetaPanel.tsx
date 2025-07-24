import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Info } from "lucide-react";

interface ProjectMetaPanelProps {
  meta: Record<string, string>;
  className?: string;
}

/**
 * A panel that displays project metadata in a compact, professional format.
 */
export function ProjectMetaPanel({ meta, className }: ProjectMetaPanelProps) {
  if (!meta || Object.keys(meta).length === 0) {
    return null;
  }

  // Define a preferred order for some important fields
  const orderedKeys = [
    'Project Name',
    'Project Number',
    'Project Description',
    'Manufacturer',
    'Created On',
    'Input Voltage',
    'Input Phase',
    'Input Frequency',
    'Input Current',
    'Control Voltage',
    'Output Power',
    'Enclosure Rating',
  ];

  const sortedMeta = Object.entries(meta).sort(([keyA], [keyB]) => {
    const indexA = orderedKeys.indexOf(keyA);
    const indexB = orderedKeys.indexOf(keyB);
    if (indexA !== -1 && indexB !== -1) return indexA - indexB;
    if (indexA !== -1) return -1;
    if (indexB !== -1) return 1;
    return keyA.localeCompare(keyB);
  });

  const projectName = meta['Project Name'] || 'Project Information';
  const projectNumber = meta['Project Number'];
  const manufacturer = meta['Manufacturer'];
  const createdOn = meta['Created On'];

  let formattedDate = '';
  if (createdOn) {
    const date = new Date(createdOn);
    // Check if the date is valid before formatting
    if (!isNaN(date.getTime())) {
      formattedDate = date.toLocaleString('default', { month: 'long', year: 'numeric' });
    }
  }

  const electricalFields = sortedMeta.filter(([key]) =>
    !['Project Name', 'Project Number', 'Project Description', 'Manufacturer', 'Created On'].includes(key)
  );

  return (
    <Card className={className}>
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Info className="h-5 w-5 text-primary" />
          <span>{projectName}</span>
        </CardTitle>
        <div className="space-y-1 pt-1 text-sm text-muted-foreground">
          {meta['Project Description'] && <p>{meta['Project Description']}</p>}
          {manufacturer && <p>Manufactured by: {manufacturer}{formattedDate && ` (${formattedDate})`}</p>}
          {projectNumber && <p>Serial #{projectNumber}</p>}
        </div>
      </CardHeader>
      {electricalFields.length > 0 && (
        <CardContent className="pt-0">
          <div className="border-t pt-4">
            <h3 className="text-xs uppercase tracking-wider text-muted-foreground mb-3">Electrical Specifications</h3>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-x-6 gap-y-3 text-sm">
              {electricalFields.map(([key, value]) => (
                  <div key={key} className="flex flex-col">
                    <span className="text-xs text-muted-foreground">{key}</span>
                    <span className="font-semibold text-foreground truncate" title={value}>{value}</span>
                  </div>
                ))}
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  );
}