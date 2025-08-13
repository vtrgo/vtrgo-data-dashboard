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

  // Define the main project fields that have a specific display order.
  const mainKeys = [
    'Project Name',
    'Project Number',
    'Project Description',
    'Manufacturer',
    'Created On',
  ];

  // Sort the metadata. Main keys come first in their specified order,
  // followed by all other keys (electrical specs) sorted alphabetically.
  const sortedMeta = Object.entries(meta).sort(([keyA], [keyB]) => {
    const indexA = mainKeys.indexOf(keyA);
    const indexB = mainKeys.indexOf(keyB);

    if (indexA !== -1 && indexB !== -1) {
      return indexA - indexB; // Both are main keys, sort by predefined order
    }
    if (indexA !== -1) {
      return -1; // A is a main key, so it comes first
    }
    if (indexB !== -1) {
      return 1; // B is a main key, so it comes first
    }
    return keyA.localeCompare(keyB); // Neither are main keys, sort alphabetically
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

  // Any field not in mainKeys is considered an electrical field.
  const electricalFields = sortedMeta.filter(([key]) => !mainKeys.includes(key));

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